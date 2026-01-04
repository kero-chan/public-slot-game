package freespins

import (
	"context"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/spin"
)

// Service defines the interface for free spins business logic
type Service interface {
	// TriggerFreeSpins triggers a new free spins session
	TriggerFreeSpins(ctx context.Context, playerID, spinID uuid.UUID, scatterCount int, betAmount float64) (*FreeSpinsSession, error)

	// ExecuteFreeSpin executes a spin in a free spins session
	// clientSeed is optional: for provably fair sessions, client provides their own seed per-spin
	ExecuteFreeSpin(ctx context.Context, freeSpinsSessionID uuid.UUID, clientSeed string) (*spin.SpinResult, error)

	// GetStatus retrieves the status of a free spins session
	GetStatus(ctx context.Context, freeSpinsSessionID uuid.UUID) (*FreeSpinsStatus, error)

	// GetActiveSession retrieves the active free spins session for a player
	GetActiveSession(ctx context.Context, playerID uuid.UUID) (*FreeSpinsSession, error)

	// RetriggerFreeSpins adds additional spins to an active session
	RetriggerFreeSpins(ctx context.Context, freeSpinsSessionID uuid.UUID, scatterCount int) error
}

// FreeSpinsStatus represents the status of a free spins session
type FreeSpinsStatus struct {
	Active             bool      `json:"active"`
	FreeSpinsSessionID uuid.UUID `json:"free_spins_session_id,omitempty"`
	TotalSpinsAwarded  int       `json:"total_spins_awarded"`
	SpinsCompleted     int       `json:"spins_completed"`
	RemainingSpins     int       `json:"remaining_spins"`
	LockedBetAmount    float64   `json:"locked_bet_amount"`
	TotalWon           float64   `json:"total_won"`
}
