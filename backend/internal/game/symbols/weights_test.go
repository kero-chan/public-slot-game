package symbols

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBaseGameWeights(t *testing.T) {
	t.Run("should return correct weights for each reel", func(t *testing.T) {
		reel0 := GetBaseGameWeights(0)
		reel1 := GetBaseGameWeights(1)
		reel2 := GetBaseGameWeights(2)
		reel3 := GetBaseGameWeights(3)
		reel4 := GetBaseGameWeights(4)

		assert.Equal(t, BaseGameWeights.Reel1, reel0)
		assert.Equal(t, BaseGameWeights.Reel2, reel1)
		assert.Equal(t, BaseGameWeights.Reel3, reel2)
		assert.Equal(t, BaseGameWeights.Reel4, reel3)
		assert.Equal(t, BaseGameWeights.Reel5, reel4)
	})

	t.Run("should return default for out of range index", func(t *testing.T) {
		// Out of range should return Reel1 as default
		assert.Equal(t, BaseGameWeights.Reel1, GetBaseGameWeights(-1))
		assert.Equal(t, BaseGameWeights.Reel1, GetBaseGameWeights(5))
		assert.Equal(t, BaseGameWeights.Reel1, GetBaseGameWeights(10))
	})

	t.Run("should return non-nil weights for all valid indices", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			weights := GetBaseGameWeights(i)
			assert.NotNil(t, weights, "Reel %d should have weights", i)
			assert.Greater(t, len(weights), 0, "Reel %d should have at least one symbol", i)
		}
	})
}

func TestGetFreeSpinsWeights(t *testing.T) {
	t.Run("should return correct weights for each reel", func(t *testing.T) {
		reel0 := GetFreeSpinsWeights(0)
		reel1 := GetFreeSpinsWeights(1)
		reel2 := GetFreeSpinsWeights(2)
		reel3 := GetFreeSpinsWeights(3)
		reel4 := GetFreeSpinsWeights(4)

		assert.Equal(t, FreeSpinsWeights.Reel1, reel0)
		assert.Equal(t, FreeSpinsWeights.Reel2, reel1)
		assert.Equal(t, FreeSpinsWeights.Reel3, reel2)
		assert.Equal(t, FreeSpinsWeights.Reel4, reel3)
		assert.Equal(t, FreeSpinsWeights.Reel5, reel4)
	})

	t.Run("should return default for out of range index", func(t *testing.T) {
		// Out of range should return Reel1 as default
		assert.Equal(t, FreeSpinsWeights.Reel1, GetFreeSpinsWeights(-1))
		assert.Equal(t, FreeSpinsWeights.Reel1, GetFreeSpinsWeights(5))
		assert.Equal(t, FreeSpinsWeights.Reel1, GetFreeSpinsWeights(10))
	})

	t.Run("should return non-nil weights for all valid indices", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			weights := GetFreeSpinsWeights(i)
			assert.NotNil(t, weights, "Reel %d should have weights", i)
			assert.Greater(t, len(weights), 0, "Reel %d should have at least one symbol", i)
		}
	})
}

func TestBaseGameWeights_TotalWeights(t *testing.T) {
	t.Run("should have positive total weight for reel 1", func(t *testing.T) {
		total := calculateTotalWeight(BaseGameWeights.Reel1)
		assert.Greater(t, total, 0, "Reel 1 should have positive total weight")
	})

	t.Run("should have positive total weight for reel 2", func(t *testing.T) {
		total := calculateTotalWeight(BaseGameWeights.Reel2)
		assert.Greater(t, total, 0, "Reel 2 should have positive total weight")
	})

	t.Run("should have positive total weight for reel 3", func(t *testing.T) {
		total := calculateTotalWeight(BaseGameWeights.Reel3)
		assert.Greater(t, total, 0, "Reel 3 should have positive total weight")
	})

	t.Run("should have positive total weight for reel 4", func(t *testing.T) {
		total := calculateTotalWeight(BaseGameWeights.Reel4)
		assert.Greater(t, total, 0, "Reel 4 should have positive total weight")
	})

	t.Run("should have positive total weight for reel 5", func(t *testing.T) {
		total := calculateTotalWeight(BaseGameWeights.Reel5)
		assert.Greater(t, total, 0, "Reel 5 should have positive total weight")
	})

	t.Run("all reels should have same total weight", func(t *testing.T) {
		reel1Total := calculateTotalWeight(BaseGameWeights.Reel1)
		reel2Total := calculateTotalWeight(BaseGameWeights.Reel2)
		reel3Total := calculateTotalWeight(BaseGameWeights.Reel3)
		reel4Total := calculateTotalWeight(BaseGameWeights.Reel4)
		reel5Total := calculateTotalWeight(BaseGameWeights.Reel5)

		assert.Equal(t, reel1Total, reel2Total, "Reel 1 and 2 should have same total weight")
		assert.Equal(t, reel2Total, reel3Total, "Reel 2 and 3 should have same total weight")
		assert.Equal(t, reel3Total, reel4Total, "Reel 3 and 4 should have same total weight")
		assert.Equal(t, reel4Total, reel5Total, "Reel 4 and 5 should have same total weight")
	})
}

