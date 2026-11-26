package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type marketplaceSearchResponse struct {
	Quotes []struct {
		QuoteReference        string `json:"quote_reference"`
		ProviderName          string `json:"provider_name"`
		ProviderSlug          string `json:"provider_slug"`
		ServiceMode           string `json:"service_mode"`
		ServiceKey            string `json:"service_key"`
		OriginCity            string `json:"origin_city"`
		OriginState           string `json:"origin_state"`
		OriginZoneName        string `json:"origin_zone_name"`
		DestinationCity       string `json:"destination_city"`
		DestinationZoneName   string `json:"destination_zone_name"`
		DropoffBranchName     string `json:"dropoff_branch_name"`
		DropoffBranchAddress  string `json:"dropoff_branch_address"`
		TotalPriceMinor       int64  `json:"total_price_minor"`
		PriceConfirmationType string `json:"price_confirmation_type"`
		EstimatedPickup       string `json:"estimated_pickup"`
		EstimatedDelivery     string `json:"estimated_delivery"`
		DeliveryMethod        string `json:"delivery_method"`
		Status                string `json:"status"`
	} `json:"quotes"`
	Meta struct {
		ProvidersConsidered int  `json:"providers_considered"`
		ResultCount         int  `json:"result_count"`
		Partial             bool `json:"partial"`
	} `json:"meta"`
}

type geoZonesResponse struct {
	Zones []struct {
		ID       string  `json:"id"`
		State    string  `json:"state"`
		Lga      string  `json:"lga"`
		ParentID *string `json:"parent_id"`
	} `json:"zones"`
}

