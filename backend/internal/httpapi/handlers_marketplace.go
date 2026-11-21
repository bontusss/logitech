package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/bol-platform/backend/internal/db/sqlc"
	"github.com/bol-platform/backend/internal/service"
)

const (
	marketplaceGuestCookie = "bol_marketplace_guest"
	marketplaceWorkers     = 8
	marketplaceTimeout     = 4 * time.Second
)

type marketplaceSearchInput struct {
	ServiceType      string   `json:"service_type" binding:"omitempty,oneof=intercity last_mile"`
	OriginState      string   `json:"origin_state" binding:"required,min=2,max=100"`
	OriginCity       string   `json:"origin_city" binding:"required,min=2,max=100"`
	DestinationState string   `json:"destination_state" binding:"required,min=2,max=100"`
	DestinationCity  string   `json:"destination_city" binding:"required,min=2,max=100"`
	PickupAddress    string   `json:"pickup_address" binding:"omitempty,max=500"`
	DropoffAddress   string   `json:"dropoff_address" binding:"omitempty,max=500"`
	PickupLatitude   *float64 `json:"pickup_latitude"`
	PickupLongitude  *float64 `json:"pickup_longitude"`
	DropoffLatitude  *float64 `json:"dropoff_latitude"`
	DropoffLongitude *float64 `json:"dropoff_longitude"`
	DistanceKm       float64  `json:"distance_km" binding:"gte=0,lte=500"`
	WeightKg         float64  `json:"weight_kg" binding:"required,gt=0,lte=500"`
	LengthCM         *float64 `json:"length_cm"`
	WidthCM          *float64 `json:"width_cm"`
	HeightCM         *float64 `json:"height_cm"`
	Category         string   `json:"category" binding:"omitempty,max=50"`
	Speed            string   `json:"speed" binding:"omitempty,oneof=standard express same_day"`
	PickupRequired   bool     `json:"pickup_required"`
	DeliveryRequired bool     `json:"delivery_required"`
	Sort             string   `json:"sort" binding:"omitempty,oneof=cheapest fastest recommended"`
}

func (req marketplaceSearchInput) toRouteRequest() (service.MarketplaceRouteRequest, error) {
	length, width, height := 0.0, 0.0, 0.0
	provided := 0
	for _, value := range []*float64{req.LengthCM, req.WidthCM, req.HeightCM} {
		if value != nil {
			provided++
			if *value <= 0 || *value > 300 {
				return service.MarketplaceRouteRequest{}, errors.New("Package dimensions must be between 0 and 300 cm.")
			}
		}
	}
	if provided != 0 && provided != 3 {
		return service.MarketplaceRouteRequest{}, errors.New("Enter length, width and height together.")
	}
	if req.LengthCM != nil {
		length, width, height = *req.LengthCM, *req.WidthCM, *req.HeightCM
	}
	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = "other"
	}
	serviceType := strings.TrimSpace(req.ServiceType)
	if serviceType == "" {
		serviceType = "intercity"
	}
	if serviceType == "last_mile" {
		if strings.TrimSpace(req.PickupAddress) == "" || strings.TrimSpace(req.DropoffAddress) == "" {
			return service.MarketplaceRouteRequest{}, errors.New("Choose verified pickup and drop-off addresses.")
		}
		if !validCoordinatePair(req.PickupLatitude, req.PickupLongitude) ||
			!validCoordinatePair(req.DropoffLatitude, req.DropoffLongitude) {
			return service.MarketplaceRouteRequest{}, errors.New("Choose pickup and drop-off locations from the map suggestions.")
		}
		if req.DistanceKm <= 0 || req.DistanceKm > 200 {
			return service.MarketplaceRouteRequest{}, errors.New("The last-mile route distance must be between 0 and 200 km.")
		}
	}
	return service.MarketplaceRouteRequest{
		ServiceType:      serviceType,
		OriginState:      strings.TrimSpace(req.OriginState),
		OriginCity:       strings.TrimSpace(req.OriginCity),
		DestinationState: strings.TrimSpace(req.DestinationState),
		DestinationCity:  strings.TrimSpace(req.DestinationCity),
		PickupAddress:    strings.TrimSpace(req.PickupAddress),
		DropoffAddress:   strings.TrimSpace(req.DropoffAddress),
		DistanceKm:       req.DistanceKm,
		Speed:            req.Speed,
		PickupRequired:   req.PickupRequired,
		DeliveryRequired: req.DeliveryRequired,
		Package: service.MarketplacePackage{
			WeightKg: req.WeightKg, LengthCM: length, WidthCM: width, HeightCM: height, Category: category,
		},
	}, nil
}

type marketplaceProviderFailure struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
}

type marketplaceSearchResult struct {
	quotes     []service.MarketplaceQuote
	searchID   uuid.UUID
	considered int
	failures   []marketplaceProviderFailure
}

func (a *API) marketplaceLocations(c *gin.Context) {
	rows, err := a.svc.Q.ListMarketplaceLocations(c.Request.Context())
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	locations := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		locations = append(locations, gin.H{"state": row.State, "city": row.City})
	}
	c.JSON(http.StatusOK, gin.H{"locations": locations})
}

