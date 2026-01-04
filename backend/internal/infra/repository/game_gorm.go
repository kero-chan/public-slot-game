package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/game"
	"gorm.io/gorm"
)

// GameGormRepository implements game.Repository using GORM
type GameGormRepository struct {
	db *gorm.DB
}

// NewGameGormRepository creates a new GORM game repository
func NewGameGormRepository(db *gorm.DB) game.Repository {
	return &GameGormRepository{
		db: db,
	}
}

// ============== Game Methods ==============

// GetGameByID retrieves a game by ID (includes inactive for admin)
func (r *GameGormRepository) GetGameByID(ctx context.Context, id uuid.UUID) (*game.Game, error) {
	var g game.Game
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&g).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, game.ErrGameNotFound
		}
		return nil, fmt.Errorf("failed to get game by ID: %w", err)
	}
	return &g, nil
}

// GetGamesByIDs retrieves multiple games by their IDs in a single query (batch fetch)
func (r *GameGormRepository) GetGamesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*game.Game, error) {
	result := make(map[uuid.UUID]*game.Game)
	if len(ids) == 0 {
		return result, nil
	}

	var games []*game.Game
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&games).Error; err != nil {
		return nil, fmt.Errorf("failed to get games by IDs: %w", err)
	}

	for _, g := range games {
		result[g.ID] = g
	}

	return result, nil
}

// ListGames lists all games with pagination
func (r *GameGormRepository) ListGames(ctx context.Context, page, pageSize int) ([]*game.Game, int64, error) {
	var games []*game.Game
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&game.Game{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count games: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&games).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list games: %w", err)
	}

	return games, total, nil
}

// CreateGame creates a new game
func (r *GameGormRepository) CreateGame(ctx context.Context, g *game.Game) error {
	g.CreatedAt = time.Now()
	g.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}
	return nil
}

// UpdateGame updates a game
func (r *GameGormRepository) UpdateGame(ctx context.Context, id uuid.UUID, update *game.GameUpdate) (*game.Game, error) {
	var g game.Game
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&g).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, game.ErrGameNotFound
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if update.Name != nil {
		updates["name"] = *update.Name
	}
	if update.Description != nil {
		updates["description"] = update.Description
	}
	if update.DevURL != nil {
		updates["dev_url"] = update.DevURL
	}
	if update.ProdURL != nil {
		updates["prod_url"] = update.ProdURL
	}
	if update.IsActive != nil {
		updates["is_active"] = *update.IsActive
	}

	if err := r.db.WithContext(ctx).Model(&g).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	// Reload the game
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&g).Error; err != nil {
		return nil, fmt.Errorf("failed to reload game: %w", err)
	}

	return &g, nil
}

// DeleteGame deletes a game
func (r *GameGormRepository) DeleteGame(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&game.Game{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete game: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return game.ErrGameNotFound
	}
	return nil
}

// ============== Asset Methods ==============

// GetAssetByID retrieves an asset by ID (includes inactive for admin)
func (r *GameGormRepository) GetAssetByID(ctx context.Context, id uuid.UUID) (*game.Asset, error) {
	var a game.Asset
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, game.ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to get asset by ID: %w", err)
	}
	return &a, nil
}

// ListAssets lists all assets with pagination
func (r *GameGormRepository) ListAssets(ctx context.Context, page, pageSize int) ([]*game.Asset, int64, error) {
	var assets []*game.Asset
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&game.Asset{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count assets: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&assets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list assets: %w", err)
	}

	return assets, total, nil
}

// CreateAsset creates a new asset
func (r *GameGormRepository) CreateAsset(ctx context.Context, a *game.Asset) error {
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return fmt.Errorf("failed to create asset: %w", err)
	}
	return nil
}

// UpdateAsset updates an asset
func (r *GameGormRepository) UpdateAsset(ctx context.Context, id uuid.UUID, update *game.AssetUpdate) (*game.Asset, error) {
	var a game.Asset
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, game.ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if update.Name != nil {
		updates["name"] = *update.Name
	}
	if update.Description != nil {
		updates["description"] = update.Description
	}
	if update.ObjectName != nil {
		updates["object_name"] = *update.ObjectName
	}
	if update.BaseURL != nil {
		updates["base_url"] = *update.BaseURL
	}
	if update.SpritesheetJSON != nil {
		updates["spritesheet_json"] = update.SpritesheetJSON
	}
	if update.Images != nil {
		updates["images"] = update.Images
	}
	if update.Audios != nil {
		updates["audios"] = update.Audios
	}
	if update.Videos != nil {
		updates["videos"] = update.Videos
	}
	if update.IsActive != nil {
		updates["is_active"] = *update.IsActive
	}

	if err := r.db.WithContext(ctx).Model(&a).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update asset: %w", err)
	}

	// Reload the asset
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		return nil, fmt.Errorf("failed to reload asset: %w", err)
	}

	return &a, nil
}

