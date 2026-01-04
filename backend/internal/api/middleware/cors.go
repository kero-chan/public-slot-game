package middleware

// Note: CORS is handled by Fiber's built-in middleware in server/fiber.go
// This file is kept for reference and custom CORS logic if needed

import (
	"github.com/gofiber/fiber/v2"
)

// CustomCORSMiddleware provides custom CORS handling if needed
func CustomCORSMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Custom CORS logic can be added here if needed
		// For now, we use Fiber's built-in CORS middleware
		return c.Next()
	}
}
