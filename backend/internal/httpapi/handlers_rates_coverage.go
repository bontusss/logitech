package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/bol-platform/backend/internal/db/sqlc"
)

func ptrString(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func nullableInt8Minor(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

func int8Value(v pgtype.Int8) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func floatNumericPtr(v *float64) pgtype.Numeric {
	if v == nil {
		return pgtype.Numeric{}
	}
	return float64ToNumeric(*v)
}

func lastMileAreaDTO(a db.LastMileServiceArea) gin.H {
	return gin.H{
		"id": a.ID, "name": a.Name, "state": a.State, "city": a.City, "coverage_status": a.CoverageStatus,
		"pricing_model": a.PricingModel, "flat_price_minor": int8Value(a.FlatPriceMinor),
		"base_delivery_fee_minor": int8Value(a.BaseDeliveryFeeMinor), "price_per_km_minor": int8Value(a.PricePerKmMinor),
		"minimum_fee_minor": int8Value(a.MinimumFeeMinor), "maximum_distance_km": numericToFloat64(a.MaximumDistanceKm),
		"maximum_fee_minor": int8Value(a.MaximumFeeMinor), "volumetric_divisor": numericToFloat64(a.VolumetricDivisor),
		"price_is_estimate": a.PriceIsEstimate, "max_weight_kg": numericToFloat64(a.MaxWeightKg),
		"max_length_cm": numericToFloat64(a.MaxLengthCm), "max_width_cm": numericToFloat64(a.MaxWidthCm), "max_height_cm": numericToFloat64(a.MaxHeightCm),
		"express_available": a.ExpressAvailable, "pickup_available": a.PickupAvailable, "delivery_available": a.DeliveryAvailable,
		"operating_hours": a.OperatingHours, "supported_parcel_types": a.SupportedParcelTypes, "created_at": a.CreatedAt, "updated_at": a.UpdatedAt,
	}
}

type lastMileAreaInput struct {
	Name                 string          `json:"name" binding:"required,min=2,max=160"`
	State                *string         `json:"state"`
	City                 *string         `json:"city"`
	CoverageStatus       string          `json:"coverage_status" binding:"omitempty,oneof=draft active inactive"`
	PricingModel         string          `json:"pricing_model" binding:"required,oneof=flat zone distance manual"`
	FlatPriceMinor       *int64          `json:"flat_price_minor"`
	BaseDeliveryFeeMinor *int64          `json:"base_delivery_fee_minor"`
	PricePerKmMinor      *int64          `json:"price_per_km_minor"`
	MinimumFeeMinor      *int64          `json:"minimum_fee_minor"`
	MaximumDistanceKm    *float64        `json:"maximum_distance_km"`
	MaximumFeeMinor      *int64          `json:"maximum_fee_minor"`
	VolumetricDivisor    *float64        `json:"volumetric_divisor"`
	PriceIsEstimate      bool            `json:"price_is_estimate"`
	MaxWeightKg          *float64        `json:"max_weight_kg"`
	MaxLengthCm          *float64        `json:"max_length_cm"`
	MaxWidthCm           *float64        `json:"max_width_cm"`
	MaxHeightCm          *float64        `json:"max_height_cm"`
	ExpressAvailable     bool            `json:"express_available"`
	PickupAvailable      bool            `json:"pickup_available"`
	DeliveryAvailable    bool            `json:"delivery_available"`
	SupportedParcelTypes []string        `json:"supported_parcel_types"`
	OperatingHours       json.RawMessage `json:"operating_hours"`
}

func (r lastMileAreaInput) validate() bool {
	for _, minor := range []*int64{r.FlatPriceMinor, r.BaseDeliveryFeeMinor, r.PricePerKmMinor, r.MinimumFeeMinor, r.MaximumFeeMinor} {
		if minor != nil && *minor < 0 {
			return false
		}
	}
	if r.PricingModel == "flat" && r.FlatPriceMinor == nil {
		return false
	}
	if r.PricingModel == "distance" && (r.BaseDeliveryFeeMinor == nil || r.PricePerKmMinor == nil) {
		return false
	}
	if r.PricingModel == "zone" { /* zone matrix is configured separately */
	}
	return true
}

func (a *API) providerOffersLastMile(c *gin.Context) bool {
	tenant, _ := currentTenant(c)
	company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		return false
	}
	businessType := strings.ToLower(company.BusinessType)
	return strings.Contains(businessType, "lastmile") || strings.Contains(businessType, "last mile")
}

func (a *API) listLastMileAreas(c *gin.Context) {
	tenant, _ := currentTenant(c)
	areas, err := a.svc.Q.ListLastMileAreas(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]gin.H, 0, len(areas))
	for _, area := range areas {
		result = append(result, lastMileAreaDTO(area))
	}
	c.JSON(http.StatusOK, gin.H{"areas": result})
}

