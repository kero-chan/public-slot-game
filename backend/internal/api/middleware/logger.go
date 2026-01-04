package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// LoggerMiddleware logs all HTTP requests
func LoggerMiddleware(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Log request
		duration := time.Since(start)
		status := c.Response().StatusCode()
		clientIP := c.Get("x-real-ip")
		if clientIP == "" {
			clientIP = c.IP()
		}
		if c.Path() != "/health" {
			log.Info().
				Str("method", c.Method()).
				Str("path", c.Path()).
				Int("status", status).
				Int64("duration", duration.Milliseconds()).
				Str("ip", clientIP).
				Str("user_agent", c.Get("User-Agent")).
				Msg("HTTP Request")
		}

		return err
	}
}
