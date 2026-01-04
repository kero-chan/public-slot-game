package symbols

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymbolNumber(t *testing.T) {
	t.Run("should return correct numbers for special symbols", func(t *testing.T) {
		assert.Equal(t, 0, SymbolNumber("wild"))
		assert.Equal(t, 1, SymbolNumber("bonus"))
	})

	t.Run("should return correct numbers for base symbols", func(t *testing.T) {
		testCases := map[string]int{
			"fa":          2,
			"zhong":       3,
			"bai":         4,
			"bawan":       5,
			"wusuo":       6,
			"wutong":      7,
			"liangsuo":    8,
			"liangtong":   9,
		}

		for symbol, expected := range testCases {
			result := SymbolNumber(symbol)
			assert.Equal(t, expected, result,
				"Symbol %s should map to number %d", symbol, expected)
		}
	})

	t.Run("should return correct numbers for gold variants", func(t *testing.T) {
		testCases := map[string]int{
			"fa_gold":          12,
			"zhong_gold":       13,
			"bai_gold":         14,
			"bawan_gold":       15,
			"wusuo_gold":       16,
			"wutong_gold":      17,
			"liangsuo_gold":    18,
			"liangtong_gold":   19,
		}

		for symbol, expected := range testCases {
			result := SymbolNumber(symbol)
			assert.Equal(t, expected, result,
				"Gold variant %s should map to number %d", symbol, expected)
		}
	})

	t.Run("should return default for unknown symbol", func(t *testing.T) {
		// Unknown symbols default to 9 (liangtong)
		assert.Equal(t, 9, SymbolNumber("unknown"))
		assert.Equal(t, 9, SymbolNumber("invalid_symbol"))
		assert.Equal(t, 9, SymbolNumber(""))
	})

	t.Run("should validate gold variants offset by 10", func(t *testing.T) {
		baseSymbols := []string{
			"fa", "zhong", "bai", "bawan",
			"wusuo", "wutong", "liangsuo", "liangtong",
		}

		for _, base := range baseSymbols {
			baseNum := SymbolNumber(base)
			goldNum := SymbolNumber(base + "_gold")

			// Gold variants should be base + 10
			assert.Equal(t, baseNum+10, goldNum,
				"Gold variant of %s should be base number + 10", base)
		}
	})
}

func TestPayingSymbols(t *testing.T) {
	t.Run("should return all 8 paying symbols", func(t *testing.T) {
		symbols := PayingSymbols()
		assert.Len(t, symbols, 8, "Should return exactly 8 paying symbols")
	})

	t.Run("should include all expected paying symbols", func(t *testing.T) {
		symbols := PayingSymbols()

		expectedSymbols := []Symbol{
			SymbolFa, SymbolZhong, SymbolBai, SymbolBawan,
			SymbolWusuo, SymbolWutong, SymbolLiangsuo, SymbolLiangtong,
		}

		for _, expected := range expectedSymbols {
			assert.Contains(t, symbols, expected,
				"PayingSymbols should include %s", expected)
		}
	})

	t.Run("should not include special symbols", func(t *testing.T) {
		symbols := PayingSymbols()

		assert.NotContains(t, symbols, SymbolWild)
		assert.NotContains(t, symbols, SymbolBonus)
		assert.NotContains(t, symbols, SymbolGold)
	})

	t.Run("should validate all returned symbols have payouts", func(t *testing.T) {
		symbols := PayingSymbols()

		for _, sym := range symbols {
			payout := GetPayout(sym, 3)
			assert.Greater(t, payout, 0.0,
				"PayingSymbol %s should have a payout", sym)
		}
	})
}

func TestAllSymbols(t *testing.T) {
	t.Run("should return all 11 symbols", func(t *testing.T) {
		symbols := AllSymbols()
		assert.Len(t, symbols, 11, "Should return exactly 11 symbols (3 special + 8 paying)")
	})

	t.Run("should include special symbols", func(t *testing.T) {
		symbols := AllSymbols()

		assert.Contains(t, symbols, SymbolWild)
		assert.Contains(t, symbols, SymbolBonus)
		assert.Contains(t, symbols, SymbolGold)
	})

	t.Run("should include all paying symbols", func(t *testing.T) {
		allSyms := AllSymbols()
		payingSyms := PayingSymbols()

		for _, paying := range payingSyms {
			assert.Contains(t, allSyms, paying,
				"AllSymbols should include paying symbol %s", paying)
		}
	})
}

