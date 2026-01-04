package handler

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/game"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// GameHandler handles game-related endpoints
type GameHandler struct {
	gameRepo game.Repository
	logger   *logger.Logger
}

// NewGameHandler creates a new game handler
func NewGameHandler(
	gameRepo game.Repository,
	log *logger.Logger,
) *GameHandler {
	return &GameHandler{
		gameRepo: gameRepo,
		logger:   log,
	}
}

// GetGameAssets retrieves the assets for a game
func (h *GameHandler) GetGameAssets(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse game ID from URL parameter
	gameIDStr := c.Get("x-game-id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		log.Warn().Err(err).Str("game_id", gameIDStr).Msg("Invalid game ID format")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_game_id",
			Message: "Invalid game ID format",
		})
	}

	log.Debug().Str("game_id", gameID.String()).Msg("Fetching game assets")

	// Get the game first to retrieve its name
	gameEntity, err := h.gameRepo.GetGameByID(c.Context(), gameID)
	if err != nil {
		if errors.Is(err, game.ErrGameNotFound) {
			log.Warn().Str("game_id", gameID.String()).Msg("Game not found")
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "game_not_found",
				Message: "Game not found",
			})
		}
		log.Error().Err(err).Str("game_id", gameID.String()).Msg("Failed to get game")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve game",
		})
	}

	// Get active asset for the game
	asset, err := h.gameRepo.GetActiveAssetForGame(c.Context(), gameID)
	if err != nil {
		if errors.Is(err, game.ErrGameNotFound) {
			log.Warn().Str("game_id", gameID.String()).Msg("Game not found")
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "game_not_found",
				Message: "Game not found",
			})
		}
		if errors.Is(err, game.ErrNoActiveConfig) {
			log.Warn().Str("game_id", gameID.String()).Msg("No active asset configuration")
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "no_active_config",
				Message: "No active asset configuration for this game",
			})
		}
		log.Error().Err(err).Str("game_id", gameID.String()).Msg("Failed to get game assets")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve game assets",
		})
	}

	// Parse the images JSON and build full URLs
	var images map[string]string
	if err := json.Unmarshal(asset.Images, &images); err != nil {
		log.Error().Err(err).Msg("Failed to parse images JSON")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to process asset images",
		})
	}

	// Parse audios JSON
	var audios map[string]any
	if asset.Audios != nil && len(asset.Audios) > 0 {
		if err := json.Unmarshal(asset.Audios, &audios); err != nil {
			log.Warn().Err(err).Msg("Failed to parse audios JSON, using empty map")
			audios = make(map[string]any)
		}
	} else {
		audios = make(map[string]any)
	}

	// Parse videos JSON (supports both string and nested {win: [], loop: []} structure)
	var videos map[string]any
	if asset.Videos != nil && len(asset.Videos) > 0 {
		if err := json.Unmarshal(asset.Videos, &videos); err != nil {
			log.Warn().Err(err).Msg("Failed to parse videos JSON, using empty map")
			videos = make(map[string]any)
		}
	} else {
		videos = make(map[string]any)
	}

	// Build full URLs using stored base_url (set once at asset creation)
	baseURL := strings.TrimSuffix(asset.BaseURL, "/")

	// Build full URLs for images
	fullURLImages := make(map[string]string)
	for key, path := range images {
		fullURLImages[key] = baseURL + "/" + path
	}

	// Build full URLs for audios (handle both string and []string values)
	fullURLAudios := make(map[string]any)
	for key, value := range audios {
		switch v := value.(type) {
		case string:
			fullURLAudios[key] = baseURL + "/" + v
		case []interface{}:
			urls := make([]string, len(v))
			for i, item := range v {
				if s, ok := item.(string); ok {
					urls[i] = baseURL + "/" + s
				}
			}
			fullURLAudios[key] = urls
		default:
			fullURLAudios[key] = value
		}
	}

	// Build full URLs for videos/animations
	// Supports multiple formats:
	// 1. Simple string: "default": "videos/default.mp4"
	// 2. Spritesheet object: "default": { "png": "path.png", "json": "path.json" }
	// 3. Symbol animations: "fa": { "win": [{ "png": "...", "json": "..." }], "loop": [...] }
	fullURLVideos := make(map[string]any)
	for key, value := range videos {
		switch v := value.(type) {
		case string:
			// Simple string path (e.g., "default": "videos/default.mp4")
			fullURLVideos[key] = baseURL + "/" + v
		case map[string]interface{}:
			// Check if it's a spritesheet object with png/json keys
			if pngPath, hasPng := v["png"].(string); hasPng {
				if jsonPath, hasJson := v["json"].(string); hasJson {
					// Spritesheet animation object: { png: "...", json: "..." }
					fullURLVideos[key] = map[string]string{
						"png":  baseURL + "/" + pngPath,
						"json": baseURL + "/" + jsonPath,
					}
					continue
				}
			}

			// Symbol animations structure: { win: [...], loop: [...] }
			symbolAnims := make(map[string]any)
			for subKey, subValue := range v {
				if paths, ok := subValue.([]interface{}); ok {
					// Check if array contains objects (spritesheet) or strings
					if len(paths) > 0 {
						if _, isObj := paths[0].(map[string]interface{}); isObj {
							// Array of spritesheet objects
							spritesheets := make([]map[string]string, 0, len(paths))
							for _, item := range paths {
								if animObj, ok := item.(map[string]interface{}); ok {
									spritesheet := make(map[string]string)
									if png, ok := animObj["png"].(string); ok {
										spritesheet["png"] = baseURL + "/" + png
									}
									if jsonPath, ok := animObj["json"].(string); ok {
										spritesheet["json"] = baseURL + "/" + jsonPath
									}
									spritesheets = append(spritesheets, spritesheet)
								}
							}
							symbolAnims[subKey] = spritesheets
						} else {
							// Array of string paths (legacy format)
							urls := make([]string, 0, len(paths))
							for _, item := range paths {
								if s, ok := item.(string); ok {
									urls = append(urls, baseURL+"/"+s)
								}
							}
							symbolAnims[subKey] = urls
						}
					} else {
						symbolAnims[subKey] = []string{}
					}
				}
			}
			fullURLVideos[key] = symbolAnims
		default:
			fullURLVideos[key] = value
		}
	}

	log.Info().
		Str("game_id", gameID.String()).
		Str("game_name", gameEntity.Name).
		Str("asset_id", asset.ID.String()).
		Str("asset_name", asset.Name).
		Msg("Game assets retrieved successfully")

	// Return the response with game name (not asset name)
	response := game.GameAssetsResponse{
		ID:              asset.ID,
		Name:            gameEntity.Name,
		SpritesheetJSON: asset.SpritesheetJSON,
		Images:          fullURLImages,
		Audios:          fullURLAudios,
		Videos:          fullURLVideos,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
