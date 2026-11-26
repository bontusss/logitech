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
	CompanyID              string
	CompanyName            string
	CompanySlug            string
	BookingSlug            string
	LogoURL                string
	PublicDescription      string
	RouteID                string
	FromZoneID             string
	FromZoneName           string
	FromState              string
	FromLGA                string
	ToZoneID               string
	ToZoneName             string
	ToState                string
	ToLGA                  string
	DropoffBranchName      string
	DropoffBranchAddress   string
	Pricing                ZoneRoutePricing
	PickupAvailable        bool
	DropoffAvailable       bool
	EstimatedPickupHours   *float64
	EstimatedDeliveryHours *float64
	AcceptedCategories     []string
	ProhibitedCategories   []string
	ProviderRating         float64
}

// BuildZoneRouteQuote converts a matched directional route into a marketplace
// quote. It never falls back to a reverse route, a provider's home address, or
// a service mode the route does not offer.
func BuildZoneRouteQuote(provider ZoneRouteProvider, req MarketplaceRouteRequest, expiresAt time.Time) (MarketplaceQuote, error) {
	mode := req.ServiceMode
	if mode == "" {
		mode = "pickup"
	}
	switch mode {
	case "pickup":
		if !provider.PickupAvailable {
			return MarketplaceQuote{}, ErrRouteUnavailable
		}
	case "dropoff":
		if !provider.DropoffAvailable {
			return MarketplaceQuote{}, ErrRouteUnavailable
		}
	default:
		return MarketplaceQuote{}, ErrRouteUnavailable
	}
	if !ItemCategoryAllowed(req.Package.Category, provider.AcceptedCategories, provider.ProhibitedCategories) {
		return MarketplaceQuote{}, ErrCategoryNotAllowed
	}
	price, err := CalculateZoneRoutePrice(provider.Pricing, req.Package.WeightKg)
	if err != nil {
		return MarketplaceQuote{}, err
	}

	deliveryMethod := "door-to-door"
	if mode == "dropoff" {
		deliveryMethod = "terminal-to-door"
	}

	estimatedDelivery := "3–5 days"
	estimatedDurationDays := 4
	if provider.EstimatedDeliveryHours != nil && *provider.EstimatedDeliveryHours > 0 {
		hours := *provider.EstimatedDeliveryHours
		estimatedDelivery = "Delivery in " + formatLegDuration(hours)
		estimatedDurationDays = int(math.Ceil(hours / 24))
	}
	estimatedPickup := ""
	var pickupHours *float64
	if provider.EstimatedPickupHours != nil && *provider.EstimatedPickupHours > 0 {
		hours := *provider.EstimatedPickupHours
		pickupHours = &hours
		estimatedPickup = "Pickup within " + formatLegDuration(hours)
	}
	var deliveryHours *float64
	if provider.EstimatedDeliveryHours != nil && *provider.EstimatedDeliveryHours > 0 {
		hours := *provider.EstimatedDeliveryHours
		deliveryHours = &hours
	}

	return MarketplaceQuote{
		QuoteReference:         NewMarketplaceQuoteReference(),
		ProviderID:             provider.CompanyID,
		ProviderName:           provider.CompanyName,
		ProviderLogo:           provider.LogoURL,
		ProviderSlug:           provider.CompanySlug,
		ProviderRating:         provider.ProviderRating,
		BookingSlug:            provider.BookingSlug,
		ServiceMode:            mode,
		ServiceKey:             "zone_route",
		ServiceName:            "Zone route",
		OriginState:            provider.FromState,
		OriginCity:             provider.FromLGA,
		OriginZoneName:         provider.FromZoneName,
		DestinationState:       provider.ToState,
		DestinationCity:        provider.ToLGA,
		DestinationZoneName:    provider.ToZoneName,
		DropoffBranchName:      provider.DropoffBranchName,
		DropoffBranchAddress:   provider.DropoffBranchAddress,
		BasePriceMinor:         price.BasePriceMinor,
		ChargeableWeightKg:     price.ChargeableWeightKg,
		AdditionalChargesMinor: price.AdditionalChargesMinor,
		TotalPriceMinor:        price.TotalMinor,
		Currency:               "NGN",
		EstimatedDelivery:      estimatedDelivery,
		EstimatedDurationDays:  estimatedDurationDays,
		EstimatedPickup:        estimatedPickup,
		EstimatedPickupHours:   pickupHours,
		EstimatedDeliveryHours: deliveryHours,
		PickupAvailable:        provider.PickupAvailable,
		DropoffAvailable:       provider.DropoffAvailable,
		DeliveryAvailable:      true,
		DeliveryMethod:         deliveryMethod,
		PriceConfirmationType:  "confirmed",
		Status:                 "active",
		ExpiresAt:              expiresAt,
		Breakdown:              price.Breakdown,
		DistanceKm:             0,
		Conditions:             strings.TrimSpace(provider.PublicDescription),
	}, nil
}

// formatLegDuration renders a configured hourly window for display.
func formatLegDuration(hours float64) string {
	switch {
	case hours <= 0:
		return ""
	case hours < 1:
		return "under 1h"
	case hours < 24:
		rounded := math.Round(hours)
		if rounded < 1 {
			rounded = 1
		}
		return fmt.Sprintf("%.0fh", rounded)
	default:
		days := math.Ceil(hours/24 - 1e-9)
		if days <= 1 {
			return "1 day"
		}
		return fmt.Sprintf("%.0f days", days)
	}
}
