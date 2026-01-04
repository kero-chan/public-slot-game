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

// AdminPlayerAssignmentHandler handles admin endpoints for player reel strip assignments
type AdminPlayerAssignmentHandler struct {
	reelStripService reelstrip.Service
	logger           *logger.Logger
	cache            *cache.Cache
}

// NewAdminPlayerAssignmentHandler creates a new admin player assignment handler
func NewAdminPlayerAssignmentHandler(
	reelStripService reelstrip.Service,
	log *logger.Logger,
	cache *cache.Cache,
) *AdminPlayerAssignmentHandler {
	return &AdminPlayerAssignmentHandler{
		reelStripService: reelStripService,
		logger:           log,
		cache:            cache,
	}
}

// CreateAssignment creates a new player reel strip assignment
// POST /admin/player-assignments
func (h *AdminPlayerAssignmentHandler) CreateAssignment(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	var req dto.CreatePlayerAssignmentRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate that at least one config is provided
	if req.BaseGameConfigID == nil && req.FreeSpinsConfigID == nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "At least one config (base_game or free_spins) must be provided",
		})
	}

	log.Info().
		Str("player_id", req.PlayerID.String()).
		Msg("Creating player assignment")

	// Check for existing assignment first
	existing, _ := h.reelStripService.GetPlayerAssignment(c.Context(), req.PlayerID)

	// Determine which configs need to be assigned
	needBaseGame := req.BaseGameConfigID != nil
	needFreeSpins := req.FreeSpinsConfigID != nil

	// If assignment exists, we only need to update the configs that changed
	if existing != nil && existing.IsActive {
		if needBaseGame && (existing.BaseGameConfigID == nil || *existing.BaseGameConfigID != *req.BaseGameConfigID) {
			if err := h.reelStripService.AssignConfigToPlayer(
				c.Context(),
				req.PlayerID,
				*req.BaseGameConfigID,
				"base_game",
				req.Reason,
				req.AssignedBy,
				req.ExpiresAt,
			); err != nil {
				log.Error().Err(err).Msg("Failed to update base game config")
				return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
					Error:   "failed_to_update_base_game",
					Message: "Failed to update base game configuration",
				})
			}
		}

		if needFreeSpins && (existing.FreeSpinsConfigID == nil || *existing.FreeSpinsConfigID != *req.FreeSpinsConfigID) {
			if err := h.reelStripService.AssignConfigToPlayer(
				c.Context(),
				req.PlayerID,
				*req.FreeSpinsConfigID,
				"free_spins",
				req.Reason,
				req.AssignedBy,
				req.ExpiresAt,
			); err != nil {
				log.Error().Err(err).Msg("Failed to update free spins config")
				return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
					Error:   "failed_to_update_free_spins",
					Message: "Failed to update free spins configuration",
				})
			}
		}
	} else {
		// No existing assignment - create new one
		// Assign base game first if provided
		if needBaseGame {
			if err := h.reelStripService.AssignConfigToPlayer(
				c.Context(),
				req.PlayerID,
				*req.BaseGameConfigID,
				"base_game",
				req.Reason,
				req.AssignedBy,
				req.ExpiresAt,
			); err != nil {
				log.Error().Err(err).Msg("Failed to assign base game config")
				return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
					Error:   "failed_to_assign_base_game",
					Message: "Failed to assign base game configuration",
				})
			}
		}

		// Then assign free spins if provided (this will update the same assignment)
		if needFreeSpins {
			if err := h.reelStripService.AssignConfigToPlayer(
				c.Context(),
				req.PlayerID,
				*req.FreeSpinsConfigID,
				"free_spins",
				req.Reason,
				req.AssignedBy,
				req.ExpiresAt,
			); err != nil {
				log.Error().Err(err).Msg("Failed to assign free spins config")
				return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
					Error:   "failed_to_assign_free_spins",
					Message: "Failed to assign free spins configuration",
				})
			}
		}
	}

	// Clear player assignment cache
	h.clearPlayerAssignmentCache(c, req.PlayerID)

	// Get the created/updated assignment to return in response
	createdAssignment, err := h.reelStripService.GetPlayerAssignment(c.Context(), req.PlayerID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get created assignment")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_assignment",
			Message: "Assignment created but failed to retrieve details",
		})
	}

	response := mapAssignmentToResponse(createdAssignment)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    response,
		"message": "Player assignment created successfully",
	})
}

