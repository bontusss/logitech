package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/bol-platform/backend/internal/db/sqlc"
)

type bookingRouteDTO struct {
	ID                      uuid.UUID `json:"id"`
	OriginState             string    `json:"origin_state"`
	OriginCity              string    `json:"origin_city"`
	DestinationState        string    `json:"destination_state"`
	DestinationCity         string    `json:"destination_city"`
	DistanceKm              float64   `json:"distance_km"`
	StandardPricePerKmMinor *int64    `json:"standard_price_per_km_minor"`
	ExpressPricePerKmMinor  *int64    `json:"express_price_per_km_minor"`
	SameDayPricePerKmMinor  *int64    `json:"same_day_price_per_km_minor"`
	MinimumChargeMinor      int64     `json:"minimum_charge_minor"`
	Status                  string    `json:"status"`
}

func bookingRouteFromDB(route db.CompanyBookingRoute) bookingRouteDTO {
	return bookingRouteDTO{
		ID:                      route.ID,
		OriginState:             route.OriginState,
		OriginCity:              route.OriginCity,
		DestinationState:        route.DestinationState,
		DestinationCity:         route.DestinationCity,
		DistanceKm:              numericToFloat64(route.DistanceKm),
		StandardPricePerKmMinor: int8Pointer(route.StandardPricePerKmMinor),
		ExpressPricePerKmMinor:  int8Pointer(route.ExpressPricePerKmMinor),
		SameDayPricePerKmMinor:  int8Pointer(route.SameDayPricePerKmMinor),
		MinimumChargeMinor:      route.MinimumChargeMinor,
		Status:                  route.Status,
	}
}

func (a *API) listCompanyBookingRoutes(c *gin.Context) {
	tenant, _ := currentTenant(c)
	routes, err := a.svc.Q.ListCompanyBookingRoutes(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]bookingRouteDTO, 0, len(routes))
	for _, route := range routes {
		result = append(result, bookingRouteFromDB(route))
	}
	c.JSON(http.StatusOK, gin.H{"routes": result})
}

func (a *API) providerOffersService(c *gin.Context, service string) bool {
	tenant, _ := currentTenant(c)
	company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		return false
	}
	businessType := strings.ToLower(company.BusinessType)
	return strings.Contains(businessType, service)
}

func (a *API) createCompanyBookingRoute(c *gin.Context) {
	tenant, _ := currentTenant(c)
	if !a.providerOffersService(c, "intercity") {
		fail(c, http.StatusForbidden, "service_not_offered", "This provider does not offer inter-city services.")
		return
	}
	var req bookingRouteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_booking_route", "Add the origin, destination, distance and at least one service price.")
		return
	}
	params, ok := req.toDBParams(tenant.CompanyID)
	if !ok {
		fail(c, http.StatusBadRequest, "invalid_booking_route", "Check the route fields and enter at least one valid price per kilometre.")
		return
	}
	created, err := a.svc.Q.CreateCompanyBookingRoute(c.Request.Context(), params)
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "booking_route_exists", "This origin and destination route already exists.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"route": bookingRouteFromDB(created)})
}

func (a *API) updateCompanyBookingRoute(c *gin.Context) {
	tenant, _ := currentTenant(c)
	if !a.providerOffersService(c, "intercity") {
		fail(c, http.StatusForbidden, "service_not_offered", "This provider does not offer inter-city services.")
		return
	}
	routeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_booking_route", "That route id is invalid.")
		return
	}
	var req bookingRouteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_booking_route", "Add the origin, destination, distance and at least one service price.")
		return
	}
	params, ok := req.toDBParams(tenant.CompanyID)
	if !ok {
		fail(c, http.StatusBadRequest, "invalid_booking_route", "Check the route fields and enter at least one valid price per kilometre.")
		return
	}
	updated, err := a.svc.Q.UpdateCompanyBookingRoute(c.Request.Context(), db.UpdateCompanyBookingRouteParams{
		CompanyID:               tenant.CompanyID,
		ID:                      routeID,
		OriginState:             params.OriginState,
		OriginCity:              params.OriginCity,
		DestinationState:        params.DestinationState,
		DestinationCity:         params.DestinationCity,
		DistanceKm:              params.DistanceKm,
		StandardPricePerKmMinor: params.StandardPricePerKmMinor,
		ExpressPricePerKmMinor:  params.ExpressPricePerKmMinor,
		SameDayPricePerKmMinor:  params.SameDayPricePerKmMinor,
		MinimumChargeMinor:      params.MinimumChargeMinor,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "booking_route_not_found", "That route was not found.")
			return
		}
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "booking_route_exists", "This origin and destination route already exists.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"route": bookingRouteFromDB(updated)})
}

