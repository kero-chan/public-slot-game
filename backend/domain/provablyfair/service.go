package provablyfair

import (
	"context"

	"github.com/google/uuid"
)

// Service defines the interface for provably fair operations
type Service interface {
	// StartSession creates a new provably fair session with Dual Commitment Protocol
	// thetaCommitment is SHA256(theta_seed) - client's commitment sent BEFORE seeing server_seed
	// Returns session ID, server_seed_hash, and nonce_start
	// Client seed is now provided per-spin, not per-session
	StartSession(ctx context.Context, playerID, gameSessionID uuid.UUID, thetaCommitment string) (*StartSessionResult, error)

	// GetSessionState retrieves the current session state for a spin
	// Used internally by spin service
	GetSessionState(ctx context.Context, gameSessionID uuid.UUID) (*PFSessionState, error)

	// RecordSpin records a spin and returns the spin hash
	// This should be called after the spin is executed
	RecordSpin(ctx context.Context, input *RecordSpinInput) (*SpinResult, error)

	// EndSession ends the session and reveals the server seed
	// Returns verification data for client
	EndSession(ctx context.Context, gameSessionID uuid.UUID) (*EndSessionResult, error)

	// GetVerificationData returns all data needed to verify a completed session
	GetVerificationData(ctx context.Context, pfSessionID uuid.UUID) (*VerificationData, error)

	// VerifySession verifies a completed session's hash chain
	// Used by clients to verify fairness
	VerifySession(ctx context.Context, pfSessionID uuid.UUID, serverSeed string) (bool, error)

	// VerifySpin verifies a single spin's hash
	// This is a stateless verification - no database access required
	// Returns the calculated spin hash and whether it matches
	VerifySpin(ctx context.Context, input *VerifySpinInput) (*VerifySpinResult, error)

	// VerifySpinWithReel verifies a spin hash and its reel positions
	// This verifies both the hash chain and the RNG-generated reel positions
	VerifySpinWithReel(ctx context.Context, input *VerifySpinWithReelInput) (*VerifySpinWithReelResult, error)

	// VerifyActiveSpin verifies a spin in an active session using server's stored server_seed
	// Client doesn't need to know server_seed - useful for real-time verification during play
	VerifyActiveSpin(ctx context.Context, gameSessionID uuid.UUID, input *VerifyActiveSpinInput) (*VerifyActiveSpinResult, error)

	// GetActiveSessionByPlayer retrieves the active PF session state by player ID
	// Used for verify-spin endpoint where we need to find the player's active session
	GetActiveSessionByPlayer(ctx context.Context, playerID uuid.UUID) (*PFSessionState, error)
}

// RecordSpinInput contains all data needed to record a spin in the PF system
type RecordSpinInput struct {
	GameSessionID     uuid.UUID  // Active game session
	SpinID            uuid.UUID  // The spin ID from spins table
	ClientSeed        string     // Client-provided seed for this spin (required for provably fair)
	ReelPositions     []int      // Array of 5 reel positions from RNG
	ReelStripConfigID *uuid.UUID // Which reel strip config was used
	GameMode          *string    // Game mode: nil for normal, or bonus_spin_trigger, etc.
	IsFreeSpin        bool       // Whether this was a free spin
	// Dual Commitment Protocol: theta_seed is revealed on first spin
	ThetaSeed string // Client's session seed - only required for first spin (nonce=1)
}

// StartSessionResult contains the result of starting a new PF session
// Client seed is now provided per-spin, not per-session
type StartSessionResult struct {
	SessionID      uuid.UUID `json:"session_id"`
	ServerSeedHash string    `json:"server_seed_hash"` // SHA256(server_seed) - commitment shown to player
	NonceStart     int64     `json:"nonce_start"`      // Always 1
}

// SpinResult contains the provably fair data for a spin
type SpinResult struct {
	SpinIndex    int64  `json:"spin_index"`
	Nonce        int64  `json:"nonce"`
	SpinHash     string `json:"spin_hash"`
	PrevSpinHash string `json:"prev_spin_hash"`
}

