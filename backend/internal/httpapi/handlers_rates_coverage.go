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

func zoneChain(id uuid.UUID, parents map[uuid.UUID]*uuid.UUID) []uuid.UUID {
	chain := []uuid.UUID{id}
	seen := map[uuid.UUID]bool{id: true}
	for {
		parent, ok := parents[chain[len(chain)-1]]
		if !ok || parent == nil || seen[*parent] {
			return chain
		}
		seen[*parent] = true
		chain = append(chain, *parent)
	}
}

// lookupZoneRate checks the exact direction first, then falls back through a
// sub-zone's parent for whichever end has no explicit price. Routes stay
// directional: the reverse direction is never used as a fallback.
func lookupZoneRate(from, to uuid.UUID, parents map[uuid.UUID]*uuid.UUID, rates map[string]int64) (int64, bool) {
	fromChain := zoneChain(from, parents)
	toChain := zoneChain(to, parents)
	for total := 0; total <= len(fromChain)+len(toChain); total++ {
		for i := 0; i < len(fromChain); i++ {
			j := total - i
			if j < 0 || j >= len(toChain) {
				continue
			}
			if price, ok := rates[fromChain[i].String()+":"+toChain[j].String()]; ok {
				return price, true
			}
		}
	}
	return 0, false
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
		ExpressAvailable: req.ExpressAvailable, PickupAvailable: req.PickupAvailable, DeliveryAvailable: req.DeliveryAvailable, OperatingHours: hours, SupportedParcelTypes: nonNilStrings(req.SupportedParcelTypes),
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
	zones, err := a.svc.Q.ListServiceZones(c.Request.Context(), db.ListServiceZonesParams{CompanyID: tenant.CompanyID, AreaID: &id})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	rates, err := a.svc.Q.ListZoneRates(c.Request.Context(), db.ListZoneRatesParams{CompanyID: tenant.CompanyID, AreaID: &id})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"area": lastMileAreaDTO(area), "zones": zones, "zone_rates": rates})
}

