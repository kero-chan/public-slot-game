package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Context keys for trace information
const (
	TraceIDKey  = "trace_id"
	ClientIPKey = "client_ip"
)

// TraceMiddleware adds traceID and client IP to request context
func TraceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Generate or extract trace ID
		traceID := c.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Extract client IP
		clientIP := c.Get("x-real-ip")
		if clientIP == "" {
			clientIP = c.IP()
		}

		// Store in context
		c.Locals(TraceIDKey, traceID)
		c.Locals(ClientIPKey, clientIP)

		// Add to response headers
		c.Set("X-Trace-ID", traceID)

		return c.Next()
	}
}

// GetTraceID extracts trace ID from fiber context
func GetTraceID(c *fiber.Ctx) string {
	if traceID, ok := c.Locals(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// GetClientIP extracts client IP from fiber context
func GetClientIP(c *fiber.Ctx) string {
	if clientIP, ok := c.Locals(ClientIPKey).(string); ok {
		return clientIP
	}
	return ""
}
