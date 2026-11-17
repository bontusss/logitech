package httpapi

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bol-platform/backend/internal/auth"
	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

func (a *API) optionalUser(c *gin.Context) *db.User {
	raw := accessTokenFromRequest(c)
	if raw == "" {
		return nil
	}
	claims, err := a.auth.Parse(raw)
	if err != nil || a.svc.IsAccessTokenRevoked(c.Request.Context(), claims.ID) {
		return nil
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil
	}
	user, err := a.svc.Q.GetUserByID(c.Request.Context(), userID)
	if err != nil || user.Status != "active" {
		return nil
	}
	return &user
}

func (a *API) branchIDsForCompany(c *gin.Context, companyID uuid.UUID, raw []string) ([]uuid.UUID, bool) {
	ids := make([]uuid.UUID, 0, len(raw))
	seen := map[uuid.UUID]bool{}
	for _, value := range raw {
		branchID, err := uuid.Parse(strings.TrimSpace(value))
		if err != nil {
			fail(c, http.StatusUnprocessableEntity, "invalid_branch", "One of the selected branches is invalid.")
			return nil, false
		}
		if seen[branchID] {
			continue
		}
		if _, err := a.svc.Q.GetBranch(c.Request.Context(), db.GetBranchParams{
			ID:        branchID,
			CompanyID: companyID,
		}); err != nil {
			fail(c, http.StatusUnprocessableEntity, "invalid_branch", "One of the selected branches does not belong to this company.")
			return nil, false
		}
		seen[branchID] = true
		ids = append(ids, branchID)
	}
	return ids, true
}