func (a *API) updateLastMileArea(c *gin.Context) {
	tenant, _ := currentTenant(c)
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
		ExpressAvailable: req.ExpressAvailable, PickupAvailable: req.PickupAvailable, DeliveryAvailable: req.DeliveryAvailable, OperatingHours: hours, SupportedParcelTypes: nonNilStrings(req.SupportedParcelTypes),
	})
	if err != nil {
		fail(c, http.StatusNotFound, "last_mile_area_not_found", "That service area was not found.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"area": lastMileAreaDTO(area)})
}

func (a *API) deleteLastMileArea(c *gin.Context) {
	tenant, _ := currentTenant(c)
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
	ID             *string  `json:"id"`
	Name           string   `json:"name" binding:"required,min=2,max=160"`
	Neighbourhoods []string `json:"neighbourhoods"`
	State          *string  `json:"state"`
	City           *string  `json:"city"`
	Lga            *string  `json:"lga"`
	CustomLocation *string  `json:"custom_location"`
	ParentZoneID   *string  `json:"parent_zone_id"`
}

type zoneUpsert struct {
	id             uuid.UUID
	isNew          bool
	name           string
	neighbourhoods []string
	state          *string
	city           *string
	lga            *string
	customLocation *string
	parentID       *uuid.UUID
}

// replaceLastMileZones saves the zone list for an area. Existing zones are
// updated in place so their pricing routes stay attached; omitted zones are
// deleted, which also removes any routes that referenced them.
func (a *API) replaceLastMileZones(c *gin.Context) {
	tenant, _ := currentTenant(c)
	areaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	if _, err := a.svc.Q.GetLastMileArea(c.Request.Context(), db.GetLastMileAreaParams{CompanyID: tenant.CompanyID, ID: areaID}); err != nil {
		fail(c, http.StatusNotFound, "last_mile_area_not_found", "That service area was not found.")
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
	existingRows, err := a.svc.Q.ListServiceZones(c.Request.Context(), db.ListServiceZonesParams{CompanyID: tenant.CompanyID, AreaID: &areaID})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	existing := make(map[uuid.UUID]bool, len(existingRows))
	for _, zone := range existingRows {
		existing[zone.ID] = true
	}

	seenNames := map[string]bool{}
	kept := map[uuid.UUID]bool{}
	upserts := make([]zoneUpsert, 0, len(req.Zones))
	for _, zone := range req.Zones {
		name := strings.TrimSpace(zone.Name)
		key := strings.ToLower(name)
		if key == "" || seenNames[key] {
			fail(c, http.StatusBadRequest, "invalid_service_zones", "Zone names must be unique.")
			return
		}
		seenNames[key] = true

		upsert := zoneUpsert{
			isNew:          true,
			name:           name,
			neighbourhoods: nonNilStrings(zone.Neighbourhoods),
			state:          zone.State,
			city:           zone.City,
			lga:            zone.Lga,
			customLocation: zone.CustomLocation,
		}
		if zone.ID != nil && strings.TrimSpace(*zone.ID) != "" {
			id, err := uuid.Parse(strings.TrimSpace(*zone.ID))
			if err != nil || !existing[id] {
				fail(c, http.StatusBadRequest, "invalid_service_zones", "That zone does not belong to this service area.")
				return
			}
			upsert.id = id
			upsert.isNew = false
			kept[id] = true
		}
		if zone.ParentZoneID != nil && strings.TrimSpace(*zone.ParentZoneID) != "" {
			parentID, err := uuid.Parse(strings.TrimSpace(*zone.ParentZoneID))
			if err != nil || !existing[parentID] {
				fail(c, http.StatusBadRequest, "invalid_service_zones", "Choose a parent zone already saved in this area.")
				return
			}
			upsert.parentID = &parentID
		}
		upserts = append(upserts, upsert)
	}

	parentOf := map[uuid.UUID]*uuid.UUID{}
	for _, zone := range upserts {
		if zone.parentID != nil && (!kept[*zone.parentID] || (!zone.isNew && *zone.parentID == zone.id)) {
			fail(c, http.StatusBadRequest, "invalid_service_zones", "A sub-zone must have a different parent zone that is kept in this area.")
			return
		}
		if !zone.isNew {
			parentOf[zone.id] = zone.parentID
		}
	}
	// Reject parent chains that loop back on themselves.
	for id := range parentOf {
		seen := map[uuid.UUID]bool{id: true}
		for current := id; parentOf[current] != nil; {
			current = *parentOf[current]
			if seen[current] {
				fail(c, http.StatusBadRequest, "invalid_service_zones", "A zone cannot be its own parent or form a loop.")
				return
			}
			seen[current] = true
		}
	}

	tx, err := a.svc.DB.Begin(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()
	qtx := a.svc.Q.WithTx(tx)

	// Remove omitted zones first so a zone can be deleted and recreated with
	// the same name in one save. The transaction rolls back on any failure.
	for id := range existing {
		if kept[id] {
			continue
		}
		if err := qtx.DeleteServiceZone(c.Request.Context(), db.DeleteServiceZoneParams{CompanyID: tenant.CompanyID, ID: id}); err != nil {
			a.handleServiceError(c, err)
			return
		}
	}
	for _, zone := range upserts {
		if zone.isNew {
			_, err = qtx.CreateServiceZone(c.Request.Context(), db.CreateServiceZoneParams{
				AreaID: &areaID, Name: zone.name, Neighbourhoods: zone.neighbourhoods, State: zone.state,
				City: zone.city, Lga: zone.lga, CustomLocation: zone.customLocation, ParentZoneID: zone.parentID, CompanyID: tenant.CompanyID,
			})
		} else {
			_, err = qtx.UpdateServiceZone(c.Request.Context(), db.UpdateServiceZoneParams{
				CompanyID: tenant.CompanyID, AreaID: &areaID, ID: zone.id, Name: zone.name, Neighbourhoods: zone.neighbourhoods,
				State: zone.state, City: zone.city, Lga: zone.lga, CustomLocation: zone.customLocation, ParentZoneID: zone.parentID,
			})
		}
		if err != nil {
			a.handleServiceError(c, err)
			return
		}
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		a.handleServiceError(c, err)
		return
	}
	zones, err := a.svc.Q.ListServiceZones(c.Request.Context(), db.ListServiceZonesParams{CompanyID: tenant.CompanyID, AreaID: &areaID})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"zones": zones})
}

func (a *API) serviceZoneParent(c *gin.Context, companyID, areaID uuid.UUID, raw *string, self uuid.UUID) (*uuid.UUID, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, true
	}
	parentID, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil || (self != uuid.Nil && parentID == self) {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "Choose a different parent zone from this area.")
		return nil, false
	}
	if _, err := a.svc.Q.GetServiceZone(c.Request.Context(), db.GetServiceZoneParams{CompanyID: companyID, AreaID: &areaID, ID: parentID}); err != nil {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "Choose a parent zone from this area.")
		return nil, false
	}
	return &parentID, true
}

