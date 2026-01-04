package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/trial"
	"github.com/slotmachine/backend/internal/infra/cache"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// TrialService manages trial session lifecycle
type TrialService struct {
	cache  *cache.RedisClient
	logger *logger.Logger
}

// NewTrialService creates a new trial service
func NewTrialService(cache *cache.RedisClient, logger *logger.Logger) *TrialService {
	return &TrialService{
		cache:  cache,
		logger: logger,
	}
}

// TrialSessionResult represents the result of starting a trial session
type TrialSessionResult struct {
	Session   *trial.TrialSession
	ExpiresAt int64 // Unix timestamp
}

// generateTrialToken generates a unique trial session token with prefix
func generateTrialToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate trial token: %w", err)
	}
	return trial.TrialTokenPrefix + hex.EncodeToString(bytes), nil
}

// StartTrialSession creates a new trial session
func (s *TrialService) StartTrialSession(ctx context.Context, gameID *uuid.UUID) (*TrialSessionResult, error) {
	log := s.logger.WithTraceContext(ctx)

	// Check if Redis is available
	if s.cache == nil {
		return nil, fmt.Errorf("trial mode requires Redis")
	}

	// Generate trial token
	sessionToken, err := generateTrialToken()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate trial token")
		return nil, err
	}

	// Create trial session
	session := trial.NewTrialSession(gameID, sessionToken)

	// Convert to cache data format
	cacheData := &cache.TrialSessionData{
		ID:             session.ID.String(),
		Balance:        session.Balance,
		TotalSpins:     session.TotalSpins,
		TotalWagered:   session.TotalWagered,
		TotalWon:       session.TotalWon,
		CreatedAt:      session.CreatedAt.Unix(),
		LastActivityAt: session.LastActivityAt.Unix(),
		ExpiresAt:      session.ExpiresAt.Unix(),
	}
	if gameID != nil {
		cacheData.GameID = gameID.String()
	}

	// Store in Redis with TTL
	if err := s.cache.SetTrialSession(ctx, sessionToken, cacheData, trial.TrialSessionDuration); err != nil {
		log.Error().Err(err).Msg("Failed to store trial session in Redis")
		return nil, fmt.Errorf("failed to create trial session: %w", err)
	}

	log.Info().
		Str("trial_session_id", session.ID.String()).
		Float64("balance", session.Balance).
		Msg("Trial session created")

	return &TrialSessionResult{
		Session:   session,
		ExpiresAt: session.ExpiresAt.Unix(),
	}, nil
}

// ValidateTrialSession validates a trial session token and returns the session
func (s *TrialService) ValidateTrialSession(ctx context.Context, sessionToken string, requestedGameID *uuid.UUID) (*trial.TrialSession, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("trial mode requires Redis")
	}

	// Get session from Redis
	cacheData, err := s.cache.GetTrialSession(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get trial session: %w", err)
	}
	if cacheData == nil {
		return nil, fmt.Errorf("trial session not found or expired")
	}

	// Check expiration
	if time.Now().Unix() > cacheData.ExpiresAt {
		// Clean up expired session
		_ = s.cache.DeleteTrialSession(ctx, sessionToken)
		return nil, fmt.Errorf("trial session expired")
	}

	// Parse session ID
	sessionID, err := uuid.Parse(cacheData.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid trial session ID")
	}

	// Parse game ID if present
	var gameID *uuid.UUID
	if cacheData.GameID != "" {
		parsed, err := uuid.Parse(cacheData.GameID)
		if err == nil {
			gameID = &parsed
		}
	}

	// Verify game access if game ID is specified
	if gameID != nil && requestedGameID != nil && *gameID != *requestedGameID {
		return nil, fmt.Errorf("trial session not authorized for this game")
	}

	// Reconstruct session object
	session := &trial.TrialSession{
		ID:             sessionID,
		SessionToken:   sessionToken,
		Balance:        cacheData.Balance,
		GameID:         gameID,
		TotalSpins:     cacheData.TotalSpins,
		TotalWagered:   cacheData.TotalWagered,
		TotalWon:       cacheData.TotalWon,
		CreatedAt:      time.Unix(cacheData.CreatedAt, 0),
		LastActivityAt: time.Unix(cacheData.LastActivityAt, 0),
		ExpiresAt:      time.Unix(cacheData.ExpiresAt, 0),
	}

	return session, nil
}

// GetTrialBalance returns the current trial balance
func (s *TrialService) GetTrialBalance(ctx context.Context, sessionToken string) (float64, error) {
	session, err := s.ValidateTrialSession(ctx, sessionToken, nil)
	if err != nil {
		return 0, err
	}
	return session.Balance, nil
}

// UpdateTrialBalance updates the trial balance in Redis
func (s *TrialService) UpdateTrialBalance(ctx context.Context, sessionToken string, newBalance float64) error {
	if s.cache == nil {
		return fmt.Errorf("trial mode requires Redis")
	}

	cacheData, err := s.cache.GetTrialSession(ctx, sessionToken)
	if err != nil {
		return err
	}
	if cacheData == nil {
		return fmt.Errorf("trial session not found")
	}

	cacheData.Balance = newBalance
	cacheData.LastActivityAt = time.Now().Unix()

	return s.cache.UpdateTrialSession(ctx, sessionToken, cacheData)
}