func createServiceZone(t *testing.T, env *testEnv, token, companyID, geoZoneID string) string {
	t.Helper()
	rec := do(t, env.router, http.MethodPost, "/api/v1/companies/me/service-zones", token, companyID, map[string]any{
		"geo_zone_id": geoZoneID,
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var parsed struct {
		Zone struct {
			ID string `json:"id"`
		} `json:"zone"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &parsed))
	require.NotEmpty(t, parsed.Zone.ID)
	return parsed.Zone.ID
}

func marketplaceSearch(t *testing.T, env *testEnv, body map[string]any) marketplaceSearchResponse {
	t.Helper()
	rec := do(t, env.router, http.MethodPost, "/api/v1/marketplace/search", "", "", body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var parsed marketplaceSearchResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &parsed))
	return parsed
}

func TestMarketplaceSearchSelectionAndOwnership(t *testing.T) {
	env := newTestEnv(t)

	owner := registerCompany(t, env, "Marketplace Freight", "market-owner@example.com")
	companyID := owner.Companies[0].CompanyID

	_, err := env.pool.Exec(t.Context(), `
		UPDATE companies SET status = 'verified', onboarding_status = 'complete', verified_at = now()
		WHERE id = $1`, companyID)
	require.NoError(t, err)

	rec := do(t, env.router, http.MethodGet, "/api/v1/companies/me/geo/zones?state=Abia", owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var zones geoZonesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &zones))
	var originGeoZone, destinationGeoZone string
	for _, zone := range zones.Zones {
		if zone.ParentID != nil {
			continue
		}
		switch zone.Lga {
		case "Aba South":
			originGeoZone = zone.ID
		case "Umuahia North":
			destinationGeoZone = zone.ID
		}
	}
	require.NotEmpty(t, originGeoZone)
	require.NotEmpty(t, destinationGeoZone)

	originZoneID := createServiceZone(t, env, owner.AccessToken, companyID, originGeoZone)
	destinationZoneID := createServiceZone(t, env, owner.AccessToken, companyID, destinationGeoZone)

	routeBody := map[string]any{
		"from_zone_id":                originZoneID,
		"to_zone_id":                  destinationZoneID,
		"price_minor":                 500_000,
		"included_weight_kg":          5,
		"additional_weight_fee_minor": 10_000,
		"pickup_available":            true,
		"dropoff_available":           true,
		"estimated_pickup_hours":      2,
		"estimated_delivery_hours":    30,
	}
	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/zone-routes", owner.AccessToken, companyID, routeBody)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var route struct {
		Route struct {
			ID string `json:"id"`
		} `json:"route"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &route))
	require.NotEmpty(t, route.Route.ID)

	// Listing now requires an active zone route, which is already published above.
	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/marketplace", owner.AccessToken, companyID, map[string]any{
		"is_listed": true, "public_description": "Intercity parcel delivery across Nigeria.",
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/marketplace/locations", "", "", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	pickupSearch := map[string]any{
		"service_mode":          "pickup",
		"origin_zone_id":        originGeoZone,
		"destination_zone_id":   destinationGeoZone,
		"origin_area":           "Ariaria",
		"destination_area":      "Umuahia",
		"origin_latitude":       5.1065,
		"origin_longitude":      7.3667,
		"destination_latitude":  5.5250,
		"destination_longitude": 7.5100,
		"weight_kg":             2,
		"category":              "other",
	}
	results := marketplaceSearch(t, env, pickupSearch)
	require.Len(t, results.Quotes, 1)
	require.Equal(t, 1, results.Meta.ProvidersConsidered)
	require.False(t, results.Meta.Partial)
	require.Equal(t, "zone_route", results.Quotes[0].ServiceKey)
	require.Equal(t, "pickup", results.Quotes[0].ServiceMode)
	require.Equal(t, int64(500_000), results.Quotes[0].TotalPriceMinor)
	require.Equal(t, "confirmed", results.Quotes[0].PriceConfirmationType)
	require.Equal(t, "Pickup within 2h", results.Quotes[0].EstimatedPickup)
	require.Equal(t, "Delivery in 2 days", results.Quotes[0].EstimatedDelivery)
	require.Equal(t, "door-to-door", results.Quotes[0].DeliveryMethod)
	require.Equal(t, "active", results.Quotes[0].Status)

	reference := results.Quotes[0].QuoteReference

	// A drop-off search starts at the hub zone the sender is willing to visit
	// and still needs the destination zone the parcel is going to.
	dropoffSearch := map[string]any{
		"service_mode":          "dropoff",
		"origin_zone_id":        originGeoZone,
		"destination_zone_id":   destinationGeoZone,
		"destination_area":      "Umuahia",
		"destination_latitude":  5.5250,
		"destination_longitude": 7.5100,
		"weight_kg":             2,
		"category":              "other",
	}
	dropoffResults := marketplaceSearch(t, env, dropoffSearch)
	require.Len(t, dropoffResults.Quotes, 1)
	require.Equal(t, "dropoff", dropoffResults.Quotes[0].ServiceMode)
	require.Equal(t, "Aba South", dropoffResults.Quotes[0].OriginCity)
	require.Equal(t, "terminal-to-door", dropoffResults.Quotes[0].DeliveryMethod)

	// Disabling pickup removes the route from pickup searches only.
	updatedRoute := map[string]any{
		"from_zone_id":                originZoneID,
		"to_zone_id":                  destinationZoneID,
		"price_minor":                 500_000,
		"included_weight_kg":          5,
		"additional_weight_fee_minor": 10_000,
		"pickup_available":            false,
		"dropoff_available":           true,
		"estimated_pickup_hours":      2,
		"estimated_delivery_hours":    30,
		"status":                      "active",
	}
	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/zone-routes/"+route.Route.ID, owner.AccessToken, companyID, updatedRoute)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	require.Empty(t, marketplaceSearch(t, env, pickupSearch).Quotes)
	require.Len(t, marketplaceSearch(t, env, dropoffSearch).Quotes, 1)

	// A guest can select and view a quote without an account.
	rec = do(t, env.router, http.MethodPost, "/api/v1/marketplace/quotes", "", "", map[string]any{
		"quote_reference": reference,
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.NotEmpty(t, cookieValue(t, rec, "bol_marketplace_guest"))

	rec = do(t, env.router, http.MethodGet, "/api/v1/marketplace/quotes/"+reference, "", "", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "selected")

	customerA := registerCustomer(t, env, "market-a@example.com")
	customerB := registerCustomer(t, env, "market-b@example.com")

	rec = do(t, env.router, http.MethodPost, "/api/v1/marketplace/quotes", customerA.AccessToken, "", map[string]any{
		"quote_reference": reference,
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/customer/quotes", customerA.AccessToken, "", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), reference)

	// Another customer cannot read or select the same saved quote.
	rec = do(t, env.router, http.MethodGet, "/api/v1/customer/quotes/"+reference, customerB.AccessToken, "", nil)
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPost, "/api/v1/customer/quotes/"+reference+"/select", customerB.AccessToken, "", nil)
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
}

// Area zones only match when the customer picks that neighbourhood. Searching
// from the whole LGA must not fall back to a route limited to one area.
func TestMarketplaceAreaZonesRestrictRoutes(t *testing.T) {
	env := newTestEnv(t)

	owner := registerCompany(t, env, "Area Freight", "area-owner@example.com")
	companyID := owner.Companies[0].CompanyID

	_, err := env.pool.Exec(t.Context(), `
		UPDATE companies SET status = 'verified', onboarding_status = 'complete', verified_at = now()
		WHERE id = $1`, companyID)
	require.NoError(t, err)

	rec := do(t, env.router, http.MethodGet, "/api/v1/companies/me/geo/zones?state=Abia", owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var zones geoZonesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &zones))
	var originGeoZone, destinationGeoZone string
	for _, zone := range zones.Zones {
		if zone.ParentID != nil {
			continue
		}
		switch zone.Lga {
		case "Aba South":
			originGeoZone = zone.ID
		case "Umuahia North":
			destinationGeoZone = zone.ID
		}
	}
	require.NotEmpty(t, originGeoZone)
	require.NotEmpty(t, destinationGeoZone)

	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/service-zones", owner.AccessToken, companyID, map[string]any{
		"geo_zone_id": originGeoZone,
		"area_name":   "Ariaria",
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var areaZone struct {
		Zone struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			GeoZoneID string `json:"geo_zone_id"`
		} `json:"zone"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &areaZone))
	require.Equal(t, "Ariaria", areaZone.Zone.Name)
	require.NotEmpty(t, areaZone.Zone.GeoZoneID)
	require.NotEqual(t, originGeoZone, areaZone.Zone.GeoZoneID)

	destinationZoneID := createServiceZone(t, env, owner.AccessToken, companyID, destinationGeoZone)

	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/branches", owner.AccessToken, companyID, map[string]any{
		"name": "Ariaria Hub", "code": "ARI", "address": "12 Market Road",
		"city": "Aba South", "state": "Abia",
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var branch struct {
		Branch struct {
			ID string `json:"id"`
		} `json:"branch"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &branch))
	require.NotEmpty(t, branch.Branch.ID)

	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/zone-routes", owner.AccessToken, companyID, map[string]any{
		"from_zone_id":                areaZone.Zone.ID,
		"to_zone_id":                  destinationZoneID,
		"price_minor":                 500_000,
		"included_weight_kg":          5,
		"additional_weight_fee_minor": 10_000,
		"pickup_available":            true,
		"dropoff_available":           true,
		"dropoff_branch_id":           branch.Branch.ID,
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/marketplace", owner.AccessToken, companyID, map[string]any{
		"is_listed": true, "public_description": "Area-limited parcel delivery.",
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	baseSearch := map[string]any{
		"service_mode":          "pickup",
		"destination_zone_id":   destinationGeoZone,
		"destination_latitude":  5.5250,
		"destination_longitude": 7.5100,
		"origin_latitude":       5.1065,
		"origin_longitude":      7.3667,
		"weight_kg":             2,
		"category":              "other",
	}

	lgaSearch := map[string]any{"origin_zone_id": originGeoZone}
	for key, value := range baseSearch {
		lgaSearch[key] = value
	}
	parsed := marketplaceSearch(t, env, lgaSearch)
	require.Empty(t, parsed.Quotes, "an LGA search must not match an area-only route")

	areaSearch := map[string]any{"origin_zone_id": areaZone.Zone.GeoZoneID}
	for key, value := range baseSearch {
		areaSearch[key] = value
	}
	parsed = marketplaceSearch(t, env, areaSearch)
	require.Len(t, parsed.Quotes, 1)
	require.Equal(t, "Ariaria", parsed.Quotes[0].OriginZoneName)

	dropoff := marketplaceSearch(t, env, map[string]any{
		"service_mode":          "dropoff",
		"origin_zone_id":        originGeoZone,
		"destination_zone_id":   destinationGeoZone,
		"destination_latitude":  5.5250,
		"destination_longitude": 7.5100,
		"weight_kg":             2,
		"category":              "other",
	})
	require.Len(t, dropoff.Quotes, 1)
	require.Equal(t, "Ariaria Hub", dropoff.Quotes[0].DropoffBranchName)
	require.Equal(t, "12 Market Road", dropoff.Quotes[0].DropoffBranchAddress)
}

