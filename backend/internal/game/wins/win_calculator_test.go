package wins

import (
	"testing"

	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateCascadeWin(t *testing.T) {
	t.Run("should calculate win for 3-of-a-kind in base game", func(t *testing.T) {
		// Create grid with "fa" 3-of-a-kind (1 way)
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}

		betAmount := 20.0
		cascadeNumber := 1
		isFreeSpin := false

		winDetails, symbolWins, totalWin := CalculateCascadeWin(grid, betAmount, cascadeNumber, isFreeSpin)

		// Should have 1 win for "fa"
		require.Len(t, winDetails, 1)
		require.Len(t, symbolWins, 1)

		// Check win detail
		detail := winDetails[0]
		assert.Equal(t, symbols.SymbolFa, detail.Symbol)
		assert.Equal(t, 3, detail.Count)
		assert.Equal(t, 1, detail.Ways)
		assert.Equal(t, 10.0, detail.Payout) // fa 3-of-a-kind = 10x

		// Win calculation: payout × ways × multiplier × betPerWay
		// = 10 × 1 × 1 × (20/20) = 10 × 1 × 1 × 1 = 10
		expectedWin := 10.0
		assert.Equal(t, expectedWin, detail.WinAmount)
		assert.Equal(t, expectedWin, totalWin)
	})

	t.Run("should calculate win with cascade multiplier", func(t *testing.T) {
		// Create grid with "fa" 3-of-a-kind
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}

		betAmount := 20.0
		cascadeNumber := 2 // 2x multiplier in base game
		isFreeSpin := false

		winDetails, _, totalWin := CalculateCascadeWin(grid, betAmount, cascadeNumber, isFreeSpin)

		require.Len(t, winDetails, 1)

		// Win = 10 × 1 × 2 × 1 = 20
		expectedWin := 20.0
		assert.Equal(t, expectedWin, winDetails[0].WinAmount)
		assert.Equal(t, expectedWin, totalWin)
	})

	t.Run("should calculate win with free spin multiplier", func(t *testing.T) {
		// Create grid with "fa" 3-of-a-kind
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}

		betAmount := 20.0
		cascadeNumber := 1 // 2x multiplier in free spins
		isFreeSpin := true

		winDetails, _, totalWin := CalculateCascadeWin(grid, betAmount, cascadeNumber, isFreeSpin)

		require.Len(t, winDetails, 1)

		// Win = 10 × 1 × 2 × 1 = 20 (free spin cascade 1 = 2x)
		expectedWin := 20.0
		assert.Equal(t, expectedWin, winDetails[0].WinAmount)
		assert.Equal(t, expectedWin, totalWin)
	})

	t.Run("should calculate win with multiple ways", func(t *testing.T) {
		// Create grid with "fa" appearing 2 times in rows 5-8 on reels 0-2
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}

		betAmount := 20.0
		cascadeNumber := 1
		isFreeSpin := false

		winDetails, _, totalWin := CalculateCascadeWin(grid, betAmount, cascadeNumber, isFreeSpin)

		require.Len(t, winDetails, 1)

		detail := winDetails[0]
		assert.Equal(t, 8, detail.Ways) // 2 × 2 × 2 = 8 ways

		// Win = 10 × 8 × 1 × 1 = 80
		expectedWin := 80.0
		assert.Equal(t, expectedWin, detail.WinAmount)
		assert.Equal(t, expectedWin, totalWin)
	})

	t.Run("should handle multiple winning symbols", func(t *testing.T) {
		// Create grid with both "fa" and "zhong" 3-of-a-kind
		// Rows 5-8 (visible):
		// Reel 0: fa, bai, wusuo, wutong (row 5-8)
		// Reel 1: fa, bai, wusuo, wutong
		// Reel 2: fa, bai, wusuo, wutong
		// Reel 3: zhong, bai, wusuo, wutong (breaks fa chain, but zhong can start)
		// Reel 4: liangtong, bai, wusuo, wutong
		grid := reels.Grid{
			{"cai", "fu", "shu", "liangsuo", "liangtong", "fa", "bai", "wusuo", "wutong", "cai"},
			{"cai", "fu", "shu", "liangsuo", "liangtong", "fa", "bai", "wusuo", "wutong", "cai"},
			{"cai", "fu", "shu", "liangsuo", "liangtong", "fa", "bai", "wusuo", "wutong", "cai"},
			{"cai", "fu", "shu", "liangsuo", "liangtong", "zhong", "bai", "wusuo", "wutong", "cai"},
			{"cai", "fu", "shu", "liangsuo", "liangtong", "liangtong", "bai", "wusuo", "wutong", "cai"},
		}

		betAmount := 20.0
		cascadeNumber := 1
		isFreeSpin := false

		winDetails, _, totalWin := CalculateCascadeWin(grid, betAmount, cascadeNumber, isFreeSpin)

		// Should have at least 1 win (fa 3-of-a-kind)
		require.GreaterOrEqual(t, len(winDetails), 1)

		// Find fa win
		var faWin *CascadeWinDetail
		for i := range winDetails {
			if winDetails[i].Symbol == symbols.SymbolFa {
				faWin = &winDetails[i]
				break
			}
		}
		require.NotNil(t, faWin)

		// fa: 10 × 1 × 1 × 1 = 10 (payout × ways × multiplier × betPerWay)
		assert.Equal(t, 10.0, faWin.WinAmount)

		// Total win includes all winning symbols
		assert.Greater(t, totalWin, 0.0)
	})

	t.Run("should handle no wins", func(t *testing.T) {
		// Create grid with no wins
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "bawan", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "liangsuo"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "fa", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "cai", "liangtong", "fa", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fu", "cai", "liangtong", "shu", "zhong"},
		}

		betAmount := 20.0
		cascadeNumber := 1
		isFreeSpin := false

		winDetails, symbolWins, totalWin := CalculateCascadeWin(grid, betAmount, cascadeNumber, isFreeSpin)

		assert.Len(t, winDetails, 0)
		assert.Len(t, symbolWins, 0)
		assert.Equal(t, 0.0, totalWin)
	})

	t.Run("should calculate bet per way correctly", func(t *testing.T) {
		// Test with different bet amounts
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}

		// Test with bet = 100
		betAmount := 100.0
		cascadeNumber := 1
		isFreeSpin := false

		winDetails, _, totalWin := CalculateCascadeWin(grid, betAmount, cascadeNumber, isFreeSpin)

		require.Len(t, winDetails, 1)

		// Win = 10 × 1 × 1 × (100/20) = 10 × 1 × 1 × 5 = 50
		expectedWin := 50.0
		assert.Equal(t, expectedWin, totalWin)
	})
}

