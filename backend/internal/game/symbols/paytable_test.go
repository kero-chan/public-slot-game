package symbols

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPayout(t *testing.T) {
	t.Run("should return correct payout for high-value symbols", func(t *testing.T) {
		testCases := []struct {
			symbol   Symbol
			count    int
			expected float64
		}{
			// SymbolFa - Highest value
			{SymbolFa, 3, 10.0},
			{SymbolFa, 4, 25.0},
			{SymbolFa, 5, 50.0},

			// SymbolZhong - Premium high
			{SymbolZhong, 3, 8.0},
			{SymbolZhong, 4, 20.0},
			{SymbolZhong, 5, 40.0},

			// SymbolBai - Premium mid
			{SymbolBai, 3, 6.0},
			{SymbolBai, 4, 15.0},
			{SymbolBai, 5, 30.0},

			// SymbolBawan - Medium high
			{SymbolBawan, 3, 5.0},
			{SymbolBawan, 4, 10.0},
			{SymbolBawan, 5, 15.0},
		}

		for _, tc := range testCases {
			result := GetPayout(tc.symbol, tc.count)
			assert.Equal(t, tc.expected, result,
				"Symbol %s with count %d should have payout %.1f",
				tc.symbol, tc.count, tc.expected)
		}
	})

	t.Run("should return correct payout for low-value symbols", func(t *testing.T) {
		testCases := []struct {
			symbol   Symbol
			count    int
			expected float64
		}{
			// SymbolWusuo - Medium
			{SymbolWusuo, 3, 3.0},
			{SymbolWusuo, 4, 5.0},
			{SymbolWusuo, 5, 12.0},

			// SymbolWutong - Medium
			{SymbolWutong, 3, 3.0},
			{SymbolWutong, 4, 5.0},
			{SymbolWutong, 5, 12.0},

			// SymbolLiangsuo - Low
			{SymbolLiangsuo, 3, 2.0},
			{SymbolLiangsuo, 4, 4.0},
			{SymbolLiangsuo, 5, 10.0},

			// SymbolLiangtong - Lowest
			{SymbolLiangtong, 3, 1.0},
			{SymbolLiangtong, 4, 3.0},
			{SymbolLiangtong, 5, 6.0},
		}

		for _, tc := range testCases {
			result := GetPayout(tc.symbol, tc.count)
			assert.Equal(t, tc.expected, result,
				"Symbol %s with count %d should have payout %.1f",
				tc.symbol, tc.count, tc.expected)
		}
	})

	t.Run("should return 0 for special symbols", func(t *testing.T) {
		// Wild, bonus, and gold don't have direct payouts
		assert.Equal(t, 0.0, GetPayout(SymbolWild, 3))
		assert.Equal(t, 0.0, GetPayout(SymbolWild, 4))
		assert.Equal(t, 0.0, GetPayout(SymbolWild, 5))

		assert.Equal(t, 0.0, GetPayout(SymbolBonus, 3))
		assert.Equal(t, 0.0, GetPayout(SymbolBonus, 4))
		assert.Equal(t, 0.0, GetPayout(SymbolBonus, 5))

		assert.Equal(t, 0.0, GetPayout(SymbolGold, 3))
		assert.Equal(t, 0.0, GetPayout(SymbolGold, 4))
		assert.Equal(t, 0.0, GetPayout(SymbolGold, 5))
	})

	t.Run("should return 0 for invalid symbol", func(t *testing.T) {
		invalidSymbol := Symbol("invalid_symbol")
		assert.Equal(t, 0.0, GetPayout(invalidSymbol, 3))
		assert.Equal(t, 0.0, GetPayout(invalidSymbol, 4))
		assert.Equal(t, 0.0, GetPayout(invalidSymbol, 5))
	})

	t.Run("should return 0 for count below minimum", func(t *testing.T) {
		// Count < 3 should not award payout
		assert.Equal(t, 0.0, GetPayout(SymbolFa, 0))
		assert.Equal(t, 0.0, GetPayout(SymbolFa, 1))
		assert.Equal(t, 0.0, GetPayout(SymbolFa, 2))

		assert.Equal(t, 0.0, GetPayout(SymbolLiangtong, 0))
		assert.Equal(t, 0.0, GetPayout(SymbolLiangtong, 1))
		assert.Equal(t, 0.0, GetPayout(SymbolLiangtong, 2))
	})

	t.Run("should return 0 for count above maximum", func(t *testing.T) {
		// Count > 5 should not have defined payouts
		assert.Equal(t, 0.0, GetPayout(SymbolFa, 6))
		assert.Equal(t, 0.0, GetPayout(SymbolFa, 7))
		assert.Equal(t, 0.0, GetPayout(SymbolFa, 10))
	})

	t.Run("should validate payout progression", func(t *testing.T) {
		// For each symbol, payout should increase with count
		payingSymbols := []Symbol{
			SymbolFa, SymbolZhong, SymbolBai, SymbolBawan,
			SymbolWusuo, SymbolWutong, SymbolLiangsuo, SymbolLiangtong,
		}

		for _, sym := range payingSymbols {
			payout3 := GetPayout(sym, 3)
			payout4 := GetPayout(sym, 4)
			payout5 := GetPayout(sym, 5)

			assert.Greater(t, payout4, payout3,
				"Symbol %s: 4-of-a-kind payout should be greater than 3-of-a-kind", sym)
			assert.Greater(t, payout5, payout4,
				"Symbol %s: 5-of-a-kind payout should be greater than 4-of-a-kind", sym)
		}
	})
}

