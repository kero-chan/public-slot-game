package game

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// GameUpdate represents updatable fields for a game
type GameUpdate struct {
	Name        *string
	Description *string
	DevURL      *string
	ProdURL     *string
	IsActive    *bool
}

// AssetUpdate represents updatable fields for an asset
type AssetUpdate struct {
	Name            *string
	Description     *string
	ObjectName      *string
	BaseURL         *string
	SpritesheetJSON json.RawMessage
	Images          json.RawMessage
	Audios          json.RawMessage
	Videos          json.RawMessage
	IsActive        *bool
}

// Repository defines the interface for game data access
type Repository interface {
	// Game methods
	GetGameByID(ctx context.Context, id uuid.UUID) (*Game, error)
	GetGamesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*Game, error)
	ListGames(ctx context.Context, page, pageSize int) ([]*Game, int64, error)
	CreateGame(ctx context.Context, g *Game) error
	UpdateGame(ctx context.Context, id uuid.UUID, update *GameUpdate) (*Game, error)
	DeleteGame(ctx context.Context, id uuid.UUID) error

	// Asset methods
	GetAssetByID(ctx context.Context, id uuid.UUID) (*Asset, error)
	ListAssets(ctx context.Context, page, pageSize int) ([]*Asset, int64, error)
	CreateAsset(ctx context.Context, a *Asset) error
	UpdateAsset(ctx context.Context, id uuid.UUID, update *AssetUpdate) (*Asset, error)
	DeleteAsset(ctx context.Context, id uuid.UUID) error

	// GameConfig methods
	GetActiveAssetForGame(ctx context.Context, gameID uuid.UUID) (*Asset, error)
	GetGameConfigByID(ctx context.Context, id uuid.UUID) (*GameConfig, error)
	ListGameConfigs(ctx context.Context, page, pageSize int) ([]*GameConfig, int64, error)
	CreateGameConfig(ctx context.Context, c *GameConfig) error
	DeleteGameConfig(ctx context.Context, id uuid.UUID) error
	ActivateGameConfig(ctx context.Context, id uuid.UUID) (*GameConfig, error)
	DeactivateGameConfig(ctx context.Context, id uuid.UUID) (*GameConfig, error)
}
