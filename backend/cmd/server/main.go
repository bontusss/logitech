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

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.

// @host localhost:8080
// @BasePath /

package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "backend/docs"
	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/db/sqlc"
	"backend/internal/handler"
	"backend/internal/redis"
	"backend/internal/web"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	cfg := config.Load()

	pool, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Warn("db open failed, /health will report db=down", "err", err)
		pool = nil
	} else {
		pool.SetMaxOpenConns(10)
	}

	var queries *sqlc.Queries
	if pool != nil {
		queries = sqlc.New(pool)
	}

	redisStore := redis.NewStore(cfg.RedisURL)
	defer redisStore.Close()

	authHandler := handler.NewAuthHandler(cfg, queries, redisStore)
	staffHandler := handler.NewStaffHandler(queries)
	branchHandler := handler.NewBranchHandler(queries)

	r := gin.Default()
	r.Use(cors())

	r.GET("/health", health(pool, cfg.DatabaseURL))

	// API routes
	api := r.Group("/api")
	{
		api.GET("/hello", hello)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/admin/login", authHandler.AdminLogin)

		// Protected endpoints
		protected := api.Group("")
		protected.Use(auth.AuthRequired(cfg.JWTSecret, redisStore))
		{
			protected.GET("/auth/me", authHandler.GetMe)
			protected.POST("/auth/logout", authHandler.Logout)

			protected.POST("/staff", staffHandler.CreateStaff)
			protected.GET("/staff", auth.RequireRoles(auth.RoleSystemAdmin, auth.RoleOwner, auth.RoleManager), staffHandler.ListStaff)
			protected.GET("/staff/:id", staffHandler.GetStaff)

			protected.GET("/branches", branchHandler.ListBranches)
		}
	}

	// Swagger API documentation
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Embedded Nuxt frontend fallback
	frontend := gin.WrapH(web.Handler())
	r.GET("/", frontend)
	r.NoRoute(frontend)

	slog.Info("server listening", "port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
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
func health(pool *sql.DB, _ string) gin.HandlerFunc {
	return func(c *gin.Context) {
		out := HealthResponse{Status: "ok", DB: "down"}
		if pool != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := db.Ping(ctx, pool); err == nil {
				out.DB = "up"
			} else {
				slog.Warn("db ping failed", "err", err)
			}
		}
		c.JSON(http.StatusOK, out)
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
func hello(c *gin.Context) {
	c.JSON(http.StatusOK, HelloResponse{Message: "hello from Go"})
}

// cors allows the Nuxt dev server (http://localhost:3000) to call /api and /health.
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "http://localhost:3000" || origin == "http://localhost:8080" || origin == "http://localhost:3001" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			if c.Request.Method == http.MethodOptions {
				c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type")
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}
		c.Next()
	}
}
