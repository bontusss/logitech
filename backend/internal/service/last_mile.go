package service

import (
	"math"
	"strings"
	"time"
)

// LastMileProviderRate is one active last-mile service area, reduced to the
// fields needed to quote a verified pickup and drop-off.
type LastMileProviderRate struct {
	CompanyID            string
	CompanyName          string
	CompanySlug          string
	BookingSlug          string
	LogoURL              string
	PublicDescription    string
	AreaID               string
	AreaName             string
	State                string
	City                 string
	PricingModel         string
	FlatPriceMinor       *int64
	BaseDeliveryFeeMinor *int64
	PricePerKmMinor      *int64
	MinimumFeeMinor      *int64
	MaximumDistanceKm    *float64
	MaximumFeeMinor      *int64
	VolumetricDivisor    float64
	PriceIsEstimate      bool
	MaxWeightKg          *float64
	MaxLengthCM          *float64
	MaxWidthCM           *float64
	MaxHeightCM          *float64
	ExpressAvailable     bool
	PickupAvailable      bool
	DeliveryAvailable    bool
	DistanceKm           float64
	AcceptedCategories   []string
	ProhibitedCategories []string
}

// BuildLastMileQuote turns a provider's coverage area and a verified Mapbox
// route distance into a transparent local delivery quote.
func BuildLastMileQuote(provider LastMileProviderRate, req MarketplaceRouteRequest, expiresAt time.Time) (MarketplaceQuote, error) {
	if !ItemCategoryAllowed(req.Package.Category, provider.AcceptedCategories, provider.ProhibitedCategories) {
		return MarketplaceQuote{}, ErrCategoryNotAllowed
	}
	if err := CheckParcelCompatibility(
		req.Package.WeightKg, req.Package.LengthCM, req.Package.WidthCM, req.Package.HeightCM,
		provider.MaxWeightKg, provider.MaxLengthCM, provider.MaxWidthCM, provider.MaxHeightCM,
	); err != nil {
		return MarketplaceQuote{}, err
	}
	if provider.DistanceKm <= 0 {
		return MarketplaceQuote{}, ErrNoRateConfigured
	}
	if provider.MaximumDistanceKm != nil && *provider.MaximumDistanceKm > 0 &&
		provider.DistanceKm > *provider.MaximumDistanceKm {
		return MarketplaceQuote{}, ErrRouteUnavailable
	}

	total, breakdown, err := calculateLastMilePrice(provider)
	if err != nil {
		return MarketplaceQuote{}, err
	}

	confirmation := "confirmed"
	note := ""
	if provider.PriceIsEstimate {
		confirmation = "estimated"
		note = "Estimated price. Final price will be confirmed by the logistics provider after package inspection."
	}

	method := "terminal-to-terminal"
	switch {
	case provider.PickupAvailable && provider.DeliveryAvailable:
		method = "door-to-door"
	case provider.DeliveryAvailable:
		method = "terminal-to-door"
	case provider.PickupAvailable:
		method = "door-to-terminal"
	}

	return MarketplaceQuote{
		QuoteReference:        NewMarketplaceQuoteReference(),
		ProviderID:            provider.CompanyID,
		ProviderName:          provider.CompanyName,
		ProviderLogo:          provider.LogoURL,
		ProviderSlug:          provider.CompanySlug,
		BookingSlug:           provider.BookingSlug,
		ServiceKey:            "last_mile",
		ServiceName:           "Last-mile delivery",
		OriginState:           req.OriginState,
		OriginCity:            req.OriginCity,
		DestinationState:      req.DestinationState,
		DestinationCity:       req.DestinationCity,
		BasePriceMinor:        total,
		ChargeableWeightKg:    ChargeableWeight(req.Package.WeightKg, req.Package.LengthCM, req.Package.WidthCM, req.Package.HeightCM, provider.VolumetricDivisor),
		TotalPriceMinor:       total,
		Currency:              "NGN",
		EstimatedDelivery:     "Local delivery",
		EstimatedDurationDays: 0,
		PickupAvailable:       provider.PickupAvailable,
		DeliveryAvailable:     provider.DeliveryAvailable,
		DeliveryMethod:        method,
		PriceConfirmationType: confirmation,
		PriceConfirmationNote: note,
		Status:                "active",
		ExpiresAt:             expiresAt,
		Conditions:            strings.TrimSpace(provider.PublicDescription),
		Breakdown:             breakdown,
		DistanceKm:            provider.DistanceKm,
	}, nil
}

func calculateLastMilePrice(provider LastMileProviderRate) (int64, []RateCharge, error) {
	var total int64
	breakdown := make([]RateCharge, 0, 3)

	switch provider.PricingModel {
	case "flat":
		if provider.FlatPriceMinor == nil || *provider.FlatPriceMinor < 0 {
			return 0, nil, ErrNoRateConfigured
		}
		total = *provider.FlatPriceMinor
		breakdown = append(breakdown, RateCharge{
			Name:        "Flat delivery fee",
			Description: "Provider flat rate for this service area",
			AmountMinor: total,
		})
	case "distance":
		if provider.BaseDeliveryFeeMinor == nil || provider.PricePerKmMinor == nil ||
			*provider.BaseDeliveryFeeMinor < 0 || *provider.PricePerKmMinor < 0 {
			return 0, nil, ErrNoRateConfigured
		}
		base := *provider.BaseDeliveryFeeMinor
		distanceCharge := int64(math.Ceil(float64(*provider.PricePerKmMinor) * provider.DistanceKm))
		total = base + distanceCharge
		breakdown = append(breakdown,
			RateCharge{Name: "Base delivery fee", AmountMinor: base},
			RateCharge{
				Name:        "Distance charge",
				Description: formatDistanceCharge(provider.DistanceKm),
				AmountMinor: distanceCharge,
			},
		)
		if provider.MinimumFeeMinor != nil && total < *provider.MinimumFeeMinor {
			adjustment := *provider.MinimumFeeMinor - total
			breakdown = append(breakdown, RateCharge{
				Name:        "Minimum fee applied",
				Description: "Provider minimum for this service area",
				AmountMinor: adjustment,
			})
			total = *provider.MinimumFeeMinor
		}
		if provider.MaximumFeeMinor != nil && *provider.MaximumFeeMinor >= 0 && total > *provider.MaximumFeeMinor {
			adjustment := *provider.MaximumFeeMinor - total
			breakdown = append(breakdown, RateCharge{
				Name:        "Maximum fee applied",
				Description: "Provider cap for this service area",
				AmountMinor: adjustment,
			})
			total = *provider.MaximumFeeMinor
		}
	default:
		return 0, nil, ErrNoRateConfigured
	}
	return total, breakdown, nil
}
