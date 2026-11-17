package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PricingError represents a pricing calculation error
type PricingError struct {
	Code    string
	Message string
}

func (e *PricingError) Error() string { return e.Message }

var (
	// ErrUnsupportedRoute indicates the origin and destination are the same or no route exists
	ErrUnsupportedRoute = errors.New("service: unsupported route - origin and destination are identical or no matching route found")

	// ErrInactiveService indicates the selected shipping service is inactive
	ErrInactiveService = errors.New("service: the selected shipping service is inactive")

	// ErrExpiredRateCard indicates the rate card version is expired
	ErrExpiredRateCard = errors.New("service: the rate card version is expired")

	// ErrMissingPricingRule indicates no pricing rule matches the shipment
	ErrMissingPricingRule = errors.New("service: no pricing rule matches the shipment configuration")

	// ErrWeightExceedsLimit indicates the parcel weight exceeds the supported limit
	ErrWeightExceedsLimit = errors.New("service: parcel weight exceeds the supported limit")

	// ErrInvalidDimensions indicates invalid package dimensions
	ErrInvalidDimensions = errors.New("service: invalid package dimensions")

	// ErrCategoryNotFound indicates the package category is not configured
	ErrCategoryNotFound = errors.New("service: package category not found in pricing rule")
)

// ShipmentSpec contains the shipment details needed for quote calculation
type ShipmentSpec struct {
	OriginLocationType      string
	OriginLocationID        uuid.UUID
	DestinationLocationType string
	DestinationLocationID   uuid.UUID
	ServiceID               uuid.UUID
	WeightKg                float64
	LengthCM                float64
	WidthCM                 float64
	HeightCM                float64
	Category                string
}

// QuoteResult contains the result of a pricing calculation
type QuoteResult struct {
	CompanyID              uuid.UUID
	CompanyName            string
	Service                *ShippingService
	RateCardVersion        *RateCardVersion
	PricingRule            *PricingRule
	ChargeableWeightKg     float64
	BasePriceCents         int64
	TotalPriceCents        int64
	Currency               string
	EstimatedDurationDays  int
	Breakdown              []ChargeBreakdown
	QuoteReference         string
	EstimatedDelivery      string
	CanBookImmediately     bool
}

// ChargeBreakdown represents a single charge in the quote breakdown
type ChargeBreakdown struct {
	Name       string
	ChargeType string // "fixed" or "percentage"
	AmountCents int64
	IsMandatory bool
	AppliedOn  string // "base", "weight", "dimensions", etc.
}

// ShippingService represents a company's shipping service configuration
type ShippingService struct {
	ID              uuid.UUID
	Name            string
	ServiceCode     string
	DeliveryType    string
	EstimatedDuration int
	MinWeightKg     float64
	MaxWeightKg     *float64
	MaxDimensionCM  float64
	VolumetricDivisor float64
	Status          string
	CompanyName     string
}

// RateCardVersion represents a historical snapshot of a rate card
type RateCardVersion struct {
	ID             uuid.UUID
	Version        int
	IsActive       bool
	EffectiveFrom  time.Time
	EffectiveUntil *time.Time
	PricingRules   []PricingRule
	AdditionalCharges []AdditionalCharge
}

// PricingRule represents a single pricing rule from a rate card version
type PricingRule struct {
	ID                uuid.UUID
	RuleType          string // "fixed", "weight_based", "per_kg", "volumetric", "category"
	Name              string
	FixedAmountCents  *int64
	WeightTier        *[]PricingTier
	PerKgBaseCents    *int64
	PerKgRateCents    *int64
	PerKgMinimumCents *int64
	PerKgMaximumCents *int64
	VolumetricDivisor float64
	CategoryID        *uuid.UUID
	ApplicableRoutes  []string
	ApplicableServices []string
	ApplicableCategories []string
	IsPercentage      bool
	PercentageValue   float64
}

