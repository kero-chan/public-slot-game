package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/freespins"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/provablyfair"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/domain/spin"
	"github.com/slotmachine/backend/internal/game/engine"
	freespinsEngine "github.com/slotmachine/backend/internal/game/freespins"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// FreeSpinsService implements the freespins.Service interface
type FreeSpinsService struct {
	sessionRepo   session.Repository
	freespinsRepo freespins.Repository
	spinRepo      spin.Repository
	playerRepo    player.Repository
	gameEngine    *engine.GameEngine
	pfService     *ProvablyFairService // Required: always use HKDF RNG for provably fair
	logger        *logger.Logger
}

// NewFreeSpinsService creates a new free spins service
// pfService is required - all spins use HKDF RNG for provably fair outcomes
func NewFreeSpinsService(
	sessionRepo session.Repository,
	freespinsRepo freespins.Repository,
	spinRepo spin.Repository,
	playerRepo player.Repository,
	gameEngine *engine.GameEngine,
	pfService *ProvablyFairService,
	log *logger.Logger,
) freespins.Service {
	return &FreeSpinsService{
		sessionRepo:   sessionRepo,
		freespinsRepo: freespinsRepo,
		spinRepo:      spinRepo,
		playerRepo:    playerRepo,
		gameEngine:    gameEngine,
		pfService:     pfService,
		logger:        log,
	}
}

// TriggerFreeSpins triggers a new free spins session
func (s *FreeSpinsService) TriggerFreeSpins(
	ctx context.Context,
	playerID, spinID uuid.UUID,
	scatterCount int,
	betAmount float64,
) (*freespins.FreeSpinsSession, error) {
	// Validate scatter count
	if scatterCount < 3 {
		return nil, fmt.Errorf("insufficient scatters to trigger free spins")
	}

	// Check if player already has an active free spins session
	existingSession, _ := s.freespinsRepo.GetActiveByPlayer(ctx, playerID)
	if existingSession != nil {
		return nil, freespins.ErrActiveFreeSpinsExists
	}

	// Calculate free spins awarded
	spinsAwarded := freespinsEngine.CalculateFreeSpinsAward(scatterCount)

	// Create new free spins session
	newSession := &freespins.FreeSpinsSession{
		ID:                uuid.New(),
		PlayerID:          playerID,
		TriggeredBySpinID: &spinID,
		ScatterCount:      scatterCount,
		TotalSpinsAwarded: spinsAwarded,
		SpinsCompleted:    0,
		RemainingSpins:    spinsAwarded,
		LockedBetAmount:   betAmount,
		TotalWon:          0.0,
		IsActive:          true,
		IsCompleted:       false,
		CreatedAt:         time.Now().UTC(),
		CompletedAt:       nil,
	}

	// Save to database
	if err := s.freespinsRepo.Create(ctx, newSession); err != nil {
		s.logger.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to create free spins session")
		return nil, fmt.Errorf("failed to create free spins session: %w", err)
	}

	s.logger.Info().
		Str("free_spins_session_id", newSession.ID.String()).
		Str("player_id", playerID.String()).
		Int("scatter_count", scatterCount).
		Int("spins_awarded", spinsAwarded).
		Msg("Free spins triggered successfully")

	return newSession, nil
}

