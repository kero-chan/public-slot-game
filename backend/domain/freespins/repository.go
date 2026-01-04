package freespins

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for free spins data access
type Repository interface {
	// Create creates a new free spins session
	Create(ctx context.Context, session *FreeSpinsSession) error

	// GetByID retrieves a free spins session by ID
	GetAvailableSessionByID(ctx context.Context, id uuid.UUID) (*FreeSpinsSession, error)
	GetByID(ctx context.Context, id uuid.UUID) (*FreeSpinsSession, error)

	// GetActiveByPlayer retrieves the active free spins session for a player
	GetActiveByPlayer(ctx context.Context, playerID uuid.UUID) (*FreeSpinsSession, error)

	// Update updates a free spins session
	Update(ctx context.Context, session *FreeSpinsSession) error

	// UpdateSpins updates spins completed and remaining
	RollbackSpin(ctx context.Context, id uuid.UUID, additionalSpins int) error
	ExecuteSpinWithLock(ctx context.Context, id uuid.UUID, additionalSpins int, lockVersion int) error
	UpdateSpins(ctx context.Context, id uuid.UUID, spinsCompleted, remainingSpins int) error

	// UpdateTotalWon updates total won amount
	AddTotalWon(ctx context.Context, id uuid.UUID, totalWon float64) error

	// CompleteSession marks a free spins session as completed
	CompleteSession(ctx context.Context, id uuid.UUID) error

	// AddSpins adds additional spins (for retrigger)
	AddSpins(ctx context.Context, id uuid.UUID, additionalSpins int) error

	// GetByPlayer retrieves all free spins sessions for a player
	GetByPlayer(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*FreeSpinsSession, error)
}
