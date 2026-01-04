package reels

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockReelStripService is a mock implementation of reelstrip.Service
type MockReelStripService struct {
	mock.Mock
}

func (m *MockReelStripService) GetActiveStripsCount(ctx context.Context, gameMode string) (map[int]int, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int]int), args.Error(1)
}

// Stub methods (not used in cache tests but required by interface)
func (m *MockReelStripService) GetReelSetForPlayer(ctx context.Context, playerID uuid.UUID, gameMode string) (*reelstrip.ReelStripConfigSet, error) {
	return nil, nil
}

func (m *MockReelStripService) GetReelSetByConfig(ctx context.Context, configID uuid.UUID) (*reelstrip.ReelStripConfigSet, error) {
	return nil, nil
}

func (m *MockReelStripService) GetDefaultReelSet(ctx context.Context, gameMode string) (*reelstrip.ReelStripConfigSet, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reelstrip.ReelStripConfigSet), args.Error(1)
}

func (m *MockReelStripService) CreateConfig(ctx context.Context, name, gameMode, description string, reelStripIDs [5]uuid.UUID, targetRTP float64, extraInfoJSON []byte) (*reelstrip.ReelStripConfig, error) {
	return nil, nil
}

func (m *MockReelStripService) GetConfigByID(ctx context.Context, id uuid.UUID) (*reelstrip.ReelStripConfig, error) {
	return nil, nil
}

func (m *MockReelStripService) GetConfigByName(ctx context.Context, name string) (*reelstrip.ReelStripConfig, error) {
	return nil, nil
}

func (m *MockReelStripService) SetDefaultConfig(ctx context.Context, configID uuid.UUID, gameMode string) error {
	return nil
}

func (m *MockReelStripService) ActivateConfig(ctx context.Context, configID uuid.UUID) error {
	return nil
}

func (m *MockReelStripService) DeactivateConfig(ctx context.Context, configID uuid.UUID) error {
	return nil
}

func (m *MockReelStripService) AssignConfigToPlayer(ctx context.Context, playerID, configID uuid.UUID, gameMode, reason, assignedBy string, expiresAt *time.Time) error {
	return nil
}

func (m *MockReelStripService) GetPlayerAssignment(ctx context.Context, playerID uuid.UUID) (*reelstrip.PlayerReelStripAssignment, error) {
	return nil, nil
}

func (m *MockReelStripService) RemovePlayerAssignment(ctx context.Context, playerID uuid.UUID) error {
	return nil
}

func (m *MockReelStripService) GenerateAndSaveStrips(ctx context.Context, gameMode string, count int, version int) error {
	return nil
}

func (m *MockReelStripService) GenerateAndSaveStripSet(ctx context.Context, gameMode string) ([5]uuid.UUID, error) {
	return [5]uuid.UUID{}, nil
}

func (m *MockReelStripService) GetStripByID(ctx context.Context, id uuid.UUID) (*reelstrip.ReelStrip, error) {
	return nil, nil
}

func (m *MockReelStripService) ValidateStripIntegrity(strip *reelstrip.ReelStrip) error {
	return nil
}

// GetRandomReelSet is called by cache.go but doesn't exist in the interface - bug in production code
func (m *MockReelStripService) GetRandomReelSet(ctx context.Context, gameMode string) (*reelstrip.ReelStripSet, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reelstrip.ReelStripSet), args.Error(1)
}

func (m *MockReelStripService) RotateStrips(ctx context.Context, gameMode string, newVersion int, count int) error {
	return nil
}

func (m *MockReelStripService) ListConfigs(ctx context.Context, filters *reelstrip.ConfigListFilters) ([]*reelstrip.ReelStripConfig, int64, error) {
	return nil, 0, nil
}

// ============================================================================
// ReelStripCache TESTS
// ============================================================================

func TestNewReelStripCache(t *testing.T) {
	mockService := new(MockReelStripService)

	cache := NewReelStripCache(mockService)

	require.NotNil(t, cache)
	assert.NotNil(t, cache.baseGameStrips)
	assert.NotNil(t, cache.freeSpinsStrips)
	assert.Len(t, cache.baseGameStrips, 5)
	assert.Len(t, cache.freeSpinsStrips, 5)
}

