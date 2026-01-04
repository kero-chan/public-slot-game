package rng

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateServerSeed(t *testing.T) {
	t.Run("should generate valid server seed", func(t *testing.T) {
		seed, err := GenerateServerSeed()

		require.NoError(t, err)
		assert.NotEmpty(t, seed)

		// Should be 64 hex characters (32 bytes * 2)
		assert.Len(t, seed, 64)

		// Should be valid hex
		_, err = hex.DecodeString(seed)
		assert.NoError(t, err)
	})

	t.Run("should generate unique seeds", func(t *testing.T) {
		seeds := make(map[string]bool)
		iterations := 10000

		for i := 0; i < iterations; i++ {
			seed, err := GenerateServerSeed()
			require.NoError(t, err)

			// Check for duplicates
			if seeds[seed] {
				t.Fatalf("Duplicate seed generated: %s", seed)
			}
			seeds[seed] = true
		}

		assert.Equal(t, iterations, len(seeds), "All seeds should be unique")
	})

	t.Run("should generate 256-bit seeds", func(t *testing.T) {
		seed, err := GenerateServerSeed()
		require.NoError(t, err)

		// Decode hex to bytes
		bytes, err := hex.DecodeString(seed)
		require.NoError(t, err)

		// Should be 32 bytes (256 bits)
		assert.Len(t, bytes, 32)
	})
}

func TestGenerateNonce(t *testing.T) {
	t.Run("should generate nonce", func(t *testing.T) {
		nonce := GenerateNonce()
		assert.Greater(t, nonce, int64(0))
	})

	t.Run("should generate unique nonces", func(t *testing.T) {
		nonce1 := GenerateNonce()
		time.Sleep(1 * time.Millisecond)
		nonce2 := GenerateNonce()

		assert.NotEqual(t, nonce1, nonce2, "Nonces should be unique")
	})

	t.Run("should generate increasing nonces", func(t *testing.T) {
		nonce1 := GenerateNonce()
		time.Sleep(1 * time.Millisecond)
		nonce2 := GenerateNonce()

		assert.Greater(t, nonce2, nonce1, "Later nonce should be greater")
	})
}

func TestCalculateChecksum(t *testing.T) {
	spinID := uuid.New()
	playerID := uuid.New()

	t.Run("should calculate checksum", func(t *testing.T) {
		spin := &ProvablyFairSpin{
			SpinID:      spinID,
			PlayerID:    playerID,
			ServerSeed:  "test_seed",
			Nonce:       12345,
			ReelStops:   []int{10, 20, 30, 40, 50},
			GameVersion: "1.0.0",
		}

		checksum := CalculateChecksum(spin)

		assert.NotEmpty(t, checksum)
		assert.Len(t, checksum, 64) // SHA256 hex = 64 chars

		// Should be valid hex
		_, err := hex.DecodeString(checksum)
		assert.NoError(t, err)
	})

	t.Run("should be deterministic", func(t *testing.T) {
		spin := &ProvablyFairSpin{
			SpinID:      spinID,
			PlayerID:    playerID,
			ServerSeed:  "test_seed",
			Nonce:       12345,
			ReelStops:   []int{10, 20, 30, 40, 50},
			GameVersion: "1.0.0",
		}

		checksum1 := CalculateChecksum(spin)
		checksum2 := CalculateChecksum(spin)

		assert.Equal(t, checksum1, checksum2, "Checksum should be deterministic")
	})

	t.Run("should change with different server seed", func(t *testing.T) {
		spin1 := &ProvablyFairSpin{
			SpinID:      spinID,
			PlayerID:    playerID,
			ServerSeed:  "seed1",
			Nonce:       12345,
			ReelStops:   []int{10, 20, 30, 40, 50},
			GameVersion: "1.0.0",
		}

		spin2 := &ProvablyFairSpin{
			SpinID:      spinID,
			PlayerID:    playerID,
			ServerSeed:  "seed2",
			Nonce:       12345,
			ReelStops:   []int{10, 20, 30, 40, 50},
			GameVersion: "1.0.0",
		}

		checksum1 := CalculateChecksum(spin1)
		checksum2 := CalculateChecksum(spin2)

		assert.NotEqual(t, checksum1, checksum2, "Different seeds should produce different checksums")
	})

	t.Run("should change with different reel stops", func(t *testing.T) {
		spin1 := &ProvablyFairSpin{
			SpinID:      spinID,
			PlayerID:    playerID,
			ServerSeed:  "test_seed",
			Nonce:       12345,
			ReelStops:   []int{10, 20, 30, 40, 50},
			GameVersion: "1.0.0",
		}

		spin2 := &ProvablyFairSpin{
			SpinID:      spinID,
			PlayerID:    playerID,
			ServerSeed:  "test_seed",
			Nonce:       12345,
			ReelStops:   []int{15, 25, 35, 45, 55},
			GameVersion: "1.0.0",
		}

		checksum1 := CalculateChecksum(spin1)
		checksum2 := CalculateChecksum(spin2)

		assert.NotEqual(t, checksum1, checksum2, "Different reel stops should produce different checksums")
	})

	t.Run("should change with different nonce", func(t *testing.T) {
		spin1 := &ProvablyFairSpin{
			SpinID:      spinID,
			PlayerID:    playerID,
			ServerSeed:  "test_seed",
			Nonce:       12345,
			ReelStops:   []int{10, 20, 30, 40, 50},
			GameVersion: "1.0.0",
		}

		spin2 := &ProvablyFairSpin{
			SpinID:      spinID,
			PlayerID:    playerID,
			ServerSeed:  "test_seed",
			Nonce:       54321,
			ReelStops:   []int{10, 20, 30, 40, 50},
			GameVersion: "1.0.0",
		}

		checksum1 := CalculateChecksum(spin1)
		checksum2 := CalculateChecksum(spin2)

		assert.NotEqual(t, checksum1, checksum2, "Different nonces should produce different checksums")
	})
}

