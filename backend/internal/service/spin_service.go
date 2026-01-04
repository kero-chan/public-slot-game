package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/freespins"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/provablyfair"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/domain/spin"
	"github.com/slotmachine/backend/internal/game/engine"
	"github.com/slotmachine/backend/internal/infra/repository"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// SpinService implements the spin.Service interface
type SpinService struct {
	spinRepo      spin.Repository
	playerRepo    player.Repository
	sessionRepo   session.Repository
	gameEngine    *engine.GameEngine
	freespinsRepo freespins.Repository
	reelstripRepo reelstrip.Repository
	txManager     *repository.TxManager
	pfService     *ProvablyFairService // Required: always use HKDF RNG for provably fair
	trialService  *TrialService        // Optional: nil if trials are disabled
	logger        *logger.Logger
}

// NewSpinService creates a new spin service
// pfService is required - all spins use HKDF RNG for provably fair outcomes
func NewSpinService(
	spinRepo spin.Repository,
	playerRepo player.Repository,
	sessionRepo session.Repository,
	gameEngine *engine.GameEngine,
	freespinsRepo freespins.Repository,
	reelstripRepo reelstrip.Repository,
	txManager *repository.TxManager,
	pfService *ProvablyFairService,
	log *logger.Logger,
) spin.Service {
	return &SpinService{
		spinRepo:      spinRepo,
		playerRepo:    playerRepo,
		sessionRepo:   sessionRepo,
		gameEngine:    gameEngine,
		freespinsRepo: freespinsRepo,
		reelstripRepo: reelstripRepo,
		txManager:     txManager,
		pfService:     pfService,
		logger:        log,
	}
}

// SetTrialService sets the trial service for trial mode support
func (s *SpinService) SetTrialService(trialService *TrialService) {
	s.trialService = trialService
}

// GenerateInitialGrid generates a demo grid for initial display
// This ensures frontend has zero RNG - all symbol generation is backend-controlled
func (s *SpinService) GenerateInitialGrid(ctx context.Context) (spin.Grid, error) {
	// Generate initial grid using game engine (now uses DB-backed reel strips)
	grid, err := s.gameEngine.GenerateInitialGrid(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate initial grid")
		return nil, fmt.Errorf("failed to generate initial grid: %w", err)
	}

	s.logger.Debug().Msg("Generated initial grid for frontend display")

	return convertGrid(grid), nil
}

type GameModeCost struct {
	BuyCost   float64
	BetAmount float64
}

// Game mode costs
var gameModeCosts = map[string]GameModeCost{
	"bonus_spin_trigger": {
		BuyCost:   750,
		BetAmount: 20,
	},
}

