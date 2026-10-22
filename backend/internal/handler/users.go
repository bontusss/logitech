package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/db/sqlc"
	"backend/internal/generator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserRequest defines the common body for creating owner, manager and
// staff accounts.
type CreateUserRequest struct {
	BranchID string `json:"branch_id" binding:"required" example:"11111111-1111-4111-8111-111111111111"`
	FullName string `json:"full_name" binding:"required" example:"Tunde Bello"`
	Phone    string `json:"phone,omitempty" example:"08031110002"`
	Email    string `json:"email,omitempty" example:"staff1@logitech.example"`
}

// CreateUserResponse represents the result of user creation. The one-time
// password is returned exactly once, at creation time, and must be handed to
// the user so they can log in and set their personal PIN.
type CreateUserResponse struct {
	ID              string `json:"id" example:"ffffffff-ffff-4fff-8fff-ffffffffffff"`
	BranchID        string `json:"branch_id" example:"11111111-1111-4111-8111-111111111111"`
	FullName        string `json:"full_name" example:"Tunde Bello"`
	Phone           string `json:"phone,omitempty" example:"08031110002"`
	Email           string `json:"email,omitempty" example:"staff1@logitech.example"`
	StaffID         string `json:"staff_id" example:"GAL-005"`
	OneTimePassword string `json:"one_time_password" example:"7Y3B9QK2"`
	Role            string `json:"role" example:"staff"`
	JobRole         string `json:"job_role,omitempty" example:"operations"`
	Status          string `json:"status" example:"active"`
	FirstLoginDone  bool   `json:"first_login_done" example:"false"`
	CreatedAt       string `json:"created_at"`
}

// parseJobRole validates a job role string for staff accounts.
func parseJobRole(raw string) (sqlc.NullStaffJobRole, *ErrorResponse) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "operations":
		return sqlc.NullStaffJobRole{StaffJobRole: sqlc.StaffJobRoleOperations, Valid: true}, nil
	case "dispatch":
		return sqlc.NullStaffJobRole{StaffJobRole: sqlc.StaffJobRoleDispatch, Valid: true}, nil
	case "receiving":
		return sqlc.NullStaffJobRole{StaffJobRole: sqlc.StaffJobRoleReceiving, Valid: true}, nil
	default:
		return sqlc.NullStaffJobRole{}, &ErrorResponse{
			Error: "job_role is required for staff and must be one of: operations, dispatch, receiving",
		}
	}
}

// createUser validates the request, generates the staff_id and one-time
// password, and inserts the user record. role must already be authorized and
// one of owner, manager, staff. For staff accounts jobRole must be valid.
func (h *StaffHandler) createUser(c *gin.Context, claims *auth.Claims, role sqlc.StaffRole, jobRole sqlc.NullStaffJobRole, req CreateUserRequest) (sqlc.Staff, string, *ErrorResponse) {
	// Verify branch exists and is active.
	branchUUID, err := uuid.Parse(req.BranchID)
	if err != nil {
		return sqlc.Staff{}, "", &ErrorResponse{Error: "invalid branch_id format (must be UUID)"}
	}

	branch, err := h.queries.GetBranchByID(c.Request.Context(), branchUUID)
	if err != nil {
		return sqlc.Staff{}, "", &ErrorResponse{Error: "branch not found with specified branch_id"}
	}
	if branch.Status != sqlc.RecordStatusActive {
		return sqlc.Staff{}, "", &ErrorResponse{Error: "branch is inactive"}
	}

	// Check email uniqueness if provided
	if strings.TrimSpace(req.Email) != "" {
		existing, err := h.queries.GetStaffByEmail(c.Request.Context(), sql.NullString{String: req.Email, Valid: true})
		if err == nil && existing.ID != uuid.Nil {
			return sqlc.Staff{}, "", &ErrorResponse{Error: "email already in use by another staff member"}
		}
	}

	// 1. Auto-generate sequential staff_id (e.g. GAL-001).
	staffID, err := generator.GenerateStaffID(c.Request.Context(), h.queries, h.cfg.StaffIDPrefix)
	if err != nil {
		return sqlc.Staff{}, "", &ErrorResponse{Error: "failed to generate unique staff_id: " + err.Error()}
	}

	// 2. Generate a one-time password (returned once in the response body).
	oneTimePassword, err := generator.GenerateOneTimePassword()
	if err != nil {
		return sqlc.Staff{}, "", &ErrorResponse{Error: "failed to generate one-time password: " + err.Error()}
	}

	// 3. Hash the one-time password; the user picks their own PIN on first login.
	otpHash, err := bcrypt.GenerateFromPassword([]byte(oneTimePassword), bcrypt.DefaultCost)
	if err != nil {
		return sqlc.Staff{}, "", &ErrorResponse{Error: "failed to hash credentials: " + err.Error()}
	}

	creatorUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		creatorUUID = uuid.Nil
	}

	createdStaff, err := h.queries.CreateStaff(c.Request.Context(), sqlc.CreateStaffParams{
		ID:        uuid.New(),
		BranchID:  branchUUID,
		FullName:  strings.TrimSpace(req.FullName),
		Phone:     sql.NullString{String: strings.TrimSpace(req.Phone), Valid: strings.TrimSpace(req.Phone) != ""},
		Email:     sql.NullString{String: strings.TrimSpace(req.Email), Valid: strings.TrimSpace(req.Email) != ""},
		StaffID:   staffID,
		OtpHash:   sql.NullString{String: string(otpHash), Valid: true},
		Role:      role,
		JobRole:   jobRole,
		Status:    sqlc.RecordStatusActive,
		CreatedBy: uuid.NullUUID{UUID: creatorUUID, Valid: creatorUUID != uuid.Nil},
	})
	if err != nil {
		return sqlc.Staff{}, "", &ErrorResponse{Error: "failed to save staff record: " + err.Error()}
	}

	return createdStaff, oneTimePassword, nil
}

