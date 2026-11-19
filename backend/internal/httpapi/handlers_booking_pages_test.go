package httpapi

import (
	"testing"

	db "github.com/bol-platform/backend/internal/db/sqlc"
)

func TestDefaultBookingPageSlug(t *testing.T) {
	tests := []struct {
		name        string
		companyName string
		companySlug string
		want        string
	}{
		{name: "uses first company name word", companyName: "2seconds logistics", companySlug: "2seconds-logistics", want: "2seconds"},
		{name: "normalizes punctuation", companyName: "A&B Freight", companySlug: "a-b-freight", want: "a-b"},
		{name: "falls back to company slug", companyName: "🚚 Logistics", companySlug: "bol-logistics", want: "bol-logistics"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := defaultBookingPageSlug(db.Company{Name: test.companyName, Slug: test.companySlug})
			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestValidateBookingPageSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "single character", input: "a", want: "a"},
		{name: "company name first word", input: "2seconds", want: "2seconds"},
		{name: "normalizes case and whitespace", input: "  Fleet-Line  ", want: "fleet-line"},
		{name: "rejects repeated hyphens", input: "fleet--line", wantErr: true},
		{name: "rejects trailing hyphen", input: "fleet-", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, reason := validateBookingPageSlug(test.input)
			if test.wantErr {
				if reason == "" {
					t.Fatalf("expected slug to be rejected, got %q", got)
				}
				return
			}
			if reason != "" {
				t.Fatalf("unexpected validation error: %s", reason)
			}
			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestValidCoordinatePair(t *testing.T) {
	tests := []struct {
		name      string
		latitude  *float64
		longitude *float64
		want      bool
	}{
		{name: "valid point", latitude: floatPointer(5.1), longitude: floatPointer(7.4), want: true},
		{name: "zero coordinates", latitude: floatPointer(0), longitude: floatPointer(0), want: true},
		{name: "missing longitude", latitude: floatPointer(5.1), want: false},
		{name: "latitude out of range", latitude: floatPointer(91), longitude: floatPointer(7.4), want: false},
		{name: "longitude out of range", latitude: floatPointer(5.1), longitude: floatPointer(181), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validCoordinatePair(test.latitude, test.longitude); got != test.want {
				t.Fatalf("got %t, want %t", got, test.want)
			}
		})
	}
}

func floatPointer(value float64) *float64 {
	return &value
}
