package handler

import (
	"net/http"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BranchHandler handles branch lookup endpoints.
type BranchHandler struct {
	queries *sqlc.Queries
}

// NewBranchHandler constructs a BranchHandler.
func NewBranchHandler(queries *sqlc.Queries) *BranchHandler {
	return &BranchHandler{
		queries: queries,
	}
}

// BranchResponse represents a branch summary.
type BranchResponse struct {
	ID        string `json:"id" example:"11111111-1111-4111-8111-111111111111"`
	Name      string `json:"name" example:"Lagos Main Branch"`
	Code      string `json:"code" example:"LAG-01"`
	Address   string `json:"address,omitempty" example:"12 Apapa Road"`
	State     string `json:"state,omitempty" example:"Lagos"`
	Phone     string `json:"phone,omitempty" example:"08030000001"`
	Status    string `json:"status" example:"active"`
	CreatedAt string `json:"created_at"`
}

// ListBranches godoc
// @Summary List All Branches
// @Description Retrieve a list of all branches.
// @Tags branches
// @Produce json
// @Security BearerAuth
// @Success 200 {array} BranchResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/branches [get]
func (h *BranchHandler) ListBranches(c *gin.Context) {
	rows, err := h.queries.ListBranches(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve branches: " + err.Error()})
		return
	}

	results := make([]BranchResponse, 0, len(rows))
	for _, b := range rows {
		results = append(results, BranchResponse{
			ID:        b.ID.String(),
			Name:      b.Name,
			Code:      b.Code,
			Address:   b.Address.String,
			State:     b.State.String,
			Phone:     b.Phone.String,
			Status:    string(b.Status),
			CreatedAt: b.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, results)
}

// getTargetBranch loads the branch record targeted by :id in the URL.
func (h *BranchHandler) getTargetBranch(c *gin.Context) (sqlc.Branch, bool) {
	idParam := c.Param("id")
	branchUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid branch id format"})
		return sqlc.Branch{}, false
	}

	branch, err := h.queries.GetBranchByID(c.Request.Context(), branchUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "branch not found"})
		return sqlc.Branch{}, false
	}
	return branch, true
}

// newBranchResponse maps a branch row onto the response representation.
func newBranchResponse(b sqlc.Branch) BranchResponse {
	return BranchResponse{
		ID:        b.ID.String(),
		Name:      b.Name,
		Code:      b.Code,
		Address:   b.Address.String,
		State:     b.State.String,
		Phone:     b.Phone.String,
		Status:    string(b.Status),
		CreatedAt: b.CreatedAt.Format(time.RFC3339),
	}
}

// CreateBranchRequest defines the body for creating a branch.
type CreateBranchRequest struct {
	Name    string `json:"name" binding:"required" example:"Kano Branch"`
	Code    string `json:"code" binding:"required" example:"KAN-01"`
	Address string `json:"address,omitempty" example:"5 Zoo Road"`
	State   string `json:"state,omitempty" example:"Kano"`
	Phone   string `json:"phone,omitempty" example:"08030000004"`
}

// CreateBranch godoc
// @Summary Create a branch
// @Description Creates a new branch. Only the owner (or system admin) can manage branches.
// @Tags branches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateBranchRequest true "Branch Details"
// @Success 201 {object} BranchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/branches [post]
func (h *BranchHandler) CreateBranch(c *gin.Context) {
	var req CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	code := strings.TrimSpace(req.Code)
	if existing, err := h.queries.GetBranchByCode(c.Request.Context(), code); err == nil && existing.ID != uuid.Nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "branch code already in use"})
		return
	}

	created, err := h.queries.CreateBranch(c.Request.Context(), sqlc.CreateBranchParams{
		ID:      uuid.New(),
		Name:    strings.TrimSpace(req.Name),
		Code:    code,
		Address: nullIfEmpty(req.Address),
		State:   nullIfEmpty(req.State),
		Phone:   nullIfEmpty(req.Phone),
		Status:  sqlc.RecordStatusActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create branch: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newBranchResponse(created))
}

// UpdateBranchRequest defines the fields that may be updated on a branch.
type UpdateBranchRequest struct {
	Name    string `json:"name,omitempty" example:"Kano Branch"`
	Address string `json:"address,omitempty" example:"5 Zoo Road"`
	State   string `json:"state,omitempty" example:"Kano"`
	Phone   string `json:"phone,omitempty" example:"08030000004"`
	Status  string `json:"status,omitempty" example:"active"`
}

// UpdateBranch godoc
// @Summary Update a branch
// @Description Updates branch details. Only the owner (or system admin) can manage branches.
// @Tags branches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Branch UUID"
// @Param request body UpdateBranchRequest true "Fields to update"
// @Success 200 {object} BranchResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/branches/{id} [put]
func (h *BranchHandler) UpdateBranch(c *gin.Context) {
	target, ok := h.getTargetBranch(c)
	if !ok {
		return
	}

	var req UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	params := sqlc.UpdateBranchParams{
		ID:      target.ID,
		Name:    nullIfEmpty(req.Name),
		Address: nullIfEmpty(req.Address),
		State:   nullIfEmpty(req.State),
		Phone:   nullIfEmpty(req.Phone),
	}
	reqStatus := strings.ToLower(strings.TrimSpace(req.Status))
	switch reqStatus {
	case "":
	case "active", "inactive":
		params.Status = sqlc.NullRecordStatus{RecordStatus: sqlc.RecordStatus(reqStatus), Valid: true}
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status must be active or inactive"})
		return
	}

	updated, err := h.queries.UpdateBranch(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update branch: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, newBranchResponse(updated))
}

// DeleteBranch godoc
// @Summary Delete a branch
// @Description Marks a branch inactive (soft delete, its data is retained). Only the owner (or system admin) can manage branches.
// @Tags branches
// @Produce json
// @Security BearerAuth
// @Param id path string true "Branch UUID"
// @Success 200 {object} BranchResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/branches/{id} [delete]
func (h *BranchHandler) DeleteBranch(c *gin.Context) {
	target, ok := h.getTargetBranch(c)
	if !ok {
		return
	}

	updated, err := h.queries.SoftDeleteBranch(c.Request.Context(), target.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete branch: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, newBranchResponse(updated))
}

// unusedAuthImport references the auth package for the route guards defined in main.
var _ = auth.RequireRoles
