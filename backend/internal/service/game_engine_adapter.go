package service

import (
	"github.com/slotmachine/backend/domain/spin"
	"github.com/slotmachine/backend/internal/game/cascade"
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/slotmachine/backend/internal/game/wins"
)

// Helper functions for conversion
// convertGrid converts engine grid to domain grid
func convertGrid(engineGrid reels.Grid) spin.Grid {
	output := make(spin.Grid, len(engineGrid))
	for i, row := range engineGrid {
		output[i] = make([]string, len(row))
		copy(output[i], row)
	}
	return output
}

// High-value symbols in priority order (highest first)
var highValueSymbolPriority = []symbols.Symbol{
	symbols.SymbolFa,
	symbols.SymbolZhong,
	symbols.SymbolBai,
	symbols.SymbolBawan,
}

// convertCascades converts engine cascades to domain cascades
func convertCascades(engineCascades []cascade.CascadeResult) spin.Cascades {
	result := make(spin.Cascades, len(engineCascades))
	for i, cascadeResult := range engineCascades {
		result[i] = spin.Cascade{
			CascadeNumber:   cascadeResult.CascadeNumber,
			GridAfter:       convertGrid(cascadeResult.GridAfter),
			Multiplier:      cascadeResult.Multiplier,
			Wins:            convertCascadeWins(cascadeResult.Wins),
			TotalCascadeWin: cascadeResult.TotalCascadeWin,
			WinningTileKind: extractHighestPriorityWinningSymbol(cascadeResult.Wins),
		}
	}
	return result
}

// extractHighestPriorityWinningSymbol extracts the highest priority winning symbol from cascade wins
// Priority: fa > zhong > bai > bawan
// Gold variants are normalized to base symbols, wild symbols are excluded
// Returns empty string if no high-value symbol wins
func extractHighestPriorityWinningSymbol(engineWins []wins.CascadeWinDetail) string {
	// Collect all winning base symbols
	winningSymbols := make(map[symbols.Symbol]bool)

	for _, win := range engineWins {
		// Get base symbol (removes _gold suffix if present)
		baseSym := symbols.GetBaseSymbol(string(win.Symbol))

		// Skip wild symbols
		if baseSym == symbols.SymbolWild {
			continue
		}

		winningSymbols[baseSym] = true
	}

	// Find the highest priority symbol that was won
	for _, prioritySym := range highValueSymbolPriority {
		if winningSymbols[prioritySym] {
			return string(prioritySym)
		}
	}

	// No high-value symbol found
	return ""
}

// convertCascadeWins converts engine wins to domain wins
func convertCascadeWins(engineWins []wins.CascadeWinDetail) []spin.CascadeWin {
	result := make([]spin.CascadeWin, len(engineWins))
	for i, win := range engineWins {
		// Convert positions from wins.Position to spin.Position
		positions := make([]spin.Position, len(win.Positions))
		for j, pos := range win.Positions {
			positions[j] = spin.Position{
				Reel:         pos.Reel,
				Row:          pos.Row,
				IsGoldToWild: pos.IsGoldToWild,
			}
		}

		result[i] = spin.CascadeWin{
			Symbol:    string(win.Symbol), // Convert symbols.Symbol to string
			Count:     win.Count,
			Ways:      win.Ways,
			Payout:    win.Payout,
			WinAmount: win.WinAmount,
			Positions: positions,
		}
	}
	return result
}