// GetPlayerAssignment retrieves a player's active assignment
// GET /admin/player-assignments/:playerId
func (h *AdminPlayerAssignmentHandler) GetPlayerAssignment(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	playerID, err := uuid.Parse(c.Params("playerId"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid player ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_player_id",
			Message: "Invalid player ID",
		})
	}

	assignment, err := h.reelStripService.GetPlayerAssignment(c.Context(), playerID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get player assignment")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_assignment",
			Message: "Failed to retrieve player assignment",
		})
	}

	if assignment == nil || !assignment.IsActive {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "assignment_not_found",
			Message: "No active assignment found for this player",
		})
	}

	response := mapAssignmentToResponse(assignment)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// UpdateAssignment updates a player assignment
// PUT /admin/player-assignments/:playerId
func (h *AdminPlayerAssignmentHandler) UpdateAssignment(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	playerID, err := uuid.Parse(c.Params("playerId"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid player ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_player_id",
			Message: "Invalid player ID",
		})
	}

	var req dto.UpdatePlayerAssignmentRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Get existing assignment
	assignment, err := h.reelStripService.GetPlayerAssignment(c.Context(), playerID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get player assignment")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_assignment",
			Message: "Failed to retrieve player assignment",
		})
	}

	if assignment == nil || !assignment.IsActive {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "assignment_not_found",
			Message: "No active assignment found for this player",
		})
	}

	// Build reason string
	reason := assignment.Reason
	if req.Reason != nil {
		reason = *req.Reason
	}

	// Keep existing assignedBy if not provided in update
	assignedBy := assignment.AssignedBy

	// Use provided expiresAt or keep existing
	expiresAt := assignment.ExpiresAt
	if req.ExpiresAt != nil {
		expiresAt = req.ExpiresAt
	}

	// Update base game config if provided
	if req.BaseGameConfigID != nil {
		if err := h.reelStripService.AssignConfigToPlayer(
			c.Context(),
			playerID,
			*req.BaseGameConfigID,
			"base_game",
			reason,
			assignedBy,
			expiresAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to update base game config")
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "failed_to_update_base_game",
				Message: "Failed to update base game configuration",
			})
		}
	}

	// Update free spins config if provided
	if req.FreeSpinsConfigID != nil {
		if err := h.reelStripService.AssignConfigToPlayer(
			c.Context(),
			playerID,
			*req.FreeSpinsConfigID,
			"free_spins",
			reason,
			assignedBy,
			expiresAt,
		); err != nil {
			log.Error().Err(err).Msg("Failed to update free spins config")
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "failed_to_update_free_spins",
				Message: "Failed to update free spins configuration",
			})
		}
	}

	// If IsActive is being set to false, remove the assignment
	if req.IsActive != nil && !*req.IsActive {
		if err := h.reelStripService.RemovePlayerAssignment(c.Context(), playerID); err != nil {
			log.Error().Err(err).Msg("Failed to deactivate assignment")
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "failed_to_deactivate",
				Message: "Failed to deactivate player assignment",
			})
		}

		// Clear cache and return success
		h.clearPlayerAssignmentCache(c, playerID)

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Player assignment deactivated successfully",
		})
	}

	// Clear cache
	h.clearPlayerAssignmentCache(c, playerID)

	// Get updated assignment
	updatedAssignment, err := h.reelStripService.GetPlayerAssignment(c.Context(), playerID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get updated assignment")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_assignment",
			Message: "Assignment updated but failed to retrieve details",
		})
	}

	response := mapAssignmentToResponse(updatedAssignment)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
		"message": "Player assignment updated successfully",
	})
}

