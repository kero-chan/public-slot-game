package middleware

import (
	"github.com/google/wire"
	"github.com/slotmachine/backend/internal/config"
	infraCache "github.com/slotmachine/backend/internal/infra/cache"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// ProviderSet is the Wire provider set for middleware
var ProviderSet = wire.NewSet(
	ProvideRateLimiter,
	ProvideTrialRateLimiter,
)

// ProvideRateLimiter creates a new rate limiter instance
func ProvideRateLimiter(cfg *config.Config, log *logger.Logger) *RateLimiter {
	// Initialize Redis client for rate limiting
	redisClient, err := infraCache.NewRedisClient(cfg, log)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to Redis for rate limiting, rate limiting will be disabled")
		// Return rate limiter with nil Redis (will skip rate limiting)
		return NewRateLimiter(nil, RateLimiterConfig{
			AuthRPS:   50,
			PublicRPS: 50,
		}, log)
	}

	log.Info().Msg("Rate limiter initialized with Redis")

	return NewRateLimiter(redisClient, RateLimiterConfig{
		AuthRPS:   50, // 50 requests per second for authenticated endpoints
		PublicRPS: 50, // 50 requests per second for public endpoints
	}, log)
}

// ProvideTrialRateLimiter creates a new trial rate limiter instance
// SECURITY: This limiter prevents DoS attacks via trial session creation
func ProvideTrialRateLimiter(cfg *config.Config, redis *infraCache.RedisClient, log *logger.Logger) *TrialRateLimiter {
	log.Info().
		Int("max_sessions_per_ip", cfg.Trial.MaxSessionsPerIP).
		Int("cooldown_seconds", cfg.Trial.SessionCooldownSeconds).
		Bool("enabled", cfg.Trial.Enabled).
		Msg("Trial rate limiter initialized")

	return NewTrialRateLimiter(redis, &cfg.Trial, log)
}
