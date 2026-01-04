package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/player"
	"gorm.io/gorm"
)

// PlayerGormRepository implements player.Repository using GORM
type PlayerGormRepository struct {
	db *gorm.DB
}

// NewPlayerGormRepository creates a new GORM player repository
func NewPlayerGormRepository(db *gorm.DB) player.Repository {
	return &PlayerGormRepository{
		db: db,
	}
}

// Create creates a new player
func (r *PlayerGormRepository) Create(ctx context.Context, p *player.Player) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return fmt.Errorf("failed to create player: %w", err)
	}
	return nil
}

// GetByID retrieves a player by ID
func (r *PlayerGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*player.Player, error) {
	var p player.Player
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, player.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to get player by ID: %w", err)
	}
	return &p, nil
}

// GetByUsername retrieves a player by username (case-insensitive)
func (r *PlayerGormRepository) GetByUsername(ctx context.Context, username string) (*player.Player, error) {
	var p player.Player
	if err := r.db.WithContext(ctx).Where("LOWER(username) = LOWER(?)", username).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, player.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to get player by username: %w", err)
	}
	return &p, nil
}

// GetByEmail retrieves a player by email (case-insensitive)
func (r *PlayerGormRepository) GetByEmail(ctx context.Context, email string) (*player.Player, error) {
	var p player.Player
	if err := r.db.WithContext(ctx).Where("LOWER(email) = LOWER(?)", email).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, player.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to get player by email: %w", err)
	}
	return &p, nil
}

// GetByUsernameAndGame retrieves a player by username and game_id (case-insensitive)
func (r *PlayerGormRepository) GetByUsernameAndGame(ctx context.Context, username string, gameID *uuid.UUID) (*player.Player, error) {
	var p player.Player
	query := r.db.WithContext(ctx).Where("LOWER(username) = LOWER(?)", username)

	if gameID != nil {
		query = query.Where("game_id = ?", *gameID)
	} else {
		query = query.Where("game_id IS NULL")
	}

	if err := query.First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, player.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to get player by username and game: %w", err)
	}
	return &p, nil
}

// GetByEmailAndGame retrieves a player by email and game_id (case-insensitive)
func (r *PlayerGormRepository) GetByEmailAndGame(ctx context.Context, email string, gameID *uuid.UUID) (*player.Player, error) {
	var p player.Player
	query := r.db.WithContext(ctx).Where("LOWER(email) = LOWER(?)", email)

	if gameID != nil {
		query = query.Where("game_id = ?", *gameID)
	} else {
		query = query.Where("game_id IS NULL")
	}

	if err := query.First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, player.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to get player by email and game: %w", err)
	}
	return &p, nil
}

// FindLoginCandidate finds a player who can login with given username and game (case-insensitive)
// Returns player if: exact game match OR cross-game account (game_id = NULL)
// Prefers game-specific account over cross-game account if both exist
func (r *PlayerGormRepository) FindLoginCandidate(ctx context.Context, username string, gameID *uuid.UUID) (*player.Player, error) {
	var p player.Player

	query := r.db.WithContext(ctx).Where("LOWER(username) = LOWER(?)", username)

	if gameID != nil {
		// Find player with exact game_id match OR cross-game account (NULL)
		query = query.Where("game_id = ? OR game_id IS NULL", *gameID)
		// Order to prefer game-specific over cross-game (NOT NULL first)
		query = query.Order("game_id IS NULL ASC")
	} else {
		// If no game_id specified, only find cross-game accounts
		query = query.Where("game_id IS NULL")
	}

	if err := query.First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, player.ErrPlayerNotFound
		}
		return nil, fmt.Errorf("failed to find login candidate: %w", err)
	}
	return &p, nil
}

