package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/bol-platform/backend/internal/auth"
	"github.com/bol-platform/backend/internal/config"
	"github.com/bol-platform/backend/internal/database"
	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/httpapi"
	"github.com/bol-platform/backend/internal/mailer"
	"github.com/bol-platform/backend/internal/ratelimit"
	"github.com/bol-platform/backend/internal/service"
)

type authResponse struct {
	User struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	} `json:"user"`
	Companies []struct {
		CompanyID string   `json:"company_id"`
		Roles     []string `json:"roles"`
	} `json:"companies"`
	AccessToken          string `json:"access_token"`
	DevVerificationToken string `json:"dev_verification_token"`
}

type testEnv struct {
	router *gin.Engine
	pool   *pgxpool.Pool
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	_, file, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(file), "..", "..", "migrations")
	require.NoError(t, database.Migrate("up", 0, "file://"+migrationsDir, databaseURL))

	pool, err := database.Connect(t.Context(), databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	// Delete tenant data only. TRUNCATE ... CASCADE would also wipe the global
	// role templates that the seed migration installs once.
	_, err = pool.Exec(t.Context(), `
		DELETE FROM audit_logs;
		DELETE FROM branch_members;
		DELETE FROM invitation_branches;
		DELETE FROM branches;
		DELETE FROM company_subdomains;
		DELETE FROM company_invitations;
		DELETE FROM refresh_tokens;
		DELETE FROM password_resets;
		DELETE FROM email_verifications;
		DELETE FROM member_roles;
		DELETE FROM role_permissions WHERE role_id IN (SELECT id FROM roles WHERE company_id IS NOT NULL);
		DELETE FROM roles WHERE company_id IS NOT NULL;
		DELETE FROM company_members;
		DELETE FROM company_branding;
		DELETE FROM company_settings;
		DELETE FROM companies;
		DELETE FROM users;`)
	require.NoError(t, err)

	cfg := &config.Config{
		Env:           "test",
		Port:          "0",
		AppURL:        "http://localhost:3000",
		DatabaseURL:   databaseURL,
		JWTSecret:     "integration-test-secret-at-least-32-characters",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    24 * time.Hour,
		CORSOrigins:   []string{"http://localhost:3000"},
		ExposeDevKeys: true,
		UploadDir:     filepath.Join(t.TempDir(), "uploads"),
		APIPublicURL:  "http://localhost:8080",
		AppDomain:     "localhost",
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	manager := auth.NewManager(cfg.JWTSecret, cfg.AccessTTL)
	svc := service.New(pool, manager, nil, cfg, &mailer.LogMailer{Logger: logger, AppURL: cfg.AppURL}, logger)

	return &testEnv{
		router: httpapi.New(cfg, svc, manager, ratelimit.New(nil), logger),
		pool:   pool,
	}
}

func do(t *testing.T, router http.Handler, method, path, token, companyID string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(encoded)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if companyID != "" {
		req.Header.Set("X-Company-Id", companyID)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeAuth(t *testing.T, rec *httptest.ResponseRecorder) authResponse {
	t.Helper()
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var parsed authResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &parsed))
	require.NotEmpty(t, parsed.AccessToken)
	return parsed
}

func registerCompany(t *testing.T, env *testEnv, name, email string) authResponse {
	t.Helper()
	rec := do(t, env.router, http.MethodPost, "/api/v1/companies", "", "", map[string]any{
		"company_name":   name,
		"business_email": email,
		"phone":          "+2348000000000",
		"business_type":  "intercity",
		"country":        "NG",
		"state":          "Abia",
		"city":           "Aba",
		"address":        "12 Port Harcourt Road",
		"contact_name":   "Owner " + name,
		"password":       "supersecret123",
	})
	return decodeAuth(t, rec)
}

func registerCustomer(t *testing.T, env *testEnv, email string) authResponse {
	t.Helper()
	rec := do(t, env.router, http.MethodPost, "/api/v1/auth/register", "", "", map[string]any{
		"full_name": "Customer " + email,
		"email":     email,
		"password":  "supersecret123",
		"phone":     "+2348111111111",
	})
	return decodeAuth(t, rec)
}

func TestCompanyRegistrationAndTenantIsolation(t *testing.T) {
	env := newTestEnv(t)

	ownerA := registerCompany(t, env, "Aba Freight", "owner-a@example.com")
	ownerB := registerCompany(t, env, "Lagos Linehaul", "owner-b@example.com")

	require.Len(t, ownerA.Companies, 1)
	require.Len(t, ownerB.Companies, 1)
	companyA := ownerA.Companies[0].CompanyID
	companyB := ownerB.Companies[0].CompanyID
	require.NotEqual(t, companyA, companyB)
	require.Contains(t, ownerA.Companies[0].Roles, "owner")

	// A valid token cannot read another company even when it supplies that id.
	rec := do(t, env.router, http.MethodGet, "/api/v1/companies/me", ownerA.AccessToken, companyB, nil)
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me", ownerA.AccessToken, "", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me", ownerA.AccessToken, companyB, map[string]any{
		"name": "Taken over",
	})
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me", ownerA.AccessToken, companyA, map[string]any{
		"city": "Umuahia",
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me/members", ownerA.AccessToken, companyA, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var members struct {
		Members []struct {
			Roles []string `json:"roles"`
		} `json:"members"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &members))
	require.Len(t, members.Members, 1)
	require.Contains(t, members.Members[0].Roles, "owner")
}

func TestRefreshRotationRevokesReusedTokenFamily(t *testing.T) {
	env := newTestEnv(t)
	session := registerCompany(t, env, "Rotation Freight", "rotate@example.com")

	login := do(t, env.router, http.MethodPost, "/api/v1/auth/login", "", "", map[string]any{
		"email":    "rotate@example.com",
		"password": "supersecret123",
	})
	require.Equal(t, http.StatusOK, login.Code, login.Body.String())
	oldRefresh := cookieValue(t, login, "bol_refresh")
	require.NotEmpty(t, oldRefresh)

	rotate := doWithCookie(t, env.router, "/api/v1/auth/refresh", "bol_refresh", oldRefresh)
	require.Equal(t, http.StatusOK, rotate.Code, rotate.Body.String())
	newRefresh := cookieValue(t, rotate, "bol_refresh")
	require.NotEmpty(t, newRefresh)
	require.NotEqual(t, oldRefresh, newRefresh)

	// Replaying the old token is treated as theft and kills the family.
	replay := doWithCookie(t, env.router, "/api/v1/auth/refresh", "bol_refresh", oldRefresh)
	require.Equal(t, http.StatusUnauthorized, replay.Code, replay.Body.String())

	// The rotated token is now unusable too.
	afterReuse := doWithCookie(t, env.router, "/api/v1/auth/refresh", "bol_refresh", newRefresh)
	require.Equal(t, http.StatusUnauthorized, afterReuse.Code, afterReuse.Body.String())

	require.NotEmpty(t, session.AccessToken)
}

func TestMemberPermissionsAreEnforced(t *testing.T) {
	env := newTestEnv(t)

	owner := registerCompany(t, env, "Permission Freight", "perm-owner@example.com")
	companyID := owner.Companies[0].CompanyID
	member := registerCustomer(t, env, "perm-staff@example.com")

	roleID := systemRoleID(t, env, "operations_staff")
	memberID := addMembership(t, env, companyID, member.User.ID, roleID)

	// A plain member can read the company but cannot change settings.
	rec := do(t, env.router, http.MethodGet, "/api/v1/companies/me", member.AccessToken, companyID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/settings", member.AccessToken, companyID, map[string]any{
		"currency": "USD",
	})
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

	// The owner has the settings permission.
	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/settings", owner.AccessToken, companyID, map[string]any{
		"currency": "NGN",
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	// A member of company A still cannot act on company B.
	other := registerCompany(t, env, "Other Freight", "perm-other@example.com")
	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me", member.AccessToken, other.Companies[0].CompanyID, nil)
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

	require.NotEqual(t, uuid.Nil, memberID)
}

func TestCompanyVerificationRequiresVerifiedEmailAndWhatsApp(t *testing.T) {
	env := newTestEnv(t)
	owner := registerCompany(t, env, "Verify Freight", "verify@example.com")
	companyID := owner.Companies[0].CompanyID
	require.NotEmpty(t, owner.DevVerificationToken)

	rec := do(t, env.router, http.MethodPost, "/api/v1/companies/me/submit-verification", owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPost, "/api/v1/auth/verify-email", "", "", map[string]any{
		"token": owner.DevVerificationToken,
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/submit-verification", owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "whatsapp_unverified")

	_, err := env.pool.Exec(t.Context(), `
		INSERT INTO user_phone_verifications
			(user_id, phone, token_hash, expires_at, verified_at, last_sent_at, window_started_at)
		VALUES ($1, '+2348000000000', '', now() + interval '10 minutes', now(), now(), now())`,
		owner.User.ID)
	require.NoError(t, err)

	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/submit-verification", owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var body struct {
		Company struct {
			Status           string `json:"status"`
			OnboardingStatus string `json:"onboarding_status"`
		} `json:"company"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "under_review", body.Company.Status)
	require.Equal(t, "pending_review", body.Company.OnboardingStatus)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func doWithCookie(t *testing.T, router http.Handler, path, name, value string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: name, Value: value})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func cookieValue(t *testing.T, rec *httptest.ResponseRecorder, name string) string {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func systemRoleID(t *testing.T, env *testEnv, key string) uuid.UUID {
	t.Helper()
	role, err := db.New(env.pool).GetSystemRoleByKey(t.Context(), key)
	require.NoError(t, err)
	return role.ID
}

func addMembership(t *testing.T, env *testEnv, companyID, userID string, roleID uuid.UUID) uuid.UUID {
	t.Helper()
	var memberID uuid.UUID
	err := env.pool.QueryRow(t.Context(), `INSERT INTO company_members (company_id, user_id, status, joined_at)
		VALUES ($1, $2, 'active', now()) RETURNING id`, companyID, userID).Scan(&memberID)
	require.NoError(t, err)

	_, err = env.pool.Exec(t.Context(), `INSERT INTO member_roles (member_id, role_id) VALUES ($1, $2)`, memberID, roleID)
	require.NoError(t, err)
	return memberID
}

var _ = pgx.ErrNoRows