// ExecuteFreeSpin executes a spin in a free spins session
// clientSeed is optional: for provably fair sessions, client provides their own seed per-spin
func (s *FreeSpinsService) ExecuteFreeSpin(ctx context.Context, freeSpinsSessionID uuid.UUID, clientSeed string) (*spin.SpinResult, error) {
	log := s.logger.WithTraceContext(ctx)

	// Get free spins session
	freeSpinsSession, err := s.freespinsRepo.GetAvailableSessionByID(ctx, freeSpinsSessionID)
	if err != nil {
		log.Error().Err(err).Str("free_spins_session_id", freeSpinsSessionID.String()).Msg("Failed to get free available spins session")
		return nil, freespins.ErrFreeSpinsNotFound
	}

	var p *player.Player
	if ctx.Value("player") != nil {
		p = ctx.Value("player").(*player.Player)
	} else {
		// Get player
		p, err = s.playerRepo.GetByID(ctx, freeSpinsSession.PlayerID)
		if err != nil {
			log.Error().Err(err).Str("player_id", freeSpinsSession.PlayerID.String()).Msg("Failed to get player for spin")
			return nil, player.ErrPlayerNotFound
		}
	}

	// Record balance before
	balanceBefore := p.Balance

	// Create engine session for free spin execution
	engineSession := &freespinsEngine.Session{
		ID:                freeSpinsSession.ID,
		PlayerID:          freeSpinsSession.PlayerID,
		ReelStripConfigID: freeSpinsSession.ReelStripConfigID,
		TotalSpinsAwarded: freeSpinsSession.TotalSpinsAwarded,
		SpinsCompleted:    freeSpinsSession.SpinsCompleted,
		RemainingSpins:    freeSpinsSession.RemainingSpins,
		LockedBetAmount:   freeSpinsSession.LockedBetAmount,
		TotalWon:          freeSpinsSession.TotalWon,
		IsActive:          freeSpinsSession.IsActive,
		CreatedAt:         freeSpinsSession.CreatedAt,
	}

	// Deduct remaining spins
	if err := s.freespinsRepo.ExecuteSpinWithLock(ctx, freeSpinsSession.ID, -1, freeSpinsSession.LockVersion); err != nil {
		log.Error().Err(err).Str("player_id", freeSpinsSession.PlayerID.String()).Msg("Failed to deduct remaining spins")
		return nil, fmt.Errorf("failed to deduct remaining spins: %w", err)
	}
	freeSpinsSession.LockVersion++
	freeSpinsSession.RemainingSpins--
	freeSpinsSession.SpinsCompleted++

	// Verify active provably fair session exists (required for all spins)
	var pfResult *provablyfair.SpinResult
	if _, err := s.pfService.GetSessionState(ctx, freeSpinsSession.SessionID); err != nil {
		log.Error().Err(err).Str("session_id", freeSpinsSession.SessionID.String()).Msg("No active provably fair session")
		return nil, fmt.Errorf("provably fair session required: start a PF session first")
	}

	// Execute free spin using game engine with HKDF RNG (always provably fair)
	spinNumber := freeSpinsSession.SpinsCompleted
	var engineResult *engine.FreeSpinResult

	// Use HKDF-based RNG for provably fair outcomes (RFC 5869 with per-reel key derivation)
	// Note: thetaSeed is empty for free spins - theta verification only happens on first spin
	hkdfRNG, _, err := s.pfService.GetHKDFStreamRNG(ctx, freeSpinsSession.SessionID, clientSeed, "")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get HKDF stream RNG for free spin")
		return nil, fmt.Errorf("failed to get HKDF stream RNG: %w", err)
	}
	engineResult, err = s.gameEngine.ExecuteFreeSpinWithRNG(ctx, freeSpinsSession.PlayerID, engineSession, spinNumber, hkdfRNG)
	if err != nil {
		s.freespinsRepo.RollbackSpin(ctx, freeSpinsSession.ID, 1)
		log.Error().Err(err).Msg("Failed to execute free spin with HKDF RNG")
		return nil, fmt.Errorf("failed to execute free spin: %w", err)
	}

	// Credit win to balance if any
	newBalance := balanceBefore
	if engineResult.TotalWin > 0 {
		newBalance = newBalance + engineResult.TotalWin
		if err := s.playerRepo.UpdateBalance(ctx, freeSpinsSession.PlayerID, engineResult.TotalWin); err != nil {
			log.Error().Err(err).Str("player_id", freeSpinsSession.PlayerID.String()).Msg("Failed to credit free spin win")
			// Continue anyway, spin already executed
		}
	}

	// Update free spins session
	newRemainingSpins := engineResult.RemainingSpins
	newTotalWon := freeSpinsSession.TotalWon + engineResult.TotalWin

	// Handle retrigger
	if engineResult.Retriggered {
		if err := s.freespinsRepo.AddSpins(ctx, freeSpinsSessionID, engineResult.AdditionalSpins); err != nil {
			log.Error().Err(err).Str("free_spins_session_id", freeSpinsSessionID.String()).Msg("Failed to add retrigger spins")
		}

		log.Info().
			Str("free_spins_session_id", freeSpinsSessionID.String()).
			Int("additional_spins", engineResult.AdditionalSpins).
			Msg("Free spins retriggered")
	}

	if err := s.freespinsRepo.AddTotalWon(ctx, freeSpinsSessionID, engineResult.TotalWin); err != nil {
		log.Error().Err(err).Str("free_spins_session_id", freeSpinsSessionID.String()).Msg("Failed to update total won")
	}

	// Check if session is complete
	if newRemainingSpins <= 0 {
		if err := s.freespinsRepo.CompleteSession(ctx, freeSpinsSessionID); err != nil {
			log.Error().Err(err).Str("free_spins_session_id", freeSpinsSessionID.String()).Msg("Failed to complete free spins session")
		}
		log.Info().
			Str("free_spins_session_id", freeSpinsSessionID.String()).
			Float64("total_won", newTotalWon).
			Msg("Free spins session completed")
	}

	// Convert grid and cascades
	grid := convertGrid(engineResult.Grid)
	cascades := convertCascades(engineResult.Cascades)

	// Save spin record
	spinRecord := &spin.Spin{
		ID:                 engineResult.SpinID,
		SessionID:          freeSpinsSession.SessionID, // Free spins don't have a session ID in traditional sense
		PlayerID:           freeSpinsSession.PlayerID,
		BetAmount:          freeSpinsSession.LockedBetAmount,
		BalanceBefore:      balanceBefore,
		BalanceAfter:       newBalance,
		Grid:               grid,
		Cascades:           cascades,
		TotalWin:           engineResult.TotalWin,
		ScatterCount:       engineResult.ScatterCount,
		IsFreeSpin:         true,
		FreeSpinsSessionID: &freeSpinsSessionID,
		FreeSpinsTriggered: false,
		ReelPositions:      engineResult.ReelPositions,
		CreatedAt:          engineResult.Timestamp,
	}

	if err := s.spinRepo.Create(ctx, spinRecord); err != nil {
		log.Error().Err(err).Str("player_id", freeSpinsSession.PlayerID.String()).Msg("Failed to save free spin record")
	}

	// Record spin in provably fair system (always required)
	pfResult, err = s.pfService.RecordSpin(ctx, &provablyfair.RecordSpinInput{
		GameSessionID: freeSpinsSession.SessionID,
		SpinID:        engineResult.SpinID,
		ReelPositions: engineResult.ReelPositions,
		ClientSeed:    clientSeed,
		IsFreeSpin:    true,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to record free spin in provably fair system")
		// Continue anyway, spin already executed
	}

	// Update session statistics
	if err := s.sessionRepo.UpdateStatistics(ctx, freeSpinsSession.SessionID, 1, 0, engineResult.TotalWin); err != nil {
		log.Error().Err(err).Str("session_id", freeSpinsSession.SessionID.String()).Msg("Failed to update session statistics")
		// Don't return error
	}

	if err := s.playerRepo.UpdateStatistics(ctx, freeSpinsSession.PlayerID, 1, 0, engineResult.TotalWin); err != nil {
		log.Error().Err(err).Str("player_id", freeSpinsSession.PlayerID.String()).Msg("Failed to update player statistics")
		// Don't return error
	}

	log.Info().
		Str("spin_id", spinRecord.ID.String()).
		Str("free_spins_session_id", freeSpinsSessionID.String()).
		Int("spin_number", spinNumber).
		Float64("total_win", engineResult.TotalWin).
		Bool("retriggered", engineResult.Retriggered).
		Int("remaining_spins", newRemainingSpins).
		Msg("Free spin executed successfully")

	// Build result
	result := &spin.SpinResult{
		SpinID:                  engineResult.SpinID,
		SessionID:               freeSpinsSession.SessionID,
		BetAmount:               freeSpinsSession.LockedBetAmount,
		BalanceBefore:           balanceBefore,
		BalanceAfterBet:         balanceBefore,
		NewBalance:              newBalance,
		Grid:                    grid,
		Cascades:                cascades,
		SpinTotalWin:            engineResult.TotalWin,
		ScatterCount:            engineResult.ScatterCount,
		IsFreeSpin:              true,
		FreeSpinsTriggered:      false,
		FreeSpinsRetriggered:    engineResult.Retriggered,
		FreeSpinsAdditional:     engineResult.AdditionalSpins,
		FreeSpinsSessionID:      freeSpinsSession.ID.String(),
		FreeSpinsRemainingSpins: newRemainingSpins,
		FreeSessionTotalWin:     newTotalWon,
		Timestamp:               engineResult.Timestamp.Format(time.RFC3339),
	}

	// Add provably fair data if available
	if pfResult != nil {
		result.ProvablyFair = &spin.SpinProvablyFairData{
			SpinIndex:    pfResult.SpinIndex,
			Nonce:        pfResult.Nonce,
			SpinHash:     pfResult.SpinHash,
			PrevSpinHash: pfResult.PrevSpinHash,
		}
	}

	return result, nil
}

