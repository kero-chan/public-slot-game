package player

import (
	"time"

	"github.com/google/uuid"
)

// Player represents a player account
type Player struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Username     string    `gorm:"type:varchar(50);not null"`
	Email        string    `gorm:"type:varchar(255);not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	Balance      float64   `gorm:"type:decimal(15,2);default:100000.00;not null"`

	// Game association - NULL means cross-game account (can login to any game)
	GameID *uuid.UUID `gorm:"type:uuid;index"`

	// Statistics
	TotalSpins   int     `gorm:"default:0"`
	TotalWagered float64 `gorm:"type:decimal(15,2);default:0.00"`
	TotalWon     float64 `gorm:"type:decimal(15,2);default:0.00"`

	// Status
	IsActive   bool `gorm:"default:true"`
	IsVerified bool `gorm:"default:false"`

	// Timestamps
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	LockVersion int       `gorm:"default:0"`
	LastLoginAt *time.Time
}

// TableName specifies the table name for GORM
func (Player) TableName() string {
	return "players"
}
