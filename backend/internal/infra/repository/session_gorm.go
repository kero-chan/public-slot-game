package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/session"
	"gorm.io/gorm"
)

// SessionGormRepository implements session.Repository using GORM
type SessionGormRepository struct {
	db *gorm.DB
}

// NewSessionGormRepository creates a new GORM session repository
func NewSessionGormRepository(db *gorm.DB) session.Repository {
	return &SessionGormRepository{
		db: db,
	}
}

// Create creates a new game session
func (r *SessionGormRepository) Create(ctx context.Context, s *session.GameSession) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetByID retrieves a session by ID
func (r *SessionGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*session.GameSession, error) {
	var s session.GameSession
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, session.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}
	return &s, nil
}

// GetActiveSessionByPlayer retrieves the active session for a player
func (r *SessionGormRepository) GetActiveSessionByPlayer(ctx context.Context, playerID uuid.UUID) (*session.GameSession, error) {
	var s session.GameSession
	err := r.db.WithContext(ctx).
		Where("player_id = ? AND ended_at IS NULL", playerID).
		Order("created_at DESC").
		First(&s).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, session.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}
	return &s, nil
}

// Update updates a session
func (r *SessionGormRepository) Update(ctx context.Context, s *session.GameSession) error {
	result := r.db.WithContext(ctx).Save(s)
	if result.Error != nil {
		return fmt.Errorf("failed to update session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return session.ErrSessionNotFound
	}
	return nil
}

// EndSession marks a session as ended
func (r *SessionGormRepository) EndSession(ctx context.Context, id uuid.UUID, endingBalance float64) error {
	now := time.Now().UTC()

	// First, get the session to calculate net change
	var s session.GameSession
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return session.ErrSessionNotFound
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Calculate net change
	netChange := endingBalance - s.StartingBalance

	// Update session
	result := r.db.WithContext(ctx).
		Model(&session.GameSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"ended_at":       now,
			"ending_balance": endingBalance,
			"net_change":     netChange,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to end session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return session.ErrSessionNotFound
	}
	return nil
}

// UpdateStatistics updates session statistics
func (r *SessionGormRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, spins int, wagered, won float64) error {
	result := r.db.WithContext(ctx).
		Model(&session.GameSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_spins":   gorm.Expr("total_spins + ?", spins),
			"total_wagered": gorm.Expr("total_wagered + ?", wagered),
			"total_won":     gorm.Expr("total_won + ?", won),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update session statistics: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return session.ErrSessionNotFound
	}
	return nil
}

// GetByPlayer retrieves all sessions for a player (paginated)
func (r *SessionGormRepository) GetByPlayer(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*session.GameSession, error) {
	var sessions []*session.GameSession
	err := r.db.WithContext(ctx).
		Where("player_id = ?", playerID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by player: %w", err)
	}
	return sessions, nil
}

// PlayerSessionGormRepository implements session.PlayerSessionRepository using GORM
type PlayerSessionGormRepository struct {
	db *gorm.DB
}

// NewPlayerSessionGormRepository creates a new GORM player session repository
func NewPlayerSessionGormRepository(db *gorm.DB) session.PlayerSessionRepository {
	return &PlayerSessionGormRepository{
		db: db,
	}
}

// Create creates a new player session
func (r *PlayerSessionGormRepository) Create(ctx context.Context, s *session.PlayerSession) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return fmt.Errorf("failed to create player session: %w", err)
	}
	return nil
}

// GetByToken retrieves a session by session token
func (r *PlayerSessionGormRepository) GetByToken(ctx context.Context, token string) (*session.PlayerSession, error) {
	var s session.PlayerSession
	if err := r.db.WithContext(ctx).
		Where("session_token = ? AND is_active = true", token).
		First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, session.ErrPlayerSessionNotFound
		}
		return nil, fmt.Errorf("failed to get player session by token: %w", err)
	}
	return &s, nil
}

// GetActiveByPlayerAndGame retrieves active session for a player and game
func (r *PlayerSessionGormRepository) GetActiveByPlayerAndGame(ctx context.Context, playerID uuid.UUID, gameID *uuid.UUID) (*session.PlayerSession, error) {
	var s session.PlayerSession
	query := r.db.WithContext(ctx).
		Where("player_id = ? AND is_active = true", playerID)

	if gameID != nil {
		query = query.Where("game_id = ?", *gameID)
	} else {
		query = query.Where("game_id IS NULL")
	}

	if err := query.First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, session.ErrPlayerSessionNotFound
		}
		return nil, fmt.Errorf("failed to get active player session: %w", err)
	}
	return &s, nil
}

// DeactivateSession marks a session as inactive with logout reason
func (r *PlayerSessionGormRepository) DeactivateSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&session.PlayerSession{}).
		Where("id = ? AND is_active = true", sessionID).
		Updates(map[string]interface{}{
			"is_active":     false,
			"logged_out_at": now,
			"logout_reason": reason,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate session: %w", result.Error)
	}
	return nil
}

// DeactivateAllPlayerSessions deactivates all active sessions for a player
func (r *PlayerSessionGormRepository) DeactivateAllPlayerSessions(ctx context.Context, playerID uuid.UUID, reason string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&session.PlayerSession{}).
		Where("player_id = ? AND is_active = true", playerID).
		Updates(map[string]interface{}{
			"is_active":     false,
			"logged_out_at": now,
			"logout_reason": reason,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate all player sessions: %w", result.Error)
	}
	return nil
}

// DeactivatePlayerGameSession deactivates active session for a player in specific game
func (r *PlayerSessionGormRepository) DeactivatePlayerGameSession(ctx context.Context, playerID uuid.UUID, gameID *uuid.UUID, reason string) error {
	now := time.Now().UTC()
	query := r.db.WithContext(ctx).
		Model(&session.PlayerSession{}).
		Where("player_id = ? AND is_active = true", playerID)

	if gameID != nil {
		query = query.Where("game_id = ?", *gameID)
	} else {
		query = query.Where("game_id IS NULL")
	}

	result := query.Updates(map[string]interface{}{
		"is_active":     false,
		"logged_out_at": now,
		"logout_reason": reason,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate player game session: %w", result.Error)
	}
	return nil
}

// UpdateLastActivity updates the last activity timestamp
func (r *PlayerSessionGormRepository) UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Model(&session.PlayerSession{}).
		Where("id = ? AND is_active = true", sessionID).
		Update("last_activity_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to update last activity: %w", result.Error)
	}
	return nil
}

// CleanupExpiredSessions marks expired sessions as inactive
func (r *PlayerSessionGormRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	reason := session.LogoutReasonExpired

	result := r.db.WithContext(ctx).
		Model(&session.PlayerSession{}).
		Where("is_active = true AND expires_at < ?", now).
		Updates(map[string]interface{}{
			"is_active":     false,
			"logged_out_at": now,
			"logout_reason": reason,
		})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}
	return result.RowsAffected, nil
}
