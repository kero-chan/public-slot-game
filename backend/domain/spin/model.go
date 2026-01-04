package spin

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Spin represents a single spin execution
type Spin struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	SessionID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	PlayerID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	BetAmount          float64    `gorm:"type:decimal(10,2);not null"`
	BalanceBefore      float64    `gorm:"type:decimal(15,2);not null"`
	BalanceAfter       float64    `gorm:"type:decimal(15,2);not null"`
	Grid               Grid       `gorm:"type:jsonb;not null"`
	Cascades           Cascades   `gorm:"type:jsonb"`
	TotalWin           float64    `gorm:"type:decimal(15,2);default:0.00"`
	ScatterCount       int        `gorm:"default:0"`
	ReelPositions      []int      `gorm:"type:jsonb;not null;serializer:json" json:"-"`
	IsFreeSpin         bool       `gorm:"default:false;index"`
	FreeSpinsSessionID *uuid.UUID `gorm:"type:uuid;index"`
	FreeSpinsTriggered bool       `gorm:"default:false"`
	GameMode           *string    `gorm:"type:varchar(32);index"` // NULL for normal spin, or: free_spin_trigger, wild_spin_trigger, bonus_spin_trigger
	GameModeCost       *float64   `gorm:"type:decimal(10,2)"`     // Cost paid for game mode (500, 750, 1000), NULL for normal spin
	CreatedAt          time.Time  `gorm:"default:CURRENT_TIMESTAMP;index"`
}

// Grid represents the game grid (5x6)
type Grid [][]string

// Cascades represents all cascades in a spin
type Cascades []Cascade

// Cascade represents a single cascade
type Cascade struct {
	CascadeNumber   int          `json:"cascade_number"`
	GridAfter       Grid         `json:"grid_after"`
	Multiplier      int          `json:"multiplier"`
	Wins            []CascadeWin `json:"wins"`
	TotalCascadeWin float64      `json:"total_cascade_win"`
	WinningTileKind string       `json:"winning_tile_kind,omitempty"` // Highest priority winning symbol (fa > zhong > bai > bawan), empty if no high-value win
}

// Position represents a grid position [reel, row]
type Position struct {
	Reel         int  `json:"reel"`
	Row          int  `json:"row"`
	IsGoldToWild bool `json:"is_gold_to_wild,omitempty"` // True if this gold tile transforms to wild
}

// CascadeWin represents a win in a cascade
type CascadeWin struct {
	Symbol    string     `json:"symbol"`
	Count     int        `json:"count"`
	Ways      int        `json:"ways"`
	Payout    float64    `json:"payout"`
	WinAmount float64    `json:"win_amount"`
	Positions []Position `json:"positions"` // Grid positions that form this win
}

// TableName specifies the table name for GORM
func (Spin) TableName() string {
	return "spins"
}

// Scan implements the sql.Scanner interface for Grid
func (g *Grid) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, g)
}

// Value implements the driver.Valuer interface for Grid
func (g Grid) Value() (driver.Value, error) {
	return json.Marshal(g)
}

// Scan implements the sql.Scanner interface for Cascades
func (c *Cascades) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Value implements the driver.Valuer interface for Cascades
func (c Cascades) Value() (driver.Value, error) {
	return json.Marshal(c)
}
