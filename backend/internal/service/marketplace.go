package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

// MarketplacePackage is the parcel information captured by the marketplace
// search. It is snapshotted onto every generated quote.
type MarketplacePackage struct {
	WeightKg float64 `json:"weight_kg"`
	LengthCM float64 `json:"length_cm"`
	WidthCM  float64 `json:"width_cm"`
	HeightCM float64 `json:"height_cm"`
	Category string  `json:"category"`
}

// MarketplaceRouteRequest is the resolved search request.
type MarketplaceRouteRequest struct {
	ServiceType      string             `json:"service_type,omitempty"`
	OriginState      string             `json:"origin_state"`
	OriginCity       string             `json:"origin_city"`
	DestinationState string             `json:"destination_state"`
	DestinationCity  string             `json:"destination_city"`
	PickupAddress    string             `json:"pickup_address,omitempty"`
	DropoffAddress   string             `json:"dropoff_address,omitempty"`
	DistanceKm       float64            `json:"distance_km,omitempty"`
	Speed            string             `json:"speed,omitempty"`
	Package          MarketplacePackage `json:"package"`
	PickupRequired   bool               `json:"pickup_required"`
	DeliveryRequired bool               `json:"delivery_required"`
}

// ProviderRate is one provider lane, reduced to the fields needed to quote it.
// It mirrors the company_booking_routes row without leaking other tenant data.
type ProviderRate struct {
	CompanyID            string
	CompanyName          string
	CompanySlug          string
	BookingSlug          string
	LogoURL              string
	PublicDescription    string
	RouteID              string
	OriginState          string
	OriginCity           string
	DestinationState     string
	DestinationCity      string
	DistanceKm           float64
	StandardPerKmMinor   *int64
	ExpressPerKmMinor    *int64
	SameDayPerKmMinor    *int64
	MinimumChargeMinor   int64
	EstimatedTransitTime *string
	DepartureSchedule    *string
	PickupAvailable      bool
	DeliveryAvailable    bool
	Description          *string
	MaxWeightKg          *float64
	MaxLengthCM          *float64
	MaxWidthCM           *float64
	MaxHeightCM          *float64
	PriceIsEstimate      bool
	AcceptedCategories   []string
	ProhibitedCategories []string
}

// MarketplaceQuote is the normalized quote returned by the marketplace and
// persisted verbatim as the quote snapshot.
type MarketplaceQuote struct {
	QuoteReference         string       `json:"quote_reference"`
	ProviderID             string       `json:"provider_id"`
	ProviderName           string       `json:"provider_name"`
	ProviderLogo           string       `json:"provider_logo"`
	ProviderSlug           string       `json:"provider_slug"`
	BookingSlug            string       `json:"booking_slug,omitempty"`
	RouteID                string       `json:"-"`
	ServiceKey             string       `json:"service_key"`
	ServiceName            string       `json:"service_name"`
	OriginState            string       `json:"origin_state"`
	OriginCity             string       `json:"origin_city"`
	DestinationState       string       `json:"destination_state"`
	DestinationCity        string       `json:"destination_city"`
	BasePriceMinor         int64        `json:"base_price_minor"`
	ChargeableWeightKg     float64      `json:"chargeable_weight_kg"`
	AdditionalChargesMinor int64        `json:"additional_charges_minor"`
	DiscountMinor          int64        `json:"discount_minor"`
	TotalPriceMinor        int64        `json:"total_price_minor"`
	Currency               string       `json:"currency"`
	EstimatedDelivery      string       `json:"estimated_delivery,omitempty"`
	EstimatedDurationDays  int          `json:"estimated_duration_days"`
	PickupAvailable        bool         `json:"pickup_available"`
	DeliveryAvailable      bool         `json:"delivery_available"`
	DeliveryMethod         string       `json:"delivery_method"`
	PriceConfirmationType  string       `json:"price_confirmation_type"`
	PriceConfirmationNote  string       `json:"price_confirmation_note,omitempty"`
	Status                 string       `json:"status"`
	ExpiresAt              time.Time    `json:"expires_at"`
	Conditions             string       `json:"conditions,omitempty"`
	Breakdown              []RateCharge `json:"breakdown"`
	DistanceKm             float64      `json:"distance_km"`
	DepartureSchedule      string       `json:"departure_schedule,omitempty"`
}