func (a *API) inviteStaff(c *gin.Context) {
	tenant, _ := currentTenant(c)
	var req struct {
		Email     string   `json:"email" binding:"required,email,max=254"`
		RoleKey   string   `json:"role_key" binding:"required"`
		BranchIDs []string `json:"branch_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Enter a valid email and role.")
		return
	}
	role, ok := a.resolveAssignableRole(c, req.RoleKey)
	if !ok {
		return
	}

	branchIDs, ok := a.branchIDsForCompany(c, tenant.CompanyID, req.BranchIDs)
	if !ok {
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if target, lookupErr := a.svc.Q.GetUserByEmail(c.Request.Context(), email); lookupErr == nil {
		if _, memberErr := a.svc.Q.GetActiveCompanyMember(c.Request.Context(), db.GetActiveCompanyMemberParams{
			CompanyID: tenant.CompanyID,
			UserID:    target.ID,
		}); memberErr == nil {
			fail(c, http.StatusConflict, "already_member", "That person is already a member of this company.")
			return
		}
	}

	company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	raw, hash, err := auth.GenerateOpaqueToken()
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	var invitation db.CompanyInvitation
	err = pgx.BeginFunc(c.Request.Context(), a.svc.DB, func(tx pgx.Tx) error {
		q := a.svc.Q.WithTx(tx)
		if err := q.RevokePendingInvitations(c.Request.Context(), db.RevokePendingInvitationsParams{
			CompanyID: tenant.CompanyID,
			Email:     email,
		}); err != nil {
			return err
		}
		created, err := q.CreateInvitation(c.Request.Context(), db.CreateInvitationParams{
			CompanyID: tenant.CompanyID,
			Email:     email,
			RoleID:    role.ID,
			TokenHash: hash,
			InvitedBy: actorID(c),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		})
		if err != nil {
			return err
		}
		invitation = created
		for _, branchID := range branchIDs {
			if err := q.AddInvitationBranch(c.Request.Context(), db.AddInvitationBranchParams{
				InvitationID: created.ID,
				BranchID:     branchID,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	if err := a.svc.Mailer.SendInvitation(c.Request.Context(), email, company.Name, raw); err != nil {
		a.log.Warn("could not send invitation email", "error", err, "company_id", tenant.CompanyID)
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.member_invited",
		EntityType: "company_invitation",
		EntityID:   &invitation.ID,
		Metadata:   map[string]any{"email": email, "role": role.Key, "branches": len(branchIDs)},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	response := gin.H{
		"invitation": gin.H{
			"id": invitation.ID, "email": invitation.Email, "role_key": role.Key,
			"branch_ids": branchIDs, "expires_at": invitation.ExpiresAt,
		},
	}
	if a.cfg.ExposeDevKeys {
		response["dev_invite_token"] = raw
	}
	c.JSON(http.StatusCreated, response)
}

func (a *API) listInvitations(c *gin.Context) {
	tenant, _ := currentTenant(c)
	invitations, err := a.svc.Q.ListCompanyInvitations(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"invitations": invitations})
}

func (a *API) revokeInvitation(c *gin.Context) {
	tenant, _ := currentTenant(c)
	invitationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_invitation", "That invitation id is invalid.")
		return
	}

	invitation, err := a.svc.Q.GetCompanyInvitation(c.Request.Context(), db.GetCompanyInvitationParams{
		ID:        invitationID,
		CompanyID: tenant.CompanyID,
	})
	if err != nil {
		fail(c, http.StatusNotFound, "invitation_not_found", "That invitation was not found.")
		return
	}
	if invitation.AcceptedAt != nil || invitation.RevokedAt != nil {
		fail(c, http.StatusConflict, "invitation_closed", "This invitation can no longer be revoked.")
		return
	}

	if err := a.svc.Q.RevokeInvitation(c.Request.Context(), db.RevokeInvitationParams{
		ID:        invitation.ID,
		CompanyID: tenant.CompanyID,
	}); err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.invitation_revoked",
		EntityType: "company_invitation",
		EntityID:   &invitation.ID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.Status(http.StatusNoContent)
}

func (a *API) resendInvitation(c *gin.Context) {
	tenant, _ := currentTenant(c)
	invitationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_invitation", "That invitation id is invalid.")
		return
	}

	invitation, err := a.svc.Q.GetCompanyInvitation(c.Request.Context(), db.GetCompanyInvitationParams{
		ID:        invitationID,
		CompanyID: tenant.CompanyID,
	})
	if err != nil {
		fail(c, http.StatusNotFound, "invitation_not_found", "That invitation was not found.")
		return
	}
	if invitation.AcceptedAt != nil || invitation.RevokedAt != nil {
		fail(c, http.StatusConflict, "invitation_closed", "This invitation has already been used or revoked.")
		return
	}

	raw, hash, err := auth.GenerateOpaqueToken()
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	updated, err := a.svc.Q.UpdateInvitationToken(c.Request.Context(), db.UpdateInvitationTokenParams{
		ID:        invitation.ID,
		CompanyID: tenant.CompanyID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	company, _ := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err := a.svc.Mailer.SendInvitation(c.Request.Context(), invitation.Email, company.Name, raw); err != nil {
		a.log.Warn("could not resend invitation email", "error", err, "company_id", tenant.CompanyID)
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.invitation_resent",
		EntityType: "company_invitation",
		EntityID:   &invitation.ID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})

	response := gin.H{"invitation": gin.H{"id": updated.ID, "expires_at": updated.ExpiresAt}}
	if a.cfg.ExposeDevKeys {
		response["dev_invite_token"] = raw
	}
	c.JSON(http.StatusOK, response)
}

func (a *API) previewInvitation(c *gin.Context) {
	preview, err := a.svc.InvitationPreview(c.Request.Context(), c.Param("token"))
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	branches := make([]gin.H, 0, len(preview.BranchIDs))
	for _, branchID := range preview.BranchIDs {
		branch, branchErr := a.svc.Q.GetBranch(c.Request.Context(), db.GetBranchParams{
			ID:        branchID,
			CompanyID: preview.CompanyID,
		})
		if branchErr == nil {
			branches = append(branches, gin.H{"id": branch.ID, "name": branch.Name, "code": branch.Code})
		}
	}

	actor := a.optionalUser(c)
	requiresSignIn := preview.AccountExists &&
		(actor == nil || !strings.EqualFold(actor.Email, preview.Email))

	c.JSON(http.StatusOK, gin.H{
		"invitation": gin.H{
			"email":            preview.Email,
			"expires_at":       preview.ExpiresAt,
			"role_key":         preview.RoleKey,
			"role_name":        preview.RoleName,
			"branch_ids":       preview.BranchIDs,
			"branches":         branches,
			"account_exists":   preview.AccountExists,
			"requires_sign_in": requiresSignIn,
		},
		"company": gin.H{
			"id": preview.CompanyID, "name": preview.CompanyName, "slug": preview.CompanySlug,
		},
	})
}

func (a *API) acceptInvitation(c *gin.Context) {
	rawToken := c.Param("token")
	var req struct {
		FullName string `json:"full_name" binding:"omitempty,min=2,max=120"`
		Password string `json:"password" binding:"omitempty,min=8,max=128"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the name and password you entered.")
		return
	}

	actor := a.optionalUser(c)
	preview, err := a.svc.InvitationPreview(c.Request.Context(), rawToken)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	if preview.AccountExists && actor == nil {
		fail(c, http.StatusConflict, "sign_in_required", "Sign in to the invited account to accept this invitation.")
		return
	}
	if actor == nil && (strings.TrimSpace(req.FullName) == "" || len(req.Password) < 8) {
		fail(c, http.StatusUnprocessableEntity, "profile_required", "Enter your full name and a password of at least 8 characters.")
		return
	}

	session, err := a.svc.AcceptInvitation(
		c.Request.Context(), rawToken, actor, req.FullName, req.Password,
		c.ClientIP(), c.Request.UserAgent(),
	)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.setSessionCookies(c, session)
	c.JSON(http.StatusCreated, toSessionDTO(session))
}

