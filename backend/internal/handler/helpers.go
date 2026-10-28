package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// execer is satisfied by *sql.DB and *sql.Tx.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// formatMoney renders an amount as a decimal string with two decimals, which
// is how numeric columns are passed to PostgreSQL.
func formatMoney(value float64) string {
	return strconv.FormatFloat(math.Round(value*100)/100, 'f', 2, 64)
}

// moneyFloat parses a numeric column rendered as a string.
func moneyFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return math.Round(parsed*100) / 100
}

// parseOptionalDate parses a YYYY-MM-DD date, returning nil for an empty value.
func parseOptionalDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// nullUUIDString renders an optional UUID as an empty string when unset.
func nullUUIDString(value uuid.NullUUID) string {
	if !value.Valid {
		return ""
	}
	return value.UUID.String()
}

// logActivity appends an audit entry for an action on an entity.
func logActivity(ctx context.Context, db execer, userID uuid.UUID, entityType string, entityID uuid.UUID, action string, details map[string]any) error {
	var payload any
	if details != nil {
		encoded, err := json.Marshal(details)
		if err != nil {
			return err
		}
		payload = encoded
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO activity_logs (id, user_id, entity_type, entity_id, action, details)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), userID, entityType, entityID, action, payload,
	)
	return err
}
