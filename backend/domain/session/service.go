package session

import (
	"context"

	"github.com/google/uuid"
)

// Service defines the interface for game session business logic
type Service interface {
	// StartSession creates a new game session
	StartSession(ctx context.Context, playerID uuid.UUID, betAmount float64) (*GameSession, error)

	// EndSession ends the current game session
	EndSession(ctx context.Context, sessionID uuid.UUID) (*GameSession, error)

	// GetSession retrieves a session by ID
	GetSession(ctx context.Context, sessionID uuid.UUID) (*GameSession, error)

	// GetPlayerSessions retrieves all sessions for a player
	GetPlayerSessions(ctx context.Context, playerID uuid.UUID, page, limit int) ([]*GameSession, error)
}
