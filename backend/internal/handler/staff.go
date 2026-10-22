package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/db/sqlc"
	"backend/internal/generator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// StaffHandler handles staff management endpoints.
type StaffHandler struct {
	cfg     *config.Config
	queries *sqlc.Queries
}

// NewStaffHandler constructs a StaffHandler.
func NewStaffHandler(cfg *config.Config, queries *sqlc.Queries) *StaffHandler {
	return &StaffHandler{
		cfg:     cfg,
		queries: queries,
	}
}

// CreateStaffRequest defines the body for creating a new user/staff member.
type CreateStaffRequest struct {
	BranchID string `json:"branch_id" binding:"required" example:"11111111-1111-4111-8111-111111111111"`
	FullName string `json:"full_name" binding:"required" example:"Tunde Bello"`
	Phone    string `json:"phone,omitempty" example:"08031110002"`
	Email    string `json:"email,omitempty" example:"staff1@logitech.example"`
	Role     string `json:"role" binding:"required" example:"staff"`
	// JobRole is required when role = "staff" and must be one of
	// operations, dispatch, receiving. Ignored for owner/manager.
	JobRole string `json:"job_role,omitempty" example:"operations"`
}

// CreateStaffResponse represents the result of staff creation. The one-time
// password is returned exactly once, at creation time, and must be handed to
// the user so they can log in and set their personal PIN.
type CreateStaffResponse struct {
	ID             string `json:"id" example:"ffffffff-ffff-4fff-8fff-ffffffffffff"`
	BranchID       string `json:"branch_id" example:"11111111-1111-4111-8111-111111111111"`
	FullName       string `json:"full_name" example:"Tunde Bello"`
	Phone          string `json:"phone,omitempty" example:"08031110002"`
	Email          string `json:"email,omitempty" example:"staff1@logitech.example"`
	StaffID        string `json:"staff_id" example:"GAL-005"`
	OneTimePassword string `json:"one_time_password" example:"7Y3B9QK2"`
	Role           string `json:"role" example:"staff"`
	JobRole        string `json:"job_role,omitempty" example:"operations"`
	Status         string `json:"status" example:"active"`
	FirstLoginDone bool   `json:"first_login_done" example:"false"`
	CreatedAt      string `json:"created_at"`
}

