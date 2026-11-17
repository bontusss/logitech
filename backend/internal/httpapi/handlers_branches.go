package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

// branchDTO normalises citext email and jsonb operating hours before they are
// serialised, instead of exposing database-specific Go types to the client.
type branchDTO struct {
	ID              uuid.UUID       `json:"id"`
	CompanyID       uuid.UUID       `json:"company_id"`
	Name            string          `json:"name"`
	Code            string          `json:"code"`
	Address         string          `json:"address"`
	City            string          `json:"city"`
	State           string          `json:"state"`
	Country         string          `json:"country"`
	Phone           *string         `json:"phone"`
	Email           *string         `json:"email"`
	ManagerMemberID *uuid.UUID      `json:"manager_member_id"`
	OperatingHours  json.RawMessage `json:"operating_hours"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       *time.Time      `json:"deleted_at"`
}

func toBranchDTO(branch db.Branch) branchDTO {
	var email *string
	if branch.Email.Valid {
		value := branch.Email.String
		email = &value
	}
	return branchDTO{
		ID: branch.ID, CompanyID: branch.CompanyID, Name: branch.Name, Code: branch.Code,
		Address: branch.Address, City: branch.City, State: branch.State, Country: branch.Country,
		Phone: branch.Phone, Email: email, ManagerMemberID: branch.ManagerMemberID,
		OperatingHours: json.RawMessage(branch.OperatingHours), Status: branch.Status,
		CreatedAt: branch.CreatedAt, UpdatedAt: branch.UpdatedAt, DeletedAt: branch.DeletedAt,
	}
}

type branchInput struct {
	Name            string          `json:"name" binding:"required,min=2,max=120"`
	Code            string          `json:"code" binding:"required,min=2,max=24"`
	Address         string          `json:"address" binding:"required,max=240"`
	City            string          `json:"city" binding:"required,max=80"`
	State           string          `json:"state" binding:"required,max=80"`
	Country         string          `json:"country" binding:"omitempty,max=2"`
	Phone           string          `json:"phone" binding:"omitempty,max=30"`
	Email           string          `json:"email" binding:"omitempty,email,max=254"`
	ManagerMemberID *string         `json:"manager_member_id"`
	OperatingHours  json.RawMessage `json:"operating_hours"`
}

// hasCompanyWideBranchAccess is true for platform administrators and company
// owners/administrators. Everyone else sees only explicitly assigned branches.
func (a *API) hasCompanyWideBranchAccess(c *gin.Context) bool {
	tenant, ok := currentTenant(c)
	if !ok {
		return false
	}
	if tenant.IsPlatform {
		return true
	}
	roles, _ := a.svc.Q.ListMemberRoles(c.Request.Context(), tenant.MemberID)
	return slices.Contains(roles, "owner") || slices.Contains(roles, "company_admin")
}

func (a *API) accessibleBranch(c *gin.Context) (db.Branch, tenantContext, bool) {
	tenant, _ := currentTenant(c)
	branchID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_branch", "That branch id is invalid.")
		return db.Branch{}, tenant, false
	}

	branch, err := a.svc.Q.GetBranch(c.Request.Context(), db.GetBranchParams{
		ID:        branchID,
		CompanyID: tenant.CompanyID,
	})
	if err != nil {
		fail(c, http.StatusNotFound, "branch_not_found", "That branch was not found.")
		return db.Branch{}, tenant, false
	}

	if !a.hasCompanyWideBranchAccess(c) {
		allowed, accessErr := a.svc.Q.MemberHasBranchAccess(c.Request.Context(), db.MemberHasBranchAccessParams{
			BranchID:  branch.ID,
			MemberID:  tenant.MemberID,
			CompanyID: tenant.CompanyID,
		})
		if accessErr != nil || !allowed {
			fail(c, http.StatusForbidden, "branch_forbidden", "You are not assigned to this branch.")
			return db.Branch{}, tenant, false
		}
	}
	return branch, tenant, true
}

func normalizeHours(raw json.RawMessage) ([]byte, bool) {
	if len(raw) == 0 {
		return []byte("{}"), true
	}
	if len(raw) > 8192 {
		return nil, false
	}
	var probe map[string]any
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, false
	}
	return raw, true
}

func (a *API) resolveManager(c *gin.Context, companyID uuid.UUID, raw *string) (*uuid.UUID, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, true
	}
	memberID, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil {
		fail(c, http.StatusUnprocessableEntity, "invalid_manager", "Select a valid branch manager.")
		return nil, false
	}
	member, err := a.svc.Q.GetCompanyMemberByID(c.Request.Context(), db.GetCompanyMemberByIDParams{
		CompanyID: companyID,
		ID:        memberID,
	})
	if err != nil || member.Status != "active" {
		fail(c, http.StatusUnprocessableEntity, "invalid_manager", "The branch manager must be an active company member.")
		return nil, false
	}
	return &memberID, true
}

func (a *API) listBranches(c *gin.Context) {
	tenant, _ := currentTenant(c)
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	if status != "" && status != "active" && status != "inactive" {
		fail(c, http.StatusBadRequest, "invalid_status", "Status must be active or inactive.")
		return
	}

	var (
		branches []db.Branch
		err      error
	)
	if a.hasCompanyWideBranchAccess(c) {
		branches, err = a.svc.Q.ListBranches(c.Request.Context(), db.ListBranchesParams{
			CompanyID: tenant.CompanyID,
			Column2:   search,
			Column3:   status,
		})
	} else {
		branches, err = a.svc.Q.ListBranchesForMember(c.Request.Context(), db.ListBranchesForMemberParams{
			CompanyID: tenant.CompanyID,
			MemberID:  tenant.MemberID,
			Column3:   search,
			Column4:   status,
		})
	}
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]branchDTO, 0, len(branches))
	for _, branch := range branches {
		result = append(result, toBranchDTO(branch))
	}
	c.JSON(http.StatusOK, gin.H{"branches": result})
}

func (a *API) createBranch(c *gin.Context) {
	tenant, _ := currentTenant(c)
	var req branchInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the branch name, code and address.")
		return
	}

	hours, ok := normalizeHours(req.OperatingHours)
	if !ok {
		fail(c, http.StatusUnprocessableEntity, "invalid_hours", "Operating hours must be a JSON object.")
		return
	}
	manager, ok := a.resolveManager(c, tenant.CompanyID, req.ManagerMemberID)
	if !ok {
		return
	}

	branch, err := a.svc.Q.CreateBranch(c.Request.Context(), db.CreateBranchParams{
		CompanyID:       tenant.CompanyID,
		Name:            strings.TrimSpace(req.Name),
		Code:            strings.ToUpper(strings.TrimSpace(req.Code)),
		Address:         strings.TrimSpace(req.Address),
		City:            strings.TrimSpace(req.City),
		State:           strings.TrimSpace(req.State),
		Country:         defaultCountry(req.Country),
		Column8:         strings.TrimSpace(req.Phone),
		Column9:         strings.ToLower(strings.TrimSpace(req.Email)),
		ManagerMemberID: manager,
		OperatingHours:  hours,
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "branch_code_taken", "That branch code is already used in this company.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	if manager != nil {
		_ = a.svc.Q.AddBranchMember(c.Request.Context(), db.AddBranchMemberParams{
			BranchID: branch.ID,
			MemberID: *manager,
		})
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "branch.created",
		EntityType: "branch",
		EntityID:   &branch.ID,
		Metadata:   map[string]any{"code": branch.Code},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusCreated, gin.H{"branch": toBranchDTO(branch)})
}

func (a *API) getBranch(c *gin.Context) {
	branch, tenant, ok := a.accessibleBranch(c)
	if !ok {
		return
	}
	staff, err := a.svc.Q.ListBranchStaff(c.Request.Context(), db.ListBranchStaffParams{
		BranchID:  branch.ID,
		CompanyID: tenant.CompanyID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"branch": toBranchDTO(branch), "staff": staff})
}

func (a *API) updateBranch(c *gin.Context) {
	current, tenant, ok := a.accessibleBranch(c)
	if !ok {
		return
	}

	var req branchInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the branch details you entered.")
		return
	}

	if !a.hasCompanyWideBranchAccess(c) {
		currentManager := ""
		if current.ManagerMemberID != nil {
			currentManager = current.ManagerMemberID.String()
		}
		requestedManager := ""
		if req.ManagerMemberID != nil {
			requestedManager = strings.TrimSpace(*req.ManagerMemberID)
		}
		if requestedManager != currentManager {
			fail(c, http.StatusForbidden, "manager_forbidden", "Only a company administrator can change the branch manager.")
			return
		}
	}

	hours, valid := normalizeHours(req.OperatingHours)
	if !valid {
		fail(c, http.StatusUnprocessableEntity, "invalid_hours", "Operating hours must be a JSON object.")
		return
	}
	manager, valid := a.resolveManager(c, tenant.CompanyID, req.ManagerMemberID)
	if !valid {
		return
	}

	branch, err := a.svc.Q.UpdateBranch(c.Request.Context(), db.UpdateBranchParams{
		ID:              current.ID,
		CompanyID:       tenant.CompanyID,
		Name:            strings.TrimSpace(req.Name),
		Code:            strings.ToUpper(strings.TrimSpace(req.Code)),
		Address:         strings.TrimSpace(req.Address),
		City:            strings.TrimSpace(req.City),
		State:           strings.TrimSpace(req.State),
		Country:         defaultCountry(req.Country),
		Column9:         strings.TrimSpace(req.Phone),
		Column10:        strings.ToLower(strings.TrimSpace(req.Email)),
		ManagerMemberID: manager,
		OperatingHours:  hours,
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "branch_code_taken", "That branch code is already used in this company.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	if manager != nil {
		_ = a.svc.Q.AddBranchMember(c.Request.Context(), db.AddBranchMemberParams{
			BranchID: branch.ID,
			MemberID: *manager,
		})
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "branch.updated",
		EntityType: "branch",
		EntityID:   &branch.ID,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"branch": toBranchDTO(branch)})
}

func (a *API) setBranchStatus(c *gin.Context) {
	branch, tenant, ok := a.accessibleBranch(c)
	if !ok {
		return
	}

	status := "inactive"
	if strings.HasSuffix(c.Request.URL.Path, "/activate") {
		status = "active"
	}

	updated, err := a.svc.Q.SetBranchStatus(c.Request.Context(), db.SetBranchStatusParams{
		ID:        branch.ID,
		CompanyID: tenant.CompanyID,
		Status:    status,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	a.svc.Audit(c.Request.Context(), service.AuditEntry{
		ActorID:    actorID(c),
		CompanyID:  &tenant.CompanyID,
		Action:     "branch.status_changed",
		EntityType: "branch",
		EntityID:   &branch.ID,
		Metadata:   map[string]any{"status": status},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"branch": toBranchDTO(updated)})
}

func (a *API) listBranchStaff(c *gin.Context) {
	branch, tenant, ok := a.accessibleBranch(c)
	if !ok {
		return
	}
	staff, err := a.svc.Q.ListBranchStaff(c.Request.Context(), db.ListBranchStaffParams{
		BranchID:  branch.ID,
		CompanyID: tenant.CompanyID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"staff": staff})
}

func (a *API) setBranchStaff(c *gin.Context) {
	branch, tenant, ok := a.accessibleBranch(c)
	if !ok {
		return
	}

	var req struct {
		MemberIDs []string `json:"member_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Select the staff members to assign.")
		return
	}

	memberIDs := make([]uuid.UUID, 0, len(req.MemberIDs))
	for _, raw := range req.MemberIDs {
		memberID, err := uuid.Parse(raw)
		if err != nil {
			fail(c, http.StatusUnprocessableEntity, "invalid_member", "One of the selected members is invalid.")
			return
		}
		member, err := a.svc.Q.GetCompanyMemberByID(c.Request.Context(), db.GetCompanyMemberByIDParams{
			CompanyID: tenant.CompanyID,
			ID:        memberID,
		})
		if err != nil || member.Status != "active" {
			fail(c, http.StatusUnprocessableEntity, "invalid_member", "Only active company members can be assigned to a branch.")
			return
		}
		memberIDs = append(memberIDs, memberID)
	}
	if branch.ManagerMemberID != nil && !slices.Contains(memberIDs, *branch.ManagerMemberID) {
		memberIDs = append(memberIDs, *branch.ManagerMemberID)
	}

	err := pgx.BeginFunc(c.Request.Context(), a.svc.DB, func(tx pgx.Tx) error {
		q := a.svc.Q.WithTx(tx)
		if err := q.DeleteBranchMembers(c.Request.Context(), branch.ID); err != nil {
			return err
		}
		for _, memberID := range memberIDs {
			if err := q.AddBranchMember(c.Request.Context(), db.AddBranchMemberParams{
				BranchID: branch.ID,
				MemberID: memberID,
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
		Action:     "branch.staff_assigned",
		EntityType: "branch",
		EntityID:   &branch.ID,
		Metadata:   map[string]any{"members": len(memberIDs)},
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	c.JSON(http.StatusOK, gin.H{"assigned": len(memberIDs)})
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func defaultCountry(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "NG"
	}
	return value
}