func (a *API) marketplaceProviders(c *gin.Context) {
	state := strings.TrimSpace(c.Query("state"))
	city := strings.TrimSpace(c.Query("city"))
	params := db.ListMarketplaceProvidersParams{}
	if state != "" {
		params.State = &state
	}
	if city != "" {
		params.City = &city
	}
	rows, err := a.svc.Q.ListMarketplaceProviders(c.Request.Context(), params)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	providers := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		providers = append(providers, gin.H{
			"id":               row.ID,
			"name":             row.Name,
			"slug":             row.Slug,
			"logo_url":         row.LogoUrl,
			"cover_image_url":  row.CoverImageUrl,
			"description":      row.PublicDescription,
			"route_count":      row.RouteCount,
			"from_price_minor": row.FromPriceMinor,
			"verified":         true,
		})
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

func (a *API) marketplaceProvider(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	provider, err := a.svc.Q.GetMarketplaceProviderBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "provider_not_found", "That logistics provider was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	routes, err := a.svc.Q.ListMarketplaceProviderRoutes(c.Request.Context(), provider.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}

	services, coverage, lanes := summarizeProviderRoutes(routes)
	c.JSON(http.StatusOK, gin.H{
		"provider": gin.H{
			"id":              provider.ID,
			"name":            provider.Name,
			"slug":            provider.Slug,
			"logo_url":        provider.LogoUrl,
			"cover_image_url": provider.CoverImageUrl,
			"primary_color":   provider.PrimaryColor,
			"secondary_color": provider.SecondaryColor,
			"description":     provider.PublicDescription,
			"business_type":   provider.BusinessType,
			"state":           provider.State,
			"city":            provider.City,
			"country":         provider.Country,
			"email":           provider.BusinessEmail,
			"phone":           provider.Phone,
			"subdomain":       provider.Subdomain,
			"verified":        true,
			"verified_at":     provider.VerifiedAt,
			"published_at":    provider.PublishedAt,
			"currency":        provider.Currency,
		},
		"services": services,
		"coverage": coverage,
		"lanes":    lanes,
	})
}

func (a *API) marketplaceProviderServices(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	provider, err := a.svc.Q.GetMarketplaceProviderBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "provider_not_found", "That logistics provider was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	routes, err := a.svc.Q.ListMarketplaceProviderRoutes(c.Request.Context(), provider.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	services, coverage, lanes := summarizeProviderRoutes(routes)
	c.JSON(http.StatusOK, gin.H{"services": services, "coverage": coverage, "lanes": lanes})
}

func summarizeProviderRoutes(rows []db.ListMarketplaceProviderRoutesRow) (services []gin.H, coverage []gin.H, lanes []gin.H) {
	type summary struct {
		key       string
		name      string
		routes    int
		minCharge int64
		transit   string
		estimated bool
	}
	defs := []struct{ key, name string }{
		{"standard", "Standard"},
		{"express", "Express"},
		{"same_day", "Same-day"},
	}
	byKey := map[string]*summary{}
	for _, def := range defs {
		byKey[def.key] = &summary{key: def.key, name: def.name}
	}
	seenLocations := map[string]bool{}

	for _, row := range rows {
		speeds := []struct {
			key  string
			rate pgtype.Int8
		}{
			{"standard", row.StandardPricePerKmMinor},
			{"express", row.ExpressPricePerKmMinor},
			{"same_day", row.SameDayPricePerKmMinor},
		}
		available := make([]gin.H, 0, 3)
		for _, speed := range speeds {
			if !speed.rate.Valid || speed.rate.Int64 <= 0 {
				continue
			}
			item := byKey[speed.key]
			item.routes++
			if item.minCharge == 0 || row.MinimumChargeMinor < item.minCharge {
				item.minCharge = row.MinimumChargeMinor
			}
			if item.transit == "" && row.EstimatedTransitTime != nil {
				item.transit = strings.TrimSpace(*row.EstimatedTransitTime)
			}
			item.estimated = item.estimated || row.PriceIsEstimate
			available = append(available, gin.H{"key": speed.key, "price_per_km_minor": speed.rate.Int64})
		}
		if len(available) > 0 {
			lanes = append(lanes, gin.H{
				"id":                   row.ID,
				"origin_state":         row.OriginState,
				"origin_city":          row.OriginCity,
				"destination_state":    row.DestinationState,
				"destination_city":     row.DestinationCity,
				"distance_km":          numericToFloat64(row.DistanceKm),
				"minimum_charge_minor": row.MinimumChargeMinor,
				"transit":              row.EstimatedTransitTime,
				"departure_schedule":   row.DepartureSchedule,
				"pickup_available":     row.PickupAvailable,
				"delivery_available":   row.DeliveryAvailable,
				"price_is_estimate":    row.PriceIsEstimate,
				"speeds":               available,
			})
		}
		for _, loc := range [][2]string{{row.OriginState, row.OriginCity}, {row.DestinationState, row.DestinationCity}} {
			key := strings.ToLower(loc[0]) + "|" + strings.ToLower(loc[1])
			if !seenLocations[key] {
				seenLocations[key] = true
				coverage = append(coverage, gin.H{"state": loc[0], "city": loc[1]})
			}
		}
	}

	for _, def := range defs {
		item := byKey[def.key]
		if item.routes == 0 {
			continue
		}
		services = append(services, gin.H{
			"key":                       item.key,
			"name":                      item.name,
			"route_count":               item.routes,
			"from_minimum_charge_minor": item.minCharge,
			"estimated_delivery":        item.transit,
			"price_is_estimate":         item.estimated,
		})
	}
	return services, coverage, lanes
}

func (a *API) marketplaceSearch(c *gin.Context) {
	var input marketplaceSearchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fail(c, http.StatusBadRequest, "invalid_search", "Add an origin, destination and package weight to search.")
		return
	}
	req, err := input.toRouteRequest()
	if err != nil {
		fail(c, http.StatusUnprocessableEntity, "invalid_search", err.Error())
		return
	}
	if req.ServiceType != "last_mile" && strings.EqualFold(req.OriginCity, req.DestinationCity) && strings.EqualFold(req.OriginState, req.DestinationState) {
		fail(c, http.StatusUnprocessableEntity, "same_location", "Origin and destination must be different.")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), marketplaceTimeout)
	defer cancel()

	customerID := (*uuid.UUID)(nil)
	if user, ok := currentUser(c); ok {
		customerID = &user.ID
	}

	result, err := a.runMarketplaceSearch(ctx, req, customerID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	order := input.Sort
	if order == "" {
		order = "cheapest"
	}
	service.SortMarketplaceQuotes(result.quotes, order)

	partial := len(result.failures) > 0 && len(result.quotes) > 0
	c.JSON(http.StatusOK, gin.H{
		"search_id": result.searchID,
		"quotes":    result.quotes,
		"meta": gin.H{
			"providers_considered": result.considered,
			"providers_quoted":     countProviders(result.quotes),
			"result_count":         len(result.quotes),
			"failures":             result.failures,
			"partial":              partial,
			"sort":                 order,
			"expires_at":           time.Now().Add(service.MarketplaceQuoteExpiry),
		},
	})
}

