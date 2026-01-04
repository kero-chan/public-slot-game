package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// PlayerHandler handles player-related endpoints
type PlayerHandler struct {
	playerService player.Service
	logger        *logger.Logger
}

// NewPlayerHandler creates a new player handler
func NewPlayerHandler(
	playerService player.Service,
	log *logger.Logger,
) *PlayerHandler {
	return &PlayerHandler{
		playerService: playerService,
		logger:        log,
	}
}

// GetBalance retrieves the player's balance
func (h *PlayerHandler) GetBalance(c *fiber.Ctx) error {
	// Create traced logger (includes traceID and clientIP)
	log := h.logger.WithTrace(c)

	// Get player ID from context (set by auth middleware)
	playerIDStr := c.Locals("user_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid player ID in token")
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid player ID in token",
		})
	}

	log.Debug().Str("player_id", playerID.String()).Msg("Fetching player balance")

	// Get balance
	balance, err := h.playerService.GetBalance(c.Context(), playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get balance")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_balance",
			Message: "Failed to retrieve balance",
		})
	}

	log.Info().Str("player_id", playerID.String()).Float64("balance", balance).Msg("Balance retrieved successfully")

	response := dto.GetBalanceResponse{
		Balance: balance,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
