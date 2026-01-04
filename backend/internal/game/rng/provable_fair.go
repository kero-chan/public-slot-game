package rng

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ProvablyFairSpin contains data for provably fair verification
type ProvablyFairSpin struct {
	SpinID       uuid.UUID `json:"spin_id"`
	PlayerID     uuid.UUID `json:"player_id"`
	ServerSeed   string    `json:"server_seed"`
	ClientSeed   string    `json:"client_seed,omitempty"` // Optional for now
	Nonce        int64     `json:"nonce"`
	ReelStops    []int     `json:"reel_stops"`
	Timestamp    time.Time `json:"timestamp"`
	Checksum     string    `json:"checksum"`
	GameVersion  string    `json:"game_version"`
}

// GenerateServerSeed generates a cryptographically secure server seed
func GenerateServerSeed() (string, error) {
	rng := NewCryptoRNG()
	bytes := make([]byte, 32) // 256 bits
	if err := rng.Bytes(bytes); err != nil {
		return "", fmt.Errorf("failed to generate server seed: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateNonce generates a unique nonce (timestamp-based)
func GenerateNonce() int64 {
	return time.Now().UnixNano()
}

// CalculateChecksum calculates SHA256 checksum for verification
func CalculateChecksum(spinData *ProvablyFairSpin) string {
	data := fmt.Sprintf("%s:%s:%s:%d:%v:%s",
		spinData.SpinID.String(),
		spinData.PlayerID.String(),
		spinData.ServerSeed,
		spinData.Nonce,
		spinData.ReelStops,
		spinData.GameVersion,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// NewProvablyFairSpin creates a new provably fair spin record
func NewProvablyFairSpin(spinID, playerID uuid.UUID, reelStops []int, gameVersion string) (*ProvablyFairSpin, error) {
	serverSeed, err := GenerateServerSeed()
	if err != nil {
		return nil, err
	}

	spin := &ProvablyFairSpin{
		SpinID:      spinID,
		PlayerID:    playerID,
		ServerSeed:  serverSeed,
		Nonce:       GenerateNonce(),
		ReelStops:   reelStops,
		Timestamp:   time.Now().UTC(),
		GameVersion: gameVersion,
	}

	// Calculate checksum
	spin.Checksum = CalculateChecksum(spin)

	return spin, nil
}

// Verify verifies the integrity of a provably fair spin
func (pf *ProvablyFairSpin) Verify() bool {
	expectedChecksum := CalculateChecksum(pf)
	return pf.Checksum == expectedChecksum
}

// ProvablyFairLogger logs provably fair data for auditing
type ProvablyFairLogger interface {
	Log(spin *ProvablyFairSpin) error
}

// Note: Actual logging implementation would be in the service layer
// This could log to:
// - Database (audit_logs table)
// - File system (for regulatory compliance)
// - External audit service
