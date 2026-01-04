package db

import (
	"github.com/google/wire"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"gorm.io/gorm"
)

// ProviderSet is the Wire provider set for database
var ProviderSet = wire.NewSet(
	ProvideDatabase,
)

// ProvideDatabase creates a new GORM database connection
func ProvideDatabase(cfg *config.Config, log *logger.Logger) (*gorm.DB, error) {
	return NewGormDB(cfg, log)
}
