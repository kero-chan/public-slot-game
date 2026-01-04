package freespins

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TRIGGER TESTS
// ============================================================================

func TestCheckTrigger(t *testing.T) {
	t.Run("should trigger with 3 scatters", func(t *testing.T) {
		// Create grid with 3 bonus symbols in visible rows (5-8)
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}

		result := CheckTrigger(grid)

		assert.True(t, result.Triggered)
		assert.Equal(t, 3, result.ScatterCount)
		assert.Equal(t, 12, result.SpinsAwarded) // 3 scatters = 12 spins
	})

	t.Run("should trigger with 4 scatters", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}

		result := CheckTrigger(grid)

		assert.True(t, result.Triggered)
		assert.Equal(t, 4, result.ScatterCount)
		assert.Equal(t, 14, result.SpinsAwarded) // 4 scatters = 14 spins
	})

	t.Run("should trigger with 5 scatters", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
		}

		result := CheckTrigger(grid)

		assert.True(t, result.Triggered)
		assert.Equal(t, 5, result.ScatterCount)
		assert.Equal(t, 16, result.SpinsAwarded) // 5 scatters = 16 spins
	})

	t.Run("should NOT trigger with 2 scatters", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}

		result := CheckTrigger(grid)

		assert.False(t, result.Triggered)
		assert.Equal(t, 2, result.ScatterCount)
		assert.Equal(t, 0, result.SpinsAwarded)
	})

	t.Run("should NOT trigger with 0 scatters", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}

		result := CheckTrigger(grid)

		assert.False(t, result.Triggered)
		assert.Equal(t, 0, result.ScatterCount)
		assert.Equal(t, 0, result.SpinsAwarded)
	})

	t.Run("should handle multiple scatters per reel", func(t *testing.T) {
		// Create grid with 2 bonus symbols per reel in visible rows = 10 scatters total
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "bonus", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "bonus", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "bonus", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "bonus", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "bonus", "bai", "wusuo", "wutong"},
		}

		result := CheckTrigger(grid)

		assert.True(t, result.Triggered)
		assert.Equal(t, 10, result.ScatterCount)
		// Formula: 12 + (2 × (10 - 3)) = 12 + 14 = 26
		assert.Equal(t, 26, result.SpinsAwarded)
	})

	t.Run("should only count scatters in visible rows", func(t *testing.T) {
		// Put bonus in buffer rows (0-4) and bottom row (9) - should not count
		grid := reels.Grid{
			{"bonus", "bonus", "bonus", "bonus", "bonus", "fa", "bai", "wusuo", "wutong", "bonus"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}

		result := CheckTrigger(grid)

		// Should NOT trigger because bonus symbols are outside visible rows
		assert.False(t, result.Triggered)
		assert.Equal(t, 0, result.ScatterCount)
	})
}

func TestGetScatterPositions(t *testing.T) {
	t.Run("should return correct positions for 3 scatters", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}

		positions := GetScatterPositions(grid)

		assert.Len(t, positions, 3)

		// Verify all positions are in visible rows (5-8)
		for _, pos := range positions {
			assert.GreaterOrEqual(t, pos.Row, 5)
			assert.LessOrEqual(t, pos.Row, 8)
		}

		// Verify positions are on different reels
		reels := make(map[int]bool)
		for _, pos := range positions {
			reels[pos.Reel] = true
		}
		assert.Len(t, reels, 3, "Scatters should be on 3 different reels")
	})

	t.Run("should return empty for no scatters", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}

		positions := GetScatterPositions(grid)

		assert.Len(t, positions, 0)
	})
}

func TestCalculateFreeSpinsAward(t *testing.T) {
	t.Run("should match symbols package calculation", func(t *testing.T) {
		for count := 0; count <= 10; count++ {
			expected := symbols.GetFreeSpinsAward(count)
			actual := CalculateFreeSpinsAward(count)
			assert.Equal(t, expected, actual,
				"CalculateFreeSpinsAward should match symbols.GetFreeSpinsAward for %d scatters", count)
		}
	})
}

// ============================================================================
// SESSION TESTS
// ============================================================================

