package reelstrip

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ReelStripConfig represents a named configuration of reel strips (e.g., "Version A", "Version B", "High RTP", "Low RTP")
// This allows you to manage multiple reel strip configurations and assign them to different players
type ReelStripConfig struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"` // e.g., "v1.0-standard", "v1.0-high-rtp", "v2.0-default"
	GameMode    string    `gorm:"type:varchar(20);not null" json:"game_mode"`         // base_game or free_spins
	Description string    `gorm:"type:text" json:"description,omitempty"`

	// Reel strip references - these are the actual reel strip IDs to use
	Reel0StripID uuid.UUID `gorm:"type:uuid;not null" json:"reel_0_strip_id"`
	Reel1StripID uuid.UUID `gorm:"type:uuid;not null" json:"reel_1_strip_id"`
	Reel2StripID uuid.UUID `gorm:"type:uuid;not null" json:"reel_2_strip_id"`
	Reel3StripID uuid.UUID `gorm:"type:uuid;not null" json:"reel_3_strip_id"`
	Reel4StripID uuid.UUID `gorm:"type:uuid;not null" json:"reel_4_strip_id"`

	// Metadata
	TargetRTP     float64         `gorm:"type:decimal(5,2)" json:"target_rtp,omitempty"` // e.g., 96.50
	IsActive      bool            `gorm:"default:true;index" json:"is_active"`
	IsDefault     bool            `gorm:"default:false;index" json:"is_default"` // Default config for new players
	ActivatedAt   *time.Time      `json:"activated_at,omitempty"`
	DeactivatedAt *time.Time      `json:"deactivated_at,omitempty"`
	CreatedAt     time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy     string          `gorm:"type:varchar(100)" json:"created_by,omitempty"` // Admin username who created this config
	Notes         string          `gorm:"type:text" json:"notes,omitempty"`
	Options       json.RawMessage `gorm:"type:jsonb" json:"options,omitempty"`
}

// TableName specifies the table name for GORM
func (ReelStripConfig) TableName() string {
	return "reel_strip_configs"
}

// ConfigListFilters represents filters for listing reel strip configurations
type ConfigListFilters struct {
	GameMode  *string // Filter by game mode (base_game, free_spins, both)
	IsActive  *bool   // Filter by active status
	IsDefault *bool   // Filter by default status
	Name      *string // Filter by name (partial match)
	Page      int     // Page number (1-indexed)
	Limit     int     // Items per page
}

// PlayerReelStripAssignment assigns a specific reel strip configuration to a player
// This allows A/B testing, VIP configurations, or custom player experiences
type PlayerReelStripAssignment struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	PlayerID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_player_active_assignment" json:"player_id"`

	// Base game configuration
	BaseGameConfigID *uuid.UUID `gorm:"type:uuid" json:"base_game_config_id,omitempty"`

	// Free spins configuration (can be different from base game)
	FreeSpinsConfigID *uuid.UUID `gorm:"type:uuid" json:"free_spins_config_id,omitempty"`

	// Assignment metadata
	AssignedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"assigned_at"`
	AssignedBy string     `gorm:"type:varchar(100)" json:"assigned_by,omitempty"` // Admin who made the assignment
	Reason     string     `gorm:"type:varchar(255)" json:"reason,omitempty"`      // e.g., "A/B Test Group A", "VIP Player"
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`                           // Optional expiration for temporary assignments
	IsActive   bool       `gorm:"default:true;index" json:"is_active"`
}

// TableName specifies the table name for GORM
func (PlayerReelStripAssignment) TableName() string {
	return "player_reel_strip_assignments"
}

// ReelStripConfigSet represents a complete configuration with loaded reel strips
type ReelStripConfigSet struct {
	Config *ReelStripConfig
	Strips [5]*ReelStrip
	TTL    *time.Duration
}

// IsComplete checks if all 5 reel strips are loaded
func (s *ReelStripConfigSet) IsComplete() bool {
	if s.Config == nil {
		return false
	}
	for i := 0; i < 5; i++ {
		if s.Strips[i] == nil {
			return false
		}
	}
	return true
}