func TestFreeSpinsWeights_TotalWeights(t *testing.T) {
	t.Run("should have positive total weight for reel 1", func(t *testing.T) {
		total := calculateTotalWeight(FreeSpinsWeights.Reel1)
		assert.Greater(t, total, 0, "Free spins reel 1 should have positive total weight")
	})

	t.Run("should have positive total weight for reel 2", func(t *testing.T) {
		total := calculateTotalWeight(FreeSpinsWeights.Reel2)
		assert.Greater(t, total, 0, "Free spins reel 2 should have positive total weight")
	})

	t.Run("should have positive total weight for reel 3", func(t *testing.T) {
		total := calculateTotalWeight(FreeSpinsWeights.Reel3)
		assert.Greater(t, total, 0, "Free spins reel 3 should have positive total weight")
	})

	t.Run("should have positive total weight for reel 4", func(t *testing.T) {
		total := calculateTotalWeight(FreeSpinsWeights.Reel4)
		assert.Greater(t, total, 0, "Free spins reel 4 should have positive total weight")
	})

	t.Run("should have positive total weight for reel 5", func(t *testing.T) {
		total := calculateTotalWeight(FreeSpinsWeights.Reel5)
		assert.Greater(t, total, 0, "Free spins reel 5 should have positive total weight")
	})

	t.Run("all reels should have same total weight", func(t *testing.T) {
		reel1Total := calculateTotalWeight(FreeSpinsWeights.Reel1)
		reel2Total := calculateTotalWeight(FreeSpinsWeights.Reel2)
		reel3Total := calculateTotalWeight(FreeSpinsWeights.Reel3)
		reel4Total := calculateTotalWeight(FreeSpinsWeights.Reel4)
		reel5Total := calculateTotalWeight(FreeSpinsWeights.Reel5)

		assert.Equal(t, reel1Total, reel2Total, "Reel 1 and 2 should have same total weight")
		assert.Equal(t, reel2Total, reel3Total, "Reel 2 and 3 should have same total weight")
		assert.Equal(t, reel3Total, reel4Total, "Reel 3 and 4 should have same total weight")
		assert.Equal(t, reel4Total, reel5Total, "Reel 4 and 5 should have same total weight")
	})
}

func TestBaseGameWeights_GoldVariants(t *testing.T) {
	t.Run("reel 1 should have zero weight for gold variants", func(t *testing.T) {
		for symbol, weight := range BaseGameWeights.Reel1 {
			if IsGoldVariant(symbol) {
				assert.Equal(t, 0, weight,
					"Reel 1 should have zero weight for gold variant %s", symbol)
			}
		}
	})

	t.Run("reel 2 should have gold variants with positive weight", func(t *testing.T) {
		hasGoldWithWeight := false
		for symbol, weight := range BaseGameWeights.Reel2 {
			if IsGoldVariant(symbol) && weight > 0 {
				hasGoldWithWeight = true
				break
			}
		}
		assert.True(t, hasGoldWithWeight, "Reel 2 should have gold variants with positive weight")
	})

	t.Run("reel 3 should have gold variants with positive weight", func(t *testing.T) {
		hasGoldWithWeight := false
		for symbol, weight := range BaseGameWeights.Reel3 {
			if IsGoldVariant(symbol) && weight > 0 {
				hasGoldWithWeight = true
				break
			}
		}
		assert.True(t, hasGoldWithWeight, "Reel 3 should have gold variants with positive weight")
	})

	t.Run("reel 4 should have gold variants with positive weight", func(t *testing.T) {
		hasGoldWithWeight := false
		for symbol, weight := range BaseGameWeights.Reel4 {
			if IsGoldVariant(symbol) && weight > 0 {
				hasGoldWithWeight = true
				break
			}
		}
		assert.True(t, hasGoldWithWeight, "Reel 4 should have gold variants with positive weight")
	})

	t.Run("reel 5 should have zero weight for gold variants", func(t *testing.T) {
		for symbol, weight := range BaseGameWeights.Reel5 {
			if IsGoldVariant(symbol) {
				assert.Equal(t, 0, weight,
					"Reel 5 should have zero weight for gold variant %s", symbol)
			}
		}
	})

	t.Run("middle reels should have all 8 paying symbols with gold variants", func(t *testing.T) {
		middleReels := []ReelWeights{
			BaseGameWeights.Reel2,
			BaseGameWeights.Reel3,
			BaseGameWeights.Reel4,
		}

		payingSymbols := PayingSymbols()

		for reelIdx, reel := range middleReels {
			for _, payingSym := range payingSymbols {
				// Check base symbol exists
				_, hasBase := reel[string(payingSym)]

				// Check gold variant exists
				goldVariant := string(payingSym) + "_gold"
				_, hasGold := reel[goldVariant]

				assert.True(t, hasBase,
					"Middle reel %d should have base symbol %s", reelIdx+2, payingSym)
				assert.True(t, hasGold,
					"Middle reel %d should have gold variant %s", reelIdx+2, goldVariant)
			}
		}
	})
}

