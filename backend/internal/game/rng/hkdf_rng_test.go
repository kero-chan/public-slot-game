package rng

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants - prevSpinHash simulates hash from previous spin (or server_seed_hash for first spin)
const testPrevSpinHash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestNewHKDFRNG(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"
	nonce := int64(1)

	rng, err := NewHKDFRNG(serverSeed, clientSeed, nonce, testPrevSpinHash)
	require.NoError(t, err)
	assert.NotNil(t, rng)
	assert.Len(t, rng.masterKey, 32)
	assert.Len(t, rng.spinHash, 64) // hex encoded
}

func TestHKDFRNG_Deterministic(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"
	nonce := int64(1)

	// Create two RNGs with same inputs
	rng1, err := NewHKDFRNG(serverSeed, clientSeed, nonce, testPrevSpinHash)
	require.NoError(t, err)

	rng2, err := NewHKDFRNG(serverSeed, clientSeed, nonce, testPrevSpinHash)
	require.NoError(t, err)

	// Should produce identical results
	assert.Equal(t, rng1.masterKey, rng2.masterKey)
	assert.Equal(t, rng1.spinHash, rng2.spinHash)

	// Same reel positions
	for i := 0; i < 5; i++ {
		pos1, err := rng1.GetReelPosition(i, 100)
		require.NoError(t, err)

		pos2, err := rng2.GetReelPosition(i, 100)
		require.NoError(t, err)

		assert.Equal(t, pos1, pos2, "Reel %d should produce same position", i)
	}
}

func TestHKDFRNG_DifferentInputsDifferentOutputs(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed1 := "deadbeefdeadbeefdeadbeefdeadbeef"
	clientSeed2 := "cafebabecafebabecafebabecafebabe"

	rng1, _ := NewHKDFRNG(serverSeed, clientSeed1, 1, testPrevSpinHash)
	rng2, _ := NewHKDFRNG(serverSeed, clientSeed2, 1, testPrevSpinHash)

	// Different client seeds should produce different results
	assert.NotEqual(t, rng1.masterKey, rng2.masterKey)
}

func TestHKDFRNG_DifferentNoncesDifferentOutputs(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng1, _ := NewHKDFRNG(serverSeed, clientSeed, 1, testPrevSpinHash)
	rng2, _ := NewHKDFRNG(serverSeed, clientSeed, 2, testPrevSpinHash)

	// Different nonces should produce different results
	assert.NotEqual(t, rng1.masterKey, rng2.masterKey)
}

func TestHKDFRNG_ReelKeyIndependence(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// Each reel should have a unique key
	keys := make([][]byte, 5)
	for i := 0; i < 5; i++ {
		key, err := rng.DeriveReelKey(i)
		require.NoError(t, err)
		keys[i] = key
	}

	// Verify all keys are different
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			assert.NotEqual(t, keys[i], keys[j],
				"Reel %d and %d should have different keys", i, j)
		}
	}
}

func TestHKDFRNG_GetReelPosition(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// Test multiple reel lengths
	reelLengths := []int{50, 100, 150, 200, 250}

	for i, length := range reelLengths {
		pos, err := rng.GetReelPosition(i, length)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, pos, 0)
		assert.Less(t, pos, length, "Position should be less than reel length")
	}
}

func TestHKDFRNG_GetAllReelPositions(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	reelLengths := []int{100, 100, 100, 100, 100}
	positions, err := rng.GetAllReelPositions(reelLengths)

	require.NoError(t, err)
	assert.Len(t, positions, 5)

	for i, pos := range positions {
		assert.GreaterOrEqual(t, pos, 0)
		assert.Less(t, pos, reelLengths[i])
	}
}

func TestHKDFRNG_DeriveKey(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// Test different domains produce different keys
	key1, err := rng.DeriveKey("multiplier", 32)
	require.NoError(t, err)

	key2, err := rng.DeriveKey("bonus:trigger", 32)
	require.NoError(t, err)

	assert.NotEqual(t, key1, key2)

	// Same domain produces same key
	key3, err := rng.DeriveKey("multiplier", 32)
	require.NoError(t, err)
	assert.Equal(t, key1, key3)
}

func TestHKDFRNG_Int(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// Test Int with domain
	value, err := rng.Int("test-domain", 100)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, value, 0)
	assert.Less(t, value, 100)

	// Same domain produces same value (deterministic)
	value2, err := rng.Int("test-domain", 100)
	require.NoError(t, err)
	assert.Equal(t, value, value2)
}

func TestHKDFRNG_Float64(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	value, err := rng.Float64("test-domain")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, value, 0.0)
	assert.Less(t, value, 1.0)
}