// PricingTier represents a weight bracket for weight-based pricing
type PricingTier struct {
	MinWeightKg  float64
	MaxWeightKg  float64
	PriceCents   int64
}

// AdditionalCharge represents an additional charge (fixed or percentage)
type AdditionalCharge struct {
	ID              uuid.UUID
	Name            string
	ChargeType      string // "fixed" or "percentage"
	AmountCents     int64
	PercentageValue float64
	IsMandatory     bool
	AppliesToRoutes  []string
	AppliesToServices []string
	AppliesToCategories []string
	IsIncludedInPrice bool
	AppliedOn       string
}

// CalculateQuote calculates a shipping quote based on the shipment specifications
func (s *Service) CalculateQuote(shipment ShipmentSpec) (*QuoteResult, error) {
	// 1. Validate shipment information
	if err := validateShipment(shipment); err != nil {
		return nil, err
	}

	// 2. Calculate volumetric weight
	volumetricWeightKg := calculateVolumetricWeight(shipment.LengthCM, shipment.WidthCM, shipment.HeightCM, 5000.0)

	// 3. Determine chargeable weight (max of actual and volumetric)
	chargeableWeightKg := determineChargeableWeight(shipment.WeightKg, volumetricWeightKg)

	// 4. For now, use a simplified pricing approach
	// In a full implementation, we would:
	// - Look up applicable routes
	// - Find the shipping service
	// - Get the active rate card version
	// - Match the pricing rule
	// - Calculate base price
	// - Apply additional charges

	// Simplified: return a basic quote with per-kg pricing
	basePriceCents := int64(chargeableWeightKg * 100) // 100 kobo per kg base rate
	totalPriceCents := basePriceCents

	// Generate quote reference
	quoteReference := generateQuoteReference()
	estimatedDuration := 3 // default 3 days

	return &QuoteResult{
		CompanyID:              uuid.New(),
		CompanyName:            "Company",
		Service:                nil,
		RateCardVersion:        nil,
		PricingRule:            nil,
		ChargeableWeightKg:     chargeableWeightKg,
		BasePriceCents:         basePriceCents,
		TotalPriceCents:        totalPriceCents,
		Currency:               "NGN",
		EstimatedDurationDays: estimatedDuration,
		Breakdown:              nil,
		QuoteReference:        quoteReference,
		EstimatedDelivery:     "",
		CanBookImmediately:    true,
	}, nil
}

// validateShipment checks the shipment specifications
func validateShipment(shipment ShipmentSpec) error {
	if shipment.WeightKg <= 0 {
		return errors.New("service: weight must be greater than zero")
	}
	if shipment.WeightKg > 1000 {
		return errors.New("service: weight exceeds maximum supported limit")
	}
	if shipment.LengthCM <= 0 || shipment.WidthCM <= 0 || shipment.HeightCM <= 0 {
		return ErrInvalidDimensions
	}
	if shipment.Category == "" {
		shipment.Category = "parcel" // default category
	}
	return nil
}

// calculateVolumetricWeight calculates the volumetric weight in kg
func calculateVolumetricWeight(lengthCM, widthCM, heightCM float64, volumetricDivisor float64) float64 {
	// Volume in cm³
	volume := lengthCM * widthCM * heightCM

	// Volumetric weight = volume / divisor
	volumetricWeight := volume / volumetricDivisor

	return volumetricWeight
}

// determineChargeableWeight returns the maximum of actual weight and volumetric weight
func determineChargeableWeight(actualWeightKg, volumetricWeightKg float64) float64 {
	// Chargeable weight is the maximum of actual and volumetric weight
	if volumetricWeightKg > actualWeightKg {
		return volumetricWeightKg
	}
	return actualWeightKg
}

// generateQuoteReference generates a unique quote reference
func generateQuoteReference() string {
	// In production, this would use a UUID or a unique string generation
	return "BOL-" + uuid.New().String()[:8]
}