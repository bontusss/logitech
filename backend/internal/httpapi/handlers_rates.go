package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/bol-platform/backend/internal/db/sqlc"
)

type countryDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	NumericID int       `json:"numeric_code"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toCountryDTO(country db.Country) countryDTO {
	return countryDTO{
		ID:        country.ID,
		Name:      country.Name,
		Code:      country.Code,
		NumericID: int(country.NumericCode),
		Status:    string(country.Status),
		CreatedAt: country.CreatedAt,
		UpdatedAt: country.UpdatedAt,
	}
}

func optionalString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func nullableTimeString(v *time.Time) *string {
	if v == nil {
		return nil
	}
	formatted := v.Format(time.RFC3339)
	return &formatted
}

func numericToFloat64(v pgtype.Numeric) float64 {
	if !v.Valid {
		return 0
	}
	f, err := v.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

func float64ToNumeric(v float64) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(strconv.FormatFloat(v, 'f', -1, 64)); err != nil {
		return pgtype.Numeric{}
	}
	return n
}

type stateDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CountryID uuid.UUID `json:"country_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toStateDTO(state db.State) stateDTO {
	code := ""
	if state.Code != nil {
		code = *state.Code
	}
	return stateDTO{
		ID:        state.ID,
		Name:      state.Name,
		Code:      code,
		CountryID: state.CountryID,
		Status:    string(state.Status),
		CreatedAt: state.CreatedAt,
		UpdatedAt: state.UpdatedAt,
	}
}

type cityDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	StateID   uuid.UUID `json:"state_id"`
	CountryID uuid.UUID `json:"country_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toCityDTO(city db.City) cityDTO {
	stateID := uuid.Nil
	if city.StateID != nil {
		stateID = *city.StateID
	}
	return cityDTO{
		ID:        city.ID,
		Name:      city.Name,
		StateID:   stateID,
		CountryID: city.CountryID,
		Status:    string(city.Status),
		CreatedAt: city.CreatedAt,
		UpdatedAt: city.UpdatedAt,
	}
}

type terminalDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CityID    uuid.UUID `json:"city_id"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toTerminalDTO(terminal db.Terminal) terminalDTO {
	cityID := uuid.Nil
	if terminal.CityID != nil {
		cityID = *terminal.CityID
	}
	address := ""
	if terminal.Address != nil {
		address = *terminal.Address
	}
	return terminalDTO{
		ID:        terminal.ID,
		Name:      terminal.Name,
		Code:      terminal.Code,
		CityID:    cityID,
		Address:   address,
		Status:    string(terminal.Status),
		CreatedAt: terminal.CreatedAt,
		UpdatedAt: terminal.UpdatedAt,
	}
}

// shippingServiceDTO for the shipping services catalogue
type shippingServiceDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ServiceType string    `json:"service_type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toShippingServiceDTO(service db.ShippingService) shippingServiceDTO {
	return shippingServiceDTO{
		ID:          service.ID,
		Name:        service.Name,
		Description: optionalString(service.Description),
		ServiceType: string(service.DeliveryType),
		Status:      string(service.Status),
		CreatedAt:   service.CreatedAt,
		UpdatedAt:   service.UpdatedAt,
	}
}

// companyRouteDTO for routes between locations
type companyRouteDTO struct {
	ID                      uuid.UUID `json:"id"`
	CompanyID               uuid.UUID `json:"company_id"`
	OriginLocationType      string    `json:"origin_location_type"`
	OriginLocationID        uuid.UUID `json:"origin_location_id"`
	DestinationLocationType string    `json:"destination_location_type"`
	DestinationLocationID   uuid.UUID `json:"destination_location_id"`
	IsDirectional           bool      `json:"is_directionality"`
	Status                  string    `json:"status"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

func toCompanyRouteDTO(route db.CompanyRoute) companyRouteDTO {
	return companyRouteDTO{
		ID:                      route.ID,
		CompanyID:               route.CompanyID,
		OriginLocationType:      string(route.OriginLocationType),
		OriginLocationID:        route.OriginLocationID,
		DestinationLocationType: string(route.DestinationLocationType),
		DestinationLocationID:   route.DestinationLocationID,
		IsDirectional:           route.IsRoundTrip,
		Status:                  string(route.Status),
		CreatedAt:               route.CreatedAt,
		UpdatedAt:               route.UpdatedAt,
	}
}

