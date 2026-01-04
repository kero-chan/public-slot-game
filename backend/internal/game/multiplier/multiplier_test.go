package multiplier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMultiplier(t *testing.T) {
	t.Run("should return correct base game multipliers", func(t *testing.T) {
		testCases := []struct {
			cascade  int
			expected int
		}{
			{1, 1}, // First cascade: 1x
			{2, 2}, // Second cascade: 2x
			{3, 3}, // Third cascade: 3x
			{4, 5}, // Fourth cascade: 5x
			{5, 5}, // Fifth+ cascade: 5x (capped)
			{10, 5}, // Any cascade 4+: 5x
		}

		for _, tc := range testCases {
			result := GetMultiplier(tc.cascade, false)
			assert.Equal(t, tc.expected, result,
				"Base game cascade %d should have multiplier %dx", tc.cascade, tc.expected)
		}
	})

	t.Run("should return correct free spins multipliers", func(t *testing.T) {
		testCases := []struct {
			cascade  int
			expected int
		}{
			{1, 2},  // First cascade: 2x
			{2, 4},  // Second cascade: 4x
			{3, 6},  // Third cascade: 6x
			{4, 10}, // Fourth cascade: 10x
			{5, 10}, // Fifth+ cascade: 10x (capped)
			{10, 10}, // Any cascade 4+: 10x
		}

		for _, tc := range testCases {
			result := GetMultiplier(tc.cascade, true)
			assert.Equal(t, tc.expected, result,
				"Free spins cascade %d should have multiplier %dx", tc.cascade, tc.expected)
		}
	})

	t.Run("should handle cascade number < 1", func(t *testing.T) {
		// Cascade < 1 should be treated as cascade 1
		assert.Equal(t, 1, GetMultiplier(0, false))
		assert.Equal(t, 1, GetMultiplier(-1, false))
		assert.Equal(t, 1, GetMultiplier(-10, false))

		assert.Equal(t, 2, GetMultiplier(0, true))
		assert.Equal(t, 2, GetMultiplier(-1, true))
		assert.Equal(t, 2, GetMultiplier(-10, true))
	})

	t.Run("should cap at maximum multiplier", func(t *testing.T) {
		// Base game caps at 5x
		for cascade := 4; cascade <= 100; cascade++ {
			result := GetMultiplier(cascade, false)
			assert.Equal(t, 5, result,
				"Base game cascade %d should be capped at 5x", cascade)
		}

		// Free spins cap at 10x
		for cascade := 4; cascade <= 100; cascade++ {
			result := GetMultiplier(cascade, true)
			assert.Equal(t, 10, result,
				"Free spins cascade %d should be capped at 10x", cascade)
		}
	})

	t.Run("should validate multiplier progression", func(t *testing.T) {
		// Base game: 1x -> 2x -> 3x -> 5x (increases but not uniformly)
		mult1 := GetMultiplier(1, false)
		mult2 := GetMultiplier(2, false)
		mult3 := GetMultiplier(3, false)
		mult4 := GetMultiplier(4, false)

		assert.Less(t, mult1, mult2, "Multiplier should increase")
		assert.Less(t, mult2, mult3, "Multiplier should increase")
		assert.Less(t, mult3, mult4, "Multiplier should increase")

		// Free spins: 2x -> 4x -> 6x -> 10x (increases but not uniformly)
		fsMult1 := GetMultiplier(1, true)
		fsMult2 := GetMultiplier(2, true)
		fsMult3 := GetMultiplier(3, true)
		fsMult4 := GetMultiplier(4, true)

		assert.Less(t, fsMult1, fsMult2, "Free spins multiplier should increase")
		assert.Less(t, fsMult2, fsMult3, "Free spins multiplier should increase")
		assert.Less(t, fsMult3, fsMult4, "Free spins multiplier should increase")
	})

	t.Run("should have higher multipliers in free spins", func(t *testing.T) {
		// At each cascade level, free spins should have higher multiplier
		for cascade := 1; cascade <= 4; cascade++ {
			baseMultiplier := GetMultiplier(cascade, false)
			freeSpinsMultiplier := GetMultiplier(cascade, true)

			assert.Greater(t, freeSpinsMultiplier, baseMultiplier,
				"Free spins cascade %d multiplier should be higher than base game", cascade)
		}
	})
}