// DeleteAsset deletes an asset
func (r *GameGormRepository) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&game.Asset{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete asset: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return game.ErrAssetNotFound
	}
	return nil
}

// ============== GameConfig Methods ==============

// GetActiveAssetForGame retrieves the active asset configuration for a game
func (r *GameGormRepository) GetActiveAssetForGame(ctx context.Context, gameID uuid.UUID) (*game.Asset, error) {
	// First check if the game exists
	var g game.Game
	if err := r.db.WithContext(ctx).Where("id = ? AND is_active = true", gameID).First(&g).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, game.ErrGameNotFound
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Get the active config with its asset
	var config game.GameConfig
	if err := r.db.WithContext(ctx).
		Preload("Asset").
		Where("game_id = ? AND is_active = true", gameID).
		First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, game.ErrNoActiveConfig
		}
		return nil, fmt.Errorf("failed to get game config: %w", err)
	}

	if config.Asset == nil || !config.Asset.IsActive {
		return nil, game.ErrNoActiveConfig
	}

	return config.Asset, nil
}

// GetGameConfigByID retrieves a game config by ID
func (r *GameGormRepository) GetGameConfigByID(ctx context.Context, id uuid.UUID) (*game.GameConfig, error) {
	var config game.GameConfig
	if err := r.db.WithContext(ctx).
		Preload("Game").
		Preload("Asset").
		Where("id = ?", id).
		First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, game.ErrGameConfigNotFound
		}
		return nil, fmt.Errorf("failed to get game config by ID: %w", err)
	}
	return &config, nil
}

// ListGameConfigs lists all game configs with pagination
func (r *GameGormRepository) ListGameConfigs(ctx context.Context, page, pageSize int) ([]*game.GameConfig, int64, error) {
	var configs []*game.GameConfig
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.WithContext(ctx).Model(&game.GameConfig{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count game configs: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Preload("Game").
		Preload("Asset").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list game configs: %w", err)
	}

	return configs, total, nil
}

// CreateGameConfig creates a new game config
func (r *GameGormRepository) CreateGameConfig(ctx context.Context, c *game.GameConfig) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return fmt.Errorf("failed to create game config: %w", err)
	}
	return nil
}

// DeleteGameConfig deletes a game config
func (r *GameGormRepository) DeleteGameConfig(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&game.GameConfig{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete game config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return game.ErrGameConfigNotFound
	}
	return nil
}

// ActivateGameConfig activates a game config (and deactivates others for the same game)
func (r *GameGormRepository) ActivateGameConfig(ctx context.Context, id uuid.UUID) (*game.GameConfig, error) {
	var config game.GameConfig
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, game.ErrGameConfigNotFound
		}
		return nil, fmt.Errorf("failed to get game config: %w", err)
	}

	// Deactivate all other configs for the same game
	if err := r.db.WithContext(ctx).
		Model(&game.GameConfig{}).
		Where("game_id = ? AND id != ?", config.GameID, id).
		Update("is_active", false).Error; err != nil {
		return nil, fmt.Errorf("failed to deactivate other configs: %w", err)
	}

	// Activate this config
	if err := r.db.WithContext(ctx).
		Model(&config).
		Updates(map[string]interface{}{
			"is_active":  true,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return nil, fmt.Errorf("failed to activate config: %w", err)
	}

	// Reload with associations
	if err := r.db.WithContext(ctx).
		Preload("Game").
		Preload("Asset").
		Where("id = ?", id).
		First(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to reload config: %w", err)
	}

	return &config, nil
}

// DeactivateGameConfig deactivates a game config
func (r *GameGormRepository) DeactivateGameConfig(ctx context.Context, id uuid.UUID) (*game.GameConfig, error) {
	var config game.GameConfig
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, game.ErrGameConfigNotFound
		}
		return nil, fmt.Errorf("failed to get game config: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Model(&config).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return nil, fmt.Errorf("failed to deactivate config: %w", err)
	}

	// Reload with associations
	if err := r.db.WithContext(ctx).
		Preload("Game").
		Preload("Asset").
		Where("id = ?", id).
		First(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to reload config: %w", err)
	}

	return &config, nil
}
