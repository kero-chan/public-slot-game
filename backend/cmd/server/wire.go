//go:build wireinject
// +build wireinject

package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	adminDomain "github.com/slotmachine/backend/domain/admin"
	playerDomain "github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/internal/api/handler"
	"github.com/slotmachine/backend/internal/api/middleware"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/db"
	"github.com/slotmachine/backend/internal/game/engine"
	"github.com/slotmachine/backend/internal/infra/repository"
	"github.com/slotmachine/backend/internal/infra/storage"
	"github.com/slotmachine/backend/internal/pkg/cache"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/server"
	"github.com/slotmachine/backend/internal/service"
	"gorm.io/gorm"
)

// Application holds all application dependencies
type Application struct {
	Config                       *config.Config
	Logger                       *logger.Logger
	DB                           *gorm.DB
	Cache                        *cache.Cache
	App                          *fiber.App
	RateLimiter                  *middleware.RateLimiter
	TrialRateLimiter             *middleware.TrialRateLimiter // Security: DoS protection for trial mode
	AuthHandler                  *handler.AuthHandler
	PlayerHandler                *handler.PlayerHandler
	SessionHandler               *handler.SessionHandler
	SpinHandler                  *handler.SpinHandler
	FreeSpinsHandler             *handler.FreeSpinsHandler
	ProvablyFairHandler          *handler.ProvablyFairHandler
	ProvablyFairService          *service.ProvablyFairService
	SpinService                  *service.SpinService      // For PF injection
	FreeSpinsService             *service.FreeSpinsService // For PF injection
	AdminReelStripHandler        *handler.AdminReelStripHandler
	AdminPlayerAssignmentHandler *handler.AdminPlayerAssignmentHandler
	AdminAuthHandler             *handler.AdminAuthHandler
	AdminManagementHandler       *handler.AdminManagementHandler
	AdminPlayerHandler           *handler.AdminPlayerHandler
	GameHandler                  *handler.GameHandler
	AdminGameHandler             *handler.AdminGameHandler
	AdminUploadHandler           *handler.AdminUploadHandler
	AdminChunkedUploadHandler    *handler.AdminChunkedUploadHandler
	AdminDirectUploadHandler     *handler.AdminDirectUploadHandler
	TrialHandler                 *handler.TrialHandler
	// Trial-specific handlers (separate from production)
	TrialSpinHandler             *handler.TrialSpinHandler
	TrialFreeSpinsHandler        *handler.TrialFreeSpinsHandler
	TrialSessionHandler          *handler.TrialSessionHandler
	TrialPlayerHandler           *handler.TrialPlayerHandler
	AdminService                 adminDomain.Service
	PlayerService                playerDomain.Service
	TrialService                 *service.TrialService
	Storage                      storage.Storage
}

// InitializeApplication creates a fully initialized application using Wire
func InitializeApplication() (*Application, error) {
	wire.Build(
		// Config
		config.ProviderSet,

		// Logger
		logger.ProviderSet,

		// Database
		db.ProviderSet,

		// Game Engine
		engine.ProviderSet,

		// Repositories
		repository.ProviderSet,

		// Storage
		storage.ProviderSet,

		// Services
		service.ProviderSet,

		// Handlers
		handler.ProviderSet,

		// Fiber App
		server.ProviderSet,

		// Cache
		cache.ProviderSet,

		// Middleware
		middleware.ProviderSet,

		// Application struct
		wire.Struct(new(Application), "*"),
	)

	return &Application{}, nil
}

// Shutdown gracefully shuts down all application resources
func (a *Application) Shutdown() error {
	a.Logger.Info().Msg("Starting graceful shutdown...")

	// Shutdown Fiber server
	if err := a.App.Shutdown(); err != nil {
		a.Logger.Error().Err(err).Msg("Failed to shutdown Fiber server")
	} else {
		a.Logger.Info().Msg("Fiber server shutdown complete")
	}

	// Close cache (which includes Redis pub/sub cleanup)
	if a.Cache != nil {
		a.Cache.Close()
		a.Logger.Info().Msg("Cache closed")
	}

	// Close database connection
	if a.DB != nil {
		if err := db.Close(a.DB, a.Logger); err != nil {
			a.Logger.Error().Err(err).Msg("Failed to close database")
			return err
		}
	}

	a.Logger.Info().Msg("Graceful shutdown complete")
	return nil
}
