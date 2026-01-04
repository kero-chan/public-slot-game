package rng

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// RNG is the interface for random number generation
// Supports both crypto-secure (production) and fast (tuning) implementations
type RNG interface {
	Int(max int) (int, error)
	IntRange(min, max int) (int, error)
	Intn(max int) (int, error)
	Float64() (float64, error)
	Bytes(b []byte) error
	Shuffle(n int, swap func(i, j int)) error
	WeightedChoice(weights []int) (int, error)
}

// CryptoRNG provides cryptographically secure random number generation
// CRITICAL: Uses crypto/rand ONLY - NEVER math/rand for gaming compliance
type CryptoRNG struct{}

// NewCryptoRNG creates a new cryptographically secure RNG
func NewCryptoRNG() *CryptoRNG {
	return &CryptoRNG{}
}

// Int generates a random integer in range [0, max)
// Uses crypto/rand for cryptographic security
func (r *CryptoRNG) Int(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max must be positive, got %d", max)
	}

	// crypto/rand.Int returns a uniform random value in [0, max)
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, fmt.Errorf("crypto RNG failed: %w", err)
	}

	return int(nBig.Int64()), nil
}

// IntRange generates a random integer in range [min, max]
func (r *CryptoRNG) IntRange(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min (%d) must be <= max (%d)", min, max)
	}

	rangeSize := max - min + 1
	n, err := r.Int(rangeSize)
	if err != nil {
		return 0, err
	}

	return min + n, nil
}

// Intn is an alias for Int for convenience
func (r *CryptoRNG) Intn(max int) (int, error) {
	return r.Int(max)
}

// Float64 generates a random float64 in range [0.0, 1.0)
func (r *CryptoRNG) Float64() (float64, error) {
	// Generate a random 53-bit integer (mantissa of float64)
	const precision = 1 << 53
	n, err := r.Int(precision)
	if err != nil {
		return 0, err
	}

	return float64(n) / float64(precision), nil
}

// Bytes fills the provided byte slice with random bytes
func (r *CryptoRNG) Bytes(b []byte) error {
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Errorf("crypto RNG read failed: %w", err)
	}
	return nil
}

// Shuffle randomly shuffles a slice using Fisher-Yates algorithm
// with cryptographically secure random numbers
func (r *CryptoRNG) Shuffle(n int, swap func(i, j int)) error {
	for i := n - 1; i > 0; i-- {
		j, err := r.Int(i + 1)
		if err != nil {
			return err
		}
		swap(i, j)
	}
	return nil
}

// WeightedChoice selects an index based on weights
// weights[i] represents the probability weight for index i
// Returns the selected index
func (r *CryptoRNG) WeightedChoice(weights []int) (int, error) {
	if len(weights) == 0 {
		return 0, fmt.Errorf("weights cannot be empty")
	}

	// Calculate total weight
	totalWeight := 0
	for _, w := range weights {
		if w < 0 {
			return 0, fmt.Errorf("weights must be non-negative")
		}
		totalWeight += w
	}

	if totalWeight == 0 {
		return 0, fmt.Errorf("total weight must be positive")
	}

	// Generate random number in [0, totalWeight)
	randValue, err := r.Int(totalWeight)
	if err != nil {
		return 0, err
	}

	// Find the index corresponding to the random value
	cumulative := 0
	for i, w := range weights {
		cumulative += w
		if randValue < cumulative {
			return i, nil
		}
	}

	// Should never reach here, but return last index as failsafe
	return len(weights) - 1, nil
}
