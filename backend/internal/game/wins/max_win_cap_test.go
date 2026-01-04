package wins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyMaxWinCap(t *testing.T) {
	t.Run("should allow wins under cap", func(t *testing.T) {
		winAmount := 1000.0
		betAmount := 10.0

		result := ApplyMaxWinCap(winAmount, betAmount)

		assert.Equal(t, winAmount, result, "Wins under cap should not be modified")
	})

	t.Run("should cap wins at 25,000x bet", func(t *testing.T) {
		winAmount := 500000.0 // Exceeds cap
		betAmount := 10.0

		result := ApplyMaxWinCap(winAmount, betAmount)

		expectedMax := 10.0 * 25000 // 250,000
		assert.Equal(t, expectedMax, result, "Wins should be capped at 25,000x bet")
	})

	t.Run("should handle exactly at cap", func(t *testing.T) {
		betAmount := 10.0
		winAmount := betAmount * 25000 // Exactly at cap

		result := ApplyMaxWinCap(winAmount, betAmount)

		assert.Equal(t, winAmount, result, "Wins exactly at cap should not be modified")
	})

	t.Run("should handle zero bet", func(t *testing.T) {
		winAmount := 1000.0
		betAmount := 0.0

		result := ApplyMaxWinCap(winAmount, betAmount)

		assert.Equal(t, 0.0, result, "Zero bet should result in zero max win")
	})

	t.Run("should handle zero win", func(t *testing.T) {
		winAmount := 0.0
		betAmount := 10.0

		result := ApplyMaxWinCap(winAmount, betAmount)

		assert.Equal(t, 0.0, result, "Zero win should remain zero")
	})

	t.Run("should handle small bets", func(t *testing.T) {
		winAmount := 100.0
		betAmount := 0.01

		result := ApplyMaxWinCap(winAmount, betAmount)

		// Win (100) is less than cap (250 = 0.01 * 25000), so should not be modified
		assert.Equal(t, winAmount, result)
	})

	t.Run("should handle large bets", func(t *testing.T) {
		winAmount := 30000000.0 // 30 million
		betAmount := 1000.0

		result := ApplyMaxWinCap(winAmount, betAmount)

		expectedMax := 1000.0 * 25000 // 25 million
		assert.Equal(t, expectedMax, result, "Large wins should be capped correctly")
	})

	t.Run("should handle fractional bets", func(t *testing.T) {
		winAmount := 500.0
		betAmount := 0.25

		result := ApplyMaxWinCap(winAmount, betAmount)

		// Win (500) is less than cap (6,250 = 0.25 * 25000), so should not be modified
		assert.Equal(t, winAmount, result)
	})

	t.Run("should cap fractional wins", func(t *testing.T) {
		winAmount := 1000.5
		betAmount := 0.02

		result := ApplyMaxWinCap(winAmount, betAmount)

		expectedMax := 0.02 * 25000 // 500
		assert.Equal(t, expectedMax, result)
	})
}

func TestIsWinCapped(t *testing.T) {
	t.Run("should return false for wins under cap", func(t *testing.T) {
		winAmount := 1000.0
		betAmount := 10.0

		isCapped := IsWinCapped(winAmount, betAmount)

		assert.False(t, isCapped)
	})

	t.Run("should return true for wins over cap", func(t *testing.T) {
		winAmount := 500000.0
		betAmount := 10.0

		isCapped := IsWinCapped(winAmount, betAmount)

		assert.True(t, isCapped, "Wins exceeding 25,000x bet should be flagged as capped")
	})

	t.Run("should return false for wins exactly at cap", func(t *testing.T) {
		betAmount := 10.0
		winAmount := betAmount * 25000 // Exactly 25,000x

		isCapped := IsWinCapped(winAmount, betAmount)

		assert.False(t, isCapped, "Wins exactly at cap should not be flagged as capped")
	})

	t.Run("should return false for zero win", func(t *testing.T) {
		winAmount := 0.0
		betAmount := 10.0

		isCapped := IsWinCapped(winAmount, betAmount)

		assert.False(t, isCapped)
	})

	t.Run("should handle edge case just over cap", func(t *testing.T) {
		betAmount := 10.0
		winAmount := (betAmount * 25000) + 0.01 // Just over cap

		isCapped := IsWinCapped(winAmount, betAmount)

		assert.True(t, isCapped)
	})
}