func TestIsPayingSymbol(t *testing.T) {
	t.Run("should return true for all paying symbols", func(t *testing.T) {
		payingSymbols := []Symbol{
			SymbolFa, SymbolZhong, SymbolBai, SymbolBawan,
			SymbolWusuo, SymbolWutong, SymbolLiangsuo, SymbolLiangtong,
		}

		for _, sym := range payingSymbols {
			assert.True(t, IsPayingSymbol(sym),
				"Symbol %s should be a paying symbol", sym)
		}
	})

	t.Run("should return false for special symbols", func(t *testing.T) {
		assert.False(t, IsPayingSymbol(SymbolWild))
		assert.False(t, IsPayingSymbol(SymbolBonus))
		assert.False(t, IsPayingSymbol(SymbolGold))
	})

	t.Run("should return false for unknown symbols", func(t *testing.T) {
		assert.False(t, IsPayingSymbol(Symbol("unknown")))
		assert.False(t, IsPayingSymbol(Symbol("invalid_symbol")))
		assert.False(t, IsPayingSymbol(Symbol("")))
	})

	t.Run("should match PayingSymbols list", func(t *testing.T) {
		payingList := PayingSymbols()

		for _, sym := range payingList {
			assert.True(t, IsPayingSymbol(sym),
				"Symbol %s from PayingSymbols() should return true in IsPayingSymbol()", sym)
		}
	})
}

func TestIsSpecialSymbol(t *testing.T) {
	t.Run("should return true for special symbols", func(t *testing.T) {
		assert.True(t, IsSpecialSymbol(SymbolWild))
		assert.True(t, IsSpecialSymbol(SymbolBonus))
		assert.True(t, IsSpecialSymbol(SymbolGold))
	})

	t.Run("should return false for paying symbols", func(t *testing.T) {
		payingSymbols := []Symbol{
			SymbolFa, SymbolZhong, SymbolBai, SymbolBawan,
			SymbolWusuo, SymbolWutong, SymbolLiangsuo, SymbolLiangtong,
		}

		for _, sym := range payingSymbols {
			assert.False(t, IsSpecialSymbol(sym),
				"Symbol %s should not be a special symbol", sym)
		}
	})

	t.Run("should return false for unknown symbols", func(t *testing.T) {
		assert.False(t, IsSpecialSymbol(Symbol("unknown")))
		assert.False(t, IsSpecialSymbol(Symbol("invalid_symbol")))
		assert.False(t, IsSpecialSymbol(Symbol("")))
	})

	t.Run("should be mutually exclusive with IsPayingSymbol", func(t *testing.T) {
		allSyms := AllSymbols()

		for _, sym := range allSyms {
			// A symbol should be either special OR paying, not both
			isSpecial := IsSpecialSymbol(sym)
			isPaying := IsPayingSymbol(sym)

			// XOR: exactly one should be true
			assert.NotEqual(t, isSpecial, isPaying,
				"Symbol %s should be either special XOR paying, not both or neither", sym)
		}
	})
}