func (a *API) createLastMileArea(c *gin.Context) {
	tenant, _ := currentTenant(c)
	if !a.providerOffersLastMile(c) {
		fail(c, http.StatusForbidden, "service_not_offered", "This provider does not offer last-mile services.")
		return
	}
	var req lastMileAreaInput
	if err := c.ShouldBindJSON(&req); err != nil || !req.validate() {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "Check the service area and pricing details.")
		return
	}
	status := req.CoverageStatus
	if status == "" {
		status = "draft"
	}
	var hours json.RawMessage = req.OperatingHours
	if len(hours) == 0 {
		hours = []byte(`{}`)
	}
	area, err := a.svc.Q.CreateLastMileArea(c.Request.Context(), db.CreateLastMileAreaParams{
		CompanyID: tenant.CompanyID, Name: strings.TrimSpace(req.Name), State: req.State, City: req.City, CoverageStatus: status, PricingModel: req.PricingModel,
		FlatPriceMinor: nullInt8Minor(req.FlatPriceMinor), BaseDeliveryFeeMinor: nullInt8Minor(req.BaseDeliveryFeeMinor), PricePerKmMinor: nullInt8Minor(req.PricePerKmMinor), MinimumFeeMinor: nullInt8Minor(req.MinimumFeeMinor),
		MaximumDistanceKm: floatNumericPtr(req.MaximumDistanceKm), MaximumFeeMinor: nullInt8Minor(req.MaximumFeeMinor), VolumetricDivisor: floatNumericPtr(req.VolumetricDivisor), PriceIsEstimate: req.PriceIsEstimate,
		MaxWeightKg: floatNumericPtr(req.MaxWeightKg), MaxLengthCm: floatNumericPtr(req.MaxLengthCm), MaxWidthCm: floatNumericPtr(req.MaxWidthCm), MaxHeightCm: floatNumericPtr(req.MaxHeightCm),
		ExpressAvailable: req.ExpressAvailable, PickupAvailable: req.PickupAvailable, DeliveryAvailable: req.DeliveryAvailable, OperatingHours: hours, SupportedParcelTypes: req.SupportedParcelTypes,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"area": lastMileAreaDTO(area)})
}

func nullInt8Minor(v *int64) pgtype.Int8 { return nullableInt8Minor(v) }

func (a *API) getLastMileArea(c *gin.Context) {
	tenant, _ := currentTenant(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	area, err := a.svc.Q.GetLastMileArea(c.Request.Context(), db.GetLastMileAreaParams{CompanyID: tenant.CompanyID, ID: id})
	if err != nil {
		fail(c, http.StatusNotFound, "last_mile_area_not_found", "That service area was not found.")
		return
	}
	zones, err := a.svc.Q.ListServiceZones(c.Request.Context(), db.ListServiceZonesParams{CompanyID: tenant.CompanyID, AreaID: id})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	rates, err := a.svc.Q.ListZoneRates(c.Request.Context(), db.ListZoneRatesParams{CompanyID: tenant.CompanyID, AreaID: id})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"area": lastMileAreaDTO(area), "zones": zones, "zone_rates": rates})
}