// CreateStaff godoc
// @Summary Create a staff member or a manager
// @Description Creates a new user. Creation rules:
// @Description - Managers can ONLY create staff (role: staff) in their own branch; a job_role (operations, dispatch or receiving) is required.
// @Description - The owner can ONLY create managers. The system admin can ONLY create owners.
// @Description A staff_id (e.g. GAL-001, prefix configurable via STAFF_ID_PREFIX) is auto-generated, and a one-time password is returned once so the user can log in and choose their personal PIN (max 10 chars).
// @Tags staff
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateStaffRequest true "Staff Creation Details"
// @Success 201 {object} CreateStaffResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/staff [post]
func (h *StaffHandler) CreateStaff(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	var req CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	targetRole := strings.ToLower(strings.TrimSpace(req.Role))

	// Enforce role-based creation rules:
	// 1. System admin can only create an owner.
	// 2. Owner can only create a manager.
	// 3. Manager can only create staff, and only in their own branch.
	switch claims.Role {
	case auth.RoleSystemAdmin:
		if targetRole != auth.RoleOwner {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "system admin can only create an owner; managers are created by the owner and staff by a manager",
			})
			return
		}

	case auth.RoleOwner:
		if targetRole != auth.RoleManager {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "only the owner can create managers; staff members are created by their branch manager",
			})
			return
		}

	case auth.RoleManager:
		if targetRole != auth.RoleStaff {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "managers can only create staff",
			})
			return
		}
		if claims.BranchID == "" {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "your account is not assigned to a branch; cannot create staff"})
			return
		}
		// Manager operations are limited to their assigned branch.
		req.BranchID = claims.BranchID

	default:
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: "permission denied: only managers can create staff, and only the owner can create managers",
		})
		return
	}

	// Staff must always carry a job role.
	var jobRole sqlc.NullStaffJobRole
	if targetRole == auth.RoleStaff {
		switch strings.ToLower(strings.TrimSpace(req.JobRole)) {
		case "operations":
			jobRole = sqlc.NullStaffJobRole{StaffJobRole: sqlc.StaffJobRoleOperations, Valid: true}
		case "dispatch":
			jobRole = sqlc.NullStaffJobRole{StaffJobRole: sqlc.StaffJobRoleDispatch, Valid: true}
		case "receiving":
			jobRole = sqlc.NullStaffJobRole{StaffJobRole: sqlc.StaffJobRoleReceiving, Valid: true}
		default:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "job_role is required for staff and must be one of: operations, dispatch, receiving",
			})
			return
		}
	}

	// Verify branch exists and is active.
	branchUUID, err := uuid.Parse(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid branch_id format (must be UUID)"})
		return
	}

	branch, err := h.queries.GetBranchByID(c.Request.Context(), branchUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "branch not found with specified branch_id"})
		return
	}
	if branch.Status != sqlc.RecordStatusActive {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "branch is inactive"})
		return
	}

	// Check email uniqueness if provided
	if strings.TrimSpace(req.Email) != "" {
		existing, err := h.queries.GetStaffByEmail(c.Request.Context(), sql.NullString{String: req.Email, Valid: true})
		if err == nil && existing.ID != uuid.Nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "email already in use by another staff member"})
			return
		}
	}

	// 1. Auto-generate sequential staff_id (e.g. GAL-001).
	staffID, err := generator.GenerateStaffID(c.Request.Context(), h.queries, h.cfg.StaffIDPrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate unique staff_id: " + err.Error()})
		return
	}

	// 2. Generate a one-time password (returned once in the response body).
	oneTimePassword, err := generator.GenerateOneTimePassword()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate one-time password: " + err.Error()})
		return
	}

	// 3. Hash the one-time password; the user picks their own PIN on first login.
	otpHash, err := bcrypt.GenerateFromPassword([]byte(oneTimePassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to hash credentials: " + err.Error()})
		return
	}

	creatorUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		creatorUUID = uuid.Nil
	}

	newID := uuid.New()
	createdStaff, err := h.queries.CreateStaff(c.Request.Context(), sqlc.CreateStaffParams{
		ID:        newID,
		BranchID:  branchUUID,
		FullName:  strings.TrimSpace(req.FullName),
		Phone:     sql.NullString{String: strings.TrimSpace(req.Phone), Valid: strings.TrimSpace(req.Phone) != ""},
		Email:     sql.NullString{String: strings.TrimSpace(req.Email), Valid: strings.TrimSpace(req.Email) != ""},
		StaffID:   staffID,
		OtpHash:   sql.NullString{String: string(otpHash), Valid: true},
		Role:      sqlc.StaffRole(targetRole),
		JobRole:   jobRole,
		Status:    sqlc.RecordStatusActive,
		CreatedBy: uuid.NullUUID{UUID: creatorUUID, Valid: creatorUUID != uuid.Nil},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save staff record: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CreateStaffResponse{
		ID:              createdStaff.ID.String(),
		BranchID:        createdStaff.BranchID.String(),
		FullName:        createdStaff.FullName,
		Phone:           createdStaff.Phone.String,
		Email:           createdStaff.Email.String,
		StaffID:         createdStaff.StaffID,
		OneTimePassword: oneTimePassword,
		Role:            string(createdStaff.Role),
		JobRole:         jobRoleString(createdStaff.JobRole),
		Status:          string(createdStaff.Status),
		FirstLoginDone:  createdStaff.FirstLoginDone,
		CreatedAt:       createdStaff.CreatedAt.Format(time.RFC3339),
	})
}

