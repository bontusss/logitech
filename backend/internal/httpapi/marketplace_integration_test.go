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
		ServiceKey            string `json:"service_key"`
		TotalPriceMinor       int64  `json:"total_price_minor"`
		PriceConfirmationType string `json:"price_confirmation_type"`
		Status                string `json:"status"`
	} `json:"quotes"`
	Meta struct {
		ProvidersConsidered int  `json:"providers_considered"`
		ResultCount         int  `json:"result_count"`
		Partial             bool `json:"partial"`
	} `json:"meta"`
}

func TestMarketplaceSearchSelectionAndOwnership(t *testing.T) {
	env := newTestEnv(t)

	owner := registerCompany(t, env, "Marketplace Freight", "market-owner@example.com")
	companyID := owner.Companies[0].CompanyID

	_, err := env.pool.Exec(t.Context(), `
		UPDATE companies SET status = 'verified', onboarding_status = 'complete', verified_at = now()
		WHERE id = $1`, companyID)
	require.NoError(t, err)

	rec := do(t, env.router, http.MethodPost, "/api/v1/companies/me/booking-routes", owner.AccessToken, companyID, map[string]any{
		"origin_state": "Abia", "origin_city": "Aba",
		"destination_state": "Lagos", "destination_city": "Lagos",
		"distance_km": 500, "standard_price_per_km": 100,
		"express_price_per_km": 180, "minimum_charge": 5000,
	})
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	searchBody := map[string]any{
		"origin_state": "Abia", "origin_city": "Aba",
		"destination_state": "Lagos", "destination_city": "Lagos",
		"weight_kg": 2, "category": "parcel",
	}

	// The provider is verified but not listed yet, so it must not surface.
	rec = do(t, env.router, http.MethodPost, "/api/v1/marketplace/search", "", "", searchBody)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var hidden marketplaceSearchResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &hidden))
	require.Empty(t, hidden.Quotes)
	require.Equal(t, 0, hidden.Meta.ProvidersConsidered)

	rec = do(t, env.router, http.MethodPut, "/api/v1/companies/me/marketplace", owner.AccessToken, companyID, map[string]any{
		"is_listed": true, "public_description": "Intercity parcel delivery across Nigeria.",
	})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/marketplace/locations", "", "", nil)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Contains(t, rec.Body.String(), "Aba")

	rec = do(t, env.router, http.MethodPost, "/api/v1/marketplace/search", "", "", searchBody)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var results marketplaceSearchResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &results))
	require.Len(t, results.Quotes, 2)
	require.Equal(t, 1, results.Meta.ProvidersConsidered)
	require.False(t, results.Meta.Partial)

	byService := map[string]int64{}
	for _, quote := range results.Quotes {
		byService[quote.ServiceKey] = quote.TotalPriceMinor
		require.Equal(t, "estimated", quote.PriceConfirmationType)
		require.Equal(t, "active", quote.Status)
	}
	require.Equal(t, int64(5_000_000), byService["standard"])
	require.Equal(t, int64(9_000_000), byService["express"])

	reference := results.Quotes[0].QuoteReference

	// A guest can select and view the quote without an account.
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

func TestMarketplaceRejectsInvalidAndExpiredQuotes(t *testing.T) {
	env := newTestEnv(t)

	rec := do(t, env.router, http.MethodPost, "/api/v1/marketplace/search", "", "", map[string]any{
		"origin_state": "Abia", "origin_city": "Aba",
		"destination_state": "Lagos", "destination_city": "Lagos",
		"weight_kg": 0,
	})
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPost, "/api/v1/marketplace/search", "", "", map[string]any{
		"origin_state": "Abia", "origin_city": "Aba",
		"destination_state": "Abia", "destination_city": "Aba",
		"weight_kg": 2,
	})
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodGet, "/api/v1/marketplace/quotes/BOL-MKT-MISSING", "", "", nil)
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())

	rec = do(t, env.router, http.MethodPost, "/api/v1/marketplace/quotes", "", "", map[string]any{
		"quote_reference": "BOL-MKT-MISSING",
	})
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
}