func TestCalculateTotalSpinWin(t *testing.T) {
	t.Run("should sum cascade wins", func(t *testing.T) {
		cascadeWins := []float64{10.0, 20.0, 30.0}
		betAmount := 20.0

		totalWin := CalculateTotalSpinWin(cascadeWins, betAmount)

		expectedTotal := 60.0
		assert.Equal(t, expectedTotal, totalWin)
	})

	t.Run("should handle no cascades", func(t *testing.T) {
		cascadeWins := []float64{}
		betAmount := 20.0

		totalWin := CalculateTotalSpinWin(cascadeWins, betAmount)

		assert.Equal(t, 0.0, totalWin)
	})

	t.Run("should apply max win cap", func(t *testing.T) {
		// Create wins that exceed max cap (25,000x bet)
		cascadeWins := []float64{100000.0, 200000.0, 300000.0}
		betAmount := 10.0

		totalWin := CalculateTotalSpinWin(cascadeWins, betAmount)

		// Max win = 10 * 25,000 = 250,000
		expectedMax := 250000.0
		assert.Equal(t, expectedMax, totalWin)
	})

	t.Run("should not cap wins below limit", func(t *testing.T) {
		cascadeWins := []float64{1000.0, 2000.0}
		betAmount := 10.0

		totalWin := CalculateTotalSpinWin(cascadeWins, betAmount)

		// Total = 3,000 (well below 250,000 cap)
		expectedTotal := 3000.0
		assert.Equal(t, expectedTotal, totalWin)
	})
}

func TestGetWinningPositions(t *testing.T) {
	t.Run("should get positions for 3-of-a-kind", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}

		positions := GetWinningPositions(grid, symbols.SymbolFa, 3)

		// Should have 3 positions (one on each of first 3 reels)
		assert.Len(t, positions, 3)

		// Check reel indices
		assert.Equal(t, 0, positions[0].Reel)
		assert.Equal(t, 1, positions[1].Reel)
		assert.Equal(t, 2, positions[2].Reel)

		// All should be in visible rows (5-8)
		for _, pos := range positions {
			assert.GreaterOrEqual(t, pos.Row, 5)
			assert.LessOrEqual(t, pos.Row, 8)
		}
	})

	t.Run("should include wild substitutions", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "wild", "cai", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}

		positions := GetWinningPositions(grid, symbols.SymbolFa, 3)

		// Should have 3 positions (fa + wild + fa)
		assert.Len(t, positions, 3)
	})

	t.Run("should handle multiple symbols per reel", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}

		positions := GetWinningPositions(grid, symbols.SymbolFa, 3)

		// Should have 6 positions (2 per reel × 3 reels)
		assert.Len(t, positions, 6)
	})

	t.Run("should respect count limit", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
		}

		// Request only first 3 reels even though fa appears in all 5
		positions := GetWinningPositions(grid, symbols.SymbolFa, 3)

		// Should only return positions from first 3 reels
		assert.Len(t, positions, 3)
		for _, pos := range positions {
			assert.Less(t, pos.Reel, 3)
		}
	})

	t.Run("should handle gold variants", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa_gold", "zhong", "liangtong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "cai", "fu", "shu", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa_gold", "liangtong", "zhong", "cai", "fu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
			{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
		}

		positions := GetWinningPositions(grid, symbols.SymbolFa, 3)

		// Gold variants should match base symbol
		assert.Len(t, positions, 3)
	})
}

func BenchmarkCalculateCascadeWin(b *testing.B) {
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "liangtong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fu", "shu", "zhong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "zhong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
	}

	betAmount := 20.0
	cascadeNumber := 1
	isFreeSpin := false

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = CalculateCascadeWin(grid, betAmount, cascadeNumber, isFreeSpin)
	}
}

func BenchmarkGetWinningPositions(b *testing.B) {
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "liangtong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "fu", "shu", "zhong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "fa", "zhong", "cai", "fu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "zhong", "liangtong", "cai", "fu", "shu"},
		{"cai", "fu", "shu", "zhong", "liangtong", "liangtong", "cai", "fu", "shu", "zhong"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetWinningPositions(grid, symbols.SymbolFa, 3)
	}
}
