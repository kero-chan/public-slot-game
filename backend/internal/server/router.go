package server

import (
	"github.com/gofiber/fiber/v2"
	adminDomain "github.com/slotmachine/backend/domain/admin"
	playerDomain "github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/internal/api/handler"
	"github.com/slotmachine/backend/internal/api/middleware"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/service"
)

// SetupRoutes sets up all application routes
func SetupRoutes(
	app *fiber.App,
	cfg *config.Config,
	log *logger.Logger,
	rateLimiter *middleware.RateLimiter,
	trialRateLimiter *middleware.TrialRateLimiter,
	authHandler *handler.AuthHandler,
	playerHandler *handler.PlayerHandler,
	sessionHandler *handler.SessionHandler,
	spinHandler *handler.SpinHandler,
	freeSpinsHandler *handler.FreeSpinsHandler,
	provablyFairHandler *handler.ProvablyFairHandler,
	adminReelStripHandler *handler.AdminReelStripHandler,
	adminPlayerAssignmentHandler *handler.AdminPlayerAssignmentHandler,
	adminAuthHandler *handler.AdminAuthHandler,
	adminManagementHandler *handler.AdminManagementHandler,
	adminPlayerHandler *handler.AdminPlayerHandler,
	gameHandler *handler.GameHandler,
	adminGameHandler *handler.AdminGameHandler,
	adminUploadHandler *handler.AdminUploadHandler,
	adminChunkedUploadHandler *handler.AdminChunkedUploadHandler,
	adminDirectUploadHandler *handler.AdminDirectUploadHandler,
	trialHandler *handler.TrialHandler,
	// Trial-specific handlers (separate endpoints from production)
	trialSpinHandler *handler.TrialSpinHandler,
	trialFreeSpinsHandler *handler.TrialFreeSpinsHandler,
	trialSessionHandler *handler.TrialSessionHandler,
	trialPlayerHandler *handler.TrialPlayerHandler,
	adminService adminDomain.Service,
	playerService playerDomain.Service,
	trialService *service.TrialService,
) {
	// Health check endpoint (no auth required)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"status": "healthy",
			},
		})
	})

	// API v1 routes
	v1 := app.Group("/v1")

	// Rate limiters
	publicRateLimiter := rateLimiter.PublicMiddleware()
	authRateLimiter := rateLimiter.AuthenticatedMiddleware()

	// Session-based auth middleware (pure session, no JWT)
	// Now supports trial tokens (prefixed with "trial_")
	sessionAuthMiddleware := middleware.SessionAuthMiddleware(log, playerService, trialService)

	// Public routes (no auth) - Apply public rate limiter
	auth := v1.Group("/auth")
	auth.Use(publicRateLimiter)
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", sessionAuthMiddleware, authHandler.Logout)
	// Trial mode - start a trial session (no auth required)
	// SECURITY: Protected by TrialRateLimiter to prevent DoS attacks
	// - Max 3 concurrent sessions per IP
	// - 5 minute cooldown between session creation
	// - Global session limit of 100,000
	auth.Post("/trial", trialRateLimiter.TrialCreationMiddleware(), trialHandler.StartTrial)

	// Trial routes (require trial auth) - completely separate from production
	trial := v1.Group("/trial")
	trial.Use(sessionAuthMiddleware, authRateLimiter)
	// Trial profile/balance (original)
	trial.Get("/profile", trialHandler.GetTrialProfile)
	trial.Get("/balance", trialHandler.GetTrialBalance)
	// Trial player (new dedicated handler)
	trial.Get("/player/balance", trialPlayerHandler.GetBalance)
	// Trial session
	trial.Post("/session/start", trialSessionHandler.StartSession)
	// Trial spin
	trial.Post("/spin", trialSpinHandler.ExecuteSpin)
	// Trial free spins
	trial.Get("/free-spins/status", trialFreeSpinsHandler.GetStatus)
	trial.Post("/free-spins/spin", trialFreeSpinsHandler.ExecuteFreeSpin)

	// Initial grid for display (no auth required - cosmetic only)
	v1.Get("/initial-grid", publicRateLimiter, spinHandler.GetInitialGrid)

	// Game assets (no auth required - needed for game initialization)
	v1.Get("/game-assets", publicRateLimiter, gameHandler.GetGameAssets)

	// Protected routes (require session auth) - Apply authenticated rate limiter

	// Player routes
	player := v1.Group("/player")
	player.Use(sessionAuthMiddleware, authRateLimiter)
	player.Get("/profile", authHandler.GetProfile)
	player.Get("/balance", playerHandler.GetBalance)

	// Session routes
	session := v1.Group("/session")
	session.Use(sessionAuthMiddleware, authRateLimiter)
	session.Post("/start", sessionHandler.StartSession)
	session.Post("/:sessionId/end", sessionHandler.EndSession)
	session.Get("/history", sessionHandler.GetSessionHistory)

	// Spin routes
	spin := v1.Group("/base-spins")
	spin.Use(sessionAuthMiddleware, authRateLimiter)
	spin.Post("/spin", spinHandler.ExecuteSpin)
	spin.Get("/histories", spinHandler.GetSpinHistory)

	// Free spins routes
	freeSpins := v1.Group("/free-spins")
	freeSpins.Use(sessionAuthMiddleware, authRateLimiter)
	freeSpins.Get("/status", freeSpinsHandler.GetStatus)
	freeSpins.Post("/spin", freeSpinsHandler.ExecuteFreeSpin)

	// Provably Fair routes
	pf := v1.Group("/pf")
	pf.Use(sessionAuthMiddleware, authRateLimiter)
	pf.Post("/sessions", provablyFairHandler.StartPFSession)              // Start a new PF session
	pf.Post("/sessions/end", provablyFairHandler.EndPFSession)            // End session and reveal seed
	pf.Get("/sessions/status", provablyFairHandler.GetPFSessionStatus)    // Get current session status
	pf.Post("/sessions/verify-spin", provablyFairHandler.VerifyActiveSpin) // Verify spin in active session

	// Verification routes (can be public for third-party verification)
	pfVerify := v1.Group("/pf/verify")
	pfVerify.Use(authRateLimiter)                                            // Only rate limit, no auth required for verification
	pfVerify.Get("/:sessionId", provablyFairHandler.GetVerificationData)     // Get verification data
	pfVerify.Post("/spin", provablyFairHandler.VerifySpin)                   // Verify single spin hash
	pfVerify.Post("/spin-with-reel", provablyFairHandler.VerifySpinWithReel) // Verify spin + reel positions
	pfVerify.Post("/:sessionId", provablyFairHandler.VerifySession)          // Verify session hash chain

	// Admin routes
	admin := v1.Group("/admin")

	// Admin Auth (no auth required for login) - Apply public rate limiter
	adminAuth := admin.Group("/auth")
	adminAuth.Post("/login", publicRateLimiter, adminAuthHandler.Login)

	// Admin auth middleware
	adminAuthMiddleware := middleware.AdminAuthMiddleware(cfg, log, adminService)

	// Protected admin routes (require admin auth) - Apply authenticated rate limiter
	adminAuth.Get("/profile", adminAuthMiddleware, authRateLimiter, adminAuthHandler.GetProfile)
	adminAuth.Post("/change-password", adminAuthMiddleware, authRateLimiter, adminAuthHandler.ChangePassword)

	// Admin Management (require admin auth)
	adminMgmt := admin.Group("/management")
	adminMgmt.Use(adminAuthMiddleware, authRateLimiter)

	// Admin user management
	adminUsers := adminMgmt.Group("/admins")
	adminUsers.Post("/", adminManagementHandler.CreateAdmin)
	adminUsers.Get("/", adminManagementHandler.ListAdmins)
	adminUsers.Get("/:id", adminManagementHandler.GetAdmin)
	adminUsers.Put("/:id", adminManagementHandler.UpdateAdmin)
	adminUsers.Delete("/:id", adminManagementHandler.DeleteAdmin)
	adminUsers.Post("/:id/reset-password", adminManagementHandler.ResetPassword)
	adminUsers.Post("/:id/activate", adminManagementHandler.ActivateAdmin)
	adminUsers.Post("/:id/deactivate", adminManagementHandler.DeactivateAdmin)
	adminUsers.Post("/:id/suspend", adminManagementHandler.SuspendAdmin)

	// Admin - Reel Strip Config Management
	adminReelConfigs := admin.Group("/reel-strip-configs")
	adminReelConfigs.Use(adminAuthMiddleware, authRateLimiter)
	adminReelConfigs.Post("/", adminReelStripHandler.CreateConfig)
	adminReelConfigs.Get("/", adminReelStripHandler.ListConfigs)
	adminReelConfigs.Get("/:id", adminReelStripHandler.GetConfig)
	adminReelConfigs.Put("/:id", adminReelStripHandler.UpdateConfig)
	adminReelConfigs.Post("/:id/activate", adminReelStripHandler.ActivateConfig)
	adminReelConfigs.Post("/:id/deactivate", adminReelStripHandler.DeactivateConfig)
	adminReelConfigs.Post("/set-default", adminReelStripHandler.SetDefaultConfig)

	// Admin - Player Assignment Management
	adminAssignments := admin.Group("/player-assignments")
	adminAssignments.Use(adminAuthMiddleware, authRateLimiter)
	adminAssignments.Post("/", adminPlayerAssignmentHandler.CreateAssignment)
	adminAssignments.Get("/:playerId", adminPlayerAssignmentHandler.GetPlayerAssignment)
	adminAssignments.Put("/:playerId", adminPlayerAssignmentHandler.UpdateAssignment)
	adminAssignments.Post("/:playerId/assign", adminPlayerAssignmentHandler.AssignConfigToPlayer)
	adminAssignments.Delete("/:playerId", adminPlayerAssignmentHandler.RemoveAssignment)

	// Admin - Player Management
	adminPlayers := admin.Group("/players")
	adminPlayers.Use(adminAuthMiddleware, authRateLimiter)
	adminPlayers.Post("/", adminPlayerHandler.CreatePlayer)
	adminPlayers.Get("/", adminPlayerHandler.ListPlayers)
	adminPlayers.Get("/:id", adminPlayerHandler.GetPlayer)
	adminPlayers.Post("/:id/activate", adminPlayerHandler.ActivatePlayer)
	adminPlayers.Post("/:id/deactivate", adminPlayerHandler.DeactivatePlayer)
	adminPlayers.Post("/:id/force-logout", adminPlayerHandler.ForceLogoutPlayer)

	// Admin - Game Management
	adminGames := admin.Group("/games")
	adminGames.Use(adminAuthMiddleware, authRateLimiter)
	adminGames.Get("/", adminGameHandler.ListGames)
	adminGames.Get("/:id", adminGameHandler.GetGame)
	adminGames.Post("/", adminGameHandler.CreateGame)
	adminGames.Put("/:id", adminGameHandler.UpdateGame)
	adminGames.Delete("/:id", adminGameHandler.DeleteGame)
	adminGames.Post("/:id/activate", adminGameHandler.ActivateGame)
	adminGames.Post("/:id/deactivate", adminGameHandler.DeactivateGame)

	// Admin - Asset Management
	adminAssets := admin.Group("/assets")
	adminAssets.Use(adminAuthMiddleware, authRateLimiter)
	adminAssets.Get("/", adminGameHandler.ListAssets)
	adminAssets.Get("/:id", adminGameHandler.GetAsset)
	adminAssets.Post("/", adminGameHandler.CreateAsset)
	adminAssets.Put("/:id", adminGameHandler.UpdateAsset)
	adminAssets.Delete("/:id", adminGameHandler.DeleteAsset)
	adminAssets.Post("/:id/activate", adminGameHandler.ActivateAsset)
	adminAssets.Post("/:id/deactivate", adminGameHandler.DeactivateAsset)

	// Admin - Game Config Management (link games to assets)
	adminGameConfigs := admin.Group("/game-configs")
	adminGameConfigs.Use(adminAuthMiddleware, authRateLimiter)
	adminGameConfigs.Get("/", adminGameHandler.ListGameConfigs)
	adminGameConfigs.Get("/:id", adminGameHandler.GetGameConfig)
	adminGameConfigs.Post("/", adminGameHandler.CreateGameConfig)
	adminGameConfigs.Delete("/:id", adminGameHandler.DeleteGameConfig)
	adminGameConfigs.Post("/:id/activate", adminGameHandler.ActivateGameConfig)
	adminGameConfigs.Post("/:id/deactivate", adminGameHandler.DeactivateGameConfig)

	// Hide route upload use only direct-upload for now
	// Admin - File Upload Management
	adminUpload := admin.Group("/upload")
	adminUpload.Use(adminAuthMiddleware, authRateLimiter)
	// adminUpload.Post("/:theme", adminUploadHandler.UploadFile)
	// adminUpload.Post("/:theme/batch", adminUploadHandler.UploadMultipleFiles)
	adminUpload.Get("/:theme/files", adminUploadHandler.ListFiles)
	adminUpload.Delete("/:theme", adminUploadHandler.DeleteTheme)
	adminUpload.Delete("/:theme/*", adminUploadHandler.DeleteFile)

	// // Admin - Chunked File Upload (for large files)
	// adminUpload.Post("/:theme/chunked/init", adminChunkedUploadHandler.InitChunkedUpload)
	// adminUpload.Post("/:theme/chunked/:uploadId/chunk", adminChunkedUploadHandler.UploadChunk)
	// adminUpload.Post("/:theme/chunked/:uploadId/complete", adminChunkedUploadHandler.CompleteChunkedUpload)
	// adminUpload.Get("/:theme/chunked/:uploadId/status", adminChunkedUploadHandler.GetUploadStatus)
	// adminUpload.Delete("/:theme/chunked/:uploadId", adminChunkedUploadHandler.AbortChunkedUpload)

	// // Admin - Processing Status (for background upload processing)
	// adminUpload.Get("/status/:uploadId", adminChunkedUploadHandler.GetProcessingStatus)

	// Admin - Direct Upload (presigned URL for client-side upload to storage)
	adminDirectUpload := admin.Group("/direct-upload")
	adminDirectUpload.Use(adminAuthMiddleware)
	adminDirectUpload.Post("/:theme/presign", adminDirectUploadHandler.GeneratePresignedURL)
	adminDirectUpload.Post("/:theme/presign/batch", adminDirectUploadHandler.GenerateBatchPresignedURLs)
	adminDirectUpload.Post("/:theme/confirm", adminDirectUploadHandler.ConfirmUpload)
	adminDirectUpload.Post("/:theme/confirm/batch", adminDirectUploadHandler.ConfirmBatchUpload)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "Route not found",
			},
		})
	})
}
