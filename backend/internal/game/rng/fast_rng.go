package rng

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// FastRNG provides fast pseudo-random number generation using math/rand
// WARNING: NOT cryptographically secure - ONLY use for RTP tuning/simulation
// NEVER use in production game logic
type FastRNG struct {
	rng *rand.Rand
	mu  sync.Mutex
}

// NewFastRNG creates a new fast RNG for tuning/simulation
// Uses math/rand with time-based seed for performance
func NewFastRNG() *FastRNG {
	return &FastRNG{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewFastRNGWithSeed creates a fast RNG with a specific seed
func NewFastRNGWithSeed(seed int64) *FastRNG {
	return &FastRNG{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Int generates a random integer in range [0, max)
func (r *FastRNG) Int(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max must be positive, got %d", max)
	}

	r.mu.Lock()
	n := r.rng.Intn(max)
	r.mu.Unlock()

	return n, nil
}

// IntRange generates a random integer in range [min, max]
func (r *FastRNG) IntRange(min, max int) (int, error) {
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
func (r *FastRNG) Intn(max int) (int, error) {
	return r.Int(max)
}

// Float64 generates a random float64 in range [0.0, 1.0)
func (r *FastRNG) Float64() (float64, error) {
	r.mu.Lock()
	f := r.rng.Float64()
	r.mu.Unlock()

	return f, nil
}

// Bytes fills the provided byte slice with random bytes
func (r *FastRNG) Bytes(b []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range b {
		b[i] = byte(r.rng.Intn(256))
	}

	return nil
}

// Shuffle randomly shuffles a slice using Fisher-Yates algorithm
func (r *FastRNG) Shuffle(n int, swap func(i, j int)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := n - 1; i > 0; i-- {
		j := r.rng.Intn(i + 1)
		swap(i, j)
	}

	return nil
}

// WeightedChoice selects an index based on weights
func (r *FastRNG) WeightedChoice(weights []int) (int, error) {
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
	r.mu.Lock()
	randValue := r.rng.Intn(totalWeight)
	r.mu.Unlock()

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
