package db

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/slotmachine/backend/internal/pkg/logger"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger implements GORM's logger interface with traceID and clientIP support
type GormLogger struct {
	logger               *logger.Logger
	SlowThreshold        time.Duration
	IgnoreRecordNotFound bool
	ParameterizedQueries bool
	LogLevel             gormlogger.LogLevel
}

// NewGormLogger creates a new GORM logger
func NewGormLogger(log *logger.Logger, slowThreshold time.Duration, parameterizedQueries bool, logLevel gormlogger.LogLevel) *GormLogger {
	return &GormLogger{
		logger:               log,
		SlowThreshold:        slowThreshold,
		IgnoreRecordNotFound: true,
		ParameterizedQueries: parameterizedQueries,
		LogLevel:             logLevel,
	}
}

// LogMode sets the log level
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info logs info level messages
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		log := l.logger.WithTraceContext(ctx)
		log.Info().Msgf(msg, data...)
	}
}

// Warn logs warning level messages
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		log := l.logger.WithTraceContext(ctx)
		log.Warn().Msgf(msg, data...)
	}
}

// Error logs error level messages
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		log := l.logger.WithTraceContext(ctx)
		log.Error().Msgf(msg, data...)
	}
}

// Trace logs SQL queries with execution time
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	sql = cleanSQL(sql)

	// Get trace info from context
	log := l.logger.WithTraceContext(ctx)

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFound):
		// Log SQL errors
		log.Error().
			Err(err).
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Msg("SQL error")

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		// Log slow queries
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		log.Warn().
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Str("threshold", l.SlowThreshold.String()).
			Msg(slowLog)

	case l.LogLevel == gormlogger.Info:
		// Log all queries in debug mode
		log.Info().
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Msg("SQL query")
	}
}

func (l *GormLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.ParameterizedQueries {
		return sql, nil
	}

	return sql, params
}

func cleanSQL(sql string) string {
	// Replace \n and \t with space
	clean := strings.ReplaceAll(sql, "\"", "")
	clean = strings.ReplaceAll(clean, "\n", " ")
	clean = strings.ReplaceAll(clean, "\t", " ")

	// Replace multiple spaces with one space
	space := regexp.MustCompile(`\s+`)
	clean = space.ReplaceAllString(clean, " ")

	return strings.TrimSpace(clean)
}
