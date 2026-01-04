package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/provablyfair"
	"gorm.io/gorm"
)

// ProvablyFairGormRepository implements provablyfair.Repository using GORM
type ProvablyFairGormRepository struct {
	db *gorm.DB
}

// NewProvablyFairGormRepository creates a new GORM provably fair repository
func NewProvablyFairGormRepository(db *gorm.DB) *ProvablyFairGormRepository {
	return &ProvablyFairGormRepository{
		db: db,
	}
}

// Ensure ProvablyFairGormRepository implements Repository
var _ provablyfair.Repository = (*ProvablyFairGormRepository)(nil)

// ==================== PFSession Operations ====================

// CreateSession creates a new provably fair session
func (r *ProvablyFairGormRepository) CreateSession(ctx context.Context, session *provablyfair.PFSession) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("failed to create PF session: %w", err)
	}
	return nil
}

// GetSessionByID retrieves a PF session by ID
func (r *ProvablyFairGormRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*provablyfair.PFSession, error) {
	var session provablyfair.PFSession
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, provablyfair.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get PF session by ID: %w", err)
	}
	return &session, nil
}

// GetActiveSessionByPlayer retrieves the active PF session for a player
func (r *ProvablyFairGormRepository) GetActiveSessionByPlayer(ctx context.Context, playerID uuid.UUID) (*provablyfair.PFSession, error) {
	var session provablyfair.PFSession
	err := r.db.WithContext(ctx).
		Where("player_id = ? AND status = ?", playerID, provablyfair.SessionStatusActive).
		Order("created_at DESC").
		First(&session).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, provablyfair.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get active PF session: %w", err)
	}
	return &session, nil
}

// GetActiveSessionByGameSession retrieves the active PF session for a game session
func (r *ProvablyFairGormRepository) GetActiveSessionByGameSession(ctx context.Context, gameSessionID uuid.UUID) (*provablyfair.PFSession, error) {
	var session provablyfair.PFSession
	err := r.db.WithContext(ctx).
		Where("game_session_id = ? AND status = ?", gameSessionID, provablyfair.SessionStatusActive).
		First(&session).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, provablyfair.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get active PF session by game session: %w", err)
	}
	return &session, nil
}

// UpdateSession updates a PF session
func (r *ProvablyFairGormRepository) UpdateSession(ctx context.Context, session *provablyfair.PFSession) error {
	result := r.db.WithContext(ctx).Save(session)
	if result.Error != nil {
		return fmt.Errorf("failed to update PF session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return provablyfair.ErrSessionNotFound
	}
	return nil
}

// EndSession marks a PF session as ended
func (r *ProvablyFairGormRepository) EndSession(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&provablyfair.PFSession{}).
		Where("id = ? AND status = ?", id, provablyfair.SessionStatusActive).
		Updates(map[string]any{
			"status":   provablyfair.SessionStatusEnded,
			"ended_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to end PF session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return provablyfair.ErrSessionNotFound
	}
	return nil
}

// ==================== SpinLog Operations ====================

// CreateSpinLog creates a new spin log entry (append-only)
func (r *ProvablyFairGormRepository) CreateSpinLog(ctx context.Context, log *provablyfair.SpinLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("failed to create spin log: %w", err)
	}
	return nil
}

// GetSpinLogsBySession retrieves all spin logs for a PF session
func (r *ProvablyFairGormRepository) GetSpinLogsBySession(ctx context.Context, pfSessionID uuid.UUID) ([]provablyfair.SpinLog, error) {
	var logs []provablyfair.SpinLog
	err := r.db.WithContext(ctx).
		Where("pf_session_id = ?", pfSessionID).
		Order("spin_index ASC").
		Find(&logs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get spin logs: %w", err)
	}
	return logs, nil
}

// GetSpinLogByIndex retrieves a specific spin log by session and index
func (r *ProvablyFairGormRepository) GetSpinLogByIndex(ctx context.Context, pfSessionID uuid.UUID, spinIndex int64) (*provablyfair.SpinLog, error) {
	var log provablyfair.SpinLog
	err := r.db.WithContext(ctx).
		Where("pf_session_id = ? AND spin_index = ?", pfSessionID, spinIndex).
		First(&log).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, provablyfair.ErrSpinNotFound
		}
		return nil, fmt.Errorf("failed to get spin log by index: %w", err)
	}
	return &log, nil
}

// GetLastSpinLog retrieves the most recent spin log for a session
func (r *ProvablyFairGormRepository) GetLastSpinLog(ctx context.Context, pfSessionID uuid.UUID) (*provablyfair.SpinLog, error) {
	var log provablyfair.SpinLog
	err := r.db.WithContext(ctx).
		Where("pf_session_id = ?", pfSessionID).
		Order("spin_index DESC").
		First(&log).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, provablyfair.ErrSpinNotFound
		}
		return nil, fmt.Errorf("failed to get last spin log: %w", err)
	}
	return &log, nil
}

// ==================== SessionAudit Operations ====================

// CreateSessionAudit creates a new session audit entry (reveals server seed)
func (r *ProvablyFairGormRepository) CreateSessionAudit(ctx context.Context, audit *provablyfair.SessionAudit) error {
	if err := r.db.WithContext(ctx).Create(audit).Error; err != nil {
		return fmt.Errorf("failed to create session audit: %w", err)
	}
	return nil
}

// GetSessionAudit retrieves the session audit for a PF session
func (r *ProvablyFairGormRepository) GetSessionAudit(ctx context.Context, pfSessionID uuid.UUID) (*provablyfair.SessionAudit, error) {
	var audit provablyfair.SessionAudit
	err := r.db.WithContext(ctx).
		Where("pf_session_id = ?", pfSessionID).
		First(&audit).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, provablyfair.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session audit: %w", err)
	}
	return &audit, nil
}
