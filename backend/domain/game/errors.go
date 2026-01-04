package game

import "errors"

var (
	// ErrGameNotFound is returned when a game is not found
	ErrGameNotFound = errors.New("game not found")

	// ErrAssetNotFound is returned when an asset is not found
	ErrAssetNotFound = errors.New("asset not found")

	// ErrGameConfigNotFound is returned when a game config is not found
	ErrGameConfigNotFound = errors.New("game config not found")

	// ErrNoActiveConfig is returned when no active config exists for a game
	ErrNoActiveConfig = errors.New("no active asset configuration for this game")

	// ErrInvalidGameID is returned when the game ID format is invalid
	ErrInvalidGameID = errors.New("invalid game ID format")
)
