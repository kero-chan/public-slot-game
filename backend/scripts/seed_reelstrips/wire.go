//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/db"
	"github.com/slotmachine/backend/internal/infra/repository"
	"github.com/slotmachine/backend/internal/pkg/cache"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/service"
)

// SeedApplication holds dependencies for the seed script
type SeedApplication struct {
	Config              *config.Config
	Logger              *logger.Logger
	ReelStripService    reelstrip.Service
	ReelStripRepository reelstrip.Repository
}

// InitializeSeedApplication creates a fully initialized seed application using Wire
func InitializeSeedApplication() (*SeedApplication, error) {
	wire.Build(
		// Config
		config.ProviderSet,

		// Logger
		logger.ProviderSet,

		// Database
		db.ProviderSet,

		// Repository
		repository.NewReelStripGormRepository,

		// Service
		service.NewReelStripService,

		// cache
		cache.ProvideCache,

		// Application struct
		wire.Struct(new(SeedApplication), "*"),
	)

	return &SeedApplication{}, nil
}
