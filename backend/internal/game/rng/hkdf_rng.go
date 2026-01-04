package rng

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// HKDFRNG implements RFC 5869 HKDF for deriving independent keys per reel
// This provides cryptographic domain separation - each reel has its own
// independent key derived from the master seed, ensuring no correlation
// between reel outcomes.
//
// Benefits over simple SHA256:
// - Each reel key is cryptographically independent
// - Standard RFC 5869 compliance for audit/compliance
// - Domain separation via 'info' parameter
// - Extensible to any number of keys needed
type HKDFRNG struct {
	masterKey []byte
	spinHash  string
}

// NewHKDFRNG creates a new HKDF-based RNG from combined seeds
//
// Algorithm:
//  1. IKM (Input Keying Material) = prevSpinHash || clientSeed || nonce
//  2. Salt = serverSeed
//  3. PRK = HMAC-SHA256(salt, IKM)  [Extract step]
//  4. masterKey = HKDF-Expand(PRK, "spin-master-v1", 32)  [Expand step]
//
// The masterKey is then used to derive per-reel keys via DeriveReelKey()
//
// IMPORTANT: prevSpinHash MUST be included in IKM to ensure:
// - Each spin's RNG depends on ALL previous spins (hash chain)
// - Entropy accumulation across the session
// - Server cannot pre-compute outcomes for multiple spins
func NewHKDFRNG(serverSeed, clientSeed string, nonce int64, prevSpinHash string) (*HKDFRNG, error) {
	// IKM = prevSpinHash || clientSeed || nonce (as string)
	// prevSpinHash links this spin to the entire previous chain
	ikm := []byte(fmt.Sprintf("%s%s%d", prevSpinHash, clientSeed, nonce))

	// Salt = serverSeed
	salt := []byte(serverSeed)

	// HKDF Extract + Expand to get master key
	// Using SHA256 as the underlying hash function
	hkdfReader := hkdf.New(sha256.New, ikm, salt, []byte("spin-master-v1"))

	masterKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, masterKey); err != nil {
		return nil, fmt.Errorf("HKDF failed: %w", err)
	}

	return &HKDFRNG{
		masterKey: masterKey,
		spinHash:  hex.EncodeToString(masterKey),
	}, nil
}

// DeriveReelKey derives an independent 32-byte key for a specific reel
//
// Algorithm:
//
//	reelKey = HKDF-Expand(masterKey, "reel:<index>", 32)
//
// Each reel gets its own cryptographically independent key.
// Compromise of one reel's key does not affect others.
func (r *HKDFRNG) DeriveReelKey(reelIndex int) ([]byte, error) {
	// Domain separation: each reel has unique 'info' context
	info := []byte(fmt.Sprintf("reel:%d", reelIndex))

	// HKDF-Expand using masterKey as PRK
	hkdfReader := hkdf.New(sha256.New, r.masterKey, nil, info)

	reelKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, reelKey); err != nil {
		return nil, fmt.Errorf("HKDF expand for reel %d failed: %w", reelIndex, err)
	}

	return reelKey, nil
}

// GetReelPosition derives position for a specific reel using HKDF
// Returns a uniform random value in range [0, reelLength)
// Uses rejection sampling to eliminate modulo bias
func (r *HKDFRNG) GetReelPosition(reelIndex, reelLength int) (int, error) {
	if reelLength <= 0 {
		return 0, fmt.Errorf("reelLength must be positive, got %d", reelLength)
	}

	// Calculate rejection threshold to eliminate modulo bias
	// threshold = (2^64 - reelLength) % reelLength = -reelLength % reelLength
	umax := uint64(reelLength)
	threshold := -umax % umax

	// Use counter suffix for rejection sampling with deterministic HKDF
	for attempt := 0; attempt < 100; attempt++ {
		// Domain includes attempt counter for unique derivation on retry
		info := []byte(fmt.Sprintf("reel:%d:%d", reelIndex, attempt))
		hkdfReader := hkdf.New(sha256.New, r.masterKey, nil, info)

		key := make([]byte, 8)
		if _, err := io.ReadFull(hkdfReader, key); err != nil {
			return 0, fmt.Errorf("HKDF expand for reel %d failed: %w", reelIndex, err)
		}

		value := binary.BigEndian.Uint64(key)

		// Reject values in the biased zone (values < threshold)
		if value >= threshold {
			return int(value % umax), nil
		}
		// Continue to next attempt with different domain
	}

	return 0, fmt.Errorf("rejection sampling failed after 100 attempts for reel %d", reelIndex)
}

