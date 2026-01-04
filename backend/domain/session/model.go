package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// GameSession represents a game session
type GameSession struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PlayerID        uuid.UUID `gorm:"type:uuid;not null;index"`
	BetAmount       float64   `gorm:"type:decimal(10,2);not null"`
	StartingBalance float64   `gorm:"type:decimal(15,2);not null"`
	EndingBalance   *float64  `gorm:"type:decimal(15,2)"`

	// Statistics
	TotalSpins   int     `gorm:"default:0"`
	TotalWagered float64 `gorm:"type:decimal(15,2);default:0.00"`
	TotalWon     float64 `gorm:"type:decimal(15,2);default:0.00"`
	NetChange    float64 `gorm:"type:decimal(15,2);default:0.00"`

	// Timestamps
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	EndedAt   *time.Time `gorm:"index"`
}

// TableName specifies the table name for GORM
func (GameSession) TableName() string {
	return "game_sessions"
}

// PlayerSession represents an active login session for a player
// Used for single-device enforcement and force logout capability
type PlayerSession struct {
	ID       uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PlayerID uuid.UUID  `gorm:"type:uuid;not null;index"`
	GameID   *uuid.UUID `gorm:"type:uuid;index"` // nil for cross-game accounts

	// Session token (unique identifier stored in JWT and Redis)
	SessionToken string `gorm:"type:varchar(64);not null;uniqueIndex"`

	// Device/client info
	DeviceInfo *string `gorm:"type:varchar(255)"`
	IPAddress  *string `gorm:"type:varchar(45)"`
	UserAgent  *string `gorm:"type:text"`

	// Session status
	IsActive bool `gorm:"not null;default:true"`

	// Timestamps
	CreatedAt      time.Time  `gorm:"not null;default:now()"`
	ExpiresAt      time.Time  `gorm:"not null"`
	LastActivityAt time.Time  `gorm:"not null;default:now()"`
	LoggedOutAt    *time.Time `gorm:""`

	// Logout reason: manual, forced, expired
	LogoutReason *string `gorm:"type:varchar(20)"`
}

// TableName specifies the table name for GORM
func (PlayerSession) TableName() string {
	return "player_sessions"
}

// Logout reasons
const (
	LogoutReasonManual  = "manual"  // User clicked logout
	LogoutReasonForced  = "forced"  // New device login forced logout
	LogoutReasonExpired = "expired" // Token expired
)

// PlayerSession errors
var (
	ErrPlayerSessionNotFound     = errors.New("player session not found")
	ErrPlayerSessionExpired      = errors.New("player session expired")
	ErrPlayerSessionInactive     = errors.New("player session is inactive")
	ErrPlayerSessionForcedLogout = errors.New("player session is forced logout because logged in from another device")
	ErrPlayerAlreadyLoggedIn     = errors.New("player is already logged in on another device")
	ErrPlayerSessionGameMismatch = errors.New("session game does not match requested game")
)
