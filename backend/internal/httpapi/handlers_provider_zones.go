package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

func providerZoneDTO(row db.ListCompanyServiceZonesRow) gin.H {
	return gin.H{
		"id":             row.ID,
		"name":           row.Name,
		"geo_zone_id":    row.GeoZoneID,
		"state":          row.GeoState,
		"lga":            row.GeoLga,
		"geo_name":       row.GeoName,
		"parent_zone_id": row.ParentZoneID,
		"created_at":     row.CreatedAt,
		"updated_at":     row.UpdatedAt,
	}
}

func providerZoneRouteDTO(row db.ListCompanyZoneRoutesRow) gin.H {
	return gin.H{
		"id":                          row.ID,
		"from_zone_id":                row.FromZoneID,
		"from_zone_name":              row.FromZoneName,
		"from_state":                  row.FromState,
		"from_lga":                    row.FromLga,
		"to_zone_id":                  row.ToZoneID,
		"to_zone_name":                row.ToZoneName,
		"to_state":                    row.ToState,
		"to_lga":                      row.ToLga,
		"price_minor":                 row.PriceMinor,
		"included_weight_kg":          numericToFloat64(row.IncludedWeightKg),
		"additional_weight_fee_minor": row.AdditionalWeightFeeMinor,
		"status":                      row.Status,
		"created_at":                  row.CreatedAt,
		"updated_at":                  row.UpdatedAt,
	}
}

func (a *API) listGeoStates(c *gin.Context) {
	states, err := a.svc.Q.ListGeoStates(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"states": nonNilStrings(states)})
}

func (a *API) listGeoLgas(c *gin.Context) {
	state := strings.TrimSpace(c.Query("state"))
	if state == "" {
		fail(c, http.StatusBadRequest, "invalid_state", "Choose a state.")
		return
	}
	lgas, err := a.svc.Q.ListGeoLgas(c.Request.Context(), state)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"state": state, "lgas": nonNilStrings(lgas)})
}

func (a *API) listGeoZones(c *gin.Context) {
	params := db.ListGeoZonesParams{}
	if state := strings.TrimSpace(c.Query("state")); state != "" {
		params.State = &state
	}
	if lga := strings.TrimSpace(c.Query("lga")); lga != "" {
		params.Lga = &lga
	}
	rows, err := a.svc.Q.ListGeoZones(c.Request.Context(), params)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	zones := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		zones = append(zones, gin.H{
			"id": row.ID, "state": row.State, "lga": row.Lga, "name": row.Name,
			"parent_id": row.ParentID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"zones": zones})
}

func (a *API) listReachableZones(c *gin.Context) {
	originID, err := uuid.Parse(c.Query("origin_zone_id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone", "Choose a pickup zone.")
		return
	}
	rows, err := a.svc.Q.ListReachableDestinationZones(c.Request.Context(), originID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	zones := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		zones = append(zones, gin.H{"id": row.ID, "state": row.State, "lga": row.Lga, "name": row.Name})
	}
	c.JSON(http.StatusOK, gin.H{"zones": zones})
}

func (a *API) listProviderServiceZones(c *gin.Context) {
	tenant, _ := currentTenant(c)
	rows, err := a.svc.Q.ListCompanyServiceZones(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	zones := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		zones = append(zones, providerZoneDTO(row))
	}
	c.JSON(http.StatusOK, gin.H{"zones": zones})
}

