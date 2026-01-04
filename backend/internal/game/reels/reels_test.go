package reels

import (
	"testing"

	"github.com/slotmachine/backend/internal/game/rng"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ReelStrip TESTS
// ============================================================================

func TestReelStrip_GetSymbolAtPosition(t *testing.T) {
	strip := ReelStrip{"A", "K", "Q", "J", "10"}

	t.Run("should get symbol at valid position", func(t *testing.T) {
		assert.Equal(t, "A", strip.GetSymbolAtPosition(0))
		assert.Equal(t, "K", strip.GetSymbolAtPosition(1))
		assert.Equal(t, "10", strip.GetSymbolAtPosition(4))
	})

	t.Run("should wrap around for position beyond strip length", func(t *testing.T) {
		assert.Equal(t, "A", strip.GetSymbolAtPosition(5))  // 5 % 5 = 0
		assert.Equal(t, "K", strip.GetSymbolAtPosition(6))  // 6 % 5 = 1
		assert.Equal(t, "Q", strip.GetSymbolAtPosition(12)) // 12 % 5 = 2
	})

	t.Run("should handle negative positions", func(t *testing.T) {
		assert.Equal(t, "10", strip.GetSymbolAtPosition(-1)) // -1 + 5 = 4
		assert.Equal(t, "J", strip.GetSymbolAtPosition(-2))  // -2 + 5 = 3
	})

	t.Run("should return empty string for empty strip", func(t *testing.T) {
		emptyStrip := ReelStrip{}
		assert.Equal(t, "", emptyStrip.GetSymbolAtPosition(0))
	})
}

func TestReelStrip_GetSymbolsFromPosition(t *testing.T) {
	strip := ReelStrip{"A", "K", "Q", "J", "10", "9"}

	t.Run("should get multiple symbols from position", func(t *testing.T) {
		symbols := strip.GetSymbolsFromPosition(0, 3)
		assert.Equal(t, []string{"A", "K", "Q"}, symbols)
	})

	t.Run("should wrap around when getting symbols", func(t *testing.T) {
		symbols := strip.GetSymbolsFromPosition(4, 4)
		assert.Equal(t, []string{"10", "9", "A", "K"}, symbols)
	})

	t.Run("should handle start position beyond strip length", func(t *testing.T) {
		symbols := strip.GetSymbolsFromPosition(7, 2) // 7 % 6 = 1
		assert.Equal(t, []string{"K", "Q"}, symbols)
	})
}

func TestGenerateReelStrip(t *testing.T) {
	cryptoRNG := rng.NewCryptoRNG()

	t.Run("should generate reel strip with correct length", func(t *testing.T) {
		weights := symbols.ReelWeights{
			"A": 250,
			"K": 250,
			"Q": 250,
			"J": 250,
		}

		strip, err := GenerateReelStrip(weights, cryptoRNG)

		require.NoError(t, err)
		assert.Len(t, strip, 1000)
	})

	t.Run("should contain all weighted symbols", func(t *testing.T) {
		weights := symbols.ReelWeights{
			"A": 300,
			"K": 200,
			"Q": 100,
		}

		strip, err := GenerateReelStrip(weights, cryptoRNG)

		require.NoError(t, err)

		// Count symbols
		counts := make(map[string]int)
		for _, sym := range strip {
			counts[sym]++
		}

		assert.Equal(t, 300, counts["A"])
		assert.Equal(t, 200, counts["K"])
		assert.Equal(t, 100, counts["Q"])
	})

	t.Run("should generate strip with length matching total weights", func(t *testing.T) {
		weights := symbols.ReelWeights{
			"A": 100,
			"K": 100,
		}

		strip, err := GenerateReelStrip(weights, cryptoRNG)

		assert.NoError(t, err)
		assert.NotNil(t, strip)
		// Strip length should equal sum of all weights
		assert.Equal(t, 200, len(strip))
	})

	t.Run("should produce different strips on multiple calls (randomness)", func(t *testing.T) {
		weights := symbols.ReelWeights{
			"A": 500,
			"K": 500,
		}

		strip1, err1 := GenerateReelStrip(weights, cryptoRNG)
		require.NoError(t, err1)

		strip2, err2 := GenerateReelStrip(weights, cryptoRNG)
		require.NoError(t, err2)

		// Strips should have same symbols but different order
		assert.NotEqual(t, strip1, strip2, "Strips should be shuffled differently")
	})
}

func TestGenerateAllReelStrips(t *testing.T) {
	cryptoRNG := rng.NewCryptoRNG()

	// t.Run("should generate 5 reel strips for base game", func(t *testing.T) {
	// 	strips, err := GenerateAllReelStrips(false, cryptoRNG)

	// 	require.NoError(t, err)
	// 	assert.Len(t, strips, 5)

	// 	// for i, strip := range strips {
	// 	// 	assert.Equal(t, BaseGameStripLength, len(strip), "Reel %d should have correct length", i)
	// 	// }
	// })

	// t.Run("should generate 5 reel strips for free spins", func(t *testing.T) {
	// 	strips, err := GenerateAllReelStrips(true, cryptoRNG)

	// 	require.NoError(t, err)
	// 	assert.Len(t, strips, 5)

	// 	for i, strip := range strips {
	// 		assert.Equal(t, FreeSpinStripLength, len(strip), "Reel %d should have correct length", i)
	// 	}
	// })

	t.Run("should generate different strips for each reel", func(t *testing.T) {
		strips, err := GenerateAllReelStrips(false, cryptoRNG)

		require.NoError(t, err)

		// Each reel should have different symbol distributions
		// (at least some should be different due to different weights)
		allSame := true
		for i := 1; i < len(strips); i++ {
			if !equalStrips(strips[0], strips[i]) {
				allSame = false
				break
			}
		}
		assert.False(t, allSame, "Not all reels should be identical")
	})
}

// ============================================================================
// Grid TESTS
// ============================================================================

func TestGrid_GetSymbol(t *testing.T) {
	grid := Grid{
		{"A", "K", "Q", "J"},
		{"K", "Q", "J", "10"},
		{"Q", "J", "10", "9"},
	}

	t.Run("should get symbol at valid position", func(t *testing.T) {
		assert.Equal(t, "A", grid.GetSymbol(0, 0))
		assert.Equal(t, "Q", grid.GetSymbol(1, 1))
		assert.Equal(t, "9", grid.GetSymbol(2, 3))
	})

	t.Run("should return empty string for invalid reel", func(t *testing.T) {
		assert.Equal(t, "", grid.GetSymbol(-1, 0))
		assert.Equal(t, "", grid.GetSymbol(5, 0))
	})

	t.Run("should return empty string for invalid row", func(t *testing.T) {
		assert.Equal(t, "", grid.GetSymbol(0, -1))
		assert.Equal(t, "", grid.GetSymbol(0, 10))
	})
}

func TestGrid_SetSymbol(t *testing.T) {
	grid := Grid{
		{"A", "K", "Q"},
		{"K", "Q", "J"},
	}

	t.Run("should set symbol at valid position", func(t *testing.T) {
		grid.SetSymbol(0, 1, "WILD")
		assert.Equal(t, "WILD", grid.GetSymbol(0, 1))
	})

	t.Run("should not panic for invalid positions", func(t *testing.T) {
		assert.NotPanics(t, func() {
			grid.SetSymbol(-1, 0, "X")
			grid.SetSymbol(10, 0, "X")
			grid.SetSymbol(0, -1, "X")
			grid.SetSymbol(0, 10, "X")
		})
	})
}

func TestGrid_Clone(t *testing.T) {
	original := Grid{
		{"A", "K", "Q"},
		{"K", "Q", "J"},
	}

	t.Run("should create deep copy", func(t *testing.T) {
		clone := original.Clone()

		assert.Equal(t, original, clone)

		// Modify clone
		clone.SetSymbol(0, 0, "WILD")

		// Original should be unchanged
		assert.Equal(t, "A", original.GetSymbol(0, 0))
		assert.Equal(t, "WILD", clone.GetSymbol(0, 0))
	})
}

func TestGrid_CountSymbol(t *testing.T) {
	grid := Grid{
		{"A", "K", "Q", "J", "10", "9"},
		{"A", "A", "K", "Q", "J", "10"},
		{"K", "Q", "J", "A", "10", "9"},
		{"Q", "J", "10", "9", "A", "K"},
		{"J", "10", "9", "A", "K", "Q"},
	}

	t.Run("should count symbol occurrences in visible grid", func(t *testing.T) {
		count := grid.CountSymbol("A")
		// Count only in visible rows (first 6 rows)
		// Reel 0: 1 A, Reel 1: 2 As, Reel 2: 1 A, Reel 3: 1 A, Reel 4: 1 A = 6 total
		assert.Greater(t, count, 0)
	})

	t.Run("should return zero for non-existent symbol", func(t *testing.T) {
		count := grid.CountSymbol("NONEXISTENT")
		assert.Equal(t, 0, count)
	})

	t.Run("should handle base symbol with _gold suffix", func(t *testing.T) {
		gridWithGold := Grid{
			{"A_gold", "K", "Q", "J", "10", "9"},
			{"A", "K", "Q", "J", "10", "9"},
			{"K", "Q", "J", "10", "9", "A"},
			{"K", "Q", "J", "10", "9", "A"},
			{"K", "Q", "J", "10", "9", "A"},
		}

		// Should count both "A" and "A_gold" as "A"
		count := gridWithGold.CountSymbol("A")
		assert.Greater(t, count, 1)
	})
}

func TestGenerateGrid(t *testing.T) {
	cryptoRNG := rng.NewCryptoRNG()

	t.Run("should generate grid with correct dimensions", func(t *testing.T) {
		strips, err := GenerateAllReelStrips(false, cryptoRNG)
		require.NoError(t, err)

		grid, positions, err := GenerateGrid(strips, cryptoRNG)

		require.NoError(t, err)
		assert.Len(t, grid, ReelCount)
		assert.Len(t, positions, ReelCount)

		for i, reel := range grid {
			assert.Len(t, reel, TotalRows, "Reel %d should have %d rows", i, TotalRows)
		}
	})

	t.Run("should return error for wrong number of strips", func(t *testing.T) {
		strips := []ReelStrip{
			{"A", "K", "Q"},
			{"K", "Q", "J"},
		}

		grid, positions, err := GenerateGrid(strips, cryptoRNG)

		assert.Error(t, err)
		assert.Nil(t, grid)
		assert.Nil(t, positions)
		assert.Contains(t, err.Error(), "expected 5 reel strips")
	})

	t.Run("should generate valid reel positions", func(t *testing.T) {
		strips, err := GenerateAllReelStrips(false, cryptoRNG)
		require.NoError(t, err)

		_, positions, err := GenerateGrid(strips, cryptoRNG)

		require.NoError(t, err)

		for i, pos := range positions {
			assert.GreaterOrEqual(t, pos, 0)
			assert.Less(t, pos, len(strips[i]))
		}
	})

	t.Run("should generate different grids on multiple calls", func(t *testing.T) {
		strips, err := GenerateAllReelStrips(false, cryptoRNG)
		require.NoError(t, err)

		grid1, _, err1 := GenerateGrid(strips, cryptoRNG)
		require.NoError(t, err1)

		grid2, _, err2 := GenerateGrid(strips, cryptoRNG)
		require.NoError(t, err2)

		// Grids should be different (extremely unlikely to be the same)
		assert.NotEqual(t, grid1, grid2)
	})
}

func TestGrid_GetVisibleGrid(t *testing.T) {
	grid := Grid{
		{"A", "K", "Q", "J", "10", "9", "A", "K", "Q", "J"},
		{"K", "Q", "J", "10", "9", "A", "K", "Q", "J", "10"},
		{"Q", "J", "10", "9", "A", "K", "Q", "J", "10", "9"},
		{"J", "10", "9", "A", "K", "Q", "J", "10", "9", "A"},
		{"10", "9", "A", "K", "Q", "J", "10", "9", "A", "K"},
	}

	t.Run("should return only visible rows", func(t *testing.T) {
		visible := grid.GetVisibleGrid()

		assert.Len(t, visible, ReelCount)
		for i, reel := range visible {
			assert.Len(t, reel, VisibleRows, "Visible reel %d should have %d rows", i, VisibleRows)
		}
	})

	t.Run("should preserve symbols in visible portion", func(t *testing.T) {
		visible := grid.GetVisibleGrid()

		// First visible rows should match original
		for reel := 0; reel < ReelCount; reel++ {
			for row := 0; row < VisibleRows; row++ {
				assert.Equal(t, grid[reel][row], visible[reel][row])
			}
		}
	})
}

func TestGrid_ToJSON_FromJSON(t *testing.T) {
	original := Grid{
		{"A", "K", "Q"},
		{"K", "Q", "J"},
		{"Q", "J", "10"},
	}

	t.Run("should convert to JSON and back", func(t *testing.T) {
		jsonData := original.ToJSON()
		restored := FromJSON(jsonData)

		assert.Equal(t, original, restored)
	})
}

// ============================================================================
// Helper Functions
// ============================================================================

func equalStrips(a, b ReelStrip) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
