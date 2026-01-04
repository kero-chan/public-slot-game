package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/internal/pkg/cache"
	"gorm.io/gorm"
)

// ReelStripGormRepository implements reelstrip.Repository using GORM
type ReelStripGormRepository struct {
	cache *cache.Cache
	db    *gorm.DB
}

// NewReelStripGormRepository creates a new GORM reel strip repository
func NewReelStripGormRepository(db *gorm.DB, cache *cache.Cache) reelstrip.Repository {
	return &ReelStripGormRepository{
		db:    db,
		cache: cache,
	}
}

// Create inserts a new reel strip into the database
func (r *ReelStripGormRepository) Create(ctx context.Context, strip *reelstrip.ReelStrip) error {
	if err := r.db.WithContext(ctx).Create(strip).Error; err != nil {
		return fmt.Errorf("failed to create reel strip: %w", err)
	}
	return nil
}

// CreateBatch inserts multiple reel strips in a single transaction
func (r *ReelStripGormRepository) CreateBatch(ctx context.Context, strips []*reelstrip.ReelStrip) error {
	if len(strips) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(strips, 100).Error; err != nil {
			return fmt.Errorf("failed to create reel strips in batch: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a reel strip by its ID
func (r *ReelStripGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*reelstrip.ReelStrip, error) {
	var strip reelstrip.ReelStrip
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&strip).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, reelstrip.ErrReelStripNotFound
		}
		return nil, fmt.Errorf("failed to get reel strip by ID: %w", err)
	}
	return &strip, nil
}

// GetAllActive retrieves all active reel strips for a game mode
func (r *ReelStripGormRepository) GetAllActive(ctx context.Context, gameMode string) ([]*reelstrip.ReelStrip, error) {
	var strips []*reelstrip.ReelStrip

	err := r.db.WithContext(ctx).
		Where("game_mode = ? AND is_active = ?", gameMode, true).
		Order("reel_number ASC, created_at DESC").
		Find(&strips).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get all active reel strips: %w", err)
	}

	return strips, nil
}

// GetByGameModeAndReel retrieves all active reel strips for a specific game mode and reel number
func (r *ReelStripGormRepository) GetByGameModeAndReel(ctx context.Context, gameMode string, reelNumber int) ([]*reelstrip.ReelStrip, error) {
	if reelNumber < 0 || reelNumber > 4 {
		return nil, reelstrip.ErrInvalidReelNumber
	}

	var strips []*reelstrip.ReelStrip

	err := r.db.WithContext(ctx).
		Where("game_mode = ? AND reel_number = ? AND is_active = ?", gameMode, reelNumber, true).
		Order("created_at DESC").
		Find(&strips).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get reel strips: %w", err)
	}

	return strips, nil
}

// CountActive returns the count of active reel strips for each reel in a game mode
func (r *ReelStripGormRepository) CountActive(ctx context.Context, gameMode string) (map[int]int, error) {
	type Result struct {
		ReelNumber int
		Count      int64
	}

	var results []Result
	err := r.db.WithContext(ctx).
		Model(&reelstrip.ReelStrip{}).
		Select("reel_number, COUNT(*) as count").
		Where("game_mode = ? AND is_active = ?", gameMode, true).
		Group("reel_number").
		Order("reel_number").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count active strips: %w", err)
	}

	countMap := make(map[int]int)
	for _, result := range results {
		countMap[result.ReelNumber] = int(result.Count)
	}

	return countMap, nil
}

// DeactivateOldVersions is deprecated - version control is now handled at config level
// This method is kept for backward compatibility but does nothing
func (r *ReelStripGormRepository) DeactivateOldVersions(ctx context.Context, gameMode string, keepVersion int) error {
	// No-op: version control is now managed at ReelStripConfig level
	return nil
}

// Update updates a reel strip
func (r *ReelStripGormRepository) Update(ctx context.Context, strip *reelstrip.ReelStrip) error {
	result := r.db.WithContext(ctx).Save(strip)
	if result.Error != nil {
		return fmt.Errorf("failed to update reel strip: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return reelstrip.ErrReelStripNotFound
	}
	return nil
}

// Delete soft deletes a reel strip by marking it as inactive
func (r *ReelStripGormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&reelstrip.ReelStrip{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to delete reel strip: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return reelstrip.ErrReelStripNotFound
	}
	return nil
}

// GetByIDs retrieves multiple reel strips by their IDs
func (r *ReelStripGormRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*reelstrip.ReelStrip, error) {
	if len(ids) == 0 {
		return []*reelstrip.ReelStrip{}, nil
	}

	var strips []*reelstrip.ReelStrip
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&strips).Error; err != nil {
		return nil, fmt.Errorf("failed to get reel strips by IDs: %w", err)
	}

	return strips, nil
}

// ===== ReelStripConfig Methods =====

// CreateConfig inserts a new reel strip configuration
func (r *ReelStripGormRepository) CreateConfig(ctx context.Context, config *reelstrip.ReelStripConfig) error {
	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		return fmt.Errorf("failed to create reel strip config: %w", err)
	}
	return nil
}

