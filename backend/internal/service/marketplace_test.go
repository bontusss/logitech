package service

import (
	"errors"
	"testing"
	"time"
)

func testProvider() ProviderRate {
	transit := "2 days"
	description := "Careful handling for fragile items."
	return ProviderRate{
		CompanyID:            "11111111-1111-1111-1111-111111111111",
		CompanyName:          "Aba Freight",
		CompanySlug:          "aba-freight",
		BookingSlug:          "aba-freight-booking",
		LogoURL:              "https://cdn.example/logo.png",
		RouteID:              "22222222-2222-2222-2222-222222222222",
		OriginState:          "Abia",
		OriginCity:           "Aba",
		DestinationState:     "Lagos",
		DestinationCity:      "Lagos",
		DistanceKm:           500,
		StandardPerKmMinor:   int64Ptr(100),
		ExpressPerKmMinor:    int64Ptr(180),
		MinimumChargeMinor:   5000,
		EstimatedTransitTime: &transit,
		PickupAvailable:      true,
		DeliveryAvailable:    true,
		Description:          &description,
	}
}

func testRequest() MarketplaceRouteRequest {
	return MarketplaceRouteRequest{
		OriginState:      "Abia",
		OriginCity:       "Aba",
		DestinationState: "Lagos",
		DestinationCity:  "Lagos",
		Package: MarketplacePackage{
			WeightKg: 2, LengthCM: 20, WidthCM: 20, HeightCM: 10, Category: "parcel",
		},
	}
}

func TestBuildProviderQuotesNormalizesServices(t *testing.T) {
	quotes, err := BuildProviderQuotes(testProvider(), testRequest(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(quotes) != 2 {
		t.Fatalf("expected two configured services, got %d", len(quotes))
	}
	if quotes[0].TotalPriceMinor != 50000 {
		t.Fatalf("expected 50000 for standard, got %d", quotes[0].TotalPriceMinor)
	}
	if quotes[0].PriceConfirmationType != "confirmed" {
		t.Fatalf("expected confirmed price, got %s", quotes[0].PriceConfirmationType)
	}
	if quotes[0].DeliveryMethod != "door-to-door" {
		t.Fatalf("expected door-to-door, got %s", quotes[0].DeliveryMethod)
	}
	if quotes[0].EstimatedDelivery != "2 days" {
		t.Fatalf("expected provider transit time, got %s", quotes[0].EstimatedDelivery)
	}
	if quotes[0].BookingSlug != "aba-freight-booking" {
		t.Fatalf("expected booking slug to be carried on the quote, got %s", quotes[0].BookingSlug)
	}
}

func TestBuildProviderQuotesRespectsItemCategories(t *testing.T) {
	provider := testProvider()
	provider.ProhibitedCategories = []string{"electronics"}
	req := testRequest()
	req.Package.Category = "electronics"
	if _, err := BuildProviderQuotes(provider, req, time.Now().Add(time.Minute)); !errors.Is(err, ErrCategoryNotAllowed) {
		t.Fatalf("expected prohibited category to be rejected, got %v", err)
	}

	req.Package.Category = "document"
	if _, err := BuildProviderQuotes(provider, req, time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("expected accepted category to be quoted, got %v", err)
	}
}

func TestBuildProviderQuotesMarksEstimates(t *testing.T) {
	provider := testProvider()
	provider.PriceIsEstimate = true
	quotes, err := BuildProviderQuotes(provider, testRequest(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quotes[0].PriceConfirmationType != "estimated" {
		t.Fatalf("expected estimated price, got %s", quotes[0].PriceConfirmationType)
	}
	if quotes[0].PriceConfirmationNote == "" {
		t.Fatal("expected an explanatory note for estimated prices")
	}
}

func TestBuildProviderQuotesSkipsUnconfiguredSpeeds(t *testing.T) {
	provider := testProvider()
	provider.ExpressPerKmMinor = nil
	req := testRequest()
	req.Speed = "express"
	if _, err := BuildProviderQuotes(provider, req, time.Now().Add(time.Minute)); !errors.Is(err, ErrNoRateConfigured) {
		t.Fatalf("expected ErrNoRateConfigured, got %v", err)
	}
}

func TestBuildProviderQuotesEnforcesRestrictions(t *testing.T) {
	provider := testProvider()
	provider.MaxWeightKg = float64Ptr(1)
	provider.MinimumChargeMinor = 0
	if _, err := BuildProviderQuotes(provider, testRequest(), time.Now().Add(time.Minute)); !errors.Is(err, ErrWeightExceeded) {
		t.Fatalf("expected weight error, got %v", err)
	}

	provider = testProvider()
	provider.PickupAvailable = false
	req := testRequest()
	req.PickupRequired = true
	if _, err := BuildProviderQuotes(provider, req, time.Now().Add(time.Minute)); !errors.Is(err, ErrNoRateConfigured) {
		t.Fatalf("expected no compatible service, got %v", err)
	}
}

func TestSortMarketplaceQuotes(t *testing.T) {
	quotes := []MarketplaceQuote{
		{ServiceKey: "standard", TotalPriceMinor: 500, EstimatedDurationDays: 4, PriceConfirmationType: "estimated"},
		{ServiceKey: "express", TotalPriceMinor: 900, EstimatedDurationDays: 1, PriceConfirmationType: "confirmed"},
		{ServiceKey: "same_day", TotalPriceMinor: 1200, EstimatedDurationDays: 0, PriceConfirmationType: "estimated"},
	}

	SortMarketplaceQuotes(quotes, "cheapest")
	if quotes[0].ServiceKey != "standard" {
		t.Fatalf("expected cheapest first, got %s", quotes[0].ServiceKey)
	}

	SortMarketplaceQuotes(quotes, "fastest")
	if quotes[0].ServiceKey != "same_day" {
		t.Fatalf("expected fastest first, got %s", quotes[0].ServiceKey)
	}

	SortMarketplaceQuotes(quotes, "recommended")
	if quotes[0].ServiceKey != "express" {
		t.Fatalf("expected confirmed price first for recommended, got %s", quotes[0].ServiceKey)
	}
}

func TestQuoteFilters(t *testing.T) {
	quote := MarketplaceQuote{
		ProviderSlug:          "aba-freight",
		ServiceKey:            "standard",
		TotalPriceMinor:       50000,
		PriceConfirmationType: "estimated",
		PickupAvailable:       true,
		DeliveryMethod:        "door-to-door",
	}
	if !QuoteMatchesFilters(quote, []string{"estimated"}, nil, nil, "aba-freight", "standard", true, true) {
		t.Fatal("expected quote to match filters")
	}
	if QuoteMatchesFilters(quote, []string{"confirmed"}, nil, nil, "", "", false, false) {
		t.Fatal("expected confirmed-only filter to exclude the quote")
	}
	max := int64(100)
	if QuoteMatchesFilters(quote, nil, nil, &max, "", "", false, false) {
		t.Fatal("expected max price filter to exclude the quote")
	}
}

func TestNewMarketplaceQuoteReferenceIsUnique(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		ref := NewMarketplaceQuoteReference()
		if seen[ref] {
			t.Fatalf("duplicate quote reference %s", ref)
		}
		seen[ref] = true
	}
}