// GetAllReelPositions derives positions for all reels
func (r *HKDFRNG) GetAllReelPositions(reelLengths []int) ([]int, error) {
	positions := make([]int, len(reelLengths))

	for i, length := range reelLengths {
		pos, err := r.GetReelPosition(i, length)
		if err != nil {
			return nil, err
		}
		positions[i] = pos
	}

	return positions, nil
}

// GetSpinHash returns the master key hash for verification/logging
func (r *HKDFRNG) GetSpinHash() string {
	return r.spinHash
}

// GetMasterKey returns the raw master key bytes (for advanced use cases)
func (r *HKDFRNG) GetMasterKey() []byte {
	return r.masterKey
}

// DeriveKey derives a key for any purpose using domain separation
// This is a generic method for extending HKDF to other use cases
//
// Example domains:
//   - "reel:0", "reel:1", etc. for reel positions
//   - "multiplier" for multiplier selection
//   - "bonus:trigger" for bonus game triggers
func (r *HKDFRNG) DeriveKey(domain string, length int) ([]byte, error) {
	if length <= 0 || length > 255*32 {
		return nil, fmt.Errorf("invalid key length: %d (must be 1-%d)", length, 255*32)
	}

	hkdfReader := hkdf.New(sha256.New, r.masterKey, nil, []byte(domain))

	key := make([]byte, length)
	if _, err := io.ReadFull(hkdfReader, key); err != nil {
		return nil, fmt.Errorf("HKDF expand for domain '%s' failed: %w", domain, err)
	}

	return key, nil
}

// Int derives a random integer in range [0, max) for a specific domain
// Uses rejection sampling to eliminate modulo bias
func (r *HKDFRNG) Int(domain string, max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max must be positive, got %d", max)
	}

	// Calculate rejection threshold to eliminate modulo bias
	umax := uint64(max)
	threshold := -umax % umax

	// Use counter suffix for rejection sampling with deterministic HKDF
	for attempt := 0; attempt < 100; attempt++ {
		// Domain includes attempt counter for unique derivation on retry
		info := []byte(fmt.Sprintf("%s:%d", domain, attempt))
		hkdfReader := hkdf.New(sha256.New, r.masterKey, nil, info)

		key := make([]byte, 8)
		if _, err := io.ReadFull(hkdfReader, key); err != nil {
			return 0, fmt.Errorf("HKDF expand for domain '%s' failed: %w", domain, err)
		}

		value := binary.BigEndian.Uint64(key)

		// Reject values in the biased zone
		if value >= threshold {
			return int(value % umax), nil
		}
	}

	return 0, fmt.Errorf("rejection sampling failed after 100 attempts for domain '%s'", domain)
}

// Float64 derives a random float64 in range [0.0, 1.0) for a specific domain
func (r *HKDFRNG) Float64(domain string) (float64, error) {
	key, err := r.DeriveKey(domain, 8)
	if err != nil {
		return 0, err
	}

	value := binary.BigEndian.Uint64(key)
	const precision = 1 << 53
	return float64(value%precision) / float64(precision), nil
}

// VerifyReelPosition verifies that a reel position was correctly derived
// Used by clients to verify server's claimed positions
//
// prevSpinHash is required to maintain hash chain integrity:
// - For first spin (nonce=1): use server_seed_hash
// - For subsequent spins: use previous spin's hash
func VerifyReelPosition(serverSeed, clientSeed string, nonce int64, prevSpinHash string, reelIndex, reelLength, expectedPosition int) (bool, error) {
	rng, err := NewHKDFRNG(serverSeed, clientSeed, nonce, prevSpinHash)
	if err != nil {
		return false, err
	}

	actualPosition, err := rng.GetReelPosition(reelIndex, reelLength)
	if err != nil {
		return false, err
	}

	return actualPosition == expectedPosition, nil
}

// VerifyAllReelPositions verifies all reel positions for a spin
//
// prevSpinHash is required to maintain hash chain integrity:
// - For first spin (nonce=1): use server_seed_hash
// - For subsequent spins: use previous spin's hash
func VerifyAllReelPositions(serverSeed, clientSeed string, nonce int64, prevSpinHash string, reelLengths, expectedPositions []int) (bool, error) {
	if len(reelLengths) != len(expectedPositions) {
		return false, fmt.Errorf("reelLengths and expectedPositions must have same length")
	}

	rng, err := NewHKDFRNG(serverSeed, clientSeed, nonce, prevSpinHash)
	if err != nil {
		return false, err
	}

	actualPositions, err := rng.GetAllReelPositions(reelLengths)
	if err != nil {
		return false, err
	}

	for i, expected := range expectedPositions {
		if actualPositions[i] != expected {
			return false, nil
		}
	}

	return true, nil
}

