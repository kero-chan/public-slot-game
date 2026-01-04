package game

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Game represents a game definition
type Game struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description"`
	DevURL      *string   `gorm:"column:dev_url;type:text" json:"dev_url"`
	ProdURL     *string   `gorm:"column:prod_url;type:text" json:"prod_url"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Game) TableName() string {
	return "games"
}

// Asset represents an asset set
type Asset struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name            string          `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description     *string         `gorm:"type:text" json:"description"`
	ObjectName      string          `gorm:"column:object_name;type:varchar(100);uniqueIndex;not null" json:"object_name"`
	BaseURL         string          `gorm:"column:base_url;type:text;not null" json:"base_url"`
	SpritesheetJSON json.RawMessage `gorm:"column:spritesheet_json;type:jsonb;not null" json:"spritesheet_json"`
	Images          json.RawMessage `gorm:"type:jsonb;not null" json:"images"`
	Audios          json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"audios"`
	Videos          json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"videos"`
	IsActive        bool            `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time       `gorm:"default:now()" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"default:now()" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Asset) TableName() string {
	return "assets"
}

// GetBaseURL computes the base URL for this asset from storage config
// Format: {publicURL}/{bucketName}/{objectName}
func (a *Asset) GetBaseURL(publicURL, bucketName string) string {
	publicURL = strings.TrimSuffix(publicURL, "/")
	return fmt.Sprintf("%s/%s/%s", publicURL, bucketName, a.ObjectName)
}

// GenerateObjectName creates a unique object name from asset name
// Format: lowercase-name-with-dashes-randomsuffix
func GenerateObjectName(assetName string) string {
	// Convert to lowercase
	name := strings.ToLower(assetName)

	// Replace spaces and special characters with dashes
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	name = reg.ReplaceAllString(name, "-")

	// Remove leading/trailing dashes
	name = strings.Trim(name, "-")

	// Generate random suffix (6 characters)
	suffix := make([]byte, 3)
	rand.Read(suffix)
	randomSuffix := hex.EncodeToString(suffix)

	return fmt.Sprintf("%s-%s", name, randomSuffix)
}

// GameConfig represents the configuration linking a game to its assets
type GameConfig struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	GameID    uuid.UUID `gorm:"type:uuid;not null" json:"game_id"`
	AssetID   uuid.UUID `gorm:"type:uuid;not null" json:"asset_id"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	// Relations
	Game  *Game  `gorm:"foreignKey:GameID" json:"game,omitempty"`
	Asset *Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

// TableName specifies the table name for GORM
func (GameConfig) TableName() string {
	return "game_configs"
}

// GameAssetsResponse is the API response for game assets
type GameAssetsResponse struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	SpritesheetJSON json.RawMessage   `json:"spritesheetJson"`
	Images          map[string]string `json:"images"`
	Audios          map[string]any    `json:"audios"`
	Videos          map[string]any    `json:"videos"`
}

// SymbolVideos represents win and loop videos for a symbol
type SymbolVideos struct {
	Win  []string `json:"win"`
	Loop []string `json:"loop"`
}

// ImageURLs represents the image URLs in the assets
type ImageURLs struct {
	Backgrounds      string `json:"backgrounds"`
	Glyphs           string `json:"glyphs"`
	Icons            string `json:"icons"`
	Tiles            string `json:"tiles"`
	WinAnnouncements string `json:"winAnnouncements"`
	BackgroundMain   string `json:"backgroundMain"`
	BackgroundStart  string `json:"backgroundStart"`
	StartBtn         string `json:"startBtn"`
	Preparing        string `json:"preparing"`
	PreparingSound   string `json:"preparingSound"`
	LoadingResources string `json:"loadingResources"`
	LoadingComplete  string `json:"loadingComplete"`
	IconAvatar       string `json:"iconAvatar"`
	IconExit         string `json:"iconExit"`
	IconHistory      string `json:"iconHistory"`
}
