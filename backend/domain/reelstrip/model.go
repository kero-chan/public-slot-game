package reelstrip

import (
	"time"

	"github.com/google/uuid"
)

// GameMode represents the game mode type
type GameMode string

const (
	BaseGame         GameMode = "base_game"
	FreeSpins        GameMode = "free_spins"
	BonusSpinTrigger GameMode = "bonus_spin_trigger"
	Both             GameMode = "both"
)

// ReelStrip represents a pre-generated reel strip stored in database
// Version control is managed at the ReelStripConfig level, not here
type ReelStrip struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	GameMode   string    `gorm:"type:varchar(20);not null" json:"game_mode"`
	ReelNumber int       `gorm:"not null" json:"reel_number"`

	// Strip data stored as JSONB array
	StripData   []string `gorm:"type:jsonb;not null;serializer:json" json:"strip_data"`
	Checksum    string   `gorm:"type:varchar(64);not null;uniqueIndex" json:"checksum"`
	StripLength int      `gorm:"not null" json:"strip_length"`

	// Metadata
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	IsActive  bool      `gorm:"default:true;index:idx_reel_strips_active" json:"is_active"`
	Notes     string    `gorm:"type:text" json:"notes,omitempty"`
}

// TableName specifies the table name for GORM
func (ReelStrip) TableName() string {
	return "reel_strips"
}

// ReelStripSet represents a complete set of 5 reel strips for a game mode
type ReelStripSet struct {
	GameMode GameMode
	Strips   [5]*ReelStrip // 5 reels (indexed 0-4)
}

// IsComplete checks if the set has all 5 reels
func (s *ReelStripSet) IsComplete() bool {
	for i := 0; i < 5; i++ {
		if s.Strips[i] == nil {
			return false
		}
	}
	return true
}

// GetStripData returns the strip data arrays for all 5 reels
func (s *ReelStripSet) GetStripData() [][]string {
	result := make([][]string, 5)
	for i := 0; i < 5; i++ {
		if s.Strips[i] != nil {
			result[i] = s.Strips[i].StripData
		}
	}
	return result
}