func zoneParentCreatesCycle(zones []db.ServiceZone, zoneID uuid.UUID, parentID *uuid.UUID) bool {
	if parentID == nil {
		return false
	}
	parents := make(map[uuid.UUID]*uuid.UUID, len(zones))
	for _, zone := range zones {
		parents[zone.ID] = zone.ParentZoneID
	}
	seen := map[uuid.UUID]bool{zoneID: true}
	current := *parentID
	for {
		if current == zoneID {
			return true
		}
		if seen[current] {
			return true
		}
		seen[current] = true
		next, ok := parents[current]
		if !ok || next == nil {
			return false
		}
		current = *next
	}
}

func (a *API) createLastMileZone(c *gin.Context) {
	tenant, _ := currentTenant(c)
	areaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	if _, err := a.svc.Q.GetLastMileArea(c.Request.Context(), db.GetLastMileAreaParams{CompanyID: tenant.CompanyID, ID: areaID}); err != nil {
		fail(c, http.StatusNotFound, "last_mile_area_not_found", "That service area was not found.")
		return
	}
	var req serviceZoneInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "Check the zone details.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if len(name) < 2 {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "Enter a zone name with at least two characters.")
		return
	}
	parentID, ok := a.serviceZoneParent(c, tenant.CompanyID, areaID, req.ParentZoneID, uuid.Nil)
	if !ok {
		return
	}
	zone, err := a.svc.Q.CreateServiceZone(c.Request.Context(), db.CreateServiceZoneParams{
		AreaID: &areaID, Name: name, Neighbourhoods: nonNilStrings(req.Neighbourhoods), State: req.State,
		City: req.City, Lga: req.Lga, CustomLocation: req.CustomLocation, ParentZoneID: parentID, CompanyID: tenant.CompanyID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"zone": zone})
}

func (a *API) updateLastMileZone(c *gin.Context) {
	tenant, _ := currentTenant(c)
	areaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	zoneID, err := uuid.Parse(c.Param("zoneId"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "That zone id is invalid.")
		return
	}
	if _, err := a.svc.Q.GetServiceZone(c.Request.Context(), db.GetServiceZoneParams{CompanyID: tenant.CompanyID, AreaID: &areaID, ID: zoneID}); err != nil {
		fail(c, http.StatusNotFound, "service_zone_not_found", "That zone was not found.")
		return
	}
	var req serviceZoneInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "Check the zone details.")
		return
	}
	name := strings.TrimSpace(req.Name)
	if len(name) < 2 {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "Enter a zone name with at least two characters.")
		return
	}
	parentID, ok := a.serviceZoneParent(c, tenant.CompanyID, areaID, req.ParentZoneID, zoneID)
	if !ok {
		return
	}
	if parentID != nil {
		zones, err := a.svc.Q.ListServiceZones(c.Request.Context(), db.ListServiceZonesParams{CompanyID: tenant.CompanyID, AreaID: &areaID})
		if err != nil {
			a.handleServiceError(c, err)
			return
		}
		if zoneParentCreatesCycle(zones, zoneID, parentID) {
			fail(c, http.StatusBadRequest, "invalid_service_zone", "A zone cannot be its own parent or form a loop.")
			return
		}
	}
	zone, err := a.svc.Q.UpdateServiceZone(c.Request.Context(), db.UpdateServiceZoneParams{
		CompanyID: tenant.CompanyID, AreaID: &areaID, ID: zoneID, Name: name, Neighbourhoods: nonNilStrings(req.Neighbourhoods),
		State: req.State, City: req.City, Lga: req.Lga, CustomLocation: req.CustomLocation, ParentZoneID: parentID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"zone": zone})
}

