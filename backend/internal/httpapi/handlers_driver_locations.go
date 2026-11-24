package httpapi

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/bol-platform/backend/internal/auth"
	db "github.com/bol-platform/backend/internal/db/sqlc"
)

type driverPortalIdentity struct {
	DriverID     uuid.UUID
	CompanyID    uuid.UUID
	FullName     string
	VehiclePlate string
	CompanyName  string
}

// resolveDriverPortalIdentity accepts both the current signed portal token and
// the legacy hashed token issued before signed links were introduced.
func (a *API) resolveDriverPortalIdentity(ctx context.Context, token string) (driverPortalIdentity, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return driverPortalIdentity{}, pgx.ErrNoRows
	}

	if driverID, err := a.auth.ParseDriverPortalToken(token); err == nil {
		driver, queryErr := a.svc.Q.GetDriverProfileForPortalByID(ctx, driverID)
		if queryErr != nil {
			return driverPortalIdentity{}, queryErr
		}
		return driverPortalIdentity{
			DriverID:     driver.ID,
			CompanyID:    driver.CompanyID,
			FullName:     driver.FullName,
			VehiclePlate: driver.VehiclePlate,
			CompanyName:  driver.CompanyName,
		}, nil
	}

	driver, err := a.svc.Q.GetDriverProfileForPortal(ctx, auth.HashToken(token))
	if err != nil {
		return driverPortalIdentity{}, err
	}
	return driverPortalIdentity{
		DriverID:     driver.ID,
		CompanyID:    driver.CompanyID,
		FullName:     driver.FullName,
		VehiclePlate: driver.VehiclePlate,
		CompanyName:  driver.CompanyName,
	}, nil
}

func (a *API) driverIdentityOrFail(c *gin.Context) (driverPortalIdentity, bool) {
	identity, err := a.resolveDriverPortalIdentity(c.Request.Context(), c.Param("token"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fail(c, http.StatusNotFound, "driver_portal_not_found", "This driver portal link is invalid or inactive.")
			return driverPortalIdentity{}, false
		}
		a.handleServiceError(c, err)
		return driverPortalIdentity{}, false
	}
	return identity, true
}

type liveDriverLocationDTO struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	UpdatedAt time.Time `json:"updated_at"`
}

type liveDriverDTO struct {
	DriverID       uuid.UUID              `json:"driver_id"`
	FullName       string                 `json:"full_name"`
	VehiclePlate   string                 `json:"vehicle_plate"`
	DeliveryID     uuid.UUID              `json:"delivery_id"`
	DeliveryStatus string                 `json:"delivery_status"`
	TrackingCode   *string                `json:"tracking_code"`
	PickupAddress  string                 `json:"pickup_address"`
	DropoffAddress string                 `json:"dropoff_address"`
	Location       *liveDriverLocationDTO `json:"location"`
}

func (a *API) listLiveDrivers(c *gin.Context) {
	tenant, _ := currentTenant(c)
	rows, err := a.svc.Q.ListDriversOnRoad(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	result := make([]liveDriverDTO, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.DriverID]; ok {
			continue
		}
		seen[row.DriverID] = struct{}{}

		driver := liveDriverDTO{
			DriverID:       row.DriverID,
			FullName:       row.FullName,
			VehiclePlate:   row.VehiclePlate,
			DeliveryID:     row.DeliveryID,
			DeliveryStatus: row.DeliveryStatus,
			TrackingCode:   row.TrackingCode,
			PickupAddress:  row.PickupAddress,
			DropoffAddress: row.DropoffAddress,
		}
		if row.Latitude.Valid && row.Longitude.Valid && row.LocationUpdatedAt != nil {
			driver.Location = &liveDriverLocationDTO{
				Latitude:  row.Latitude.Float64,
				Longitude: row.Longitude.Float64,
				UpdatedAt: *row.LocationUpdatedAt,
			}
		}
		result = append(result, driver)
	}

	c.JSON(http.StatusOK, gin.H{"drivers": result})
}

func (a *API) updateDriverLocation(c *gin.Context) {
	identity, ok := a.driverIdentityOrFail(c)
	if !ok {
		return
	}

	var req struct {
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		AccuracyM *float64 `json:"accuracy_m"`
		Heading   *float64 `json:"heading"`
		SpeedKph  *float64 `json:"speed_kph"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_location", "Send a valid latitude and longitude.")
		return
	}
	if req.Latitude == nil || req.Longitude == nil ||
		!finiteNumber(*req.Latitude) || !finiteNumber(*req.Longitude) ||
		*req.Latitude < -90 || *req.Latitude > 90 ||
		*req.Longitude < -180 || *req.Longitude > 180 {
		fail(c, http.StatusBadRequest, "invalid_location", "Send a valid latitude and longitude.")
		return
	}
	if req.AccuracyM != nil && (!finiteNumber(*req.AccuracyM) || *req.AccuracyM < 0) {
		fail(c, http.StatusBadRequest, "invalid_location", "Location accuracy cannot be negative.")
		return
	}
	if req.Heading != nil && (!finiteNumber(*req.Heading) || *req.Heading < 0 || *req.Heading >= 360) {
		fail(c, http.StatusBadRequest, "invalid_location", "Heading must be between 0 and 359 degrees.")
		return
	}
	if req.SpeedKph != nil && (!finiteNumber(*req.SpeedKph) || *req.SpeedKph < 0) {
		fail(c, http.StatusBadRequest, "invalid_location", "Speed cannot be negative.")
		return
	}

	if _, err := a.svc.Q.GetDriverActiveAssignment(c.Request.Context(), db.GetDriverActiveAssignmentParams{
		CompanyID:        identity.CompanyID,
		AssignedDriverID: &identity.DriverID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fail(c, http.StatusConflict, "no_active_delivery", "Location sharing is only available during an active assigned delivery.")
			return
		}
		a.handleServiceError(c, err)
		return
	}

	location, err := a.svc.Q.UpsertDriverLocation(c.Request.Context(), db.UpsertDriverLocationParams{
		DriverID:  identity.DriverID,
		CompanyID: identity.CompanyID,
		Latitude:  *req.Latitude,
		Longitude: *req.Longitude,
		AccuracyM: nullableFloat8(req.AccuracyM),
		Heading:   nullableFloat8(req.Heading),
		SpeedKph:  nullableFloat8(req.SpeedKph),
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"location": liveDriverLocationDTO{
			Latitude:  location.Latitude,
			Longitude: location.Longitude,
			UpdatedAt: location.UpdatedAt,
		},
	})
}

func (a *API) stopDriverLocation(c *gin.Context) {
	identity, ok := a.driverIdentityOrFail(c)
	if !ok {
		return
	}
	if err := a.svc.Q.StopDriverLocation(c.Request.Context(), db.StopDriverLocationParams{
		DriverID:  identity.DriverID,
		CompanyID: identity.CompanyID,
	}); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func finiteNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func nullableFloat8(value *float64) pgtype.Float8 {
	if value == nil {
		return pgtype.Float8{}
	}
	return pgtype.Float8{Float64: *value, Valid: true}
}
