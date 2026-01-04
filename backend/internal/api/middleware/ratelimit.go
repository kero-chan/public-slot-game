package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slotmachine/backend/internal/infra/cache"
	"github.com/slotmachine/backend/internal/pkg/errors"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	AuthRPS   int // Requests per second for authenticated endpoints
	PublicRPS int // Requests per second for public endpoints
}

// NewRateLimiter creates a new rate limiter with Redis backend
func NewRateLimiter(redis *cache.RedisClient, config RateLimiterConfig, log *logger.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		config: config,
		logger: log,
	}
}

// RateLimiter implements Redis-based rate limiting
type RateLimiter struct {
	redis  *cache.RedisClient
	config RateLimiterConfig
	logger *logger.Logger
}

// AuthenticatedMiddleware returns middleware for authenticated endpoints
// Rate limit: 10 RPS per path per user
func (rl *RateLimiter) AuthenticatedMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := rl.logger.WithTrace(c)

		// If Redis is not available, skip rate limiting
		if rl.redis == nil {
			return c.Next()
		}

		// Get user ID from context (set by auth middleware)
		userID, ok := c.Locals("user_id").(string)
		username, _ := c.Locals("username").(string)

		// Extract client IP
		clientIP := c.Get("x-real-ip")
		if clientIP == "" {
			clientIP = c.IP()
		}

		if !ok {
			// If no user ID in context, this shouldn't happen on auth routes
			// Fall back to IP to prevent bypass
			userID = "ip:" + clientIP
		}

		path := c.Path()
		limit := rl.config.AuthRPS
		window := time.Second

		// Create Redis key: ratelimit:auth:{userID}:{path}:{timestamp}
		timestamp := time.Now().Unix()
		key := fmt.Sprintf("ratelimit:auth:%s:%s:%d", userID, path, timestamp)

		allowed, remaining, resetTime := rl.checkLimit(key, limit, window)

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			// Log rate limit exceeded for monitoring and potential user blocking
			log.Warn().
				Str("user_id", userID).
				Str("username", username).
				Str("ip", clientIP).
				Str("path", path).
				Int("limit", limit).
				Str("method", c.Method()).
				Str("user_agent", c.Get("User-Agent")).
				Msg("Rate limit exceeded for authenticated user")

			return respondError(c, errors.RateLimitExceeded(int(window.Seconds())))
		}

		return c.Next()
	}
}

// PublicMiddleware returns middleware for public endpoints
// Rate limit: 50 RPS per path per IP
func (rl *RateLimiter) PublicMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := rl.logger.WithTrace(c)

		// If Redis is not available, skip rate limiting
		if rl.redis == nil {
			return c.Next()
		}

		// Extract client IP
		clientIP := c.Get("x-real-ip")
		if clientIP == "" {
			clientIP = c.IP()
		}

		path := c.Path()
		limit := rl.config.PublicRPS
		window := time.Second

		// Create Redis key: ratelimit:public:{ip}:{path}:{timestamp}
		timestamp := time.Now().Unix()
		key := fmt.Sprintf("ratelimit:public:%s:%s:%d", clientIP, path, timestamp)

		allowed, remaining, resetTime := rl.checkLimit(key, limit, window)

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			// Log rate limit exceeded for monitoring and potential IP blocking
			log.Warn().
				Str("ip", clientIP).
				Str("path", path).
				Int("limit", limit).
				Str("method", c.Method()).
				Str("user_agent", c.Get("User-Agent")).
				Msg("Rate limit exceeded for public endpoint")

			return respondError(c, errors.RateLimitExceeded(int(window.Seconds())))
		}

		return c.Next()
	}
}

// checkLimit checks if the request is within rate limit
func (rl *RateLimiter) checkLimit(key string, limit int, window time.Duration) (allowed bool, remaining int, resetTime int64) {
	ctx := context.Background()

	// Increment counter
	count, err := rl.redis.Incr(ctx, key)
	if err != nil {
		// If Redis fails, allow the request (fail open)
		return true, limit, time.Now().Add(window).Unix()
	}

	// Set expiration on first request
	if count == 1 {
		if err := rl.redis.Expire(ctx, key, window); err != nil {
			// Continue even if expire fails
		}
	}

	resetTime = time.Now().Add(window).Unix()

	// Check if limit exceeded
	if count > int64(limit) {
		return false, 0, resetTime
	}

	remaining = limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, resetTime
}
