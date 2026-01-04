package cascade

import (
	"testing"

	"github.com/slotmachine/backend/internal/game/multiplier"
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/rng"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/slotmachine/backend/internal/game/wins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// removeWinningSymbols TESTS
// ============================================================================

func TestRemoveWinningSymbols(t *testing.T) {
	t.Run("should remove all winning positions", func(t *testing.T) {
		// Create a grid with 3 "fa" symbols on first 3 reels
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		symbolWins := []wins.SymbolWin{
			{
				Symbol: symbols.SymbolFa,
				Count:  3,
				Ways:   1,
			},
		}

		result := removeWinningSymbols(grid, symbolWins)

		// Check that winning symbols in visible rows (5-8) are removed
		// Rows 5-8 should have empty symbols where "fa" was
		assert.Equal(t, EmptySymbol, result.GetSymbol(0, 5))
		assert.Equal(t, EmptySymbol, result.GetSymbol(0, 6))
		assert.Equal(t, EmptySymbol, result.GetSymbol(0, 7))

		assert.Equal(t, EmptySymbol, result.GetSymbol(1, 5))
		assert.Equal(t, EmptySymbol, result.GetSymbol(1, 6))
		assert.Equal(t, EmptySymbol, result.GetSymbol(1, 7))

		assert.Equal(t, EmptySymbol, result.GetSymbol(2, 5))
		assert.Equal(t, EmptySymbol, result.GetSymbol(2, 6))
		assert.Equal(t, EmptySymbol, result.GetSymbol(2, 7))
	})

	t.Run("should preserve non-winning symbols", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		symbolWins := []wins.SymbolWin{
			{
				Symbol: symbols.SymbolFa,
				Count:  3,
				Ways:   1,
			},
		}

		result := removeWinningSymbols(grid, symbolWins)

		// Non-winning symbols on reel 3 and 4 should be preserved
		assert.Equal(t, "zhong", result.GetSymbol(3, 5))
		assert.Equal(t, "liangtong", result.GetSymbol(3, 6))
		assert.Equal(t, "cai", result.GetSymbol(3, 7))

		assert.Equal(t, "zhong", result.GetSymbol(4, 5))
		assert.Equal(t, "liangtong", result.GetSymbol(4, 6))
		assert.Equal(t, "cai", result.GetSymbol(4, 7))
	})

	t.Run("should convert gold variants to wild", func(t *testing.T) {
		// Create a grid with gold variants
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa_gold", "fa_gold", "fa_gold", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa_gold", "fa_gold", "fa_gold", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa_gold", "fa_gold", "fa_gold", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		symbolWins := []wins.SymbolWin{
			{
				Symbol: symbols.SymbolFa, // Base symbol
				Count:  3,
				Ways:   1,
			},
		}

		result := removeWinningSymbols(grid, symbolWins)

		// Gold variants should be converted to wild, not removed
		assert.Equal(t, string(symbols.SymbolWild), result.GetSymbol(0, 5))
		assert.Equal(t, string(symbols.SymbolWild), result.GetSymbol(0, 6))
		assert.Equal(t, string(symbols.SymbolWild), result.GetSymbol(0, 7))

		assert.Equal(t, string(symbols.SymbolWild), result.GetSymbol(1, 5))
		assert.Equal(t, string(symbols.SymbolWild), result.GetSymbol(1, 6))
		assert.Equal(t, string(symbols.SymbolWild), result.GetSymbol(1, 7))
	})

	t.Run("should handle multiple winning symbols", func(t *testing.T) {
		// Grid with two different winning symbols - fa in first 3 reels
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bai", "wusuo", "wutong", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bai", "wusuo", "wutong", "fu", "shu"},
		}

		symbolWins := []wins.SymbolWin{
			{
				Symbol: symbols.SymbolFa,
				Count:  3,
				Ways:   1,
			},
		}

		result := removeWinningSymbols(grid, symbolWins)

		// Winning "fa" symbols in visible rows (5-8) should be removed in first 3 reels
		assert.Equal(t, EmptySymbol, result.GetSymbol(0, 5))
		assert.Equal(t, EmptySymbol, result.GetSymbol(0, 6))
		assert.Equal(t, EmptySymbol, result.GetSymbol(0, 7))

		assert.Equal(t, EmptySymbol, result.GetSymbol(1, 5))
		assert.Equal(t, EmptySymbol, result.GetSymbol(1, 6))
		assert.Equal(t, EmptySymbol, result.GetSymbol(1, 7))

		assert.Equal(t, EmptySymbol, result.GetSymbol(2, 5))
		assert.Equal(t, EmptySymbol, result.GetSymbol(2, 6))
		assert.Equal(t, EmptySymbol, result.GetSymbol(2, 7))

		// Reel 3 and 4 should not be affected
		assert.NotEqual(t, EmptySymbol, result.GetSymbol(3, 5))
		assert.NotEqual(t, EmptySymbol, result.GetSymbol(4, 5))
	})

	t.Run("should handle overlapping wins", func(t *testing.T) {
		// Grid where same position might be counted in multiple wins
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "fu", "shu"},
		}

		symbolWins := []wins.SymbolWin{
			{
				Symbol: symbols.SymbolFa,
				Count:  5, // 5-of-a-kind
				Ways:   1,
			},
		}

		result := removeWinningSymbols(grid, symbolWins)

		// All winning positions should be removed
		for reel := 0; reel < 5; reel++ {
			assert.Equal(t, EmptySymbol, result.GetSymbol(reel, 5))
			assert.Equal(t, EmptySymbol, result.GetSymbol(reel, 6))
			assert.Equal(t, EmptySymbol, result.GetSymbol(reel, 7))
		}
	})

	t.Run("should not modify original grid", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		originalSymbol := grid.GetSymbol(0, 5)

		symbolWins := []wins.SymbolWin{
			{
				Symbol: symbols.SymbolFa,
				Count:  3,
				Ways:   1,
			},
		}

		_ = removeWinningSymbols(grid, symbolWins)

		// Original grid should be unchanged
		assert.Equal(t, originalSymbol, grid.GetSymbol(0, 5))
	})
}

