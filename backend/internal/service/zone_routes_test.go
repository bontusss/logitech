package service

import (
	"errors"
	"testing"
	"time"
)

func TestCalculateZoneRoutePriceUsesIncludedWeight(t *testing.T) {
	pricing := ZoneRoutePricing{BasePriceMinor: 500_000, IncludedWeightKg: 5, AdditionalWeightFeeMinor: 10_000}

	result, err := CalculateZoneRoutePrice(pricing, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalMinor != 500_000 {
		t.Fatalf("expected base price only, got %d", result.TotalMinor)
	}
	if result.AdditionalChargesMinor != 0 {
		t.Fatalf("expected no additional charges, got %d", result.AdditionalChargesMinor)
	}
	if len(result.Breakdown) != 1 {
		t.Fatalf("expected one breakdown line, got %d", len(result.Breakdown))
	}
}

func TestCalculateZoneRoutePriceChargesAdditionalKilograms(t *testing.T) {
	pricing := ZoneRoutePricing{BasePriceMinor: 500_000, IncludedWeightKg: 5, AdditionalWeightFeeMinor: 10_000}

	result, err := CalculateZoneRoutePrice(pricing, 6.2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AdditionalWeightKg != 1.2 {
		t.Fatalf("expected 1.2 additional kg, got %v", result.AdditionalWeightKg)
	}
	if result.AdditionalChargesMinor != 12_000 {
		t.Fatalf("expected 12,000 minor additional, got %d", result.AdditionalChargesMinor)
	}
	if result.TotalMinor != 512_000 {
		t.Fatalf("expected 512,000 total, got %d", result.TotalMinor)
	}
}

func TestCalculateZoneRoutePriceRejectsInvalidWeight(t *testing.T) {
	for _, weight := range []float64{0, -1, -0.01} {
		_, err := CalculateZoneRoutePrice(ZoneRoutePricing{BasePriceMinor: 100}, weight)
		if !errors.Is(err, ErrInvalidWeight) {
			t.Fatalf("weight %v: expected ErrInvalidWeight, got %v", weight, err)
		}
	}
}

func TestBuildZoneRouteQuoteIsDirectionalAndWeightAware(t *testing.T) {
	provider := ZoneRouteProvider{
		CompanyID:          "company-1",
		CompanyName:        "Northline",
		CompanySlug:        "northline",
		RouteID:            "route-1",
		FromZoneID:         "from-1",
		FromZoneName:       "Aba North",
		FromState:          "Abia",
		FromLGA:            "Aba North",
		ToZoneID:           "to-1",
		ToZoneName:         "Ikeja",
		ToState:            "Lagos",
		ToLGA:              "Ikeja",
		ProviderRating:     4.6,
		Pricing:            ZoneRoutePricing{BasePriceMinor: 700_000, IncludedWeightKg: 5, AdditionalWeightFeeMinor: 20_000},
		AcceptedCategories: []string{"other"},
	}
	req := MarketplaceRouteRequest{Package: MarketplacePackage{WeightKg: 8, Category: "other"}}

	quote, err := BuildZoneRouteQuote(provider, req, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.ServiceKey != "zone_route" {
		t.Fatalf("expected zone_route service, got %q", quote.ServiceKey)
	}
	if quote.OriginState != "Abia" || quote.DestinationCity != "Ikeja" {
		t.Fatalf("unexpected direction: %s/%s -> %s/%s", quote.OriginState, quote.OriginCity, quote.DestinationState, quote.DestinationCity)
	}
	if quote.TotalPriceMinor != 760_000 {
		t.Fatalf("expected 760,000 total, got %d", quote.TotalPriceMinor)
	}
	if quote.ProviderRating != 4.6 {
		t.Fatalf("expected provider rating to carry through, got %v", quote.ProviderRating)
	}
}

func TestBuildZoneRouteQuoteRejectsProhibitedCategory(t *testing.T) {
	provider := ZoneRouteProvider{
		CompanyID: "company-1", CompanyName: "Northline", RouteID: "route-1",
		ProhibitedCategories: []string{"jewelry"},
		Pricing:              ZoneRoutePricing{BasePriceMinor: 100},
	}
	_, err := BuildZoneRouteQuote(provider, MarketplaceRouteRequest{
		Package: MarketplacePackage{WeightKg: 1, Category: "jewelry"},
	}, time.Now().Add(time.Minute))
	if !errors.Is(err, ErrCategoryNotAllowed) {
		t.Fatalf("expected ErrCategoryNotAllowed, got %v", err)
	}
}

func TestSortMarketplaceQuotesByRating(t *testing.T) {
	quotes := []MarketplaceQuote{
		{ProviderID: "a", ProviderRating: 4.1, TotalPriceMinor: 100},
		{ProviderID: "b", ProviderRating: 4.9, TotalPriceMinor: 300},
		{ProviderID: "c", ProviderRating: 4.9, TotalPriceMinor: 200},
	}
	SortMarketplaceQuotes(quotes, "rating")
	if quotes[0].ProviderID != "c" || quotes[1].ProviderID != "b" || quotes[2].ProviderID != "a" {
		t.Fatalf("unexpected rating order: %s, %s, %s", quotes[0].ProviderID, quotes[1].ProviderID, quotes[2].ProviderID)
	}
}