var marketplaceServices = []struct {
	Key     string
	Name    string
	Days    int
	Pick    func(ProviderRate) *int64
	Default string
}{
	{Key: "standard", Name: "Standard", Days: 4, Pick: func(p ProviderRate) *int64 { return p.StandardPerKmMinor }, Default: "3–5 days"},
	{Key: "express", Name: "Express", Days: 2, Pick: func(p ProviderRate) *int64 { return p.ExpressPerKmMinor }, Default: "1–2 days"},
	{Key: "same_day", Name: "Same-day", Days: 0, Pick: func(p ProviderRate) *int64 { return p.SameDayPerKmMinor }, Default: "Same day"},
}

// BuildProviderQuotes normalizes every configured service on a lane into a
// marketplace quote. It never invents a price: services without a rate are
// skipped, and a provider with no usable service returns ErrNoRateConfigured.
func BuildProviderQuotes(provider ProviderRate, req MarketplaceRouteRequest, expiresAt time.Time) ([]MarketplaceQuote, error) {
	if !ItemCategoryAllowed(req.Package.Category, provider.AcceptedCategories, provider.ProhibitedCategories) {
		return nil, ErrCategoryNotAllowed
	}
	if strings.EqualFold(strings.TrimSpace(req.OriginCity), strings.TrimSpace(req.DestinationCity)) &&
		strings.EqualFold(strings.TrimSpace(req.OriginState), strings.TrimSpace(req.DestinationState)) {
		return nil, ErrRouteUnavailable
	}
	if err := CheckParcelCompatibility(
		req.Package.WeightKg, req.Package.LengthCM, req.Package.WidthCM, req.Package.HeightCM,
		provider.MaxWeightKg, provider.MaxLengthCM, provider.MaxWidthCM, provider.MaxHeightCM,
	); err != nil {
		return nil, err
	}

	chargeable := ChargeableWeight(req.Package.WeightKg, req.Package.LengthCM, req.Package.WidthCM, req.Package.HeightCM, 5000)
	quotes := make([]MarketplaceQuote, 0, len(marketplaceServices))
	for _, svc := range marketplaceServices {
		if req.Speed != "" && req.Speed != svc.Key {
			continue
		}
		perKm := svc.Pick(provider)
		if perKm == nil || *perKm <= 0 {
			continue
		}
		if req.PickupRequired && !provider.PickupAvailable {
			continue
		}
		if req.DeliveryRequired && !provider.DeliveryAvailable {
			continue
		}

		rate, err := CalculateIntercityRate(IntercityRateInput{
			DistanceKm:         provider.DistanceKm,
			PerKmMinor:         *perKm,
			MinimumChargeMinor: provider.MinimumChargeMinor,
		})
		if err != nil {
			continue
		}

		confirmation := "confirmed"
		note := ""
		if provider.PriceIsEstimate {
			confirmation = "estimated"
			note = "Estimated price. Final price will be confirmed by the logistics provider after package inspection."
		}

		eta := svc.Default
		if provider.EstimatedTransitTime != nil && strings.TrimSpace(*provider.EstimatedTransitTime) != "" {
			eta = strings.TrimSpace(*provider.EstimatedTransitTime)
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

		quote := MarketplaceQuote{
			QuoteReference:        NewMarketplaceQuoteReference(),
			ProviderID:            provider.CompanyID,
			ProviderName:          provider.CompanyName,
			ProviderLogo:          provider.LogoURL,
			ProviderSlug:          provider.CompanySlug,
			BookingSlug:           provider.BookingSlug,
			ServiceKey:            svc.Key,
			ServiceName:           svc.Name,
			OriginState:           provider.OriginState,
			OriginCity:            provider.OriginCity,
			DestinationState:      provider.DestinationState,
			DestinationCity:       provider.DestinationCity,
			BasePriceMinor:        rate.BasePriceMinor,
			ChargeableWeightKg:    chargeable,
			TotalPriceMinor:       rate.TotalMinor,
			Currency:              "NGN",
			EstimatedDelivery:     eta,
			EstimatedDurationDays: svc.Days,
			PickupAvailable:       provider.PickupAvailable,
			DeliveryAvailable:     provider.DeliveryAvailable,
			DeliveryMethod:        method,
			PriceConfirmationType: confirmation,
			PriceConfirmationNote: note,
			Status:                "active",
			ExpiresAt:             expiresAt,
			Breakdown:             rate.Breakdown,
			DistanceKm:            provider.DistanceKm,
		}
		if provider.Description != nil {
			quote.Conditions = strings.TrimSpace(*provider.Description)
		}
		if provider.DepartureSchedule != nil {
			quote.DepartureSchedule = strings.TrimSpace(*provider.DepartureSchedule)
		}
		quotes = append(quotes, quote)
	}

	if len(quotes) == 0 {
		return nil, ErrNoRateConfigured
	}
	return quotes, nil
}

// SortMarketplaceQuotes applies the user-selected order. "recommended" is a
// transparent score: confirmed prices first, then faster service, then price.
func SortMarketplaceQuotes(quotes []MarketplaceQuote, order string) {
	switch order {
	case "fastest":
		sort.SliceStable(quotes, func(i, j int) bool {
			if quotes[i].EstimatedDurationDays != quotes[j].EstimatedDurationDays {
				return quotes[i].EstimatedDurationDays < quotes[j].EstimatedDurationDays
			}
			return quotes[i].TotalPriceMinor < quotes[j].TotalPriceMinor
		})
	case "recommended":
		sort.SliceStable(quotes, func(i, j int) bool {
			ci := quotes[i].PriceConfirmationType == "confirmed"
			cj := quotes[j].PriceConfirmationType == "confirmed"
			if ci != cj {
				return ci
			}
			if quotes[i].EstimatedDurationDays != quotes[j].EstimatedDurationDays {
				return quotes[i].EstimatedDurationDays < quotes[j].EstimatedDurationDays
			}
			return quotes[i].TotalPriceMinor < quotes[j].TotalPriceMinor
		})
	default:
		sort.SliceStable(quotes, func(i, j int) bool {
			return quotes[i].TotalPriceMinor < quotes[j].TotalPriceMinor
		})
	}
}

// QuoteMatchesFilters implements the UI filters over normalized quotes.
func QuoteMatchesFilters(q MarketplaceQuote, confirmation []string, minPrice, maxPrice *int64, providerSlug, serviceKey string, pickupOnly, doorToDoorOnly bool) bool {
	if len(confirmation) > 0 {
		match := false
		for _, value := range confirmation {
			if value == q.PriceConfirmationType {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	if minPrice != nil && q.TotalPriceMinor < *minPrice {
		return false
	}
	if maxPrice != nil && q.TotalPriceMinor > *maxPrice {
		return false
	}
	if providerSlug != "" && q.ProviderSlug != providerSlug {
		return false
	}
	if serviceKey != "" && q.ServiceKey != serviceKey {
		return false
	}
	if pickupOnly && !q.PickupAvailable {
		return false
	}
	if doorToDoorOnly && q.DeliveryMethod != "door-to-door" {
		return false
	}
	return true
}

// NewMarketplaceQuoteReference returns an unguessable, human-readable quote id.
func NewMarketplaceQuoteReference() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("BOL-MKT-%d", time.Now().UnixNano())
	}
	return "BOL-MKT-" + strings.ToUpper(hex.EncodeToString(buf))
}

// MarketplaceQuoteExpiry is the validity window for a marketplace quote.
const MarketplaceQuoteExpiry = 15 * time.Minute