// ExecuteSpin executes a regular spin
// gameMode is optional: bonus_spin_trigger (guaranteed free spins)
// clientSeed is optional: for provably fair sessions, client provides their own seed per-spin
// thetaSeed is optional: for Dual Commitment Protocol, revealed on first spin
func (s *SpinService) ExecuteSpin(ctx context.Context, playerID, sessionID uuid.UUID, betAmount float64, gameMode, clientSeed, thetaSeed string) (*spin.SpinResult, error) {
	log := s.logger.WithTraceContext(ctx)

	// Validate bet amount
	if betAmount <= 0 {
		return nil, fmt.Errorf("bet amount must be positive")
	}

	// Calculate deduction based on game mode
	// If game mode is specified, only deduct the game mode cost (not bet amount)
	// If no game mode, deduct the bet amount as normal
	var totalDeduction float64
	if gameMode != "" {
		cost, ok := gameModeCosts[gameMode]
		if !ok {
			return nil, fmt.Errorf("invalid game mode: %s", gameMode)
		}
		totalDeduction = cost.BuyCost // Only game mode cost, no bet amount
		betAmount = cost.BetAmount    // Deduct bet amount from total
	} else {
		totalDeduction = betAmount // Normal spin: deduct bet amount
	}

	var p *player.Player
	var err error

	if ctx.Value("player") != nil {
		p = ctx.Value("player").(*player.Player)
	} else {
		// Get player to check balance
		p, err = s.playerRepo.GetByID(ctx, playerID)
		if err != nil {
			log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player for spin")
			return nil, player.ErrPlayerNotFound
		}
	}

	// Get session to verify it exists and is active
	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to get session for spin")
		return nil, session.ErrSessionNotFound
	}

	// Verify session belongs to player
	if sess.PlayerID != playerID {
		return nil, fmt.Errorf("session does not belong to player")
	}

	// Verify session is active
	if sess.EndedAt != nil {
		return nil, session.ErrSessionAlreadyEnded
	}

	// Check if player has sufficient balance
	if p.Balance < totalDeduction {
		log.Warn().
			Str("player_id", playerID.String()).
			Float64("balance", p.Balance).
			Float64("bet_amount", betAmount).
			Str("game_mode", gameMode).
			Float64("total_deduction", totalDeduction).
			Msg("Insufficient balance for spin")
		return nil, player.ErrInsufficientBalance
	}

	// Record balance before
	balanceBefore := p.Balance
	balanceAfterBet := balanceBefore - totalDeduction
	newBalance := balanceAfterBet
	lockVersion := p.LockVersion

	// Verify active provably fair session exists (required for all spins)
	var pfResult *provablyfair.SpinResult
	if _, err := s.pfService.GetSessionState(ctx, sessionID); err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("No active provably fair session")
		return nil, fmt.Errorf("provably fair session required: start a PF session first")
	}

	// Execute spin within a transaction to ensure atomicity
	var engineResult *engine.SpinResult
	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Deduct bet + game mode cost from balance with optimistic lock
		if err := s.playerRepo.UpdateBalanceWithLockAndTx(txCtx, playerID, -totalDeduction, lockVersion); err != nil {
			log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to deduct bet")
			return fmt.Errorf("failed to deduct bet: %w", err)
		}

		// Get HKDF stream RNG from PF service (client provides seed per-spin)
		// This implements RFC 5869 HKDF for per-reel key derivation
		// Dual Commitment Protocol: thetaSeed is verified BEFORE RNG generation
		hkdfRNG, _, err := s.pfService.GetHKDFStreamRNG(txCtx, sessionID, clientSeed, thetaSeed)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get HKDF stream RNG")
			return fmt.Errorf("failed to get HKDF stream RNG: %w", err)
		}

		// Execute spin with HKDF-based RNG (provably fair with per-reel key derivation)
		engineResult, err = s.gameEngine.ExecuteBaseSpinWithRNG(txCtx, playerID, betAmount, gameMode, hkdfRNG)
		if err != nil {
			log.Error().Err(err).Msg("Failed to execute base spin with HKDF RNG")
			return fmt.Errorf("failed to execute spin: %w", err)
		}

		// Credit win to balance if any
		if engineResult.TotalWin > 0 {
			newBalance = newBalance + engineResult.TotalWin
			if err := s.playerRepo.UpdateBalanceWithTx(txCtx, playerID, engineResult.TotalWin); err != nil {
				log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to credit win")
				return fmt.Errorf("failed to credit win: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Prepare game mode fields (nil for normal spins)
	var gameModePtr *string
	var gameModeCostPtr *float64
	if gameMode != "" {
		gameModePtr = &gameMode
		gameModeCostPtr = &totalDeduction
	}

	// Convert engine result to domain spin model
	spinRecord := &spin.Spin{
		ID:                 engineResult.SpinID,
		SessionID:          sessionID,
		PlayerID:           playerID,
		BetAmount:          betAmount,
		BalanceBefore:      balanceBefore,
		BalanceAfter:       newBalance,
		Grid:               convertGrid(engineResult.Grid),
		Cascades:           convertCascades(engineResult.Cascades),
		TotalWin:           engineResult.TotalWin,
		ScatterCount:       engineResult.ScatterCount,
		IsFreeSpin:         false,
		FreeSpinsSessionID: nil,
		FreeSpinsTriggered: engineResult.FreeSpinsTriggered,
		ReelPositions:      engineResult.ReelPositions,
		GameMode:           gameModePtr,
		GameModeCost:       gameModeCostPtr,
		CreatedAt:          engineResult.Timestamp,
	}

	// Save spin to database
	if err := s.spinRepo.Create(ctx, spinRecord); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to save spin")
		// Don't return error, spin was executed successfully
	}

	// Record spin in provably fair system (always required)
	// Dual Commitment Protocol: thetaSeed is passed on first spin for verification
	pfResult, err = s.pfService.RecordSpin(ctx, &provablyfair.RecordSpinInput{
		GameSessionID:     sessionID,
		SpinID:            spinRecord.ID,
		ClientSeed:        clientSeed, // Per-spin client seed
		ReelPositions:     engineResult.ReelPositions,
		ReelStripConfigID: engineResult.ReelStripConfigID,
		GameMode:          gameModePtr,
		IsFreeSpin:        false,
		ThetaSeed:         thetaSeed, // Dual Commitment Protocol: revealed on first spin
	})
	if err != nil {
		log.Error().Err(err).Str("spin_id", spinRecord.ID.String()).Msg("Failed to record spin in PF system")
		// Don't return error, spin was executed successfully
	} else {
		log.Debug().
			Str("spin_id", spinRecord.ID.String()).
			Int64("nonce", pfResult.Nonce).
			Str("spin_hash", pfResult.SpinHash).
			Msg("Spin recorded in PF system")
	}

	// Update session statistics
	if err := s.sessionRepo.UpdateStatistics(ctx, sessionID, 1, betAmount, engineResult.TotalWin); err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to update session statistics")
		// Don't return error
	}

	// Update player statistics
	if err := s.playerRepo.UpdateStatistics(ctx, playerID, 1, betAmount, engineResult.TotalWin); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to update player statistics")
		// Don't return error
	}

	log.Info().
		Str("spin_id", spinRecord.ID.String()).
		Str("player_id", playerID.String()).
		Float64("bet_amount", betAmount).
		Float64("total_win", engineResult.TotalWin).
		Bool("free_spins_triggered", engineResult.FreeSpinsTriggered).
		Msg("Spin executed successfully")

	// Create free spins session if triggered
	var freeSpinsSessionID *uuid.UUID
	var freeSpinsAwarded int
	if engineResult.FreeSpinsTriggered {
		freeSpinsSession, err := s.createFreeSpinsSession(ctx, sess.ID, playerID, spinRecord.ID, engineResult.ScatterCount, betAmount, gameMode)
		if err != nil {
			log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to create free spins session")
			// Don't return error, spin was executed successfully
		} else {
			freeSpinsSessionID = &freeSpinsSession.ID
			freeSpinsAwarded = freeSpinsSession.TotalSpinsAwarded

			// Update spin record with free spins session ID
			if err := s.spinRepo.UpdateFreeSpinsSessionId(ctx, spinRecord.ID, *freeSpinsSessionID); err != nil {
				log.Error().Err(err).Str("spin_id", spinRecord.ID.String()).Msg("Failed to update spin with free spins session ID")
			} else {
				spinRecord.FreeSpinsSessionID = freeSpinsSessionID
			}

			log.Info().
				Str("free_spins_session_id", freeSpinsSession.ID.String()).
				Str("player_id", playerID.String()).
				Int("scatter_count", engineResult.ScatterCount).
				Int("spins_awarded", freeSpinsAwarded).
				Msg("Free spins session created")
		}
	}

	// Build result
	result := &spin.SpinResult{
		SpinID:                  spinRecord.ID,
		SessionID:               sessionID,
		BetAmount:               betAmount,
		BalanceBefore:           balanceBefore,
		BalanceAfterBet:         balanceAfterBet,
		NewBalance:              newBalance,
		Grid:                    spinRecord.Grid,
		Cascades:                spinRecord.Cascades,
		SpinTotalWin:            engineResult.TotalWin,
		ScatterCount:            engineResult.ScatterCount,
		IsFreeSpin:              false,
		FreeSpinsTriggered:      engineResult.FreeSpinsTriggered,
		FreeSpinsRemainingSpins: freeSpinsAwarded,
		FreeSessionTotalWin:     0,
		GameMode:                gameMode,
		GameModeCost:            totalDeduction,
		Timestamp:               spinRecord.CreatedAt.Format(time.RFC3339),
	}

	// Add provably fair data if PF session was active
	if pfResult != nil {
		result.ProvablyFair = &spin.SpinProvablyFairData{
			SpinIndex:    pfResult.SpinIndex,
			Nonce:        pfResult.Nonce,
			SpinHash:     pfResult.SpinHash,
			PrevSpinHash: pfResult.PrevSpinHash,
		}
	}

	// Only set GameModeCost if a game mode was used
	if gameMode == "" {
		result.GameModeCost = 0
	}

	if freeSpinsSessionID != nil {
		result.FreeSpinsSessionID = freeSpinsSessionID.String()
	}

	return result, nil
}