// EndSessionResult contains the result of ending a PF session
// Client seed is stored per-spin in SpinVerification, not per-session
type EndSessionResult struct {
	SessionID      uuid.UUID          `json:"session_id"`
	ServerSeed     string             `json:"server_seed"` // Revealed plaintext
	ServerSeedHash string             `json:"server_seed_hash"`
	TotalSpins     int64              `json:"total_spins"`
	Spins          []SpinVerification `json:"spins"` // Each spin has its own client_seed
}

// VerifySpinInput contains all data needed to verify a single spin
type VerifySpinInput struct {
	ServerSeed   string // 256-bit hex-encoded server seed
	ClientSeed   string // Per-spin client seed
	Nonce        int64  // Spin nonce (sequential, starting from 1)
	PrevSpinHash string // Previous spin hash (or server_seed_hash for first spin)
	SpinHash     string // The spin hash to verify
}

// VerifySpinResult contains the result of spin verification
type VerifySpinResult struct {
	Valid            bool   // Whether the spin hash matches
	ExpectedSpinHash string // The calculated spin hash
	ServerSeedHash   string // SHA256(server_seed) for reference
}

// VerifySpinWithReelInput contains data to verify spin and reel positions
type VerifySpinWithReelInput struct {
	ServerSeed        string    // 256-bit hex-encoded server seed
	ClientSeed        string    // Per-spin client seed
	Nonce             int64     // Spin nonce
	PrevSpinHash      string    // Previous spin hash
	SpinHash          string    // The spin hash to verify
	ReelPositions     []int     // The reel positions to verify
	ReelStripConfigID uuid.UUID // Config ID to lookup strip lengths
}

// VerifySpinWithReelResult contains the result including reel verification
type VerifySpinWithReelResult struct {
	Valid                 bool  // Whether both hash and reel positions match
	SpinHashValid         bool  // Whether the spin hash matches
	ReelPositionsValid    bool  // Whether the reel positions match
	ExpectedSpinHash      string
	ExpectedReelPositions []int
	ServerSeedHash        string
}

// VerifyActiveSpinInput contains data to verify a spin in an active session
type VerifyActiveSpinInput struct {
	ClientSeed        string    // Per-spin client seed
	Nonce             int64     // Spin nonce
	SpinHash          string    // The spin hash to verify
	PrevSpinHash      string    // Optional: if provided, use this instead of server-calculated value
	ReelPositions     []int     // Optional: reel positions to verify
	ReelStripConfigID uuid.UUID // Required if ReelPositions provided
}

// VerifyActiveSpinResult contains the result of active session spin verification
type VerifyActiveSpinResult struct {
	Valid                 bool   // Whether verification passed
	SpinHashValid         bool   // Whether the spin hash matches
	ReelPositionsValid    *bool  // nil if not checking, true/false otherwise
	ExpectedSpinHash      string // Calculated spin hash
	ExpectedReelPositions []int  // Only if ReelPositions was provided
	ServerSeedHash        string // Server seed hash (not plaintext)
	PrevSpinHash          string // The prev_spin_hash used for calculation (for debugging)
}

// HashGenerator defines the interface for hash chain operations
type HashGenerator interface {
	// GenerateServerSeed generates a cryptographically secure server seed (256-bit)
	GenerateServerSeed() (string, error)

	// HashServerSeed creates SHA256 hash of server seed
	HashServerSeed(serverSeed string) string

	// GenerateSpinHash creates spin hash from previous hash + seeds + nonce
	// Formula: SHA256(prev_spin_hash + server_seed + client_seed + nonce)
	GenerateSpinHash(prevSpinHash, serverSeed, clientSeed string, nonce int64) string

	// GenerateClientSeed generates a random client seed if user doesn't provide one
	// Used as fallback when client doesn't provide a seed
	GenerateClientSeed() (string, error)

	// GenerateInitialPrevSpinHash creates the initial prevSpinHash for first spin
	// For Dual Commitment Protocol: SHA256(server_seed_hash + theta_commitment)
	// For legacy sessions: just server_seed_hash
	GenerateInitialPrevSpinHash(serverSeedHash, thetaCommitment string) string
}