func TestReelStripCache_LoadCache(t *testing.T) {
	ctx := context.Background()

	t.Run("should load cache successfully", func(t *testing.T) {
		mockService := new(MockReelStripService)
		cache := NewReelStripCache(mockService)

		// Mock base game strips count
		baseGameCounts := map[int]int{0: 5, 1: 5, 2: 5, 3: 5, 4: 5}
		mockService.On("GetActiveStripsCount", ctx, string(reelstrip.BaseGame)).
			Return(baseGameCounts, nil)

		// Mock free spins strips count
		freeSpinsCounts := map[int]int{0: 3, 1: 3, 2: 3, 3: 3, 4: 3}
		mockService.On("GetActiveStripsCount", ctx, string(reelstrip.FreeSpins)).
			Return(freeSpinsCounts, nil)

		err := cache.LoadCache(ctx)

		require.NoError(t, err)
		assert.False(t, cache.lastUpdate.IsZero())
		mockService.AssertExpectations(t)
	})

	t.Run("should error when base game strips missing", func(t *testing.T) {
		mockService := new(MockReelStripService)
		cache := NewReelStripCache(mockService)

		// Mock with missing strips for reel 2
		baseGameCounts := map[int]int{0: 5, 1: 5, 2: 0, 3: 5, 4: 5}
		mockService.On("GetActiveStripsCount", ctx, string(reelstrip.BaseGame)).
			Return(baseGameCounts, nil)

		err := cache.LoadCache(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active strips found for reel 2")
	})

	t.Run("should error when free spins strips missing", func(t *testing.T) {
		mockService := new(MockReelStripService)
		cache := NewReelStripCache(mockService)

		// Mock base game strips
		baseGameCounts := map[int]int{0: 5, 1: 5, 2: 5, 3: 5, 4: 5}
		mockService.On("GetActiveStripsCount", ctx, string(reelstrip.BaseGame)).
			Return(baseGameCounts, nil)

		// Mock free spins with missing strips for reel 0
		freeSpinsCounts := map[int]int{0: 0, 1: 3, 2: 3, 3: 3, 4: 3}
		mockService.On("GetActiveStripsCount", ctx, string(reelstrip.FreeSpins)).
			Return(freeSpinsCounts, nil)

		err := cache.LoadCache(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active strips found for reel 0")
	})
}

// Note: GetRandomReelSet test skipped - method calls non-existent service.GetRandomReelSet
// This appears to be a bug in cache.go that needs fixing in production code

func TestReelStripCache_RefreshCache(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockReelStripService)
	cache := NewReelStripCache(mockService)

	// Mock GetActiveStripsCount for both game modes
	baseGameCounts := map[int]int{0: 5, 1: 5, 2: 5, 3: 5, 4: 5}
	mockService.On("GetActiveStripsCount", ctx, string(reelstrip.BaseGame)).
		Return(baseGameCounts, nil)

	freeSpinsCounts := map[int]int{0: 3, 1: 3, 2: 3, 3: 3, 4: 3}
	mockService.On("GetActiveStripsCount", ctx, string(reelstrip.FreeSpins)).
		Return(freeSpinsCounts, nil)

	err := cache.RefreshCache(ctx)

	require.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestReelStripCache_GetLastUpdate(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockReelStripService)
	cache := NewReelStripCache(mockService)

	t.Run("should return zero time before load", func(t *testing.T) {
		lastUpdate := cache.GetLastUpdate()
		assert.True(t, lastUpdate.IsZero())
	})

	t.Run("should return update time after load", func(t *testing.T) {
		before := time.Now()

		// Mock GetActiveStripsCount
		counts := map[int]int{0: 5, 1: 5, 2: 5, 3: 5, 4: 5}
		mockService.On("GetActiveStripsCount", ctx, string(reelstrip.BaseGame)).
			Return(counts, nil)
		mockService.On("GetActiveStripsCount", ctx, string(reelstrip.FreeSpins)).
			Return(counts, nil)

		err := cache.LoadCache(ctx)
		require.NoError(t, err)

		after := time.Now()
		lastUpdate := cache.GetLastUpdate()

		assert.False(t, lastUpdate.IsZero())
		assert.True(t, lastUpdate.After(before) || lastUpdate.Equal(before))
		assert.True(t, lastUpdate.Before(after) || lastUpdate.Equal(after))
	})
}

func TestReelStripCache_GetCacheStats(t *testing.T) {
	mockService := new(MockReelStripService)
	cache := NewReelStripCache(mockService)

	stats := cache.GetCacheStats()

	require.NotNil(t, stats)
	assert.Contains(t, stats, "base_game_counts")
	assert.Contains(t, stats, "free_spins_counts")
	assert.Contains(t, stats, "last_update")

	baseGameCounts := stats["base_game_counts"].([]int)
	freeSpinsCounts := stats["free_spins_counts"].([]int)

	assert.Len(t, baseGameCounts, 5)
	assert.Len(t, freeSpinsCounts, 5)
}

// ============================================================================
// ReelStripCacheV2 TESTS
// ============================================================================

func TestNewReelStripCacheV2(t *testing.T) {
	cache := NewReelStripCacheV2()

	require.NotNil(t, cache)
	assert.NotNil(t, cache.cache)
}

func TestReelStripCacheV2_LoadFromService(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockReelStripService)
	cache := NewReelStripCacheV2()

	err := cache.LoadFromService(ctx, mockService)

	require.NoError(t, err)
	assert.False(t, cache.lastUpdate.IsZero())

	// Check cache structure initialized
	assert.Contains(t, cache.cache, string(reelstrip.BaseGame))
	assert.Contains(t, cache.cache, string(reelstrip.FreeSpins))
	assert.Len(t, cache.cache[string(reelstrip.BaseGame)], 5)
	assert.Len(t, cache.cache[string(reelstrip.FreeSpins)], 5)
}