// GetConfigByID retrieves a reel strip configuration by ID
func (r *ReelStripGormRepository) GetConfigByID(ctx context.Context, id uuid.UUID) (*reelstrip.ReelStripConfig, error) {
	var config reelstrip.ReelStripConfig
	cacheId := r.cache.ReelStripConfigById(id)
	res, err := r.cache.GetWithSingleflight(ctx, cacheId, &reelstrip.ReelStripConfig{}, func() (any, error) {
		if err := r.db.WithContext(ctx).Where("id = ?", id).First(&config).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, reelstrip.ErrConfigNotFound
			}
			return nil, fmt.Errorf("failed to get config by ID: %w", err)
		}
		return &config, nil
	})

	if err != nil {
		return nil, err
	}
	return res.(*reelstrip.ReelStripConfig), nil
}

// GetConfigByName retrieves a reel strip configuration by name
func (r *ReelStripGormRepository) GetConfigByName(ctx context.Context, name string) (*reelstrip.ReelStripConfig, error) {
	var config reelstrip.ReelStripConfig
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, reelstrip.ErrConfigNotFound
		}
		return nil, fmt.Errorf("failed to get config by name: %w", err)
	}
	return &config, nil
}

// GetDefaultConfig retrieves the default configuration for a game mode
func (r *ReelStripGormRepository) GetDefaultConfig(ctx context.Context, gameMode string) (*reelstrip.ReelStripConfig, error) {
	var config reelstrip.ReelStripConfig
	key := r.cache.DefaultReelStripConfig(gameMode)
	res, err := r.cache.GetWithSingleflight(ctx, key, &reelstrip.ReelStripConfig{}, func() (any, error) {
		if err := r.db.WithContext(ctx).
			Where("(game_mode = ? OR game_mode = ?) AND is_default = ? AND is_active = ?", gameMode, string(reelstrip.Both), true, true).
			Order(gorm.Expr("CASE game_mode WHEN ? THEN 0 WHEN ? THEN 1 END", gameMode, string(reelstrip.Both))).
			First(&config).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, reelstrip.ErrNoDefaultConfig
			}
			return nil, fmt.Errorf("failed to get default config: %w", err)
		}

		cacheId := r.cache.ReelStripConfigById(config.ID)
		r.cache.Set(ctx, cacheId, &config, 0)
		return &config, nil
	})

	if err != nil {
		return nil, err
	}
	return res.(*reelstrip.ReelStripConfig), nil
}

// ListConfigs retrieves reel strip configs with filtering and pagination
func (r *ReelStripGormRepository) ListConfigs(ctx context.Context, filters *reelstrip.ConfigListFilters) ([]*reelstrip.ReelStripConfig, int64, error) {
	query := r.db.WithContext(ctx).Model(&reelstrip.ReelStripConfig{})

	// Apply filters
	if filters.GameMode != nil && *filters.GameMode != "" {
		query = query.Where("game_mode = ?", *filters.GameMode)
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.IsDefault != nil {
		query = query.Where("is_default = ?", *filters.IsDefault)
	}
	if filters.Name != nil && *filters.Name != "" {
		query = query.Where("name ILIKE ?", "%"+*filters.Name+"%")
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count configs: %w", err)
	}

	// Apply pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 {
		filters.Limit = 20 // Default limit
	}
	offset := (filters.Page - 1) * filters.Limit

	// Fetch paginated results
	var configs []*reelstrip.ReelStripConfig
	if err := query.
		Order("created_at DESC").
		Limit(filters.Limit).
		Offset(offset).
		Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list configs: %w", err)
	}

	return configs, total, nil
}

// UpdateConfig updates a reel strip configuration
func (r *ReelStripGormRepository) UpdateConfig(ctx context.Context, config *reelstrip.ReelStripConfig) error {
	result := r.db.WithContext(ctx).Save(config)
	if result.Error != nil {
		return fmt.Errorf("failed to update config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return reelstrip.ErrConfigNotFound
	}
	return nil
}

// DeleteConfig soft deletes a configuration by marking it as inactive
func (r *ReelStripGormRepository) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&reelstrip.ReelStripConfig{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to delete config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return reelstrip.ErrConfigNotFound
	}
	return nil
}

// SetDefaultConfig sets a configuration as the default for its game mode
// This will unset any existing default for that game mode
func (r *ReelStripGormRepository) SetDefaultConfig(ctx context.Context, id uuid.UUID, gameMode string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, unset any existing default for this game mode
		if err := tx.Model(&reelstrip.ReelStripConfig{}).
			Where("game_mode = ? AND is_default = ?", gameMode, true).
			Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset existing default: %w", err)
		}

		// Set the new default
		result := tx.Model(&reelstrip.ReelStripConfig{}).
			Where("id = ? AND game_mode = ?", id, gameMode).
			Update("is_default", true)

		if result.Error != nil {
			return fmt.Errorf("failed to set default config: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return reelstrip.ErrConfigNotFound
		}

		return nil
	})
}

