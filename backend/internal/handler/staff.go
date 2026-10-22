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

// StaffHandler handles staff management endpoints.
type StaffHandler struct {
	queries *sqlc.Queries
}

// NewStaffHandler constructs a StaffHandler.
func NewStaffHandler(queries *sqlc.Queries) *StaffHandler {
	return &StaffHandler{
		queries: queries,
	}
}

// CreateStaffRequest defines the body for creating a new user/staff member.
type CreateStaffRequest struct {
	BranchID string `json:"branch_id" binding:"required" example:"11111111-1111-4111-8111-111111111111"`
	FullName string `json:"full_name" binding:"required" example:"Ada Nwosu"`
	Phone    string `json:"phone,omitempty" example:"08031110002"`
	Email    string `json:"email,omitempty" example:"ada@logitech.example"`
	Role     string `json:"role,omitempty" example:"staff"`
}

// CreateStaffResponse represents the result of staff creation, including auto-generated credentials.
type CreateStaffResponse struct {
	ID        string `json:"id" example:"ffffffff-ffff-4fff-8fff-ffffffffffff"`
	BranchID  string `json:"branch_id" example:"11111111-1111-4111-8111-111111111111"`
	FullName  string `json:"full_name" example:"Ada Nwosu"`
	Phone     string `json:"phone,omitempty" example:"08031110002"`
	Email     string `json:"email,omitempty" example:"ada@logitech.example"`
	StaffID   string `json:"staff_id" example:"STF-005"`
	Pin       string `json:"pin" example:"9A4F"`
	Role      string `json:"role" example:"staff"`
	Status    string `json:"status" example:"active"`
	CreatedAt string `json:"created_at"`
}

// CreateStaff godoc
// @Summary Create a staff, manager, or owner
// @Description Creates a new staff member. Role creation rules:
// @Description - Managers can ONLY create staff (role: staff).
// @Description - System Admin can ONLY create owners and managers (role: owner or manager).
// @Description staff_id (e.g. STF-001) and a 4-digit alphanumeric PIN are auto-generated.
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
	// 1. System admin can create owner and manager, but not staff.
	// 2. Manager can create staff, but not owner or manager.
	// 3. Other roles cannot create anyone.
	switch claims.Role {
	case auth.RoleSystemAdmin:
		if targetRole != auth.RoleOwner && targetRole != auth.RoleManager {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "system admin can only create an owner or a manager; staff members must be created by a manager",
			})
			return
		}

	case auth.RoleManager:
		if targetRole == "" {
			targetRole = auth.RoleStaff
		}
		if targetRole != auth.RoleStaff {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: "only system admin can create an owner or a manager; managers can only create staff",
			})
			return
		}

	default:
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: "permission denied: only managers can create staff, and only system admin can create owners or managers",
		})
		return
	}

	// Verify branch exists
	branchUUID, err := uuid.Parse(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid branch_id format (must be UUID)"})
		return
	}

	if _, err := h.queries.GetBranchByID(c.Request.Context(), branchUUID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "branch not found with specified branch_id"})
		return
	}

	// Check email uniqueness if email provided
	if strings.TrimSpace(req.Email) != "" {
		existing, err := h.queries.GetStaffByEmail(c.Request.Context(), sql.NullString{String: req.Email, Valid: true})
		if err == nil && existing.ID != uuid.Nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "email already in use by another staff member"})
			return
		}
	}

	// 1. Auto-generate sequential staff_id (e.g. STF-001)
	staffID, err := generator.GenerateStaffID(c.Request.Context(), h.queries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate unique staff_id: " + err.Error()})
		return
	}

	// 2. Auto-generate 4-digit alphanumeric PIN
	pin, err := generator.GenerateAlphanumericPIN(c.Request.Context(), h.queries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate unique pin: " + err.Error()})
		return
	}

	// 3. Hash PIN with bcrypt
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to hash credentials: " + err.Error()})
		return
	}

	newID := uuid.New()
	createdStaff, err := h.queries.CreateStaff(c.Request.Context(), sqlc.CreateStaffParams{
		ID:           newID,
		BranchID:     branchUUID,
		FullName:     strings.TrimSpace(req.FullName),
		Phone:        sql.NullString{String: strings.TrimSpace(req.Phone), Valid: strings.TrimSpace(req.Phone) != ""},
		Email:        sql.NullString{String: strings.TrimSpace(req.Email), Valid: strings.TrimSpace(req.Email) != ""},
		StaffID:      sql.NullString{String: staffID, Valid: true},
		Pin:          pin,
		PasswordHash: string(passwordHash),
		Role:         sqlc.StaffRole(targetRole),
		Status:       sqlc.RecordStatusActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save staff record: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CreateStaffResponse{
		ID:        createdStaff.ID.String(),
		BranchID:  createdStaff.BranchID.String(),
		FullName:  createdStaff.FullName,
		Phone:     createdStaff.Phone.String,
		Email:     createdStaff.Email.String,
		StaffID:   createdStaff.StaffID.String,
		Pin:       pin,
		Role:      string(createdStaff.Role),
		Status:    string(createdStaff.Status),
		CreatedAt: createdStaff.CreatedAt.Format(time.RFC3339),
	})
}

