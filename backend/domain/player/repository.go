package player

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for player data access
type Repository interface {
	// Create creates a new player
	Create(ctx context.Context, player *Player) error

	// GetByID retrieves a player by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Player, error)

	// GetByUsername retrieves a player by username (cross-game account only)
	GetByUsername(ctx context.Context, username string) (*Player, error)

	// GetByUsernameAndGame retrieves a player by username and game_id
	// Pass nil for gameID to find cross-game accounts
	GetByUsernameAndGame(ctx context.Context, username string, gameID *uuid.UUID) (*Player, error)

	// GetByEmail retrieves a player by email (cross-game account only)
	GetByEmail(ctx context.Context, email string) (*Player, error)

	// GetByEmailAndGame retrieves a player by email and game_id
	GetByEmailAndGame(ctx context.Context, email string, gameID *uuid.UUID) (*Player, error)

	// FindLoginCandidate finds a player who can login with given username and game
	// Returns player if exact game match OR cross-game account exists
	// Prefers game-specific account over cross-game account if both exist
	FindLoginCandidate(ctx context.Context, username string, gameID *uuid.UUID) (*Player, error)

	// Update updates a player's information
	Update(ctx context.Context, player *Player) error

	// UpdateBalance updates a player's balance
	UpdateBalance(ctx context.Context, id uuid.UUID, newBalance float64) error
	UpdateBalanceWithLock(ctx context.Context, id uuid.UUID, newBalance float64, lockVersion int) error

	// UpdateBalanceWithTx updates a player's balance within a transaction
	// The tx parameter should be passed via context using TxKey
	UpdateBalanceWithTx(ctx context.Context, id uuid.UUID, amount float64) error
	UpdateBalanceWithLockAndTx(ctx context.Context, id uuid.UUID, amount float64, lockVersion int) error

	// UpdateStatistics updates player statistics
	UpdateStatistics(ctx context.Context, id uuid.UUID, spins int, wagered, won float64) error

	// UpdateLastLogin updates the last login timestamp
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// Delete deletes a player (soft delete or hard delete based on implementation)
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves a list of players with filters and pagination
	List(ctx context.Context, filters ListFilters) ([]*Player, int64, error)
}

// ListFilters represents filters for listing players
type ListFilters struct {
	Username string
	Email    string
	GameID   *uuid.UUID
	IsActive *bool
	Page     int
	Limit    int
	SortBy   string
	SortDesc bool
}