// rateCardDTO for rate cards with version info
type rateCardDTO struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	IsPublished  bool      `json:"is_published"`
	VersionCount int       `json:"version_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toRateCardDTO(card db.RateCard) rateCardDTO {
	status := "draft"
	if card.IsPublished {
		status = "published"
	}

	return rateCardDTO{
		ID:           card.ID,
		Name:         card.Name,
		Description:  optionalString(card.Description),
		Status:       status,
		IsPublished:  card.IsPublished,
		VersionCount: 0,
		CreatedAt:    card.CreatedAt,
		UpdatedAt:    card.UpdatedAt,
	}
}

type rateCardVersionDTO struct {
	ID          uuid.UUID `json:"id"`
	RateCardID  uuid.UUID `json:"rate_card_id"`
	Version     int       `json:"version"`
	IsActive    bool      `json:"is_active"`
	Status      string    `json:"status"`
	PublishedAt *string   `json:"published_at,omitempty"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toRateCardVersionDTO(version db.RateCardVersion) rateCardVersionDTO {
	status := "draft"
	if version.IsActive {
		status = "active"
	}

	return rateCardVersionDTO{
		ID:         version.ID,
		RateCardID: version.RateCardID,
		Version:    int(version.Version),
		IsActive:   version.IsActive,
		Status:     status,
		Notes:      "",
		CreatedAt:  version.CreatedAt,
		UpdatedAt:  version.CreatedAt,
	}
}

// pricingRuleDTO for pricing rules
type pricingRuleDTO struct {
	ID                uuid.UUID `json:"id"`
	RateCardVersionID uuid.UUID `json:"rate_card_version_id"`
	RuleType          string    `json:"rule_type"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	IsActive          bool      `json:"is_active"`
	EffectiveFrom     time.Time `json:"effective_from"`
	EffectiveUntil    *string   `json:"effective_until,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func toPricingRuleDTO(rule db.PricingRule) pricingRuleDTO {
	var effectiveUntil *string
	if rule.EffectiveUntil != nil {
		formatted := rule.EffectiveUntil.Format(time.RFC3339)
		effectiveUntil = &formatted
	}

	return pricingRuleDTO{
		ID:                rule.ID,
		RateCardVersionID: rule.RateCardVersionID,
		RuleType:          string(rule.RuleType),
		Name:              rule.Name,
		Description:       optionalString(rule.Description),
		IsActive:          rule.IsActive,
		EffectiveFrom:     rule.EffectiveFrom,
		EffectiveUntil:    effectiveUntil,
		CreatedAt:         rule.CreatedAt,
		UpdatedAt:         rule.CreatedAt,
	}
}

// pricingTierDTO for weight brackets
type pricingTierDTO struct {
	ID                uuid.UUID `json:"id"`
	RateCardVersionID uuid.UUID `json:"rate_card_version_id"`
	MinWeightKG       float64   `json:"min_weight_kg"`
	MaxWeightKG       *float64  `json:"max_weight_kg,omitempty"`
	PriceCents        int       `json:"price_cents"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func toPricingTierDTO(tier db.PricingTier) pricingTierDTO {
	var maxWeight *float64
	if tier.MaxWeightKg.Valid {
		value := numericToFloat64(tier.MaxWeightKg)
		maxWeight = &value
	}

	return pricingTierDTO{
		ID:                tier.ID,
		RateCardVersionID: tier.RateCardVersionID,
		MinWeightKG:       numericToFloat64(tier.MinWeightKg),
		MaxWeightKG:       maxWeight,
		PriceCents:        int(tier.PriceCents),
		Name:              tier.Name,
		Description:       optionalString(tier.Description),
		CreatedAt:         tier.CreatedAt,
		UpdatedAt:         tier.CreatedAt,
	}
}

// packageCategoryDTO for package categories
type packageCategoryDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toPackageCategoryDTO(cat db.PackageCategory) packageCategoryDTO {
	return packageCategoryDTO{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: optionalString(cat.Description),
		SortOrder:   int(cat.SortOrder),
		IsDefault:   cat.IsDefault,
		CreatedAt:   cat.CreatedAt,
		UpdatedAt:   cat.UpdatedAt,
	}
}

// additionalChargeDTO for additional charges
type additionalChargeDTO struct {
	ID                  uuid.UUID  `json:"id"`
	RateCardVersionID   *uuid.UUID `json:"rate_card_version_id,omitempty"`
	CompanyID           *uuid.UUID `json:"company_id,omitempty"`
	Name                string     `json:"name"`
	ChargeType          string     `json:"charge_type"` // 'fixed' or 'percentage'
	AmountCents         int        `json:"amount_cents"`
	Percentage          float64    `json:"percentage,omitempty"`
	IsMandatory         bool       `json:"is_mandatory"`
	AppliesToRoutes     []string   `json:"applies_to_routes,omitempty"`
	AppliesToServices   []string   `json:"applies_to_services,omitempty"`
	AppliesToCategories []string   `json:"applies_to_categories,omitempty"`
	IsIncludedInPrice   bool       `json:"is_included_in_price"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func toAdditionalChargeDTO(charge db.AdditionalCharge) additionalChargeDTO {
	appliesToRoutes := []string{}
	appliesToServices := []string{}
	appliesToCategories := []string{}

	if len(charge.AppliesToRoutes) > 0 {
		var routes []string
		if err := json.Unmarshal(charge.AppliesToRoutes, &routes); err == nil {
			appliesToRoutes = routes
		}
	}
	if len(charge.AppliesToServices) > 0 {
		var services []string
		if err := json.Unmarshal(charge.AppliesToServices, &services); err == nil {
			appliesToServices = services
		}
	}
	if len(charge.AppliesToCategories) > 0 {
		var cats []string
		if err := json.Unmarshal(charge.AppliesToCategories, &cats); err == nil {
			appliesToCategories = cats
		}
	}

	pct := 0.0
	if charge.Percentage.Valid {
		pct = numericToFloat64(charge.Percentage)
	}

	return additionalChargeDTO{
		ID:                  charge.ID,
		RateCardVersionID:   charge.RateCardVersionID,
		CompanyID:           charge.CompanyID,
		Name:                charge.Name,
		ChargeType:          string(charge.ChargeType),
		AmountCents:         int(charge.AmountCents),
		Percentage:          pct,
		IsMandatory:         charge.IsMandatory,
		AppliesToRoutes:     appliesToRoutes,
		AppliesToServices:   appliesToServices,
		AppliesToCategories: appliesToCategories,
		IsIncludedInPrice:   charge.IsIncludedInPrice,
		CreatedAt:           charge.CreatedAt,
		UpdatedAt:           charge.UpdatedAt,
	}
}

// shippingQuoteDTO for quote responses
type shippingQuoteDTO struct {
	ID                      uuid.UUID `json:"id"`
	QuoteReference          string    `json:"quote_reference"`
	OriginLocationType      string    `json:"origin_location_type"`
	DestinationLocationType string    `json:"destination_location_type"`
	SelectedService         string    `json:"selected_service"`
	ChargeableWeightKG      float64   `json:"chargeable_weight_kg"`
	BasePriceNGN            float64   `json:"base_price_ngn"`
	TotalPriceNGN           float64   `json:"total_price_ngn"`
	Status                  string    `json:"status"`
	EstimatedDuration       int       `json:"estimated_duration_days"`
	CreatedAt               time.Time `json:"created_at"`
}

// quoteChargeDTO for individual charges in a quote
type quoteChargeDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	ChargeType  string    `json:"charge_type"` // 'fixed' or 'percentage'
	AmountNGN   float64   `json:"amount_ngn"`
	Percentage  float64   `json:"percentage,omitempty"`
	IsMandatory bool      `json:"is_mandatory"`
	AppliedOn   string    `json:"applied_on"`
}