// GetSpinDetails retrieves details of a specific spin
func (s *SpinService) GetSpinDetails(ctx context.Context, spinID uuid.UUID) (*spin.Spin, error) {
	spinRecord, err := s.spinRepo.GetByID(ctx, spinID)
	if err != nil {
		s.logger.Error().Err(err).Str("spin_id", spinID.String()).Msg("Failed to get spin details")
		return nil, spin.ErrSpinNotFound
	}

	return spinRecord, nil
}

// createFreeSpinsSession creates a new free spins session
func (s *SpinService) createFreeSpinsSession(
	ctx context.Context,
	sessionId uuid.UUID,
	playerID, spinID uuid.UUID,
	scatterCount int,
	betAmount float64,
	gameMode string,
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

	// Calculate free spins awarded (base: 12 for 3 scatters, +2 for each additional)
	spinsAwarded := 12 + (scatterCount-3)*2

	// Lookup reel strip config by gameMode (active, default)
	// If gameMode is empty or config not found, reelStripConfigID will be nil
	var reelStripConfigID *uuid.UUID
	if gameMode != "" && s.reelstripRepo != nil {
		config, err := s.reelstripRepo.GetDefaultConfig(ctx, gameMode)
		if err == nil && config != nil {
			reelStripConfigID = &config.ID
		} else {
			s.logger.Debug().
				Str("game_mode", gameMode).
				Msg("No default reel strip config found for game mode, using nil")
		}
	}

	// Create new free spins session
	newSession := &freespins.FreeSpinsSession{
		ID:                uuid.New(),
		PlayerID:          playerID,
		SessionID:         sessionId,
		TriggeredBySpinID: &spinID,
		ScatterCount:      scatterCount,
		TotalSpinsAwarded: spinsAwarded,
		SpinsCompleted:    0,
		RemainingSpins:    spinsAwarded,
		LockedBetAmount:   betAmount,
		TotalWon:          0.0,
		IsActive:          true,
		IsCompleted:       false,
		ReelStripConfigID: reelStripConfigID,
		CreatedAt:         time.Now().UTC(),
		CompletedAt:       nil,
	}

	// Save to database
	if err := s.freespinsRepo.Create(ctx, newSession); err != nil {
		s.logger.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to create free spins session")
		return nil, fmt.Errorf("failed to create free spins session: %w", err)
	}

	return newSession, nil
}