func countProviders(quotes []service.MarketplaceQuote) int {
	seen := map[string]bool{}
	for _, quote := range quotes {
		seen[quote.ProviderID] = true
	}
	return len(seen)
}

type providerQuoteJob struct {
	provider service.ProviderRate
}

type providerQuoteResult struct {
	quotes  []service.MarketplaceQuote
	failure *marketplaceProviderFailure
}

func (a *API) runMarketplaceSearch(ctx context.Context, req service.MarketplaceRouteRequest, customerID *uuid.UUID) (marketplaceSearchResult, error) {
	if req.ServiceType == "last_mile" {
		return a.runLastMileMarketplaceSearch(ctx, req, customerID)
	}
	rows, err := a.svc.Q.ListMarketplaceRoutesForLane(ctx, db.ListMarketplaceRoutesForLaneParams{
		OriginState:      req.OriginState,
		OriginCity:       req.OriginCity,
		DestinationState: req.DestinationState,
		DestinationCity:  req.DestinationCity,
	})
	if err != nil {
		return marketplaceSearchResult{}, err
	}

	expiresAt := time.Now().Add(service.MarketplaceQuoteExpiry)
	jobs := make(chan providerQuoteJob)
	results := make(chan providerQuoteResult, len(rows))
	sem := make(chan struct{}, marketplaceWorkers)
	var wg sync.WaitGroup

	for _, row := range rows {
		job := providerQuoteJob{provider: providerRateFromRow(row)}
		wg.Add(1)
		go func(job providerQuoteJob) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results <- providerQuoteResult{failure: &marketplaceProviderFailure{Provider: job.provider.CompanyName, Code: "timeout"}}
				return
			}
			quotes, err := service.BuildProviderQuotes(job.provider, req, expiresAt)
			if err != nil {
				a.log.Warn("marketplace provider quote failed", "company_id", job.provider.CompanyID, "error", err)
				results <- providerQuoteResult{failure: &marketplaceProviderFailure{Provider: job.provider.CompanyName, Code: quoteFailureCode(err)}}
				return
			}
			results <- providerQuoteResult{quotes: quotes}
		}(job)
	}
	go func() {
		wg.Wait()
		close(results)
		close(jobs)
	}()

	quotes := make([]service.MarketplaceQuote, 0)
	failures := make([]marketplaceProviderFailure, 0)
	seen := map[string]bool{}
	considered := map[string]bool{}
	for _, row := range rows {
		considered[row.CompanyID.String()] = true
	}
	for result := range results {
		if result.failure != nil {
			failures = append(failures, *result.failure)
			continue
		}
		for _, quote := range result.quotes {
			key := quote.ProviderID + "|" + quote.ServiceKey
			if seen[key] {
				continue
			}
			seen[key] = true
			quotes = append(quotes, quote)
		}
	}

	metadata, _ := json.Marshal(map[string]any{
		"package":           req.Package,
		"speed":             req.Speed,
		"pickup_required":   req.PickupRequired,
		"delivery_required": req.DeliveryRequired,
	})
	logRow, err := a.svc.Q.CreateMarketplaceSearchLog(ctx, db.CreateMarketplaceSearchLogParams{
		CustomerID:       customerID,
		OriginState:      req.OriginState,
		OriginCity:       req.OriginCity,
		DestinationState: req.DestinationState,
		DestinationCity:  req.DestinationCity,
		SearchMetadata:   metadata,
		ResultCount:      int32(len(quotes)),
		FailureCount:     int32(len(failures)),
	})
	if err != nil {
		return marketplaceSearchResult{}, err
	}

	persisted := make([]service.MarketplaceQuote, 0, len(quotes))
	for _, quote := range quotes {
		saved, err := a.persistMarketplaceQuote(ctx, logRow.ID, quote, req)
		if err != nil {
			a.log.Error("marketplace quote persistence failed", "reference", quote.QuoteReference, "error", err)
			failures = append(failures, marketplaceProviderFailure{Provider: quote.ProviderName, Code: "quote_failed"})
			continue
		}
		persisted = append(persisted, saved)
	}

	return marketplaceSearchResult{
		quotes:     persisted,
		searchID:   logRow.ID,
		considered: len(considered),
		failures:   failures,
	}, nil
}

