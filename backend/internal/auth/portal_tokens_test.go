package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDriverPortalTokenRoundTrip(t *testing.T) {
	manager := NewManager(testSecret, 15*time.Minute)
	driverID := uuid.New()
	token, err := manager.DriverPortalToken(driverID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parsed, err := manager.ParseDriverPortalToken(token)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if parsed != driverID {
		t.Fatalf("expected %s, got %s", driverID, parsed)
	}
}

func TestDriverPortalTokenRejectsTampering(t *testing.T) {
	manager := NewManager(testSecret, 15*time.Minute)
	token, err := manager.DriverPortalToken(uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := manager.ParseDriverPortalToken(token + "x"); err == nil {
		t.Fatal("expected a modified token to be rejected")
	}
	other := NewManager("another-secret-that-is-at-least-32-characters", 15*time.Minute)
	if _, err := other.ParseDriverPortalToken(token); err == nil {
		t.Fatal("expected a token signed with another secret to be rejected")
	}
}
