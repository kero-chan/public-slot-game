package freespins

import (
	"time"

	"github.com/google/uuid"
)

// FreeSpinsSession represents a free spins bonus session
type FreeSpinsSession struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	PlayerID            uuid.UUID  `gorm:"type:uuid;not null;index"`
	SessionID           uuid.UUID  `gorm:"type:uuid;not null"`
	TriggeredBySpinID   *uuid.UUID `gorm:"type:uuid"`
	ScatterCount        int        `gorm:"not null"`
	TotalSpinsAwarded   int        `gorm:"not null"`
	SpinsCompleted      int        `gorm:"default:0"`
	RemainingSpins      int        `gorm:"not null"`
	LockedBetAmount     float64    `gorm:"type:decimal(10,2);not null"`
	TotalWon            float64    `gorm:"type:decimal(15,2);default:0.00"`
	IsActive            bool       `gorm:"default:true;index"`
	IsCompleted         bool       `gorm:"default:false"`
	ReelStripConfigID   *uuid.UUID `gorm:"type:uuid"`
	CreatedAt           time.Time  `gorm:"default:CURRENT_TIMESTAMP;index"`
	UpdatedAt           time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	LockVersion         int        `gorm:"default:0"`
	CompletedAt         *time.Time
}

// TableName specifies the table name for GORM
func (FreeSpinsSession) TableName() string {
	return "free_spins_sessions"
}
