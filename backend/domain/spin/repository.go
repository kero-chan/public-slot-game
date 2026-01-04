package spin

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for spin data access
type Repository interface {
	// Create creates a new spin record
	Create(ctx context.Context, spin *Spin) error

	// GetByID retrieves a spin by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Spin, error)

	// GetBySession retrieves all spins for a session
	GetBySession(ctx context.Context, sessionID uuid.UUID) ([]*Spin, error)

	// GetByPlayer retrieves spins for a player (paginated)
	GetByPlayer(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*Spin, error)

	// GetByPlayerInTimeRange retrieves spins for a player in a time range
	GetByPlayerInTimeRange(ctx context.Context, playerID uuid.UUID, start, end time.Time, limit, offset int) ([]*Spin, error)

	// GetByFreeSpinsSession retrieves all spins for a free spins session
	GetByFreeSpinsSession(ctx context.Context, freeSpinsSessionID uuid.UUID) ([]*Spin, error)

	// Count counts total spins for a player
	Count(ctx context.Context, playerID uuid.UUID) (int64, error)

	// CountInTimeRange counts spins for a player in a time range
	CountInTimeRange(ctx context.Context, playerID uuid.UUID, start, end time.Time) (int64, error)

	// UpdateFreeSpinsSessionId updates the free spins session ID for a spin
	UpdateFreeSpinsSessionId(ctx context.Context, id uuid.UUID, freeSpinsSessionID uuid.UUID) error
}