func TestMinSymbolsForPayout(t *testing.T) {
	t.Run("should return 3 as minimum", func(t *testing.T) {
		minSymbols := MinSymbolsForPayout()
		assert.Equal(t, 3, minSymbols, "Minimum symbols for payout should be 3")
	})

	t.Run("should validate all symbols respect minimum", func(t *testing.T) {
		payingSymbols := []Symbol{
			SymbolFa, SymbolZhong, SymbolBai, SymbolBawan,
			SymbolWusuo, SymbolWutong, SymbolLiangsuo, SymbolLiangtong,
		}

		minSymbols := MinSymbolsForPayout()

		for _, sym := range payingSymbols {
			// Should have payout at minimum
			payoutAtMin := GetPayout(sym, minSymbols)
			assert.Greater(t, payoutAtMin, 0.0,
				"Symbol %s should have payout at minimum count %d", sym, minSymbols)

			// Should NOT have payout below minimum
			payoutBelowMin := GetPayout(sym, minSymbols-1)
			assert.Equal(t, 0.0, payoutBelowMin,
				"Symbol %s should not have payout below minimum count %d", sym, minSymbols)
		}
	})
}

func TestMaxSymbolsForPayout(t *testing.T) {
	t.Run("should return 5 as maximum", func(t *testing.T) {
		maxSymbols := MaxSymbolsForPayout()
		assert.Equal(t, 5, maxSymbols, "Maximum symbols for payout should be 5 (number of reels)")
	})

	t.Run("should validate all symbols have payout at maximum", func(t *testing.T) {
		payingSymbols := []Symbol{
			SymbolFa, SymbolZhong, SymbolBai, SymbolBawan,
			SymbolWusuo, SymbolWutong, SymbolLiangsuo, SymbolLiangtong,
		}

		maxSymbols := MaxSymbolsForPayout()

		for _, sym := range payingSymbols {
			// Should have payout at maximum
			payoutAtMax := GetPayout(sym, maxSymbols)
			assert.Greater(t, payoutAtMax, 0.0,
				"Symbol %s should have payout at maximum count %d", sym, maxSymbols)
		}
	})
}

func TestGetFreeSpinsAward(t *testing.T) {
	t.Run("should return correct free spins for valid scatter counts", func(t *testing.T) {
		testCases := []struct {
			scatterCount int
			expected     int
		}{
			{3, 12}, // Base: 12 free spins
			{4, 14}, // 12 + (2 × 1) = 14
			{5, 16}, // 12 + (2 × 2) = 16
		}

		for _, tc := range testCases {
			result := GetFreeSpinsAward(tc.scatterCount)
			assert.Equal(t, tc.expected, result,
				"%d scatters should award %d free spins", tc.scatterCount, tc.expected)
		}
	})

	t.Run("should return 0 for scatter count below minimum", func(t *testing.T) {
		assert.Equal(t, 0, GetFreeSpinsAward(0))
		assert.Equal(t, 0, GetFreeSpinsAward(1))
		assert.Equal(t, 0, GetFreeSpinsAward(2))
	})

	t.Run("should use formula for scatter counts above 5", func(t *testing.T) {
		// Formula: 12 + (2 × (scatter_count - 3))
		testCases := []struct {
			scatterCount int
			expected     int
		}{
			{6, 18}, // 12 + (2 × 3) = 18
			{7, 20}, // 12 + (2 × 4) = 20
			{8, 22}, // 12 + (2 × 5) = 22
			{10, 26}, // 12 + (2 × 7) = 26
		}

		for _, tc := range testCases {
			result := GetFreeSpinsAward(tc.scatterCount)
			assert.Equal(t, tc.expected, result,
				"%d scatters should award %d free spins", tc.scatterCount, tc.expected)
		}
	})

	t.Run("should validate formula consistency", func(t *testing.T) {
		// Verify the formula works correctly
		for count := 3; count <= 10; count++ {
			expected := 12 + (2 * (count - 3))
			result := GetFreeSpinsAward(count)
			assert.Equal(t, expected, result,
				"Formula should be consistent for %d scatters", count)
		}
	})

	t.Run("should award increasing free spins with more scatters", func(t *testing.T) {
		// Each additional scatter should award more free spins
		for count := 3; count < 10; count++ {
			current := GetFreeSpinsAward(count)
			next := GetFreeSpinsAward(count + 1)
			assert.Greater(t, next, current,
				"%d scatters should award more free spins than %d scatters", count+1, count)
		}
	})
}