// GetSpinHistory retrieves spin history for a player
func (s *SpinService) GetSpinHistory(ctx context.Context, playerID uuid.UUID, page, limit int) (*spin.SpinHistoryResult, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // Default limit
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get total count
	total, err := s.spinRepo.Count(ctx, playerID)
	if err != nil {
		s.logger.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to count spins")
		return nil, fmt.Errorf("failed to count spins: %w", err)
	}

	// Get spins
	spins, err := s.spinRepo.GetByPlayer(ctx, playerID, limit, offset)
	if err != nil {
		s.logger.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get spin history")
		return nil, fmt.Errorf("failed to get spin history: %w", err)
	}

	result := &spin.SpinHistoryResult{
		Page:  page,
		Limit: limit,
		Total: total,
		Spins: spins,
	}

	return result, nil
}

// ExecuteTrialSpin executes a spin for a trial session
// Uses trial-specific reel strips with HUGE RTP for better winning experience
// Balance is managed in Redis only (no database writes)
func (s *SpinService) ExecuteTrialSpin(ctx context.Context, sessionToken string, trialSessionID uuid.UUID, betAmount float64, gameMode string) (*spin.SpinResult, error) {
	log := s.logger.WithTraceContext(ctx)

	if s.trialService == nil {
		return nil, fmt.Errorf("trial service not configured")
	}

	// Validate bet amount
	if betAmount <= 0 {
		return nil, fmt.Errorf("bet amount must be positive")
	}

	// Calculate deduction (same game mode logic as regular spin)
	var totalDeduction float64
	if gameMode != "" {
		cost, ok := gameModeCosts[gameMode]
		if !ok {
			return nil, fmt.Errorf("invalid game mode: %s", gameMode)
		}
		totalDeduction = cost.BuyCost
		betAmount = cost.BetAmount
	} else {
		totalDeduction = betAmount
	}

	// Deduct bet from trial balance in Redis
	balanceBefore, err := s.trialService.GetTrialBalance(ctx, sessionToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get trial balance")
		return nil, fmt.Errorf("failed to get trial balance: %w", err)
	}

	if balanceBefore < totalDeduction {
		log.Warn().
			Float64("balance", balanceBefore).
			Float64("bet_amount", betAmount).
			Float64("total_deduction", totalDeduction).
			Msg("Insufficient trial balance")
		return nil, player.ErrInsufficientBalance
	}

	// Execute trial spin using game engine with HUGE RTP
	engineResult, err := s.gameEngine.ExecuteTrialSpin(ctx, betAmount, gameMode)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute trial spin")
		return nil, fmt.Errorf("failed to execute spin: %w", err)
	}

	// Calculate new balance: deduct bet, add winnings
	balanceAfterBet := balanceBefore - totalDeduction
	newBalance := balanceAfterBet + engineResult.TotalWin

	// Update trial balance in Redis (single atomic update)
	if err := s.trialService.UpdateTrialBalance(ctx, sessionToken, newBalance); err != nil {
		log.Error().Err(err).Msg("Failed to update trial balance")
		// Don't return error, spin was executed
	}

	// Update trial stats (no DB write, just Redis)
	if err := s.trialService.UpdateTrialStats(ctx, sessionToken, 1, betAmount, engineResult.TotalWin); err != nil {
		log.Error().Err(err).Msg("Failed to update trial stats")
	}

	log.Info().
		Str("spin_id", engineResult.SpinID.String()).
		Str("trial_session", trialSessionID.String()).
		Float64("bet_amount", betAmount).
		Float64("total_win", engineResult.TotalWin).
		Float64("new_balance", newBalance).
		Bool("free_spins_triggered", engineResult.FreeSpinsTriggered).
		Msg("Trial spin executed successfully")

	// Handle free spins trigger for trial mode
	var freeSpinsSessionID *uuid.UUID
	var freeSpinsAwarded int
	if engineResult.FreeSpinsTriggered && s.trialService != nil {
		// Create trial free spins session in Redis
		// We need a game session ID - use the spin ID as a pseudo game session
		freeSpinsSession, err := s.trialService.StartTrialFreeSpins(
			ctx,
			trialSessionID,
			engineResult.SpinID, // Use spin ID as game session ID for trial
			engineResult.FreeSpinsAwarded,
			betAmount,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create trial free spins session")
		} else {
			freeSpinsSessionID = &freeSpinsSession.ID
			freeSpinsAwarded = freeSpinsSession.TotalSpins

			log.Info().
				Str("free_spins_session_id", freeSpinsSession.ID.String()).
				Int("scatter_count", engineResult.ScatterCount).
				Int("spins_awarded", freeSpinsAwarded).
				Msg("Trial free spins session created")
		}
	}

	// Build result (similar structure to regular spin, but no DB writes)
	result := &spin.SpinResult{
		SpinID:                  engineResult.SpinID,
		SessionID:               trialSessionID, // Use trial session ID
		BetAmount:               betAmount,
		BalanceBefore:           balanceBefore,
		BalanceAfterBet:         balanceAfterBet,
		NewBalance:              newBalance,
		Grid:                    convertGrid(engineResult.Grid),
		Cascades:                convertCascades(engineResult.Cascades),
		SpinTotalWin:            engineResult.TotalWin,
		ScatterCount:            engineResult.ScatterCount,
		IsFreeSpin:              false,
		FreeSpinsTriggered:      engineResult.FreeSpinsTriggered,
		FreeSpinsRemainingSpins: freeSpinsAwarded,
		FreeSessionTotalWin:     0,
		GameMode:                gameMode,
		GameModeCost:            totalDeduction,
		Timestamp:               engineResult.Timestamp.Format(time.RFC3339),
	}

	if gameMode == "" {
		result.GameModeCost = 0
	}

	if freeSpinsSessionID != nil {
		result.FreeSpinsSessionID = freeSpinsSessionID.String()
	}

	return result, nil
}
