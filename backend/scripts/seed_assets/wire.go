//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/slotmachine/backend/domain/game"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/db"
	"github.com/slotmachine/backend/internal/infra/repository"
	"github.com/slotmachine/backend/internal/infra/storage"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// SeedApplication holds dependencies for the seed script
type SeedApplication struct {
	Config         *config.Config
	Logger         *logger.Logger
	GameRepository game.Repository
	Storage        storage.Storage
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
		repository.NewGameGormRepository,

		// Storage
		storage.ProviderSet,

		// Application struct
		wire.Struct(new(SeedApplication), "*"),
	)

	return &SeedApplication{}, nil
}
