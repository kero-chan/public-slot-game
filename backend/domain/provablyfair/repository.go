package provablyfair

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for provably fair data persistence
type Repository interface {
	// PFSession operations
	CreateSession(ctx context.Context, session *PFSession) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*PFSession, error)
	GetActiveSessionByPlayer(ctx context.Context, playerID uuid.UUID) (*PFSession, error)
	GetActiveSessionByGameSession(ctx context.Context, gameSessionID uuid.UUID) (*PFSession, error)
	UpdateSession(ctx context.Context, session *PFSession) error
	EndSession(ctx context.Context, id uuid.UUID) error

	// SpinLog operations (append-only)
	CreateSpinLog(ctx context.Context, log *SpinLog) error
	GetSpinLogsBySession(ctx context.Context, pfSessionID uuid.UUID) ([]SpinLog, error)
	GetSpinLogByIndex(ctx context.Context, pfSessionID uuid.UUID, spinIndex int64) (*SpinLog, error)
	GetLastSpinLog(ctx context.Context, pfSessionID uuid.UUID) (*SpinLog, error)

	// SessionAudit operations
	CreateSessionAudit(ctx context.Context, audit *SessionAudit) error
	GetSessionAudit(ctx context.Context, pfSessionID uuid.UUID) (*SessionAudit, error)
}

// CacheRepository defines the interface for Redis-based session state
type CacheRepository interface {
	// Session state operations
	SetSessionState(ctx context.Context, state *PFSessionState) error
	GetSessionState(ctx context.Context, sessionID uuid.UUID) (*PFSessionState, error)
	GetSessionStateByPlayer(ctx context.Context, playerID uuid.UUID) (*PFSessionState, error)
	GetSessionStateByGameSession(ctx context.Context, gameSessionID uuid.UUID) (*PFSessionState, error)
	UpdateSessionState(ctx context.Context, state *PFSessionState) error
	DeleteSessionState(ctx context.Context, sessionID uuid.UUID) error

	// Atomic nonce increment
	IncrementNonce(ctx context.Context, sessionID uuid.UUID) (int64, error)
}
