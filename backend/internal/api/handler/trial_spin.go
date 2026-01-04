package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/trial"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/service"
)

// TrialSpinHandler handles spin endpoints for trial sessions
type TrialSpinHandler struct {
	spinService *service.SpinService
	logger      *logger.Logger
}

// NewTrialSpinHandler creates a new trial spin handler
func NewTrialSpinHandler(
	spinService *service.SpinService,
	log *logger.Logger,
) *TrialSpinHandler {
	return &TrialSpinHandler{
		spinService: spinService,
		logger:      log,
	}
}

// ExecuteSpin executes a trial spin
// POST /v1/trial/spin
func (h *TrialSpinHandler) ExecuteSpin(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get trial session from context (set by auth middleware)
	trialSession := c.Locals("trial_session").(*trial.TrialSession)
	sessionToken := c.Locals("session_token").(string)

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
		Str("trial_session_id", trialSession.ID.String()).
		Msg("Trial spin request received")

	// Execute trial spin
	result, err := h.spinService.ExecuteTrialSpin(
		c.Context(),
		sessionToken,
		trialSession.ID,
		req.BetAmount,
		req.GameMode,
	)
	if err != nil {
		log.Error().Err(err).Str("trial_session", trialSession.ID.String()).Msg("Failed to execute trial spin")

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

	return c.Status(fiber.StatusOK).JSON(response)
}