// ListStaff godoc
// @Summary List Staff Members
// @Description Retrieve a list of staff members. Accessible by system_admin, owner, and manager.
// @Description Managers are always limited to the staff of their own assigned branch.
// @Tags staff
// @Produce json
// @Security BearerAuth
// @Param branch_id query string false "Filter by Branch UUID"
// @Success 200 {array} StaffProfile
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/staff [get]
func (h *StaffHandler) ListStaff(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	branchFilter := c.Query("branch_id")

	// Manager operations are limited to their assigned branch: ignore any
	// branch_id filter pointing at another branch.
	if claims.Role == auth.RoleManager {
		if claims.BranchID == "" {
			c.JSON(http.StatusOK, []StaffProfile{})
			return
		}
		branchFilter = claims.BranchID
	}

	var rows []sqlc.Staff

	if branchFilter != "" {
		bUUID, err := uuid.Parse(branchFilter)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid branch_id query parameter"})
			return
		}
		rows, err = h.queries.ListStaffByBranch(c.Request.Context(), bUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve staff: " + err.Error()})
			return
		}
	} else {
		var err error
		rows, err = h.queries.ListStaff(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve staff: " + err.Error()})
			return
		}
	}

	results := make([]StaffProfile, 0, len(rows))
	for _, s := range rows {
		results = append(results, newStaffProfile(s))
	}

	c.JSON(http.StatusOK, results)
}

// getTargetStaff loads the staff record targeted by :id in the URL.
func (h *StaffHandler) getTargetStaff(c *gin.Context) (sqlc.Staff, bool) {
	idParam := c.Param("id")
	targetUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid staff id format"})
		return sqlc.Staff{}, false
	}

	staff, err := h.queries.GetStaffByID(c.Request.Context(), targetUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "staff record not found"})
		return sqlc.Staff{}, false
	}
	return staff, true
}

// canManageTarget enforces who may manage which staff records:
// - system_admin and owner may manage anyone.
// - managers may only manage staff in their own branch.
func canManageTarget(claims *auth.Claims, target sqlc.Staff) *ErrorResponse {
	switch claims.Role {
	case auth.RoleSystemAdmin, auth.RoleOwner:
		return nil
	case auth.RoleManager:
		if target.Role != sqlc.StaffRoleStaff {
			return &ErrorResponse{Error: "only the owner can manage manager or owner accounts"}
		}
		if claims.BranchID == "" || target.BranchID.String() != claims.BranchID {
			return &ErrorResponse{Error: "manager operations are limited to their assigned branch"}
		}
		return nil
	default:
		return &ErrorResponse{Error: "permission denied"}
	}
}

// GetStaff godoc
// @Summary Get Staff Details
// @Description Get details for a specific staff member by UUID. Staff can only fetch their own profile; managers only staff in their own branch.
// @Tags staff
// @Produce json
// @Security BearerAuth
// @Param id path string true "Staff UUID"
// @Success 200 {object} StaffProfile
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/staff/{id} [get]
func (h *StaffHandler) GetStaff(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	staff, ok := h.getTargetStaff(c)
	if !ok {
		return
	}

	// Staff can only view their own profile; managers only staff in their branch.
	if claims.Role == auth.RoleStaff && claims.UserID != staff.ID.String() {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "permission denied"})
		return
	}
	if claims.Role == auth.RoleManager && staff.BranchID.String() != claims.BranchID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "manager operations are limited to their assigned branch"})
		return
	}

	c.JSON(http.StatusOK, newStaffProfile(staff))
}

// UpdateStaffRequest defines the fields that may be updated on a staff/user record.
type UpdateStaffRequest struct {
	FullName string `json:"full_name,omitempty" example:"Tunde Bello"`
	Phone    string `json:"phone,omitempty" example:"08031110002"`
	Email    string `json:"email,omitempty" example:"staff1@logitech.example"`
	// JobRole may be changed only for staff members; manager/owner accounts have none.
	JobRole string `json:"job_role,omitempty" example:"operations"`
	// Status: "active" re-grants access (e.g. after a revoke); owner/system admin only.
	Status string `json:"status,omitempty" example:"active"`
}