// GetStatus retrieves the status of a free spins session
func (s *FreeSpinsService) GetStatus(ctx context.Context, freeSpinsSessionID uuid.UUID) (*freespins.FreeSpinsStatus, error) {
	session, err := s.freespinsRepo.GetByID(ctx, freeSpinsSessionID)
	if err != nil {
		s.logger.Error().Err(err).Str("free_spins_session_id", freeSpinsSessionID.String()).Msg("Failed to get free spins status")
		return nil, freespins.ErrFreeSpinsNotFound
	}

	status := &freespins.FreeSpinsStatus{
		Active:             session.IsActive,
		FreeSpinsSessionID: session.ID,
		TotalSpinsAwarded:  session.TotalSpinsAwarded,
		SpinsCompleted:     session.SpinsCompleted,
		RemainingSpins:     session.RemainingSpins,
		LockedBetAmount:    session.LockedBetAmount,
		TotalWon:           session.TotalWon,
	}

	return status, nil
}

// GetActiveSession retrieves the active free spins session for a player
func (s *FreeSpinsService) GetActiveSession(ctx context.Context, playerID uuid.UUID) (*freespins.FreeSpinsSession, error) {
	session, err := s.freespinsRepo.GetActiveByPlayer(ctx, playerID)
	if err != nil {
		s.logger.Debug().Str("player_id", playerID.String()).Msg("No active free spins session")
		return nil, freespins.ErrFreeSpinsNotFound
	}

	return session, nil
}