func (a *API) updateMember(c *gin.Context) {
	tenant, _ := currentTenant(c)
	memberID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_member", "That member id is invalid.")
		return
	}

	member, err := a.svc.Q.GetCompanyMemberByID(c.Request.Context(), db.GetCompanyMemberByIDParams{
		CompanyID: tenant.CompanyID,
		ID:        memberID,
	})
	if err != nil {
		fail(c, http.StatusNotFound, "member_not_found", "That member was not found.")
		return
	}

	roles, _ := a.svc.Q.ListMemberRoles(c.Request.Context(), member.ID)
	if slices.Contains(roles, "owner") {
		fail(c, http.StatusConflict, "owner_immutable", "The company owner cannot be suspended or edited here.")
		return
	}

	var req struct {
		Title  *string `json:"title" binding:"omitempty,max=80"`
		Status *string `json:"status" binding:"omitempty,oneof=active suspended"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Title or status is invalid.")
		return
	}

	status := member.Status
	if req.Status != nil {
		status = *req.Status
	}
	if status != "active" && status != "suspended" {
		fail(c, http.StatusUnprocessableEntity, "invalid_status", "Status must be active or suspended.")
		return
	}
	if member.UserID == mustUserID(c) && status != "active" {
		fail(c, http.StatusConflict, "self_suspend", "You cannot suspend your own membership.")
		return
	}

	updated, err := a.svc.Q.UpdateMemberProfile(c.Request.Context(), db.UpdateMemberProfileParams{
		ID:        member.ID,
		CompanyID: tenant.CompanyID,
		Column3:   valueOr(req.Title, deref(member.Title)),
		Status:    status,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	if status == "suspended" {
		_ = a.svc.Q.RevokeUserRefreshTokens(c.Request.Context(), member.UserID)
		_ = a.svc.Q.DeleteMemberBranches(c.Request.Context(), member.ID)
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.member_updated",
		EntityType: "company_member",
		EntityID:   &member.ID,
		Metadata:   map[string]any{"status": status},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"member": updated})
}

func (a *API) setMemberBranches(c *gin.Context) {
	tenant, _ := currentTenant(c)
	memberID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_member", "That member id is invalid.")
		return
	}

	member, err := a.svc.Q.GetCompanyMemberByID(c.Request.Context(), db.GetCompanyMemberByIDParams{
		CompanyID: tenant.CompanyID,
		ID:        memberID,
	})
	if err != nil {
		fail(c, http.StatusNotFound, "member_not_found", "That member was not found.")
		return
	}

	var req struct {
		BranchIDs []string `json:"branch_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Select the branches for this member.")
		return
	}
	branchIDs, ok := a.branchIDsForCompany(c, tenant.CompanyID, req.BranchIDs)
	if !ok {
		return
	}

	err = pgx.BeginFunc(c.Request.Context(), a.svc.DB, func(tx pgx.Tx) error {
		q := a.svc.Q.WithTx(tx)
		if err := q.DeleteMemberBranches(c.Request.Context(), member.ID); err != nil {
			return err
		}
		for _, branchID := range branchIDs {
			if err := q.AddMemberBranch(c.Request.Context(), db.AddMemberBranchParams{
				BranchID: branchID,
				MemberID: member.ID,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "company.member_branches_updated",
		EntityType: "company_member",
		EntityID:   &member.ID,
		Metadata:   map[string]any{"branches": len(branchIDs)},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"member_id": member.ID, "branch_ids": branchIDs})
}

var _ = errors.Is
