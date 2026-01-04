package wins

import (
	"github.com/slotmachine/backend/internal/game/multiplier"
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/symbols"
)

// CascadeWinDetail represents win details for a single cascade
type CascadeWinDetail struct {
	Symbol    symbols.Symbol `json:"symbol"`
	Count     int            `json:"count"`      // 3, 4, or 5
	Ways      int            `json:"ways"`       // Number of ways
	Payout    float64        `json:"payout"`     // Base payout multiplier from paytable
	WinAmount float64        `json:"win_amount"` // Actual win amount for this symbol
	Positions []Position     `json:"positions"`  // Grid positions that form this win
}

// CalculateCascadeWin calculates the total win for a single cascade
// Returns individual symbol wins and total cascade win
func CalculateCascadeWin(grid reels.Grid, betAmount float64, cascadeNumber int, isFreeSpin bool) ([]CascadeWinDetail, []SymbolWin, float64) {
	// Get multiplier for this cascade
	cascadeMultiplier := multiplier.GetMultiplier(cascadeNumber, isFreeSpin)

	// Calculate ways for all symbols
	symbolWins := CalculateWays(grid)

	// Calculate win for each symbol
	winDetails := make([]CascadeWinDetail, 0)
	totalCascadeWin := 0.0

	for _, win := range symbolWins {
		// Get payout multiplier from paytable
		payoutMultiplier := symbols.GetPayout(win.Symbol, win.Count)
		if payoutMultiplier == 0 {
			continue
		}

		// Win calculation formula (spec 04-rtp-mathematics.md):
		// Win = Symbol_Payout × Ways_Count × Cascade_Multiplier × Bet_Per_Way
		// where Bet_Per_Way = Total_Bet / 20
		betPerWay := betAmount / 20.0
		winAmount := payoutMultiplier * float64(win.Ways) * float64(cascadeMultiplier) * betPerWay

		// Create positions with gold transformation flag set
		positions := make([]Position, len(win.Positions))
		for j, pos := range win.Positions {
			symbolStr := grid.GetSymbol(pos.Reel, pos.Row)
			positions[j] = Position{
				Reel:         pos.Reel,
				Row:          pos.Row,
				IsGoldToWild: symbols.IsGoldVariant(symbolStr),
			}
		}

		winDetail := CascadeWinDetail{
			Symbol:    win.Symbol,
			Count:     win.Count,
			Ways:      win.Ways,
			Payout:    payoutMultiplier,
			WinAmount: winAmount,
			Positions: positions, // Include positions with gold flag for frontend
		}

		winDetails = append(winDetails, winDetail)
		totalCascadeWin += winAmount
	}

	return winDetails, symbolWins, totalCascadeWin
}

// CalculateTotalSpinWin calculates the total win across all cascades
// and applies the max win cap
func CalculateTotalSpinWin(cascadeWins []float64, betAmount float64) float64 {
	totalWin := 0.0
	for _, cascadeWin := range cascadeWins {
		totalWin += cascadeWin
	}

	// Apply max win cap (25,000x bet)
	return ApplyMaxWinCap(totalWin, betAmount)
}

// GetWinningPositions returns positions of winning symbols
// Used for highlighting/animation
func GetWinningPositions(grid reels.Grid, targetSymbol symbols.Symbol, count int) []Position {
	positions := make([]Position, 0)

	// Get matching positions for each reel up to 'count' reels
	// Only check middle 4 fully visible rows (5-8)
	for reelIdx := 0; reelIdx < count && reelIdx < reels.ReelCount; reelIdx++ {
		for row := reels.WinCheckStartRow; row <= reels.WinCheckEndRow; row++ {
			symbolStr := grid.GetSymbol(reelIdx, row)
			baseSymbol := symbols.GetBaseSymbol(symbolStr)

			if baseSymbol == targetSymbol {
				positions = append(positions, Position{Reel: reelIdx, Row: row})
			} else if baseSymbol == symbols.SymbolWild && symbols.CanBeSubstituted(targetSymbol) {
				// Include wild substitutions
				positions = append(positions, Position{Reel: reelIdx, Row: row})
			}
		}
	}

	return positions
}