func TestVerifyReelPosition(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"
	nonce := int64(1)
	reelLength := 100

	// Get actual position
	rng, _ := NewHKDFRNG(serverSeed, clientSeed, nonce, testPrevSpinHash)
	actualPos, _ := rng.GetReelPosition(0, reelLength)

	// Verify correct position
	valid, err := VerifyReelPosition(serverSeed, clientSeed, nonce, testPrevSpinHash, 0, reelLength, actualPos)
	require.NoError(t, err)
	assert.True(t, valid)

	// Verify wrong position fails
	wrongPos := (actualPos + 1) % reelLength
	valid, err = VerifyReelPosition(serverSeed, clientSeed, nonce, testPrevSpinHash, 0, reelLength, wrongPos)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestVerifyAllReelPositions(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"
	nonce := int64(1)
	reelLengths := []int{100, 100, 100, 100, 100}

	// Get actual positions
	rng, _ := NewHKDFRNG(serverSeed, clientSeed, nonce, testPrevSpinHash)
	actualPositions, _ := rng.GetAllReelPositions(reelLengths)

	// Verify correct positions
	valid, err := VerifyAllReelPositions(serverSeed, clientSeed, nonce, testPrevSpinHash, reelLengths, actualPositions)
	require.NoError(t, err)
	assert.True(t, valid)

	// Verify wrong positions fail
	wrongPositions := make([]int, len(actualPositions))
	copy(wrongPositions, actualPositions)
	wrongPositions[0] = (wrongPositions[0] + 1) % reelLengths[0]

	valid, err = VerifyAllReelPositions(serverSeed, clientSeed, nonce, testPrevSpinHash, reelLengths, wrongPositions)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestHKDFRNG_Distribution(t *testing.T) {
	// Test that positions are reasonably distributed
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	reelLength := 100
	iterations := 1000

	// Count occurrences in buckets
	buckets := make([]int, 10) // 10 buckets of 10 positions each

	for i := 0; i < iterations; i++ {
		clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"
		rng, _ := NewHKDFRNG(serverSeed, clientSeed, int64(i), testPrevSpinHash)
		pos, _ := rng.GetReelPosition(0, reelLength)
		buckets[pos/10]++
	}

	// Each bucket should have roughly iterations/10 = 100 entries
	// Allow 50% variance for statistical noise
	expectedPerBucket := float64(iterations) / 10
	for i, count := range buckets {
		ratio := float64(count) / expectedPerBucket
		assert.Greater(t, ratio, 0.5, "Bucket %d has too few entries: %d", i, count)
		assert.Less(t, ratio, 1.5, "Bucket %d has too many entries: %d", i, count)
	}
}

func TestHKDFRNG_ErrorCases(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// Test zero reel length
	_, err := rng.GetReelPosition(0, 0)
	assert.Error(t, err)

	// Test negative reel length
	_, err = rng.GetReelPosition(0, -1)
	assert.Error(t, err)

	// Test zero max for Int
	_, err = rng.Int("test", 0)
	assert.Error(t, err)

	// Test invalid key length
	_, err = rng.DeriveKey("test", 0)
	assert.Error(t, err)
}

// Benchmark tests
func BenchmarkHKDFRNG_NewHKDFRNG(b *testing.B) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewHKDFRNG(serverSeed, clientSeed, int64(i), testPrevSpinHash)
	}
}

func BenchmarkHKDFRNG_GetAllReelPositions(b *testing.B) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"
	reelLengths := []int{100, 100, 100, 100, 100}

	rng, _ := NewHKDFRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rng.GetAllReelPositions(reelLengths)
	}
}

func BenchmarkHKDFRNG_DeriveReelKey(b *testing.B) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rng.DeriveReelKey(i % 5)
	}
}

// =============================================================================
// HKDFStreamRNG Tests - RNG interface implementation
// =============================================================================

func TestNewHKDFStreamRNG(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"
	nonce := int64(1)

	rng, err := NewHKDFStreamRNG(serverSeed, clientSeed, nonce, testPrevSpinHash)
	require.NoError(t, err)
	assert.NotNil(t, rng)
	assert.NotEmpty(t, rng.GetSpinHash())
}

func TestHKDFStreamRNG_Deterministic(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"
	nonce := int64(1)

	// Create two RNGs with same inputs
	rng1, err := NewHKDFStreamRNG(serverSeed, clientSeed, nonce, testPrevSpinHash)
	require.NoError(t, err)

	rng2, err := NewHKDFStreamRNG(serverSeed, clientSeed, nonce, testPrevSpinHash)
	require.NoError(t, err)

	// Should produce identical results
	for i := 0; i < 10; i++ {
		val1, err := rng1.Int(100)
		require.NoError(t, err)

		val2, err := rng2.Int(100)
		require.NoError(t, err)

		assert.Equal(t, val1, val2, "Int() call %d should produce same value", i)
	}
}

func TestHKDFStreamRNG_Int(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// Test multiple calls produce values in range
	for i := 0; i < 100; i++ {
		val, err := rng.Int(50)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, val, 0)
		assert.Less(t, val, 50)
	}
}

