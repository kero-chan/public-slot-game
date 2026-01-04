package reels

import (
	"fmt"

	"github.com/slotmachine/backend/internal/game/rng"
	"github.com/slotmachine/backend/internal/game/symbols"
)

const (
	ReelCount   = 5  // Number of reels
	StartRow    = 4  // First visible row (after buffer rows)
	VisibleRows = 6  // Visible rows in the grid (including partial top/bottom)
	BufferRows  = 4  // Buffer rows above visible area
	TotalRows   = 10 // Total rows (4 buffer + 6 visible)

	// Winning check range - only the middle 4 fully visible rows
	// Row 4: top partial (visible but not checked for wins)
	// Row 5-8: fully visible middle 4 rows (checked for wins)
	// Row 9: bottom partial (visible but not checked for wins)
	WinCheckStartRow = 5 // First fully visible row for win checks
	WinCheckEndRow   = 8 // Last fully visible row for win checks (inclusive)
)

// ReelStrip represents a single reel strip
type ReelStrip []string

// GenerateReelStrip generates a reel strip from symbol weights
// using cryptographically secure RNG
func GenerateReelStrip(weights symbols.ReelWeights, rng *rng.CryptoRNG) (ReelStrip, error) {
	stripLength := 0
	for _, weight := range weights {
		stripLength += weight
	}
	strip := make(ReelStrip, 0, stripLength)

	// Create a pool of symbols based on weights
	symbolPool := make([]string, 0, stripLength)
	for symbol, weight := range weights {
		for i := 0; i < weight; i++ {
			symbolPool = append(symbolPool, symbol)
		}
	}

	if len(symbolPool) != stripLength {
		fmt.Println("weight", weights)
		return nil, fmt.Errorf("symbol pool size %d does not match strip length %d", len(symbolPool), stripLength)
	}

	// Shuffle the pool using Fisher-Yates with crypto RNG
	err := rng.Shuffle(len(symbolPool), func(i, j int) {
		symbolPool[i], symbolPool[j] = symbolPool[j], symbolPool[i]
	})
	if err != nil {
		return nil, fmt.Errorf("failed to shuffle reel strip: %w", err)
	}

	strip = symbolPool
	return strip, nil
}

// GenerateAllReelStrips generates all 5 reel strips for base game or free spins
func GenerateAllReelStrips(isFreeSpin bool, rng *rng.CryptoRNG) ([]ReelStrip, error) {
	strips := make([]ReelStrip, ReelCount)

	for i := 0; i < ReelCount; i++ {
		var weights symbols.ReelWeights
		if isFreeSpin {
			weights = symbols.GetFreeSpinsWeights(i)
		} else {
			weights = symbols.GetBaseGameWeights(i)
		}

		strip, err := GenerateReelStrip(weights, rng)
		if err != nil {
			return nil, fmt.Errorf("failed to generate reel %d: %w", i+1, err)
		}
		strips[i] = strip
	}

	return strips, nil
}

// GenerateTrialReelStrips generates all 5 reel strips with HUGE RTP for trial mode
// Uses trial-specific weights with higher bonus/wild rates for better winning experience
func GenerateTrialReelStrips(isFreeSpin bool, rng *rng.CryptoRNG) ([]ReelStrip, error) {
	strips := make([]ReelStrip, ReelCount)

	for i := 0; i < ReelCount; i++ {
		var weights symbols.ReelWeights
		if isFreeSpin {
			weights = symbols.GetTrialFreeSpinsWeights(i)
		} else {
			weights = symbols.GetTrialWeights(i)
		}

		strip, err := GenerateReelStrip(weights, rng)
		if err != nil {
			return nil, fmt.Errorf("failed to generate trial reel %d: %w", i+1, err)
		}
		strips[i] = strip
	}

	return strips, nil
}

// GetSymbolAtPosition returns the symbol at a specific position on a reel strip
func (rs ReelStrip) GetSymbolAtPosition(position int) string {
	if len(rs) == 0 {
		return ""
	}
	// Wrap around if position exceeds strip length
	position = position % len(rs)
	if position < 0 {
		position += len(rs)
	}
	return rs[position]
}

// GetSymbolsFromPosition returns N symbols starting from a position
func (rs ReelStrip) GetSymbolsFromPosition(startPos, count int) []string {
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = rs.GetSymbolAtPosition(startPos + i)
	}
	return result
}