func TestNewSession(t *testing.T) {
	playerID := uuid.New()
	scatterCount := 3
	betAmount := 20.0

	t.Run("should create session with correct initial values", func(t *testing.T) {
		session := NewSession(playerID, scatterCount, betAmount, nil)

		assert.NotEqual(t, uuid.Nil, session.ID)
		assert.Equal(t, playerID, session.PlayerID)
		assert.Equal(t, 12, session.TotalSpinsAwarded) // 3 scatters = 12 spins
		assert.Equal(t, 0, session.SpinsCompleted)
		assert.Equal(t, 12, session.RemainingSpins)
		assert.Equal(t, betAmount, session.LockedBetAmount)
		assert.Equal(t, 0.0, session.TotalWon)
		assert.True(t, session.IsActive)
		assert.False(t, session.CreatedAt.IsZero())
	})

	t.Run("should generate unique session IDs", func(t *testing.T) {
		session1 := NewSession(playerID, scatterCount, betAmount, nil)
		session2 := NewSession(playerID, scatterCount, betAmount, nil)

		assert.NotEqual(t, session1.ID, session2.ID)
	})

	t.Run("should set CreatedAt to current time", func(t *testing.T) {
		before := time.Now().UTC()
		session := NewSession(playerID, scatterCount, betAmount, nil)
		after := time.Now().UTC()

		assert.True(t, session.CreatedAt.After(before) || session.CreatedAt.Equal(before))
		assert.True(t, session.CreatedAt.Before(after) || session.CreatedAt.Equal(after))
	})

	t.Run("should calculate correct spins for different scatter counts", func(t *testing.T) {
		testCases := []struct {
			scatterCount  int
			expectedSpins int
		}{
			{3, 12},
			{4, 14},
			{5, 16},
			{6, 18},
		}

		for _, tc := range testCases {
			session := NewSession(playerID, tc.scatterCount, betAmount, nil)
			assert.Equal(t, tc.expectedSpins, session.TotalSpinsAwarded,
				"%d scatters should award %d spins", tc.scatterCount, tc.expectedSpins)
			assert.Equal(t, tc.expectedSpins, session.RemainingSpins)
		}
	})
}

func TestSession_ExecuteSpin(t *testing.T) {
	playerID := uuid.New()
	session := NewSession(playerID, 3, 20.0, nil)

	t.Run("should update session state correctly", func(t *testing.T) {
		initialRemaining := session.RemainingSpins
		winAmount := 50.0

		session.ExecuteSpin(winAmount)

		assert.Equal(t, 1, session.SpinsCompleted)
		assert.Equal(t, initialRemaining-1, session.RemainingSpins)
		assert.Equal(t, winAmount, session.TotalWon)
		assert.True(t, session.IsActive)
	})

	t.Run("should accumulate wins over multiple spins", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil)

		session.ExecuteSpin(10.0)
		session.ExecuteSpin(20.0)
		session.ExecuteSpin(30.0)

		assert.Equal(t, 3, session.SpinsCompleted)
		assert.Equal(t, 9, session.RemainingSpins) // 12 - 3 = 9
		assert.Equal(t, 60.0, session.TotalWon)
	})

	t.Run("should handle zero win", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil)

		session.ExecuteSpin(0.0)

		assert.Equal(t, 1, session.SpinsCompleted)
		assert.Equal(t, 11, session.RemainingSpins)
		assert.Equal(t, 0.0, session.TotalWon)
	})

	t.Run("should deactivate when all spins completed", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil) // 12 spins

		// Execute 11 spins - should still be active
		for i := 0; i < 11; i++ {
			session.ExecuteSpin(10.0)
		}
		assert.True(t, session.IsActive)
		assert.Equal(t, 1, session.RemainingSpins)

		// Execute last spin - should deactivate
		session.ExecuteSpin(10.0)
		assert.False(t, session.IsActive)
		assert.Equal(t, 0, session.RemainingSpins)
		assert.Equal(t, 12, session.SpinsCompleted)
	})
}