func (a *API) updateLastMileArea(c *gin.Context) {
	tenant, _ := currentTenant(c)
	if !a.providerOffersLastMile(c) {
		fail(c, http.StatusForbidden, "service_not_offered", "This provider does not offer last-mile services.")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	var req lastMileAreaInput
	if err := c.ShouldBindJSON(&req); err != nil || !req.validate() {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "Check the service area and pricing details.")
		return
	}
	status := req.CoverageStatus
	if status == "" {
		status = "draft"
	}
	var hours json.RawMessage = req.OperatingHours
	if len(hours) == 0 {
		hours = []byte(`{}`)
	}
	area, err := a.svc.Q.UpdateLastMileArea(c.Request.Context(), db.UpdateLastMileAreaParams{
		CompanyID: tenant.CompanyID, ID: id, Name: strings.TrimSpace(req.Name), State: req.State, City: req.City, CoverageStatus: status, PricingModel: req.PricingModel,
		FlatPriceMinor: nullInt8Minor(req.FlatPriceMinor), BaseDeliveryFeeMinor: nullInt8Minor(req.BaseDeliveryFeeMinor), PricePerKmMinor: nullInt8Minor(req.PricePerKmMinor), MinimumFeeMinor: nullInt8Minor(req.MinimumFeeMinor),
		MaximumDistanceKm: floatNumericPtr(req.MaximumDistanceKm), MaximumFeeMinor: nullInt8Minor(req.MaximumFeeMinor), VolumetricDivisor: floatNumericPtr(req.VolumetricDivisor), PriceIsEstimate: req.PriceIsEstimate,
		MaxWeightKg: floatNumericPtr(req.MaxWeightKg), MaxLengthCm: floatNumericPtr(req.MaxLengthCm), MaxWidthCm: floatNumericPtr(req.MaxWidthCm), MaxHeightCm: floatNumericPtr(req.MaxHeightCm),
		ExpressAvailable: req.ExpressAvailable, PickupAvailable: req.PickupAvailable, DeliveryAvailable: req.DeliveryAvailable, OperatingHours: hours, SupportedParcelTypes: req.SupportedParcelTypes,
	})
	if err != nil {
		fail(c, http.StatusNotFound, "last_mile_area_not_found", "That service area was not found.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"area": lastMileAreaDTO(area)})
}

func (a *API) deleteLastMileArea(c *gin.Context) {
	tenant, _ := currentTenant(c)
	if !a.providerOffersLastMile(c) {
		fail(c, http.StatusForbidden, "service_not_offered", "This provider does not offer last-mile services.")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	if err := a.svc.Q.DeleteLastMileArea(c.Request.Context(), db.DeleteLastMileAreaParams{CompanyID: tenant.CompanyID, ID: id}); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

type serviceZoneInput struct {
	Name           string   `json:"name" binding:"required,min=2,max=160"`
	Neighbourhoods []string `json:"neighbourhoods"`
	State          *string  `json:"state"`
	City           *string  `json:"city"`
	Lga            *string  `json:"lga"`
	CustomLocation *string  `json:"custom_location"`
}

func (a *API) replaceLastMileZones(c *gin.Context) {
	tenant, _ := currentTenant(c)
	if !a.providerOffersLastMile(c) {
		fail(c, http.StatusForbidden, "service_not_offered", "This provider does not offer last-mile services.")
		return
	}
	areaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	var req struct {
		Zones []serviceZoneInput `json:"zones"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_service_zones", "Check the zone details.")
		return
	}
	if len(req.Zones) == 0 {
		fail(c, http.StatusBadRequest, "invalid_service_zones", "Add at least one zone.")
		return
	}
	seen := map[string]bool{}
	for _, z := range req.Zones {
		name := strings.ToLower(strings.TrimSpace(z.Name))
		if name == "" || seen[name] {
			fail(c, http.StatusBadRequest, "invalid_service_zones", "Zone names must be unique.")
			return
		}
		seen[name] = true
	}
	for _, old := range func() []db.ServiceZone {
		z, _ := a.svc.Q.ListServiceZones(c.Request.Context(), db.ListServiceZonesParams{CompanyID: tenant.CompanyID, AreaID: areaID})
		return z
	}() {
		_ = old
	}
	// Delete existing zones safely through an area-owned query, then recreate.
	_, err = a.svc.DB.Exec(c.Request.Context(), `DELETE FROM service_zones z USING last_mile_service_areas a WHERE z.area_id=a.id AND z.area_id=$1 AND a.company_id=$2`, areaID, tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	zones := make([]db.ServiceZone, 0, len(req.Zones))
	for _, z := range req.Zones {
		created, err := a.svc.Q.CreateServiceZone(c.Request.Context(), db.CreateServiceZoneParams{AreaID: areaID, Name: strings.TrimSpace(z.Name), Neighbourhoods: z.Neighbourhoods, State: z.State, City: z.City, Lga: z.Lga, CustomLocation: z.CustomLocation, CompanyID: tenant.CompanyID})
		if err != nil {
			a.handleServiceError(c, err)
			return
		}
		zones = append(zones, created)
	}
	c.JSON(http.StatusOK, gin.H{"zones": zones})
}

type zoneRateInput struct {
	FromZoneID string `json:"from_zone_id"`
	ToZoneID   string `json:"to_zone_id"`
	PriceMinor int64  `json:"price_minor"`
}

func (a *API) replaceZoneRates(c *gin.Context) {
	tenant, _ := currentTenant(c)
	if !a.providerOffersLastMile(c) {
		fail(c, http.StatusForbidden, "service_not_offered", "This provider does not offer last-mile services.")
		return
	}
	areaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	var req struct {
		Rates []zoneRateInput `json:"rates"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_rates", "Check the zone pricing matrix.")
		return
	}
	if err := a.svc.Q.ReplacePricingZoneRates(c.Request.Context(), db.ReplacePricingZoneRatesParams{CompanyID: tenant.CompanyID, AreaID: areaID}); err != nil {
		a.handleServiceError(c, err)
		return
	}
	seen := map[string]bool{}
	result := make([]db.PricingZoneRate, 0, len(req.Rates))
	for _, rate := range req.Rates {
		if rate.PriceMinor < 0 {
			fail(c, http.StatusBadRequest, "invalid_zone_rates", "Prices cannot be negative.")
			return
		}
		from, err := uuid.Parse(rate.FromZoneID)
		if err != nil {
			fail(c, http.StatusBadRequest, "invalid_zone_rates", "Choose valid zones.")
			return
		}
		to, err := uuid.Parse(rate.ToZoneID)
		if err != nil {
			fail(c, http.StatusBadRequest, "invalid_zone_rates", "Choose valid zones.")
			return
		}
		key := from.String() + ":" + to.String()
		if seen[key] {
			fail(c, http.StatusBadRequest, "invalid_zone_rates", "Duplicate direction in the matrix.")
			return
		}
		seen[key] = true
		created, err := a.svc.Q.CreatePricingZoneRate(c.Request.Context(), db.CreatePricingZoneRateParams{AreaID: areaID, FromZoneID: from, ToZoneID: to, PriceMinor: rate.PriceMinor, CompanyID: tenant.CompanyID})
		if err != nil {
			a.handleServiceError(c, err)
			return
		}
		result = append(result, created)
	}
	c.JSON(http.StatusOK, gin.H{"zone_rates": result})
}

func (a *API) previewLastMilePrice(c *gin.Context) {
	tenant, _ := currentTenant(c)
	areaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	area, err := a.svc.Q.GetLastMileArea(c.Request.Context(), db.GetLastMileAreaParams{CompanyID: tenant.CompanyID, ID: areaID})
	if err != nil {
		fail(c, http.StatusNotFound, "last_mile_area_not_found", "That service area was not found.")
		return
	}
	var req struct {
		WeightKg   *float64 `json:"weight_kg"`
		LengthCm   *float64 `json:"length_cm"`
		WidthCm    *float64 `json:"width_cm"`
		HeightCm   *float64 `json:"height_cm"`
		DistanceKm *float64 `json:"distance_km"`
		FromZoneID *string  `json:"from_zone_id"`
		ToZoneID   *string  `json:"to_zone_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_pricing_preview", "Check the parcel details.")
		return
	}
	var base *int64
	switch area.PricingModel {
	case "manual":
		c.JSON(http.StatusOK, gin.H{"confirmation_required": true, "physical_inspection_required": true, "base_price_minor": nil, "additional_charges_minor": 0, "estimated_total_minor": nil, "currency": "NGN", "pricing_model": "manual", "is_estimate": true})
		return
	case "flat":
		base = int8Value(area.FlatPriceMinor)
	case "distance":
		base = int8Value(area.BaseDeliveryFeeMinor)
		per := int8Value(area.PricePerKmMinor)
		if base == nil || per == nil || req.DistanceKm == nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": gin.H{"code": "missing_pricing_configuration", "message": "This area needs a base fee and per-kilometre rate."},
			})
			return
		}
		calculated := *base + int64(*req.DistanceKm*float64(*per))
		if area.MinimumFeeMinor.Valid && calculated < area.MinimumFeeMinor.Int64 {
			calculated = area.MinimumFeeMinor.Int64
		}
		if area.MaximumFeeMinor.Valid && calculated > area.MaximumFeeMinor.Int64 {
			calculated = area.MaximumFeeMinor.Int64
		}
		base = &calculated
	case "zone":
		if req.FromZoneID == nil || req.ToZoneID == nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": gin.H{"code": "missing_zone_selection", "message": "Choose the origin and destination zones."},
			})
			return
		}
		from, err := uuid.Parse(*req.FromZoneID)
		if err != nil {
			base = nil
		} else {
			to, _ := uuid.Parse(*req.ToZoneID)
			for _, r := range func() []db.PricingZoneRate {
				x, _ := a.svc.Q.ListZoneRates(c.Request.Context(), db.ListZoneRatesParams{CompanyID: tenant.CompanyID, AreaID: areaID})
				return x
			}() {
				if r.FromZoneID == from && r.ToZoneID == to {
					value := r.PriceMinor
					base = &value
					break
				}
			}
		}
	}
	if base == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": gin.H{"code": "missing_pricing_configuration", "message": "No valid price is configured for this service area."},
		})
		return
	}
	if area.MaxWeightKg.Valid && req.WeightKg != nil && *req.WeightKg > numericToFloat64(area.MaxWeightKg) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": gin.H{"code": "parcel_restriction", "message": "Parcel exceeds the maximum weight."},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"base_price_minor": *base, "additional_charges_minor": 0, "estimated_total_minor": *base, "currency": "NGN", "pricing_model": area.PricingModel, "is_estimate": area.PriceIsEstimate, "confirmation_required": false, "physical_inspection_required": false})
}
