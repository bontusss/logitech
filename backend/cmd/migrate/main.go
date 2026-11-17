package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/bol-platform/backend/internal/config"
	"github.com/bol-platform/backend/internal/database"
)

func main() {
	_ = godotenv.Load("../.env")
	_ = godotenv.Load()

	direction := "up"
	steps := 0

	if len(os.Args) > 1 {
		direction = os.Args[1]
	}
	if len(os.Args) > 2 && os.Args[2] != "" {
		if os.Args[2] == "all" {
			steps = 0
		} else {
			parsed, err := strconv.Atoi(os.Args[2])
			if err != nil || parsed < 0 {
				fmt.Fprintln(os.Stderr, "usage: migrate [up|down|version] [steps|all]")
				os.Exit(1)
			}
			steps = parsed
		}
	}
	// A bare `down` reverts one migration, matching the Makefile contract.
	if direction == "down" && steps == 0 && len(os.Args) <= 2 {
		steps = 1
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	}

	migrationsDir, err := filepath.Abs("migrations")
	if err != nil {
		fmt.Fprintln(os.Stderr, "migrations path:", err)
		os.Exit(1)
	}

	if err := database.Migrate(direction, steps, "file://"+migrationsDir, cfg.DatabaseURL); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
	fmt.Printf("migrations %s complete\n", direction)
}