func TestBaseGameWeights_BonusSymbol(t *testing.T) {
	t.Run("all reels should have bonus symbol", func(t *testing.T) {
		reels := []ReelWeights{
			BaseGameWeights.Reel1,
			BaseGameWeights.Reel2,
			BaseGameWeights.Reel3,
			BaseGameWeights.Reel4,
			BaseGameWeights.Reel5,
		}

		for i, reel := range reels {
			weight, hasBonus := reel["bonus"]
			assert.True(t, hasBonus, "Reel %d should have bonus symbol", i+1)
			assert.Greater(t, weight, 0, "Reel %d bonus weight should be > 0", i+1)
		}
	})

	t.Run("bonus weight should be consistent across reels", func(t *testing.T) {
		bonusWeights := []int{
			BaseGameWeights.Reel1["bonus"],
			BaseGameWeights.Reel2["bonus"],
			BaseGameWeights.Reel3["bonus"],
			BaseGameWeights.Reel4["bonus"],
			BaseGameWeights.Reel5["bonus"],
		}

		// All bonus weights should be the same (calculated from same rate)
		first := bonusWeights[0]
		for i, w := range bonusWeights[1:] {
			assert.Equal(t, first, w, "Reel %d bonus weight should equal reel 1", i+2)
		}
	})
}

func TestFreeSpinsWeights_BonusSymbol(t *testing.T) {
	t.Run("all reels should have bonus symbol in free spins", func(t *testing.T) {
		reels := []ReelWeights{
			FreeSpinsWeights.Reel1,
			FreeSpinsWeights.Reel2,
			FreeSpinsWeights.Reel3,
			FreeSpinsWeights.Reel4,
			FreeSpinsWeights.Reel5,
		}

		for i, reel := range reels {
			weight, hasBonus := reel["bonus"]
			assert.True(t, hasBonus, "Free spins reel %d should have bonus symbol", i+1)
			assert.Greater(t, weight, 0, "Free spins reel %d bonus weight should be > 0", i+1)
		}
	})
}

func TestWeights_SymbolCoverage(t *testing.T) {
	t.Run("base game outer reels should have all paying symbols", func(t *testing.T) {
		outerReels := []ReelWeights{
			BaseGameWeights.Reel1,
			BaseGameWeights.Reel5,
		}

		payingSymbols := PayingSymbols()

		for reelIdx, reel := range outerReels {
			reelName := []string{"Reel1", "Reel5"}[reelIdx]

			for _, payingSym := range payingSymbols {
				weight, exists := reel[string(payingSym)]
				assert.True(t, exists,
					"%s should have paying symbol %s", reelName, payingSym)
				assert.Greater(t, weight, 0,
					"%s weight for %s should be > 0", reelName, payingSym)
			}
		}
	})

	t.Run("base game should not have wild symbol", func(t *testing.T) {
		reels := []ReelWeights{
			BaseGameWeights.Reel1,
			BaseGameWeights.Reel2,
			BaseGameWeights.Reel3,
			BaseGameWeights.Reel4,
			BaseGameWeights.Reel5,
		}

		for i, reel := range reels {
			_, hasWild := reel["wild"]
			assert.False(t, hasWild, "Base game reel %d should NOT have wild symbol", i+1)
		}
	})

	t.Run("free spins should not have wild symbol", func(t *testing.T) {
		reels := []ReelWeights{
			FreeSpinsWeights.Reel1,
			FreeSpinsWeights.Reel2,
			FreeSpinsWeights.Reel3,
			FreeSpinsWeights.Reel4,
			FreeSpinsWeights.Reel5,
		}

		for i, reel := range reels {
			_, hasWild := reel["wild"]
			assert.False(t, hasWild, "Free spins reel %d should NOT have wild symbol", i+1)
		}
	})
}

