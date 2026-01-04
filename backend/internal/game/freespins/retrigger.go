package freespins

import (
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/symbols"
)

// RetriggerResult represents the result of checking for retrigger
type RetriggerResult struct {
	Retriggered       bool `json:"retriggered"`
	ScatterCount      int  `json:"scatter_count"`
	AdditionalSpins   int  `json:"additional_spins"`
	NewTotalRemaining int  `json:"new_total_remaining"`
}

// CheckRetrigger checks if free spins are retriggered during a free spin
// Retrigger adds additional spins to the current session
func CheckRetrigger(grid reels.Grid, currentRemainingSpins int) RetriggerResult {
	scatterCount := countScatters(grid)

	if scatterCount >= symbols.MinScattersForFreeSpin() {
		additionalSpins := symbols.GetFreeSpinsAward(scatterCount)
		newTotalRemaining := currentRemainingSpins + additionalSpins

		return RetriggerResult{
			Retriggered:       true,
			ScatterCount:      scatterCount,
			AdditionalSpins:   additionalSpins,
			NewTotalRemaining: newTotalRemaining,
		}
	}

	return RetriggerResult{
		Retriggered:       false,
		ScatterCount:      scatterCount,
		AdditionalSpins:   0,
		NewTotalRemaining: currentRemainingSpins,
	}
}

// IsRetriggerPossible checks if retrigger is possible based on scatter count
func IsRetriggerPossible(scatterCount int) bool {
	return scatterCount >= symbols.MinScattersForFreeSpin()
}

// GetRetriggerMessage generates a message for retrigger event
func GetRetriggerMessage(scatterCount, additionalSpins int) string {
	return "Free Spins Retriggered! +" + string(rune(additionalSpins)) + " spins awarded!"
}
