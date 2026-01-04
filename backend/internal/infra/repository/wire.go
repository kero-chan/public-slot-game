package repository

import (
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet is the Wire provider set for repositories
var ProviderSet = wire.NewSet(
	NewPlayerGormRepository,
	NewSessionGormRepository,
	NewPlayerSessionGormRepository,
	NewSpinGormRepository,
	NewFreeSpinsGormRepository,
	NewReelStripGormRepository,
	NewAdminGormRepository,
	NewGameGormRepository,
	NewProvablyFairGormRepository,
	NewTxManager,
)

// ProvideDB is a provider function for *gorm.DB
// This is used when you need to inject the database separately
func ProvideDB(db *gorm.DB) *gorm.DB {
	return db
}