// createCountryInput for creating countries (catalogue - platform admin only)
type createCountryInput struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	Code      string `json:"code" binding:"required,alphaunicode,min=2,max=10"`
	NumericID int    `json:"numeric_code" binding:"required,min=1,max=9999"`
	Status    string `json:"status" binding:"required,oneof=active retired"`
}

// createStateInput for creating states
type createStateInput struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	Code      string `json:"code" binding:"required,alphaunicode,min=2,max=10"`
	CountryID string `json:"country_id" binding:"required,uuid"`
	Status    string `json:"status" binding:"required,oneof=active retired"`
}

// createCityInput for creating cities
type createCityInput struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	StateID   string `json:"state_id" binding:"required,uuid"`
	CountryID string `json:"country_id" binding:"required,uuid"`
	Status    string `json:"status" binding:"required,oneof=active retired"`
}

// createTerminalInput for creating terminals
type createTerminalInput struct {
	Name    string `json:"name" binding:"required,min=2,max=100"`
	Code    string `json:"code" binding:"required,alphanum,min=2,max=20"`
	CityID  string `json:"city_id" binding:"required,uuid"`
	Address string `json:"address" binding:"required,max=240"`
	Status  string `json:"status" binding:"required,oneof=active retired"`
}

// createShippingServiceInput for creating shipping services
type createShippingServiceInput struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"required"`
	ServiceType string `json:"service_type" binding:"required,oneof=standard express overnight same_day economy"`
	Status      string `json:"status" binding:"required,oneof=active inactive"`
}

// createCompanyRouteInput for creating routes
type createCompanyRouteInput struct {
	OriginLocationType      string `json:"origin_location_type" binding:"required,oneof=country state city terminal"`
	OriginLocationID        string `json:"origin_location_id" binding:"required,uuid"`
	DestinationLocationType string `json:"destination_location_type" binding:"required,oneof=country state city terminal"`
	DestinationLocationID   string `json:"destination_location_id" binding:"required,uuid"`
	IsDirectional           bool   `json:"is_directionality" binding:"required"`
	Status                  string `json:"status" binding:"required,oneof=active inactive"`
}

// createRateCardInput for creating rate cards
type createRateCardInput struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"required,oneof=draft published"`
}

// createRateCardVersionInput for creating rate card versions
type createRateCardVersionInput struct {
	RateCardID string `json:"rate_card_id" binding:"required,uuid"`
	Version    int    `json:"version" binding:"required,min=1"`
	IsActive   bool   `json:"is_active" binding:"required"`
	Notes      string `json:"notes"`
}

