package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/spin"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/api/testdata"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// Blank import to keep testdata available for manual testing
// Uncomment the return statement in ExecuteSpin to use test scenarios
var _ = testdata.SpinScenarios

// SpinHandler handles spin-related endpoints
type SpinHandler struct {
	spinService spin.Service
	logger      *logger.Logger
}

// NewSpinHandler creates a new spin handler
func NewSpinHandler(
	spinService spin.Service,
	log *logger.Logger,
) *SpinHandler {
	return &SpinHandler{
		spinService: spinService,
		logger:      log,
	}
}

// GetInitialGrid generates a demo grid for initial display
// This ensures frontend has zero RNG - all symbol generation is backend-controlled
// No authentication required as this is just for visual display
func (h *SpinHandler) GetInitialGrid(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Generate initial grid
	grid, err := h.spinService.GenerateInitialGrid(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate initial grid")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_generate_grid",
			Message: "Failed to generate initial grid",
		})
	}

	// Build response
	response := dto.InitialGridResponse{
		Grid: convertGrid(grid),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// ExecuteSpin executes a regular spin
func (h *SpinHandler) ExecuteSpin(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// To test specific scenarios, uncomment one of the following lines:
	// return c.Status(fiber.StatusOK).JSON(testdata.SpinScenarios.BonusCascadeTrigger())
	// return c.Status(fiber.StatusOK).JSON(testdata.SpinScenarios.SimpleBonusTrigger())
	// return c.Status(fiber.StatusOK).JSON(testdata.SpinScenarios.BigWin())

	// Get player ID from context
	playerIDStr := c.Locals("user_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid player ID in token",
		})
	}

	// Parse request
	var req dto.ExecuteSpinRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	log.Info().
		Float64("bet_amount", req.BetAmount).
		Str("game_mode", req.GameMode).
		Str("session_id", req.SessionID).
		Msg("ExecuteSpin request received")

	// Parse session ID (allow empty for auto-session creation)
	var sessionID uuid.UUID
	if req.SessionID != "" {
		parsedSessionID, err := uuid.Parse(req.SessionID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_session_id",
				Message: "Invalid session ID",
			})
		}
		sessionID = parsedSessionID
	}

	// Execute spin (uuid.Nil if no session provided, service will handle it)
	// ClientSeed is optional - for provably fair sessions, client provides per-spin seed
	// ThetaSeed is required on first spin if theta_commitment was provided (Dual Commitment Protocol)
	result, err := h.spinService.ExecuteSpin(c.Context(), playerID, sessionID, req.BetAmount, req.GameMode, req.ClientSeed, req.ThetaSeed)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to execute spin")

		if err == player.ErrInsufficientBalance {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "insufficient_balance",
				Message: "Insufficient balance for this bet",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_execute_spin",
			Message: "Failed to execute spin",
		})
	}

	// Build response
	response := dto.SpinResponse{
		SpinID:                  result.SpinID.String(),
		SessionID:               result.SessionID.String(),
		BetAmount:               result.BetAmount,
		BalanceBefore:           result.BalanceBefore,
		BalanceAfterBet:         result.BalanceAfterBet,
		NewBalance:              result.NewBalance,
		Grid:                    convertGrid(result.Grid),
		Cascades:                convertCascades(result.Cascades),
		SpinTotalWin:            result.SpinTotalWin,
		ScatterCount:            result.ScatterCount,
		IsFreeSpin:              result.IsFreeSpin,
		FreeSpinsSessionID:      result.FreeSpinsSessionID,
		FreeSpinsTriggered:      result.FreeSpinsTriggered,
		FreeSpinsRemainingSpins: result.FreeSpinsRemainingSpins,
		GameMode:                result.GameMode,
		GameModeCost:            result.GameModeCost,
		Timestamp:               result.Timestamp,
	}

	// Add provably fair data if present
	if result.ProvablyFair != nil {
		response.ProvablyFair = &dto.SpinProvablyFairData{
			SpinHash:     result.ProvablyFair.SpinHash,
			PrevSpinHash: result.ProvablyFair.PrevSpinHash,
			Nonce:        result.ProvablyFair.Nonce,
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetSpinHistory retrieves the player's spin history
func (h *SpinHandler) GetSpinHistory(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get player ID from context
	playerIDStr := c.Locals("user_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid player ID in token",
		})
	}

	// Get pagination params
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	// Get spin history
	history, err := h.spinService.GetSpinHistory(c.Context(), playerID, page, limit)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get spin history")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_history",
			Message: "Failed to retrieve spin history",
		})
	}

	// Build response
	spinSummaries := make([]dto.SpinSummary, len(history.Spins))
	for i, s := range history.Spins {
		spinSummaries[i] = dto.SpinSummary{
			SpinID:             s.ID.String(),
			SessionID:          s.SessionID.String(),
			BetAmount:          s.BetAmount,
			TotalWin:           s.TotalWin,
			ScatterCount:       s.ScatterCount,
			IsFreeSpin:         s.IsFreeSpin,
			FreeSpinsTriggered: s.FreeSpinsTriggered,
			CreatedAt:          s.CreatedAt,
		}
	}

	response := dto.SpinHistoryResponse{
		Page:  history.Page,
		Limit: history.Limit,
		Total: history.Total,
		Spins: spinSummaries,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// convertCascades converts spin.Cascades to dto.CascadeInfo
func convertCascades(cascades spin.Cascades) []dto.CascadeInfo {
	result := make([]dto.CascadeInfo, len(cascades))
	for i, cascade := range cascades {
		wins := make([]dto.WinInfo, len(cascade.Wins))
		for j, win := range cascade.Wins {
			// Convert positions from wins.Position to dto.Position
			positions := make([]dto.Position, len(win.Positions))
			for k, pos := range win.Positions {
				positions[k] = dto.Position{
					Reel: pos.Reel,
					Row:  pos.Row,
				}
			}

			wins[j] = dto.WinInfo{
				Symbol:       symbols.SymbolNumber(win.Symbol),
				Count:        win.Count,
				Ways:         win.Ways,
				Payout:       win.Payout,
				WinAmount:    win.WinAmount,
				Positions:    positions,
				WinIntensity: string(symbols.GetWinIntensity(symbols.Symbol(win.Symbol), win.Count)),
			}
		}
		result[i] = dto.CascadeInfo{
			CascadeNumber:   cascade.CascadeNumber,
			GridAfter:       convertGrid(cascade.GridAfter),
			Multiplier:      cascade.Multiplier,
			Wins:            wins,
			TotalCascadeWin: cascade.TotalCascadeWin,
			WinningTileKind: cascade.WinningTileKind,
		}
	}
	return result
}

func convertGrid(grid spin.Grid) [][]int {
	result := make([][]int, len(grid))
	for i, row := range grid {
		result[i] = make([]int, len(row))
		for j, symbol := range row {
			result[i][j] = symbols.SymbolNumber(symbol)
		}
	}
	return result
}

