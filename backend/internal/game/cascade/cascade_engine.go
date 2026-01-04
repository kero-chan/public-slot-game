package cascade

import (
	"github.com/slotmachine/backend/internal/game/multiplier"
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/rng"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/slotmachine/backend/internal/game/wins"
)

const EmptySymbol = ""

// CascadeResult represents the result of a single cascade
type CascadeResult struct {
	CascadeNumber   int                     `json:"cascade_number"`
	GridAfter       reels.Grid              `json:"grid_after"` // Grid with 10 rows (4 buffer + 6 visible)
	Wins            []wins.CascadeWinDetail `json:"wins"`
	TotalCascadeWin float64                 `json:"total_cascade_win"`
	Multiplier      int                     `json:"multiplier"`
	WinningSymbols  []symbols.Symbol        `json:"winning_symbols"`
}

// ExecuteCascades executes all cascades for a spin
// Returns all cascade results and the final grid
func ExecuteCascades(
	initialGrid reels.Grid,
	reelStrips []reels.ReelStrip,
	reelPositions []int,
	betAmount float64,
	isFreeSpin bool,
	rngInstance rng.RNG,
) ([]CascadeResult, reels.Grid, error) {
	cascadeResults := make([]CascadeResult, 0)
	currentGrid := initialGrid.Clone()
	cascadeNumber := 0

	// Execute first cascade (initial grid evaluation)
	for {
		cascadeNumber++

		// Calculate win for this cascade
		winDetails, symbolWins, totalWin := wins.CalculateCascadeWin(currentGrid, betAmount, cascadeNumber, isFreeSpin)
		if len(symbolWins) == 0 {
			// No wins, cascade sequence ends
			break
		}

		// Get winning symbols for position tracking
		winningSymbols := make([]symbols.Symbol, 0)
		for _, w := range symbolWins {
			winningSymbols = append(winningSymbols, w.Symbol)
		}

		// Remove winning symbols
		currentGrid = removeWinningSymbols(currentGrid, symbolWins)

		// Drop symbols down (gravity)
		currentGrid = dropSymbols(currentGrid)

		// Fill empty positions from reel strips AND get buffer symbols
		currentGrid, reelPositions = fillEmptyPositions(currentGrid, reelStrips, reelPositions)

		// Store cascade result with extended grid
		cascadeResult := CascadeResult{
			CascadeNumber:   cascadeNumber,
			GridAfter:       currentGrid, // Variable-length columns for smooth animation
			Wins:            winDetails,
			TotalCascadeWin: totalWin,
			Multiplier:      multiplier.GetMultiplier(cascadeNumber, isFreeSpin),
			WinningSymbols:  winningSymbols,
		}

		cascadeResults = append(cascadeResults, cascadeResult)

		// Continue to next cascade
	}

	return cascadeResults, currentGrid, nil
}

// removeWinningSymbols removes all winning symbols from the grid
func removeWinningSymbols(grid reels.Grid, symbolWins []wins.SymbolWin) reels.Grid {
	newGrid := grid.Clone()

	for _, win := range symbolWins {
		// Get all winning positions for this symbol
		positions := wins.GetWinningPositions(grid, win.Symbol, win.Count)

		// Remove symbols at winning positions
		for _, pos := range positions {
			sym := grid.GetSymbol(pos.Reel, pos.Row)
			if symbols.IsGoldVariant(sym) {
				newGrid.SetSymbol(pos.Reel, pos.Row, string(symbols.SymbolWild))
			} else {
				newGrid.SetSymbol(pos.Reel, pos.Row, EmptySymbol)
			}
		}
	}

	return newGrid
}

// dropSymbols applies gravity - symbols drop down to fill empty spaces
func dropSymbols(grid reels.Grid) reels.Grid {
	newGrid := grid.Clone()

	// Process each reel independently
	for reelIdx := 0; reelIdx < reels.ReelCount; reelIdx++ {
		// Collect non-empty symbols from bottom to top
		nonEmptySymbols := make([]string, 0)
		for row := reels.TotalRows - 1; row >= 0; row-- {
			symbol := newGrid.GetSymbol(reelIdx, row)
			if symbol != EmptySymbol {
				nonEmptySymbols = append(nonEmptySymbols, symbol)
			}
		}

		// Place non-empty symbols at bottom, empty at top
		// Fill from bottom up with non-empty symbols
		for row := reels.TotalRows - 1; row >= 0; row-- {
			bottomIndex := reels.TotalRows - 1 - row
			if bottomIndex < len(nonEmptySymbols) {
				newGrid.SetSymbol(reelIdx, row, nonEmptySymbols[bottomIndex])
			} else {
				newGrid.SetSymbol(reelIdx, row, EmptySymbol)
			}
		}
	}

	return newGrid
}

// fillEmptyPositions fills empty positions from reel strips
func fillEmptyPositions(
	grid reels.Grid,
	reelStrips []reels.ReelStrip,
	reelPositions []int,
) (reels.Grid, []int) {
	newGrid := grid.Clone()
	newPositions := make([]int, len(reelPositions))
	copy(newPositions, reelPositions)

	// Process each reel
	for reelIdx := 0; reelIdx < reels.ReelCount; reelIdx++ {
		// Count empty positions in this reel
		emptyCount := 0
		for row := 0; row < reels.TotalRows; row++ {
			if newGrid.GetSymbol(reelIdx, row) == EmptySymbol {
				emptyCount++
			}
		}

		// Fill empty positions from reel strip
		if emptyCount > 0 {
			// Advance reel position
			stripLength := len(reelStrips[reelIdx])
			newPositions[reelIdx] = (newPositions[reelIdx] - emptyCount + stripLength) % stripLength

			// Get new symbols from strip
			newSymbols := reelStrips[reelIdx].GetSymbolsFromPosition(newPositions[reelIdx], emptyCount)

			// Fill from top
			for i, newSymbol := range newSymbols {
				newGrid.SetSymbol(reelIdx, i, newSymbol)
			}
		}
	}

	return newGrid, newPositions
}

// GetTotalWinFromCascades calculates total win across all cascades
func GetTotalWinFromCascades(cascadeResults []CascadeResult, betAmount float64) float64 {
	totalWin := 0.0
	for _, cascade := range cascadeResults {
		totalWin += cascade.TotalCascadeWin
	}

	return wins.ApplyMaxWinCap(totalWin, betAmount)
}