func (a *API) deleteLastMileZone(c *gin.Context) {
	tenant, _ := currentTenant(c)
	areaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	zoneID, err := uuid.Parse(c.Param("zoneId"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_service_zone", "That zone id is invalid.")
		return
	}
	if _, err := a.svc.Q.GetServiceZone(c.Request.Context(), db.GetServiceZoneParams{CompanyID: tenant.CompanyID, AreaID: &areaID, ID: zoneID}); err != nil {
		fail(c, http.StatusNotFound, "service_zone_not_found", "That zone was not found.")
		return
	}
	if err := a.svc.Q.DeleteServiceZone(c.Request.Context(), db.DeleteServiceZoneParams{CompanyID: tenant.CompanyID, ID: zoneID}); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

type zoneRateInput struct {
	FromZoneID string `json:"from_zone_id"`
	ToZoneID   string `json:"to_zone_id"`
	PriceMinor int64  `json:"price_minor"`
}

func (a *API) replaceZoneRates(c *gin.Context) {
	tenant, _ := currentTenant(c)
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
	if _, err := a.svc.Q.GetLastMileArea(c.Request.Context(), db.GetLastMileAreaParams{CompanyID: tenant.CompanyID, ID: areaID}); err != nil {
		fail(c, http.StatusNotFound, "last_mile_area_not_found", "That service area was not found.")
		return
	}
	zones, err := a.svc.Q.ListServiceZones(c.Request.Context(), db.ListServiceZonesParams{CompanyID: tenant.CompanyID, AreaID: &areaID})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	zoneSet := make(map[uuid.UUID]bool, len(zones))
	for _, zone := range zones {
		zoneSet[zone.ID] = true
	}

	// Validate the whole matrix before deleting the current one so a bad row
	// cannot wipe a provider's configured routes.
	type parsedRate struct {
		from  uuid.UUID
		to    uuid.UUID
		price int64
	}
	seen := map[string]bool{}
	parsed := make([]parsedRate, 0, len(req.Rates))
	for _, rate := range req.Rates {
		if rate.PriceMinor < 0 {
			fail(c, http.StatusBadRequest, "invalid_zone_rates", "Prices cannot be negative.")
			return
		}
		from, err := uuid.Parse(strings.TrimSpace(rate.FromZoneID))
		if err != nil || !zoneSet[from] {
			fail(c, http.StatusBadRequest, "invalid_zone_rates", "Choose a pickup zone in this service area.")
			return
		}
		to, err := uuid.Parse(strings.TrimSpace(rate.ToZoneID))
		if err != nil || !zoneSet[to] {
			fail(c, http.StatusBadRequest, "invalid_zone_rates", "Choose a delivery zone in this service area.")
			return
		}
		key := from.String() + ":" + to.String()
		if seen[key] {
			fail(c, http.StatusBadRequest, "invalid_zone_rates", "Each direction can only be priced once.")
			return
		}
		seen[key] = true
		parsed = append(parsed, parsedRate{from: from, to: to, price: rate.PriceMinor})
	}

	tx, err := a.svc.DB.Begin(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()
	qtx := a.svc.Q.WithTx(tx)
	if err := qtx.ReplacePricingZoneRates(c.Request.Context(), db.ReplacePricingZoneRatesParams{CompanyID: tenant.CompanyID, AreaID: &areaID}); err != nil {
		a.handleServiceError(c, err)
		return
	}
	for _, rate := range parsed {
		if _, err := qtx.CreatePricingZoneRate(c.Request.Context(), db.CreatePricingZoneRateParams{AreaID: &areaID, FromZoneID: rate.from, ToZoneID: rate.to, PriceMinor: rate.price, CompanyID: tenant.CompanyID}); err != nil {
			a.handleServiceError(c, err)
			return
		}
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		a.handleServiceError(c, err)
		return
	}
	result, err := a.svc.Q.ListZoneRates(c.Request.Context(), db.ListZoneRatesParams{CompanyID: tenant.CompanyID, AreaID: &areaID})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"zone_rates": result})
}

func (a *API) zoneRateZones(c *gin.Context, companyID, areaID uuid.UUID, fromRaw, toRaw string) (uuid.UUID, uuid.UUID, bool) {
	from, fromErr := uuid.Parse(strings.TrimSpace(fromRaw))
	to, toErr := uuid.Parse(strings.TrimSpace(toRaw))
	if fromErr != nil || toErr != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_rate", "Choose the pickup and delivery zones.")
		return uuid.Nil, uuid.Nil, false
	}
	zones, err := a.svc.Q.ListServiceZones(c.Request.Context(), db.ListServiceZonesParams{CompanyID: companyID, AreaID: &areaID})
	if err != nil {
		a.handleServiceError(c, err)
		return uuid.Nil, uuid.Nil, false
	}
	known := make(map[uuid.UUID]bool, len(zones))
	for _, zone := range zones {
		known[zone.ID] = true
	}
	if !known[from] {
		fail(c, http.StatusBadRequest, "invalid_zone_rate", "Choose a pickup zone in this service area.")
		return uuid.Nil, uuid.Nil, false
	}
	if !known[to] {
		fail(c, http.StatusBadRequest, "invalid_zone_rate", "Choose a delivery zone in this service area.")
		return uuid.Nil, uuid.Nil, false
	}
	return from, to, true
}

