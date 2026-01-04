package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/spin"
	"gorm.io/gorm"
)

// SpinGormRepository implements spin.Repository using GORM
type SpinGormRepository struct {
	db *gorm.DB
}

// NewSpinGormRepository creates a new GORM spin repository
func NewSpinGormRepository(db *gorm.DB) spin.Repository {
	return &SpinGormRepository{
		db: db,
	}
}

// Create creates a new spin record
func (r *SpinGormRepository) Create(ctx context.Context, s *spin.Spin) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("failed to create spin: %w", err)
	}
	return nil
}

// GetByID retrieves a spin by ID
func (r *SpinGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*spin.Spin, error) {
	var s spin.Spin
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, spin.ErrSpinNotFound
		}
		return nil, fmt.Errorf("failed to get spin by ID: %w", err)
	}
	return &s, nil
}

// GetBySession retrieves spins for a session with a reasonable limit
// Default limit of 1000 spins per session prevents memory issues
func (r *SpinGormRepository) GetBySession(ctx context.Context, sessionID uuid.UUID) ([]*spin.Spin, error) {
	const maxSpinsPerSession = 1000

	var spins []*spin.Spin
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Limit(maxSpinsPerSession).
		Find(&spins).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get spins by session: %w", err)
	}
	return spins, nil
}

// GetByPlayer retrieves spins for a player (paginated)
func (r *SpinGormRepository) GetByPlayer(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*spin.Spin, error) {
	var spins []*spin.Spin
	err := r.db.WithContext(ctx).
		Where("player_id = ?", playerID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&spins).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get spins by player: %w", err)
	}
	return spins, nil
}

// GetByPlayerInTimeRange retrieves spins for a player in a time range
func (r *SpinGormRepository) GetByPlayerInTimeRange(
	ctx context.Context,
	playerID uuid.UUID,
	start, end time.Time,
	limit, offset int,
) ([]*spin.Spin, error) {
	var spins []*spin.Spin
	err := r.db.WithContext(ctx).
		Where("player_id = ? AND created_at >= ? AND created_at <= ?", playerID, start, end).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&spins).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get spins in time range: %w", err)
	}
	return spins, nil
}

// GetByFreeSpinsSession retrieves spins for a free spins session with a reasonable limit
// Free spin sessions typically have 12-20 spins, limit of 100 handles retriggers safely
func (r *SpinGormRepository) GetByFreeSpinsSession(ctx context.Context, freeSpinsSessionID uuid.UUID) ([]*spin.Spin, error) {
	const maxFreeSpinsPerSession = 100

	var spins []*spin.Spin
	err := r.db.WithContext(ctx).
		Where("free_spins_session_id = ?", freeSpinsSessionID).
		Order("created_at ASC").
		Limit(maxFreeSpinsPerSession).
		Find(&spins).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get spins by free spins session: %w", err)
	}
	return spins, nil
}

// Count counts total spins for a player
func (r *SpinGormRepository) Count(ctx context.Context, playerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&spin.Spin{}).
		Where("player_id = ?", playerID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count spins: %w", err)
	}
	return count, nil
}

// CountInTimeRange counts spins for a player in a time range
func (r *SpinGormRepository) CountInTimeRange(
	ctx context.Context,
	playerID uuid.UUID,
	start, end time.Time,
) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&spin.Spin{}).
		Where("player_id = ? AND created_at >= ? AND created_at <= ?", playerID, start, end).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count spins in time range: %w", err)
	}
	return count, nil
}

func (r *SpinGormRepository) UpdateFreeSpinsSessionId(ctx context.Context, id uuid.UUID, freeSpinsSessionID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&spin.Spin{}).
		Where("id = ?", id).
		Update("free_spins_session_id", freeSpinsSessionID).Error
}
