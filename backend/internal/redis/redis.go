package redis

import (
	"context"
	"log/slog"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Store handles token blacklisting using Redis with a graceful in-memory fallback.
type Store struct {
	client    *goredis.Client
	isRedis   bool
	mu        sync.RWMutex
	memTokens map[string]time.Time
}

// NewStore initializes a connection to Redis, or falls back to an in-memory cache if unavailable.
func NewStore(addr string) *Store {
	store := &Store{
		memTokens: make(map[string]time.Time),
	}

	if addr != "" {
		opt, err := goredis.ParseURL(addr)
		var client *goredis.Client
		if err != nil {
			client = goredis.NewClient(&goredis.Options{
				Addr: addr,
			})
		} else {
			client = goredis.NewClient(opt)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err == nil {
			slog.Info("connected to redis store", "addr", addr)
			store.client = client
			store.isRedis = true
			return store
		}
		slog.Warn("redis ping failed, using in-memory store fallback", "err", err)
	}

	return store
}

// BlacklistToken stores a revoked token until its expiration.
func (s *Store) BlacklistToken(ctx context.Context, token string, expiration time.Duration) error {
	if s.isRedis && s.client != nil {
		return s.client.Set(ctx, "bl:"+token, "revoked", expiration).Err()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.memTokens[token] = time.Now().Add(expiration)
	return nil
}

// IsTokenBlacklisted checks whether a token has been revoked.
func (s *Store) IsTokenBlacklisted(ctx context.Context, token string) bool {
	if s.isRedis && s.client != nil {
		val, err := s.client.Get(ctx, "bl:"+token).Result()
		return err == nil && val == "revoked"
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	exp, exists := s.memTokens[token]
	if !exists {
		return false
	}
	if time.Now().After(exp) {
		return false
	}
	return true
}

// Close closes the Redis client connection if active.
func (s *Store) Close() error {
	if s.isRedis && s.client != nil {
		return s.client.Close()
	}
	return nil
}