func TestMarketplaceListingRequiresActiveZoneRoute(t *testing.T) {
	env := newTestEnv(t)

	owner := registerCompany(t, env, "Listing Guard Freight", "listing-guard@example.com")
	companyID := owner.Companies[0].CompanyID
	_, err := env.pool.Exec(t.Context(), `
		UPDATE companies SET status = 'verified', onboarding_status = 'complete', verified_at = now()
		WHERE id = $1`, companyID)
	require.NoError(t, err)

	listBody := map[string]any{"is_listed": true, "public_description": "Coverage-first provider."}
	rec := do(t, env.router, http.MethodPut, "/api/v1/companies/me/marketplace", owner.AccessToken, companyID, listBody)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "zones_routes_required")

	rec = do(t, env.router, http.MethodGet, "/api/v1/companies/me/geo/zones?state=Abia", owner.AccessToken, companyID, nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var zones geoZonesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &zones))
	require.NotEmpty(t, zones.Zones)
	geoZoneID := zones.Zones[0].ID

	originZoneID := createServiceZone(t, env, owner.AccessToken, companyID, geoZoneID)
	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/marketplace", owner.AccessToken, companyID, listBody)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPost, "/api/v1/companies/me/zone-routes", owner.AccessToken, companyID, map[string]any{
		"from_zone_id":                originZoneID,
		"to_zone_id":                  originZoneID,
		"price_minor":                 250_000,
		"included_weight_kg":          5,
		"additional_weight_fee_minor": 5_000,
		"status":                      "inactive",
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var route struct {
		Route struct {
			ID string `json:"id"`
		} `json:"route"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &route))

	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/marketplace", owner.AccessToken, companyID, listBody)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPatch, "/api/v1/companies/me/zone-routes/"+route.Route.ID+"/status", owner.AccessToken, companyID, map[string]any{
		"status": "active",
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/marketplace", owner.AccessToken, companyID, listBody)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "\"is_listed\":true")
}

func TestMarketplaceRejectsInvalidAndExpiredQuotes(t *testing.T) {
	env := newTestEnv(t)

	rec := do(t, env.router, http.MethodPost, "/api/v1/marketplace/search", "", "", map[string]any{
		"service_mode":          "pickup",
		"origin_zone_id":        "00000000-0000-0000-0000-000000000001",
		"destination_zone_id":   "00000000-0000-0000-0000-000000000002",
		"origin_area":           "Ariaria",
		"destination_area":      "Umuahia",
		"origin_latitude":       5.1,
		"origin_longitude":      7.3,
		"destination_latitude":  5.5,
		"destination_longitude": 7.5,
		"weight_kg":             0,
	})
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())

	// A valid UUID that is not in the geography catalog is rejected.
	rec = do(t, env.router, http.MethodPost, "/api/v1/marketplace/search", "", "", map[string]any{
		"service_mode":          "dropoff",
		"destination_zone_id":   "00000000-0000-0000-0000-000000000002",
		"destination_area":      "Umuahia",
		"destination_latitude":  5.5,
		"destination_longitude": 7.5,
		"weight_kg":             2,
	})
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	// A destination pin is required for both service modes.
	rec = do(t, env.router, http.MethodPost, "/api/v1/marketplace/search", "", "", map[string]any{
		"service_mode":        "dropoff",
		"destination_zone_id": "00000000-0000-0000-0000-000000000002",
		"destination_area":    "Umuahia",
		"weight_kg":           2,
	})
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/marketplace/quotes/BOL-MKT-MISSING", "", "", nil)
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPost, "/api/v1/marketplace/quotes", "", "", map[string]any{
		"quote_reference": "BOL-MKT-MISSING",
	})
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
}