// UpdateStaff godoc
// @Summary Update staff details
// @Description Update a staff member's details. Managers can only update staff in their own branch; the owner can update anyone. Changing status (e.g. re-activating a revoked account) is owner/system admin only.
// @Tags staff
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Staff UUID"
// @Param request body UpdateStaffRequest true "Fields to update"
// @Success 200 {object} StaffProfile
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/staff/{id} [put]
func (h *StaffHandler) UpdateStaff(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	target, ok := h.getTargetStaff(c)
	if !ok {
		return
	}

	if errMsg := canManageTarget(claims, target); errMsg != nil {
		c.JSON(http.StatusForbidden, *errMsg)
		return
	}

	var req UpdateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	// Check email uniqueness if provided and changed.
	newEmail := strings.TrimSpace(req.Email)
	if newEmail != "" && newEmail != target.Email.String {
		existing, err := h.queries.GetStaffByEmail(c.Request.Context(), sql.NullString{String: newEmail, Valid: true})
		if err == nil && existing.ID != target.ID {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "email already in use by another staff member"})
			return
		}
	}

	fullName := strings.TrimSpace(req.FullName)
	params := sqlc.UpdateStaffDetailsParams{
		ID: target.ID,
		FullName: sql.NullString{String: fullName, Valid: fullName != ""},
		Phone:   nullIfEmpty(req.Phone),
		Email:   nullIfEmpty(newEmail),
	}

	// job_role may only be set on staff records.
	if req.JobRole != "" {
		if target.Role != sqlc.StaffRoleStaff {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job_role can only be set on staff members"})
			return
		}
		switch strings.ToLower(strings.TrimSpace(req.JobRole)) {
		case "operations":
			params.JobRole = sqlc.NullStaffJobRole{StaffJobRole: sqlc.StaffJobRoleOperations, Valid: true}
		case "dispatch":
			params.JobRole = sqlc.NullStaffJobRole{StaffJobRole: sqlc.StaffJobRoleDispatch, Valid: true}
		case "receiving":
			params.JobRole = sqlc.NullStaffJobRole{StaffJobRole: sqlc.StaffJobRoleReceiving, Valid: true}
		default:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "job_role must be one of: operations, dispatch, receiving"})
			return
		}
	}

	updated, err := h.queries.UpdateStaffDetails(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update staff record: " + err.Error()})
		return
	}

	// Optional status change (re-activating a revoked account, de-activating one):
	// owner / system admin only.
	reqStatus := strings.ToLower(strings.TrimSpace(req.Status))
	if reqStatus != "" && reqStatus != string(target.Status) {
		if claims.Role != auth.RoleOwner && claims.Role != auth.RoleSystemAdmin {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "only the owner can change an account's status"})
			return
		}
		switch reqStatus {
		case "active", "inactive":
			updatedRow, err := h.queries.UpdateStaffStatus(c.Request.Context(), sqlc.UpdateStaffStatusParams{
				ID:     target.ID,
				Status: sqlc.RecordStatus(reqStatus),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update staff status: " + err.Error()})
				return
			}
			updated, err = h.queries.GetStaffByID(c.Request.Context(), updatedRow.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load staff record: " + err.Error()})
				return
			}
		default:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status must be active or inactive"})
			return
		}
	}

	c.JSON(http.StatusOK, newStaffProfile(updated))
}

// RevokeStaff godoc
// @Summary Revoke staff access
// @Description Revokes a user's access by marking their account inactive. Managers can revoke staff in their own branch; the owner can revoke anyone.
// @Tags staff
// @Produce json
// @Security BearerAuth
// @Param id path string true "Staff UUID"
// @Success 200 {object} StaffProfile
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/staff/{id} [delete]
func (h *StaffHandler) RevokeStaff(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	target, ok := h.getTargetStaff(c)
	if !ok {
		return
	}

	if errMsg := canManageTarget(claims, target); errMsg != nil {
		c.JSON(http.StatusForbidden, *errMsg)
		return
	}

	if target.Status != sqlc.RecordStatusActive {
		c.JSON(http.StatusOK, newStaffProfile(target))
		return
	}

	updated, err := h.queries.UpdateStaffStatus(c.Request.Context(), sqlc.UpdateStaffStatusParams{
		ID:     target.ID,
		Status: sqlc.RecordStatusInactive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to revoke access: " + err.Error()})
		return
	}

	staff, err := h.queries.GetStaffByID(c.Request.Context(), updated.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to load staff record: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, newStaffProfile(staff))
}

// ResetStaffPinResponse is returned after a manager or owner resets a PIN.
type ResetStaffPinResponse struct {
	StaffID         string `json:"staff_id" example:"GAL-003"`
	OneTimePassword string `json:"one_time_password" example:"9XKM28LP"`
}

// ResetStaffPin godoc
// @Summary Reset a staff PIN
// @Description Generates a new one-time password for the user, invalidating their current PIN. The one-time password is returned once and must be given to the user, who uses it at login (with new_pin) to choose a new PIN. Any lockout on the account is cleared.
// @Description Managers can reset the PIN of staff in their own branch; the owner can reset any PIN.
// @Tags staff
// @Produce json
// @Security BearerAuth
// @Param id path string true "Staff UUID"
// @Success 200 {object} ResetStaffPinResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/staff/{id}/reset-pin [post]
func (h *StaffHandler) ResetStaffPin(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	target, ok := h.getTargetStaff(c)
	if !ok {
		return
	}

	if errMsg := canManageTarget(claims, target); errMsg != nil {
		c.JSON(http.StatusForbidden, *errMsg)
		return
	}

	oneTimePassword, err := generator.GenerateOneTimePassword()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate one-time password: " + err.Error()})
		return
	}

	otpHash, err := bcrypt.GenerateFromPassword([]byte(oneTimePassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to hash one-time password: " + err.Error()})
		return
	}

	if _, err := h.queries.StartPinReset(c.Request.Context(), sqlc.StartPinResetParams{
		ID:      target.ID,
		OtpHash: sql.NullString{String: string(otpHash), Valid: true},
	}); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to reset pin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, ResetStaffPinResponse{
		StaffID:         target.StaffID,
		OneTimePassword: oneTimePassword,
	})
}