func (a *API) runLastMileMarketplaceSearch(ctx context.Context, req service.MarketplaceRouteRequest, customerID *uuid.UUID) (marketplaceSearchResult, error) {
	rows, err := a.svc.Q.ListMarketplaceLastMileAreas(ctx, db.ListMarketplaceLastMileAreasParams{
		City:  req.DestinationCity,
		State: req.DestinationState,
	})
	if err != nil {
		return marketplaceSearchResult{}, err
	}

	expiresAt := time.Now().Add(service.MarketplaceQuoteExpiry)
	quotes := make([]service.MarketplaceQuote, 0, len(rows))
	quoteIndex := make(map[string]int, len(rows))
	failures := make([]marketplaceProviderFailure, 0)
	failedProviders := make(map[string]bool)
	considered := make(map[string]bool, len(rows))

	for _, row := range rows {
		provider := lastMileProviderRateFromRow(row, req.DistanceKm)
		considered[provider.CompanyID] = true
		quote, quoteErr := service.BuildLastMileQuote(provider, req, expiresAt)
		if quoteErr != nil {
			a.log.Warn("last-mile provider quote failed", "company_id", provider.CompanyID, "area_id", provider.AreaID, "error", quoteErr)
			if !failedProviders[provider.CompanyID] {
				failedProviders[provider.CompanyID] = true
				failures = append(failures, marketplaceProviderFailure{Provider: provider.CompanyName, Code: quoteFailureCode(quoteErr)})
			}
			continue
		}
		if index, exists := quoteIndex[provider.CompanyID]; exists {
			if quote.TotalPriceMinor < quotes[index].TotalPriceMinor {
				quotes[index] = quote
			}
			continue
		}
		quoteIndex[provider.CompanyID] = len(quotes)
		quotes = append(quotes, quote)
	}

	metadata, _ := json.Marshal(map[string]any{
		"service_type":      "last_mile",
		"package":           req.Package,
		"pickup_address":    req.PickupAddress,
		"dropoff_address":   req.DropoffAddress,
		"distance_km":       req.DistanceKm,
		"pickup_required":   req.PickupRequired,
		"delivery_required": req.DeliveryRequired,
	})
	logRow, err := a.svc.Q.CreateMarketplaceSearchLog(ctx, db.CreateMarketplaceSearchLogParams{
		CustomerID:       customerID,
		OriginState:      req.OriginState,
		OriginCity:       req.OriginCity,
		DestinationState: req.DestinationState,
		DestinationCity:  req.DestinationCity,
		SearchMetadata:   metadata,
		ResultCount:      int32(len(quotes)),
		FailureCount:     int32(len(failures)),
	})
	if err != nil {
		return marketplaceSearchResult{}, err
	}

	persisted := make([]service.MarketplaceQuote, 0, len(quotes))
	for _, quote := range quotes {
		saved, persistErr := a.persistMarketplaceQuote(ctx, logRow.ID, quote, req)
		if persistErr != nil {
			a.log.Error("last-mile quote persistence failed", "reference", quote.QuoteReference, "error", persistErr)
			failures = append(failures, marketplaceProviderFailure{Provider: quote.ProviderName, Code: "quote_failed"})
			continue
		}
		persisted = append(persisted, saved)
	}

	return marketplaceSearchResult{
		quotes:     persisted,
		searchID:   logRow.ID,
		considered: len(considered),
		failures:   failures,
	}, nil
}