func TestCanBeSubstituted(t *testing.T) {
	t.Run("should return true for all paying symbols", func(t *testing.T) {
		payingSymbols := []Symbol{
			SymbolFa, SymbolZhong, SymbolBai, SymbolBawan,
			SymbolWusuo, SymbolWutong, SymbolLiangsuo, SymbolLiangtong,
		}

		for _, sym := range payingSymbols {
			assert.True(t, CanBeSubstituted(sym),
				"Wild should be able to substitute for paying symbol %s", sym)
		}
	})

	t.Run("should return false for bonus (scatter)", func(t *testing.T) {
		assert.False(t, CanBeSubstituted(SymbolBonus),
			"Wild should NOT substitute for bonus (scatter)")
	})

	t.Run("should return false for wild itself", func(t *testing.T) {
		assert.False(t, CanBeSubstituted(SymbolWild),
			"Wild should NOT substitute for wild")
	})

	t.Run("should return false for gold", func(t *testing.T) {
		assert.False(t, CanBeSubstituted(SymbolGold),
			"Wild should NOT substitute for gold")
	})

	t.Run("should match IsPayingSymbol logic", func(t *testing.T) {
		allSyms := AllSymbols()

		for _, sym := range allSyms {
			expected := IsPayingSymbol(sym)
			actual := CanBeSubstituted(sym)

			assert.Equal(t, expected, actual,
				"CanBeSubstituted(%s) should match IsPayingSymbol(%s)", sym, sym)
		}
	})
}

func TestHasGoldVariant(t *testing.T) {
	t.Run("should return true for all paying symbols", func(t *testing.T) {
		payingSymbols := []Symbol{
			SymbolFa, SymbolZhong, SymbolBai, SymbolBawan,
			SymbolWusuo, SymbolWutong, SymbolLiangsuo, SymbolLiangtong,
		}

		for _, sym := range payingSymbols {
			assert.True(t, HasGoldVariant(sym),
				"Symbol %s should have a gold variant", sym)
		}
	})

	t.Run("should return false for special symbols", func(t *testing.T) {
		assert.False(t, HasGoldVariant(SymbolWild))
		assert.False(t, HasGoldVariant(SymbolBonus))
		assert.False(t, HasGoldVariant(SymbolGold))
	})

	t.Run("should match IsPayingSymbol logic", func(t *testing.T) {
		allSyms := AllSymbols()

		for _, sym := range allSyms {
			expected := IsPayingSymbol(sym)
			actual := HasGoldVariant(sym)

			assert.Equal(t, expected, actual,
				"HasGoldVariant(%s) should match IsPayingSymbol(%s)", sym, sym)
		}
	})
}

func TestGetBaseSymbol(t *testing.T) {
	t.Run("should remove _gold suffix from gold variants", func(t *testing.T) {
		testCases := map[string]Symbol{
			"fa_gold":         SymbolFa,
			"zhong_gold":      SymbolZhong,
			"bai_gold":        SymbolBai,
			"bawan_gold":      SymbolBawan,
			"wusuo_gold":      SymbolWusuo,
			"wutong_gold":     SymbolWutong,
			"liangsuo_gold":   SymbolLiangsuo,
			"liangtong_gold":  SymbolLiangtong,
		}

		for goldVariant, expectedBase := range testCases {
			result := GetBaseSymbol(goldVariant)
			assert.Equal(t, expectedBase, result,
				"Gold variant %s should return base symbol %s", goldVariant, expectedBase)
		}
	})

	t.Run("should return symbol unchanged if not gold variant", func(t *testing.T) {
		baseSymbols := []string{
			"wild", "bonus", "fa", "zhong", "bai", "bawan",
			"wusuo", "wutong", "liangsuo", "liangtong",
		}

		for _, sym := range baseSymbols {
			result := GetBaseSymbol(sym)
			assert.Equal(t, Symbol(sym), result,
				"Base symbol %s should be returned unchanged", sym)
		}
	})

	t.Run("should handle edge cases", func(t *testing.T) {
		// Short strings that can't be gold variants
		assert.Equal(t, Symbol("gold"), GetBaseSymbol("gold"))
		assert.Equal(t, Symbol("abc"), GetBaseSymbol("abc"))
		assert.Equal(t, Symbol(""), GetBaseSymbol(""))

		// Strings ending with _gold - the function strips _gold suffix regardless
		// This is by design - it's a simple string manipulation function
		assert.Equal(t, Symbol("x"), GetBaseSymbol("x_gold"))
		assert.Equal(t, Symbol("invalid"), GetBaseSymbol("invalid_gold"))
	})

	t.Run("should be consistent with IsGoldVariant", func(t *testing.T) {
		goldVariants := []string{
			"fa_gold", "zhong_gold", "bai_gold", "bawan_gold",
			"wusuo_gold", "wutong_gold", "liangsuo_gold", "liangtong_gold",
		}

		for _, variant := range goldVariants {
			base := GetBaseSymbol(variant)

			// Base symbol should not be a gold variant
			assert.False(t, IsGoldVariant(string(base)),
				"GetBaseSymbol(%s) should return a non-gold symbol", variant)
		}
	})
}

