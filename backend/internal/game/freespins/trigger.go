package freespins

import (
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/symbols"
)

// TriggerResult represents the result of checking for free spins trigger
type TriggerResult struct {
	Triggered    bool `json:"triggered"`
	ScatterCount int  `json:"scatter_count"`
	SpinsAwarded int  `json:"spins_awarded"`
}

// CheckTrigger checks if free spins are triggered
// Free spins trigger when 3+ bonus (scatter) symbols appear
func CheckTrigger(grid reels.Grid) TriggerResult {
	scatterCount := countScatters(grid)

	if scatterCount >= symbols.MinScattersForFreeSpin() {
		spinsAwarded := symbols.GetFreeSpinsAward(scatterCount)
		return TriggerResult{
			Triggered:    true,
			ScatterCount: scatterCount,
			SpinsAwarded: spinsAwarded,
		}
	}

	return TriggerResult{
		Triggered:    false,
		ScatterCount: scatterCount,
		SpinsAwarded: 0,
	}
}

// countScatters counts the number of bonus (scatter) symbols in the visible grid
func countScatters(grid reels.Grid) int {
	count := 0

	for reel := 0; reel < reels.ReelCount; reel++ {
		for row := reels.WinCheckStartRow; row <= reels.WinCheckEndRow; row++ {
			symbolStr := grid.GetSymbol(reel, row)
			baseSymbol := symbols.GetBaseSymbol(symbolStr)

			if baseSymbol == symbols.SymbolBonus {
				count++
			}
		}
	}

	return count
}

// GetScatterPositions returns positions of all scatter symbols
func GetScatterPositions(grid reels.Grid) []Position {
	positions := make([]Position, 0)

	for reel := 0; reel < reels.ReelCount; reel++ {
		for row := reels.WinCheckStartRow; row <= reels.WinCheckEndRow; row++ {
			symbolStr := grid.GetSymbol(reel, row)
			baseSymbol := symbols.GetBaseSymbol(symbolStr)

			if baseSymbol == symbols.SymbolBonus {
				positions = append(positions, Position{
					Reel: reel,
					Row:  row,
				})
			}
		}
	}

	return positions
}

// Position represents a grid position
type Position struct {
	Reel int `json:"reel"`
	Row  int `json:"row"`
}

// CalculateFreeSpinsAward calculates free spins award for scatter count
// Uses the formula from symbols package
func CalculateFreeSpinsAward(scatterCount int) int {
	return symbols.GetFreeSpinsAward(scatterCount)
}