// Update updates a player's information
func (r *PlayerGormRepository) Update(ctx context.Context, p *player.Player) error {
	result := r.db.WithContext(ctx).Save(p)
	if result.Error != nil {
		return fmt.Errorf("failed to update player: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return player.ErrPlayerNotFound
	}
	return nil
}

// UpdateBalance updates a player's balance
func (r *PlayerGormRepository) UpdateBalanceWithLock(ctx context.Context, id uuid.UUID, amount float64, lockVersion int) error {
	result := r.db.WithContext(ctx).
		Model(&player.Player{}).
		Where("id = ? and lock_version = ?", id, lockVersion).
		Updates(map[string]any{
			"balance":      gorm.Expr("balance + ?", amount),
			"lock_version": gorm.Expr("lock_version + 1"),
			"updated_at":   time.Now().UTC(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update balance: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return player.ErrNotFoundOrLockChanged
	}
	return nil
}

// UpdateBalance updates a player's balance
func (r *PlayerGormRepository) UpdateBalance(ctx context.Context, id uuid.UUID, amount float64) error {
	result := r.db.WithContext(ctx).
		Model(&player.Player{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"balance":    gorm.Expr("balance + ?", amount),
			"updated_at": time.Now().UTC(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update balance: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return player.ErrPlayerNotFound
	}
	return nil
}

// UpdateBalanceWithTx updates a player's balance using transaction from context
func (r *PlayerGormRepository) UpdateBalanceWithTx(ctx context.Context, id uuid.UUID, amount float64) error {
	db := GetDBOrTx(ctx, r.db)
	result := db.
		Model(&player.Player{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"balance":    gorm.Expr("balance + ?", amount),
			"updated_at": time.Now().UTC(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update balance: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return player.ErrPlayerNotFound
	}
	return nil
}

// UpdateBalanceWithLockAndTx updates a player's balance with optimistic locking using transaction from context
func (r *PlayerGormRepository) UpdateBalanceWithLockAndTx(ctx context.Context, id uuid.UUID, amount float64, lockVersion int) error {
	db := GetDBOrTx(ctx, r.db)
	result := db.
		Model(&player.Player{}).
		Where("id = ? AND lock_version = ?", id, lockVersion).
		Updates(map[string]any{
			"balance":      gorm.Expr("balance + ?", amount),
			"lock_version": gorm.Expr("lock_version + 1"),
			"updated_at":   time.Now().UTC(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update balance: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return player.ErrNotFoundOrLockChanged
	}
	return nil
}

// UpdateStatistics updates player statistics
func (r *PlayerGormRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, spins int, wagered, won float64) error {
	result := r.db.WithContext(ctx).
		Model(&player.Player{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_spins":   gorm.Expr("total_spins + ?", spins),
			"total_wagered": gorm.Expr("total_wagered + ?", wagered),
			"total_won":     gorm.Expr("total_won + ?", won),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update statistics: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return player.ErrPlayerNotFound
	}
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *PlayerGormRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&player.Player{}).
		Where("id = ?", id).
		Update("last_login_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return player.ErrPlayerNotFound
	}
	return nil
}

// Delete deletes a player (hard delete)
func (r *PlayerGormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&player.Player{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete player: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return player.ErrPlayerNotFound
	}
	return nil
}

// List retrieves a list of players with filters and pagination
func (r *PlayerGormRepository) List(ctx context.Context, filters player.ListFilters) ([]*player.Player, int64, error) {
	var players []*player.Player
	var total int64

	query := r.db.WithContext(ctx).Model(&player.Player{})

	// Apply filters
	if filters.Username != "" {
		query = query.Where("username ILIKE ?", "%"+filters.Username+"%")
	}
	if filters.Email != "" {
		query = query.Where("email ILIKE ?", "%"+filters.Email+"%")
	}
	if filters.GameID != nil {
		query = query.Where("game_id = ?", *filters.GameID)
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count players: %w", err)
	}

	// Apply sorting
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "DESC"
	if !filters.SortDesc {
		sortOrder = "ASC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	} else {
		query = query.Limit(20) // Default limit
	}
	if filters.Page > 0 {
		offset := (filters.Page - 1) * filters.Limit
		query = query.Offset(offset)
	}

	// Execute query
	if err := query.Find(&players).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list players: %w", err)
	}

	return players, total, nil
}
