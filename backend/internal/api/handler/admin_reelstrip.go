package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/cache"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// AdminReelStripHandler handles admin endpoints for reel strip configuration management
type AdminReelStripHandler struct {
	reelStripService reelstrip.Service
	logger           *logger.Logger
	cache            *cache.Cache
}

// NewAdminReelStripHandler creates a new admin reel strip handler
func NewAdminReelStripHandler(
	reelStripService reelstrip.Service,
	log *logger.Logger,
	cache *cache.Cache,
) *AdminReelStripHandler {
	return &AdminReelStripHandler{
		reelStripService: reelStripService,
		logger:           log,
		cache:            cache,
	}
}

// CreateConfig creates a new reel strip configuration
// POST /admin/reel-strip-configs
func (h *AdminReelStripHandler) CreateConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	var req dto.CreateReelStripConfigRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	log.Info().
		Str("name", req.Name).
		Str("game_mode", req.GameMode).
		Msg("Creating reel strip config")

	config, err := h.reelStripService.CreateConfig(
		c.Context(),
		req.Name,
		req.GameMode,
		req.Description,
		req.ReelStripIDs,
		req.TargetRTP,
		nil,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_create_config",
			Message: "Failed to create reel strip configuration",
		})
	}

	// Update created_by and notes if provided
	if req.CreatedBy != "" {
		config.CreatedBy = req.CreatedBy
	}
	if req.Notes != "" {
		config.Notes = req.Notes
	}

	response := mapConfigToResponse(config)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetConfig retrieves a reel strip configuration by ID
// GET /admin/reel-strip-configs/:id
func (h *AdminReelStripHandler) GetConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	configID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid config ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_config_id",
			Message: "Invalid configuration ID",
		})
	}

	config, err := h.reelStripService.GetConfigByID(c.Context(), configID)
	if err != nil {
		if err == reelstrip.ErrConfigNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "config_not_found",
				Message: "Reel strip configuration not found",
			})
		}
		log.Error().Err(err).Msg("Failed to get config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_config",
			Message: "Failed to retrieve configuration",
		})
	}

	response := mapConfigToResponse(config)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// ListConfigs retrieves reel strip configurations with filtering and pagination
// GET /admin/reel-strip-configs?game_mode=base_game&is_active=true&page=1&limit=20
func (h *AdminReelStripHandler) ListConfigs(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse query parameters
	filters := &reelstrip.ConfigListFilters{}

	// Game mode filter
	if gameMode := c.Query("game_mode"); gameMode != "" {
		filters.GameMode = &gameMode
	}

	// is_active filter
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		filters.IsActive = &isActive
	}

	// is_default filter
	if isDefaultStr := c.Query("is_default"); isDefaultStr != "" {
		isDefault := isDefaultStr == "true"
		filters.IsDefault = &isDefault
	}

	// name filter
	if name := c.Query("name"); name != "" {
		filters.Name = &name
	}

	// Pagination
	filters.Page = c.QueryInt("page", 1)
	filters.Limit = c.QueryInt("limit", 20)

	configs, total, err := h.reelStripService.ListConfigs(c.Context(), filters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list configs")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_list_configs",
			Message: "Failed to retrieve configurations",
		})
	}

	responses := make([]dto.ReelStripConfigResponse, len(configs))
	for i, config := range configs {
		responses[i] = mapConfigToResponse(config)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"configs": responses,
			"total":   total,
			"page":    filters.Page,
			"limit":   filters.Limit,
		},
	})
}

// UpdateConfig updates a reel strip configuration
// PUT /admin/reel-strip-configs/:id
func (h *AdminReelStripHandler) UpdateConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	configID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid config ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_config_id",
			Message: "Invalid configuration ID",
		})
	}

	var req dto.UpdateReelStripConfigRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	config, err := h.reelStripService.GetConfigByID(c.Context(), configID)
	if err != nil {
		if err == reelstrip.ErrConfigNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "config_not_found",
				Message: "Reel strip configuration not found",
			})
		}
		log.Error().Err(err).Msg("Failed to get config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_config",
			Message: "Failed to retrieve configuration",
		})
	}

	// Update fields if provided
	if req.Name != nil {
		config.Name = *req.Name
	}
	if req.Description != nil {
		config.Description = *req.Description
	}
	if req.TargetRTP != nil {
		config.TargetRTP = *req.TargetRTP
	}
	if req.Notes != nil {
		config.Notes = *req.Notes
	}

	config.UpdatedAt = time.Now()

	// Clear cache for this config and related data
	h.clearConfigCache(c, configID, config.GameMode)

	// Save the updated config (you'll need to implement UpdateConfig in the service)
	// For now, this assumes you have the repo method available
	// You may need to add this to the service interface

	response := mapConfigToResponse(config)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// ActivateConfig activates a reel strip configuration
