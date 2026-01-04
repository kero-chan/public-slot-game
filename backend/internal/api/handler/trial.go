package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/trial"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/api/middleware"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/service"
)

// TrialHandler handles trial mode endpoints
type TrialHandler struct {
	trialService     *service.TrialService
	trialRateLimiter *middleware.TrialRateLimiter
	logger           *logger.Logger
}

// NewTrialHandler creates a new trial handler
func NewTrialHandler(
	trialService *service.TrialService,
	trialRateLimiter *middleware.TrialRateLimiter,
	log *logger.Logger,
) *TrialHandler {
	return &TrialHandler{
		trialService:     trialService,
		trialRateLimiter: trialRateLimiter,
		logger:           log,
	}
}

// StartTrial starts a new trial session
// POST /v1/auth/trial
// Security: Protected by TrialRateLimiter middleware (IP cooldown + max sessions)
// Memory optimization: Auto-cleans old session when same device creates new one
func (h *TrialHandler) StartTrial(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	var req dto.StartTrialRequest
	if err := c.BodyParser(&req); err != nil {
		// Body is optional, continue without it
		req = dto.StartTrialRequest{}
	}

	// Parse game_id if provided
	var gameID *uuid.UUID
	if req.GameID != "" {
		parsed, err := uuid.Parse(req.GameID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_game_id",
				Message: "Invalid game ID format",
			})
		}
		gameID = &parsed
	}

	// Get client IP from context (set by TrialRateLimiter middleware)
	clientIP, _ := c.Locals("client_ip").(string)
	if clientIP == "" {
		clientIP = c.Get("x-real-ip")
		if clientIP == "" {
			clientIP = c.IP()
		}
	}

	// Generate fingerprint from IP + User-Agent for per-browser limiting
	var fingerprint string
	if h.trialRateLimiter != nil {
		userAgent := c.Get("User-Agent")
		fingerprint = h.trialRateLimiter.GenerateDeviceFingerprint(clientIP, userAgent)
	}

	// Start trial session
	result, err := h.trialService.StartTrialSession(c.Context(), gameID)
	if err != nil {
		log.Error().Err(err).Str("ip", clientIP).Msg("Failed to start trial session")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "trial_error",
			Message: "Failed to start trial session",
		})
	}

	// Register new session for IP/fingerprint tracking
	if h.trialRateLimiter != nil {
		sessionTTL := trial.TrialSessionDuration
		if err := h.trialRateLimiter.RegisterTrialSession(c.Context(), clientIP, result.Session.SessionToken, fingerprint, sessionTTL); err != nil {
			// Log but don't fail - session was created successfully
			log.Warn().Err(err).Str("ip", clientIP).Msg("Failed to register trial session for tracking")
		}
	}

	log.Info().
		Str("trial_session_id", result.Session.ID.String()).
		Str("ip", clientIP).
		Str("fingerprint", fingerprint).
		Msg("Trial session started")

	// Return trial response
	return c.Status(fiber.StatusOK).JSON(dto.TrialResponse{
		SessionToken: result.Session.SessionToken,
		ExpiresAt:    result.ExpiresAt,
		Player: dto.TrialProfile{
			ID:           result.Session.ID.String(),
			Username:     "Trial Player",
			Balance:      result.Session.Balance,
			TotalSpins:   result.Session.TotalSpins,
			TotalWagered: result.Session.TotalWagered,
			TotalWon:     result.Session.TotalWon,
			IsTrial:      true,
		},
	})
}

// GetTrialProfile returns trial player profile
// GET /v1/trial/profile
func (h *TrialHandler) GetTrialProfile(c *fiber.Ctx) error {
	// Get trial session from context (set by middleware)
	trialSession, ok := c.Locals("trial_session").(*trial.TrialSession)
	if !ok || trialSession == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid trial session",
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.TrialProfile{
		ID:           trialSession.ID.String(),
		Username:     "Trial Player",
		Balance:      trialSession.Balance,
		TotalSpins:   trialSession.TotalSpins,
		TotalWagered: trialSession.TotalWagered,
		TotalWon:     trialSession.TotalWon,
		IsTrial:      true,
	})
}

// GetTrialBalance returns current trial balance
// GET /v1/trial/balance
func (h *TrialHandler) GetTrialBalance(c *fiber.Ctx) error {
	// Get trial session from context (set by middleware)
	trialSession, ok := c.Locals("trial_session").(*trial.TrialSession)
	if !ok || trialSession == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid trial session",
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.TrialBalanceResponse{
		Balance: trialSession.Balance,
	})
}
