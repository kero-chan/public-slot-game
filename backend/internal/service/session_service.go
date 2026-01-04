package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// SessionService implements the session.Service interface
type SessionService struct {
	sessionRepo session.Repository
	playerRepo  player.Repository
	logger      *logger.Logger
}

// NewSessionService creates a new session service
func NewSessionService(
	sessionRepo session.Repository,
	playerRepo player.Repository,
	log *logger.Logger,
) session.Service {
	return &SessionService{
		sessionRepo: sessionRepo,
		playerRepo:  playerRepo,
		logger:      log,
	}
}

// StartSession creates a new game session
func (s *SessionService) StartSession(ctx context.Context, playerID uuid.UUID, betAmount float64) (*session.GameSession, error) {
	log := s.logger.WithTraceContext(ctx)

	// Validate bet amount
	if betAmount <= 0 {
		return nil, fmt.Errorf("bet amount must be positive")
	}

	// Check if player exists and get current balance
	p, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player for session start")
		return nil, player.ErrPlayerNotFound
	}

	// Check if player is active
	if !p.IsActive {
		return nil, fmt.Errorf("player account is not active")
	}

	// Check if there's already an active session
	existingSession, _ := s.sessionRepo.GetActiveSessionByPlayer(ctx, playerID)
	if existingSession != nil {
		log.Warn().
			Str("player_id", playerID.String()).
			Str("existing_session_id", existingSession.ID.String()).
			Msg("Player already has an active session")
		return nil, session.ErrActiveSessionExists
	}

	// Create new session
	newSession := &session.GameSession{
		ID:              uuid.New(),
		PlayerID:        playerID,
		BetAmount:       betAmount,
		StartingBalance: p.Balance,
		EndingBalance:   nil,
		TotalSpins:      0,
		TotalWagered:    0.0,
		TotalWon:        0.0,
		NetChange:       0.0,
		CreatedAt:       time.Now().UTC(),
		EndedAt:         nil,
	}

	// Save to database
	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to create session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	log.Info().
		Str("session_id", newSession.ID.String()).
		Str("player_id", playerID.String()).
		Float64("bet_amount", betAmount).
		Msg("Session started successfully")

	return newSession, nil
}

// EndSession ends the current game session
func (s *SessionService) EndSession(ctx context.Context, sessionID uuid.UUID) (*session.GameSession, error) {
	log := s.logger.WithTraceContext(ctx)

	// Get session
	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to get session for ending")
		return nil, session.ErrSessionNotFound
	}

	// Check if session is already ended
	if sess.EndedAt != nil {
		return nil, session.ErrSessionAlreadyEnded
	}

	// Get current player balance
	p, err := s.playerRepo.GetByID(ctx, sess.PlayerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", sess.PlayerID.String()).Msg("Failed to get player balance for session end")
		return nil, player.ErrPlayerNotFound
	}

	// End session with current balance
	if err := s.sessionRepo.EndSession(ctx, sessionID, p.Balance); err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to end session")
		return nil, fmt.Errorf("failed to end session: %w", err)
	}

	// Get updated session
	updatedSession, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to get updated session")
		return nil, session.ErrSessionNotFound
	}

	log.Info().
		Str("session_id", sessionID.String()).
		Str("player_id", sess.PlayerID.String()).
		Float64("starting_balance", sess.StartingBalance).
		Float64("ending_balance", p.Balance).
		Float64("net_change", updatedSession.NetChange).
		Int("total_spins", updatedSession.TotalSpins).
		Msg("Session ended successfully")

	return updatedSession, nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(ctx context.Context, sessionID uuid.UUID) (*session.GameSession, error) {
	log := s.logger.WithTraceContext(ctx)

	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to get session")
		return nil, session.ErrSessionNotFound
	}

	return sess, nil
}

// GetPlayerSessions retrieves all sessions for a player
func (s *SessionService) GetPlayerSessions(ctx context.Context, playerID uuid.UUID, page, limit int) ([]*session.GameSession, error) {
	log := s.logger.WithTraceContext(ctx)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // Default limit
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get sessions from repository
	sessions, err := s.sessionRepo.GetByPlayer(ctx, playerID, limit, offset)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player sessions")
		return nil, fmt.Errorf("failed to get player sessions: %w", err)
	}

	return sessions, nil
}
