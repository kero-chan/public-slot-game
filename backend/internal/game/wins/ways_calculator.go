package wins

import (
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/slotmachine/backend/internal/pkg/util"
)

// Position represents a grid position [reel, row]
type Position struct {
	Reel         int  `json:"reel"`
	Row          int  `json:"row"`
	IsGoldToWild bool `json:"is_gold_to_wild,omitempty"` // True if this gold tile transforms to wild
}

// SymbolWin represents a winning combination
type SymbolWin struct {
	Symbol    symbols.Symbol // Base symbol (without _gold suffix)
	Count     int            // Number of consecutive reels (3, 4, or 5)
	Ways      int            // Number of ways to achieve this win
	Positions []Position     // All matching positions that form this win
}

// CalculateWays calculates all winning combinations in a grid
// Returns a slice of SymbolWin for each winning symbol
func CalculateWays(grid reels.Grid) []SymbolWin {
	wins := make([]SymbolWin, 0)

	// Symbol must appear in reel 0 to check win
	// Only check middle 4 fully visible rows (5-8)
	needCheckWinSyms := util.UniqueSlice(grid[0][reels.WinCheckStartRow : reels.WinCheckEndRow+1])
	// Check each paying symbol (skip bonus/wild/gold - they don't create way wins)
	for _, sym := range needCheckWinSyms {
		baseSymbol := symbols.GetBaseSymbol(sym)

		// Only process paying symbols
		if !symbols.IsPayingSymbol(baseSymbol) {
			continue
		}

		win := calculateWaysForSymbol(grid, baseSymbol)
		if win.Ways > 0 && win.Count >= symbols.MinSymbolsForPayout() {
			wins = append(wins, win)
		}
	}

	return wins
}

// calculateWaysForSymbol calculates ways for a specific symbol
func calculateWaysForSymbol(grid reels.Grid, targetSymbol symbols.Symbol) SymbolWin {
	// Count matching positions on each reel starting from reel 0
	matchingPositions := make([][]int, reels.ReelCount)

	for reelIdx := 0; reelIdx < reels.ReelCount; reelIdx++ {
		matchingPositions[reelIdx] = getMatchingPositions(grid, reelIdx, targetSymbol)

		if len(matchingPositions[reelIdx]) == 0 {
			// Calculate win for previous consecutive reels
			if reelIdx >= symbols.MinSymbolsForPayout() {
				ways := calculateWaysProduct(matchingPositions[:reelIdx])
				positions := collectPositions(matchingPositions[:reelIdx])
				return SymbolWin{
					Symbol:    targetSymbol,
					Count:     reelIdx,
					Ways:      ways,
					Positions: positions,
				}
			}
			// Not enough consecutive reels for a win
			return SymbolWin{Symbol: targetSymbol, Count: 0, Ways: 0, Positions: nil}
		}
	}

	// All 5 reels have matches - 5-of-a-kind
	ways := calculateWaysProduct(matchingPositions)
	positions := collectPositions(matchingPositions)
	return SymbolWin{
		Symbol:    targetSymbol,
		Count:     reels.ReelCount,
		Ways:      ways,
		Positions: positions,
	}
}

// getMatchingPositions returns row indices where the symbol matches (including wild substitution)
func getMatchingPositions(grid reels.Grid, reelIdx int, targetSymbol symbols.Symbol) []int {
	positions := make([]int, 0)

	// Only check middle 4 fully visible rows (5-8) for winning positions
	for row := reels.WinCheckStartRow; row <= reels.WinCheckEndRow; row++ {
		symbolStr := grid.GetSymbol(reelIdx, row)
		baseSymbol := symbols.GetBaseSymbol(symbolStr)

		// Check if symbol matches or wild substitutes
		if baseSymbol == targetSymbol {
			positions = append(positions, row)
		} else if baseSymbol == symbols.SymbolWild && symbols.CanBeSubstituted(targetSymbol) {
			// Wild substitutes for paying symbols
			positions = append(positions, row)
		}
	}

	return positions
}

// calculateWaysProduct calculates the product of matching positions
// Ways = matches_reel1 × matches_reel2 × matches_reel3 × ...
func calculateWaysProduct(matchingPositions [][]int) int {
	if len(matchingPositions) == 0 {
		return 0
	}

	ways := 1
	for _, positions := range matchingPositions {
		ways *= len(positions)
	}

	return ways
}

// GetWinningSymbols returns all symbols that have wins in the grid
func GetWinningSymbols(grid reels.Grid) []symbols.Symbol {
	wins := CalculateWays(grid)
	winningSymbols := make([]symbols.Symbol, 0, len(wins))

	for _, win := range wins {
		if win.Ways > 0 {
			winningSymbols = append(winningSymbols, win.Symbol)
		}
	}

	return winningSymbols
}

// HasAnyWins checks if there are any wins in the grid
func HasAnyWins(grid reels.Grid) bool {
	wins := CalculateWays(grid)
	return len(wins) > 0
}

// collectPositions converts matchingPositions ([][]int) to []Position
// matchingPositions[reelIdx] = [row1, row2, ...] for each reel
// Returns all [reel, row] positions across all reels
func collectPositions(matchingPositions [][]int) []Position {
	positions := make([]Position, 0)

	for reelIdx, rows := range matchingPositions {
		for _, row := range rows {
			positions = append(positions, Position{
				Reel: reelIdx,
				Row:  row,
			})
		}
	}

	return positions
}
