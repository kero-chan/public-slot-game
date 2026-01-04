package logger

import (
	"github.com/google/wire"
	"github.com/slotmachine/backend/internal/config"
)

// ProviderSet is the Wire provider set for logger
var ProviderSet = wire.NewSet(
	ProvideLogger,
)

// ProvideLogger creates a new logger from config
func ProvideLogger(cfg *config.Config) *Logger {
	return New(cfg.Logging.Level, cfg.Logging.Format)
}
