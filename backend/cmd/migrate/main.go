// Command migrate runs SQL schema migrations with golang-migrate.
//
// Usage (from backend/):
//
//	go run ./cmd/migrate up        # apply all pending migrations
//	go run ./cmd/migrate up 2      # apply up to 2 pending migrations
//	go run ./cmd/migrate down      # revert 1 migration
//	go run ./cmd/migrate down 2    # revert 2 migrations
//	go run ./cmd/migrate version   # print current schema version
//
// DSN comes from DATABASE_URL or db.DefaultDSN.
package main

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"backend/internal/db"
)

const migrationsDir = "file://migrations"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	dir, n := os.Args[1], 0
	if len(os.Args) > 2 {
		if _, err := fmt.Sscanf(os.Args[2], "%d", &n); err != nil {
			fmt.Fprintf(os.Stderr, "invalid count %q\n", os.Args[2])
			os.Exit(2)
		}
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = db.DefaultDSN
	}

	m, err := migrate.New(migrationsDir, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate init: %v\n", err)
		os.Exit(1)
	}
	defer func() { srcErr, dbErr := m.Close(); fmt.Println(closeMsg(srcErr, dbErr)) }()

	switch dir {
	case "up":
		if n <= 0 {
			err = m.Up()
		} else {
			err = m.Steps(n)
		}
	case "down":
		if n <= 0 {
			n = 1
		}
		err = m.Steps(-n)
	case "version":
		v, dirty, verr := m.Version()
		if verr != nil {
			fmt.Println("no migrations applied yet")
			return
		}
		fmt.Printf("version %d dirty=%v\n", v, dirty)
		return
	default:
		usage()
		os.Exit(2)
	}

	if err != nil && err != migrate.ErrNoChange {
		fmt.Fprintf(os.Stderr, "migrate %s: %v\n", dir, err)
		os.Exit(1)
	}
	v, dirty, _ := m.Version()
	fmt.Printf("migrate %s done: version %d dirty=%v\n", dir, v, dirty)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: go run ./cmd/migrate up|down [n] | version")
}

func closeMsg(src, dbErr error) string {
	if src != nil || dbErr != nil {
		return fmt.Sprintf("close: src=%v db=%v", src, dbErr)
	}
	return ""
}
