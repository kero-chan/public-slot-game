package reelstrip

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Service defines the business logic interface for reel strips
type Service interface {
	// GetReelSetForPlayer retrieves the reel strip set for a specific player
	// This is the main method to use - it handles player assignments, defaults, and fallbacks
	GetReelSetForPlayer(ctx context.Context, playerID uuid.UUID, gameMode string) (*ReelStripConfigSet, error)

	// GetReelSetByConfig retrieves a reel strip set by configuration ID
	GetReelSetByConfig(ctx context.Context, configID uuid.UUID) (*ReelStripConfigSet, error)

	// GetDefaultReelSet retrieves the default reel strip set for a game mode
	GetDefaultReelSet(ctx context.Context, gameMode string) (*ReelStripConfigSet, error)

	// ReelStripConfig management
	CreateConfig(ctx context.Context, name, gameMode, description string, reelStripIDs [5]uuid.UUID, targetRTP float64, extraInfoJSON []byte) (*ReelStripConfig, error)
	GetConfigByID(ctx context.Context, id uuid.UUID) (*ReelStripConfig, error)
	GetConfigByName(ctx context.Context, name string) (*ReelStripConfig, error)
	ListConfigs(ctx context.Context, filters *ConfigListFilters) ([]*ReelStripConfig, int64, error)
	SetDefaultConfig(ctx context.Context, configID uuid.UUID, gameMode string) error
	ActivateConfig(ctx context.Context, configID uuid.UUID) error
	DeactivateConfig(ctx context.Context, configID uuid.UUID) error

	// Player assignment management
	AssignConfigToPlayer(ctx context.Context, playerID, configID uuid.UUID, gameMode, reason, assignedBy string, expiresAt *time.Time) error
	GetPlayerAssignment(ctx context.Context, playerID uuid.UUID) (*PlayerReelStripAssignment, error)
	RemovePlayerAssignment(ctx context.Context, playerID uuid.UUID) error

	// ReelStrip operations (for creating configs)
	GenerateAndSaveStrips(ctx context.Context, gameMode string, count int, version int) error
	GenerateAndSaveStripSet(ctx context.Context, gameMode string) ([5]uuid.UUID, error) // Generates one complete set and returns strip IDs
	GetStripByID(ctx context.Context, id uuid.UUID) (*ReelStrip, error)
	GetActiveStripsCount(ctx context.Context, gameMode string) (map[int]int, error)
	ValidateStripIntegrity(strip *ReelStrip) error

	// Legacy support (deprecated)
	GetRandomReelSet(ctx context.Context, gameMode string) (*ReelStripSet, error)
	RotateStrips(ctx context.Context, gameMode string, newVersion int, count int) error
}
