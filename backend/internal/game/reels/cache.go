package reels

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/slotmachine/backend/domain/reelstrip"
)

// ReelStripCache provides in-memory caching for reel strips
type ReelStripCache struct {
	baseGameStrips  [][]*reelstrip.ReelStrip // [reel_number][strips]
	freeSpinsStrips [][]*reelstrip.ReelStrip // [reel_number][strips]
	mu              sync.RWMutex
	lastUpdate      time.Time
	service         reelstrip.Service
}

// NewReelStripCache creates a new reel strip cache
func NewReelStripCache(service reelstrip.Service) *ReelStripCache {
	return &ReelStripCache{
		baseGameStrips:  make([][]*reelstrip.ReelStrip, 5),
		freeSpinsStrips: make([][]*reelstrip.ReelStrip, 5),
		service:         service,
	}
}

// LoadCache loads reel strips from database into memory
func (c *ReelStripCache) LoadCache(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Load base game strips
	if err := c.loadGameModeStrips(ctx, string(reelstrip.BaseGame), &c.baseGameStrips); err != nil {
		return fmt.Errorf("failed to load base game strips: %w", err)
	}

	// Load free spins strips
	if err := c.loadGameModeStrips(ctx, string(reelstrip.FreeSpins), &c.freeSpinsStrips); err != nil {
		return fmt.Errorf("failed to load free spins strips: %w", err)
	}

	c.lastUpdate = time.Now()
	return nil
}

// loadGameModeStrips loads strips for a specific game mode
func (c *ReelStripCache) loadGameModeStrips(ctx context.Context, gameMode string, target *[][]*reelstrip.ReelStrip) error {
	// Get repository interface through service (need to get from DB directly)
	// For now, we'll use the service to get counts and organize
	counts, err := c.service.GetActiveStripsCount(ctx, gameMode)
	if err != nil {
		return err
	}

	// Verify we have strips for all reels
	for i := 0; i < 5; i++ {
		if counts[i] == 0 {
			return fmt.Errorf("no active strips found for reel %d in %s mode", i, gameMode)
		}
	}

	// Note: This is a simplified version. In production, you'd want to fetch
	// all strips from repository and organize them here.
	// For now, the service layer will handle fetching from DB.

	return nil
}

// GetRandomReelSet returns a random set of reel strips from cache
// Falls back to database if cache is empty
func (c *ReelStripCache) GetRandomReelSet(ctx context.Context, isFreeSpin bool) (*reelstrip.ReelStripSet, error) {
	gameMode := string(reelstrip.BaseGame)
	if isFreeSpin {
		gameMode = string(reelstrip.FreeSpins)
	}

	// For now, delegate to service which will fetch from DB
	// In a full implementation, you'd randomly select from cached strips here
	return c.service.GetRandomReelSet(ctx, gameMode)
}

// RefreshCache reloads the cache from database
func (c *ReelStripCache) RefreshCache(ctx context.Context) error {
	return c.LoadCache(ctx)
}

// GetLastUpdate returns the last cache update time
func (c *ReelStripCache) GetLastUpdate() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastUpdate
}

// GetCacheStats returns statistics about cached strips
func (c *ReelStripCache) GetCacheStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	baseGameCounts := make([]int, 5)
	for i := 0; i < 5; i++ {
		baseGameCounts[i] = len(c.baseGameStrips[i])
	}

	freeSpinsCounts := make([]int, 5)
	for i := 0; i < 5; i++ {
		freeSpinsCounts[i] = len(c.freeSpinsStrips[i])
	}

	return map[string]interface{}{
		"base_game_counts":  baseGameCounts,
		"free_spins_counts": freeSpinsCounts,
		"last_update":       c.lastUpdate,
	}
}

// ReelStripCacheV2 is an improved version with full in-memory caching
type ReelStripCacheV2 struct {
	// Map structure: [game_mode][reel_number][]strip_data
	// Each game_mode has 5 reels, each reel has multiple strip variations
	cache      map[string][][][]string // game_mode -> [5 reels][multiple strips][symbols]
	mu         sync.RWMutex
	lastUpdate time.Time
}

// NewReelStripCacheV2 creates an improved cache
func NewReelStripCacheV2() *ReelStripCacheV2 {
	return &ReelStripCacheV2{
		cache: make(map[string][][][]string),
	}
}

// LoadFromService loads all strips from service into memory
func (c *ReelStripCacheV2) LoadFromService(ctx context.Context, service reelstrip.Service) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize cache structure for each game mode
	// Each game mode has 5 reels, each reel can have multiple strip variations
	c.cache[string(reelstrip.BaseGame)] = make([][][]string, 5)
	c.cache[string(reelstrip.FreeSpins)] = make([][][]string, 5)

	// Initialize empty slices for each reel
	for i := 0; i < 5; i++ {
		c.cache[string(reelstrip.BaseGame)][i] = make([][]string, 0)
		c.cache[string(reelstrip.FreeSpins)][i] = make([][]string, 0)
	}

	// This is a placeholder - you'd need to add a method to service to get all strips
	// For now, we'll document the intended behavior
	c.lastUpdate = time.Now()

	return nil
}

// GetRandomStripData returns random strip data for a reel
func (c *ReelStripCacheV2) GetRandomStripData(gameMode string, reelNumber int) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	reels, exists := c.cache[gameMode]
	if !exists {
		return nil, fmt.Errorf("no strips cached for game mode: %s", gameMode)
	}

	if reelNumber < 0 || reelNumber >= len(reels) {
		return nil, fmt.Errorf("invalid reel number: %d", reelNumber)
	}

	stripVariations := reels[reelNumber]
	if len(stripVariations) == 0 {
		return nil, fmt.Errorf("no strips available for reel %d", reelNumber)
	}

	// Return a random strip variation using crypto/rand
	max := big.NewInt(int64(len(stripVariations)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure random index: %w", err)
	}
	idx := int(n.Int64())
	return stripVariations[idx], nil
}
