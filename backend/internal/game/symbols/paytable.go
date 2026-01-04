package symbols

// SymbolPayout represents payouts for different symbol counts
type SymbolPayout struct {
	Symbol  Symbol
	Payouts map[int]float64 // map[symbolCount]multiplier
}

// Paytable defines the complete paytable for the game
var Paytable = map[Symbol]map[int]float64{
	// High-value symbols
	SymbolFa: {
		3: 10.0,
		4: 25.0,
		5: 50.0,
	},
	SymbolZhong: {
		3: 8.0,
		4: 20.0,
		5: 40.0,
	},
	SymbolBai: {
		3: 6.0,
		4: 15.0,
		5: 30.0,
	},
	SymbolBawan: {
		3: 5.0,
		4: 10.0,
		5: 15.0,
	},

	// Low-value symbols
	SymbolWusuo: {
		3: 3.0,
		4: 5.0,
		5: 12.0,
	},
	SymbolWutong: {
		3: 3.0,
		4: 5.0,
		5: 12.0,
	},
	SymbolLiangsuo: {
		3: 2.0,
		4: 4.0,
		5: 10.0,
	},
	SymbolLiangtong: {
		3: 1.0,
		4: 3.0,
		5: 6.0,
	},
}

// GetPayout returns the payout multiplier for a symbol and count
func GetPayout(sym Symbol, count int) float64 {
	if payouts, ok := Paytable[sym]; ok {
		if payout, ok := payouts[count]; ok {
			return payout
		}
	}
	return 0.0
}

// MinSymbolsForPayout returns the minimum symbols needed for a payout
func MinSymbolsForPayout() int {
	return 3
}

// MaxSymbolsForPayout returns the maximum symbols that can award payout
func MaxSymbolsForPayout() int {
	return 5 // 5 reels
}

// FreeSpinsAward defines free spins awarded for scatter counts
var FreeSpinsAward = map[int]int{
	3: 12, // 3 scatters = 12 free spins
	4: 14, // 4 scatters = 14 free spins (12 + 2)
	5: 16, // 5 scatters = 16 free spins (12 + 4)
}

// GetFreeSpinsAward returns the number of free spins for scatter count
// Formula: 12 + (2 × (scatter_count - 3))
func GetFreeSpinsAward(scatterCount int) int {
	if scatterCount < 3 {
		return 0
	}
	if award, ok := FreeSpinsAward[scatterCount]; ok {
		return award
	}
	// For counts > 5, use formula
	return 12 + (2 * (scatterCount - 3))
}

// MinScattersForFreeSpin returns minimum scatters to trigger free spins
func MinScattersForFreeSpin() int {
	return 3
}