// =============================================================================
// HKDFStreamRNG - RNG interface implementation using HKDF
// =============================================================================

// HKDFStreamRNG wraps HKDFRNG to implement the RNG interface
// It uses an internal counter for domain separation, making it compatible
// with the existing game engine that expects stateful RNG calls.
//
// Each call to Int(), Float64(), etc. uses a unique domain "stream:<counter>"
// This provides deterministic, reproducible random streams while maintaining
// cryptographic independence between operations.
type HKDFStreamRNG struct {
	hkdf    *HKDFRNG
	counter int
}

// NewHKDFStreamRNG creates a new HKDF-based RNG that implements the RNG interface
// This is the main entry point for provably fair gaming
//
// prevSpinHash MUST be included to maintain hash chain integrity:
// - For first spin (nonce=1): use server_seed_hash
// - For subsequent spins: use previous spin's hash
// This ensures each spin's RNG depends on all previous spins
func NewHKDFStreamRNG(serverSeed, clientSeed string, nonce int64, prevSpinHash string) (*HKDFStreamRNG, error) {
	hkdf, err := NewHKDFRNG(serverSeed, clientSeed, nonce, prevSpinHash)
	if err != nil {
		return nil, err
	}

	return &HKDFStreamRNG{
		hkdf:    hkdf,
		counter: 0,
	}, nil
}

// nextDomain generates a unique domain string for each RNG operation
func (r *HKDFStreamRNG) nextDomain() string {
	domain := fmt.Sprintf("stream:%d", r.counter)
	r.counter++
	return domain
}

// Int generates a random integer in range [0, max)
// Each call uses a unique domain for cryptographic independence
func (r *HKDFStreamRNG) Int(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max must be positive, got %d", max)
	}

	domain := r.nextDomain()
	return r.hkdf.Int(domain, max)
}

// IntRange generates a random integer in range [min, max]
func (r *HKDFStreamRNG) IntRange(min, max int) (int, error) {
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

// Intn is an alias for Int
func (r *HKDFStreamRNG) Intn(max int) (int, error) {
	return r.Int(max)
}

// Float64 generates a random float64 in range [0.0, 1.0)
func (r *HKDFStreamRNG) Float64() (float64, error) {
	domain := r.nextDomain()
	return r.hkdf.Float64(domain)
}

// Bytes fills the provided byte slice with random bytes
func (r *HKDFStreamRNG) Bytes(b []byte) error {
	domain := r.nextDomain()
	key, err := r.hkdf.DeriveKey(domain, len(b))
	if err != nil {
		return err
	}
	copy(b, key)
	return nil
}

// Shuffle randomly shuffles using deterministic RNG
func (r *HKDFStreamRNG) Shuffle(n int, swap func(i, j int)) error {
	for i := n - 1; i > 0; i-- {
		j, err := r.Int(i + 1)
		if err != nil {
			return err
		}
		swap(i, j)
	}
	return nil
}

// WeightedChoice selects an index based on weights using deterministic RNG
func (r *HKDFStreamRNG) WeightedChoice(weights []int) (int, error) {
	if len(weights) == 0 {
		return 0, fmt.Errorf("weights cannot be empty")
	}

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

	randValue, err := r.Int(totalWeight)
	if err != nil {
		return 0, err
	}

	cumulative := 0
	for i, w := range weights {
		cumulative += w
		if randValue < cumulative {
			return i, nil
		}
	}

	return len(weights) - 1, nil
}

// GetSpinHash returns the master key hash for verification/logging
func (r *HKDFStreamRNG) GetSpinHash() string {
	return r.hkdf.GetSpinHash()
}

// GetHKDFRNG returns the underlying HKDFRNG for direct reel position access
func (r *HKDFStreamRNG) GetHKDFRNG() *HKDFRNG {
	return r.hkdf
}

// GetReelPosition derives position for a specific reel using HKDF
// This delegates to the underlying HKDFRNG for consistent reel position generation
func (r *HKDFStreamRNG) GetReelPosition(reelIndex, reelLength int) (int, error) {
	return r.hkdf.GetReelPosition(reelIndex, reelLength)
}

// GetAllReelPositions derives positions for all reels
func (r *HKDFStreamRNG) GetAllReelPositions(reelLengths []int) ([]int, error) {
	return r.hkdf.GetAllReelPositions(reelLengths)
}

// Ensure HKDFStreamRNG implements RNG interface
var _ RNG = (*HKDFStreamRNG)(nil)