func TestGetBaseMultiplier(t *testing.T) {
	t.Run("should return 1x for base game", func(t *testing.T) {
		result := GetBaseMultiplier(false)
		assert.Equal(t, 1, result, "Base game should start at 1x")
	})

	t.Run("should return 2x for free spins", func(t *testing.T) {
		result := GetBaseMultiplier(true)
		assert.Equal(t, 2, result, "Free spins should start at 2x")
	})

	t.Run("should match first cascade multiplier", func(t *testing.T) {
		// Base multiplier should equal first cascade multiplier
		baseGameStart := GetBaseMultiplier(false)
		baseGameFirst := GetMultiplier(1, false)
		assert.Equal(t, baseGameStart, baseGameFirst,
			"Base multiplier should match first cascade multiplier")

		freeSpinsStart := GetBaseMultiplier(true)
		freeSpinsFirst := GetMultiplier(1, true)
		assert.Equal(t, freeSpinsStart, freeSpinsFirst,
			"Free spins base multiplier should match first cascade multiplier")
	})
}

func TestGetIncrement(t *testing.T) {
	t.Run("should return 1 for base game", func(t *testing.T) {
		result := GetIncrement(false)
		assert.Equal(t, 1, result, "Base game should increment by 1")
	})

	t.Run("should return 2 for free spins", func(t *testing.T) {
		result := GetIncrement(true)
		assert.Equal(t, 2, result, "Free spins should increment by 2")
	})

	t.Run("should reflect actual multiplier differences", func(t *testing.T) {
		// Base game: 1 -> 2 -> 3 (increments by 1 each time, until cap)
		mult1 := GetMultiplier(1, false)
		mult2 := GetMultiplier(2, false)
		mult3 := GetMultiplier(3, false)

		assert.Equal(t, 1, mult2-mult1, "Base game cascade 1->2 should increment by 1")
		assert.Equal(t, 1, mult3-mult2, "Base game cascade 2->3 should increment by 1")

		// Free spins: 2 -> 4 -> 6 (increments by 2 each time, until cap)
		fsMult1 := GetMultiplier(1, true)
		fsMult2 := GetMultiplier(2, true)
		fsMult3 := GetMultiplier(3, true)

		assert.Equal(t, 2, fsMult2-fsMult1, "Free spins cascade 1->2 should increment by 2")
		assert.Equal(t, 2, fsMult3-fsMult2, "Free spins cascade 2->3 should increment by 2")
	})
}

func TestCalculateMultiplierProgression(t *testing.T) {
	t.Run("should return correct progression for base game", func(t *testing.T) {
		progression := CalculateMultiplierProgression(5, false)

		expectedProgression := []int{1, 2, 3, 5, 5}
		assert.Equal(t, expectedProgression, progression,
			"Base game progression should be [1, 2, 3, 5, 5]")
	})

	t.Run("should return correct progression for free spins", func(t *testing.T) {
		progression := CalculateMultiplierProgression(5, true)

		expectedProgression := []int{2, 4, 6, 10, 10}
		assert.Equal(t, expectedProgression, progression,
			"Free spins progression should be [2, 4, 6, 10, 10]")
	})

	t.Run("should handle single cascade", func(t *testing.T) {
		baseProgression := CalculateMultiplierProgression(1, false)
		assert.Equal(t, []int{1}, baseProgression)

		freeSpinsProgression := CalculateMultiplierProgression(1, true)
		assert.Equal(t, []int{2}, freeSpinsProgression)
	})

	t.Run("should handle zero cascades", func(t *testing.T) {
		baseProgression := CalculateMultiplierProgression(0, false)
		assert.Equal(t, []int{}, baseProgression)

		freeSpinsProgression := CalculateMultiplierProgression(0, true)
		assert.Equal(t, []int{}, freeSpinsProgression)
	})

	t.Run("should handle many cascades", func(t *testing.T) {
		progression := CalculateMultiplierProgression(10, false)

		assert.Len(t, progression, 10)
		// First 3 should be [1, 2, 3]
		assert.Equal(t, 1, progression[0])
		assert.Equal(t, 2, progression[1])
		assert.Equal(t, 3, progression[2])
		// Rest should be 5 (capped)
		for i := 3; i < 10; i++ {
			assert.Equal(t, 5, progression[i],
				"Base game cascade %d should be capped at 5x", i+1)
		}
	})

	t.Run("should validate progression length", func(t *testing.T) {
		for numCascades := 1; numCascades <= 20; numCascades++ {
			baseProgression := CalculateMultiplierProgression(numCascades, false)
			freeSpinsProgression := CalculateMultiplierProgression(numCascades, true)

			assert.Len(t, baseProgression, numCascades,
				"Progression should have length equal to number of cascades")
			assert.Len(t, freeSpinsProgression, numCascades,
				"Free spins progression should have length equal to number of cascades")
		}
	})

	t.Run("should never decrease", func(t *testing.T) {
		// Base game progression should never decrease
		baseProgression := CalculateMultiplierProgression(10, false)
		for i := 1; i < len(baseProgression); i++ {
			assert.GreaterOrEqual(t, baseProgression[i], baseProgression[i-1],
				"Base game multiplier should never decrease at cascade %d", i+1)
		}

		// Free spins progression should never decrease
		freeSpinsProgression := CalculateMultiplierProgression(10, true)
		for i := 1; i < len(freeSpinsProgression); i++ {
			assert.GreaterOrEqual(t, freeSpinsProgression[i], freeSpinsProgression[i-1],
				"Free spins multiplier should never decrease at cascade %d", i+1)
		}
	})
}