func TestNewProvablyFairSpin(t *testing.T) {
	spinID := uuid.New()
	playerID := uuid.New()
	reelStops := []int{10, 20, 30, 40, 50}
	gameVersion := "1.0.0"

	t.Run("should create provably fair spin", func(t *testing.T) {
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)

		require.NoError(t, err)
		assert.NotNil(t, spin)

		assert.Equal(t, spinID, spin.SpinID)
		assert.Equal(t, playerID, spin.PlayerID)
		assert.Equal(t, reelStops, spin.ReelStops)
		assert.Equal(t, gameVersion, spin.GameVersion)
	})

	t.Run("should generate server seed", func(t *testing.T) {
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)

		require.NoError(t, err)
		assert.NotEmpty(t, spin.ServerSeed)
		assert.Len(t, spin.ServerSeed, 64)
	})

	t.Run("should generate nonce", func(t *testing.T) {
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)

		require.NoError(t, err)
		assert.Greater(t, spin.Nonce, int64(0))
	})

	t.Run("should set timestamp", func(t *testing.T) {
		before := time.Now().UTC()
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
		after := time.Now().UTC()

		require.NoError(t, err)
		assert.True(t, spin.Timestamp.After(before) || spin.Timestamp.Equal(before))
		assert.True(t, spin.Timestamp.Before(after) || spin.Timestamp.Equal(after))
	})

	t.Run("should calculate checksum", func(t *testing.T) {
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)

		require.NoError(t, err)
		assert.NotEmpty(t, spin.Checksum)
		assert.Len(t, spin.Checksum, 64)

		// Verify checksum is correct
		expectedChecksum := CalculateChecksum(spin)
		assert.Equal(t, expectedChecksum, spin.Checksum)
	})

	t.Run("should create unique spins", func(t *testing.T) {
		spin1, err1 := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
		time.Sleep(1 * time.Millisecond)
		spin2, err2 := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)

		require.NoError(t, err1)
		require.NoError(t, err2)

		// Different server seeds
		assert.NotEqual(t, spin1.ServerSeed, spin2.ServerSeed)

		// Different nonces
		assert.NotEqual(t, spin1.Nonce, spin2.Nonce)

		// Different checksums
		assert.NotEqual(t, spin1.Checksum, spin2.Checksum)
	})
}