func lastMileProviderRateFromRow(row db.ListMarketplaceLastMileAreasRow, distanceKm float64) service.LastMileProviderRate {
	publicDescription := ""
	if row.PublicDescription != nil {
		publicDescription = strings.TrimSpace(*row.PublicDescription)
	}
	state, city := "", ""
	if row.State != nil {
		state = *row.State
	}
	if row.City != nil {
		city = *row.City
	}
	return service.LastMileProviderRate{
		CompanyID:            row.CompanyID.String(),
		CompanyName:          row.CompanyName,
		CompanySlug:          row.CompanySlug,
		BookingSlug:          row.BookingSlug,
		LogoURL:              row.LogoUrl,
		PublicDescription:    publicDescription,
		AreaID:               row.AreaID.String(),
		AreaName:             row.AreaName,
		State:                state,
		City:                 city,
		PricingModel:         row.PricingModel,
		FlatPriceMinor:       int8Pointer(row.FlatPriceMinor),
		BaseDeliveryFeeMinor: int8Pointer(row.BaseDeliveryFeeMinor),
		PricePerKmMinor:      int8Pointer(row.PricePerKmMinor),
		MinimumFeeMinor:      int8Pointer(row.MinimumFeeMinor),
		MaximumDistanceKm:    numericPointer(row.MaximumDistanceKm),
		MaximumFeeMinor:      int8Pointer(row.MaximumFeeMinor),
		VolumetricDivisor:    numericToFloat64(row.VolumetricDivisor),
		PriceIsEstimate:      row.PriceIsEstimate,
		MaxWeightKg:          numericPointer(row.MaxWeightKg),
		MaxLengthCM:          numericPointer(row.MaxLengthCm),
		MaxWidthCM:           numericPointer(row.MaxWidthCm),
		MaxHeightCM:          numericPointer(row.MaxHeightCm),
		ExpressAvailable:     row.ExpressAvailable,
		PickupAvailable:      row.PickupAvailable,
		DeliveryAvailable:    row.DeliveryAvailable,
		DistanceKm:           distanceKm,
		AcceptedCategories:   row.AcceptedItemCategories,
		ProhibitedCategories: row.ProhibitedItemCategories,
	}
}

func quoteFailureCode(err error) string {
	switch {
	case errors.Is(err, service.ErrWeightExceeded), errors.Is(err, service.ErrParcelTooLarge):
		return "parcel_restriction"
	case errors.Is(err, service.ErrCategoryNotAllowed):
		return "category_not_allowed"
	case errors.Is(err, service.ErrRouteUnavailable):
		return "route_unavailable"
	case errors.Is(err, service.ErrNoRateConfigured):
		return "no_matching_rate"
	default:
		return "quote_failed"
	}
}

func providerRateFromRow(row db.ListMarketplaceRoutesForLaneRow) service.ProviderRate {
	return service.ProviderRate{
		CompanyID:            row.CompanyID.String(),
		CompanyName:          row.CompanyName,
		CompanySlug:          row.CompanySlug,
		BookingSlug:          row.BookingSlug,
		LogoURL:              row.LogoUrl,
		PublicDescription:    row.PublicDescription,
		RouteID:              row.RouteID.String(),
		OriginState:          row.OriginState,
		OriginCity:           row.OriginCity,
		DestinationState:     row.DestinationState,
		DestinationCity:      row.DestinationCity,
		DistanceKm:           numericToFloat64(row.DistanceKm),
		StandardPerKmMinor:   int8Pointer(row.StandardPricePerKmMinor),
		ExpressPerKmMinor:    int8Pointer(row.ExpressPricePerKmMinor),
		SameDayPerKmMinor:    int8Pointer(row.SameDayPricePerKmMinor),
		MinimumChargeMinor:   row.MinimumChargeMinor,
		EstimatedTransitTime: row.EstimatedTransitTime,
		DepartureSchedule:    row.DepartureSchedule,
		PickupAvailable:      row.PickupAvailable,
		DeliveryAvailable:    row.DeliveryAvailable,
		Description:          row.Description,
		MaxWeightKg:          numericPointer(row.MaxWeightKg),
		MaxLengthCM:          numericPointer(row.MaxLengthCm),
		MaxWidthCM:           numericPointer(row.MaxWidthCm),
		MaxHeightCM:          numericPointer(row.MaxHeightCm),
		PriceIsEstimate:      row.PriceIsEstimate,
		AcceptedCategories:   row.AcceptedItemCategories,
		ProhibitedCategories: row.ProhibitedItemCategories,
	}
}

func numericPointer(v pgtype.Numeric) *float64 {
	if !v.Valid {
		return nil
	}
	value := numericToFloat64(v)
	if value <= 0 {
		return nil
	}
	return &value
}