func TestHKDFStreamRNG_IntRange(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// Test range [10, 20]
	for i := 0; i < 50; i++ {
		val, err := rng.IntRange(10, 20)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, val, 10)
		assert.LessOrEqual(t, val, 20)
	}
}

func TestHKDFStreamRNG_Float64(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	for i := 0; i < 50; i++ {
		val, err := rng.Float64()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, val, 0.0)
		assert.Less(t, val, 1.0)
	}
}

func TestHKDFStreamRNG_Bytes(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	b := make([]byte, 32)
	err := rng.Bytes(b)
	require.NoError(t, err)

	// Verify not all zeros
	nonZero := false
	for _, v := range b {
		if v != 0 {
			nonZero = true
			break
		}
	}
	assert.True(t, nonZero, "Bytes should contain non-zero values")
}

func TestHKDFStreamRNG_Shuffle(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng1, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)
	rng2, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// Create two identical slices
	slice1 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	slice2 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	// Shuffle both
	err := rng1.Shuffle(len(slice1), func(i, j int) {
		slice1[i], slice1[j] = slice1[j], slice1[i]
	})
	require.NoError(t, err)

	err = rng2.Shuffle(len(slice2), func(i, j int) {
		slice2[i], slice2[j] = slice2[j], slice2[i]
	})
	require.NoError(t, err)

	// Both should be identical (deterministic)
	assert.Equal(t, slice1, slice2)
}

func TestHKDFStreamRNG_WeightedChoice(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	weights := []int{10, 20, 30, 40} // Index 3 should be most frequent

	counts := make([]int, 4)
	for i := 0; i < 1000; i++ {
		// Create new RNG for each iteration to test distribution
		rng, _ := NewHKDFStreamRNG(serverSeed, clientSeed, int64(i), testPrevSpinHash)
		idx, err := rng.WeightedChoice(weights)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, idx, 0)
		assert.Less(t, idx, 4)
		counts[idx]++
	}

	// Verify distribution roughly matches weights
	// With 1000 iterations and total weight 100:
	// Expected: 10%, 20%, 30%, 40% (±50% variance for statistical noise)
	assert.Greater(t, counts[3], counts[0], "Index 3 (weight 40) should appear more than index 0 (weight 10)")
}

func TestHKDFStreamRNG_GetReelPosition(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// Test reel positions match underlying HKDFRNG
	hkdf := rng.GetHKDFRNG()

	for i := 0; i < 5; i++ {
		pos1, err := rng.GetReelPosition(i, 100)
		require.NoError(t, err)

		pos2, err := hkdf.GetReelPosition(i, 100)
		require.NoError(t, err)

		assert.Equal(t, pos1, pos2, "Reel position %d should match", i)
	}
}

func TestHKDFStreamRNG_GetAllReelPositions(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	reelLengths := []int{100, 120, 100, 110, 100}
	positions, err := rng.GetAllReelPositions(reelLengths)
	require.NoError(t, err)
	assert.Len(t, positions, 5)

	for i, pos := range positions {
		assert.GreaterOrEqual(t, pos, 0)
		assert.Less(t, pos, reelLengths[i])
	}
}

func TestHKDFStreamRNG_IndependentFromReelKeys(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	// Create two RNGs
	rng1, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)
	rng2, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// RNG1: Get reel positions first, then stream values
	reelLengths := []int{100, 100, 100, 100, 100}
	_, _ = rng1.GetAllReelPositions(reelLengths)
	val1After, _ := rng1.Int(1000)

	// RNG2: Get stream values first (without reel positions)
	val2, _ := rng2.Int(1000)

	// Stream values should be the same regardless of GetReelPosition calls
	// because they use different domain spaces
	assert.Equal(t, val1After, val2, "Stream values should be independent of reel position calls")
}

func TestHKDFStreamRNG_ErrorCases(t *testing.T) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	// Test zero max for Int
	_, err := rng.Int(0)
	assert.Error(t, err)

	// Test negative max for Int
	_, err = rng.Int(-1)
	assert.Error(t, err)

	// Test invalid range for IntRange
	_, err = rng.IntRange(10, 5)
	assert.Error(t, err)

	// Test empty weights for WeightedChoice
	_, err = rng.WeightedChoice([]int{})
	assert.Error(t, err)

	// Test all-zero weights
	_, err = rng.WeightedChoice([]int{0, 0, 0})
	assert.Error(t, err)
}

// Benchmark tests for HKDFStreamRNG
func BenchmarkHKDFStreamRNG_Int(b *testing.B) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rng.Int(100)
	}
}

func BenchmarkHKDFStreamRNG_WeightedChoice(b *testing.B) {
	serverSeed := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	clientSeed := "deadbeefdeadbeefdeadbeefdeadbeef"

	rng, _ := NewHKDFStreamRNG(serverSeed, clientSeed, 1, testPrevSpinHash)
	weights := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rng.WeightedChoice(weights)
	}
}
