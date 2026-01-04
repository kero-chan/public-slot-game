package reels

import (
	"fmt"

	"github.com/slotmachine/backend/internal/game/rng"
	"github.com/slotmachine/backend/internal/game/symbols"
)

// Grid represents the game grid (5 reels × 10 rows)
// Stored as [reel][row] (column-major format)
// Rows 0-3 are buffer rows, rows 4-9 are visible (with partial top/bottom)
type Grid [][]string

// GenerateGrid generates a new grid from reel strips using random reel positions
func GenerateGrid(strips []ReelStrip, rngInstance rng.RNG) (Grid, []int, error) {
	if len(strips) != ReelCount {
		return nil, nil, fmt.Errorf("expected %d reel strips, got %d", ReelCount, len(strips))
	}

	grid := make(Grid, ReelCount)
	reelPositions := make([]int, ReelCount)

	// Generate random position for each reel
	for reelIdx := 0; reelIdx < ReelCount; reelIdx++ {
		// Generate random start position (0 to strip length-1)
		stripLength := len(strips[reelIdx])
		position, err := rngInstance.Int(stripLength)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate random position for reel %d: %w", reelIdx, err)
		}

		reelPositions[reelIdx] = position

		// Get TotalRows symbols starting from this position
		grid[reelIdx] = strips[reelIdx].GetSymbolsFromPosition(position, TotalRows)
	}

	return grid, reelPositions, nil
}

// GetSymbol returns the symbol at [reel][row]
func (g Grid) GetSymbol(reel, row int) string {
	if reel < 0 || reel >= len(g) {
		return ""
	}
	if row < 0 || row >= len(g[reel]) {
		return ""
	}
	return g[reel][row]
}

// SetSymbol sets the symbol at [reel][row]
func (g Grid) SetSymbol(reel, row int, symbol string) {
	if reel >= 0 && reel < len(g) && row >= 0 && row < len(g[reel]) {
		g[reel][row] = symbol
	}
}

// Clone creates a deep copy of the grid
func (g Grid) Clone() Grid {
	clone := make(Grid, len(g))
	for i := range g {
		clone[i] = make([]string, len(g[i]))
		copy(clone[i], g[i])
	}
	return clone
}

// ToJSON converts grid to JSON-serializable format (5x10 array)
func (g Grid) ToJSON() [][]string {
	return [][]string(g)
}

// FromJSON creates a grid from JSON format
func FromJSON(data [][]string) Grid {
	return Grid(data)
}

// GetVisibleGrid returns only the visible portion (4 rows)
func (g Grid) GetVisibleGrid() Grid {
	visible := make(Grid, ReelCount)
	for i := 0; i < ReelCount; i++ {
		visible[i] = g[i][:VisibleRows]
	}
	return visible
}

// CountSymbol counts occurrences of a symbol in the visible grid
func (g Grid) CountSymbol(symbol string) int {
	count := 0
	for reel := 0; reel < ReelCount; reel++ {
		for row := 0; row < VisibleRows; row++ {
			// Compare base symbol (remove _gold suffix if present)
			gridSymbol := g.GetSymbol(reel, row)
			baseSymbol := symbols.GetBaseSymbol(gridSymbol)
			if string(baseSymbol) == symbol {
				count++
			}
		}
	}
	return count
}