// UpdateTrialStats updates trial session statistics
func (s *TrialService) UpdateTrialStats(ctx context.Context, sessionToken string, spins int, wagered, won float64) error {
	if s.cache == nil {
		return fmt.Errorf("trial mode requires Redis")
	}

	cacheData, err := s.cache.GetTrialSession(ctx, sessionToken)
	if err != nil {
		return err
	}
	if cacheData == nil {
		return fmt.Errorf("trial session not found")
	}

	cacheData.TotalSpins += spins
	cacheData.TotalWagered += wagered
	cacheData.TotalWon += won
	cacheData.LastActivityAt = time.Now().Unix()

	return s.cache.UpdateTrialSession(ctx, sessionToken, cacheData)
}

// DeductTrialBalance deducts bet amount from trial balance
func (s *TrialService) DeductTrialBalance(ctx context.Context, sessionToken string, amount float64) (float64, error) {
	if s.cache == nil {
		return 0, fmt.Errorf("trial mode requires Redis")
	}

	cacheData, err := s.cache.GetTrialSession(ctx, sessionToken)
	if err != nil {
		return 0, err
	}
	if cacheData == nil {
		return 0, fmt.Errorf("trial session not found")
	}

	if cacheData.Balance < amount {
		return cacheData.Balance, fmt.Errorf("insufficient trial balance")
	}

	cacheData.Balance -= amount
	cacheData.TotalWagered += amount
	cacheData.LastActivityAt = time.Now().Unix()

	if err := s.cache.UpdateTrialSession(ctx, sessionToken, cacheData); err != nil {
		return 0, err
	}

	return cacheData.Balance, nil
}

// CreditTrialWin adds win amount to trial balance (win already includes trial multiplier)
func (s *TrialService) CreditTrialWin(ctx context.Context, sessionToken string, amount float64) (float64, error) {
	if s.cache == nil {
		return 0, fmt.Errorf("trial mode requires Redis")
	}

	cacheData, err := s.cache.GetTrialSession(ctx, sessionToken)
	if err != nil {
		return 0, err
	}
	if cacheData == nil {
		return 0, fmt.Errorf("trial session not found")
	}

	cacheData.Balance += amount
	cacheData.TotalWon += amount
	cacheData.TotalSpins++
	cacheData.LastActivityAt = time.Now().Unix()

	if err := s.cache.UpdateTrialSession(ctx, sessionToken, cacheData); err != nil {
		return 0, err
	}

	return cacheData.Balance, nil
}

// EndTrialSession ends a trial session (optional, sessions auto-expire)
func (s *TrialService) EndTrialSession(ctx context.Context, sessionToken string) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.DeleteTrialSession(ctx, sessionToken)
}

// Trial Game Session Methods

// StartTrialGameSession creates a new trial game session
func (s *TrialService) StartTrialGameSession(ctx context.Context, trialSession *trial.TrialSession, betAmount float64) (*trial.TrialGameSession, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("trial mode requires Redis")
	}

	gameSession := trial.NewTrialGameSession(trialSession.ID, betAmount, trialSession.Balance)

	cacheData := &cache.TrialGameSessionData{
		ID:              gameSession.ID.String(),
		TrialSessionID:  trialSession.ID.String(),
		BetAmount:       gameSession.BetAmount,
		StartingBalance: gameSession.StartingBalance,
		TotalSpins:      0,
		TotalWagered:    0,
		TotalWon:        0,
		CreatedAt:       gameSession.CreatedAt.Unix(),
	}

	if err := s.cache.SetTrialGameSession(ctx, gameSession.ID.String(), cacheData, trial.TrialSessionDuration); err != nil {
		return nil, fmt.Errorf("failed to create trial game session: %w", err)
	}

	return gameSession, nil
}

// GetTrialGameSession retrieves a trial game session
func (s *TrialService) GetTrialGameSession(ctx context.Context, sessionID uuid.UUID) (*trial.TrialGameSession, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("trial mode requires Redis")
	}

	cacheData, err := s.cache.GetTrialGameSession(ctx, sessionID.String())
	if err != nil {
		return nil, err
	}
	if cacheData == nil {
		return nil, fmt.Errorf("trial game session not found")
	}

	trialSessionID, _ := uuid.Parse(cacheData.TrialSessionID)

	return &trial.TrialGameSession{
		ID:              sessionID,
		TrialSessionID:  trialSessionID,
		BetAmount:       cacheData.BetAmount,
		StartingBalance: cacheData.StartingBalance,
		TotalSpins:      cacheData.TotalSpins,
		TotalWagered:    cacheData.TotalWagered,
		TotalWon:        cacheData.TotalWon,
		CreatedAt:       time.Unix(cacheData.CreatedAt, 0),
	}, nil
}

// Trial Free Spins Methods