func TestProvablyFairSpin_Verify(t *testing.T) {
	spinID := uuid.New()
	playerID := uuid.New()
	reelStops := []int{10, 20, 30, 40, 50}
	gameVersion := "1.0.0"

	t.Run("should verify valid spin", func(t *testing.T) {
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
		require.NoError(t, err)

		isValid := spin.Verify()
		assert.True(t, isValid, "Valid spin should pass verification")
	})

	t.Run("should reject tampered server seed", func(t *testing.T) {
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
		require.NoError(t, err)

		// Tamper with server seed
		spin.ServerSeed = "tampered_seed"

		isValid := spin.Verify()
		assert.False(t, isValid, "Tampered server seed should fail verification")
	})

	t.Run("should reject tampered reel stops", func(t *testing.T) {
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
		require.NoError(t, err)

		// Tamper with reel stops
		spin.ReelStops = []int{99, 99, 99, 99, 99}

		isValid := spin.Verify()
		assert.False(t, isValid, "Tampered reel stops should fail verification")
	})

	t.Run("should reject tampered nonce", func(t *testing.T) {
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
		require.NoError(t, err)

		// Tamper with nonce
		spin.Nonce = 99999

		isValid := spin.Verify()
		assert.False(t, isValid, "Tampered nonce should fail verification")
	})

	t.Run("should reject tampered checksum", func(t *testing.T) {
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
		require.NoError(t, err)

		// Tamper with checksum
		spin.Checksum = "0000000000000000000000000000000000000000000000000000000000000000"

		isValid := spin.Verify()
		assert.False(t, isValid, "Tampered checksum should fail verification")
	})

	t.Run("should verify after multiple checks", func(t *testing.T) {
		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
		require.NoError(t, err)

		// Verify multiple times
		assert.True(t, spin.Verify())
		assert.True(t, spin.Verify())
		assert.True(t, spin.Verify())
	})
}

func TestProvablyFairSpin_SecurityProperties(t *testing.T) {
	t.Run("should not be predictable", func(t *testing.T) {
		spinID := uuid.New()
		playerID := uuid.New()
		reelStops := []int{10, 20, 30, 40, 50}
		gameVersion := "1.0.0"

		// Generate multiple spins
		spins := make([]*ProvablyFairSpin, 100)
		for i := 0; i < 100; i++ {
			spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
			require.NoError(t, err)
			spins[i] = spin
		}

		// All server seeds should be unique
		seeds := make(map[string]bool)
		for _, spin := range spins {
			if seeds[spin.ServerSeed] {
				t.Fatal("Server seeds should be unique and unpredictable")
			}
			seeds[spin.ServerSeed] = true
		}
	})

	t.Run("should have cryptographic randomness", func(t *testing.T) {
		spinID := uuid.New()
		playerID := uuid.New()
		reelStops := []int{10, 20, 30, 40, 50}
		gameVersion := "1.0.0"

		spin, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
		require.NoError(t, err)

		// Decode server seed
		seedBytes, err := hex.DecodeString(spin.ServerSeed)
		require.NoError(t, err)

		// Check that not all bytes are the same (basic randomness test)
		allSame := true
		firstByte := seedBytes[0]
		for _, b := range seedBytes {
			if b != firstByte {
				allSame = false
				break
			}
		}

		assert.False(t, allSame, "Server seed should have cryptographic randomness")
	})
}

func BenchmarkGenerateServerSeed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateServerSeed()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCalculateChecksum(b *testing.B) {
	spin := &ProvablyFairSpin{
		SpinID:      uuid.New(),
		PlayerID:    uuid.New(),
		ServerSeed:  "test_seed",
		Nonce:       12345,
		ReelStops:   []int{10, 20, 30, 40, 50},
		GameVersion: "1.0.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateChecksum(spin)
	}
}

func BenchmarkNewProvablyFairSpin(b *testing.B) {
	spinID := uuid.New()
	playerID := uuid.New()
	reelStops := []int{10, 20, 30, 40, 50}
	gameVersion := "1.0.0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewProvablyFairSpin(spinID, playerID, reelStops, gameVersion)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerify(b *testing.B) {
	spin, err := NewProvablyFairSpin(uuid.New(), uuid.New(), []int{10, 20, 30, 40, 50}, "1.0.0")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = spin.Verify()
	}
}
