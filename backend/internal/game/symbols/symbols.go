package symbols

// Symbol type represents a symbol ID
type Symbol string

// WinIntensity represents the visual intensity level of a win
type WinIntensity string

// WinIntensity constants for animation/sound purposes
const (
	WinIntensitySmall  WinIntensity = "small"
	WinIntensityMedium WinIntensity = "medium"
	WinIntensityBig    WinIntensity = "big"
	WinIntensityMega   WinIntensity = "mega"
)

// Symbol constants
const (
	// Special symbols
	SymbolWild  Symbol = "wild"
	SymbolBonus Symbol = "bonus" // Scatter (triggers free spins)
	SymbolGold  Symbol = "gold"  // Mystery symbol

	// Aliases for easier access
	Wild  = SymbolWild
	Bonus = SymbolBonus
	Gold  = SymbolGold

	// High-value symbols (can have _gold variant)
	SymbolFa    Symbol = "fa"    // 发 - Highest value
	SymbolZhong Symbol = "zhong" // 中 - Premium high
	SymbolBai   Symbol = "bai"   // 白/百 - Premium mid
	SymbolBawan Symbol = "bawan" // 八萬 - Medium high

	// Low-value symbols (can have _gold variant)
	SymbolWusuo     Symbol = "wusuo"     // 五索 - Medium
	SymbolWutong    Symbol = "wutong"    // 五筒 - Medium
	SymbolLiangsuo  Symbol = "liangsuo"  // 两索 - Low
	SymbolLiangtong Symbol = "liangtong" // 两筒 - Low
)

func SymbolNumber(sym string) int {
	switch sym {
	case "wild":
		return 0
	case "bonus":
		return 1
	case "fa":
		return 2
	case "fa_gold":
		return 12
	case "zhong":
		return 3
	case "zhong_gold":
		return 13
	case "bai":
		return 4
	case "bai_gold":
		return 14
	case "bawan":
		return 5
	case "bawan_gold":
		return 15
	case "wusuo":
		return 6
	case "wusuo_gold":
		return 16
	case "wutong":
		return 7
	case "wutong_gold":
		return 17
	case "liangsuo":
		return 8
	case "liangsuo_gold":
		return 18
	case "liangtong":
		return 9
	case "liangtong_gold":
		return 19
	default:
		return 9
	}
}

// PayingSymbols returns all symbols that can award payouts
func PayingSymbols() []Symbol {
	return []Symbol{
		SymbolFa,
		SymbolZhong,
		SymbolBai,
		SymbolBawan,
		SymbolWusuo,
		SymbolWutong,
		SymbolLiangsuo,
		SymbolLiangtong,
	}
}

// AllSymbols returns all symbols including special ones
func AllSymbols() []Symbol {
	return []Symbol{
		SymbolWild,
		SymbolBonus,
		SymbolGold,
		SymbolFa,
		SymbolZhong,
		SymbolBai,
		SymbolBawan,
		SymbolWusuo,
		SymbolWutong,
		SymbolLiangsuo,
		SymbolLiangtong,
	}
}

// IsPayingSymbol checks if a symbol awards payouts
func IsPayingSymbol(sym Symbol) bool {
	switch sym {
	case SymbolFa, SymbolZhong, SymbolBai, SymbolBawan,
		SymbolWusuo, SymbolWutong, SymbolLiangsuo, SymbolLiangtong:
		return true
	default:
		return false
	}
}

// IsSpecialSymbol checks if a symbol is special (wild, bonus, gold)
func IsSpecialSymbol(sym Symbol) bool {
	switch sym {
	case SymbolWild, SymbolBonus, SymbolGold:
		return true
	default:
		return false
	}
}

// CanBeSubstituted checks if wild can substitute for this symbol
func CanBeSubstituted(sym Symbol) bool {
	// Wild substitutes for all paying symbols but NOT bonus or gold
	return IsPayingSymbol(sym)
}

// HasGoldVariant checks if a symbol can have a gold variant
func HasGoldVariant(sym Symbol) bool {
	// All paying symbols can have gold variants
	// Gold variants only appear on reels 2, 3, 4
	return IsPayingSymbol(sym)
}

// GetBaseSymbol removes the _gold suffix if present
func GetBaseSymbol(symbolStr string) Symbol {
	// Remove _gold suffix
	if len(symbolStr) > 5 && symbolStr[len(symbolStr)-5:] == "_gold" {
		return Symbol(symbolStr[:len(symbolStr)-5])
	}
	return Symbol(symbolStr)
}

// IsGoldVariant checks if a symbol string is a gold variant
func IsGoldVariant(symbolStr string) bool {
	return len(symbolStr) > 5 && symbolStr[len(symbolStr)-5:] == "_gold"
}

// IsHighValueSymbol checks if a symbol is high-value (for win intensity calculation)
// High-value symbols: fa, zhong, bai
func IsHighValueSymbol(sym Symbol) bool {
	switch sym {
	case SymbolFa, SymbolZhong, SymbolBai:
		return true
	default:
		return false
	}
}

// GetWinIntensity calculates the win intensity based on symbol and count
// Rules:
//   - small:  3 of any symbol
//   - medium: 4 of a low-value symbol
//   - big:    5+ of a low-value symbol, OR 4 of a high-value symbol
//   - mega:   5+ of a high-value symbol
func GetWinIntensity(sym Symbol, count int) WinIntensity {
	// Get base symbol (remove _gold suffix if present)
	baseSym := GetBaseSymbol(string(sym))
	isHighValue := IsHighValueSymbol(baseSym)

	if count >= 5 && isHighValue {
		return WinIntensityMega
	} else if count >= 5 || (count >= 4 && isHighValue) {
		return WinIntensityBig
	} else if count >= 4 {
		return WinIntensityMedium
	}
	return WinIntensitySmall
}

// RandomGenerator interface for random number generation
type RandomGenerator interface {
	Intn(n int) int
}

// NonBonusSymbols returns all symbols except bonus (for replacement purposes)
func NonBonusSymbols() []Symbol {
	return []Symbol{
		SymbolWild,
		SymbolGold,
		SymbolFa,
		SymbolZhong,
		SymbolBai,
		SymbolBawan,
		SymbolWusuo,
		SymbolWutong,
		SymbolLiangsuo,
		SymbolLiangtong,
	}
}

// GetRandomNonBonusSymbol returns a random symbol that is not bonus
func GetRandomNonBonusSymbol(rng RandomGenerator) Symbol {
	symbols := NonBonusSymbols()
	return symbols[rng.Intn(len(symbols))]
}
