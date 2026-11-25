package service

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// ZoneRoutePricing is one directional provider rate between two service zones.
// The route price covers IncludedWeightKg; every additional kilogram is billed
// at AdditionalWeightFeeMinor.
type ZoneRoutePricing struct {
	BasePriceMinor           int64
	IncludedWeightKg         float64
	AdditionalWeightFeeMinor int64
}

// ZoneRoutePriceResult is the server-calculated price for a parcel.
type ZoneRoutePriceResult struct {
	BasePriceMinor         int64        `json:"base_price_minor"`
	ChargeableWeightKg     float64      `json:"chargeable_weight_kg"`
	IncludedWeightKg       float64      `json:"included_weight_kg"`
	AdditionalWeightKg     float64      `json:"additional_weight_kg"`
	AdditionalChargesMinor int64        `json:"additional_charges_minor"`
	TotalMinor             int64        `json:"total_minor"`
	Breakdown              []RateCharge `json:"breakdown"`
}

// CalculateZoneRoutePrice applies the provider's included-weight allowance and
// per-kilogram fee. Fractions of a kilogram are rounded up so a provider is
// never underpaid for a partial kilogram.
func CalculateZoneRoutePrice(pricing ZoneRoutePricing, weightKg float64) (ZoneRoutePriceResult, error) {
	if pricing.BasePriceMinor < 0 || pricing.AdditionalWeightFeeMinor < 0 || pricing.IncludedWeightKg < 0 {
		return ZoneRoutePriceResult{}, ErrNoRateConfigured
	}
	if weightKg <= 0 || math.IsNaN(weightKg) || math.IsInf(weightKg, 0) {
		return ZoneRoutePriceResult{}, ErrInvalidWeight
	}

	additionalKg := math.Max(0, weightKg-pricing.IncludedWeightKg)
	// Round partial kilograms up, with a small tolerance for float noise
	// such as 6.2 - 5 = 1.2000000000000002.
	additionalKg = math.Ceil(additionalKg*100-1e-9) / 100
	additionalMinor := int64(math.Ceil(additionalKg*float64(pricing.AdditionalWeightFeeMinor) - 1e-9))

	breakdown := []RateCharge{{
		Name:        "Route price",
		Description: fmt.Sprintf("Covers up to %s kg", formatKg(pricing.IncludedWeightKg)),
		AmountMinor: pricing.BasePriceMinor,
	}}
	if additionalMinor > 0 {
		breakdown = append(breakdown, RateCharge{
			Name:        "Additional weight",
			Description: fmt.Sprintf("%s kg beyond the included allowance", formatKg(additionalKg)),
			AmountMinor: additionalMinor,
		})
	}

	return ZoneRoutePriceResult{
		BasePriceMinor:         pricing.BasePriceMinor,
		ChargeableWeightKg:     weightKg,
		IncludedWeightKg:       pricing.IncludedWeightKg,
		AdditionalWeightKg:     additionalKg,
		AdditionalChargesMinor: additionalMinor,
		TotalMinor:             pricing.BasePriceMinor + additionalMinor,
		Breakdown:              breakdown,
	}, nil
}

func formatKg(value float64) string {
	if math.Abs(value-math.Round(value)) < 0.005 {
		return fmt.Sprintf("%.0f", value)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

// ZoneRouteProvider is one active directional provider route matched for a
// marketplace lane.
type ZoneRouteProvider struct {
	CompanyID            string
	CompanyName          string
	CompanySlug          string
	BookingSlug          string
	LogoURL              string
	PublicDescription    string
	RouteID              string
	FromZoneID           string
	FromZoneName         string
	FromState            string
	FromLGA              string
	ToZoneID             string
	ToZoneName           string
	ToState              string
	ToLGA                string
	Pricing              ZoneRoutePricing
	AcceptedCategories   []string
	ProhibitedCategories []string
	ProviderRating       float64
}

// BuildZoneRouteQuote converts a matched directional route into a marketplace
// quote. It never falls back to a reverse route or a provider's home address.
func BuildZoneRouteQuote(provider ZoneRouteProvider, req MarketplaceRouteRequest, expiresAt time.Time) (MarketplaceQuote, error) {
	if !ItemCategoryAllowed(req.Package.Category, provider.AcceptedCategories, provider.ProhibitedCategories) {
		return MarketplaceQuote{}, ErrCategoryNotAllowed
	}
	price, err := CalculateZoneRoutePrice(provider.Pricing, req.Package.WeightKg)
	if err != nil {
		return MarketplaceQuote{}, err
	}

	return MarketplaceQuote{
		QuoteReference:         NewMarketplaceQuoteReference(),
		ProviderID:             provider.CompanyID,
		ProviderName:           provider.CompanyName,
		ProviderLogo:           provider.LogoURL,
		ProviderSlug:           provider.CompanySlug,
		ProviderRating:         provider.ProviderRating,
		BookingSlug:            provider.BookingSlug,
		ServiceKey:             "zone_route",
		ServiceName:            "Zone route",
		OriginState:            provider.FromState,
		OriginCity:             provider.FromLGA,
		DestinationState:       provider.ToState,
		DestinationCity:        provider.ToLGA,
		BasePriceMinor:         price.BasePriceMinor,
		ChargeableWeightKg:     price.ChargeableWeightKg,
		AdditionalChargesMinor: price.AdditionalChargesMinor,
		TotalPriceMinor:        price.TotalMinor,
		Currency:               "NGN",
		EstimatedDelivery:      "3–5 days",
		EstimatedDurationDays:  4,
		PickupAvailable:        true,
		DeliveryAvailable:      true,
		DeliveryMethod:         "door-to-door",
		PriceConfirmationType:  "confirmed",
		Status:                 "active",
		ExpiresAt:              expiresAt,
		Breakdown:              price.Breakdown,
		DistanceKm:             0,
		Conditions:             strings.TrimSpace(provider.PublicDescription),
	}, nil
}
