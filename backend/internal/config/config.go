package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	App          AppConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	JWT          JWTConfig
	Logging      LoggingConfig
	CORS         CORSConfig
	RateLimit    RateLimitConfig
	Trial        TrialConfig
	Game         GameConfig
	Storage      StorageConfig
	ProvablyFair ProvablyFairConfig
}

// AppConfig holds application-level settings
type AppConfig struct {
	Env  string
	Addr string
	Name string
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis connection settings
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Enabled  bool
}

// JWTConfig holds JWT authentication settings
type JWTConfig struct {
	Secret          string
	ExpirationHours int
	// ServiceToken is a static token for service-to-service auth (never expires)
	// Used by internal services
	ServiceToken string
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level                    string
	Format                   string
	SQLThresholdMilliSeconds int
	SQLParameterizedQueries  bool
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	AllowedOrigins string
	AllowedMethods string
	AllowedHeaders string
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	SpinLimit    int
	GeneralLimit int
}

// TrialConfig holds trial mode security settings
type TrialConfig struct {
	// MaxSessionsPerIP is the maximum number of concurrent trial sessions per IP
	// Set higher for office environments with shared IP (NAT)
	MaxSessionsPerIP int
	// SessionCooldownSeconds is the minimum time between creating new trial sessions from same IP
	SessionCooldownSeconds int
	// Enabled controls whether trial mode is available
	Enabled bool
	// WhitelistedIPs is a comma-separated list of IPs that bypass rate limiting
	// Useful for internal testing, office networks, CI/CD pipelines
	// Example: "10.0.0.0/8,192.168.0.0/16,172.16.0.0/12" for private networks
	WhitelistedIPs string
	// WhitelistedMaxSessions is max sessions for whitelisted IPs (0 = unlimited)
	WhitelistedMaxSessions int
	// TrustedProxies is a comma-separated list of proxy IPs/CIDRs that are trusted
	// Only trust X-Real-IP/X-Forwarded-For headers from these proxies
	// Example: "10.0.0.0/8,172.16.0.0/12" for internal load balancers
	// If empty, falls back to direct connection IP (c.IP())
	TrustedProxies string
}

// GameConfig holds game-specific settings
type GameConfig struct {
	MinBet           float64
	MaxBet           float64
	BetStep          float64
	DefaultBalance   float64
	TargetRTP        float64
	MaxWinMultiplier int
}

// StorageConfig holds S3/MinIO/GCS storage settings
type StorageConfig struct {
	// Provider can be "minio" or "gcs"
	Provider string
	// MinIO/S3 specific settings
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	// Common settings
	BucketName string
	UseSSL     bool
	PublicURL  string
}

// ProvablyFairConfig holds provably fair gaming settings
type ProvablyFairConfig struct {
	// EncryptionKey is the 32-byte key for AES-256-GCM encryption of server seeds
	// Used to encrypt server_seed before storing in database for recovery
	EncryptionKey string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if in development
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			// .env file is optional, so just log a warning
			fmt.Println("Warning: .env file not found, using environment variables")
		}
	}

	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Addr: getEnv("APP_ADDR", ":8080"),
			Name: getEnv("APP_NAME", "SlotMachine"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "slotmachine"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			Enabled:  getEnvAsBool("REDIS_ENABLED", true),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "change-this-secret-in-production"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			ServiceToken:    getEnv("SERVICE_TOKEN", ""),
		},
		Logging: LoggingConfig{
			Level:                    getEnv("LOG_LEVEL", "debug"),
			Format:                   getEnv("LOG_FORMAT", "json"),
			SQLThresholdMilliSeconds: getEnvAsInt("LOG_SQL_THRESHOLD_MILLI_SECONDS", 200),
			SQLParameterizedQueries:  getEnvAsBool("LOG_SQL_PARAMETERIZED_QUERIES", false),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
			AllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
			AllowedHeaders: getEnv("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization,X-Game-ID"),
		},
		RateLimit: RateLimitConfig{
			SpinLimit:    getEnvAsInt("RATE_LIMIT_SPIN", 10),
			GeneralLimit: getEnvAsInt("RATE_LIMIT_GENERAL", 100),
		},
		Trial: TrialConfig{
			MaxSessionsPerIP:       getEnvAsInt("TRIAL_MAX_SESSIONS_PER_IP", 5),
			SessionCooldownSeconds: getEnvAsInt("TRIAL_SESSION_COOLDOWN_SECONDS", 60), // 1 minute
			Enabled:                getEnvAsBool("TRIAL_ENABLED", true),
			// Default: whitelist private networks (RFC 1918) for office/internal testing
			WhitelistedIPs:         getEnv("TRIAL_WHITELISTED_IPS", "10.0.0.0/8,192.168.0.0/16,172.16.0.0/12,127.0.0.1"),
			WhitelistedMaxSessions: getEnvAsInt("TRIAL_WHITELISTED_MAX_SESSIONS", 50), // Higher limit for internal
			// Default: trust private networks as reverse proxies (common in Docker/K8s)
			TrustedProxies: getEnv("TRIAL_TRUSTED_PROXIES", "10.0.0.0/8,192.168.0.0/16,172.16.0.0/12,127.0.0.1"),
		},
		Game: GameConfig{
			MinBet:           getEnvAsFloat("MIN_BET", 1.00),
			MaxBet:           getEnvAsFloat("MAX_BET", 1000.00),
			BetStep:          getEnvAsFloat("BET_STEP", 1.00),
			DefaultBalance:   getEnvAsFloat("DEFAULT_BALANCE", 100000.00),
			TargetRTP:        getEnvAsFloat("TARGET_RTP", 96.5),
			MaxWinMultiplier: getEnvAsInt("MAX_WIN_MULTIPLIER", 25000),
		},
		Storage: StorageConfig{
			Provider:        getEnv("STORAGE_PROVIDER", "minio"), // "minio" or "gcs"
			Endpoint:        getEnv("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("STORAGE_SECRET_KEY", "minioadmin"),
			BucketName:      getEnv("STORAGE_BUCKET", "slot-assets"),
			UseSSL:          getEnvAsBool("STORAGE_USE_SSL", false),
			PublicURL:       getEnv("STORAGE_PUBLIC_URL", "http://localhost:9000"),
		},
		ProvablyFair: ProvablyFairConfig{
			// Default key for development only - MUST be overridden in production
			EncryptionKey: getEnv("PF_ENCRYPTION_KEY", "provablyfair-dev-key-32bytes!!!!"),
		},
	}

	// Validate critical settings
	if cfg.JWT.Secret == "change-this-secret-in-production" && cfg.App.Env == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	if cfg.Database.Password == "" && cfg.App.Env == "production" {
		return nil, fmt.Errorf("DB_PASSWORD must be set in production")
	}

	return cfg, nil
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
