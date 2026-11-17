package httpapi_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func doHost(t *testing.T, env *testEnv, method, path, token, companyID, host string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(encoded)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Host = host
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if companyID != "" {
		req.Header.Set("X-Company-Id", companyID)
	}
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)
	return rec
}

func createTestBranch(t *testing.T, env *testEnv, token, companyID, name, code string) string {
	t.Helper()
	rec := do(t, env.router, http.MethodPost, "/api/v1/companies/me/branches", token, companyID, map[string]any{
		"name": name, "code": code, "address": "12 Port Harcourt Road",
		"city": "Aba", "state": "Abia", "country": "NG",
		"phone": "+2348000000000", "email": strings.ToLower(code) + "@example.com",
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var body struct {
		Branch struct {
			ID             string         `json:"id"`
			Code           string         `json:"code"`
			Status         string         `json:"status"`
			Email          *string        `json:"email"`
			OperatingHours map[string]any `json:"operating_hours"`
		} `json:"branch"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NotEmpty(t, body.Branch.ID)
	require.Equal(t, "active", body.Branch.Status)
	require.NotNil(t, body.Branch.Email)
	require.Equal(t, strings.ToLower(code)+"@example.com", *body.Branch.Email)
	require.NotNil(t, body.Branch.OperatingHours)
	return body.Branch.ID
}

func memberIDForUser(t *testing.T, env *testEnv, companyID, userID string) string {
	t.Helper()
	var memberID uuid.UUID
	require.NoError(t, env.pool.QueryRow(t.Context(),
		`SELECT id FROM company_members WHERE company_id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		companyID, userID).Scan(&memberID))
	return memberID.String()
}

func TestCompanyProfileBrandingAndUploads(t *testing.T) {
	env := newTestEnv(t)
	owner := registerCompany(t, env, "Profile Freight", "profile-owner@example.com")
	companyID := owner.Companies[0].CompanyID

	rec := do(t, env.router, http.MethodPatch, "/api/v1/companies/me", owner.AccessToken, companyID, map[string]any{
		"description": "Intercity freight across the south-east.",
		"operating_hours": map[string]any{
			"mon": map[string]any{"open": "08:00", "close": "18:00"},
		},
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me", owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var detail struct {
		Company struct {
			Description    string         `json:"description"`
			OperatingHours map[string]any `json:"operating_hours"`
		} `json:"company"`
		Membership struct {
			Permissions []string `json:"permissions"`
		} `json:"membership"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &detail))
	require.Equal(t, "Intercity freight across the south-east.", detail.Company.Description)
	require.Contains(t, detail.Company.OperatingHours, "mon")
	require.Contains(t, detail.Membership.Permissions, "branches.create")

	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/branding", owner.AccessToken, companyID, map[string]any{
		"public_name":     "Profile Express",
		"pwa_short_name":  "ProfileX",
		"primary_color":   "#123456",
		"secondary_color": "#654321",
		"description":     "Public brand description",
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	// A valid PNG uploads and is served back from the local upload directory.
	pngBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=")
	require.NoError(t, err)
	rec = uploadBranding(t, env, owner.AccessToken, companyID, "logo", pngBytes)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var uploaded struct {
		Branding struct {
			LogoURL *string `json:"logo_url"`
		} `json:"branding"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &uploaded))
	require.NotNil(t, uploaded.Branding.LogoURL)
	require.Contains(t, *uploaded.Branding.LogoURL, "/uploads/")

	path := strings.TrimPrefix(*uploaded.Branding.LogoURL, "http://localhost:8080")
	fileRec := do(t, env.router, http.MethodGet, path, "", "", nil)
	require.Equal(t, http.StatusOK, fileRec.Code)

	// Read-only roles cannot change the profile or branding.
	staff := registerCustomer(t, env, "profile-staff@example.com")
	addMembership(t, env, companyID, staff.User.ID, systemRoleID(t, env, "operations_staff"))
	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me", staff.AccessToken, companyID, map[string]any{"name": "Nope"})
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/branding", staff.AccessToken, companyID, map[string]any{"primary_color": "#000000"})
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func uploadBranding(t *testing.T, env *testEnv, token, companyID, kind string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("kind", kind))
	part, err := writer.CreateFormFile("file", "brand.png")
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/companies/me/branding/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Company-Id", companyID)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)
	return rec
}

func TestBranchAccessAndCrossCompanyIsolation(t *testing.T) {
	env := newTestEnv(t)
	owner := registerCompany(t, env, "Branch Freight", "branch-owner@example.com")
	companyID := owner.Companies[0].CompanyID
	other := registerCompany(t, env, "Other Branch Freight", "branch-other@example.com")
	otherCompanyID := other.Companies[0].CompanyID

	first := createTestBranch(t, env, owner.AccessToken, companyID, "Aba HQ", "ABA")
	second := createTestBranch(t, env, owner.AccessToken, companyID, "Lagos Hub", "LOS")

	rec := do(t, env.router, http.MethodPost, "/api/v1/companies/me/branches", owner.AccessToken, companyID, map[string]any{
		"name": "Duplicate", "code": "aba", "address": "1 Test Road", "city": "Aba", "state": "Abia",
	})
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())

	manager := registerCustomer(t, env, "branch-manager@example.com")
	managerMemberID := addMembership(t, env, companyID, manager.User.ID, systemRoleID(t, env, "branch_manager"))
	_, err := env.pool.Exec(t.Context(), `INSERT INTO branch_members (branch_id, member_id) VALUES ($1, $2)`, first, managerMemberID)
	require.NoError(t, err)

	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me/branches", manager.AccessToken, companyID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var list struct {
		Branches []struct {
			ID string `json:"id"`
		} `json:"branches"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list.Branches, 1)
	require.Equal(t, first, list.Branches[0].ID)

	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me/branches/"+second, manager.AccessToken, companyID, nil)
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/branches/"+first, manager.AccessToken, companyID, map[string]any{
		"name": "Aba Headquarters", "code": "ABA", "address": "14 Port Harcourt Road",
		"city": "Aba", "state": "Abia", "country": "NG",
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/branches", manager.AccessToken, companyID, map[string]any{
		"name": "Unauthorised", "code": "BAD", "address": "1 Test Road", "city": "Aba", "state": "Abia",
	})
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

	// Managers may not reassign their branch to another member.
	ownerMemberID := memberIDForUser(t, env, companyID, owner.User.ID)
	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/branches/"+first, manager.AccessToken, companyID, map[string]any{
		"name": "Aba HQ", "code": "ABA", "address": "14 Port Harcourt Road",
		"city": "Aba", "state": "Abia", "manager_member_id": ownerMemberID,
	})
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me/branches/"+first, other.AccessToken, otherCompanyID, nil)
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/branches/"+first+"/staff", owner.AccessToken, companyID, map[string]any{
		"member_ids": []string{managerMemberID.String()},
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me/branches/"+first, owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var branchDetail struct {
		Branch struct {
			Status string `json:"status"`
		} `json:"branch"`
		Staff []struct {
			ID string `json:"id"`
		} `json:"staff"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &branchDetail))
	require.Len(t, branchDetail.Staff, 1)

	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/branches/"+first+"/deactivate", owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &branchDetail))
	require.Equal(t, "inactive", branchDetail.Branch.Status)

	// There is deliberately no delete endpoint.
	rec = do(t, env.router, http.MethodDelete, "/api/v1/companies/me/branches/"+first, owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
}

