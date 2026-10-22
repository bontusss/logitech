// Command server is the standalone API + embedded-frontend server.
//
// Endpoints:
//   - GET /health  -> {"status":"ok","db":"up|down"}
//   - GET /api/... -> JSON API
//   - everything else -> embedded Nuxt build (backend/internal/web/dist), if present
//
// The server can be started independently of the frontend via `make run-be`
// (or `cd backend && go run ./cmd/server`).
// @title Logitech API
// @version 1.0
// @description Logistics and Manifest Management API for Logitech.
// @termsOfService http://swagger.io/terms/

// @contact.name Logitech API Support
// @contact.url https://github.com/bontusss/logitech
// @contact.email support@logitech.example

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "backend/docs"
	"backend/internal/db"
	"backend/internal/web"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// HealthResponse represents the health check status response.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
	DB     string `json:"db" example:"up"`
}

// HelloResponse represents the sample hello response.
type HelloResponse struct {
	Message string `json:"message" example:"hello from Go"`
}

func main() {
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = db.DefaultDSN
	}

	pool, err := db.Open(dsn)
	if err != nil {
		slog.Warn("db open failed, /health will report db=down", "err", err)
		pool = nil
	} else {
		pool.SetMaxOpenConns(10)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health(pool, dsn))
	mux.HandleFunc("GET /api/hello", hello)

	// Swagger API documentation
	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/swagger/", httpSwagger.Handler())

	mux.Handle("/", web.Handler())

	slog.Info("server listening", "addr", addr)
	if err := http.ListenAndServe(":"+addr, cors(mux)); err != nil {
		slog.Error("server exited", "err", err)
		os.Exit(1)
	}
}

// health checks server and database connectivity.
//
// @Summary Check service health
// @Description Returns the service status and PostgreSQL database connection state.
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func health(pool *sql.DB, _ string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusOK
		out := HealthResponse{Status: "ok", DB: "down"}
		if pool != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := db.Ping(ctx, pool); err == nil {
				out.DB = "up"
			} else {
				slog.Warn("db ping failed", "err", err)
			}
		}
		writeJSON(w, code, out)
	}
}

// hello returns a sample greeting message.
//
// @Summary Hello message
// @Description Returns a simple greeting message from the Logitech API server.
// @Tags example
// @Produce json
// @Success 200 {object} HelloResponse
// @Router /api/hello [get]
func hello(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HelloResponse{Message: "hello from Go"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// cors allows the Nuxt dev server (http://localhost:3000) to call /api and /health.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" || origin == "http://localhost:8080" || origin == "http://localhost:3001" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
