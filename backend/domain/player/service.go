package player

import (
	"context"

	"github.com/google/uuid"
)

// LoginResult contains the result of a successful login
type LoginResult struct {
	Player       *Player
	SessionToken string
	ExpiresAt    int64 // Unix timestamp
}

// LoginOptions contains options for login
type LoginOptions struct {
	ForceLogout bool   // If true, force logout existing session
	IPAddress   string // Client IP address
	UserAgent   string // Client user agent
	DeviceInfo  string // Device information
}

// Service defines the interface for player business logic
type Service interface {
	// Register creates a new player account (optionally bound to a game)
	// If gameID is nil, creates a cross-game account
	Register(ctx context.Context, username, email, password string, gameID *uuid.UUID) (*Player, error)

	// Login authenticates a player and creates a session
	// If gameID is provided, validates game access
	// Cross-game accounts (game_id = NULL) can login to any game
	// Returns error if player already logged in on another device (unless forceLogout is true)
	Login(ctx context.Context, username, password string, gameID *uuid.UUID, opts *LoginOptions) (*LoginResult, error)

	// Logout invalidates the player's session
	Logout(ctx context.Context, sessionToken string) error

	// ValidateSession validates a session token and returns session info
	// Also validates that the session's game_id matches the requested game
	ValidateSession(ctx context.Context, sessionToken string, requestedGameID *uuid.UUID) (*LoginResult, error)

	// GetProfile retrieves a player's profile
	GetProfile(ctx context.Context, playerID uuid.UUID) (*Player, error)

	// GetBalance retrieves a player's current balance
	GetBalance(ctx context.Context, playerID uuid.UUID) (float64, error)

	// UpdateBalance updates a player's balance (for internal use)
	UpdateBalance(ctx context.Context, playerID uuid.UUID, newBalance float64) error

	// DeductBet deducts bet amount from player balance
	DeductBet(ctx context.Context, playerID uuid.UUID, betAmount float64) error

	// CreditWin credits win amount to player balance
	CreditWin(ctx context.Context, playerID uuid.UUID, winAmount float64) error
}
