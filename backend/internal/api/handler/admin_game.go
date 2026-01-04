package handler

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/game"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/infra/storage"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// AdminGameHandler handles admin game management endpoints
type AdminGameHandler struct {
	gameRepo game.Repository
	storage  storage.Storage
	logger   *logger.Logger
}

// NewAdminGameHandler creates a new admin game handler
func NewAdminGameHandler(
	gameRepo game.Repository,
	storage storage.Storage,
	log *logger.Logger,
) *AdminGameHandler {
	return &AdminGameHandler{
		gameRepo: gameRepo,
		storage:  storage,
		logger:   log,
	}
}

// ListGames lists all games
// GET /admin/games
func (h *AdminGameHandler) ListGames(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	games, total, err := h.gameRepo.ListGames(c.Context(), page, pageSize)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list games")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_list_games",
			Message: "Failed to list games",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"items":     games,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetGame gets a game by ID
// GET /admin/games/:id
func (h *AdminGameHandler) GetGame(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid game ID",
		})
	}

	g, err := h.gameRepo.GetGameByID(c.Context(), id)
	if err != nil {
		if err == game.ErrGameNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Game not found",
			})
		}
		log.Error().Err(err).Msg("Failed to get game")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_game",
			Message: "Failed to get game",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    g,
	})
}

// CreateGame creates a new game
// POST /admin/games
func (h *AdminGameHandler) CreateGame(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	var req dto.CreateGameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	g := &game.Game{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		DevURL:      req.DevURL,
		ProdURL:     req.ProdURL,
		IsActive:    req.IsActive,
	}

	if err := h.gameRepo.CreateGame(c.Context(), g); err != nil {
		log.Error().Err(err).Msg("Failed to create game")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_create_game",
			Message: "Failed to create game",
		})
	}

	log.Info().Str("game_id", g.ID.String()).Msg("Game created")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    g,
	})
}

// UpdateGame updates a game
// PUT /admin/games/:id
func (h *AdminGameHandler) UpdateGame(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid game ID",
		})
	}

	var req dto.UpdateGameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	g, err := h.gameRepo.UpdateGame(c.Context(), id, &game.GameUpdate{
		Name:        req.Name,
		Description: req.Description,
		DevURL:      req.DevURL,
		ProdURL:     req.ProdURL,
		IsActive:    req.IsActive,
	})
	if err != nil {
		if err == game.ErrGameNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Game not found",
			})
		}
		log.Error().Err(err).Msg("Failed to update game")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_update_game",
			Message: "Failed to update game",
		})
	}

	log.Info().Str("game_id", id.String()).Msg("Game updated")

	return c.JSON(fiber.Map{
		"success": true,
		"data":    g,
	})
}

// DeleteGame deletes a game
// DELETE /admin/games/:id
func (h *AdminGameHandler) DeleteGame(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid game ID",
		})
	}

	if err := h.gameRepo.DeleteGame(c.Context(), id); err != nil {
		if err == game.ErrGameNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Game not found",
			})
		}
		log.Error().Err(err).Msg("Failed to delete game")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_delete_game",
			Message: "Failed to delete game",
		})
	}

	log.Info().Str("game_id", id.String()).Msg("Game deleted")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Game deleted",
	})
}

// ActivateGame activates a game
// POST /admin/games/:id/activate
func (h *AdminGameHandler) ActivateGame(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid game ID",
		})
	}

	isActive := true
	g, err := h.gameRepo.UpdateGame(c.Context(), id, &game.GameUpdate{IsActive: &isActive})
	if err != nil {
		log.Error().Err(err).Msg("Failed to activate game")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_activate_game",
			Message: "Failed to activate game",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    g,
	})
}

// DeactivateGame deactivates a game
// POST /admin/games/:id/deactivate
func (h *AdminGameHandler) DeactivateGame(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid game ID",
		})
	}

	isActive := false
	g, err := h.gameRepo.UpdateGame(c.Context(), id, &game.GameUpdate{IsActive: &isActive})
	if err != nil {
		log.Error().Err(err).Msg("Failed to deactivate game")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_deactivate_game",
			Message: "Failed to deactivate game",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    g,
	})
}

// ListAssets lists all assets
// GET /admin/assets
func (h *AdminGameHandler) ListAssets(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	assets, total, err := h.gameRepo.ListAssets(c.Context(), page, pageSize)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list assets")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_list_assets",
			Message: "Failed to list assets",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"items":     assets,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetAsset gets an asset by ID
// GET /admin/assets/:id
func (h *AdminGameHandler) GetAsset(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid asset ID",
		})
	}

	a, err := h.gameRepo.GetAssetByID(c.Context(), id)
	if err != nil {
		if err == game.ErrAssetNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Asset not found",
			})
		}
		log.Error().Err(err).Msg("Failed to get asset")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_asset",
			Message: "Failed to get asset",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    a,
	})
}

