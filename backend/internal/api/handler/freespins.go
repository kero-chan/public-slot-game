package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/freespins"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/api/testdata"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// Blank import to keep testdata available for manual testing
// Uncomment the return statement in ExecuteFreeSpin to use test scenarios
var _ = testdata.SpinScenarios

// FreeSpinsHandler handles free spins endpoints
type FreeSpinsHandler struct {
	freeSpinsService freespins.Service
	logger           *logger.Logger
}

// NewFreeSpinsHandler creates a new free spins handler
func NewFreeSpinsHandler(
	freeSpinsService freespins.Service,
	log *logger.Logger,
) *FreeSpinsHandler {
	return &FreeSpinsHandler{
		freeSpinsService: freeSpinsService,
		logger:           log,
	}
}

// GetStatus retrieves the free spins status for the player
func (h *FreeSpinsHandler) GetStatus(c *fiber.Ctx) error {
	// Get player ID from context
	playerIDStr := c.Locals("user_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid player ID in token",
		})
	}

	// Get active session
	session, err := h.freeSpinsService.GetActiveSession(c.Context(), playerID)
	if err != nil {
		// No active session
		response := dto.FreeSpinsStatusResponse{
			Active:            false,
			TotalSpinsAwarded: 0,
			SpinsCompleted:    0,
			RemainingSpins:    0,
			LockedBetAmount:   0,
			TotalWon:          0,
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}

	// Build response
	response := dto.FreeSpinsStatusResponse{
		Active:             session.IsActive,
		FreeSpinsSessionID: session.ID.String(),
		SessionID:          session.SessionID.String(), // Game session ID for provably fair recovery
		TotalSpinsAwarded:  session.TotalSpinsAwarded,
		SpinsCompleted:     session.SpinsCompleted,
		RemainingSpins:     session.RemainingSpins,
		LockedBetAmount:    session.LockedBetAmount,
		TotalWon:           session.TotalWon,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// ExecuteFreeSpin executes a free spin
func (h *FreeSpinsHandler) ExecuteFreeSpin(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// To test specific free spin scenarios, uncomment one of the following lines:
	// return c.Status(fiber.StatusOK).JSON(testdata.SpinScenarios.LastFreeSpinNoCascade())
	// return c.Status(fiber.StatusOK).JSON(testdata.SpinScenarios.LastFreeSpinWithCascade())
	// return c.Status(fiber.StatusOK).JSON(testdata.SpinScenarios.FreeSpinRetrigger()) // Test retrigger overlay (every spin triggers retrigger)

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

	// Execute free spin (pass client seed for provably fair)
	result, err := h.freeSpinsService.ExecuteFreeSpin(c.Context(), freeSpinsSessionID, req.ClientSeed)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to execute free spin")

		if err == freespins.ErrFreeSpinsNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "free_spins_not_found",
				Message: "Free spins session not found",
			})
		}

		if err == freespins.ErrFreeSpinsNotActive {
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "free_spins_not_active",
				Message: "Free spins session is not active",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_execute_free_spin",
			Message: "Failed to execute free spin",
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
		FreeSpinsSessionID:      result.FreeSpinsSessionID,
		Grid:                    convertGrid(result.Grid),
		Cascades:                convertCascades(result.Cascades),
		SpinTotalWin:            result.SpinTotalWin,
		ScatterCount:            result.ScatterCount,
		IsFreeSpin:              result.IsFreeSpin,
		FreeSpinsTriggered:      result.FreeSpinsTriggered,
		FreeSpinsRetriggered:    result.FreeSpinsRetriggered,
		FreeSpinsAdditional:     result.FreeSpinsAdditional,
		FreeSpinsRemainingSpins: result.FreeSpinsRemainingSpins,
		FreeSessionTotalWin:     result.FreeSessionTotalWin,
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
