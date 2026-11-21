package service

import (
	"errors"
	"math"
	"strconv"
)

// Rate errors shared by the booking page and the marketplace.
var (
	ErrRouteUnavailable   = errors.New("service: route is no longer available")
	ErrSpeedUnavailable   = errors.New("service: delivery speed is not offered on this route")
	ErrWeightExceeded     = errors.New("service: package exceeds the provider's weight limit")
	ErrParcelTooLarge     = errors.New("service: package exceeds the provider's dimension limit")
	ErrNoRateConfigured   = errors.New("service: no valid rate is configured")
	ErrCategoryNotAllowed = errors.New("service: provider does not accept this item category")
)

// IntercityRateInput is the single intercity pricing formula used everywhere:
// a per-kilometre rate multiplied by the configured distance, floored at the
// provider's minimum charge. Money is always integer minor units.
type IntercityRateInput struct {
	DistanceKm         float64
	PerKmMinor         int64
	MinimumChargeMinor int64
}

// RateCharge is a single line item in a quote breakdown.
type RateCharge struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	AmountMinor int64  `json:"amount_minor"`
}

// IntercityRateResult is the computed route price.
type IntercityRateResult struct {
	BasePriceMinor int64        `json:"base_price_minor"`
	TotalMinor     int64        `json:"total_minor"`
	Breakdown      []RateCharge `json:"breakdown"`
}

// CalculateIntercityRate applies the provider's per-kilometre rate.
func CalculateIntercityRate(in IntercityRateInput) (IntercityRateResult, error) {
	if in.DistanceKm <= 0 || in.PerKmMinor <= 0 {
		return IntercityRateResult{}, ErrNoRateConfigured
	}
	if in.MinimumChargeMinor < 0 {
		in.MinimumChargeMinor = 0
	}

	distanceCharge := int64(math.Ceil(float64(in.PerKmMinor) * in.DistanceKm))
	breakdown := []RateCharge{{
		Name:        "Distance charge",
		Description: formatDistanceCharge(in.DistanceKm),
		AmountMinor: distanceCharge,
	}}

	total := distanceCharge
	if total < in.MinimumChargeMinor {
		adjustment := in.MinimumChargeMinor - total
		breakdown = append(breakdown, RateCharge{
			Name:        "Minimum charge applied",
			Description: "Provider minimum for this lane",
			AmountMinor: adjustment,
		})
		total = in.MinimumChargeMinor
	}

	return IntercityRateResult{
		BasePriceMinor: total,
		TotalMinor:     total,
		Breakdown:      breakdown,
	}, nil
}

// CheckParcelCompatibility enforces the provider's configured route limits.
func CheckParcelCompatibility(weightKg, lengthCM, widthCM, heightCM float64, maxWeightKg, maxLengthCM, maxWidthCM, maxHeightCM *float64) error {
	if weightKg <= 0 {
		return ErrWeightExceeded
	}
	if maxWeightKg != nil && weightKg > *maxWeightKg {
		return ErrWeightExceeded
	}
	maxDimension := math.Max(lengthCM, math.Max(widthCM, heightCM))
	for _, limit := range []*float64{maxLengthCM, maxWidthCM, maxHeightCM} {
		if limit != nil && *limit > 0 && maxDimension > *limit {
			return ErrParcelTooLarge
		}
	}
	return nil
}

// ChargeableWeight returns the greater of actual and volumetric weight.
func ChargeableWeight(actualWeightKg, lengthCM, widthCM, heightCM, volumetricDivisor float64) float64 {
	if volumetricDivisor <= 0 {
		volumetricDivisor = 5000
	}
	volumetric := (lengthCM * widthCM * heightCM) / volumetricDivisor
	if volumetric > actualWeightKg {
		return math.Round(volumetric*100) / 100
	}
	return math.Round(actualWeightKg*100) / 100
}

func formatDistanceCharge(distanceKm float64) string {
	return "Per-kilometre rate over " + trimTrailingZeros(distanceKm) + " km"
}

func trimTrailingZeros(value float64) string {
	return strconv.FormatFloat(math.Round(value*10)/10, 'f', -1, 64)
}