// CreateAsset creates a new asset
// POST /admin/assets
func (h *AdminGameHandler) CreateAsset(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	var req dto.CreateAssetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	spritesheetJSON, err := json.Marshal(req.SpritesheetJSON)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_spritesheet_json",
			Message: "Invalid spritesheet JSON",
		})
	}

	imagesJSON, err := json.Marshal(req.Images)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_images_json",
			Message: "Invalid images JSON",
		})
	}

	audiosJSON, err := json.Marshal(req.Audios)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_audios_json",
			Message: "Invalid audios JSON",
		})
	}

	videosJSON, err := json.Marshal(req.Videos)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_videos_json",
			Message: "Invalid videos JSON",
		})
	}

	// Generate ObjectName from the asset name if not provided
	objectName := req.ObjectName
	if objectName == "" {
		objectName = game.GenerateObjectName(req.Name)
	}

	// Create folder on storage
	if err := h.storage.CreateFolder(c.Context(), objectName); err != nil {
		log.Error().Err(err).Str("object_name", objectName).Msg("Failed to create storage folder")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_create_storage_folder",
			Message: "Failed to create storage folder",
		})
	}

	// Get base URL from storage (set once at creation time)
	baseURL := h.storage.GetBaseURL(objectName)

	a := &game.Asset{
		ID:              uuid.New(),
		Name:            req.Name,
		Description:     req.Description,
		ObjectName:      objectName,
		BaseURL:         baseURL,
		SpritesheetJSON: spritesheetJSON,
		Images:          imagesJSON,
		Audios:          audiosJSON,
		Videos:          videosJSON,
		IsActive:        req.IsActive,
	}

	if err := h.gameRepo.CreateAsset(c.Context(), a); err != nil {
		log.Error().Err(err).Msg("Failed to create asset")
		// Try to cleanup the folder we just created
		if delErr := h.storage.DeleteTheme(c.Context(), objectName); delErr != nil {
			log.Error().Err(delErr).Str("object_name", objectName).Msg("Failed to cleanup storage folder after asset creation failed")
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_create_asset",
			Message: "Failed to create asset",
		})
	}

	log.Info().Str("asset_id", a.ID.String()).Str("object_name", objectName).Str("base_url", baseURL).Msg("Asset created")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    a,
	})
}

// UpdateAsset updates an asset
// PUT /admin/assets/:id
func (h *AdminGameHandler) UpdateAsset(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid asset ID",
		})
	}

	var req dto.UpdateAssetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Get current asset to check if object_name is changing
	currentAsset, err := h.gameRepo.GetAssetByID(c.Context(), id)
	if err != nil {
		if err == game.ErrAssetNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Asset not found",
			})
		}
		log.Error().Err(err).Msg("Failed to get current asset")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_asset",
			Message: "Failed to get asset",
		})
	}

	update := &game.AssetUpdate{
		Name:        req.Name,
		Description: req.Description,
		ObjectName:  req.ObjectName,
		IsActive:    req.IsActive,
	}

	// If object_name is being changed, rename the folder on storage
	if req.ObjectName != nil && *req.ObjectName != currentAsset.ObjectName {
		newObjectName := *req.ObjectName

		// Rename folder on storage
		if err := h.storage.RenameFolder(c.Context(), currentAsset.ObjectName, newObjectName); err != nil {
			log.Error().Err(err).
				Str("old_object_name", currentAsset.ObjectName).
				Str("new_object_name", newObjectName).
				Msg("Failed to rename storage folder")
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "failed_to_rename_storage_folder",
				Message: "Failed to rename storage folder",
			})
		}

		// Update base_url to reflect new object_name
		newBaseURL := h.storage.GetBaseURL(newObjectName)
		update.BaseURL = &newBaseURL

		log.Info().
			Str("old_object_name", currentAsset.ObjectName).
			Str("new_object_name", newObjectName).
			Str("new_base_url", newBaseURL).
			Msg("Storage folder renamed")
	}

	if req.SpritesheetJSON != nil {
		spritesheetJSON, err := json.Marshal(req.SpritesheetJSON)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_spritesheet_json",
				Message: "Invalid spritesheet JSON",
			})
		}
		update.SpritesheetJSON = spritesheetJSON
	}

	if req.Images != nil {
		imagesJSON, err := json.Marshal(req.Images)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_images_json",
				Message: "Invalid images JSON",
			})
		}
		update.Images = imagesJSON
	}

	if req.Audios != nil {
		audiosJSON, err := json.Marshal(req.Audios)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_audios_json",
				Message: "Invalid audios JSON",
			})
		}
		update.Audios = audiosJSON
	}

	if req.Videos != nil {
		videosJSON, err := json.Marshal(req.Videos)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_videos_json",
				Message: "Invalid videos JSON",
			})
		}
		update.Videos = videosJSON
	}

	a, err := h.gameRepo.UpdateAsset(c.Context(), id, update)
	if err != nil {
		if err == game.ErrAssetNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Asset not found",
			})
		}
		log.Error().Err(err).Msg("Failed to update asset")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_update_asset",
			Message: "Failed to update asset",
		})
	}

	log.Info().Str("asset_id", id.String()).Msg("Asset updated")

	return c.JSON(fiber.Map{
		"success": true,
		"data":    a,
	})
}