func TestIsGoldVariant(t *testing.T) {
	t.Run("should return true for gold variants", func(t *testing.T) {
		goldVariants := []string{
			"fa_gold", "zhong_gold", "bai_gold", "bawan_gold",
			"wusuo_gold", "wutong_gold", "liangsuo_gold", "liangtong_gold",
		}

		for _, variant := range goldVariants {
			assert.True(t, IsGoldVariant(variant),
				"%s should be identified as gold variant", variant)
		}
	})

	t.Run("should return false for base symbols", func(t *testing.T) {
		baseSymbols := []string{
			"wild", "bonus", "gold", "fa", "zhong", "bai", "bawan",
			"wusuo", "wutong", "liangsuo", "liangtong",
		}

		for _, sym := range baseSymbols {
			assert.False(t, IsGoldVariant(sym),
				"%s should NOT be identified as gold variant", sym)
		}
	})

	t.Run("should return false for edge cases", func(t *testing.T) {
		// Short strings
		assert.False(t, IsGoldVariant(""))
		assert.False(t, IsGoldVariant("gold"))
		assert.False(t, IsGoldVariant("abc"))

		// Strings with _gold but too short
		assert.False(t, IsGoldVariant("_gold")) // length = 5, not > 5
	})

	t.Run("should validate all gold variants in SymbolNumber", func(t *testing.T) {
		goldVariants := []string{
			"fa_gold", "zhong_gold", "bai_gold", "bawan_gold",
			"wusuo_gold", "wutong_gold", "liangsuo_gold", "liangtong_gold",
		}

		for _, variant := range goldVariants {
			// Should be identified as gold variant
			assert.True(t, IsGoldVariant(variant))

			// Should have a number mapping >= 12
			num := SymbolNumber(variant)
			assert.GreaterOrEqual(t, num, 12,
				"Gold variant %s should map to number >= 12", variant)
		}
	})
}

func TestSymbolConsistency(t *testing.T) {
	t.Run("should validate PayingSymbols matches Paytable", func(t *testing.T) {
		payingSyms := PayingSymbols()

		// Every paying symbol should have paytable entry
		for _, sym := range payingSyms {
			payout := GetPayout(sym, 3)
			assert.Greater(t, payout, 0.0,
				"Paying symbol %s should have paytable entry", sym)
		}

		// Every paytable symbol should be in PayingSymbols
		for sym := range Paytable {
			assert.Contains(t, payingSyms, sym,
				"Paytable symbol %s should be in PayingSymbols list", sym)
		}
	})

	t.Run("should validate symbol counts match expectations", func(t *testing.T) {
		// 3 special + 8 paying = 11 total
		assert.Len(t, AllSymbols(), 11)
		assert.Len(t, PayingSymbols(), 8)

		// Special symbols count
		specialCount := 0
		for _, sym := range AllSymbols() {
			if IsSpecialSymbol(sym) {
				specialCount++
			}
		}
		assert.Equal(t, 3, specialCount)
	})

	t.Run("should validate all symbols have unique Symbol constants", func(t *testing.T) {
		allSyms := AllSymbols()
		unique := make(map[Symbol]bool)

		for _, sym := range allSyms {
			assert.False(t, unique[sym], "Symbol %s should be unique", sym)
			unique[sym] = true
		}
	})
}

func BenchmarkSymbolNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SymbolNumber("fa_gold")
	}
}

func BenchmarkIsPayingSymbol(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = IsPayingSymbol(SymbolFa)
	}
}

func BenchmarkGetBaseSymbol(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetBaseSymbol("fa_gold")
	}
}

func BenchmarkIsGoldVariant(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = IsGoldVariant("fa_gold")
	}
}
