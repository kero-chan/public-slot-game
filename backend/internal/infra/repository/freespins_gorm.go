package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/freespins"
	"gorm.io/gorm"
)

// FreeSpinsGormRepository implements freespins.Repository using GORM
type FreeSpinsGormRepository struct {
	db *gorm.DB
}

// NewFreeSpinsGormRepository creates a new GORM free spins repository
func NewFreeSpinsGormRepository(db *gorm.DB) freespins.Repository {
	return &FreeSpinsGormRepository{
		db: db,
	}
}

// Create creates a new free spins session
func (r *FreeSpinsGormRepository) Create(ctx context.Context, session *freespins.FreeSpinsSession) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("failed to create free spins session: %w", err)
	}
	return nil
}

// GetAvailableSessionByID retrieves a free spins session active and remaining spin > 0  by ID
func (r *FreeSpinsGormRepository) GetAvailableSessionByID(ctx context.Context, id uuid.UUID) (*freespins.FreeSpinsSession, error) {
	var session freespins.FreeSpinsSession
	if err := r.db.WithContext(ctx).Where("id = ? and is_active = true and remaining_spins > 0", id).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, freespins.ErrFreeSpinsNotFound
		}
		return nil, fmt.Errorf("failed to get free spins session by ID: %w", err)
	}
	return &session, nil
}

// GetByID retrieves a free spins session by ID
func (r *FreeSpinsGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*freespins.FreeSpinsSession, error) {
	var session freespins.FreeSpinsSession
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, freespins.ErrFreeSpinsNotFound
		}
		return nil, fmt.Errorf("failed to get free spins session by ID: %w", err)
	}
	return &session, nil
}

// GetActiveByPlayer retrieves the active free spins session for a player
func (r *FreeSpinsGormRepository) GetActiveByPlayer(ctx context.Context, playerID uuid.UUID) (*freespins.FreeSpinsSession, error) {
	var session freespins.FreeSpinsSession
	err := r.db.WithContext(ctx).
		Where("player_id = ? AND is_active = ?", playerID, true).
		Order("created_at DESC").
		First(&session).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, freespins.ErrFreeSpinsNotFound
		}
		return nil, fmt.Errorf("failed to get active free spins session: %w", err)
	}
	return &session, nil
}

// Update updates a free spins session
func (r *FreeSpinsGormRepository) Update(ctx context.Context, session *freespins.FreeSpinsSession) error {
	result := r.db.WithContext(ctx).Save(session)
	if result.Error != nil {
		return fmt.Errorf("failed to update free spins session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return freespins.ErrFreeSpinsNotFound
	}
	return nil
}

// UpdateSpins updates spins completed and remaining
func (r *FreeSpinsGormRepository) UpdateSpins(ctx context.Context, id uuid.UUID, spinsCompleted, remainingSpins int) error {
	result := r.db.WithContext(ctx).
		Model(&freespins.FreeSpinsSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"spins_completed": spinsCompleted,
			"remaining_spins": remainingSpins,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update spins: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return freespins.ErrFreeSpinsNotFound
	}
	return nil
}

// AddTotalWon updates total won amount
func (r *FreeSpinsGormRepository) AddTotalWon(ctx context.Context, id uuid.UUID, amount float64) error {
	result := r.db.WithContext(ctx).
		Model(&freespins.FreeSpinsSession{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"total_won":  gorm.Expr("total_won + ?", amount),
			"updated_at": time.Now().UTC(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update total won: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return freespins.ErrFreeSpinsNotFound
	}
	return nil
}

// CompleteSession marks a free spins session as completed
func (r *FreeSpinsGormRepository) CompleteSession(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&freespins.FreeSpinsSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":    false,
			"is_completed": true,
			"completed_at": now,
			"updated_at":   now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to complete session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return freespins.ErrFreeSpinsNotFound
	}
	return nil
}

func (r *FreeSpinsGormRepository) ExecuteSpinWithLock(ctx context.Context, id uuid.UUID, additionalSpins int, lockVersion int) error {
	result := r.db.WithContext(ctx).
		Model(&freespins.FreeSpinsSession{}).
		Where("id = ? and lock_version = ?", id, lockVersion).
		Updates(map[string]interface{}{
			"spins_completed": gorm.Expr("spins_completed - ?", additionalSpins),
			"remaining_spins": gorm.Expr("remaining_spins + ?", additionalSpins),
			"updated_at":      time.Now().UTC(),
			"lock_version":    gorm.Expr("lock_version + 1"),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to add spins: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return freespins.ErrNotFoundOrLockChanged
	}
	return nil
}

func (r *FreeSpinsGormRepository) RollbackSpin(ctx context.Context, id uuid.UUID, additionalSpins int) error {
	result := r.db.WithContext(ctx).
		Model(&freespins.FreeSpinsSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"spins_completed": gorm.Expr("spins_completed - ?", additionalSpins),
			"remaining_spins": gorm.Expr("remaining_spins + ?", additionalSpins),
			"updated_at":      time.Now().UTC(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to add spins: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return freespins.ErrNotFound
	}
	return nil
}

// AddSpins adds additional spins (for retrigger)
func (r *FreeSpinsGormRepository) AddSpins(ctx context.Context, id uuid.UUID, additionalSpins int) error {
	result := r.db.WithContext(ctx).
		Model(&freespins.FreeSpinsSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_spins_awarded": gorm.Expr("total_spins_awarded + ?", additionalSpins),
			"remaining_spins":     gorm.Expr("remaining_spins + ?", additionalSpins),
			"updated_at":          time.Now().UTC(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to add spins: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return freespins.ErrFreeSpinsNotFound
	}
	return nil
}

// GetByPlayer retrieves all free spins sessions for a player
func (r *FreeSpinsGormRepository) GetByPlayer(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*freespins.FreeSpinsSession, error) {
	var sessions []*freespins.FreeSpinsSession
	err := r.db.WithContext(ctx).
		Where("player_id = ?", playerID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get free spins sessions by player: %w", err)
	}
	return sessions, nil
}
