package redis

import (
	"be-lms/config"
	"be-lms/database/db"
	"context"
	"fmt"
	"time"

	goRedis "github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	KeyPrefix  string
	MaxAttempt int
	Duration   time.Duration
}

func NewRateLimiter(maxAttempt int, duration time.Duration, prefix string) *RateLimiter {
	if prefix == "" {
		prefix = "rate_limit"
	}
	return &RateLimiter{
		KeyPrefix:  prefix,
		MaxAttempt: maxAttempt,
		Duration:   duration,
	}
}

func (r *RateLimiter) IsLimited(ip string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", r.KeyPrefix, ip)

	attempts, err := db.RedisClient.Get(ctx, key).Int()
	if err != nil && err != goRedis.Nil {
		// Nếu Redis lỗi, không chặn
		config.Log.Error(fmt.Errorf("redis GET error: %w", err))
		return false, nil
	}

	if attempts >= r.MaxAttempt {
		return true, nil
	}

	pipe := db.RedisClient.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, r.Duration)

	_, err = pipe.Exec(ctx)
	if err != nil {
		// Nếu Redis pipeline lỗi, không chặn
		config.Log.Error(fmt.Errorf("redis pipeline error: %w", err))
		return false, nil
	}

	// Nếu vượt quá giới hạn sau khi tăng
	if incr.Val() > int64(r.MaxAttempt) {
		return true, nil
	}

	return false, nil
}
