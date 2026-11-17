package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

const testSecret = "test-secret-that-is-at-least-32-characters-long"

func TestIssueAndParseAccessToken(t *testing.T) {
	manager := NewManager(testSecret, 15*time.Minute)
	userID := uuid.New()

	token, jti, expiresAt, err := manager.Issue(userID)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if time.Until(expiresAt) <= 0 {
		t.Fatal("token expiry is not in the future")
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Subject != userID.String() {
		t.Fatalf("subject = %q, want %q", claims.Subject, userID)
	}
	if claims.ID != jti.String() {
		t.Fatalf("jti = %q, want %q", claims.ID, jti)
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	manager := NewManager(testSecret, -time.Minute)
	token, _, _, err := manager.Issue(uuid.New())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := manager.Parse(token); err == nil {
		t.Fatal("expired token was accepted")
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	issuer := NewManager(testSecret, 15*time.Minute)
	token, _, _, err := issuer.Issue(uuid.New())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	other := NewManager("another-secret-that-is-at-least-32-characters", 15*time.Minute)
	if _, err := other.Parse(token); err == nil {
		t.Fatal("token signed with another secret was accepted")
	}
}