func (a *API) duplicateCompanyBookingRoute(c *gin.Context) {
	tenant, _ := currentTenant(c)
	if !a.providerOffersService(c, "intercity") {
		fail(c, http.StatusForbidden, "service_not_offered", "This provider does not offer inter-city services.")
		return
	}
	routeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_booking_route", "That route id is invalid.")
		return
	}
	duplicated, err := a.svc.Q.DuplicateCompanyBookingRoute(c.Request.Context(), db.DuplicateCompanyBookingRouteParams{CompanyID: tenant.CompanyID, ID: routeID})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "booking_route_not_found", "That route was not found.")
			return
		}
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "booking_route_exists", "This lane already exists. Edit the existing route or change the lane.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"route": bookingRouteFromDB(duplicated)})
}

func (a *API) deleteCompanyBookingRoute(c *gin.Context) {
	tenant, _ := currentTenant(c)
	if !a.providerOffersService(c, "intercity") {
		fail(c, http.StatusForbidden, "service_not_offered", "This provider does not offer inter-city services.")
		return
	}
	routeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_booking_route", "That route id is invalid.")
		return
	}
	if err := a.svc.Q.DeleteCompanyBookingRoute(c.Request.Context(), db.DeleteCompanyBookingRouteParams{CompanyID: tenant.CompanyID, ID: routeID}); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (a *API) setCompanyBookingRouteStatus(c *gin.Context) {
	tenant, _ := currentTenant(c)
	if !a.providerOffersService(c, "intercity") {
		fail(c, http.StatusForbidden, "service_not_offered", "This provider does not offer inter-city services.")
		return
	}
	routeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_booking_route", "That route id is invalid.")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_booking_route_status", "Choose active or inactive.")
		return
	}
	updated, err := a.svc.Q.SetCompanyBookingRouteStatus(c.Request.Context(), db.SetCompanyBookingRouteStatusParams{
		CompanyID: tenant.CompanyID,
		ID:        routeID,
		Status:    req.Status,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "booking_route_not_found", "That route was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"route": bookingRouteFromDB(updated)})
}

type bookingRouteInput struct {
	OriginState        string   `json:"origin_state" binding:"required,min=2,max=100"`
	OriginCity         string   `json:"origin_city" binding:"required,min=2,max=100"`
	DestinationState   string   `json:"destination_state" binding:"required,min=2,max=100"`
	DestinationCity    string   `json:"destination_city" binding:"required,min=2,max=100"`
	DistanceKm         float64  `json:"distance_km" binding:"required,gt=0,lte=10000"`
	StandardPricePerKm *float64 `json:"standard_price_per_km"`
	ExpressPricePerKm  *float64 `json:"express_price_per_km"`
	SameDayPricePerKm  *float64 `json:"same_day_price_per_km"`
	MinimumCharge      float64  `json:"minimum_charge" binding:"gte=0"`
}

func (req bookingRouteInput) toDBParams(companyID uuid.UUID) (db.CreateCompanyBookingRouteParams, bool) {
	standard, standardOK := parseRoutePrice(req.StandardPricePerKm)
	express, expressOK := parseRoutePrice(req.ExpressPricePerKm)
	sameDay, sameDayOK := parseRoutePrice(req.SameDayPricePerKm)
	minimum, minimumOK := amountToMinor(req.MinimumCharge)
	if !standardOK || !expressOK || !sameDayOK || !minimumOK ||
		(standard == nil && express == nil && sameDay == nil) {
		return db.CreateCompanyBookingRouteParams{}, false
	}
	return db.CreateCompanyBookingRouteParams{
		CompanyID:               companyID,
		OriginState:             strings.TrimSpace(req.OriginState),
		OriginCity:              strings.TrimSpace(req.OriginCity),
		DestinationState:        strings.TrimSpace(req.DestinationState),
		DestinationCity:         strings.TrimSpace(req.DestinationCity),
		DistanceKm:              float64ToNumeric(req.DistanceKm),
		StandardPricePerKmMinor: nullableInt8(standard),
		ExpressPricePerKmMinor:  nullableInt8(express),
		SameDayPricePerKmMinor:  nullableInt8(sameDay),
		MinimumChargeMinor:      minimum,
	}, true
}

func nullableInt8(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}

func int8Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func parseRoutePrice(amount *float64) (*int64, bool) {
	if amount == nil {
		return nil, true
	}
	minor, ok := amountToMinor(*amount)
	if !ok || minor == 0 {
		return nil, false
	}
	return &minor, true
}

func amountToMinor(amount float64) (int64, bool) {
	if amount < 0 || amount > 1_000_000_000 || amount*100 > float64(int64(^uint64(0)>>1)) {
		return 0, false
	}
	return int64(amount*100 + 0.5), true
}