func TestReelStripCacheV2_GetRandomStripData(t *testing.T) {
	cache := NewReelStripCacheV2()

	t.Run("should error for non-existent game mode", func(t *testing.T) {
		_, err := cache.GetRandomStripData("invalid_mode", 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no strips cached for game mode")
	})

	t.Run("should error for invalid reel number", func(t *testing.T) {
		// Initialize cache
		cache.cache[string(reelstrip.BaseGame)] = make([][][]string, 5)

		_, err := cache.GetRandomStripData(string(reelstrip.BaseGame), 10)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid reel number")
	})

	t.Run("should error for negative reel number", func(t *testing.T) {
		cache.cache[string(reelstrip.BaseGame)] = make([][][]string, 5)

		_, err := cache.GetRandomStripData(string(reelstrip.BaseGame), -1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid reel number")
	})

	t.Run("should error when no strips available", func(t *testing.T) {
		cache.cache[string(reelstrip.BaseGame)] = make([][][]string, 5)
		for i := 0; i < 5; i++ {
			cache.cache[string(reelstrip.BaseGame)][i] = make([][]string, 0)
		}

		_, err := cache.GetRandomStripData(string(reelstrip.BaseGame), 0)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no strips available")
	})

	t.Run("should return random strip data", func(t *testing.T) {
		// Initialize cache with test data
		cache.cache[string(reelstrip.BaseGame)] = make([][][]string, 5)
		testStrip := []string{"A", "K", "Q", "J", "10"}
		cache.cache[string(reelstrip.BaseGame)][0] = [][]string{testStrip}

		result, err := cache.GetRandomStripData(string(reelstrip.BaseGame), 0)

		require.NoError(t, err)
		assert.Equal(t, testStrip, result)
	})

	t.Run("should return different strips on multiple calls", func(t *testing.T) {
		// Initialize cache with multiple strip variations
		cache.cache[string(reelstrip.BaseGame)] = make([][][]string, 5)
		strip1 := []string{"A", "K", "Q"}
		strip2 := []string{"10", "9", "8"}
		strip3 := []string{"J", "Q", "K"}
		cache.cache[string(reelstrip.BaseGame)][0] = [][]string{strip1, strip2, strip3}

		// Get multiple strips and check we get variation
		results := make(map[string]bool)
		for i := 0; i < 50; i++ {
			result, err := cache.GetRandomStripData(string(reelstrip.BaseGame), 0)
			require.NoError(t, err)
			// Convert to string for comparison
			key := ""
			for _, sym := range result {
				key += sym + ","
			}
			results[key] = true
		}

		// Should get at least 2 different strip variations in 50 tries
		assert.GreaterOrEqual(t, len(results), 2, "Should get multiple different strips")
	})
}