// UnlockStaffAccount godoc
// @Summary Unlock a locked account
// @Description Removes the lockout from an account that was locked after 3 failed login attempts. Managers can unlock staff of their own branch; the owner can unlock accounts across all branches.
// @Tags staff
// @Produce json
// @Security BearerAuth
// @Param id path string true "Staff UUID"
// @Success 200 {object} StaffProfile
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/staff/{id}/unlock [post]
func (h *StaffHandler) UnlockStaffAccount(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	target, ok := h.getTargetStaff(c)
	if !ok {
		return
	}

	if errMsg := canManageTarget(claims, target); errMsg != nil {
		c.JSON(http.StatusForbidden, *errMsg)
		return
	}

	if !target.LockedAt.Valid {
		c.JSON(http.StatusOK, newStaffProfile(target))
		return
	}

	updated, err := h.queries.UnlockStaffAccount(c.Request.Context(), target.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to unlock account: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, newStaffProfile(updated))
}

// FirstLoginStatusResponse reports whether a user has completed their first login.
type FirstLoginStatusResponse struct {
	StaffID        string `json:"staff_id" example:"GAL-003"`
	FirstLoginDone bool   `json:"first_login_done" example:"true"`
	// FirstLoginAt is set once the user has completed first login (RFC3339).
	FirstLoginAt string `json:"first_login_at,omitempty" example:"2026-10-22T09:15:00Z"`
}

// FirstLoginStatus godoc
// @Summary Check if a staff has performed their first login
// @Description Reports whether the staff has completed their first login (i.e. created their personal PIN), and when. Staff can query their own status; managers only staff in their own branch; the owner can query anyone.
// @Tags staff
// @Produce json
// @Security BearerAuth
// @Param id path string true "Staff UUID"
// @Success 200 {object} FirstLoginStatusResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/staff/{id}/first-login [get]
func (h *StaffHandler) FirstLoginStatus(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	target, ok := h.getTargetStaff(c)
	if !ok {
		return
	}

	// Staff may check themselves; managers only staff in their branch.
	if claims.Role == auth.RoleStaff && claims.UserID != target.ID.String() {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "permission denied"})
		return
	}
	if claims.Role == auth.RoleManager && target.BranchID.String() != claims.BranchID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "manager operations are limited to their assigned branch"})
		return
	}

	resp := FirstLoginStatusResponse{
		StaffID:        target.StaffID,
		FirstLoginDone: target.FirstLoginDone,
	}
	if target.FirstLoginAt.Valid {
		resp.FirstLoginAt = target.FirstLoginAt.Time.Format(time.RFC3339)
	}


	c.JSON(http.StatusOK, resp)
}

// nullIfEmpty builds a nullable string column value.
func nullIfEmpty(s string) sql.NullString {
	s = strings.TrimSpace(s)
	return sql.NullString{String: s, Valid: s != ""}
}