func TestSession_AddRetriggerSpins(t *testing.T) {
	playerID := uuid.New()

	t.Run("should add spins correctly", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil) // 12 spins

		session.AddRetriggerSpins(14) // Add 14 more (from 4 scatters)

		assert.Equal(t, 26, session.TotalSpinsAwarded) // 12 + 14
		assert.Equal(t, 26, session.RemainingSpins)
	})

	t.Run("should add to remaining spins after some completed", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil) // 12 spins

		// Complete 5 spins
		for i := 0; i < 5; i++ {
			session.ExecuteSpin(10.0)
		}
		assert.Equal(t, 7, session.RemainingSpins)

		// Retrigger with 3 scatters (12 more)
		session.AddRetriggerSpins(12)

		assert.Equal(t, 24, session.TotalSpinsAwarded) // 12 + 12
		assert.Equal(t, 19, session.RemainingSpins)    // 7 + 12
		assert.Equal(t, 5, session.SpinsCompleted)
	})

	t.Run("should handle multiple retriggers", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil) // 12 spins

		session.AddRetriggerSpins(12) // First retrigger
		session.AddRetriggerSpins(14) // Second retrigger

		assert.Equal(t, 38, session.TotalSpinsAwarded) // 12 + 12 + 14
		assert.Equal(t, 38, session.RemainingSpins)
	})
}

func TestSession_IsComplete(t *testing.T) {
	playerID := uuid.New()

	t.Run("should return false for active session with remaining spins", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil)
		assert.False(t, session.IsComplete())
	})

	t.Run("should return true when all spins completed", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil) // 12 spins

		for i := 0; i < 12; i++ {
			session.ExecuteSpin(10.0)
		}

		assert.True(t, session.IsComplete())
	})

	t.Run("should return true when deactivated", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil)
		session.IsActive = false

		assert.True(t, session.IsComplete())
	})

	t.Run("should return true when remaining spins is 0", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil)
		session.RemainingSpins = 0

		assert.True(t, session.IsComplete())
	})
}

func TestSession_GetProgress(t *testing.T) {
	playerID := uuid.New()

	t.Run("should return 0% at start", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil)
		assert.Equal(t, 0.0, session.GetProgress())
	})

	t.Run("should return correct percentage after some spins", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil) // 12 spins

		session.ExecuteSpin(10.0)
		session.ExecuteSpin(10.0)
		session.ExecuteSpin(10.0)

		// 3 / 12 = 25%
		expectedProgress := 25.0
		assert.InDelta(t, expectedProgress, session.GetProgress(), 0.01)
	})

	t.Run("should return 100% when complete", func(t *testing.T) {
		session := NewSession(playerID, 3, 20.0, nil) // 12 spins

		for i := 0; i < 12; i++ {
			session.ExecuteSpin(10.0)
		}

		assert.Equal(t, 100.0, session.GetProgress())
	})

	t.Run("should handle 0 total spins awarded", func(t *testing.T) {
		session := &Session{
			TotalSpinsAwarded: 0,
			SpinsCompleted:    0,
		}

		// Should return 100% to avoid division by zero
		assert.Equal(t, 100.0, session.GetProgress())
	})
}

// ============================================================================
// RETRIGGER TESTS
// ============================================================================

func TestCheckRetrigger(t *testing.T) {
	t.Run("should retrigger with 3 scatters", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}
		currentRemaining := 5

		result := CheckRetrigger(grid, currentRemaining)

		assert.True(t, result.Retriggered)
		assert.Equal(t, 3, result.ScatterCount)
		assert.Equal(t, 12, result.AdditionalSpins)
		assert.Equal(t, 17, result.NewTotalRemaining) // 5 + 12
	})

	t.Run("should retrigger with 4 scatters", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}
		currentRemaining := 3

		result := CheckRetrigger(grid, currentRemaining)

		assert.True(t, result.Retriggered)
		assert.Equal(t, 4, result.ScatterCount)
		assert.Equal(t, 14, result.AdditionalSpins)
		assert.Equal(t, 17, result.NewTotalRemaining) // 3 + 14
	})

	t.Run("should NOT retrigger with 2 scatters", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}
		currentRemaining := 5

		result := CheckRetrigger(grid, currentRemaining)

		assert.False(t, result.Retriggered)
		assert.Equal(t, 2, result.ScatterCount)
		assert.Equal(t, 0, result.AdditionalSpins)
		assert.Equal(t, 5, result.NewTotalRemaining) // Unchanged
	})

	t.Run("should work when 1 spin remaining", func(t *testing.T) {
		grid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}
		currentRemaining := 1

		result := CheckRetrigger(grid, currentRemaining)

		assert.True(t, result.Retriggered)
		assert.Equal(t, 13, result.NewTotalRemaining) // 1 + 12
	})
}

