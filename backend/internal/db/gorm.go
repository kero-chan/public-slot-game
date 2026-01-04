package db

import (
	"fmt"
	"time"

	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewGormDB creates a new GORM database connection
func NewGormDB(cfg *config.Config, log *logger.Logger) (*gorm.DB, error) {
	// Configure GORM logger with traceID and clientIP support
	var gormLogLevel gormlogger.LogLevel
	switch cfg.Logging.Level {
	case "debug", "info":
		gormLogLevel = gormlogger.Info
	case "warn":
		gormLogLevel = gormlogger.Warn
	default:
		gormLogLevel = gormlogger.Error
	}

	// Use custom GORM logger with trace support
	customLogger := NewGormLogger(log, time.Duration(cfg.Logging.SQLThresholdMilliSeconds)*time.Millisecond, cfg.Logging.SQLParameterizedQueries, gormLogLevel)

	gormConfig := &gorm.Config{
		Logger: customLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Database.Host).
		Str("dbname", cfg.Database.DBName).
		Msg("Database connection established")

	return db, nil
}

// Close closes the database connection
func Close(db *gorm.DB, log *logger.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Info().Msg("Database connection closed")
	return nil
}