// POST /admin/reel-strip-configs/:id/activate
func (h *AdminReelStripHandler) ActivateConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	configID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid config ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_config_id",
			Message: "Invalid configuration ID",
		})
	}

	// Get config to know game mode for cache invalidation
	config, err := h.reelStripService.GetConfigByID(c.Context(), configID)
	if err != nil {
		if err == reelstrip.ErrConfigNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "config_not_found",
				Message: "Reel strip configuration not found",
			})
		}
		log.Error().Err(err).Msg("Failed to get config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_config",
			Message: "Failed to retrieve configuration",
		})
	}

	if err := h.reelStripService.ActivateConfig(c.Context(), configID); err != nil {
		log.Error().Err(err).Msg("Failed to activate config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_activate_config",
			Message: "Failed to activate configuration",
		})
	}

	// Clear cache for this config and related data
	h.clearConfigCache(c, configID, config.GameMode)

	log.Info().Str("config_id", configID.String()).Msg("Activated config")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Configuration activated successfully",
	})
}

// DeactivateConfig deactivates a reel strip configuration
// POST /admin/reel-strip-configs/:id/deactivate
func (h *AdminReelStripHandler) DeactivateConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	configID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid config ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_config_id",
			Message: "Invalid configuration ID",
		})
	}

	// Get config to know game mode for cache invalidation
	config, err := h.reelStripService.GetConfigByID(c.Context(), configID)
	if err != nil {
		if err == reelstrip.ErrConfigNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "config_not_found",
				Message: "Reel strip configuration not found",
			})
		}
		log.Error().Err(err).Msg("Failed to get config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_config",
			Message: "Failed to retrieve configuration",
		})
	}

	if err := h.reelStripService.DeactivateConfig(c.Context(), configID); err != nil {
		log.Error().Err(err).Msg("Failed to deactivate config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_deactivate_config",
			Message: "Failed to deactivate configuration",
		})
	}

	// Clear cache for this config and related data
	h.clearConfigCache(c, configID, config.GameMode)

	log.Info().Str("config_id", configID.String()).Msg("Deactivated config")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Configuration deactivated successfully",
	})
}

// SetDefaultConfig sets a configuration as the default for its game mode
// POST /admin/reel-strip-configs/set-default
func (h *AdminReelStripHandler) SetDefaultConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	var req dto.SetDefaultConfigRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.reelStripService.SetDefaultConfig(c.Context(), req.ConfigID, req.GameMode); err != nil {
		if err == reelstrip.ErrConfigNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "config_not_found",
				Message: "Reel strip configuration not found",
			})
		}
		log.Error().Err(err).Msg("Failed to set default config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_set_default",
			Message: "Failed to set default configuration",
		})
	}

	// Clear cache for default config and this specific config
	h.clearConfigCache(c, req.ConfigID, req.GameMode)

	log.Info().
		Str("config_id", req.ConfigID.String()).
		Str("game_mode", req.GameMode).
		Msg("Set default config")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Default configuration set successfully",
	})
}

// clearConfigCache clears all cache entries related to a reel strip configuration
func (h *AdminReelStripHandler) clearConfigCache(ctx *fiber.Ctx, configID uuid.UUID, gameMode string) {
	log := h.logger.WithTrace(ctx)

	// Clear default config cache for this game mode
	if err := h.cache.Expire(ctx.Context(), h.cache.ReelStripsDefaultKey(gameMode)); err != nil {
		log.Warn().Err(err).Str("game_mode", gameMode).Msg("Failed to clear default config cache")
	}

	// Clear default reel strip config cache
	if err := h.cache.Expire(ctx.Context(), h.cache.DefaultReelStripConfig(gameMode)); err != nil {
		log.Warn().Err(err).Str("game_mode", gameMode).Msg("Failed to clear default reel strip config cache")
	}

	// Clear specific config cache by ID
	if err := h.cache.Expire(ctx.Context(), h.cache.ReelStripsByConfigIdKey(configID)); err != nil {
		log.Warn().Err(err).Str("config_id", configID.String()).Msg("Failed to clear config by ID cache")
	}

	// Clear config set cache
	if err := h.cache.Expire(ctx.Context(), h.cache.ReelStripConfigSetKey(configID)); err != nil {
		log.Warn().Err(err).Str("config_id", configID.String()).Msg("Failed to clear config set cache")
	}

	// Clear config by ID cache
	if err := h.cache.Expire(ctx.Context(), h.cache.ReelStripConfigById(configID)); err != nil {
		log.Warn().Err(err).Str("config_id", configID.String()).Msg("Failed to clear config cache")
	}

	log.Debug().
		Str("config_id", configID.String()).
		Str("game_mode", gameMode).
		Msg("Cleared config cache")
}

// Helper function to map domain model to DTO
func mapConfigToResponse(config *reelstrip.ReelStripConfig) dto.ReelStripConfigResponse {
	return dto.ReelStripConfigResponse{
		ID:            config.ID,
		Name:          config.Name,
		GameMode:      config.GameMode,
		Description:   config.Description,
		Reel0StripID:  config.Reel0StripID,
		Reel1StripID:  config.Reel1StripID,
		Reel2StripID:  config.Reel2StripID,
		Reel3StripID:  config.Reel3StripID,
		Reel4StripID:  config.Reel4StripID,
		TargetRTP:     config.TargetRTP,
		IsActive:      config.IsActive,
		IsDefault:     config.IsDefault,
		ActivatedAt:   config.ActivatedAt,
		DeactivatedAt: config.DeactivatedAt,
		CreatedAt:     config.CreatedAt,
		UpdatedAt:     config.UpdatedAt,
		CreatedBy:     config.CreatedBy,
		Notes:         config.Notes,
	}
}
