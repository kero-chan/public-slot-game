package engine

import (
	"github.com/google/wire"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/internal/pkg/cache"
)

// ProviderSet is the Wire provider set for game engine
var ProviderSet = wire.NewSet(
	ProvideGameEngine,
)

// ProvideGameEngine creates a new game engine with DB support
// This is the recommended implementation that uses reel strips from database
func ProvideGameEngine(cache *cache.Cache, reelStripService reelstrip.Service) *GameEngine {
	// Enable DB strips by default (set to true)
	useDBStrips := true
	return NewGameEngine(reelStripService, cache, useDBStrips)
}