func (a *API) createProviderServiceZone(c *gin.Context) {
	tenant, _ := currentTenant(c)
	var req struct {
		GeoZoneID    string  `json:"geo_zone_id" binding:"required,uuid"`
		ParentZoneID *string `json:"parent_zone_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "Choose a state and LGA for this zone.")
		return
	}
	geoZoneID, err := uuid.Parse(req.GeoZoneID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "Choose a valid state and LGA.")
		return
	}
	var parentID *uuid.UUID
	if req.ParentZoneID != nil && strings.TrimSpace(*req.ParentZoneID) != "" {
		parsed, parseErr := uuid.Parse(strings.TrimSpace(*req.ParentZoneID))
		if parseErr != nil {
			fail(c, http.StatusBadRequest, "invalid_parent_zone", "Choose a valid parent zone.")
			return
		}
		parentID = &parsed
	}
	zone, err := a.svc.Q.CreateCompanyServiceZone(c.Request.Context(), db.CreateCompanyServiceZoneParams{
		CompanyID:    tenant.CompanyID,
		GeoZoneID:    geoZoneID,
		ParentZoneID: parentID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusUnprocessableEntity, "invalid_service_zone", "That state or LGA is not available.")
			return
		}
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "service_zone_exists", "You already cover that LGA.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"zone": gin.H{
		"id": zone.ID, "name": zone.Name, "geo_zone_id": zone.GeoZoneID,
		"parent_zone_id": zone.ParentZoneID, "state": zone.State, "lga": zone.Lga,
		"created_at": zone.CreatedAt, "updated_at": zone.UpdatedAt,
	}})
}

func (a *API) deleteProviderServiceZone(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "That zone id is invalid.")
		return
	}
	if err := a.svc.Q.DeleteCompanyServiceZone(c.Request.Context(), db.DeleteCompanyServiceZoneParams{
		CompanyID: tenant.CompanyID, ID: id,
	}); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) listProviderZoneRoutes(c *gin.Context) {
	tenant, _ := currentTenant(c)
	rows, err := a.svc.Q.ListCompanyZoneRoutes(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	routes := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		routes = append(routes, providerZoneRouteDTO(row))
	}
	c.JSON(http.StatusOK, gin.H{"routes": routes})
}

type providerZoneRouteInput struct {
	FromZoneID               string  `json:"from_zone_id" binding:"required,uuid"`
	ToZoneID                 string  `json:"to_zone_id" binding:"required,uuid"`
	PriceMinor               int64   `json:"price_minor" binding:"gte=0"`
	IncludedWeightKg         float64 `json:"included_weight_kg" binding:"gte=0,lte=500"`
	AdditionalWeightFeeMinor int64   `json:"additional_weight_fee_minor" binding:"gte=0"`
	Status                   string  `json:"status" binding:"omitempty,oneof=active inactive"`
}

func (r providerZoneRouteInput) valid() bool {
	return strings.TrimSpace(r.FromZoneID) != "" && strings.TrimSpace(r.ToZoneID) != ""
}

func (a *API) createProviderZoneRoute(c *gin.Context) {
	tenant, _ := currentTenant(c)
	var req providerZoneRouteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_route", "Choose both zones and enter valid prices.")
		return
	}
	fromID, errFrom := uuid.Parse(req.FromZoneID)
	toID, errTo := uuid.Parse(req.ToZoneID)
	if errFrom != nil || errTo != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_route", "Choose valid zones for this route.")
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	route, err := a.svc.Q.CreateCompanyZoneRoute(c.Request.Context(), db.CreateCompanyZoneRouteParams{
		CompanyID:                tenant.CompanyID,
		FromZoneID:               fromID,
		ToZoneID:                 toID,
		PriceMinor:               req.PriceMinor,
		IncludedWeightKg:         float64ToNumeric(req.IncludedWeightKg),
		AdditionalWeightFeeMinor: req.AdditionalWeightFeeMinor,
		Status:                   req.Status,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusUnprocessableEntity, "invalid_zone_route", "Both zones must belong to your company.")
			return
		}
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "zone_route_exists", "A route already exists in that direction.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"route": gin.H{
		"id": route.ID, "from_zone_id": route.FromZoneID, "to_zone_id": route.ToZoneID,
		"price_minor": route.PriceMinor, "included_weight_kg": numericToFloat64(route.IncludedWeightKg),
		"additional_weight_fee_minor": route.AdditionalWeightFeeMinor, "status": route.Status,
	}})
}

func (a *API) updateProviderZoneRoute(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_route", "That route id is invalid.")
		return
	}
	var req providerZoneRouteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_route", "Choose both zones and enter valid prices.")
		return
	}
	fromID, errFrom := uuid.Parse(req.FromZoneID)
	toID, errTo := uuid.Parse(req.ToZoneID)
	if errFrom != nil || errTo != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_route", "Choose valid zones for this route.")
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	route, err := a.svc.Q.UpdateCompanyZoneRoute(c.Request.Context(), db.UpdateCompanyZoneRouteParams{
		CompanyID:                tenant.CompanyID,
		ID:                       id,
		FromZoneID:               fromID,
		ToZoneID:                 toID,
		PriceMinor:               req.PriceMinor,
		IncludedWeightKg:         float64ToNumeric(req.IncludedWeightKg),
		AdditionalWeightFeeMinor: req.AdditionalWeightFeeMinor,
		Status:                   req.Status,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "zone_route_not_found", "That route was not found.")
			return
		}
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "zone_route_exists", "A route already exists in that direction.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"route": gin.H{
		"id": route.ID, "from_zone_id": route.FromZoneID, "to_zone_id": route.ToZoneID,
		"price_minor": route.PriceMinor, "included_weight_kg": numericToFloat64(route.IncludedWeightKg),
		"additional_weight_fee_minor": route.AdditionalWeightFeeMinor, "status": route.Status,
	}})
}

func (a *API) setProviderZoneRouteStatus(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_route", "That route id is invalid.")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_status", "Choose active or inactive.")
		return
	}
	route, err := a.svc.Q.SetCompanyZoneRouteStatus(c.Request.Context(), db.SetCompanyZoneRouteStatusParams{
		CompanyID: tenant.CompanyID, ID: id, Status: req.Status,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "zone_route_not_found", "That route was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"route": gin.H{"id": route.ID, "status": route.Status}})
}

func (a *API) deleteProviderZoneRoute(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_route", "That route id is invalid.")
		return
	}
	if err := a.svc.Q.DeleteCompanyZoneRoute(c.Request.Context(), db.DeleteCompanyZoneRouteParams{
		CompanyID: tenant.CompanyID, ID: id,
	}); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *API) previewProviderZoneRoute(c *gin.Context) {
	tenant, _ := currentTenant(c)
	fromID, errFrom := uuid.Parse(c.Query("from_zone_id"))
	toID, errTo := uuid.Parse(c.Query("to_zone_id"))
	weight, errWeight := parsePositiveFloat(c.Query("weight_kg"))
	if errFrom != nil || errTo != nil || errWeight != nil {
		fail(c, http.StatusBadRequest, "invalid_preview", "Choose both zones and enter a package weight.")
		return
	}
	rows, err := a.svc.Q.ListCompanyZoneRoutes(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	for _, row := range rows {
		if row.FromZoneID != fromID || row.ToZoneID != toID || row.Status != "active" {
			continue
		}
		price, priceErr := service.CalculateZoneRoutePrice(service.ZoneRoutePricing{
			BasePriceMinor:           row.PriceMinor,
			IncludedWeightKg:         numericToFloat64(row.IncludedWeightKg),
			AdditionalWeightFeeMinor: row.AdditionalWeightFeeMinor,
		}, weight)
		if priceErr != nil {
			fail(c, http.StatusUnprocessableEntity, "missing_rate", "This route has no usable price.")
			return
		}
		c.JSON(http.StatusOK, gin.H{"quote": price})
		return
	}
	fail(c, http.StatusNotFound, "zone_route_not_found", "No active route matches those zones.")
}

func parsePositiveFloat(value string) (float64, error) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || parsed <= 0 {
		return 0, service.ErrInvalidWeight
	}
	return parsed, nil
}