// GetSetByConfigID retrieves a complete reel strip set by configuration ID
func (r *ReelStripGormRepository) GetSetByConfigID(ctx context.Context, configID uuid.UUID) (*reelstrip.ReelStripConfigSet, error) {
	// Get the configuration
	cacheKey := r.cache.ReelStripConfigSetKey(configID)
	res, err := r.cache.GetWithSingleflight(ctx, cacheKey, &reelstrip.ReelStripConfigSet{}, func() (any, error) {
		config, err := r.GetConfigByID(ctx, configID)
		if err != nil {
			return nil, err
		}

		// Collect all reel strip IDs
		stripIDs := []uuid.UUID{
			config.Reel0StripID,
			config.Reel1StripID,
			config.Reel2StripID,
			config.Reel3StripID,
			config.Reel4StripID,
		}

		// Fetch all strips
		strips, err := r.GetByIDs(ctx, stripIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get strips for config: %w", err)
		}

		// Map strips by ID for quick lookup
		stripMap := make(map[uuid.UUID]*reelstrip.ReelStrip)
		for _, strip := range strips {
			stripMap[strip.ID] = strip
		}

		// Build the config set in correct order
		configSet := &reelstrip.ReelStripConfigSet{
			Config: config,
		}

		configSet.Strips[0] = stripMap[config.Reel0StripID]
		configSet.Strips[1] = stripMap[config.Reel1StripID]
		configSet.Strips[2] = stripMap[config.Reel2StripID]
		configSet.Strips[3] = stripMap[config.Reel3StripID]
		configSet.Strips[4] = stripMap[config.Reel4StripID]

		// Verify completeness
		if !configSet.IsComplete() {
			return nil, reelstrip.ErrIncompleteSet
		}

		return configSet, nil
	})

	if err != nil {
		return nil, err
	}

	return res.(*reelstrip.ReelStripConfigSet), nil
}

// ===== PlayerReelStripAssignment Methods =====

// CreateAssignment creates a new player reel strip assignment
func (r *ReelStripGormRepository) CreateAssignment(ctx context.Context, assignment *reelstrip.PlayerReelStripAssignment) error {
	if err := r.db.WithContext(ctx).Create(assignment).Error; err != nil {
		return fmt.Errorf("failed to create player assignment: %w", err)
	}
	return nil
}

// GetPlayerAssignment retrieves a player's active assignment
func (r *ReelStripGormRepository) GetPlayerAssignment(ctx context.Context, playerID uuid.UUID) (*reelstrip.PlayerReelStripAssignment, error) {
	var assignment reelstrip.PlayerReelStripAssignment
	cacheKey := r.cache.PlayerAssignmentKey(playerID)
	res, err := r.cache.GetWithSingleflight(ctx, cacheKey, &reelstrip.PlayerReelStripAssignment{}, func() (any, error) {
		if err := r.db.WithContext(ctx).
			Where("player_id = ? AND is_active = ?", playerID, true).
			Where("expires_at IS NULL OR expires_at > ?", time.Now()).
			First(&assignment).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return &reelstrip.PlayerReelStripAssignment{
					PlayerID: playerID,
				}, nil
			}
			return nil, fmt.Errorf("failed to get player assignment: %w", err)
		}
		return &assignment, nil
	})

	if err != nil {
		return nil, err
	}

	return res.(*reelstrip.PlayerReelStripAssignment), nil
}

// UpdateAssignment updates a player assignment
func (r *ReelStripGormRepository) UpdateAssignment(ctx context.Context, assignment *reelstrip.PlayerReelStripAssignment) error {
	result := r.db.WithContext(ctx).Save(assignment)
	if result.Error != nil {
		return fmt.Errorf("failed to update assignment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return reelstrip.ErrAssignmentNotFound
	}
	return nil
}

// GetPlayerAssignmentsByPlayerIDs retrieves active assignments for multiple players in a single query
func (r *ReelStripGormRepository) GetPlayerAssignmentsByPlayerIDs(ctx context.Context, playerIDs []uuid.UUID) (map[uuid.UUID]*reelstrip.PlayerReelStripAssignment, error) {
	if len(playerIDs) == 0 {
		return make(map[uuid.UUID]*reelstrip.PlayerReelStripAssignment), nil
	}

	var assignments []*reelstrip.PlayerReelStripAssignment
	if err := r.db.WithContext(ctx).
		Where("player_id IN ? AND is_active = ?", playerIDs, true).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&assignments).Error; err != nil {
		return nil, fmt.Errorf("failed to get player assignments: %w", err)
	}

	result := make(map[uuid.UUID]*reelstrip.PlayerReelStripAssignment, len(assignments))
	for _, a := range assignments {
		result[a.PlayerID] = a
	}
	return result, nil
}

// DeleteAssignment permanently deletes an assignment record
func (r *ReelStripGormRepository) DeleteAssignment(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Unscoped().
		Delete(&reelstrip.PlayerReelStripAssignment{}, "id = ?", id)

	if result.Error != nil {
		return fmt.Errorf("failed to delete assignment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return reelstrip.ErrAssignmentNotFound
	}
	return nil
}