// ============================================================================
// dropSymbols TESTS
// ============================================================================

func TestDropSymbols(t *testing.T) {
	t.Run("should drop symbols down to fill empty spaces", func(t *testing.T) {
		// Create a grid with empty symbols in middle (after removal)
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", EmptySymbol, EmptySymbol, EmptySymbol, "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", EmptySymbol, "fa", "bai", "cai", "fu"},
			{"cai", "fu", "shu", EmptySymbol, EmptySymbol, "fa", "zhong", "bai", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		result := dropSymbols(grid)

		// Reel 0: 3 empty symbols should be at top, others dropped
		assert.Equal(t, EmptySymbol, result.GetSymbol(0, 0))
		assert.Equal(t, EmptySymbol, result.GetSymbol(0, 1))
		assert.Equal(t, EmptySymbol, result.GetSymbol(0, 2))
		assert.Equal(t, "fu", result.GetSymbol(0, 9)) // Bottom row
		assert.Equal(t, "cai", result.GetSymbol(0, 8))
		assert.Equal(t, "fa", result.GetSymbol(0, 7))

		// Reel 1: 1 empty symbol should be at top
		assert.Equal(t, EmptySymbol, result.GetSymbol(1, 0))
		assert.Equal(t, "fu", result.GetSymbol(1, 9)) // Bottom row
		assert.Equal(t, "cai", result.GetSymbol(1, 8))
		assert.Equal(t, "bai", result.GetSymbol(1, 7))

		// Reel 2: 2 empty symbols should be at top
		assert.Equal(t, EmptySymbol, result.GetSymbol(2, 0))
		assert.Equal(t, EmptySymbol, result.GetSymbol(2, 1))
		assert.Equal(t, "fu", result.GetSymbol(2, 9)) // Bottom row
	})

	t.Run("should maintain column integrity", func(t *testing.T) {
		grid := reels.Grid{
			{"A", EmptySymbol, EmptySymbol, "B", "C", "D", "E", "F", "G", "H"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		result := dropSymbols(grid)

		// Count non-empty symbols in reel 0
		nonEmptyCount := 0
		symbols := make([]string, 0)
		for row := 0; row < reels.TotalRows; row++ {
			sym := result.GetSymbol(0, row)
			if sym != EmptySymbol {
				nonEmptyCount++
				symbols = append(symbols, sym)
			}
		}

		// Should have 8 non-empty symbols (10 total - 2 empty)
		assert.Equal(t, 8, nonEmptyCount)

		// Non-empty symbols should be at bottom
		assert.Equal(t, "H", result.GetSymbol(0, 9))
		assert.Equal(t, "G", result.GetSymbol(0, 8))
		assert.Equal(t, "F", result.GetSymbol(0, 7))
		assert.Equal(t, "E", result.GetSymbol(0, 6))
		assert.Equal(t, "D", result.GetSymbol(0, 5))
		assert.Equal(t, "C", result.GetSymbol(0, 4))
		assert.Equal(t, "B", result.GetSymbol(0, 3))
		assert.Equal(t, "A", result.GetSymbol(0, 2))
	})

	t.Run("should handle all empty reel", func(t *testing.T) {
		grid := reels.Grid{
			{EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		result := dropSymbols(grid)

		// All positions should be empty
		for row := 0; row < reels.TotalRows; row++ {
			assert.Equal(t, EmptySymbol, result.GetSymbol(0, row))
		}
	})

	t.Run("should handle no empty symbols", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		result := dropSymbols(grid)

		// Grid should be unchanged
		for reel := 0; reel < reels.ReelCount; reel++ {
			for row := 0; row < reels.TotalRows; row++ {
				assert.Equal(t, grid.GetSymbol(reel, row), result.GetSymbol(reel, row))
			}
		}
	})

	t.Run("should process each reel independently", func(t *testing.T) {
		grid := reels.Grid{
			{"A", EmptySymbol, "B", "C", "D", "E", "F", "G", "H", "I"},
			{"J", "K", "L", EmptySymbol, "M", "N", "O", "P", "Q", "R"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		result := dropSymbols(grid)

		// Reel 0: 1 empty at top
		assert.Equal(t, EmptySymbol, result.GetSymbol(0, 0))
		assert.Equal(t, "I", result.GetSymbol(0, 9))

		// Reel 1: 1 empty at top
		assert.Equal(t, EmptySymbol, result.GetSymbol(1, 0))
		assert.Equal(t, "R", result.GetSymbol(1, 9))

		// Reel 2: No empty symbols
		assert.NotEqual(t, EmptySymbol, result.GetSymbol(2, 0))
	})

	t.Run("should not modify original grid", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", EmptySymbol, EmptySymbol, "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		originalSymbol := grid.GetSymbol(0, 1)

		_ = dropSymbols(grid)

		// Original grid should be unchanged
		assert.Equal(t, originalSymbol, grid.GetSymbol(0, 1))
		assert.Equal(t, EmptySymbol, grid.GetSymbol(0, 1))
	})
}

// ============================================================================
// fillEmptyPositions TESTS
// ============================================================================

func TestFillEmptyPositions(t *testing.T) {
	t.Run("should fill empty positions from reel strips", func(t *testing.T) {
		// Create grid with 3 empty positions at top of reel 0
		grid := reels.Grid{
			{EmptySymbol, EmptySymbol, EmptySymbol, "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		// Create simple reel strips
		reelStrips := make([]reels.ReelStrip, 5)
		for i := 0; i < 5; i++ {
			strip := make(reels.ReelStrip, 100)
			for j := 0; j < 100; j++ {
				strip[j] = "new_symbol"
			}
			reelStrips[i] = strip
		}

		reelPositions := []int{50, 50, 50, 50, 50}

		result, newPositions := fillEmptyPositions(grid, reelStrips, reelPositions)

		// Top 3 positions of reel 0 should be filled
		assert.NotEqual(t, EmptySymbol, result.GetSymbol(0, 0))
		assert.NotEqual(t, EmptySymbol, result.GetSymbol(0, 1))
		assert.NotEqual(t, EmptySymbol, result.GetSymbol(0, 2))

		// Should be filled with new symbols from reel strip
		assert.Equal(t, "new_symbol", result.GetSymbol(0, 0))
		assert.Equal(t, "new_symbol", result.GetSymbol(0, 1))
		assert.Equal(t, "new_symbol", result.GetSymbol(0, 2))

		// Reel position should be updated (moved backwards by 3)
		assert.Equal(t, 47, newPositions[0])

		// Other reels should not change position
		assert.Equal(t, 50, newPositions[1])
		assert.Equal(t, 50, newPositions[2])
	})

	t.Run("should handle wrap-around of reel position", func(t *testing.T) {
		grid := reels.Grid{
			{EmptySymbol, EmptySymbol, "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		// Create reel strips
		reelStrips := make([]reels.ReelStrip, 5)
		for i := 0; i < 5; i++ {
			strip := make(reels.ReelStrip, 100)
			for j := 0; j < 100; j++ {
				strip[j] = "symbol"
			}
			reelStrips[i] = strip
		}

		reelPositions := []int{0, 50, 50, 50, 50} // Position 0 should wrap around

		_, newPositions := fillEmptyPositions(grid, reelStrips, reelPositions)

		// Position 0 - 2 should wrap to 98 (0 - 2 + 100) % 100
		assert.Equal(t, 98, newPositions[0])
	})

	t.Run("should not modify reels with no empty positions", func(t *testing.T) {
		grid := reels.Grid{
			{EmptySymbol, EmptySymbol, "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		reelStrips := make([]reels.ReelStrip, 5)
		for i := 0; i < 5; i++ {
			strip := make(reels.ReelStrip, 100)
			for j := 0; j < 100; j++ {
				strip[j] = "new_symbol"
			}
			reelStrips[i] = strip
		}

		reelPositions := []int{50, 50, 50, 50, 50}
		originalReelSymbol := grid.GetSymbol(1, 0)

		result, newPositions := fillEmptyPositions(grid, reelStrips, reelPositions)

		// Reels without empty symbols should be unchanged
		assert.Equal(t, originalReelSymbol, result.GetSymbol(1, 0))
		assert.Equal(t, 50, newPositions[1]) // Position unchanged
	})

	t.Run("should handle all empty reel", func(t *testing.T) {
		grid := reels.Grid{
			{EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol, EmptySymbol},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		reelStrips := make([]reels.ReelStrip, 5)
		for i := 0; i < 5; i++ {
			strip := make(reels.ReelStrip, 100)
			for j := 0; j < 100; j++ {
				strip[j] = "filled"
			}
			reelStrips[i] = strip
		}

		reelPositions := []int{50, 50, 50, 50, 50}

		result, newPositions := fillEmptyPositions(grid, reelStrips, reelPositions)

		// All positions should be filled
		for row := 0; row < reels.TotalRows; row++ {
			assert.Equal(t, "filled", result.GetSymbol(0, row))
		}

		// Position should move back by 10
		assert.Equal(t, 40, newPositions[0])
	})

	t.Run("should update reel positions correctly for each reel", func(t *testing.T) {
		grid := reels.Grid{
			{EmptySymbol, EmptySymbol, "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{EmptySymbol, EmptySymbol, EmptySymbol, "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{EmptySymbol, "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		reelStrips := make([]reels.ReelStrip, 5)
		for i := 0; i < 5; i++ {
			strip := make(reels.ReelStrip, 100)
			for j := 0; j < 100; j++ {
				strip[j] = "symbol"
			}
			reelStrips[i] = strip
		}

		reelPositions := []int{50, 60, 70, 80, 90}

		_, newPositions := fillEmptyPositions(grid, reelStrips, reelPositions)

		// Reel 0: 2 empty → 50 - 2 = 48
		assert.Equal(t, 48, newPositions[0])
		// Reel 1: 3 empty → 60 - 3 = 57
		assert.Equal(t, 57, newPositions[1])
		// Reel 2: 1 empty → 70 - 1 = 69
		assert.Equal(t, 69, newPositions[2])
		// Reel 3: 0 empty → 80 (unchanged)
		assert.Equal(t, 80, newPositions[3])
		// Reel 4: 0 empty → 90 (unchanged)
		assert.Equal(t, 90, newPositions[4])
	})
}

// ============================================================================
// ExecuteCascades TESTS
// ============================================================================

func TestExecuteCascades(t *testing.T) {
	t.Run("should execute single cascade with win", func(t *testing.T) {
		// Create a grid with winning combination
		initialGrid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		// Use cycling pattern to prevent infinite cascades
		reelStrips := make([]reels.ReelStrip, 5)
		symbolCycle := []string{"cai", "fu", "shu", "zhong", "liangtong", "bai", "wusuo", "wutong"}
		for i := 0; i < 5; i++ {
			strip := make(reels.ReelStrip, 100)
			for j := 0; j < 100; j++ {
				strip[j] = symbolCycle[j%len(symbolCycle)]
			}
			reelStrips[i] = strip
		}

		reelPositions := []int{0, 0, 0, 0, 0}
		betAmount := 20.0
		isFreeSpin := false

		cryptoRNG := rng.NewCryptoRNG()

		cascadeResults, finalGrid, err := ExecuteCascades(
			initialGrid,
			reelStrips,
			reelPositions,
			betAmount,
			isFreeSpin,
			cryptoRNG,
		)

		require.NoError(t, err)
		assert.Greater(t, len(cascadeResults), 0, "Should have at least one cascade")

		// First cascade should have wins
		assert.Greater(t, cascadeResults[0].TotalCascadeWin, 0.0)
		assert.Equal(t, 1, cascadeResults[0].CascadeNumber)
		assert.Equal(t, 1, cascadeResults[0].Multiplier) // Base game cascade 1 = 1x

		// Verify final grid is valid
		assert.NotNil(t, finalGrid)
		assert.Equal(t, 5, len(finalGrid))
	})

	t.Run("should stop when no more wins", func(t *testing.T) {
		// Create a grid with no winning combination (visible rows 0-3 have no 3-of-a-kind)
		// Each row (0-3) has different symbols across reels to prevent wins
		initialGrid := reels.Grid{
			{"cai", "cai", "cai", "cai", "cai", "cai", "cai", "cai", "cai", "cai"},      // Reel 0
			{"fu", "fu", "fu", "fu", "fu", "fu", "fu", "fu", "fu", "fu"},                  // Reel 1
			{"shu", "shu", "shu", "shu", "shu", "shu", "shu", "shu", "shu", "shu"},        // Reel 2 - NO match to reels 0 or 1
			{"zhong", "zhong", "zhong", "zhong", "zhong", "zhong", "zhong", "zhong", "zhong", "zhong"}, // Reel 3 - NO match to any previous reel
			{"liangtong", "liangtong", "liangtong", "liangtong", "liangtong", "liangtong", "liangtong", "liangtong", "liangtong", "liangtong"}, // Reel 4 - NO match to any previous reel
		}

		// Create reel strips with varied symbols to prevent accidental winning combinations
		// Use a cycling pattern of different symbols to ensure no 3+ of same symbol
		reelStrips := make([]reels.ReelStrip, 5)
		symbolCycle := []string{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo"}
		for i := 0; i < 5; i++ {
			strip := make(reels.ReelStrip, 100)
			for j := 0; j < 100; j++ {
				strip[j] = symbolCycle[j%len(symbolCycle)]
			}
			reelStrips[i] = strip
		}

		reelPositions := []int{0, 0, 0, 0, 0}
		betAmount := 20.0
		isFreeSpin := false

		cryptoRNG := rng.NewCryptoRNG()

		cascadeResults, _, err := ExecuteCascades(
			initialGrid,
			reelStrips,
			reelPositions,
			betAmount,
			isFreeSpin,
			cryptoRNG,
		)

		require.NoError(t, err)
		assert.Equal(t, 0, len(cascadeResults), "Should have no cascades for no wins")
	})

	t.Run("should apply correct multipliers for base game", func(t *testing.T) {
		// Create grid that will create multiple cascades
		// This is a simplified test - actual cascades depend on complex game logic
		initialGrid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "fu", "shu"},
		}

		// Use cycling pattern to prevent infinite cascades
		reelStrips := make([]reels.ReelStrip, 5)
		symbolCycle := []string{"cai", "fu", "shu", "zhong", "liangtong", "bai", "wusuo", "wutong"}
		for i := 0; i < 5; i++ {
			strip := make(reels.ReelStrip, 100)
			for j := 0; j < 100; j++ {
				strip[j] = symbolCycle[j%len(symbolCycle)]
			}
			reelStrips[i] = strip
		}

		reelPositions := []int{0, 0, 0, 0, 0}
		betAmount := 20.0
		isFreeSpin := false

		cryptoRNG := rng.NewCryptoRNG()

		cascadeResults, _, err := ExecuteCascades(
			initialGrid,
			reelStrips,
			reelPositions,
			betAmount,
			isFreeSpin,
			cryptoRNG,
		)

		require.NoError(t, err)

		if len(cascadeResults) > 0 {
			// First cascade should have 1x multiplier
			assert.Equal(t, multiplier.GetMultiplier(1, false), cascadeResults[0].Multiplier)
		}

		if len(cascadeResults) > 1 {
			// Second cascade should have 2x multiplier
			assert.Equal(t, multiplier.GetMultiplier(2, false), cascadeResults[1].Multiplier)
		}
	})

	t.Run("should apply correct multipliers for free spins", func(t *testing.T) {
		initialGrid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "fu", "shu"},
		}

		// Use cycling pattern to prevent infinite cascades
		reelStrips := make([]reels.ReelStrip, 5)
		symbolCycle := []string{"cai", "fu", "shu", "zhong", "liangtong", "bai", "wusuo", "wutong"}
		for i := 0; i < 5; i++ {
			strip := make(reels.ReelStrip, 100)
			for j := 0; j < 100; j++ {
				strip[j] = symbolCycle[j%len(symbolCycle)]
			}
			reelStrips[i] = strip
		}

		reelPositions := []int{0, 0, 0, 0, 0}
		betAmount := 20.0
		isFreeSpin := true

		cryptoRNG := rng.NewCryptoRNG()

		cascadeResults, _, err := ExecuteCascades(
			initialGrid,
			reelStrips,
			reelPositions,
			betAmount,
			isFreeSpin,
			cryptoRNG,
		)

		require.NoError(t, err)

		if len(cascadeResults) > 0 {
			// First cascade should have 2x multiplier (free spin)
			assert.Equal(t, multiplier.GetMultiplier(1, true), cascadeResults[0].Multiplier)
			assert.Equal(t, 2, cascadeResults[0].Multiplier)
		}
	})

	t.Run("should not modify initial grid", func(t *testing.T) {
		initialGrid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		}

		originalSymbol := initialGrid.GetSymbol(0, 5)

		// Use cycling pattern to prevent infinite cascades
		reelStrips := make([]reels.ReelStrip, 5)
		symbolCycle := []string{"cai", "fu", "shu", "zhong", "liangtong", "bai", "wusuo", "wutong"}
		for i := 0; i < 5; i++ {
			strip := make(reels.ReelStrip, 100)
			for j := 0; j < 100; j++ {
				strip[j] = symbolCycle[j%len(symbolCycle)]
			}
			reelStrips[i] = strip
		}

		reelPositions := []int{0, 0, 0, 0, 0}
		betAmount := 20.0
		isFreeSpin := false

		cryptoRNG := rng.NewCryptoRNG()

		_, _, err := ExecuteCascades(
			initialGrid,
			reelStrips,
			reelPositions,
			betAmount,
			isFreeSpin,
			cryptoRNG,
		)

		require.NoError(t, err)

		// Initial grid should be unchanged
		assert.Equal(t, originalSymbol, initialGrid.GetSymbol(0, 5))
	})
}

// ============================================================================
// GetTotalWinFromCascades TESTS
// ============================================================================

func TestGetTotalWinFromCascades(t *testing.T) {
	t.Run("should sum wins from all cascades", func(t *testing.T) {
		cascadeResults := []CascadeResult{
			{TotalCascadeWin: 100.0},
			{TotalCascadeWin: 50.0},
			{TotalCascadeWin: 25.0},
		}

		betAmount := 10.0
		total := GetTotalWinFromCascades(cascadeResults, betAmount)

		assert.Equal(t, 175.0, total)
	})

	t.Run("should apply max win cap", func(t *testing.T) {
		betAmount := 10.0
		maxWin := betAmount * 25000 // 250,000

		cascadeResults := []CascadeResult{
			{TotalCascadeWin: 300000.0}, // Exceeds cap
		}

		total := GetTotalWinFromCascades(cascadeResults, betAmount)

		assert.Equal(t, maxWin, total)
	})

	t.Run("should handle empty cascade results", func(t *testing.T) {
		cascadeResults := []CascadeResult{}

		betAmount := 10.0
		total := GetTotalWinFromCascades(cascadeResults, betAmount)

		assert.Equal(t, 0.0, total)
	})

	t.Run("should handle zero wins", func(t *testing.T) {
		cascadeResults := []CascadeResult{
			{TotalCascadeWin: 0.0},
			{TotalCascadeWin: 0.0},
		}

		betAmount := 10.0
		total := GetTotalWinFromCascades(cascadeResults, betAmount)

		assert.Equal(t, 0.0, total)
	})

	t.Run("should handle fractional wins", func(t *testing.T) {
		cascadeResults := []CascadeResult{
			{TotalCascadeWin: 12.50},
			{TotalCascadeWin: 7.25},
			{TotalCascadeWin: 3.75},
		}

		betAmount := 5.0
		total := GetTotalWinFromCascades(cascadeResults, betAmount)

		assert.InDelta(t, 23.50, total, 0.001)
	})
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkRemoveWinningSymbols(b *testing.B) {
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
	}

	symbolWins := []wins.SymbolWin{
		{Symbol: symbols.SymbolFa, Count: 3, Ways: 1},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = removeWinningSymbols(grid, symbolWins)
	}
}

func BenchmarkDropSymbols(b *testing.B) {
	grid := reels.Grid{
		{"cai", "fu", EmptySymbol, EmptySymbol, "liangtong", EmptySymbol, "fa", "fa", "cai", "fu"},
		{"cai", "fu", "shu", EmptySymbol, "liangtong", "fa", "fa", "fa", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dropSymbols(grid)
	}
}

func BenchmarkExecuteCascades(b *testing.B) {
	initialGrid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fa", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
	}

	reelStrips := make([]reels.ReelStrip, 5)
	for i := 0; i < 5; i++ {
		strip := make(reels.ReelStrip, 100)
		for j := 0; j < 100; j++ {
			strip[j] = "liangtong"
		}
		reelStrips[i] = strip
	}

	reelPositions := []int{0, 0, 0, 0, 0}
	betAmount := 20.0
	isFreeSpin := false

	cryptoRNG := rng.NewCryptoRNG()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ExecuteCascades(initialGrid, reelStrips, reelPositions, betAmount, isFreeSpin, cryptoRNG)
	}
}
