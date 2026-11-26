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
		PickupAvailable:    true,
		DropoffAvailable:   true,
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
		PickupAvailable:      true,
		DropoffAvailable:     true,
	}
	_, err := BuildZoneRouteQuote(provider, MarketplaceRouteRequest{
		Package: MarketplacePackage{WeightKg: 1, Category: "jewelry"},
	}, time.Now().Add(time.Minute))
	if !errors.Is(err, ErrCategoryNotAllowed) {
		t.Fatalf("expected ErrCategoryNotAllowed, got %v", err)
	}
}

func TestBuildZoneRouteQuoteRequiresOfferedServiceMode(t *testing.T) {
	provider := ZoneRouteProvider{
		CompanyID: "company-1", CompanyName: "Northline", RouteID: "route-1",
		Pricing:          ZoneRoutePricing{BasePriceMinor: 100},
		PickupAvailable:  true,
		DropoffAvailable: false,
	}
	req := MarketplaceRouteRequest{
		ServiceMode: "dropoff",
		Package:     MarketplacePackage{WeightKg: 1, Category: "other"},
	}
	if _, err := BuildZoneRouteQuote(provider, req, time.Now().Add(time.Minute)); !errors.Is(err, ErrRouteUnavailable) {
		t.Fatalf("expected ErrRouteUnavailable for an unsupported drop-off route, got %v", err)
	}

	req.ServiceMode = "pickup"
	quote, err := BuildZoneRouteQuote(provider, req, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.ServiceMode != "pickup" || !quote.PickupAvailable {
		t.Fatalf("expected a pickup quote, got mode %q", quote.ServiceMode)
	}
}

func TestBuildZoneRouteQuoteUsesConfiguredWindows(t *testing.T) {
	pickupHours := 2.0
	deliveryHours := 30.0
	provider := ZoneRouteProvider{
		CompanyID: "company-1", CompanyName: "Northline", RouteID: "route-1",
		Pricing:                ZoneRoutePricing{BasePriceMinor: 100},
		PickupAvailable:        true,
		DropoffAvailable:       true,
		EstimatedPickupHours:   &pickupHours,
		EstimatedDeliveryHours: &deliveryHours,
	}
	quote, err := BuildZoneRouteQuote(provider, MarketplaceRouteRequest{
		ServiceMode: "pickup",
		Package:     MarketplacePackage{WeightKg: 1, Category: "other"},
	}, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.EstimatedPickup != "Pickup within 2h" {
		t.Fatalf("unexpected pickup window %q", quote.EstimatedPickup)
	}
	if quote.EstimatedDelivery != "Delivery in 2 days" {
		t.Fatalf("unexpected delivery window %q", quote.EstimatedDelivery)
	}
	if quote.EstimatedDurationDays != 2 {
		t.Fatalf("expected two delivery days, got %d", quote.EstimatedDurationDays)
	}
}

func TestSortMarketplaceQuotesByPickupAndDelivery(t *testing.T) {
	fastPickup := 1.0
	slowPickup := 8.0
	fastDelivery := 12.0
	slowDelivery := 48.0
	quotes := []MarketplaceQuote{
		{ProviderID: "a", TotalPriceMinor: 100, EstimatedPickupHours: &slowPickup, EstimatedDeliveryHours: &slowDelivery},
		{ProviderID: "b", TotalPriceMinor: 300, EstimatedPickupHours: &fastPickup, EstimatedDeliveryHours: &slowDelivery},
		{ProviderID: "c", TotalPriceMinor: 200, EstimatedPickupHours: &fastPickup, EstimatedDeliveryHours: &fastDelivery},
		{ProviderID: "d", TotalPriceMinor: 50, EstimatedDurationDays: 4},
	}

	SortMarketplaceQuotes(quotes, "pickup")
	if quotes[0].ProviderID != "c" || quotes[1].ProviderID != "b" {
		t.Fatalf("unexpected pickup order: %s, %s", quotes[0].ProviderID, quotes[1].ProviderID)
	}
	if quotes[3].ProviderID != "d" {
		t.Fatalf("expected the route without a pickup window to sort last, got %s", quotes[3].ProviderID)
	}

	SortMarketplaceQuotes(quotes, "delivery")
	if quotes[0].ProviderID != "c" {
		t.Fatalf("unexpected delivery order: %s", quotes[0].ProviderID)
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