func (a *API) persistMarketplaceQuote(ctx context.Context, searchID uuid.UUID, quote service.MarketplaceQuote, req service.MarketplaceRouteRequest) (service.MarketplaceQuote, error) {
	packageSnapshot, err := json.Marshal(map[string]any{
		"package":           req.Package,
		"pickup_required":   req.PickupRequired,
		"delivery_required": req.DeliveryRequired,
	})
	if err != nil {
		return service.MarketplaceQuote{}, err
	}
	breakdown, err := json.Marshal(quote.Breakdown)
	if err != nil {
		return service.MarketplaceQuote{}, err
	}
	routeID, _ := uuid.Parse(quote.RouteID)
	rateSnapshot, err := json.Marshal(map[string]any{
		"route_id":          quote.RouteID,
		"distance_km":       quote.DistanceKm,
		"price_is_estimate": quote.PriceConfirmationType == "estimated",
		"captured_at":       time.Now().UTC(),
	})
	if err != nil {
		return service.MarketplaceQuote{}, err
	}

	row, err := a.svc.Q.InsertMarketplaceQuote(ctx, db.InsertMarketplaceQuoteParams{
		QuoteReference:         quote.QuoteReference,
		SearchID:               &searchID,
		CompanyID:              mustUUID(quote.ProviderID),
		ProviderName:           quote.ProviderName,
		ProviderLogo:           quote.ProviderLogo,
		ProviderSlug:           quote.ProviderSlug,
		RouteID:                nullableUUID(routeID),
		OriginState:            quote.OriginState,
		OriginCity:             quote.OriginCity,
		DestinationState:       quote.DestinationState,
		DestinationCity:        quote.DestinationCity,
		ServiceKey:             quote.ServiceKey,
		ServiceName:            quote.ServiceName,
		DistanceKm:             float64ToNumeric(quote.DistanceKm),
		PackageSnapshot:        packageSnapshot,
		ChargeableWeightKg:     float64ToNumeric(quote.ChargeableWeightKg),
		BasePriceMinor:         quote.BasePriceMinor,
		AdditionalChargesMinor: quote.AdditionalChargesMinor,
		DiscountMinor:          quote.DiscountMinor,
		TotalPriceMinor:        quote.TotalPriceMinor,
		Currency:               quote.Currency,
		PriceConfirmationType:  quote.PriceConfirmationType,
		PriceBreakdown:         breakdown,
		RateSnapshot:           rateSnapshot,
		EstimatedDelivery:      optionalNonEmpty(quote.EstimatedDelivery),
		EstimatedDurationDays:  pgtype.Int4{Int32: int32(quote.EstimatedDurationDays), Valid: true},
		PickupAvailable:        quote.PickupAvailable,
		DeliveryAvailable:      quote.DeliveryAvailable,
		DeliveryMethod:         quote.DeliveryMethod,
		Conditions:             optionalNonEmpty(quote.Conditions),
		ExpiresAt:              quote.ExpiresAt,
	})
	if err != nil {
		return service.MarketplaceQuote{}, err
	}
	saved := quoteFromDB(row)
	saved.BookingSlug = quote.BookingSlug
	return saved, nil
}