func TestIsRetriggerPossible(t *testing.T) {
	t.Run("should return true for 3+ scatters", func(t *testing.T) {
		assert.True(t, IsRetriggerPossible(3))
		assert.True(t, IsRetriggerPossible(4))
		assert.True(t, IsRetriggerPossible(5))
		assert.True(t, IsRetriggerPossible(10))
	})

	t.Run("should return false for less than 3 scatters", func(t *testing.T) {
		assert.False(t, IsRetriggerPossible(0))
		assert.False(t, IsRetriggerPossible(1))
		assert.False(t, IsRetriggerPossible(2))
	})
}

func TestGetRetriggerMessage(t *testing.T) {
	t.Run("should generate message", func(t *testing.T) {
		msg := GetRetriggerMessage(3, 12)
		assert.NotEmpty(t, msg)
		assert.Contains(t, msg, "Free Spins Retriggered")
	})
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

func TestFreeSpinsFullWorkflow(t *testing.T) {
	t.Run("should handle complete free spins session workflow", func(t *testing.T) {
		playerID := uuid.New()

		// 1. Initial trigger with 3 scatters
		triggerGrid := reels.Grid{
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
			{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		}

		trigger := CheckTrigger(triggerGrid)
		require.True(t, trigger.Triggered)
		require.Equal(t, 12, trigger.SpinsAwarded)

		// 2. Create session
		session := NewSession(playerID, trigger.ScatterCount, 20.0, nil)
		require.Equal(t, 12, session.RemainingSpins)

		// 3. Execute 5 spins
		for i := 0; i < 5; i++ {
			session.ExecuteSpin(50.0)
		}
		assert.Equal(t, 5, session.SpinsCompleted)
		assert.Equal(t, 7, session.RemainingSpins)
		assert.Equal(t, 250.0, session.TotalWon)

		// 4. Retrigger on 6th spin with 4 scatters
		retriggerGrid := triggerGrid // Use same grid (has 3 scatters) + 1 more
		retriggerGrid[3] = []string{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"}

		retrigger := CheckRetrigger(retriggerGrid, session.RemainingSpins)
		assert.True(t, retrigger.Retriggered)
		assert.Equal(t, 4, retrigger.ScatterCount)
		assert.Equal(t, 14, retrigger.AdditionalSpins)

		session.AddRetriggerSpins(retrigger.AdditionalSpins)
		assert.Equal(t, 26, session.TotalSpinsAwarded) // 12 + 14
		assert.Equal(t, 21, session.RemainingSpins)    // 7 + 14

		// 5. Complete all remaining spins
		for session.RemainingSpins > 0 {
			session.ExecuteSpin(30.0)
		}

		assert.True(t, session.IsComplete())
		assert.Equal(t, 26, session.SpinsCompleted)
		assert.Equal(t, 880.0, session.TotalWon) // 5*50 + 21*30 = 250 + 630
		assert.Equal(t, 100.0, session.GetProgress())
	})
}

func BenchmarkCheckTrigger(b *testing.B) {
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckTrigger(grid)
	}
}

func BenchmarkNewSession(b *testing.B) {
	playerID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSession(playerID, 3, 20.0, nil)
	}
}

func BenchmarkCheckRetrigger(b *testing.B) {
	grid := reels.Grid{
		{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "bonus", "fa", "bai", "wusuo", "wutong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
		{"cai", "fu", "shu", "zhong", "liangtong", "fa", "bai", "wusuo", "wutong", "zhong"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckRetrigger(grid, 5)
	}
}
