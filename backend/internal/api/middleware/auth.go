package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/internal/pkg/errors"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/service"
)

// SessionAuthMiddleware validates session tokens directly (no JWT)
// Extracts session token from Authorization header and validates against Redis/DB
// Also validates that the session's game matches the requested game (X-Game-ID header)
// Supports trial tokens (prefixed with "trial_")
func SessionAuthMiddleware(log *logger.Logger, playerService player.Service, trialService *service.TrialService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return respondError(c, errors.Unauthorized("Missing authorization header"))
		}

		// Check Bearer scheme and extract session token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return respondError(c, errors.Unauthorized("Invalid authorization header format"))
		}
		sessionToken := parts[1]

		if sessionToken == "" {
			return respondError(c, errors.Unauthorized("Missing session token"))
		}

		// Get client IP for logging
		clientIP := c.Get("x-real-ip")
		if clientIP == "" {
			clientIP = c.IP()
		}

		// Get requested game ID from header
		gameIDHeader := c.Get("X-Game-ID")
		var requestedGameID *uuid.UUID
		if gameIDHeader != "" {
			parsed, err := uuid.Parse(gameIDHeader)
			if err != nil {
				return respondError(c, errors.BadRequest("Invalid X-Game-ID format"))
			}
			requestedGameID = &parsed
		}

		// Check if this is a trial session token
		if service.IsTrialToken(sessionToken) {
			return handleTrialSession(c, log, trialService, sessionToken, requestedGameID, clientIP)
		}

		// Validate session with player service (checks Redis first, then DB)
		result, err := playerService.ValidateSession(c.Context(), sessionToken, requestedGameID)
		if err != nil {
			// Truncate session token for logging
			tokenPreview := sessionToken
			if len(sessionToken) > 8 {
				tokenPreview = sessionToken[:8] + "..."
			}
			log.Warn().
				Str("session_token", tokenPreview).
				Str("ip", clientIP).
				Err(err).
				Msg("Session validation failed")
			return respondError(c, errors.Unauthorized(err.Error()))
		}

		// Store user info in context for handlers
		c.Locals("user_id", result.Player.ID.String())
		c.Locals("username", result.Player.Username)
		c.Locals("session_token", sessionToken)
		c.Locals("player", result.Player)
		c.Locals("is_trial", false)
		if result.Player.GameID != nil {
			c.Locals("game_id", result.Player.GameID.String())
		}

		return c.Next()
	}
}

// handleTrialSession handles authentication for trial session tokens
func handleTrialSession(c *fiber.Ctx, log *logger.Logger, trialService *service.TrialService, sessionToken string, requestedGameID *uuid.UUID, clientIP string) error {
	if trialService == nil {
		log.Warn().Str("ip", clientIP).Msg("Trial mode attempted but trial service not available")
		return respondError(c, errors.Unauthorized("Trial mode not available"))
	}

	// Validate trial session from Redis
	trialSession, err := trialService.ValidateTrialSession(c.Context(), sessionToken, requestedGameID)
	if err != nil {
		tokenPreview := sessionToken
		if len(sessionToken) > 12 {
			tokenPreview = sessionToken[:12] + "..."
		}
		log.Warn().
			Str("session_token", tokenPreview).
			Str("ip", clientIP).
			Err(err).
			Msg("Trial session validation failed")
		return respondError(c, errors.Unauthorized(err.Error()))
	}

	// Store trial session info in context for handlers
	c.Locals("user_id", trialSession.ID.String())
	c.Locals("username", "Trial Player")
	c.Locals("session_token", sessionToken)
	c.Locals("is_trial", true)
	c.Locals("trial_session", trialSession)
	if trialSession.GameID != nil {
		c.Locals("game_id", trialSession.GameID.String())
	}

	return c.Next()
}

// respondError sends an error response
func respondError(c *fiber.Ctx, err *errors.HTTPError) error {
	return c.Status(err.StatusCode).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    err.Code,
			"message": err.Message,
			"details": err.Details,
		},
	})
}