// StartTrialFreeSpins creates a trial free spins session
func (s *TrialService) StartTrialFreeSpins(ctx context.Context, trialSessionID, gameSessionID uuid.UUID, spinsAwarded int, betAmount float64) (*trial.TrialFreeSpinsSession, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("trial mode requires Redis")
	}

	freeSpins := trial.NewTrialFreeSpinsSession(trialSessionID, gameSessionID, spinsAwarded, betAmount)

	cacheData := &cache.TrialFreeSpinsData{
		ID:              freeSpins.ID.String(),
		TrialSessionID:  trialSessionID.String(),
		GameSessionID:   gameSessionID.String(),
		TotalSpins:      freeSpins.TotalSpins,
		RemainingSpins:  freeSpins.RemainingSpins,
		CompletedSpins:  freeSpins.CompletedSpins,
		LockedBetAmount: freeSpins.LockedBetAmount,
		TotalWon:        freeSpins.TotalWon,
		IsActive:        freeSpins.IsActive,
		CreatedAt:       freeSpins.CreatedAt.Unix(),
	}

	if err := s.cache.SetTrialFreeSpins(ctx, freeSpins.ID.String(), cacheData, trial.TrialSessionDuration); err != nil {
		return nil, fmt.Errorf("failed to create trial free spins: %w", err)
	}

	s.logger.Info().
		Str("trial_session_id", trialSessionID.String()).
		Int("spins_awarded", spinsAwarded).
		Msg("Trial free spins session created")

	return freeSpins, nil
}

// GetTrialFreeSpins retrieves a trial free spins session
func (s *TrialService) GetTrialFreeSpins(ctx context.Context, sessionID uuid.UUID) (*trial.TrialFreeSpinsSession, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("trial mode requires Redis")
	}

	cacheData, err := s.cache.GetTrialFreeSpins(ctx, sessionID.String())
	if err != nil {
		return nil, err
	}
	if cacheData == nil {
		return nil, fmt.Errorf("trial free spins session not found")
	}

	trialSessionID, _ := uuid.Parse(cacheData.TrialSessionID)
	gameSessionID, _ := uuid.Parse(cacheData.GameSessionID)

	return &trial.TrialFreeSpinsSession{
		ID:              sessionID,
		TrialSessionID:  trialSessionID,
		GameSessionID:   gameSessionID,
		TotalSpins:      cacheData.TotalSpins,
		RemainingSpins:  cacheData.RemainingSpins,
		CompletedSpins:  cacheData.CompletedSpins,
		LockedBetAmount: cacheData.LockedBetAmount,
		TotalWon:        cacheData.TotalWon,
		IsActive:        cacheData.IsActive,
		CreatedAt:       time.Unix(cacheData.CreatedAt, 0),
	}, nil
}

// GetActiveTrialFreeSpins finds active free spins for a trial session
func (s *TrialService) GetActiveTrialFreeSpins(ctx context.Context, trialSessionID uuid.UUID) (*trial.TrialFreeSpinsSession, error) {
	if s.cache == nil {
		return nil, nil
	}

	cacheData, err := s.cache.GetActiveTrialFreeSpinsByTrialSession(ctx, trialSessionID.String())
	if err != nil {
		return nil, err
	}
	if cacheData == nil {
		return nil, nil
	}

	id, _ := uuid.Parse(cacheData.ID)
	gameSessionID, _ := uuid.Parse(cacheData.GameSessionID)

	return &trial.TrialFreeSpinsSession{
		ID:              id,
		TrialSessionID:  trialSessionID,
		GameSessionID:   gameSessionID,
		TotalSpins:      cacheData.TotalSpins,
		RemainingSpins:  cacheData.RemainingSpins,
		CompletedSpins:  cacheData.CompletedSpins,
		LockedBetAmount: cacheData.LockedBetAmount,
		TotalWon:        cacheData.TotalWon,
		IsActive:        cacheData.IsActive,
		CreatedAt:       time.Unix(cacheData.CreatedAt, 0),
	}, nil
}

// UpdateTrialFreeSpins updates a trial free spins session
func (s *TrialService) UpdateTrialFreeSpins(ctx context.Context, freeSpins *trial.TrialFreeSpinsSession) error {
	if s.cache == nil {
		return fmt.Errorf("trial mode requires Redis")
	}

	cacheData := &cache.TrialFreeSpinsData{
		ID:              freeSpins.ID.String(),
		TrialSessionID:  freeSpins.TrialSessionID.String(),
		GameSessionID:   freeSpins.GameSessionID.String(),
		TotalSpins:      freeSpins.TotalSpins,
		RemainingSpins:  freeSpins.RemainingSpins,
		CompletedSpins:  freeSpins.CompletedSpins,
		LockedBetAmount: freeSpins.LockedBetAmount,
		TotalWon:        freeSpins.TotalWon,
		IsActive:        freeSpins.IsActive,
		CreatedAt:       freeSpins.CreatedAt.Unix(),
	}

	return s.cache.UpdateTrialFreeSpins(ctx, freeSpins.ID.String(), cacheData)
}

// IsTrialToken checks if a token is a trial session token
func IsTrialToken(token string) bool {
	return len(token) > len(trial.TrialTokenPrefix) && token[:len(trial.TrialTokenPrefix)] == trial.TrialTokenPrefix
}
