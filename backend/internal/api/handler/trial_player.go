package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/slotmachine/backend/domain/trial"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// TrialPlayerHandler handles player endpoints for trial sessions
type TrialPlayerHandler struct {
	logger *logger.Logger
}

// NewTrialPlayerHandler creates a new trial player handler
func NewTrialPlayerHandler(log *logger.Logger) *TrialPlayerHandler {
	return &TrialPlayerHandler{
		logger: log,
	}
}

// GetBalance retrieves the trial player's balance
// GET /v1/trial/player/balance
func (h *TrialPlayerHandler) GetBalance(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	trialSession := c.Locals("trial_session").(*trial.TrialSession)

	log.Debug().
		Str("trial_session_id", trialSession.ID.String()).
		Float64("balance", trialSession.Balance).
		Msg("Trial balance retrieved")

	return c.Status(fiber.StatusOK).JSON(dto.GetBalanceResponse{
		Balance: trialSession.Balance,
	})
}