func TestGetMaxWinForBet(t *testing.T) {
	t.Run("should calculate max win for standard bet", func(t *testing.T) {
		betAmount := 10.0

		maxWin := GetMaxWinForBet(betAmount)

		expectedMax := 10.0 * 25000 // 250,000
		assert.Equal(t, expectedMax, maxWin)
	})

	t.Run("should calculate max win for small bet", func(t *testing.T) {
		betAmount := 0.01

		maxWin := GetMaxWinForBet(betAmount)

		expectedMax := 0.01 * 25000 // 250
		assert.Equal(t, expectedMax, maxWin)
	})

	t.Run("should calculate max win for large bet", func(t *testing.T) {
		betAmount := 1000.0

		maxWin := GetMaxWinForBet(betAmount)

		expectedMax := 1000.0 * 25000 // 25,000,000
		assert.Equal(t, expectedMax, maxWin)
	})

	t.Run("should return zero for zero bet", func(t *testing.T) {
		betAmount := 0.0

		maxWin := GetMaxWinForBet(betAmount)

		assert.Equal(t, 0.0, maxWin)
	})

	t.Run("should handle fractional bets", func(t *testing.T) {
		betAmount := 2.5

		maxWin := GetMaxWinForBet(betAmount)

		expectedMax := 2.5 * 25000 // 62,500
		assert.Equal(t, expectedMax, maxWin)
	})

	t.Run("should validate max win multiplier constant", func(t *testing.T) {
		// Verify the constant is correct
		assert.Equal(t, 25000, MaxWinMultiplier, "MaxWinMultiplier should be 25,000 per regulatory requirements")
	})
}

func TestMaxWinCap_RegulatoryCompliance(t *testing.T) {
	t.Run("should enforce regulatory 25000x limit", func(t *testing.T) {
		// Test multiple bet amounts to ensure compliance
		testCases := []struct {
			betAmount        float64
			winAmount        float64
			expectedCappedWin float64
		}{
			{1.0, 30000.0, 25000.0},
			{5.0, 150000.0, 125000.0},
			{10.0, 500000.0, 250000.0},
			{20.0, 1000000.0, 500000.0},
			{100.0, 5000000.0, 2500000.0},
		}

		for _, tc := range testCases {
			result := ApplyMaxWinCap(tc.winAmount, tc.betAmount)
			assert.Equal(t, tc.expectedCappedWin, result,
				"Bet %.2f with win %.2f should be capped to %.2f",
				tc.betAmount, tc.winAmount, tc.expectedCappedWin)
		}
	})

	t.Run("should log when cap is applied", func(t *testing.T) {
		// This test documents the expected behavior
		// In production, ApplyMaxWinCap should log when it caps a win
		winAmount := 500000.0
		betAmount := 10.0

		beforeCap := winAmount
		afterCap := ApplyMaxWinCap(winAmount, betAmount)

		assert.NotEqual(t, beforeCap, afterCap, "Win was capped")
		assert.True(t, IsWinCapped(winAmount, betAmount), "Win should be flagged as capped")

		// TODO: Verify logging when logging is implemented
	})
}

func TestMaxWinCap_EdgeCases(t *testing.T) {
	t.Run("should handle negative bet amounts", func(t *testing.T) {
		// Edge case: negative bet (should not occur in practice)
		winAmount := 1000.0
		betAmount := -10.0

		result := ApplyMaxWinCap(winAmount, betAmount)

		// Negative bet * 25000 = negative max win
		// Win (1000) > negative max, so it will be "capped" to the negative value
		expectedMax := betAmount * 25000
		assert.Equal(t, expectedMax, result)
	})

	t.Run("should handle very large win amounts", func(t *testing.T) {
		winAmount := 999999999999.0 // Nearly 1 trillion
		betAmount := 100.0

		result := ApplyMaxWinCap(winAmount, betAmount)

		expectedMax := 100.0 * 25000 // 2.5 million
		assert.Equal(t, expectedMax, result)
	})

	t.Run("should handle very small bet amounts", func(t *testing.T) {
		winAmount := 100.0
		betAmount := 0.000001 // 1 micro-unit

		result := ApplyMaxWinCap(winAmount, betAmount)

		expectedMax := 0.000001 * 25000 // 0.025
		assert.InDelta(t, expectedMax, result, 0.0001, "Should handle very small bet amounts with floating point precision")
	})
}

func BenchmarkApplyMaxWinCap(b *testing.B) {
	winAmount := 500000.0
	betAmount := 10.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyMaxWinCap(winAmount, betAmount)
	}
}

func BenchmarkIsWinCapped(b *testing.B) {
	winAmount := 500000.0
	betAmount := 10.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsWinCapped(winAmount, betAmount)
	}
}

func BenchmarkGetMaxWinForBet(b *testing.B) {
	betAmount := 10.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetMaxWinForBet(betAmount)
	}
}
