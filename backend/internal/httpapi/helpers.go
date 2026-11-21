package httpapi

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

// numericToFloat64 converts a numeric database value to a float64 for JSON
// responses. Monetary values are always handled as integer minor units; this is
// only used for weights, distances and dimensions.
func numericToFloat64(v pgtype.Numeric) float64 {
	if !v.Valid {
		return 0
	}
	f, err := v.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

// float64ToNumeric converts a measurement (weight, distance, dimensions) to the
// numeric type used by the database. Never use this for money.
func float64ToNumeric(v float64) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(strconv.FormatFloat(v, 'f', -1, 64)); err != nil {
		return pgtype.Numeric{}
	}
	return n
}