func (a *API) zoneRateExists(c *gin.Context, companyID, areaID, from, to, exclude uuid.UUID) (bool, bool) {
	rates, err := a.svc.Q.ListZoneRates(c.Request.Context(), db.ListZoneRatesParams{CompanyID: companyID, AreaID: &areaID})
	if err != nil {
		a.handleServiceError(c, err)
		return false, false
	}
	for _, rate := range rates {
		if rate.ID == exclude {
			continue
		}
		if rate.FromZoneID == from && rate.ToZoneID == to {
			return true, true
		}
	}
	return false, true
}

func (a *API) createZoneRate(c *gin.Context) {
	tenant, _ := currentTenant(c)
	areaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	if _, err := a.svc.Q.GetLastMileArea(c.Request.Context(), db.GetLastMileAreaParams{CompanyID: tenant.CompanyID, ID: areaID}); err != nil {
		fail(c, http.StatusNotFound, "last_mile_area_not_found", "That service area was not found.")
		return
	}
	var req zoneRateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_rate", "Check the route details.")
		return
	}
	if req.PriceMinor < 0 {
		fail(c, http.StatusBadRequest, "invalid_zone_rate", "The base price cannot be negative.")
		return
	}
	from, to, ok := a.zoneRateZones(c, tenant.CompanyID, areaID, req.FromZoneID, req.ToZoneID)
	if !ok {
		return
	}
	exists, ok := a.zoneRateExists(c, tenant.CompanyID, areaID, from, to, uuid.Nil)
	if !ok {
		return
	}
	if exists {
		fail(c, http.StatusConflict, "zone_rate_exists", "A route for this direction already exists. Edit that route instead.")
		return
	}
	rate, err := a.svc.Q.CreatePricingZoneRate(c.Request.Context(), db.CreatePricingZoneRateParams{AreaID: &areaID, FromZoneID: from, ToZoneID: to, PriceMinor: req.PriceMinor, CompanyID: tenant.CompanyID})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"zone_rate": rate})
}