func TestInvitationAcceptanceExpiryAndSuspension(t *testing.T) {
	env := newTestEnv(t)
	owner := registerCompany(t, env, "Invite Freight", "invite-owner@example.com")
	companyID := owner.Companies[0].CompanyID
	branchID := createTestBranch(t, env, owner.AccessToken, companyID, "Aba HQ", "ABA")

	customer := registerCustomer(t, env, "invited-staff@example.com")
	invitation := inviteStaff(t, env, owner.AccessToken, companyID, customer.User.Email, "operations_staff", []string{branchID})

	preview := do(t, env.router, http.MethodGet, "/api/v1/invitations/"+invitation.Token, "", "", nil)
	require.Equal(t, http.StatusOK, preview.Code, preview.Body.String())
	var previewBody struct {
		Invitation struct {
			AccountExists bool     `json:"account_exists"`
			BranchIDs     []string `json:"branch_ids"`
			RoleKey       string   `json:"role_key"`
		} `json:"invitation"`
	}
	require.NoError(t, json.Unmarshal(preview.Body.Bytes(), &previewBody))
	require.True(t, previewBody.Invitation.AccountExists)
	require.Equal(t, []string{branchID}, previewBody.Invitation.BranchIDs)

	rec := do(t, env.router, http.MethodPost, "/api/v1/invitations/"+invitation.Token+"/accept", "", "", map[string]any{})
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "sign_in_required")

	rec = do(t, env.router, http.MethodPost, "/api/v1/invitations/"+invitation.Token+"/accept", customer.AccessToken, "", map[string]any{})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var session authResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &session))
	require.Contains(t, session.Companies[0].Roles, "operations_staff")

	memberID := memberIDForUser(t, env, companyID, customer.User.ID)
	var branchCount int
	require.NoError(t, env.pool.QueryRow(t.Context(),
		`SELECT count(*) FROM branch_members WHERE branch_id = $1 AND member_id = $2`, branchID, memberID).Scan(&branchCount))
	require.Equal(t, 1, branchCount)

	rec = do(t, env.router, http.MethodPost, "/api/v1/invitations/"+invitation.Token+"/accept", customer.AccessToken, "", map[string]any{})
	require.Equal(t, http.StatusGone, rec.Code, rec.Body.String())

	// A brand new account is created from the invitation form.
	newInvite := inviteStaff(t, env, owner.AccessToken, companyID, "new-staff@example.com", "customer_service", nil)
	rec = do(t, env.router, http.MethodPost, "/api/v1/invitations/"+newInvite.Token+"/accept", "", "", map[string]any{
		"full_name": "New Staff", "password": "supersecret123",
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	// Expired invitations are rejected and resending issues a fresh token.
	expired := inviteStaff(t, env, owner.AccessToken, companyID, "expired-staff@example.com", "operations_staff", nil)
	_, err := env.pool.Exec(t.Context(), `UPDATE company_invitations SET expires_at = now() - interval '1 hour' WHERE id = $1`, expired.ID)
	require.NoError(t, err)
	rec = do(t, env.router, http.MethodPost, "/api/v1/invitations/"+expired.Token+"/accept", "", "", map[string]any{
		"full_name": "Expired Staff", "password": "supersecret123",
	})
	require.Equal(t, http.StatusGone, rec.Code, rec.Body.String())

	revoked := inviteStaff(t, env, owner.AccessToken, companyID, "revoked-staff@example.com", "operations_staff", nil)
	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/invitations/"+revoked.ID+"/resend", owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var resent struct {
		DevInviteToken string `json:"dev_invite_token"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resent))
	require.NotEmpty(t, resent.DevInviteToken)
	require.NotEqual(t, revoked.Token, resent.DevInviteToken)

	rec = do(t, env.router, http.MethodDelete, "/api/v1/companies/me/invitations/"+revoked.ID, owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	rec = do(t, env.router, http.MethodGet, "/api/v1/invitations/"+resent.DevInviteToken, "", "", nil)
	require.Equal(t, http.StatusGone, rec.Code, rec.Body.String())

	// Suspension invalidates the member's tenant access immediately.
	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/members/"+memberID, owner.AccessToken, companyID, map[string]any{
		"status": "suspended",
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me", customer.AccessToken, companyID, nil)
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

type invitationResult struct {
	ID    string
	Token string
}

func inviteStaff(t *testing.T, env *testEnv, token, companyID, email, roleKey string, branchIDs []string) invitationResult {
	t.Helper()
	rec := do(t, env.router, http.MethodPost, "/api/v1/companies/me/invitations", token, companyID, map[string]any{
		"email": email, "role_key": roleKey, "branch_ids": branchIDs,
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var body struct {
		Invitation struct {
			ID string `json:"id"`
		} `json:"invitation"`
		DevInviteToken string `json:"dev_invite_token"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NotEmpty(t, body.DevInviteToken)
	return invitationResult{ID: body.Invitation.ID, Token: body.DevInviteToken}
}

func TestCustomRolesAndGrantAuthority(t *testing.T) {
	env := newTestEnv(t)
	owner := registerCompany(t, env, "Role Freight", "role-owner@example.com")
	companyID := owner.Companies[0].CompanyID

	rec := do(t, env.router, http.MethodPost, "/api/v1/companies/me/roles", owner.AccessToken, companyID, map[string]any{
		"name": "Dispatcher", "description": "Coordinates branch dispatch",
		"permissions": []string{"branches.view", "staff.view"},
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var created struct {
		Role struct {
			ID   string `json:"id"`
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"role"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	require.Equal(t, "Dispatcher", created.Role.Name)

	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/roles", owner.AccessToken, companyID, map[string]any{
		"name": "Bad", "permissions": []string{"not.a.permission"},
	})
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	staff := registerCustomer(t, env, "role-staff@example.com")
	memberID := addMembership(t, env, companyID, staff.User.ID, systemRoleID(t, env, "operations_staff")).String()
	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/roles", staff.AccessToken, companyID, map[string]any{
		"name": "Nope", "permissions": []string{"company.view"},
	})
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

	// Custom roles can be assigned to members and then cannot be deleted in use.
	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/members/"+memberID+"/role", owner.AccessToken, companyID, map[string]any{
		"role_key": created.Role.Key,
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/roles/"+created.Role.ID, owner.AccessToken, companyID, map[string]any{
		"name": "Dispatcher Lead", "permissions": []string{"branches.view"},
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodDelete, "/api/v1/companies/me/roles/"+created.Role.ID, owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())

	// The owner role can never be granted to another member.
	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/members/"+memberID+"/role", owner.AccessToken, companyID, map[string]any{
		"role_key": "owner",
	})
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/roles", owner.AccessToken, companyID, map[string]any{
		"name": "Unused", "permissions": []string{"branches.view"},
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var unused struct {
		Role struct {
			ID string `json:"id"`
		} `json:"role"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &unused))
	rec = do(t, env.router, http.MethodDelete, "/api/v1/companies/me/roles/"+unused.Role.ID, owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
}

func TestSubdomainUniquenessResolutionAndManifest(t *testing.T) {
	env := newTestEnv(t)
	first := registerCompany(t, env, "Acme Freight", "sub-acme@example.com")
	firstID := first.Companies[0].CompanyID
	second := registerCompany(t, env, "Beta Freight", "sub-beta@example.com")
	secondID := second.Companies[0].CompanyID

	rec := do(t, env.router, http.MethodPut, "/api/v1/companies/me/subdomain", first.AccessToken, firstID, map[string]any{"subdomain": "acme"})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/subdomain", second.AccessToken, secondID, map[string]any{"subdomain": "acme"})
	require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/subdomain", second.AccessToken, secondID, map[string]any{"subdomain": "api"})
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me/subdomain/availability?subdomain=free-route", first.AccessToken, firstID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), `"available":true`)

	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me/subdomain/availability?subdomain=acme", second.AccessToken, secondID, nil)
	require.Contains(t, rec.Body.String(), `"available":false`)

	rec = doHost(t, env, http.MethodGet, "/api/v1/public/branding", "", "", "acme.localhost", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var public struct {
		Company struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"company"`
		Subdomain string `json:"subdomain"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &public))
	require.Equal(t, firstID, public.Company.ID)
	require.Equal(t, "acme", public.Subdomain)

	rec = doHost(t, env, http.MethodGet, "/api/v1/public/manifest", "", "", "acme.localhost", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, "application/manifest+json", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Body.String(), `"short_name"`)

	rec = doHost(t, env, http.MethodGet, "/api/v1/public/branding", "", "", "unknown.localhost", nil)
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())

	// The apex serves the platform defaults, not another tenant's brand.
	rec = doHost(t, env, http.MethodGet, "/api/v1/public/branding", "", "", "localhost", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), `"company":null`)

	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/subdomain", "", firstID, map[string]any{"subdomain": "hijack"})
	require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}
