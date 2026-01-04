package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/slotmachine/backend/internal/server"
)

func main() {
	// Initialize application with Wire
	application, err := InitializeApplication()
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	log := application.Logger
	cfg := application.Config

	log.Info().
		Str("env", cfg.App.Env).
		Str("addr", cfg.App.Addr).
		Msg("Starting Slot Machine Backend Server")

	// Inject PF service into session handler (for integrated PF session management)
	application.SessionHandler.SetProvablyFairService(application.ProvablyFairService)
	log.Info().Msg("PF service injected into session handler")

	// Inject trial service into spin service (for trial mode support)
	application.SpinService.SetTrialService(application.TrialService)
	log.Info().Msg("Trial service injected into spin service")

	// Setup routes
	server.SetupRoutes(
		application.App,
		cfg,
		log,
		application.RateLimiter,
		application.TrialRateLimiter, // Security: DoS protection for trial mode
		application.AuthHandler,
		application.PlayerHandler,
		application.SessionHandler,
		application.SpinHandler,
		application.FreeSpinsHandler,
		application.ProvablyFairHandler,
		application.AdminReelStripHandler,
		application.AdminPlayerAssignmentHandler,
		application.AdminAuthHandler,
		application.AdminManagementHandler,
		application.AdminPlayerHandler,
		application.GameHandler,
		application.AdminGameHandler,
		application.AdminUploadHandler,
		application.AdminChunkedUploadHandler,
		application.AdminDirectUploadHandler,
		application.TrialHandler,
		// Trial-specific handlers (separate from production)
		application.TrialSpinHandler,
		application.TrialFreeSpinsHandler,
		application.TrialSessionHandler,
		application.TrialPlayerHandler,
		application.AdminService,
		application.PlayerService,
		application.TrialService,
	)

	// Start server in a goroutine
	go func() {
		log.Info().Str("addr", cfg.App.Addr).Msg("Server listening")
		if err := application.App.Listen(cfg.App.Addr); err != nil {
			log.Error().Err(err).Msg("Failed to start server")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Gracefully shutdown all resources (Fiber, Redis, Database)
	if err := application.Shutdown(); err != nil {
		log.Error().Err(err).Msg("Shutdown error")
		os.Exit(1)
	}

	log.Info().Msg("Server stopped")
}
