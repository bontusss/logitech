package service

import (
	"errors"
	"testing"
	"time"
)

func testLastMileProvider() LastMileProviderRate {
	return LastMileProviderRate{
		CompanyID:            "33333333-3333-3333-3333-333333333333",
		CompanyName:          "Island Riders",
		CompanySlug:          "island-riders",
		BookingSlug:          "island-riders",
		AreaID:               "44444444-4444-4444-4444-444444444444",
		AreaName:             "Lagos Island",
		City:                 "Lagos",
		State:                "Lagos",
		PricingModel:         "distance",
		BaseDeliveryFeeMinor: int64Ptr(50000),
		PricePerKmMinor:      int64Ptr(10000),
		DistanceKm:           5,
		PickupAvailable:      true,
		DeliveryAvailable:    true,
		AcceptedCategories:   []string{"document", "electronics", "other"},
	}
}

func testLastMileRequest() MarketplaceRouteRequest {
	return MarketplaceRouteRequest{
		ServiceType:      "last_mile",
		OriginState:      "Lagos",
		OriginCity:       "Lagos",
		DestinationState: "Lagos",
		DestinationCity:  "Lagos",
		DistanceKm:       5,
		Package: MarketplacePackage{
			WeightKg: 2, Category: "document",
		},
	}
}

func TestBuildLastMileQuoteUsesDistance(t *testing.T) {
	quote, err := BuildLastMileQuote(testLastMileProvider(), testLastMileRequest(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.ServiceKey != "last_mile" {
		t.Fatalf("expected last_mile service, got %s", quote.ServiceKey)
	}
	if quote.TotalPriceMinor != 100000 {
		t.Fatalf("expected 50000 base + 50000 distance charge, got %d", quote.TotalPriceMinor)
	}
	if quote.DistanceKm != 5 {
		t.Fatalf("expected the verified distance on the quote, got %v", quote.DistanceKm)
	}
	if quote.BookingSlug != "island-riders" {
		t.Fatalf("expected booking slug to be carried, got %s", quote.BookingSlug)
	}
}

func TestBuildLastMileQuoteAppliesMinimumAndMaximum(t *testing.T) {
	provider := testLastMileProvider()
	provider.PricePerKmMinor = int64Ptr(1000)
	minimum := int64(75000)
	provider.MinimumFeeMinor = &minimum
	quote, err := BuildLastMileQuote(provider, testLastMileRequest(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.TotalPriceMinor != 75000 {
		t.Fatalf("expected minimum fee, got %d", quote.TotalPriceMinor)
	}

	provider = testLastMileProvider()
	provider.PricePerKmMinor = int64Ptr(100000)
	maximum := int64(80000)
	provider.MaximumFeeMinor = &maximum
	quote, err = BuildLastMileQuote(provider, testLastMileRequest(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.TotalPriceMinor != 80000 {
		t.Fatalf("expected maximum fee, got %d", quote.TotalPriceMinor)
	}
}

func TestBuildLastMileQuoteRejectsCoverageAndCategory(t *testing.T) {
	provider := testLastMileProvider()
	provider.MaximumDistanceKm = float64Ptr(3)
	if _, err := BuildLastMileQuote(provider, testLastMileRequest(), time.Now().Add(time.Minute)); !errors.Is(err, ErrRouteUnavailable) {
		t.Fatalf("expected out-of-coverage error, got %v", err)
	}

	provider = testLastMileProvider()
	provider.ProhibitedCategories = []string{"document"}
	if _, err := BuildLastMileQuote(provider, testLastMileRequest(), time.Now().Add(time.Minute)); !errors.Is(err, ErrCategoryNotAllowed) {
		t.Fatalf("expected prohibited category error, got %v", err)
	}
}
