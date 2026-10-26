package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"backend/internal/auth"
	"backend/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DriverHandler handles driver management endpoints.
type DriverHandler struct {
	queries *sqlc.Queries
}

// NewDriverHandler constructs a DriverHandler.
func NewDriverHandler(queries *sqlc.Queries) *DriverHandler {
	return &DriverHandler{queries: queries}
}

// DriverResponse represents a driver summary.
type DriverResponse struct {
	ID            string `json:"id" example:"cccccccc-cccc-4ccc-8ccc-ccccccccccc1"`
	BranchID      string `json:"branch_id" example:"11111111-1111-4111-8111-111111111111"`
	FullName      string `json:"full_name" example:"Musa Ibrahim"`
	Phone         string `json:"phone" example:"08033330001"`
	LicenseNumber string `json:"license_number,omitempty" example:"DEMO-LIC-001"`
	LicenseExpiry string `json:"license_expiry,omitempty" example:"2027-12-31"`
	Status        string `json:"status" example:"on_trip"`
	CreatedAt     string `json:"created_at"`
}

// ListDrivers godoc
// @Summary List All Drivers
// @Description Retrieve a list of all drivers.
// @Tags drivers
// @Produce json
// @Security BearerAuth
// @Success 200 {array} DriverResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/drivers [get]
func (h *DriverHandler) ListDrivers(c *gin.Context) {
	rows, err := h.queries.ListDrivers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve drivers: " + err.Error()})
		return
	}

	results := make([]DriverResponse, 0, len(rows))
	for _, d := range rows {
		results = append(results, newDriverResponse(d))
	}

	c.JSON(http.StatusOK, results)
}

// getTargetDriver loads the driver record targeted by :id in the URL.
func (h *DriverHandler) getTargetDriver(c *gin.Context) (sqlc.Driver, bool) {
	idParam := c.Param("id")
	driverUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid driver id format"})
		return sqlc.Driver{}, false
	}

	driver, err := h.queries.GetDriverByID(c.Request.Context(), driverUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "driver record not found"})
		return sqlc.Driver{}, false
	}
	return driver, true
}

// newDriverResponse maps a driver row onto the response representation.
func newDriverResponse(d sqlc.Driver) DriverResponse {
	var expiry string
	if d.LicenseExpiry.Valid {
		expiry = d.LicenseExpiry.Time.Format("2006-01-02")
	}
	return DriverResponse{
		ID:            d.ID.String(),
		BranchID:      d.BranchID.String(),
		FullName:      d.FullName,
		Phone:         d.Phone,
		LicenseNumber: d.LicenseNumber.String,
		LicenseExpiry: expiry,
		Status:        string(d.Status),
	}
}

// CreateDriverRequest defines the body for creating a driver.
type CreateDriverRequest struct {
	BranchID       string `json:"branch_id" binding:"required" example:"11111111-1111-4111-8111-111111111111"`
	FullName       string `json:"full_name" binding:"required" example:"Musa Ibrahim"`
	Phone          string `json:"phone" binding:"required" example:"08033330001"`
	LicenseNumber  string `json:"license_number,omitempty" example:"DEMO-LIC-001"`
	LicenseExpiry  string `json:"license_expiry,omitempty" example:"2027-12-31"`
	Status         string `json:"status,omitempty" example:"available"`
}

// CreateDriver godoc
// @Summary Create a Driver
// @Description Creates a new driver. Only the owner (or system admin) can manage drivers.
// @Tags drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateDriverRequest true "Driver Details"
// @Success 201 {object} DriverResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/drivers [post]
func (h *DriverHandler) CreateDriver(c *gin.Context) {
	var req CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	branchUUID, err := uuid.Parse(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid branch_id"})
		return
	}

	// Parse license expiry as date (YYYY-MM-DD format)
	var expiryDate sql.NullTime
	if req.LicenseExpiry != "" {
		if t, err := time.Parse("2006-01-02", req.LicenseExpiry); err == nil {
			expiryDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	created, err := h.queries.CreateDriver(c.Request.Context(), sqlc.CreateDriverParams{
		ID:            uuid.New(),
		BranchID:      branchUUID,
		FullName:      strings.TrimSpace(req.FullName),
		Phone:         strings.TrimSpace(req.Phone),
		LicenseNumber: nullIfEmpty(req.LicenseNumber),
		LicenseExpiry: expiryDate,
		Status:        sqlc.DriverStatusAvailable,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create driver: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newDriverResponse(created))
}

// UpdateDriverRequest defines the fields that may be updated on a driver.
type UpdateDriverRequest struct {
	BranchID      string `json:"branch_id,omitempty" example:"11111111-1111-4111-8111-111111111111"`
	FullName      string `json:"full_name,omitempty" example:"Musa Ibrahim"`
	Phone         string `json:"phone,omitempty" example:"08033330001"`
	LicenseNumber string `json:"license_number,omitempty" example:"DEMO-LIC-001"`
	LicenseExpiry string `json:"license_expiry,omitempty" example:"2027-12-31"`
	Status        string `json:"status,omitempty" example:"available"`
}

// UpdateDriver godoc
// @Summary Update a Driver
// @Description Updates driver details. Only the owner (or system admin) can manage drivers.
// @Tags drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Driver UUID"
// @Param request body UpdateDriverRequest true "Fields to update"
// @Success 200 {object} DriverResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/drivers/{id} [put]
func (h *DriverHandler) UpdateDriver(c *gin.Context) {
	_, ok := h.getTargetDriver(c)
	if !ok {
		return
	}

	var req UpdateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	params := sqlc.UpdateDriverParams{
		BranchID:      uuid.NullUUID{},
		FullName:      nullIfEmpty(req.FullName),
		Phone:         nullIfEmpty(req.Phone),
		LicenseNumber: nullIfEmpty(req.LicenseNumber),
		LicenseExpiry: sql.NullTime{},
	}
	if req.BranchID != "" {
		branchUUID, err := uuid.Parse(req.BranchID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid branch_id"})
			return
		}
		params.BranchID = uuid.NullUUID{UUID: branchUUID, Valid: true}
	}

	reqStatus := strings.ToLower(strings.TrimSpace(req.Status))
	if reqStatus != "" {
		switch reqStatus {
		case "available", "on_trip", "inactive":
			params.Status = sqlc.NullDriverStatus{DriverStatus: sqlc.DriverStatus(reqStatus), Valid: true}
		default:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status must be one of: available, on_trip, inactive"})
			return
		}
	}

	updated, err := h.queries.UpdateDriver(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update driver: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, newDriverResponse(updated))
}

// DeleteDriver godoc
// @Summary Delete a Driver
// @Description Marks a driver inactive (soft delete, its data is retained). Only the owner (or system admin) can manage drivers.
// @Tags drivers
// @Produce json
// @Security BearerAuth
// @Param id path string true "Driver UUID"
// @Success 200 {object} DriverResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/drivers/{id} [delete]
func (h *DriverHandler) DeleteDriver(c *gin.Context) {
	target, ok := h.getTargetDriver(c)
	if !ok {
		return
	}

	updated, err := h.queries.SoftDeleteDriver(c.Request.Context(), target.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete driver: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, newDriverResponse(updated))
}

// unusedAuthImport references the auth package for the route guards defined in main.
var _ = auth.RequireRoles
