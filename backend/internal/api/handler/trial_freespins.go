package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/trial"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/game/cascade"
	"github.com/slotmachine/backend/internal/game/engine"
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/slotmachine/backend/internal/game/wins"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/service"
)

// TrialFreeSpinsHandler handles free spins endpoints for trial sessions
type TrialFreeSpinsHandler struct {
	trialService *service.TrialService
	gameEngine   *engine.GameEngine
	logger       *logger.Logger
}

// NewTrialFreeSpinsHandler creates a new trial free spins handler
func NewTrialFreeSpinsHandler(
	trialService *service.TrialService,
	gameEngine *engine.GameEngine,
	log *logger.Logger,
) *TrialFreeSpinsHandler {
	return &TrialFreeSpinsHandler{
		trialService: trialService,
		gameEngine:   gameEngine,
		logger:       log,
	}
}

// GetStatus retrieves the free spins status for trial session
// GET /v1/trial/free-spins/status
func (h *TrialFreeSpinsHandler) GetStatus(c *fiber.Ctx) error {
	trialSession := c.Locals("trial_session").(*trial.TrialSession)

	// Get active trial free spins from Redis
	freeSpins, err := h.trialService.GetActiveTrialFreeSpins(c.Context(), trialSession.ID)
	if err != nil || freeSpins == nil {
		// No active free spins session
		return c.Status(fiber.StatusOK).JSON(dto.FreeSpinsStatusResponse{
			Active:            false,
			TotalSpinsAwarded: 0,
			SpinsCompleted:    0,
			RemainingSpins:    0,
			LockedBetAmount:   0,
			TotalWon:          0,
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.FreeSpinsStatusResponse{
		Active:             freeSpins.IsActive,
		FreeSpinsSessionID: freeSpins.ID.String(),
		SessionID:          freeSpins.GameSessionID.String(), // Game session ID for provably fair recovery
		TotalSpinsAwarded:  freeSpins.TotalSpins,
		SpinsCompleted:     freeSpins.CompletedSpins,
		RemainingSpins:     freeSpins.RemainingSpins,
		LockedBetAmount:    freeSpins.LockedBetAmount,
		TotalWon:           freeSpins.TotalWon,
	})
}

// ExecuteFreeSpin executes a free spin for trial session
// POST /v1/trial/free-spins/spin
func (h *TrialFreeSpinsHandler) ExecuteFreeSpin(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	trialSession := c.Locals("trial_session").(*trial.TrialSession)
	sessionToken := c.Locals("session_token").(string)

	// Parse request
	var req dto.ExecuteFreeSpinRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Parse free spins session ID
	freeSpinsSessionID, err := uuid.Parse(req.FreeSpinsSessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_free_spins_session_id",
			Message: "Invalid free spins session ID",
		})
	}

	// Get trial free spins session from Redis
	freeSpins, err := h.trialService.GetTrialFreeSpins(c.Context(), freeSpinsSessionID)
	if err != nil || freeSpins == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "free_spins_not_found",
			Message: "Free spins session not found",
		})
	}

	if !freeSpins.IsActive || freeSpins.RemainingSpins <= 0 {
		return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
			Error:   "free_spins_not_active",
			Message: "Free spins session is not active",
		})
	}

	// Get current trial balance
	balanceBefore := trialSession.Balance

	// Execute trial free spin using game engine with HUGE RTP
	spinNumber := freeSpins.CompletedSpins + 1
	engineResult, err := h.gameEngine.ExecuteTrialFreeSpin(
		freeSpins.LockedBetAmount,
		freeSpins.RemainingSpins,
		spinNumber,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute trial free spin")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_execute_free_spin",
			Message: "Failed to execute free spin",
		})
	}

	// Update trial balance with winnings (free spins don't deduct bet)
	newBalance := balanceBefore + engineResult.TotalWin
	if err := h.trialService.UpdateTrialBalance(c.Context(), sessionToken, newBalance); err != nil {
		log.Error().Err(err).Msg("Failed to update trial balance")
	}

	// Update free spins session
	freeSpins.CompletedSpins++
	freeSpins.RemainingSpins = engineResult.RemainingSpins
	freeSpins.TotalWon += engineResult.TotalWin

	// Handle retrigger
	if engineResult.Retriggered {
		freeSpins.TotalSpins += engineResult.AdditionalSpins
	}

	// Check if completed
	if freeSpins.RemainingSpins <= 0 {
		freeSpins.IsActive = false
	}

	// Save updated free spins session
	if err := h.trialService.UpdateTrialFreeSpins(c.Context(), freeSpins); err != nil {
		log.Error().Err(err).Msg("Failed to update trial free spins session")
	}

	log.Info().
		Str("trial_session_id", trialSession.ID.String()).
		Int("spin_number", spinNumber).
		Float64("win", engineResult.TotalWin).
		Int("remaining", freeSpins.RemainingSpins).
		Bool("retriggered", engineResult.Retriggered).
		Msg("Trial free spin executed")

	// Build response
	response := dto.SpinResponse{
		SpinID:                  engineResult.SpinID.String(),
		SessionID:               trialSession.ID.String(),
		BetAmount:               freeSpins.LockedBetAmount,
		BalanceBefore:           balanceBefore,
		BalanceAfterBet:         balanceBefore, // No deduction for free spins
		NewBalance:              newBalance,
		FreeSpinsSessionID:      freeSpins.ID.String(),
		Grid:                    convertTrialGrid(engineResult.Grid),
		Cascades:                convertTrialCascades(engineResult.Cascades),
		SpinTotalWin:            engineResult.TotalWin,
		ScatterCount:            engineResult.ScatterCount,
		IsFreeSpin:              true,
		FreeSpinsTriggered:      false,
		FreeSpinsRetriggered:    engineResult.Retriggered,
		FreeSpinsAdditional:     engineResult.AdditionalSpins,
		FreeSpinsRemainingSpins: freeSpins.RemainingSpins,
		FreeSessionTotalWin:     freeSpins.TotalWon,
		Timestamp:               time.Now().UTC().Format(time.RFC3339),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// Trial-specific conversion functions

// convertTrialGrid converts engine reels.Grid to [][]int for JSON response
func convertTrialGrid(grid reels.Grid) [][]int {
	result := make([][]int, len(grid))
	for i, row := range grid {
		result[i] = make([]int, len(row))
		for j, symbol := range row {
			result[i][j] = symbols.SymbolNumber(symbol)
		}
	}
	return result
}

// convertTrialCascades converts engine cascade.CascadeResult to dto.CascadeInfo
func convertTrialCascades(cascades []cascade.CascadeResult) []dto.CascadeInfo {
	result := make([]dto.CascadeInfo, len(cascades))
	for i, cascadeResult := range cascades {
		winsInfo := make([]dto.WinInfo, len(cascadeResult.Wins))
		for j, win := range cascadeResult.Wins {
			positions := make([]dto.Position, len(win.Positions))
			for k, pos := range win.Positions {
				positions[k] = dto.Position{
					Reel: pos.Reel,
					Row:  pos.Row,
				}
			}

			winsInfo[j] = dto.WinInfo{
				Symbol:       symbols.SymbolNumber(string(win.Symbol)),
				Count:        win.Count,
				Ways:         win.Ways,
				Payout:       win.Payout,
				WinAmount:    win.WinAmount,
				Positions:    positions,
				WinIntensity: string(symbols.GetWinIntensity(win.Symbol, win.Count)),
			}
		}
		result[i] = dto.CascadeInfo{
			CascadeNumber:   cascadeResult.CascadeNumber,
			GridAfter:       convertTrialGrid(cascadeResult.GridAfter),
			Multiplier:      cascadeResult.Multiplier,
			Wins:            winsInfo,
			TotalCascadeWin: cascadeResult.TotalCascadeWin,
			WinningTileKind: extractTrialHighestWinningSymbol(cascadeResult.Wins),
		}
	}
	return result
}

// extractTrialHighestWinningSymbol extracts the highest priority winning symbol
func extractTrialHighestWinningSymbol(engineWins []wins.CascadeWinDetail) string {
	highValueSymbols := []symbols.Symbol{
		symbols.SymbolFa,
		symbols.SymbolZhong,
		symbols.SymbolBai,
		symbols.SymbolBawan,
	}

	winningSymbols := make(map[symbols.Symbol]bool)
	for _, win := range engineWins {
		baseSym := symbols.GetBaseSymbol(string(win.Symbol))
		if baseSym != symbols.SymbolWild {
			winningSymbols[baseSym] = true
		}
	}

	for _, prioritySym := range highValueSymbols {
		if winningSymbols[prioritySym] {
			return string(prioritySym)
		}
	}
	return ""
}
