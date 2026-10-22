package generator

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"

	"backend/internal/db/sqlc"
)

const (
	pinCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	pinLength  = 4
)

// GenerateStaffID generates a unique sequential staff ID formatted as STF-001, STF-002, etc.
func GenerateStaffID(ctx context.Context, q *sqlc.Queries) (string, error) {
	for attempt := 0; attempt < 50; attempt++ {
		seqVal, err := q.GetNextStaffIDSequence(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get next staff id sequence: %w", err)
		}

		candidate := fmt.Sprintf("STF-%03d", seqVal)
		exists, err := q.CheckStaffIDExists(ctx, sql.NullString{String: candidate, Valid: true})
		if err != nil {
			return "", fmt.Errorf("failed to check staff id existence: %w", err)
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not generate a unique staff id after multiple attempts")
}

// GenerateAlphanumericPIN generates a cryptographically random 4-character alphanumeric PIN
// and verifies that it is unique in the staff database table.
func GenerateAlphanumericPIN(ctx context.Context, q *sqlc.Queries) (string, error) {
	charsetLen := big.NewInt(int64(len(pinCharset)))

	for attempt := 0; attempt < 30; attempt++ {
		b := make([]byte, pinLength)
		for i := 0; i < pinLength; i++ {
			n, err := rand.Int(rand.Reader, charsetLen)
			if err != nil {
				return "", fmt.Errorf("failed to read random bytes: %w", err)
			}
			b[i] = pinCharset[n.Int64()]
		}
		candidate := string(b)

		exists, err := q.CheckPinExists(ctx, candidate)
		if err != nil {
			return "", fmt.Errorf("failed to check pin uniqueness: %w", err)
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not generate a unique 4-digit alphanumeric pin after multiple attempts")
}
