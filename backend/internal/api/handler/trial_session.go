package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slotmachine/backend/domain/trial"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// TrialSessionHandler handles session endpoints for trial sessions
type TrialSessionHandler struct {
	logger *logger.Logger
}

// NewTrialSessionHandler creates a new trial session handler
func NewTrialSessionHandler(log *logger.Logger) *TrialSessionHandler {
	return &TrialSessionHandler{
		logger: log,
	}
}

// StartSession starts a virtual game session for trial
// POST /v1/trial/session/start
func (h *TrialSessionHandler) StartSession(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	trialSession := c.Locals("trial_session").(*trial.TrialSession)

	// Parse request for bet amount
	var req dto.StartSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Return a "virtual" session for trial (uses trial session ID)
	now := time.Now().UTC()
	response := dto.SessionResponse{
		ID:              trialSession.ID.String(),
		PlayerID:        trialSession.ID.String(),
		BetAmount:       req.BetAmount,
		StartingBalance: trialSession.Balance,
		EndingBalance:   nil,
		TotalSpins:      trialSession.TotalSpins,
		TotalWagered:    trialSession.TotalWagered,
		TotalWon:        trialSession.TotalWon,
		NetChange:       trialSession.TotalWon - trialSession.TotalWagered,
		CreatedAt:       now,
		EndedAt:         nil,
	}

	log.Info().
		Str("trial_session_id", trialSession.ID.String()).
		Float64("balance", trialSession.Balance).
		Msg("Trial session started (virtual)")

	return c.Status(fiber.StatusCreated).JSON(response)
}