func TestMultiplierConsistency(t *testing.T) {
	t.Run("should validate multiplier progression follows spec", func(t *testing.T) {
		// Spec from 03-game-mechanics.md:
		// Base game: 1x, 2x, 3x, 5x, 5x, ...
		// Free spins: 2x, 4x, 6x, 10x, 10x, ...

		// Base game validation
		assert.Equal(t, 1, GetMultiplier(1, false), "Base game cascade 1 should be 1x")
		assert.Equal(t, 2, GetMultiplier(2, false), "Base game cascade 2 should be 2x")
		assert.Equal(t, 3, GetMultiplier(3, false), "Base game cascade 3 should be 3x")
		assert.Equal(t, 5, GetMultiplier(4, false), "Base game cascade 4+ should be 5x")

		// Free spins validation
		assert.Equal(t, 2, GetMultiplier(1, true), "Free spins cascade 1 should be 2x")
		assert.Equal(t, 4, GetMultiplier(2, true), "Free spins cascade 2 should be 4x")
		assert.Equal(t, 6, GetMultiplier(3, true), "Free spins cascade 3 should be 6x")
		assert.Equal(t, 10, GetMultiplier(4, true), "Free spins cascade 4+ should be 10x")
	})

	t.Run("should validate free spins multiplier is exactly 2x base game", func(t *testing.T) {
		// For cascades 1-3, free spins multiplier should be exactly 2x base game
		for cascade := 1; cascade <= 3; cascade++ {
			baseMultiplier := GetMultiplier(cascade, false)
			freeSpinsMultiplier := GetMultiplier(cascade, true)

			assert.Equal(t, baseMultiplier*2, freeSpinsMultiplier,
				"Free spins cascade %d multiplier should be exactly 2x base game", cascade)
		}
	})

	t.Run("should validate all functions agree", func(t *testing.T) {
		// GetBaseMultiplier, GetIncrement, and GetMultiplier should be consistent

		// Base game
		baseStart := GetBaseMultiplier(false)
		baseInc := GetIncrement(false)
		assert.Equal(t, baseStart, GetMultiplier(1, false))
		assert.Equal(t, baseStart+baseInc, GetMultiplier(2, false))
		assert.Equal(t, baseStart+2*baseInc, GetMultiplier(3, false))

		// Free spins
		fsStart := GetBaseMultiplier(true)
		fsInc := GetIncrement(true)
		assert.Equal(t, fsStart, GetMultiplier(1, true))
		assert.Equal(t, fsStart+fsInc, GetMultiplier(2, true))
		assert.Equal(t, fsStart+2*fsInc, GetMultiplier(3, true))
	})
}

func BenchmarkGetMultiplier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetMultiplier(3, false)
	}
}

func BenchmarkGetBaseMultiplier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetBaseMultiplier(false)
	}
}

func BenchmarkCalculateMultiplierProgression(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = CalculateMultiplierProgression(10, false)
	}
}
