package wins

const (
	// MaxWinMultiplier is the maximum win as a multiple of bet amount
	// Regulatory requirement: wins cannot exceed 25,000x bet
	MaxWinMultiplier = 25000
)

// ApplyMaxWinCap applies the maximum win cap of 25,000x bet
// If win exceeds the cap, it is reduced to the maximum allowed
func ApplyMaxWinCap(winAmount, betAmount float64) float64 {
	maxWin := betAmount * MaxWinMultiplier

	if winAmount > maxWin {
		return maxWin
	}

	return winAmount
}

// IsWinCapped checks if a win amount exceeds the cap
func IsWinCapped(winAmount, betAmount float64) bool {
	maxWin := betAmount * MaxWinMultiplier
	return winAmount > maxWin
}

// GetMaxWinForBet returns the maximum possible win for a bet amount
func GetMaxWinForBet(betAmount float64) float64 {
	return betAmount * MaxWinMultiplier
}