// RetriggerFreeSpins adds additional spins to an active session
func (s *FreeSpinsService) RetriggerFreeSpins(ctx context.Context, freeSpinsSessionID uuid.UUID, scatterCount int) error {
	// Validate scatter count
	if scatterCount < 3 {
		return fmt.Errorf("insufficient scatters to retrigger free spins")
	}

	// Get session
	session, err := s.freespinsRepo.GetByID(ctx, freeSpinsSessionID)
	if err != nil {
		return freespins.ErrFreeSpinsNotFound
	}

	// Verify session is active
	if !session.IsActive {
		return freespins.ErrFreeSpinsNotActive
	}

	// Calculate additional spins
	additionalSpins := freespinsEngine.CalculateFreeSpinsAward(scatterCount)

	// Add spins to session
	if err := s.freespinsRepo.AddSpins(ctx, freeSpinsSessionID, additionalSpins); err != nil {
		s.logger.Error().Err(err).Str("free_spins_session_id", freeSpinsSessionID.String()).Msg("Failed to retrigger free spins")
		return fmt.Errorf("failed to retrigger free spins: %w", err)
	}

	s.logger.Info().
		Str("free_spins_session_id", freeSpinsSessionID.String()).
		Int("scatter_count", scatterCount).
		Int("additional_spins", additionalSpins).
		Msg("Free spins retriggered")

	return nil
}
