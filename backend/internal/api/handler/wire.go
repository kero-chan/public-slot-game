package handler

import (
	"github.com/google/wire"
)

// ProviderSet is the Wire provider set for HTTP handlers
var ProviderSet = wire.NewSet(
	NewAuthHandler,
	NewPlayerHandler,
	NewSessionHandler,
	NewSpinHandler,
	NewFreeSpinsHandler,
	NewAdminReelStripHandler,
	NewAdminPlayerAssignmentHandler,
	NewAdminAuthHandler,
	NewAdminManagementHandler,
	NewAdminPlayerHandler,
	NewGameHandler,
	NewAdminGameHandler,
	NewAdminUploadHandler,
	NewAdminChunkedUploadHandler,
	NewAdminDirectUploadHandler,
	NewProvablyFairHandler,
	NewTrialHandler,
	// Trial-specific handlers (separate from production)
	NewTrialSpinHandler,
	NewTrialFreeSpinsHandler,
	NewTrialSessionHandler,
	NewTrialPlayerHandler,
)
