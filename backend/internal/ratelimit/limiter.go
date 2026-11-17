package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter enforces a fixed window per key. Redis is used when available so
// limits hold across instances; a process-local fallback keeps development
// and tests working without Redis.
type Limiter struct {
	redis *redis.Client

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	count   int
	expires time.Time
}

func New(client *redis.Client) *Limiter {
	return &Limiter{redis: client, buckets: make(map[string]*bucket)}
}

// Allow records a hit and reports whether the caller is under the limit. When
// blocked it also returns the time until the window resets.
func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration) {
	if limit <= 0 {
		return true, 0
	}

	if l.redis != nil {
		redisKey := fmt.Sprintf("ratelimit:%s", key)
		count, err := l.redis.Incr(ctx, redisKey).Result()
		if err == nil {
			if count == 1 {
				_ = l.redis.Expire(ctx, redisKey, window).Err()
			}
			if count > int64(limit) {
				ttl, _ := l.redis.TTL(ctx, redisKey).Result()
				if ttl < 0 {
					ttl = window
				}
				return false, ttl
			}
			return true, 0
		}
	}

	return l.allowLocal(key, limit, window)
}

func (l *Limiter) allowLocal(key string, limit int, window time.Duration) (bool, time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.buckets) > 10_000 {
		for k, b := range l.buckets {
			if now.After(b.expires) {
				delete(l.buckets, k)
			}
		}
	}

	b, ok := l.buckets[key]
	if !ok || now.After(b.expires) {
		l.buckets[key] = &bucket{count: 1, expires: now.Add(window)}
		return true, 0
	}

	b.count++
	if b.count > limit {
		return false, time.Until(b.expires)
	}
	return true, 0
}
