package reelstrip

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for reel strip data access
type Repository interface {
	// ReelStrip operations
	Create(ctx context.Context, strip *ReelStrip) error
	CreateBatch(ctx context.Context, strips []*ReelStrip) error
	GetByID(ctx context.Context, id uuid.UUID) (*ReelStrip, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*ReelStrip, error)
	Update(ctx context.Context, strip *ReelStrip) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Legacy methods (deprecated - use config-based methods instead)
	GetAllActive(ctx context.Context, gameMode string) ([]*ReelStrip, error)
	GetByGameModeAndReel(ctx context.Context, gameMode string, reelNumber int) ([]*ReelStrip, error)
	CountActive(ctx context.Context, gameMode string) (map[int]int, error)
	DeactivateOldVersions(ctx context.Context, gameMode string, keepVersion int) error

	// ReelStripConfig operations
	CreateConfig(ctx context.Context, config *ReelStripConfig) error
	GetConfigByID(ctx context.Context, id uuid.UUID) (*ReelStripConfig, error)
	GetConfigByName(ctx context.Context, name string) (*ReelStripConfig, error)
	GetDefaultConfig(ctx context.Context, gameMode string) (*ReelStripConfig, error)
	ListConfigs(ctx context.Context, filters *ConfigListFilters) ([]*ReelStripConfig, int64, error)
	UpdateConfig(ctx context.Context, config *ReelStripConfig) error
	DeleteConfig(ctx context.Context, id uuid.UUID) error
	SetDefaultConfig(ctx context.Context, id uuid.UUID, gameMode string) error

	// Get complete reel strip set by config ID
	GetSetByConfigID(ctx context.Context, configID uuid.UUID) (*ReelStripConfigSet, error)

	// PlayerReelStripAssignment operations
	CreateAssignment(ctx context.Context, assignment *PlayerReelStripAssignment) error
	GetPlayerAssignment(ctx context.Context, playerID uuid.UUID) (*PlayerReelStripAssignment, error)
	GetPlayerAssignmentsByPlayerIDs(ctx context.Context, playerIDs []uuid.UUID) (map[uuid.UUID]*PlayerReelStripAssignment, error)
	UpdateAssignment(ctx context.Context, assignment *PlayerReelStripAssignment) error
	DeleteAssignment(ctx context.Context, id uuid.UUID) error
}