// ListStaff godoc
// @Summary List Staff Members
// @Description Retrieve a list of staff members. Accessible by system_admin, owner, and manager.
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

	// Managers are scoped to their branch unless a specific filter is passed
	if claims.Role == auth.RoleManager && branchFilter == "" && claims.BranchID != "" {
		branchFilter = claims.BranchID
	}

	var results []StaffProfile

	if branchFilter != "" {
		bUUID, err := uuid.Parse(branchFilter)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid branch_id query parameter"})
			return
		}
		rows, err := h.queries.ListStaffByBranch(c.Request.Context(), bUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve staff: " + err.Error()})
			return
		}
		for _, s := range rows {
			results = append(results, StaffProfile{
				ID:        s.ID.String(),
				StaffID:   s.StaffID.String,
				FullName:  s.FullName,
				Role:      string(s.Role),
				BranchID:  s.BranchID.String(),
				Phone:     s.Phone.String,
				Email:     s.Email.String,
				Status:    string(s.Status),
				CreatedAt: s.CreatedAt.Format(time.RFC3339),
			})
		}
	} else {
		rows, err := h.queries.ListStaff(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve staff: " + err.Error()})
			return
		}
		for _, s := range rows {
			results = append(results, StaffProfile{
				ID:        s.ID.String(),
				StaffID:   s.StaffID.String,
				FullName:  s.FullName,
				Role:      string(s.Role),
				BranchID:  s.BranchID.String(),
				Phone:     s.Phone.String,
				Email:     s.Email.String,
				Status:    string(s.Status),
				CreatedAt: s.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	c.JSON(http.StatusOK, results)
}

// GetStaff godoc
// @Summary Get Staff Details
// @Description Get details for a specific staff member by UUID.
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
	idParam := c.Param("id")
	targetUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid staff id format"})
		return
	}

	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthenticated"})
		return
	}

	// Staff role can only fetch their own profile
	if claims.Role == auth.RoleStaff && claims.UserID != idParam {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "permission denied"})
		return
	}

	staff, err := h.queries.GetStaffByID(c.Request.Context(), targetUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "staff record not found"})
		return
	}

	c.JSON(http.StatusOK, StaffProfile{
		ID:        staff.ID.String(),
		StaffID:   staff.StaffID.String,
		FullName:  staff.FullName,
		Role:      string(staff.Role),
		BranchID:  staff.BranchID.String(),
		Phone:     staff.Phone.String,
		Email:     staff.Email.String,
		Status:    string(staff.Status),
		CreatedAt: staff.CreatedAt.Format(time.RFC3339),
	})
}
