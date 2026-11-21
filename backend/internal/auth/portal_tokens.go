package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidPortalToken = errors.New("auth: invalid driver portal token")

// DriverPortalToken creates a stable, signed link token for a driver. The
// signing key is the application JWT secret, so no plaintext portal token is
// stored in the database. A driver whose status is not active cannot use it.
func (m *Manager) DriverPortalToken(driverID uuid.UUID) (string, error) {
	if driverID == uuid.Nil {
		return "", ErrInvalidPortalToken
	}
	payload := "v1." + driverID.String()
	return payload + "." + m.portalSignature(payload), nil
}

// ParseDriverPortalToken validates a signed token and returns the driver id.
func (m *Manager) ParseDriverPortalToken(raw string) (uuid.UUID, error) {
	parts := strings.Split(strings.TrimSpace(raw), ".")
	if len(parts) != 3 || parts[0] != "v1" {
		return uuid.Nil, ErrInvalidPortalToken
	}
	payload := parts[0] + "." + parts[1]
	expected := m.portalSignature(payload)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[2])) != 1 {
		return uuid.Nil, ErrInvalidPortalToken
	}
	driverID, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.Nil, ErrInvalidPortalToken
	}
	return driverID, nil
}

func (m *Manager) portalSignature(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte("driver-portal:" + payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
