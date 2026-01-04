package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/provablyfair"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/service"
)

// SessionHandler handles game session endpoints
type SessionHandler struct {
	sessionService session.Service
	pfService      *service.ProvablyFairService // Optional: nil if PF is disabled
	logger         *logger.Logger
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(
	sessionService session.Service,
	log *logger.Logger,
) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
		pfService:      nil, // PF service set separately via SetProvablyFairService
		logger:         log,
	}
}

// SetProvablyFairService sets the provably fair service for PF integration
func (h *SessionHandler) SetProvablyFairService(pfService *service.ProvablyFairService) {
	h.pfService = pfService
}

// StartSession starts a new game session
func (h *SessionHandler) StartSession(c *fiber.Ctx) error {
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

	// Parse request
	var req dto.StartSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Start session
	sess, err := h.sessionService.StartSession(c.Context(), playerID, req.BetAmount)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to start session")

		if err == session.ErrActiveSessionExists {
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "active_session_exists",
				Message: "Player already has an active session",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_start_session",
			Message: "Failed to start game session",
		})
	}

	// Build response
	response := dto.SessionResponse{
		ID:              sess.ID.String(),
		PlayerID:        sess.PlayerID.String(),
		BetAmount:       sess.BetAmount,
		StartingBalance: sess.StartingBalance,
		EndingBalance:   sess.EndingBalance,
		TotalSpins:      sess.TotalSpins,
		TotalWagered:    sess.TotalWagered,
		TotalWon:        sess.TotalWon,
		NetChange:       sess.NetChange,
		CreatedAt:       sess.CreatedAt,
		EndedAt:         sess.EndedAt,
	}

	// Start PF session if PF service is enabled
	// Dual Commitment Protocol: theta_commitment is sent by client BEFORE seeing server_seed
	if h.pfService != nil {
		pfResult, err := h.pfService.StartSession(c.Context(), playerID, sess.ID, req.ThetaCommitment)
		if err != nil {
			// Log error but don't fail the session start
			log.Warn().Err(err).Msg("Failed to start PF session, continuing without provably fair")
		} else {
			response.ProvablyFair = &dto.SessionProvablyFairData{
				SessionID:      pfResult.SessionID.String(),
				ServerSeedHash: pfResult.ServerSeedHash,
				NonceStart:     pfResult.NonceStart,
			}
			log.Info().
				Str("pf_session_id", pfResult.SessionID.String()).
				Str("server_seed_hash", pfResult.ServerSeedHash).
				Bool("has_theta_commitment", req.ThetaCommitment != "").
				Msg("PF session started with Dual Commitment Protocol")
		}
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// EndSession ends the current game session
func (h *SessionHandler) EndSession(c *fiber.Ctx) error {
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

	// Get session ID from URL params
	sessionIDStr := c.Params("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_session_id",
			Message: "Invalid session ID",
		})
	}

	// End session
	sess, err := h.sessionService.EndSession(c.Context(), sessionID)
	if err != nil {
		log.Error().Err(err).Str("session_id", sessionID.String()).Msg("Failed to end session")

		if err == session.ErrSessionNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "session_not_found",
				Message: "Session not found",
			})
		}

		if err == session.ErrSessionAlreadyEnded {
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "session_already_ended",
				Message: "Session already ended",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_end_session",
			Message: "Failed to end game session",
		})
	}

	// Build response
	response := dto.SessionResponse{
		ID:              sess.ID.String(),
		PlayerID:        sess.PlayerID.String(),
		BetAmount:       sess.BetAmount,
		StartingBalance: sess.StartingBalance,
		EndingBalance:   sess.EndingBalance,
		TotalSpins:      sess.TotalSpins,
		TotalWagered:    sess.TotalWagered,
		TotalWon:        sess.TotalWon,
		NetChange:       sess.NetChange,
		CreatedAt:       sess.CreatedAt,
		EndedAt:         sess.EndedAt,
	}

	// End PF session if PF service is enabled (reveals server_seed)
	if h.pfService != nil {
		pfResult, err := h.pfService.EndSession(c.Context(), sessionID)
		if err != nil {
			// Check if it's just a "not found" error (no PF session existed)
			if err != provablyfair.ErrStateNotFound && err != provablyfair.ErrSessionAlreadyEnded {
				log.Warn().Err(err).Msg("Failed to end PF session")
			}
		} else {
			// Convert spins to DTO format
			spins := make([]dto.SpinVerificationData, len(pfResult.Spins))
			for i, s := range pfResult.Spins {
				var configIDStr *string
				if s.ReelStripConfigID != nil {
					str := s.ReelStripConfigID.String()
					configIDStr = &str
				}
				spins[i] = dto.SpinVerificationData{
					SpinIndex:         s.SpinIndex,
					Nonce:             s.Nonce,
					ClientSeed:        s.ClientSeed,
					SpinHash:          s.SpinHash,
					PrevSpinHash:      s.PrevSpinHash,
					ReelPositions:     s.ReelPositions,
					ReelStripConfigID: configIDStr,
					GameMode:          s.GameMode,
					IsFreeSpin:        s.IsFreeSpin,
				}
			}

			response.ProvablyFair = &dto.SessionProvablyFairData{
				SessionID:      pfResult.SessionID.String(),
				ServerSeedHash: pfResult.ServerSeedHash,
				ServerSeed:     pfResult.ServerSeed, // Revealed!
				TotalSpins:     pfResult.TotalSpins,
				Spins:          spins,
			}
			log.Info().
				Str("pf_session_id", pfResult.SessionID.String()).
				Int64("pf_total_spins", pfResult.TotalSpins).
				Msg("PF session ended with game session, server seed revealed")
		}
	}

	log.Info().
		Str("session_id", sessionID.String()).
		Str("player_id", playerID.String()).
		Msg("Session ended successfully")

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetSessionHistory retrieves the player's session history
func (h *SessionHandler) GetSessionHistory(c *fiber.Ctx) error {
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

	// Get sessions
	sessions, err := h.sessionService.GetPlayerSessions(c.Context(), playerID, page, limit)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get session history")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_sessions",
			Message: "Failed to retrieve session history",
		})
	}

	// Build response
	sessionResponses := make([]dto.SessionResponse, len(sessions))
	for i, sess := range sessions {
		sessionResponses[i] = dto.SessionResponse{
			ID:              sess.ID.String(),
			PlayerID:        sess.PlayerID.String(),
			BetAmount:       sess.BetAmount,
			StartingBalance: sess.StartingBalance,
			EndingBalance:   sess.EndingBalance,
			TotalSpins:      sess.TotalSpins,
			TotalWagered:    sess.TotalWagered,
			TotalWon:        sess.TotalWon,
			NetChange:       sess.NetChange,
			CreatedAt:       sess.CreatedAt,
			EndedAt:         sess.EndedAt,
		}
	}

	response := dto.SessionHistoryResponse{
		Page:     page,
		Limit:    limit,
		Sessions: sessionResponses,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
