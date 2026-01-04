package rng

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/slotmachine/backend/domain/provablyfair"
)

// HashChainGenerator implements provably fair hash chain operations
type HashChainGenerator struct {
	cryptoRNG *CryptoRNG
}

// NewHashChainGenerator creates a new hash chain generator
func NewHashChainGenerator() *HashChainGenerator {
	return &HashChainGenerator{
		cryptoRNG: NewCryptoRNG(),
	}
}

// Ensure HashChainGenerator implements HashGenerator interface
var _ provablyfair.HashGenerator = (*HashChainGenerator)(nil)

// GenerateServerSeed generates a cryptographically secure 256-bit server seed
func (h *HashChainGenerator) GenerateServerSeed() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if err := h.cryptoRNG.Bytes(bytes); err != nil {
		return "", fmt.Errorf("failed to generate server seed: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// HashServerSeed creates SHA256 hash of server seed
// Used for commitment - public before session starts
func (h *HashChainGenerator) HashServerSeed(serverSeed string) string {
	hash := sha256.Sum256([]byte(serverSeed))
	return hex.EncodeToString(hash[:])
}

// GenerateSpinHash creates spin hash from previous hash + seeds + nonce
// spin_hash_n = SHA256(spin_hash_{n-1} + server_seed + client_seed + nonce)
func (h *HashChainGenerator) GenerateSpinHash(prevSpinHash, serverSeed, clientSeed string, nonce int64) string {
	// Format: prevSpinHash + serverSeed + clientSeed + nonce(as string)
	data := fmt.Sprintf("%s%s%s%d", prevSpinHash, serverSeed, clientSeed, nonce)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateClientSeed generates a random client seed if user doesn't provide one
func (h *HashChainGenerator) GenerateClientSeed() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if err := h.cryptoRNG.Bytes(bytes); err != nil {
		return "", fmt.Errorf("failed to generate client seed: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateInitialPrevSpinHash creates the initial prevSpinHash for first spin
// For Dual Commitment Protocol:
//
//	prevSpinHash = SHA256(server_seed_hash + theta_commitment)
//
// This ensures both server and client commitments are included in the RNG chain.
// If no theta_commitment (legacy sessions), falls back to just server_seed_hash.
func (h *HashChainGenerator) GenerateInitialPrevSpinHash(serverSeedHash, thetaCommitment string) string {
	if thetaCommitment == "" {
		// Legacy session without Dual Commitment - use server_seed_hash only
		return serverSeedHash
	}
	// Dual Commitment: combine both commitments
	data := serverSeedHash + thetaCommitment
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// VerifyHashChain verifies the entire hash chain for a session
// Each spin has its own client_seed provided per-spin
// First spin's prevSpinHash should be serverSeedHash
// Returns true if all hashes are valid
func (h *HashChainGenerator) VerifyHashChain(
	serverSeed, serverSeedHash string,
	spins []provablyfair.SpinVerification,
) (bool, error) {
	// Verify server seed hash matches
	expectedServerSeedHash := h.HashServerSeed(serverSeed)
	if expectedServerSeedHash != serverSeedHash {
		return false, fmt.Errorf("server seed hash mismatch: expected %s, got %s", expectedServerSeedHash, serverSeedHash)
	}

	// Verify each spin in the chain
	// First spin's prevSpinHash should be serverSeedHash
	prevHash := serverSeedHash
	for i, spin := range spins {
		// Each spin has its own client_seed
		expectedHash := h.GenerateSpinHash(prevHash, serverSeed, spin.ClientSeed, spin.Nonce)
		if expectedHash != spin.SpinHash {
			return false, fmt.Errorf("spin %d hash mismatch: expected %s, got %s", i+1, expectedHash, spin.SpinHash)
		}

		// Also verify prevSpinHash matches
		if spin.PrevSpinHash != prevHash {
			return false, fmt.Errorf("spin %d prev hash mismatch: expected %s, got %s", i+1, prevHash, spin.PrevSpinHash)
		}

		prevHash = spin.SpinHash
	}

	return true, nil
}
