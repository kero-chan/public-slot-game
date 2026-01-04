package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	playerService player.Service
	logger        *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	playerService player.Service,
	log *logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		playerService: playerService,
		logger:        log,
	}
}

// Register handles player registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Parse game_id if provided
	var gameID *uuid.UUID
	if req.GameID != nil && *req.GameID != "" {
		parsed, err := uuid.Parse(*req.GameID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_game_id",
				Message: "Invalid game ID format",
			})
		}
		gameID = &parsed
	}

	// Register player
	p, err := h.playerService.Register(c.Context(), req.Username, req.Email, req.Password, gameID)
	if err != nil {
		log.Error().Err(err).Str("username", req.Username).Msg("Registration failed")

		if err == player.ErrPlayerAlreadyExists {
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "player_exists",
				Message: "Username or email already exists",
			})
		}

		if err == player.ErrGameNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "game_not_found",
				Message: "Specified game does not exist",
			})
		}

		if err == player.ErrGameIDRequired {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "game_id_required",
				Message: "game_id is required for registration",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "registration_failed",
			Message: "Failed to register player",
		})
	}

	// Build game_id string for response
	var gameIDStr *string
	if p.GameID != nil {
		s := p.GameID.String()
		gameIDStr = &s
	}

	// Build response - registration returns player info only, user must login to get session
	response := dto.RegisterResponse{
		Message: "Registration successful. Please login to continue.",
		Player: dto.PlayerProfile{
			ID:           p.ID.String(),
			Username:     p.Username,
			Email:        p.Email,
			Balance:      p.Balance,
			GameID:       gameIDStr,
			TotalSpins:   p.TotalSpins,
			TotalWagered: p.TotalWagered,
			TotalWon:     p.TotalWon,
			IsActive:     p.IsActive,
			IsVerified:   p.IsVerified,
			CreatedAt:    p.CreatedAt,
			LastLoginAt:  p.LastLoginAt,
		},
	}

	log.Info().Str("player_id", p.ID.String()).Msg("Player registered successfully")

	return c.Status(fiber.StatusCreated).JSON(response)
}

// Login handles player login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Parse game_id if provided
	var gameID *uuid.UUID
	if req.GameID != nil && *req.GameID != "" {
		parsed, err := uuid.Parse(*req.GameID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_game_id",
				Message: "Invalid game ID format",
			})
		}
		gameID = &parsed
	}

	clientIp := c.Get("x-real-ip")
	if clientIp == "" {
		clientIp = c.IP()
	}

	// Build login options
	loginOpts := &player.LoginOptions{
		ForceLogout: req.ForceLogout,
		IPAddress:   clientIp,
		UserAgent:   string(c.Request().Header.UserAgent()),
	}
	if req.DeviceInfo != nil {
		loginOpts.DeviceInfo = *req.DeviceInfo
	}

	// Authenticate player
	result, err := h.playerService.Login(c.Context(), req.Username, req.Password, gameID, loginOpts)
	if err != nil {
		log.Warn().Str("username", req.Username).Err(err).Msg("Login failed")

		if err == player.ErrInvalidCredentials {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid username or password",
			})
		}

		if err == player.ErrGameAccessDenied {
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error:   "game_access_denied",
				Message: "Player not authorized for this game",
			})
		}

		if errors.Is(err, session.ErrPlayerAlreadyLoggedIn) {
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "already_logged_in",
				Message: "Player is already logged in on another device. Set force_logout=true to logout other device.",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "login_failed",
			Message: "Failed to authenticate player",
		})
	}

	p := result.Player

	// Build game_id string for response
	var gameIDStr *string
	if p.GameID != nil {
		s := p.GameID.String()
		gameIDStr = &s
	}

	// Build response - session-based auth, no JWT
	response := dto.AuthResponse{
		SessionToken: result.SessionToken,
		ExpiresAt:    result.ExpiresAt,
		Player: dto.PlayerProfile{
			ID:           p.ID.String(),
			Username:     p.Username,
			Email:        p.Email,
			Balance:      p.Balance,
			GameID:       gameIDStr,
			TotalSpins:   p.TotalSpins,
			TotalWagered: p.TotalWagered,
			TotalWon:     p.TotalWon,
			IsActive:     p.IsActive,
			IsVerified:   p.IsVerified,
			CreatedAt:    p.CreatedAt,
			LastLoginAt:  p.LastLoginAt,
		},
	}

	log.Info().Str("player_id", p.ID.String()).Msg("Player logged in successfully")

	return c.Status(fiber.StatusOK).JSON(response)
}

// Logout handles player logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get session token from context (set by auth middleware)
	sessionToken, ok := c.Locals("session_token").(string)
	if !ok || sessionToken == "" {
		// Try to get from request body
		var req dto.LogoutRequest
		if err := c.BodyParser(&req); err == nil && req.SessionToken != "" {
			sessionToken = req.SessionToken
		}
	}

	if sessionToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "session_token_required",
			Message: "Session token is required for logout",
		})
	}

	// Logout
	if err := h.playerService.Logout(c.Context(), sessionToken); err != nil {
		log.Error().Err(err).Msg("Logout failed")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "logout_failed",
			Message: "Failed to logout",
		})
	}

	log.Info().Msg("Player logged out successfully")

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// GetProfile retrieves the authenticated player's profile
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get player ID from context (set by auth middleware)
	playerIDStr := c.Locals("user_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid player ID in token",
		})
	}

	// Get player profile
	p, err := h.playerService.GetProfile(c.Context(), playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get profile")
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "player_not_found",
			Message: "Player not found",
		})
	}

	// Build game_id string for response
	var gameIDStr *string
	if p.GameID != nil {
		s := p.GameID.String()
		gameIDStr = &s
	}

	// Build response
	profile := dto.PlayerProfile{
		ID:           p.ID.String(),
		Username:     p.Username,
		Email:        p.Email,
		Balance:      p.Balance,
		GameID:       gameIDStr,
		TotalSpins:   p.TotalSpins,
		TotalWagered: p.TotalWagered,
		TotalWon:     p.TotalWon,
		IsActive:     p.IsActive,
		IsVerified:   p.IsVerified,
		CreatedAt:    p.CreatedAt,
		LastLoginAt:  p.LastLoginAt,
	}

	return c.Status(fiber.StatusOK).JSON(profile)
}
