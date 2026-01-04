package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	adminDomain "github.com/slotmachine/backend/domain/admin"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// AdminPlayerHandler handles admin player management endpoints
type AdminPlayerHandler struct {
	adminService adminDomain.Service
	logger       *logger.Logger
}

// NewAdminPlayerHandler creates a new admin player handler
func NewAdminPlayerHandler(
	adminService adminDomain.Service,
	logger *logger.Logger,
) *AdminPlayerHandler {
	return &AdminPlayerHandler{
		adminService: adminService,
		logger:       logger,
	}
}

// getAdminFromContext retrieves the admin from fiber context (set by auth middleware)
func getAdminFromContext(c *fiber.Ctx) *adminDomain.Admin {
	admin, _ := c.Locals("admin").(*adminDomain.Admin)
	return admin
}

// GetPlayer retrieves a player by ID
func (h *AdminPlayerHandler) GetPlayer(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse player ID from path
	playerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_player_id",
			Message: "Invalid player ID format",
		})
	}

	// Get player from service
	player, err := h.adminService.GetPlayer(c.Context(), playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "get_player_failed",
			Message: "Failed to retrieve player",
		})
	}

	return c.JSON(player)
}

// CreatePlayer creates a new player
func (h *AdminPlayerHandler) CreatePlayer(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse request body
	var req struct {
		Username string  `json:"username"`
		Email    string  `json:"email"`
		Password string  `json:"password"`
		Balance  float64 `json:"balance"`
		GameID   *string `json:"game_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "validation_error",
			Message: "Username is required",
		})
	}
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "validation_error",
			Message: "Email is required",
		})
	}
	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "validation_error",
			Message: "Password is required",
		})
	}

	// Parse game_id if provided
	var gameID *uuid.UUID
	if req.GameID != nil && *req.GameID != "" {
		parsedID, err := uuid.Parse(*req.GameID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_game_id",
				Message: "Invalid game ID format",
			})
		}
		gameID = &parsedID
	}

	// Get admin from context (set by auth middleware)
	admin := getAdminFromContext(c)
	var createdBy uuid.UUID
	if admin != nil {
		createdBy = admin.ID
	}

	// Create player request
	createReq := adminDomain.CreatePlayerRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Balance:  req.Balance,
		GameID:   gameID,
	}

	// Create player
	player, err := h.adminService.CreatePlayer(c.Context(), createReq, createdBy)
	if err != nil {
		log.Error().Err(err).Str("username", req.Username).Msg("Failed to create player")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "create_player_failed",
			Message: err.Error(),
		})
	}

	log.Info().
		Str("username", req.Username).
		Str("admin_id", createdBy.String()).
		Msg("Player created successfully")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    player,
	})
}

// ListPlayers retrieves a list of players with optional filters
func (h *AdminPlayerHandler) ListPlayers(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse query parameters
	filters := adminDomain.PlayerListFilters{
		Username: c.Query("username"),
		Email:    c.Query("email"),
		Page:     1,
		Limit:    20,
		SortBy:   "created_at",
		SortDesc: true,
	}

	// Parse page
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filters.Page = p
		}
	}

	// Parse limit
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			filters.Limit = l
		}
	}

	// Parse game_id filter
	if gameID := c.Query("game_id"); gameID != "" {
		if parsedID, err := uuid.Parse(gameID); err == nil {
			filters.GameID = &parsedID
		}
	}

	// Parse is_active filter
	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filters.IsActive = &active
		}
	}

	// Parse sort_by
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}

	// Parse sort_desc
	if sortDesc := c.Query("sort_desc"); sortDesc != "" {
		if desc, err := strconv.ParseBool(sortDesc); err == nil {
			filters.SortDesc = desc
		}
	}

	// Get players from service
	players, total, err := h.adminService.ListPlayers(c.Context(), filters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list players")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "list_players_failed",
			Message: "Failed to retrieve players list",
		})
	}

	return c.JSON(fiber.Map{
		"players": players,
		"total":   total,
		"page":    filters.Page,
		"limit":   filters.Limit,
	})
}

// ActivatePlayer activates a player account
func (h *AdminPlayerHandler) ActivatePlayer(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse player ID from path
	playerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_player_id",
			Message: "Invalid player ID format",
		})
	}

	// Get admin from context (set by auth middleware)
	admin := getAdminFromContext(c)
	var updatedBy uuid.UUID
	if admin != nil {
		updatedBy = admin.ID
	}

	// Activate player
	if err := h.adminService.ActivatePlayer(c.Context(), playerID, updatedBy); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to activate player")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "activate_player_failed",
			Message: err.Error(),
		})
	}

	log.Info().
		Str("player_id", playerID.String()).
		Str("admin_id", updatedBy.String()).
		Msg("Player activated successfully")

	return c.JSON(fiber.Map{
		"message": "Player activated successfully",
	})
}

// DeactivatePlayer deactivates a player account
func (h *AdminPlayerHandler) DeactivatePlayer(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse player ID from path
	playerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_player_id",
			Message: "Invalid player ID format",
		})
	}

	// Get admin from context (set by auth middleware)
	admin := getAdminFromContext(c)
	var updatedBy uuid.UUID
	if admin != nil {
		updatedBy = admin.ID
	}

	// Deactivate player
	if err := h.adminService.DeactivatePlayer(c.Context(), playerID, updatedBy); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to deactivate player")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "deactivate_player_failed",
			Message: err.Error(),
		})
	}

	log.Info().
		Str("player_id", playerID.String()).
		Str("admin_id", updatedBy.String()).
		Msg("Player deactivated successfully")

	return c.JSON(fiber.Map{
		"message": "Player deactivated successfully",
	})
}

// ForceLogoutPlayer force logout a player from all sessions
func (h *AdminPlayerHandler) ForceLogoutPlayer(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse player ID from path
	playerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_player_id",
			Message: "Invalid player ID format",
		})
	}

	// Get admin from context (set by auth middleware)
	admin := getAdminFromContext(c)
	var adminUUID uuid.UUID
	if admin != nil {
		adminUUID = admin.ID
	}

	// Force logout player
	if err := h.adminService.ForceLogoutPlayer(c.Context(), playerID, adminUUID); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to force logout player")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "force_logout_failed",
			Message: err.Error(),
		})
	}

	log.Info().
		Str("player_id", playerID.String()).
		Str("admin_id", adminUUID.String()).
		Msg("Player force logged out successfully")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Player logged out from all sessions",
	})
}
