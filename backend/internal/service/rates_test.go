package service

import (
	"errors"
	"testing"
)

func int64Ptr(v int64) *int64       { return &v }
func float64Ptr(v float64) *float64 { return &v }

func TestCalculateIntercityRate(t *testing.T) {
	t.Run("per kilometre price", func(t *testing.T) {
		result, err := CalculateIntercityRate(IntercityRateInput{
			DistanceKm:         400,
			PerKmMinor:         120,
			MinimumChargeMinor: 5000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.TotalMinor != 48000 {
			t.Fatalf("expected 48000, got %d", result.TotalMinor)
		}
		if len(result.Breakdown) != 1 {
			t.Fatalf("expected one breakdown line, got %d", len(result.Breakdown))
		}
	})

	t.Run("minimum charge is applied", func(t *testing.T) {
		result, err := CalculateIntercityRate(IntercityRateInput{
			DistanceKm:         10,
			PerKmMinor:         100,
			MinimumChargeMinor: 5000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.TotalMinor != 5000 {
			t.Fatalf("expected minimum 5000, got %d", result.TotalMinor)
		}
		if len(result.Breakdown) != 2 {
			t.Fatalf("expected a minimum-charge line, got %d lines", len(result.Breakdown))
		}
	})

	t.Run("missing rate", func(t *testing.T) {
		_, err := CalculateIntercityRate(IntercityRateInput{DistanceKm: 100})
		if !errors.Is(err, ErrNoRateConfigured) {
			t.Fatalf("expected ErrNoRateConfigured, got %v", err)
		}
	})
}

func TestParcelCompatibility(t *testing.T) {
	if err := CheckParcelCompatibility(3, 20, 10, 5, float64Ptr(5), nil, nil, nil); err != nil {
		t.Fatalf("expected parcel to pass, got %v", err)
	}
	if err := CheckParcelCompatibility(6, 20, 10, 5, float64Ptr(5), nil, nil, nil); !errors.Is(err, ErrWeightExceeded) {
		t.Fatalf("expected weight error, got %v", err)
	}
	if err := CheckParcelCompatibility(3, 120, 10, 5, nil, float64Ptr(100), nil, nil); !errors.Is(err, ErrParcelTooLarge) {
		t.Fatalf("expected dimension error, got %v", err)
	}
}

func TestChargeableWeightUsesVolumetric(t *testing.T) {
	if got := ChargeableWeight(2, 50, 50, 50, 5000); got != 25 {
		t.Fatalf("expected 25kg volumetric weight, got %v", got)
	}
	if got := ChargeableWeight(30, 20, 20, 20, 5000); got != 30 {
		t.Fatalf("expected actual weight, got %v", got)
	}
}