// createUserResponse maps a created user row and its one-time password onto the response.
func createUserResponse(s sqlc.Staff, oneTimePassword string) CreateUserResponse {
	return CreateUserResponse{
		ID:              s.ID.String(),
		BranchID:        s.BranchID.String(),
		FullName:        s.FullName,
		Phone:           s.Phone.String,
		Email:           s.Email.String,
		StaffID:         s.StaffID,
		OneTimePassword: oneTimePassword,
		Role:            string(s.Role),
		JobRole:         jobRoleString(s.JobRole),
		Status:          string(s.Status),
		FirstLoginDone:  s.FirstLoginDone,
		CreatedAt:       s.CreatedAt.Format(time.RFC3339),
	}
}

// CreateOwner godoc
// @Summary Create an owner account
// @Description Creates an owner account. Only the system admin can create owners. A staff_id is auto-generated and a one-time password is returned once, to be given to the owner for their first login.
// @Tags owners
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "Owner Details"
// @Success 201 {object} CreateUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/owners [post]
func (h *StaffHandler) CreateOwner(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	created, otp, errMsg := h.createUser(c, claims, sqlc.StaffRoleOwner, sqlc.NullStaffJobRole{}, req)
	if errMsg != nil {
		c.JSON(errorStatus(errMsg.Error), *errMsg)
		return
	}

	c.JSON(http.StatusCreated, createUserResponse(created, otp))
}

// CreateManager godoc
// @Summary Create a manager account
// @Description Creates a manager account for a branch. Only the owner (or system admin) can manage managers. A staff_id is auto-generated and a one-time password is returned once, to be given to the manager for their first login.
// @Tags managers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "Manager Details"
// @Success 201 {object} CreateUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/managers [post]
func (h *StaffHandler) CreateManager(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	created, otp, errMsg := h.createUser(c, claims, sqlc.StaffRoleManager, sqlc.NullStaffJobRole{}, req)
	if errMsg != nil {
		c.JSON(errorStatus(errMsg.Error), *errMsg)
		return
	}

	c.JSON(http.StatusCreated, createUserResponse(created, otp))
}

// ListManagers godoc
// @Summary List manager accounts
// @Description Retrieve a list of manager accounts. Accessible by the owner (and system admin), optionally filtered by branch.
// @Tags managers
// @Produce json
// @Security BearerAuth
// @Param branch_id query string false "Filter by Branch UUID"
// @Success 200 {array} StaffProfile
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/managers [get]
func (h *StaffHandler) ListManagers(c *gin.Context) {
	ctx := c.Request.Context()

	var rows []sqlc.Staff
	var err error

	branchFilter := c.Query("branch_id")
	if branchFilter != "" {
		bUUID, parseErr := uuid.Parse(branchFilter)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid branch_id query parameter"})
			return
		}
		rows, err = h.queries.ListManagersByBranch(ctx, bUUID)
	} else {
		rows, err = h.queries.ListManagers(ctx)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve managers: " + err.Error()})
		return
	}

	results := make([]StaffProfile, 0, len(rows))
	for _, s := range rows {
		results = append(results, newStaffProfile(s))
	}

	c.JSON(http.StatusOK, results)
}

// errorStatus maps a handler error message onto an HTTP status code.
func errorStatus(msg string) int {
	switch {
	case strings.Contains(msg, "invalid branch_id"):
		return http.StatusBadRequest
	case strings.Contains(msg, "not found"), strings.Contains(msg, "inactive"), strings.Contains(msg, "already in use"):
		return http.StatusBadRequest
	case strings.Contains(msg, "failed to generate"), strings.Contains(msg, "failed to save"), strings.Contains(msg, "failed to hash"):
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
