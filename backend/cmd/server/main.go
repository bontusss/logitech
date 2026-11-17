package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/bol-platform/backend/internal/auth"
	"github.com/bol-platform/backend/internal/config"
	"github.com/bol-platform/backend/internal/database"
	"github.com/bol-platform/backend/internal/httpapi"
	"github.com/bol-platform/backend/internal/mailer"
	"github.com/bol-platform/backend/internal/ratelimit"
	"github.com/bol-platform/backend/internal/service"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Support both backend/.env and the repository-root .env used by the Makefile.
	_ = godotenv.Load("../.env")
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database unavailable", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		options, parseErr := redis.ParseURL(cfg.RedisURL)
		if parseErr != nil {
			log.Warn("redis url invalid; continuing without redis", "error", parseErr)
		} else {
			client := redis.NewClient(options)
			pingCtx, pingCancel := context.WithTimeout(ctx, 2*time.Second)
			if pingErr := client.Ping(pingCtx).Err(); pingErr != nil {
				log.Warn("redis unavailable; rate limits will use the in-process fallback", "error", pingErr)
				_ = client.Close()
			} else {
				redisClient = client
				defer client.Close()
			}
			pingCancel()
		}
	}

	mail := &mailer.LogMailer{Logger: log, AppURL: cfg.AppURL}
	authManager := auth.NewManager(cfg.JWTSecret, cfg.AccessTTL)
	svc := service.New(pool, authManager, redisClient, cfg, mail, log)
	limiter := ratelimit.New(redisClient)
	router := httpapi.New(cfg, svc, authManager, limiter, log)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("BOL API listening", "port", cfg.Port, "env", cfg.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
	log.Info("BOL API stopped")
}
