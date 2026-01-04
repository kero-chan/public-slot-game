package ctxutil

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// Context keys
type contextKey string

const (
	TraceIDKey  contextKey = "trace_id"
	ClientIPKey contextKey = "client_ip"
)

// WithTraceInfo adds traceID and clientIP from fiber context to Go context
func WithTraceInfo(ctx context.Context, c *fiber.Ctx) context.Context {
	if traceID, ok := c.Locals("trace_id").(string); ok && traceID != "" {
		ctx = context.WithValue(ctx, TraceIDKey, traceID)
	}

	if clientIP, ok := c.Locals("client_ip").(string); ok && clientIP != "" {
		ctx = context.WithValue(ctx, ClientIPKey, clientIP)
	}

	return ctx
}

// GetTraceID extracts traceID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// GetClientIP extracts clientIP from context
func GetClientIP(ctx context.Context) string {
	if clientIP, ok := ctx.Value(ClientIPKey).(string); ok {
		return clientIP
	}
	return ""
}
