package handler

import (
	"net/http"
	"time"

	"backend/internal/db/sqlc"

	"github.com/gin-gonic/gin"
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
