package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("auth: invalid or expired access token")

const tokenIssuer = "bol-platform"

type Claims struct {
	jwt.RegisteredClaims
	Type string `json:"typ"`
}

type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// Issue creates a single access token. The returned JTI lets logout add the
// token to the Redis denylist until it expires naturally.
func (m *Manager) Issue(userID uuid.UUID) (token string, jti uuid.UUID, expiresAt time.Time, err error) {
	now := time.Now()
	expiresAt = now.Add(m.ttl)
	jti = uuid.New()

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    tokenIssuer,
			ID:        jti.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Type: "access",
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", uuid.Nil, time.Time{}, fmt.Errorf("auth: sign token: %w", err)
	}
	return signed, jti, expiresAt, nil
}

func (m *Manager) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithIssuer(tokenIssuer), jwt.WithExpirationRequired())
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Type != "access" || claims.Subject == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (m *Manager) TTL() time.Duration {
	return m.ttl
}