func TestWeights_Symmetry(t *testing.T) {
	t.Run("outer reels should have similar structure", func(t *testing.T) {
		// Reel 1 and Reel 5 are outer reels - should have same symbols
		reel1Symbols := make(map[string]bool)
		reel5Symbols := make(map[string]bool)

		for sym := range BaseGameWeights.Reel1 {
			reel1Symbols[sym] = true
		}

		for sym := range BaseGameWeights.Reel5 {
			reel5Symbols[sym] = true
		}

		// Should have same symbols (though weights may differ)
		assert.Equal(t, len(reel1Symbols), len(reel5Symbols),
			"Outer reels should have same number of symbols")

		for sym := range reel1Symbols {
			assert.True(t, reel5Symbols[sym],
				"Reel 5 should have symbol %s that Reel 1 has", sym)
		}
	})

	t.Run("middle reels should have similar structure", func(t *testing.T) {
		// Reels 2, 3, 4 are middle reels - should have same symbols
		reel2Symbols := make(map[string]bool)
		reel3Symbols := make(map[string]bool)
		reel4Symbols := make(map[string]bool)

		for sym := range BaseGameWeights.Reel2 {
			reel2Symbols[sym] = true
		}
		for sym := range BaseGameWeights.Reel3 {
			reel3Symbols[sym] = true
		}
		for sym := range BaseGameWeights.Reel4 {
			reel4Symbols[sym] = true
		}

		// Should have same symbols (though weights may differ)
		assert.Equal(t, len(reel2Symbols), len(reel3Symbols),
			"Middle reels should have same number of symbols")
		assert.Equal(t, len(reel2Symbols), len(reel4Symbols),
			"Middle reels should have same number of symbols")
	})
}

func TestWeights_NonZero(t *testing.T) {
	t.Run("all base game non-gold weights should be positive", func(t *testing.T) {
		reels := []ReelWeights{
			BaseGameWeights.Reel1,
			BaseGameWeights.Reel2,
			BaseGameWeights.Reel3,
			BaseGameWeights.Reel4,
			BaseGameWeights.Reel5,
		}

		for reelIdx, reel := range reels {
			for symbol, weight := range reel {
				// Gold variants on outer reels (1 and 5) can have 0 weight
				if IsGoldVariant(symbol) && (reelIdx == 0 || reelIdx == 4) {
					assert.GreaterOrEqual(t, weight, 0,
						"Base game reel %d symbol %s should have non-negative weight", reelIdx+1, symbol)
				} else {
					assert.Greater(t, weight, 0,
						"Base game reel %d symbol %s should have positive weight", reelIdx+1, symbol)
				}
			}
		}
	})

	t.Run("all free spins non-gold weights should be positive", func(t *testing.T) {
		reels := []ReelWeights{
			FreeSpinsWeights.Reel1,
			FreeSpinsWeights.Reel2,
			FreeSpinsWeights.Reel3,
			FreeSpinsWeights.Reel4,
			FreeSpinsWeights.Reel5,
		}

		for reelIdx, reel := range reels {
			for symbol, weight := range reel {
				// Gold variants on outer reels (1 and 5) can have 0 weight
				if IsGoldVariant(symbol) && (reelIdx == 0 || reelIdx == 4) {
					assert.GreaterOrEqual(t, weight, 0,
						"Free spins reel %d symbol %s should have non-negative weight", reelIdx+1, symbol)
				} else {
					assert.Greater(t, weight, 0,
						"Free spins reel %d symbol %s should have positive weight", reelIdx+1, symbol)
				}
			}
		}
	})
}

// Helper function to calculate total weight of a reel
func calculateTotalWeight(weights ReelWeights) int {
	total := 0
	for _, weight := range weights {
		total += weight
	}
	return total
}

func BenchmarkGetBaseGameWeights(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetBaseGameWeights(2)
	}
}

func BenchmarkGetFreeSpinsWeights(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetFreeSpinsWeights(2)
	}
}
