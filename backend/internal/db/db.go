package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// DefaultDSN matches the docker-compose service in the repository root.
const DefaultDSN = "postgres://app:app@localhost:5432/appdb?sslmode=disable"

// Open opens a connection pool using the given DSN.
func Open(dsn string) (*sql.DB, error) {
	return sql.Open("postgres", dsn)
}

// Ping verifies the database is reachable, bounded by ctx.
func Ping(ctx context.Context, d *sql.DB) error {
	if d == nil {
		return fmt.Errorf("no database handle")
	}
	return d.PingContext(ctx)
}