// DeleteAsset deletes an asset
// DELETE /admin/assets/:id
func (h *AdminGameHandler) DeleteAsset(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid asset ID",
		})
	}

	if err := h.gameRepo.DeleteAsset(c.Context(), id); err != nil {
		if err == game.ErrAssetNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Asset not found",
			})
		}
		log.Error().Err(err).Msg("Failed to delete asset")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_delete_asset",
			Message: "Failed to delete asset",
		})
	}

	log.Info().Str("asset_id", id.String()).Msg("Asset deleted")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Asset deleted",
	})
}

// ActivateAsset activates an asset
// POST /admin/assets/:id/activate
func (h *AdminGameHandler) ActivateAsset(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid asset ID",
		})
	}

	isActive := true
	a, err := h.gameRepo.UpdateAsset(c.Context(), id, &game.AssetUpdate{IsActive: &isActive})
	if err != nil {
		log.Error().Err(err).Msg("Failed to activate asset")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_activate_asset",
			Message: "Failed to activate asset",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    a,
	})
}

// DeactivateAsset deactivates an asset
// POST /admin/assets/:id/deactivate
func (h *AdminGameHandler) DeactivateAsset(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid asset ID",
		})
	}

	isActive := false
	a, err := h.gameRepo.UpdateAsset(c.Context(), id, &game.AssetUpdate{IsActive: &isActive})
	if err != nil {
		log.Error().Err(err).Msg("Failed to deactivate asset")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_deactivate_asset",
			Message: "Failed to deactivate asset",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    a,
	})
}

// ListGameConfigs lists all game configs
// GET /admin/game-configs
func (h *AdminGameHandler) ListGameConfigs(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)

	configs, total, err := h.gameRepo.ListGameConfigs(c.Context(), page, pageSize)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list game configs")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_list_game_configs",
			Message: "Failed to list game configs",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"items":     configs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetGameConfig gets a game config by ID
// GET /admin/game-configs/:id
func (h *AdminGameHandler) GetGameConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid game config ID",
		})
	}

	config, err := h.gameRepo.GetGameConfigByID(c.Context(), id)
	if err != nil {
		if err == game.ErrGameConfigNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Game config not found",
			})
		}
		log.Error().Err(err).Msg("Failed to get game config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_game_config",
			Message: "Failed to get game config",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// CreateGameConfig creates a new game config
// POST /admin/game-configs
func (h *AdminGameHandler) CreateGameConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	var req dto.CreateGameConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	gameID, err := uuid.Parse(req.GameID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_game_id",
			Message: "Invalid game ID",
		})
	}

	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_asset_id",
			Message: "Invalid asset ID",
		})
	}

	config := &game.GameConfig{
		ID:       uuid.New(),
		GameID:   gameID,
		AssetID:  assetID,
		IsActive: req.IsActive,
	}

	if err := h.gameRepo.CreateGameConfig(c.Context(), config); err != nil {
		log.Error().Err(err).Msg("Failed to create game config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_create_game_config",
			Message: "Failed to create game config",
		})
	}

	log.Info().Str("config_id", config.ID.String()).Msg("Game config created")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// DeleteGameConfig deletes a game config
// DELETE /admin/game-configs/:id
func (h *AdminGameHandler) DeleteGameConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid game config ID",
		})
	}

	if err := h.gameRepo.DeleteGameConfig(c.Context(), id); err != nil {
		if err == game.ErrGameConfigNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "not_found",
				Message: "Game config not found",
			})
		}
		log.Error().Err(err).Msg("Failed to delete game config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_delete_game_config",
			Message: "Failed to delete game config",
		})
	}

	log.Info().Str("config_id", id.String()).Msg("Game config deleted")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Game config deleted",
	})
}

// ActivateGameConfig activates a game config
// POST /admin/game-configs/:id/activate
func (h *AdminGameHandler) ActivateGameConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid game config ID",
		})
	}

	config, err := h.gameRepo.ActivateGameConfig(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to activate game config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_activate_game_config",
			Message: "Failed to activate game config",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

// DeactivateGameConfig deactivates a game config
// POST /admin/game-configs/:id/deactivate
func (h *AdminGameHandler) DeactivateGameConfig(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid game config ID",
		})
	}

	config, err := h.gameRepo.DeactivateGameConfig(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to deactivate game config")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_deactivate_game_config",
			Message: "Failed to deactivate game config",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}
