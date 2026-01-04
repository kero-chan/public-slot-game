package session

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for game session data access
type Repository interface {
	// Create creates a new game session
	Create(ctx context.Context, session *GameSession) error

	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, id uuid.UUID) (*GameSession, error)

	// GetActiveSessionByPlayer retrieves the active session for a player
	GetActiveSessionByPlayer(ctx context.Context, playerID uuid.UUID) (*GameSession, error)

	// Update updates a session
	Update(ctx context.Context, session *GameSession) error

	// EndSession marks a session as ended
	EndSession(ctx context.Context, id uuid.UUID, endingBalance float64) error

	// UpdateStatistics updates session statistics
	UpdateStatistics(ctx context.Context, id uuid.UUID, spins int, wagered, won float64) error

	// GetByPlayer retrieves all sessions for a player (paginated)
	GetByPlayer(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*GameSession, error)
}

// PlayerSessionRepository defines the interface for player login session data access
type PlayerSessionRepository interface {
	// Create creates a new player session
	Create(ctx context.Context, session *PlayerSession) error

	// GetByToken retrieves a session by session token
	GetByToken(ctx context.Context, token string) (*PlayerSession, error)

	// GetActiveByPlayerAndGame retrieves active session for a player and game
	GetActiveByPlayerAndGame(ctx context.Context, playerID uuid.UUID, gameID *uuid.UUID) (*PlayerSession, error)

	// DeactivateSession marks a session as inactive with logout reason
	DeactivateSession(ctx context.Context, sessionID uuid.UUID, reason string) error

	// DeactivateAllPlayerSessions deactivates all active sessions for a player
	DeactivateAllPlayerSessions(ctx context.Context, playerID uuid.UUID, reason string) error

	// DeactivatePlayerGameSession deactivates active session for a player in specific game
	DeactivatePlayerGameSession(ctx context.Context, playerID uuid.UUID, gameID *uuid.UUID, reason string) error

	// UpdateLastActivity updates the last activity timestamp
	UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error

	// CleanupExpiredSessions marks expired sessions as inactive
	CleanupExpiredSessions(ctx context.Context) (int64, error)
}