func optionalNonEmpty(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func mustUUID(value string) uuid.UUID {
	id, _ := uuid.Parse(value)
	return id
}

func nullableUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func quoteFromDB(row db.MarketplaceQuote) service.MarketplaceQuote {
	quote := service.MarketplaceQuote{
		QuoteReference:         row.QuoteReference,
		ProviderID:             row.CompanyID.String(),
		ProviderName:           row.ProviderName,
		ProviderLogo:           row.ProviderLogo,
		ProviderSlug:           row.ProviderSlug,
		ServiceKey:             row.ServiceKey,
		ServiceName:            row.ServiceName,
		OriginState:            row.OriginState,
		OriginCity:             row.OriginCity,
		DestinationState:       row.DestinationState,
		DestinationCity:        row.DestinationCity,
		BasePriceMinor:         row.BasePriceMinor,
		ChargeableWeightKg:     numericToFloat64(row.ChargeableWeightKg),
		AdditionalChargesMinor: row.AdditionalChargesMinor,
		DiscountMinor:          row.DiscountMinor,
		TotalPriceMinor:        row.TotalPriceMinor,
		Currency:               row.Currency,
		EstimatedDurationDays:  int(row.EstimatedDurationDays.Int32),
		PickupAvailable:        row.PickupAvailable,
		DeliveryAvailable:      row.DeliveryAvailable,
		DeliveryMethod:         row.DeliveryMethod,
		PriceConfirmationType:  row.PriceConfirmationType,
		Status:                 row.Status,
		ExpiresAt:              row.ExpiresAt,
		DistanceKm:             numericToFloat64(row.DistanceKm),
	}
	if row.EstimatedDelivery != nil {
		quote.EstimatedDelivery = *row.EstimatedDelivery
	}
	if row.Conditions != nil {
		quote.Conditions = *row.Conditions
	}
	_ = json.Unmarshal(row.PriceBreakdown, &quote.Breakdown)
	if quote.PriceConfirmationType == "estimated" {
		quote.PriceConfirmationNote = "Estimated price. Final price will be confirmed by the logistics provider after package inspection."
	}
	if time.Now().After(row.ExpiresAt) && quote.Status == "active" {
		quote.Status = "expired"
	}
	return quote
}

func (a *API) marketplaceQuote(c *gin.Context) {
	row, err := a.svc.Q.GetMarketplaceQuoteByReference(c.Request.Context(), strings.TrimSpace(c.Param("reference")))
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "quote_not_found", "That quote was not found. Run the search again.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	quote := quoteFromDB(row)
	c.JSON(http.StatusOK, gin.H{"quote": quote, "can_select": quote.Status == "active"})
}

func (a *API) selectMarketplaceQuote(c *gin.Context) {
	var req struct {
		QuoteReference string `json:"quote_reference" binding:"required,min=8,max=60"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_selection", "Choose a valid quote.")
		return
	}
	row, err := a.svc.Q.GetMarketplaceQuoteByReference(c.Request.Context(), strings.TrimSpace(req.QuoteReference))
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "quote_not_found", "That quote was not found. Run the search again.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	if time.Now().After(row.ExpiresAt) || row.Status == "expired" {
		fail(c, http.StatusGone, "quote_expired", "This quote has expired. Search again for the current price.")
		return
	}

	user, signedIn := currentUser(c)
	var customerID *uuid.UUID
	var guestHash *string
	if signedIn {
		customerID = &user.ID
	} else {
		token, err := a.marketplaceGuestToken(c)
		if err != nil {
			a.handleServiceError(c, err)
			return
		}
		hash := sha256.Sum256([]byte(token))
		encoded := hex.EncodeToString(hash[:])
		guestHash = &encoded
	}

	_, err = a.svc.Q.InsertSavedQuote(c.Request.Context(), db.InsertSavedQuoteParams{
		CustomerID:     customerID,
		GuestTokenHash: guestHash,
		QuoteID:        row.ID,
	})
	if err != nil && err != pgx.ErrNoRows {
		a.handleServiceError(c, err)
		return
	}
	_ = a.svc.Q.MarkMarketplaceQuoteSelected(c.Request.Context(), row.ID)

	quote := quoteFromDB(row)
	quote.Status = "selected"
	c.JSON(http.StatusOK, gin.H{"quote": quote, "saved": true})
}

func (a *API) marketplaceGuestToken(c *gin.Context) (string, error) {
	if existing, err := c.Cookie(marketplaceGuestCookie); err == nil && len(existing) >= 32 {
		return existing, nil
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     marketplaceGuestCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   a.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	return token, nil
}

func (a *API) listCustomerQuotes(c *gin.Context) {
	user, _ := currentUser(c)
	_ = a.svc.Q.ExpireMarketplaceQuotes(c.Request.Context())
	rows, err := a.svc.Q.ListCustomerSavedQuotes(c.Request.Context(), &user.ID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	quotes := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		quote := quoteFromDB(db.MarketplaceQuote{
			ID: row.ID, QuoteReference: row.QuoteReference, SearchID: row.SearchID,
			CompanyID: row.CompanyID, ProviderName: row.ProviderName, ProviderLogo: row.ProviderLogo,
			ProviderSlug: row.ProviderSlug, RouteID: row.RouteID, OriginState: row.OriginState,
			OriginCity: row.OriginCity, DestinationState: row.DestinationState, DestinationCity: row.DestinationCity,
			ServiceKey: row.ServiceKey, ServiceName: row.ServiceName, DistanceKm: row.DistanceKm,
			PackageSnapshot: row.PackageSnapshot, ChargeableWeightKg: row.ChargeableWeightKg,
			BasePriceMinor: row.BasePriceMinor, AdditionalChargesMinor: row.AdditionalChargesMinor,
			DiscountMinor: row.DiscountMinor, TotalPriceMinor: row.TotalPriceMinor, Currency: row.Currency,
			PriceConfirmationType: row.PriceConfirmationType, PriceBreakdown: row.PriceBreakdown,
			RateSnapshot: row.RateSnapshot, EstimatedDelivery: row.EstimatedDelivery,
			EstimatedDurationDays: row.EstimatedDurationDays, PickupAvailable: row.PickupAvailable,
			DeliveryAvailable: row.DeliveryAvailable, DeliveryMethod: row.DeliveryMethod,
			Conditions: row.Conditions, Status: row.Status, ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
		})
		quotes = append(quotes, gin.H{
			"saved_quote_id":   row.SavedQuoteID,
			"selection_status": row.SelectionStatus,
			"selected_at":      row.SelectedAt,
			"quote":            quote,
			"expired":          quote.Status == "expired",
		})
	}
	c.JSON(http.StatusOK, gin.H{"quotes": quotes})
}

func (a *API) getCustomerQuote(c *gin.Context) {
	user, _ := currentUser(c)
	row, err := a.svc.Q.GetCustomerSavedQuoteByReference(c.Request.Context(), db.GetCustomerSavedQuoteByReferenceParams{
		CustomerID:     &user.ID,
		QuoteReference: strings.TrimSpace(c.Param("reference")),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "quote_not_found", "That saved quote was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	quote := quoteFromDB(db.MarketplaceQuote{
		ID: row.ID, QuoteReference: row.QuoteReference, SearchID: row.SearchID,
		CompanyID: row.CompanyID, ProviderName: row.ProviderName, ProviderLogo: row.ProviderLogo,
		ProviderSlug: row.ProviderSlug, RouteID: row.RouteID, OriginState: row.OriginState,
		OriginCity: row.OriginCity, DestinationState: row.DestinationState, DestinationCity: row.DestinationCity,
		ServiceKey: row.ServiceKey, ServiceName: row.ServiceName, DistanceKm: row.DistanceKm,
		PackageSnapshot: row.PackageSnapshot, ChargeableWeightKg: row.ChargeableWeightKg,
		BasePriceMinor: row.BasePriceMinor, AdditionalChargesMinor: row.AdditionalChargesMinor,
		DiscountMinor: row.DiscountMinor, TotalPriceMinor: row.TotalPriceMinor, Currency: row.Currency,
		PriceConfirmationType: row.PriceConfirmationType, PriceBreakdown: row.PriceBreakdown,
		RateSnapshot: row.RateSnapshot, EstimatedDelivery: row.EstimatedDelivery,
		EstimatedDurationDays: row.EstimatedDurationDays, PickupAvailable: row.PickupAvailable,
		DeliveryAvailable: row.DeliveryAvailable, DeliveryMethod: row.DeliveryMethod,
		Conditions: row.Conditions, Status: row.Status, ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
	})
	c.JSON(http.StatusOK, gin.H{
		"saved_quote_id":   row.SavedQuoteID,
		"selection_status": row.SelectionStatus,
		"selected_at":      row.SelectedAt,
		"quote":            quote,
	})
}

func (a *API) selectCustomerQuote(c *gin.Context) {
	user, _ := currentUser(c)
	row, err := a.svc.Q.GetCustomerSavedQuoteByReference(c.Request.Context(), db.GetCustomerSavedQuoteByReferenceParams{
		CustomerID:     &user.ID,
		QuoteReference: strings.TrimSpace(c.Param("reference")),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			fail(c, http.StatusNotFound, "quote_not_found", "That saved quote was not found.")
			return
		}
		a.handleServiceError(c, err)
		return
	}
	if time.Now().After(row.ExpiresAt) || row.Status == "expired" {
		fail(c, http.StatusGone, "quote_expired", "This quote has expired. Search again for the current price.")
		return
	}
	_ = a.svc.Q.MarkMarketplaceQuoteSelected(c.Request.Context(), row.ID)
	c.JSON(http.StatusOK, gin.H{"selected": true, "quote_reference": row.QuoteReference})
}

// ---- Provider listing settings ----

type marketplaceSettingsDTO struct {
	IsListed          bool       `json:"is_listed"`
	VisibilityStatus  string     `json:"visibility_status"`
	PublicDescription string     `json:"public_description"`
	PublishedAt       *time.Time `json:"published_at"`
}

func (a *API) getCompanyMarketplaceSettings(c *gin.Context) {
	tenant, _ := currentTenant(c)
	row, err := a.svc.Q.GetMarketplaceSettings(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"marketplace": marketplaceSettingsDTO{VisibilityStatus: "draft"}})
			return
		}
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"marketplace": marketplaceSettingsDTO{
		IsListed:          row.IsListed,
		VisibilityStatus:  row.VisibilityStatus,
		PublicDescription: valueOrString(row.PublicDescription),
		PublishedAt:       row.PublishedAt,
	}})
}

func (a *API) updateCompanyMarketplaceSettings(c *gin.Context) {
	tenant, _ := currentTenant(c)
	company, err := a.svc.Q.GetCompanyByID(c.Request.Context(), tenant.CompanyID)
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	var req struct {
		IsListed          bool   `json:"is_listed"`
		PublicDescription string `json:"public_description" binding:"max=600"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid_marketplace_settings", "Check the marketplace description.")
		return
	}
	description := strings.TrimSpace(req.PublicDescription)
	if req.IsListed && company.Status != "verified" {
		fail(c, http.StatusForbidden, "provider_not_verified", "Your company must be verified before it can be listed.")
		return
	}
	status := "paused"
	if req.IsListed {
		status = "listed"
	}
	var publishedAt *time.Time
	if req.IsListed {
		now := time.Now()
		publishedAt = &now
	}
	row, err := a.svc.Q.UpsertMarketplaceSettings(c.Request.Context(), db.UpsertMarketplaceSettingsParams{
		CompanyID:         tenant.CompanyID,
		IsListed:          req.IsListed,
		VisibilityStatus:  status,
		PublicDescription: optionalNonEmpty(description),
		PublishedAt:       publishedAt,
	})
	if err != nil {
		a.handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"marketplace": marketplaceSettingsDTO{
		IsListed:          row.IsListed,
		VisibilityStatus:  row.VisibilityStatus,
		PublicDescription: valueOrString(row.PublicDescription),
		PublishedAt:       row.PublishedAt,
	}})
}

func valueOrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