// AssignConfigToPlayer assigns a specific config to a player (simplified endpoint)
// POST /admin/player-assignments/:playerId/assign
func (h *AdminPlayerAssignmentHandler) AssignConfigToPlayer(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	playerID, err := uuid.Parse(c.Params("playerId"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid player ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_player_id",
			Message: "Invalid player ID",
		})
	}

	type AssignRequest struct {
		ConfigID   uuid.UUID  `json:"config_id" validate:"required"`
		GameMode   string     `json:"game_mode" validate:"required,oneof=base_game free_spins"`
		Reason     string     `json:"reason,omitempty"`
		AssignedBy string     `json:"assigned_by,omitempty"`
		ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	}

	var req AssignRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.reelStripService.AssignConfigToPlayer(
		c.Context(),
		playerID,
		req.ConfigID,
		req.GameMode,
		req.Reason,
		req.AssignedBy,
		req.ExpiresAt,
	); err != nil {
		log.Error().Err(err).Msg("Failed to assign config to player")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_assign",
			Message: "Failed to assign configuration to player",
		})
	}

	// Clear player assignment cache
	h.clearPlayerAssignmentCache(c, playerID)

	log.Info().
		Str("player_id", playerID.String()).
		Str("config_id", req.ConfigID.String()).
		Str("game_mode", req.GameMode).
		Msg("Assigned config to player")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Configuration assigned to player successfully",
	})
}

// RemoveAssignment removes a player's assignment
// DELETE /admin/player-assignments/:playerId
func (h *AdminPlayerAssignmentHandler) RemoveAssignment(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	playerID, err := uuid.Parse(c.Params("playerId"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid player ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_player_id",
			Message: "Invalid player ID",
		})
	}

	if err := h.reelStripService.RemovePlayerAssignment(c.Context(), playerID); err != nil {
		if err == reelstrip.ErrAssignmentNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "assignment_not_found",
				Message: "No assignment found for this player",
			})
		}
		log.Error().Err(err).Msg("Failed to remove assignment")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_remove_assignment",
			Message: "Failed to remove player assignment",
		})
	}

	// Clear player assignment cache
	h.clearPlayerAssignmentCache(c, playerID)

	log.Info().Str("player_id", playerID.String()).Msg("Removed player assignment")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Player assignment removed successfully",
	})
}

// clearPlayerAssignmentCache clears cache entries for a player's assignment
func (h *AdminPlayerAssignmentHandler) clearPlayerAssignmentCache(ctx *fiber.Ctx, playerID uuid.UUID) {
	log := h.logger.WithTrace(ctx)

	// Clear player assignment cache
	if err := h.cache.Expire(ctx.Context(), h.cache.PlayerAssignmentKey(playerID)); err != nil {
		log.Warn().Err(err).Str("player_id", playerID.String()).Msg("Failed to clear player assignment cache")
	}

	log.Debug().
		Str("player_id", playerID.String()).
		Msg("Cleared player assignment cache")
}

// Helper function to map domain model to DTO
func mapAssignmentToResponse(assignment *reelstrip.PlayerReelStripAssignment) dto.PlayerAssignmentResponse {
	return dto.PlayerAssignmentResponse{
		ID:                assignment.ID,
		PlayerID:          assignment.PlayerID,
		BaseGameConfigID:  assignment.BaseGameConfigID,
		FreeSpinsConfigID: assignment.FreeSpinsConfigID,
		AssignedAt:        assignment.AssignedAt,
		AssignedBy:        assignment.AssignedBy,
		Reason:            assignment.Reason,
		ExpiresAt:         assignment.ExpiresAt,
		IsActive:          assignment.IsActive,
	}
}
