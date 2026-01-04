package logger

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog logger
type Logger struct {
	logger *zerolog.Logger
}

// New creates a new logger instance
func New(level, format string) *Logger {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// Configure output format
	var logger zerolog.Logger
	if format == "pretty" || format == "console" {
		logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Caller().Logger()
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}

	return &Logger{
		logger: &logger,
	}
}

// Info returns a zerolog event for info logging (supports chaining)
func (l *Logger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Debug returns a zerolog event for debug logging (supports chaining)
func (l *Logger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Warn returns a zerolog event for warn logging (supports chaining)
func (l *Logger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error returns a zerolog event for error logging (supports chaining)
func (l *Logger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// WithField returns a new logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &Logger{logger: &newLogger}
}

// WithFields returns a new logger with multiple additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := l.logger.With().Fields(fields).Logger()
	return &Logger{logger: &newLogger}
}

// GetZerolog returns the underlying zerolog logger
func (l *Logger) GetZerolog() *zerolog.Logger {
	return l.logger
}

// WithTrace returns a logger with traceID and clientIP from fiber context
func (l *Logger) WithTrace(c *fiber.Ctx) *Logger {
	traceID, _ := c.Locals("trace_id").(string)
	clientIP, _ := c.Locals("client_ip").(string)

	newLogger := l.logger.With().
		Str("trace_id", traceID).
		Str("client_ip", clientIP).
		Logger()

	return &Logger{logger: &newLogger}
}

// WithTraceContext returns a logger with traceID and clientIP from context
func (l *Logger) WithTraceContext(ctx context.Context) *Logger {
	traceID, _ := ctx.Value("trace_id").(string)
	clientIP, _ := ctx.Value("client_ip").(string)

	if traceID == "" && clientIP == "" {
		return l
	}

	newLogger := l.logger.With().
		Str("trace_id", traceID).
		Str("client_ip", clientIP).
		Logger()

	return &Logger{logger: &newLogger}
}