func TestMinScattersForFreeSpin(t *testing.T) {
	t.Run("should return 3 as minimum", func(t *testing.T) {
		minScatters := MinScattersForFreeSpin()
		assert.Equal(t, 3, minScatters, "Minimum scatters to trigger free spins should be 3")
	})

	t.Run("should validate award respects minimum", func(t *testing.T) {
		minScatters := MinScattersForFreeSpin()

		// Should award free spins at minimum
		award := GetFreeSpinsAward(minScatters)
		assert.Greater(t, award, 0, "Should award free spins at minimum scatter count")

		// Should NOT award free spins below minimum
		noAward := GetFreeSpinsAward(minScatters - 1)
		assert.Equal(t, 0, noAward, "Should not award free spins below minimum scatter count")
	})
}

func TestPaytableCompleteness(t *testing.T) {
	t.Run("should have complete paytable for all paying symbols", func(t *testing.T) {
		payingSymbols := []Symbol{
			SymbolFa, SymbolZhong, SymbolBai, SymbolBawan,
			SymbolWusuo, SymbolWutong, SymbolLiangsuo, SymbolLiangtong,
		}

		for _, sym := range payingSymbols {
			// Each paying symbol should have payouts for 3, 4, 5
			payout3 := GetPayout(sym, 3)
			payout4 := GetPayout(sym, 4)
			payout5 := GetPayout(sym, 5)

			assert.Greater(t, payout3, 0.0, "Symbol %s should have 3-of-a-kind payout", sym)
			assert.Greater(t, payout4, 0.0, "Symbol %s should have 4-of-a-kind payout", sym)
			assert.Greater(t, payout5, 0.0, "Symbol %s should have 5-of-a-kind payout", sym)
		}
	})

	t.Run("should validate paytable has exactly 8 paying symbols", func(t *testing.T) {
		assert.Len(t, Paytable, 8, "Paytable should have exactly 8 paying symbols")
	})

	t.Run("should validate each paying symbol has 3 payout levels", func(t *testing.T) {
		for sym, payouts := range Paytable {
			assert.Len(t, payouts, 3,
				"Symbol %s should have exactly 3 payout levels (3, 4, 5)", sym)

			// Verify keys are 3, 4, 5
			_, has3 := payouts[3]
			_, has4 := payouts[4]
			_, has5 := payouts[5]

			assert.True(t, has3, "Symbol %s should have 3-of-a-kind payout", sym)
			assert.True(t, has4, "Symbol %s should have 4-of-a-kind payout", sym)
			assert.True(t, has5, "Symbol %s should have 5-of-a-kind payout", sym)
		}
	})
}

func TestFreeSpinsAwardCompleteness(t *testing.T) {
	t.Run("should have awards defined for 3, 4, 5 scatters", func(t *testing.T) {
		assert.Contains(t, FreeSpinsAward, 3)
		assert.Contains(t, FreeSpinsAward, 4)
		assert.Contains(t, FreeSpinsAward, 5)
	})

	t.Run("should validate FreeSpinsAward has exactly 3 entries", func(t *testing.T) {
		assert.Len(t, FreeSpinsAward, 3, "FreeSpinsAward should have exactly 3 entries")
	})

	t.Run("should validate formula matches map values", func(t *testing.T) {
		for count, expectedAward := range FreeSpinsAward {
			formulaResult := 12 + (2 * (count - 3))
			assert.Equal(t, expectedAward, formulaResult,
				"FreeSpinsAward[%d] should match formula result", count)
		}
	})
}

func BenchmarkGetPayout(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetPayout(SymbolFa, 5)
	}
}

func BenchmarkGetFreeSpinsAward(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetFreeSpinsAward(4)
	}
}
