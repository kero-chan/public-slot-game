package server

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/middleware"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// NewFiberApp creates and configures a new Fiber app
func NewFiberApp(cfg *config.Config, log *logger.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               cfg.App.Name,
		ServerHeader:          cfg.App.Name,
		DisableStartupMessage: false,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           60 * time.Second,
		BodyLimit:             3 * 1024 * 1024,
		// StreamRequestBody:     true,            // Stream request body to reduce memory usage for large uploads
		ErrorHandler: customErrorHandler(log),
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	})

	// Recover middleware - must be first
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.App.Env == "development",
	}))

	// Trace middleware - add traceID and clientIP to all requests
	app.Use(middleware.TraceMiddleware())

	// Request logging middleware
	app.Use(requestLogger(log))

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     cfg.CORS.AllowedMethods,
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Compression middleware
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	return app
}

// requestLogger logs all HTTP requests with traceID and clientIP
func requestLogger(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Log request with trace info
		duration := time.Since(start)
		tracedLog := log.WithTrace(c)

		if c.Path() != "/health" {
			tracedLog.Info().
				Str("method", c.Method()).
				Str("path", c.Path()).
				Int("status", c.Response().StatusCode()).
				Dur("duration", duration).
				Str("user_agent", c.Get("User-Agent")).
				Msg("HTTP request")
		}

		return err
	}
}

// customErrorHandler handles Fiber errors with traceID and clientIP
func customErrorHandler(log *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default error code
		code := fiber.StatusInternalServerError

		// Check if it's a Fiber error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		// Log the error with trace info
		tracedLog := log.WithTrace(c)
		tracedLog.Error().
			Err(err).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", code).
			Msg("Request error")

		// Return JSON error response
		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
	}
}