func (a *API) updateZoneRate(c *gin.Context) {
	tenant, _ := currentTenant(c)
	areaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	rateID, err := uuid.Parse(c.Param("rateId"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_rate", "That route id is invalid.")
		return
	}
	if _, err := a.svc.Q.GetPricingZoneRate(c.Request.Context(), db.GetPricingZoneRateParams{CompanyID: tenant.CompanyID, AreaID: &areaID, ID: rateID}); err != nil {
		fail(c, http.StatusNotFound, "zone_rate_not_found", "That route was not found.")
		return
	}
	var req zoneRateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_rate", "Check the route details.")
		return
	}
	if req.PriceMinor < 0 {
		fail(c, http.StatusBadRequest, "invalid_zone_rate", "The base price cannot be negative.")
		return
	}
	from, to, ok := a.zoneRateZones(c, tenant.CompanyID, areaID, req.FromZoneID, req.ToZoneID)
	if !ok {
		return
	}
	exists, ok := a.zoneRateExists(c, tenant.CompanyID, areaID, from, to, rateID)
	if !ok {
		return
	}
	if exists {
		fail(c, http.StatusConflict, "zone_rate_exists", "A route for this direction already exists. Edit that route instead.")
		return
	}
	rate, err := a.svc.Q.UpdatePricingZoneRate(c.Request.Context(), db.UpdatePricingZoneRateParams{CompanyID: tenant.CompanyID, AreaID: &areaID, ID: rateID, FromZoneID: from, ToZoneID: to, PriceMinor: req.PriceMinor})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"zone_rate": rate})
}

func (a *API) deleteZoneRate(c *gin.Context) {
	tenant, _ := currentTenant(c)
	areaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_last_mile_area", "That service area id is invalid.")
		return
	}
	rateID, err := uuid.Parse(c.Param("rateId"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_zone_rate", "That route id is invalid.")
		return
	}
	if _, err := a.svc.Q.GetPricingZoneRate(c.Request.Context(), db.GetPricingZoneRateParams{CompanyID: tenant.CompanyID, AreaID: &areaID, ID: rateID}); err != nil {
		fail(c, http.StatusNotFound, "zone_rate_not_found", "That route was not found.")
		return
	}
	if err := a.svc.Q.DeletePricingZoneRate(c.Request.Context(), db.DeletePricingZoneRateParams{CompanyID: tenant.CompanyID, AreaID: &areaID, ID: rateID}); err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
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
		from, fromErr := uuid.Parse(strings.TrimSpace(*req.FromZoneID))
		to, toErr := uuid.Parse(strings.TrimSpace(*req.ToZoneID))
		if fromErr == nil && toErr == nil {
			zones, err := a.svc.Q.ListServiceZones(c.Request.Context(), db.ListServiceZonesParams{CompanyID: tenant.CompanyID, AreaID: &areaID})
			if err != nil {
				a.handleServiceError(c, err)
				return
			}
			rates, err := a.svc.Q.ListZoneRates(c.Request.Context(), db.ListZoneRatesParams{CompanyID: tenant.CompanyID, AreaID: &areaID})
			if err != nil {
				a.handleServiceError(c, err)
				return
			}
			parents := make(map[uuid.UUID]*uuid.UUID, len(zones))
			known := make(map[uuid.UUID]bool, len(zones))
			for _, zone := range zones {
				parents[zone.ID] = zone.ParentZoneID
				known[zone.ID] = true
			}
			if known[from] && known[to] {
				byDirection := make(map[string]int64, len(rates))
				for _, rate := range rates {
					byDirection[rate.FromZoneID.String()+":"+rate.ToZoneID.String()] = rate.PriceMinor
				}
				if price, ok := lookupZoneRate(from, to, parents, byDirection); ok {
					base = &price
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