// createPricingRuleInput for creating pricing rules
type createPricingRuleInput struct {
	RateCardVersionID string  `json:"rate_card_version_id" binding:"required,uuid"`
	RuleType          string  `json:"rule_type" binding:"required,oneof=fixed weight_based per_kg volumetric weight_based category_based"`
	Name              string  `json:"name" binding:"required,min=2,max=100"`
	Description       string  `json:"description"`
	IsActive          bool    `json:"is_active" binding:"required"`
	EffectiveFrom     string  `json:"effective_from" binding:"required,datetime"`
	EffectiveUntil    *string `json:"effective_until,omitempty" binding:"omitempty,datetime"`
}

// createPricingTierInput for creating pricing tiers (weight brackets)
type createPricingTierInput struct {
	RateCardVersionID string   `json:"rate_card_version_id" binding:"required,uuid"`
	MinWeightKG       float64  `json:"min_weight_kg" binding:"required,min=0"`
	MaxWeightKG       *float64 `json:"max_weight_kg" binding:"omitempty,min=0"`
	PriceCents        int      `json:"price_cents" binding:"required,min=0"`
	Name              string   `json:"name" binding:"required,min=2,max=100"`
	Description       string   `json:"description"`
}

// createPackageCategoryInput for creating package categories
type createPackageCategoryInput struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"required"`
	SortOrder   int    `json:"sort_order" binding:"min=0"`
	IsDefault   bool   `json:"is_default"`
}

// createAdditionalChargeInput for creating additional charges
type createAdditionalChargeInput struct {
	RateCardVersionID   *string  `json:"rate_card_version_id,omitempty" binding:"omitempty,uuid"`
	CompanyID           *string  `json:"company_id,omitempty" binding:"omitempty,uuid"`
	Name                string   `json:"name" binding:"required,min=2,max=100"`
	ChargeType          string   `json:"charge_type" binding:"required,oneof=fixed percentage"`
	AmountCents         int      `json:"amount_cents" binding:"required,min=0"`
	Percentage          *float64 `json:"percentage,omitempty" binding:"omitempty,min=0,max=1000"`
	IsMandatory         bool     `json:"is_mandatory" binding:"required"`
	AppliesToRoutes     string   `json:"applies_to_routes,omitempty" binding:"omitempty"`
	AppliesToServices   string   `json:"applies_to_services,omitempty" binding:"omitempty"`
	AppliesToCategories string   `json:"applies_to_categories,omitempty" binding:"omitempty"`
	IsIncludedInPrice   bool     `json:"is_included_in_price" binding:"required"`
}

// createShippingQuoteInput for creating shipping quotes
type createShippingQuoteInput struct {
	OriginLocationType      string  `json:"origin_location_type" binding:"required,oneof=country state city terminal"`
	OriginLocationID        string  `json:"origin_location_id" binding:"required,uuid"`
	DestinationLocationType string  `json:"destination_location_type" binding:"required,oneof=country state city terminal"`
	DestinationLocationID   string  `json:"destination_location_id" binding:"required,uuid"`
	SelectedServiceID       string  `json:"selected_service_id" binding:"required,uuid"`
	RateCardVersionID       string  `json:"rate_card_version_id" binding:"required,uuid"`
	QuoteReference          string  `json:"quote_reference" binding:"required,min=3,max=50"`
	ReferenceWeightKG       float64 `json:"reference_weight_kg" binding:"required,min=0"`
	ReferenceLengthCM       float64 `json:"reference_length_cm" binding:"required,min=0"`
	ReferenceWidthCM        float64 `json:"reference_width_cm" binding:"required,min=0"`
	ReferenceHeightCM       float64 `json:"reference_height_cm" binding:"required,min=0"`
	ReferenceCategory       string  `json:"reference_category" binding:"required,oneof=document small_parcel medium_parcel large_parcel fragile electronics bulk"`
	ChargeableWeightKG      float64 `json:"chargeable_weight_kg" binding:"required,min=0"`
	BasePriceCents          int     `json:"base_price_cents" binding:"required,min=0"`
	TotalPriceCents         int     `json:"total_price_cents" binding:"required,min=0"`
	EstimatedDuration       int     `json:"estimated_duration_days" binding:"required,min=1"`
}

// ---- Country handlers ----

func (a *API) listCountries(c *gin.Context) {
	countries, err := a.svc.Q.ListCountries(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]countryDTO, 0, len(countries))
	for _, c := range countries {
		result = append(result, toCountryDTO(c))
	}
	c.JSON(http.StatusOK, gin.H{"countries": result})
}

func (a *API) createCountry(c *gin.Context) {
	var req createCountryInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the country data.")
		return
	}

	countries, err := a.svc.Q.CreateCountry(c.Request.Context(), db.CreateCountryParams{
		Name:        req.Name,
		Code:        req.Code,
		NumericCode: int32(req.NumericID),
		Status:      req.Status,
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "country_code_taken", "That country code is already in use.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"country": toCountryDTO(countries)})
}

// ---- State handlers ----

func (a *API) listStates(c *gin.Context) {
	code := c.Query("country_code")
	var states []db.State
	var err error

	if code != "" {
		countryID, parseErr := uuid.Parse(code)
		if parseErr != nil {
			c.JSON(http.StatusOK, gin.H{"states": []stateDTO{}})
			return
		}
		states, err = a.svc.Q.ListStatesByCountry(c.Request.Context(), countryID)
	} else {
		states = []db.State{}
	}
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]stateDTO, 0, len(states))
	for _, s := range states {
		result = append(result, toStateDTO(s))
	}
	c.JSON(http.StatusOK, gin.H{"states": result})
}

func (a *API) createState(c *gin.Context) {
	var req createStateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the state data.")
		return
	}

	countryID, err := uuid.Parse(req.CountryID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_country", "Country ID must be a valid UUID.")
		return
	}

	code := &req.Code
	if req.Code == "" {
		code = nil
	}
	states, err := a.svc.Q.CreateState(c.Request.Context(), db.CreateStateParams{
		Name:      req.Name,
		Code:      code,
		CountryID: countryID,
		Status:    "active",
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "state_code_taken", "That state code is already used in this country.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"state": toStateDTO(states)})
}

// ---- City handlers ----

func (a *API) listCities(c *gin.Context) {
	stateID := c.Query("state_code")
	var cities []db.City
	var err error

	if stateID != "" {
		parsedID, parseErr := uuid.Parse(stateID)
		if parseErr != nil {
			c.JSON(http.StatusOK, gin.H{"cities": []cityDTO{}})
			return
		}
		cities, err = a.svc.Q.ListCitiesByState(c.Request.Context(), &parsedID)
	} else {
		cities = []db.City{}
	}
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]cityDTO, 0, len(cities))
	for _, c := range cities {
		result = append(result, toCityDTO(c))
	}
	c.JSON(http.StatusOK, gin.H{"cities": result})
}

func (a *API) createCity(c *gin.Context) {
	var req createCityInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the city data.")
		return
	}

	stateID, err := uuid.Parse(req.StateID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_state", "State ID must be a valid UUID.")
		return
	}

	countryID, err := uuid.Parse(req.CountryID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_country", "Country ID must be a valid UUID.")
		return
	}

	cities, err := a.svc.Q.CreateCity(c.Request.Context(), db.CreateCityParams{
		Name:      req.Name,
		StateID:   &stateID,
		CountryID: countryID,
		Status:    "active",
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "city_name_taken", "A city with that name already exists in this state.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"city": toCityDTO(cities)})
}

// ---- Terminal handlers ----

func (a *API) listTerminals(c *gin.Context) {
	cityID := c.Query("city_id")
	var terminals []db.Terminal
	var err error

	if cityID != "" {
		parsedID, parseErr := uuid.Parse(cityID)
		if parseErr != nil {
			c.JSON(http.StatusOK, gin.H{"terminals": []terminalDTO{}})
			return
		}
		terminals, err = a.svc.Q.ListTerminalsByCity(c.Request.Context(), &parsedID)
	} else {
		terminals = []db.Terminal{}
	}
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]terminalDTO, 0, len(terminals))
	for _, t := range terminals {
		result = append(result, toTerminalDTO(t))
	}
	c.JSON(http.StatusOK, gin.H{"terminals": result})
}

func (a *API) createTerminal(c *gin.Context) {
	var req createTerminalInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the terminal data.")
		return
	}

	cityID, err := uuid.Parse(req.CityID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_city", "City ID must be a valid UUID.")
		return
	}
	cityIDPtr := &cityID
	address := &req.Address

	terminals, err := a.svc.Q.CreateTerminal(c.Request.Context(), db.CreateTerminalParams{
		Name:    req.Name,
		Code:    req.Code,
		CityID:  cityIDPtr,
		Address: address,
		Status:  "active",
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "terminal_code_taken", "That terminal code is already in use.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"terminal": toTerminalDTO(terminals)})
}

// ---- Shipping Service handlers ----

func (a *API) listShippingServices(c *gin.Context) {
	tenant, _ := currentTenant(c)
	services, err := a.svc.Q.ListShippingServices(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]shippingServiceDTO, 0, len(services))
	for _, s := range services {
		result = append(result, toShippingServiceDTO(s))
	}
	c.JSON(http.StatusOK, gin.H{"shipping_services": result})
}

func (a *API) createShippingService(c *gin.Context) {
	var req createShippingServiceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the shipping service data.")
		return
	}

	tenant, _ := currentTenant(c)
	services, err := a.svc.Q.CreateShippingService(c.Request.Context(), db.CreateShippingServiceParams{
		CompanyID:         tenant.CompanyID,
		Name:              req.Name,
		Description:       &req.Description,
		ServiceCode:       strings.ToUpper(req.Name),
		DeliveryType:      req.ServiceType,
		EstimatedDuration: 3,
		MinWeightKg:       float64ToNumeric(0),
		MaxWeightKg:       float64ToNumeric(0),
		MaxDimensionCm:    float64ToNumeric(0),
		VolumetricDivisor: float64ToNumeric(5000),
		Status:            req.Status,
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "service_name_taken", "A service with that name already exists.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"shipping_service": toShippingServiceDTO(services)})
}

// ---- Company Route handlers ----

func (a *API) listCompanyRoutes(c *gin.Context) {
	tenant, _ := currentTenant(c)
	routes, err := a.svc.Q.ListCompanyRoutes(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]companyRouteDTO, 0, len(routes))
	for _, r := range routes {
		result = append(result, toCompanyRouteDTO(r))
	}
	c.JSON(http.StatusOK, gin.H{"company_routes": result})
}

func (a *API) createCompanyRoute(c *gin.Context) {
	var req createCompanyRouteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the route data.")
		return
	}

	tenant, _ := currentTenant(c)
	originLocID, err := uuid.Parse(req.OriginLocationID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_origin_location", "Origin location ID must be a valid UUID.")
		return
	}

	destLocID, err := uuid.Parse(req.DestinationLocationID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_destination_location", "Destination location ID must be a valid UUID.")
		return
	}

	routes, err := a.svc.Q.CreateCompanyRoute(c.Request.Context(), db.CreateCompanyRouteParams{
		CompanyID:               tenant.CompanyID,
		OriginLocationType:      req.OriginLocationType,
		OriginLocationID:        originLocID,
		DestinationLocationType: req.DestinationLocationType,
		DestinationLocationID:   destLocID,
		IsRoundTrip:             req.IsDirectional,
		Status:                  req.Status,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"company_route": toCompanyRouteDTO(routes)})
}

// ---- Rate Card handlers ----

func (a *API) listRateCards(c *gin.Context) {
	tenant, _ := currentTenant(c)
	cards, err := a.svc.Q.ListRateCards(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]rateCardDTO, 0, len(cards))
	for _, c := range cards {
		result = append(result, toRateCardDTO(c))
	}
	c.JSON(http.StatusOK, gin.H{"rate_cards": result})
}

func (a *API) createRateCard(c *gin.Context) {
	var req createRateCardInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the rate card data.")
		return
	}

	tenant, _ := currentTenant(c)
	description := &req.Description
	if req.Description == "" {
		description = nil
	}
	cards, err := a.svc.Q.CreateRateCard(c.Request.Context(), db.CreateRateCardParams{
		CompanyID:   tenant.CompanyID,
		Name:        req.Name,
		Description: description,
		Status:      req.Status,
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "rate_card_name_taken", "A rate card with that name already exists.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"rate_card": toRateCardDTO(cards)})
}

func (a *API) getRateCard(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card", "That rate card id is invalid.")
		return
	}

	tenant, _ := currentTenant(c)
	cards, err := a.svc.Q.ListRateCards(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	for _, card := range cards {
		if card.ID == id {
			c.JSON(http.StatusOK, gin.H{"rate_card": toRateCardDTO(card)})
			return
		}
	}
	fail(c, http.StatusNotFound, "rate_card_not_found", "That rate card was not found.")
}

// ---- Rate Card Version handlers ----

func (a *API) listRateCardVersions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card", "That rate card id is invalid.")
		return
	}

	versions, err := a.svc.Q.ListRateCardVersions(c.Request.Context(), id)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]rateCardVersionDTO, 0, len(versions))
	for _, v := range versions {
		result = append(result, toRateCardVersionDTO(v))
	}
	c.JSON(http.StatusOK, gin.H{"rate_card_versions": result})
}

func (a *API) createRateCardVersion(c *gin.Context) {
	if _, err := uuid.Parse(c.Param("id")); err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card", "That rate card id is invalid.")
		return
	}

	var req createRateCardVersionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the rate card version data.")
		return
	}

	rateCardID, err := uuid.Parse(req.RateCardID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card", "Rate card ID must be a valid UUID.")
		return
	}

	versions, err := a.svc.Q.ListRateCardVersions(c.Request.Context(), rateCardID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	newVersion := 1
	if len(versions) > 0 {
		newVersion = int(versions[len(versions)-1].Version) + 1
	}

	created, err := a.svc.Q.CreateRateCardVersion(c.Request.Context(), db.CreateRateCardVersionParams{
		RateCardID:     rateCardID,
		Version:        int32(newVersion),
		IsActive:       req.IsActive,
		EffectiveFrom:  time.Now(),
		EffectiveUntil: nil,
		CreatedBy:      nil,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"rate_card_version": toRateCardVersionDTO(created)})
}

// ---- Pricing Rule handlers ----

func (a *API) listPricingRules(c *gin.Context) {
	versionID, err := uuid.Parse(c.Param("version_id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card_version", "That rate card version id is invalid.")
		return
	}

	rules, err := a.svc.Q.ListPricingRules(c.Request.Context(), versionID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]pricingRuleDTO, 0, len(rules))
	for _, r := range rules {
		result = append(result, toPricingRuleDTO(r))
	}
	c.JSON(http.StatusOK, gin.H{"pricing_rules": result})
}

func (a *API) createPricingRule(c *gin.Context) {
	if _, err := uuid.Parse(c.Param("version_id")); err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card_version", "That rate card version id is invalid.")
		return
	}

	var req createPricingRuleInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the pricing rule data.")
		return
	}

	rateCardVersionID, err := uuid.Parse(req.RateCardVersionID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card_version", "Rate card version ID must be a valid UUID.")
		return
	}

	effectiveFrom, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_effective_from", "Effective from must be in YYYY-MM-DD format.")
		return
	}

	var effectiveUntil *time.Time
	if req.EffectiveUntil != nil && *req.EffectiveUntil != "" {
		parsed, err := time.Parse("2006-01-02", *req.EffectiveUntil)
		if err != nil {
			fail(c, http.StatusBadRequest, "invalid_effective_until", "Effective until must be in YYYY-MM-DD format.")
			return
		}
		effectiveUntil = &parsed
	}

	description := &req.Description
	if req.Description == "" {
		description = nil
	}

	rules, err := a.svc.Q.CreatePricingRule(c.Request.Context(), db.CreatePricingRuleParams{
		RateCardVersionID:    rateCardVersionID,
		RuleType:             req.RuleType,
		Name:                 req.Name,
		Description:          description,
		FixedAmount:          float64ToNumeric(0),
		PerKgBaseCharge:      float64ToNumeric(0),
		PerKgRate:            float64ToNumeric(0),
		PerKgMinimum:         float64ToNumeric(0),
		PerKgMaximum:         float64ToNumeric(0),
		VolumetricDivisor:    float64ToNumeric(5000),
		CategoryID:           nil,
		ApplicableRoutes:     nil,
		ApplicableServices:   nil,
		ApplicableCategories: nil,
		EffectiveFrom:        effectiveFrom,
		EffectiveUntil:       effectiveUntil,
		IsActive:             req.IsActive,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"pricing_rule": toPricingRuleDTO(rules)})
}

// ---- Pricing Tier handlers ----

func (a *API) listPricingTiers(c *gin.Context) {
	versionID, err := uuid.Parse(c.Param("version_id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card_version", "That rate card version id is invalid.")
		return
	}

	tiers, err := a.svc.Q.ListPricingTiers(c.Request.Context(), versionID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]pricingTierDTO, 0, len(tiers))
	for _, t := range tiers {
		result = append(result, toPricingTierDTO(t))
	}
	c.JSON(http.StatusOK, gin.H{"pricing_tiers": result})
}

func (a *API) createPricingTier(c *gin.Context) {
	if _, err := uuid.Parse(c.Param("version_id")); err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card_version", "That rate card version id is invalid.")
		return
	}

	var req createPricingTierInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the pricing tier data.")
		return
	}

	rateCardVersionID, err := uuid.Parse(req.RateCardVersionID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card_version", "Rate card version ID must be a valid UUID.")
		return
	}

	var maxWeight pgtype.Numeric
	if req.MaxWeightKG != nil {
		maxWeight = float64ToNumeric(*req.MaxWeightKG)
	}

	description := &req.Description
	if req.Description == "" {
		description = nil
	}

	tiers, err := a.svc.Q.CreatePricingTier(c.Request.Context(), db.CreatePricingTierParams{
		RateCardVersionID: rateCardVersionID,
		MinWeightKg:       float64ToNumeric(req.MinWeightKG),
		MaxWeightKg:       maxWeight,
		PriceCents:        int32(req.PriceCents),
		Name:              req.Name,
		Description:       description,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"pricing_tier": toPricingTierDTO(tiers)})
}

// ---- Package Category handlers ----

func (a *API) listPackageCategories(c *gin.Context) {
	categories, err := a.svc.Q.ListPackageCategories(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]packageCategoryDTO, 0, len(categories))
	for _, c := range categories {
		result = append(result, toPackageCategoryDTO(c))
	}
	c.JSON(http.StatusOK, gin.H{"package_categories": result})
}

func (a *API) createPackageCategory(c *gin.Context) {
	var req createPackageCategoryInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the package category data.")
		return
	}

	description := &req.Description
	if req.Description == "" {
		description = nil
	}

	categories, err := a.svc.Q.CreatePackageCategory(c.Request.Context(), db.CreatePackageCategoryParams{
		Name:        req.Name,
		Description: description,
		SortOrder:   int32(req.SortOrder),
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		if isUniqueViolation(err) {
			fail(c, http.StatusConflict, "package_category_name_taken", "A package category with that name already exists.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"package_category": toPackageCategoryDTO(categories)})
}

// ---- Additional Charge handlers ----

func (a *API) listAdditionalCharges(c *gin.Context) {
	versionIDStr := c.Query("version_id")
	companyIDStr := c.Query("company_id")

	var versionID *uuid.UUID
	var companyID *uuid.UUID

	if versionIDStr != "" {
		id, err := uuid.Parse(versionIDStr)
		if err == nil {
			versionID = &id
		}
	}
	if companyIDStr != "" {
		id, err := uuid.Parse(companyIDStr)
		if err == nil {
			companyID = &id
		}
	}

	charges, err := a.svc.Q.ListAdditionalCharges(c.Request.Context(), db.ListAdditionalChargesParams{
		RateCardVersionID: versionID,
		CompanyID:         companyID,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	result := make([]additionalChargeDTO, 0, len(charges))
	for _, ch := range charges {
		result = append(result, toAdditionalChargeDTO(ch))
	}
	c.JSON(http.StatusOK, gin.H{"additional_charges": result})
}

func (a *API) createAdditionalCharge(c *gin.Context) {
	var req createAdditionalChargeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the additional charge data.")
		return
	}

	var rateCardVersionID *uuid.UUID
	if req.RateCardVersionID != nil && *req.RateCardVersionID != "" {
		id, err := uuid.Parse(*req.RateCardVersionID)
		if err == nil {
			rateCardVersionID = &id
		}
	}

	var companyID *uuid.UUID
	if req.CompanyID != nil && *req.CompanyID != "" {
		id, err := uuid.Parse(*req.CompanyID)
		if err == nil {
			companyID = &id
		}
	}

	var percentage pgtype.Numeric
	if req.Percentage != nil {
		percentage = float64ToNumeric(*req.Percentage)
	}

	charges, err := a.svc.Q.CreateAdditionalCharge(c.Request.Context(), db.CreateAdditionalChargeParams{
		RateCardVersionID: rateCardVersionID,
		CompanyID:         companyID,
		Name:              req.Name,
		ChargeType:        req.ChargeType,
		AmountCents:       int32(req.AmountCents),
		Percentage:        percentage,
		IsMandatory:       req.IsMandatory,
		IsIncludedInPrice: req.IsIncludedInPrice,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"additional_charge": toAdditionalChargeDTO(charges)})
}

// ---- Shipping Quote handlers ----

func (a *API) createShippingQuote(c *gin.Context) {
	var req createShippingQuoteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_input", "Check the shipping quote data.")
		return
	}

	tenant, _ := currentTenant(c)
	originLocID, err := uuid.Parse(req.OriginLocationID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_origin_location", "Origin location ID must be a valid UUID.")
		return
	}

	destLocID, err := uuid.Parse(req.DestinationLocationID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_destination_location", "Destination location ID must be a valid UUID.")
		return
	}

	selectedServiceID, err := uuid.Parse(req.SelectedServiceID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_service", "Selected service ID must be a valid UUID.")
		return
	}

	rateCardVersionID, err := uuid.Parse(req.RateCardVersionID)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid_rate_card_version", "Rate card version ID must be a valid UUID.")
		return
	}

	quote, err := a.svc.Q.CreateShippingQuote(c.Request.Context(), db.CreateShippingQuoteParams{
		CompanyID:               tenant.CompanyID,
		CustomerID:              nil,
		OriginLocationType:      req.OriginLocationType,
		OriginLocationID:        originLocID,
		DestinationLocationType: req.DestinationLocationType,
		DestinationLocationID:   destLocID,
		SelectedServiceID:       selectedServiceID,
		RateCardVersionID:       &rateCardVersionID,
		QuoteReference:          req.QuoteReference,
		ReferenceWeightKg:       float64ToNumeric(req.ReferenceWeightKG),
		ReferenceLengthCm:       float64ToNumeric(req.ReferenceLengthCM),
		ReferenceWidthCm:        float64ToNumeric(req.ReferenceWidthCM),
		ReferenceHeightCm:       float64ToNumeric(req.ReferenceHeightCM),
		ReferenceCategory:       req.ReferenceCategory,
		ChargeableWeightKg:      float64ToNumeric(req.ChargeableWeightKG),
		BasePriceCents:          int32(req.BasePriceCents),
		TotalPriceCents:         int32(req.TotalPriceCents),
		Currency:                "NGN",
		EstimatedDuration:       int32(req.EstimatedDuration),
		Status:                  "valid",
		ExpiresAt:               time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"shipping_quote": toShippingQuoteDTO(quote)})
}

func toShippingQuoteDTO(quote db.ShippingQuote) shippingQuoteDTO {
	return shippingQuoteDTO{
		ID:                      quote.ID,
		QuoteReference:          quote.QuoteReference,
		OriginLocationType:      quote.OriginLocationType,
		DestinationLocationType: quote.DestinationLocationType,
		SelectedService:         "standard",
		ChargeableWeightKG:      numericToFloat64(quote.ChargeableWeightKg),
		BasePriceNGN:            float64(quote.BasePriceCents) / 100.0,
		TotalPriceNGN:           float64(quote.TotalPriceCents) / 100.0,
		Status:                  quote.Status,
		EstimatedDuration:       int(quote.EstimatedDuration),
		CreatedAt:               quote.CreatedAt,
	}
}
