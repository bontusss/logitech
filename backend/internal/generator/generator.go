package generator

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"backend/internal/db/sqlc"
)

const (
	// oneTimePasswordLength is the length of the generated one-time password
	// issued at user creation and on PIN reset.
	oneTimePasswordLength = 8
	otpCharset            = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"
)

// MaxPinLength is the maximum allowed length of a user-chosen PIN.
const MaxPinLength = 10

// MinPinLength is the minimum allowed length of a user-chosen PIN.
const MinPinLength = 4

// GenerateStaffID generates a unique sequential staff ID using the configured
// prefix, formatted as <prefix>-001, <prefix>-002, etc. (e.g. GAL-001).
func GenerateStaffID(ctx context.Context, q *sqlc.Queries, prefix string) (string, error) {
	for attempt := 0; attempt < 50; attempt++ {
		seqVal, err := q.GetNextStaffIDSequence(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get next staff id sequence: %w", err)
		}

		candidate := fmt.Sprintf("%s-%03d", prefix, seqVal)
		exists, err := q.CheckStaffIDExists(ctx, candidate)
		if err != nil {
			return "", fmt.Errorf("failed to check staff id existence: %w", err)
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not generate a unique staff id after multiple attempts")
}

// GenerateOneTimePassword generates a cryptographically random alphanumeric
// one-time password. The caller stores its bcrypt hash; the clear text is
// shown exactly once to the creator so it can be handed to the user.
func GenerateOneTimePassword() (string, error) {
	charsetLen := big.NewInt(int64(len(otpCharset)))

	b := make([]byte, oneTimePasswordLength)
	for i := 0; i < oneTimePasswordLength; i++ {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("failed to read random bytes: %w", err)
		}
		b[i] = otpCharset[n.Int64()]
	}
	return string(b), nil
}

// ValidatePin checks a user-chosen PIN against the length rules.
func ValidatePin(pin string) bool {
	n := len(pin)
	return n >= MinPinLength && n <= MaxPinLength
}
