package handler

import (
	"net/http"
	"strings"

	"backend/internal/auth"
	"backend/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FleetHandler handles fleet/vehicle management endpoints.
type FleetHandler struct {
	queries *sqlc.Queries
}

// NewFleetHandler constructs a FleetHandler.
func NewFleetHandler(queries *sqlc.Queries) *FleetHandler {
	return &FleetHandler{queries: queries}
}

// FleetResponse represents a fleet/vehicle summary.
type FleetResponse struct {
	ID                 string  `json:"id" example:"dddddddd-dddd-4ddd-8ddd-ddddddddddd1"`
	BranchID           string  `json:"branch_id" example:"11111111-1111-4111-8111-111111111111"`
	RegistrationNumber string  `json:"registration_number" example:"LAG-123-XY"`
	VehicleType        string  `json:"vehicle_type" example:"truck"`
	MakeModel          string  `json:"make_model,omitempty" example:"Mitsubishi Canter"`
	CapacityKg         string  `json:"capacity_kg,omitempty" example:"3500"`
	Status             string  `json:"status" example:"available"`
	CreatedAt          string  `json:"created_at"`
}

// ListFleets godoc
// @Summary List All Fleet Vehicles
// @Description Retrieve a list of all fleet vehicles.
// @Tags fleet
// @Produce json
// @Security BearerAuth
// @Success 200 {array} FleetResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/fleet [get]
func (h *FleetHandler) ListFleets(c *gin.Context) {
	rows, err := h.queries.ListFleets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve fleet: " + err.Error()})
		return
	}

	results := make([]FleetResponse, 0, len(rows))
	for _, f := range rows {
		results = append(results, newFleetResponse(f))
	}

	c.JSON(http.StatusOK, results)
}

// getTargetFleet loads the fleet record targeted by :id in the URL.
func (h *FleetHandler) getTargetFleet(c *gin.Context) (sqlc.Fleet, bool) {
	idParam := c.Param("id")
	fleetUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid fleet id format"})
		return sqlc.Fleet{}, false
	}

	fleet, err := h.queries.GetFleetByID(c.Request.Context(), fleetUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "fleet record not found"})
		return sqlc.Fleet{}, false
	}
	return fleet, true
}

// newFleetResponse maps a fleet row onto the response representation.
func newFleetResponse(f sqlc.Fleet) FleetResponse {
	return FleetResponse{
		ID:                 f.ID.String(),
		BranchID:           f.BranchID.String(),
		RegistrationNumber: f.RegistrationNumber,
		VehicleType:        f.VehicleType,
		MakeModel:          f.MakeModel.String,
		CapacityKg:         f.CapacityKg.String,
		Status:             string(f.Status),
		CreatedAt:          f.CreatedAt.Format(http.TimeFormat),
	}
}

// CreateFleetRequest defines the body for creating a fleet vehicle.
type CreateFleetRequest struct {
	BranchID           string `json:"branch_id" binding:"required" example:"11111111-1111-4111-8111-111111111111"`
	RegistrationNumber string `json:"registration_number" binding:"required" example:"LAG-123-XY"`
	VehicleType        string `json:"vehicle_type" binding:"required" example:"truck"`
	MakeModel          string `json:"make_model,omitempty" example:"Mitsubishi Canter"`
	CapacityKg         string `json:"capacity_kg,omitempty" example:"3500"`
}

// CreateFleet godoc
// @Summary Create a Fleet Vehicle
// @Description Creates a new fleet vehicle. Only the owner (or system admin) can manage fleet.
// @Tags fleet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateFleetRequest true "Fleet Vehicle Details"
// @Success 201 {object} FleetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/fleet [post]
func (h *FleetHandler) CreateFleet(c *gin.Context) {
	var req CreateFleetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	branchUUID, err := uuid.Parse(req.BranchID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid branch_id"})
		return
	}

	registrationNumber := strings.TrimSpace(req.RegistrationNumber)
	if existing, err := h.queries.GetFleetByRegistration(c.Request.Context(), registrationNumber); err == nil && existing.ID != uuid.Nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "registration number already in use"})
		return
	}

	capacityKg := strings.TrimSpace(req.CapacityKg)
	created, err := h.queries.CreateFleet(c.Request.Context(), sqlc.CreateFleetParams{
		ID:                 uuid.New(),
		BranchID:           branchUUID,
		RegistrationNumber: registrationNumber,
		VehicleType:        strings.TrimSpace(req.VehicleType),
		MakeModel:          nullIfEmpty(req.MakeModel),
		CapacityKg:         nullIfEmpty(capacityKg),
		Status:             sqlc.VehicleStatusAvailable,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create fleet vehicle: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newFleetResponse(created))
}

// UpdateFleetRequest defines the fields that may be updated on a fleet vehicle.
type UpdateFleetRequest struct {
	BranchID    string `json:"branch_id,omitempty" example:"11111111-1111-4111-8111-111111111111"`
	VehicleType string `json:"vehicle_type,omitempty" example:"truck"`
	MakeModel   string `json:"make_model,omitempty" example:"Mitsubishi Canter"`
	CapacityKg  string `json:"capacity_kg,omitempty" example:"3500"`
	Status      string `json:"status,omitempty" example:"available"`
}

// UpdateFleet godoc
// @Summary Update a Fleet Vehicle
// @Description Updates fleet vehicle details. Only the owner (or system admin) can manage fleet.
// @Tags fleet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Fleet UUID"
// @Param request body UpdateFleetRequest true "Fields to update"
// @Success 200 {object} FleetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/fleet/{id} [put]
func (h *FleetHandler) UpdateFleet(c *gin.Context) {
	_, ok := h.getTargetFleet(c)
	if !ok {
		return
	}

	var req UpdateFleetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request payload: " + err.Error()})
		return
	}

	params := sqlc.UpdateFleetParams{
		BranchID:    uuid.NullUUID{},
		VehicleType: nullIfEmpty(req.VehicleType),
		MakeModel:   nullIfEmpty(req.MakeModel),
		CapacityKg:  nullIfEmpty(req.CapacityKg),
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
	switch reqStatus {
	case "":
	case "available", "on_trip", "maintenance", "inactive":
		params.Status = sqlc.NullVehicleStatus{VehicleStatus: sqlc.VehicleStatus(reqStatus), Valid: true}
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "status must be one of: available, on_trip, maintenance, inactive"})
		return
	}

	updated, err := h.queries.UpdateFleet(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update fleet vehicle: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, newFleetResponse(updated))
}

// DeleteFleet godoc
// @Summary Delete a Fleet Vehicle
// @Description Marks a fleet vehicle inactive (soft delete, its data is retained). Only the owner (or system admin) can manage fleet.
// @Tags fleet
// @Produce json
// @Security BearerAuth
// @Param id path string true "Fleet UUID"
// @Success 200 {object} FleetResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/fleet/{id} [delete]
func (h *FleetHandler) DeleteFleet(c *gin.Context) {
	target, ok := h.getTargetFleet(c)
	if !ok {
		return
	}

	updated, err := h.queries.SoftDeleteFleet(c.Request.Context(), target.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete fleet vehicle: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, newFleetResponse(updated))
}

// unusedAuthImport references the auth package for the route guards defined in main.
var _ = auth.RequireRoles
