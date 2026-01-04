package dto

import "time"

// StartPFSessionRequest represents a request to start a provably fair session
// No client_seed here - it's provided per-spin now
type StartPFSessionRequest struct {
	// Empty - no fields needed for starting a PF session
	// Client seed is now provided per-spin
}

// StartPFSessionResponse represents the response after starting a PF session
// Client seed is now provided per-spin, not per-session
type StartPFSessionResponse struct {
	SessionID      string `json:"session_id"`
	ServerSeedHash string `json:"server_seed_hash"` // SHA256(server_seed) - commitment shown to player
	NonceStart     int64  `json:"nonce_start"`      // Always 1
}

// EndPFSessionResponse represents the response after ending a PF session
// Client seed is stored per-spin, not per-session
type EndPFSessionResponse struct {
	SessionID      string                 `json:"session_id"`
	ServerSeed     string                 `json:"server_seed"` // Revealed plaintext
	ServerSeedHash string                 `json:"server_seed_hash"`
	TotalSpins     int64                  `json:"total_spins"`
	Spins          []SpinVerificationData `json:"spins"` // Each spin has its own client_seed
}

// SpinVerificationData contains data for verifying a single spin
// Includes per-spin client_seed for provably fair verification
type SpinVerificationData struct {
	SpinIndex         int64   `json:"spin_index"`
	Nonce             int64   `json:"nonce"`
	ClientSeed        string  `json:"client_seed"`                    // Per-spin client seed
	SpinHash          string  `json:"spin_hash"`
	PrevSpinHash      string  `json:"prev_spin_hash"`
	ReelPositions     []int   `json:"reel_positions"`                 // Array of 5 reel positions from RNG
	ReelStripConfigID *string `json:"reel_strip_config_id,omitempty"` // Which reel strip config was used
	GameMode          *string `json:"game_mode,omitempty"`            // Game mode if any
	IsFreeSpin        bool    `json:"is_free_spin"`
}

// PFSessionStatusResponse represents the current status of a PF session
// Client seed is now provided per-spin, not stored in session
type PFSessionStatusResponse struct {
	SessionID      string `json:"session_id"`
	ServerSeedHash string `json:"server_seed_hash"` // SHA256(server_seed) - commitment
	CurrentNonce   int64  `json:"current_nonce"`
	LastSpinHash   string `json:"last_spin_hash"` // Last spin's hash (or serverSeedHash if no spins)
	Status         string `json:"status"`         // active, ended
}

// SpinPFDataResponse represents PF data included in spin response
type SpinPFDataResponse struct {
	SpinIndex    int64  `json:"spin_index"`
	Nonce        int64  `json:"nonce"`
	SpinHash     string `json:"spin_hash"`
	PrevSpinHash string `json:"prev_spin_hash"`
}

// VerificationDataResponse represents all data needed for client verification
// Each spin has its own client_seed stored in SpinVerificationData
type VerificationDataResponse struct {
	SessionID      string                 `json:"session_id"`
	ServerSeed     string                 `json:"server_seed"`      // Revealed after session ends
	ServerSeedHash string                 `json:"server_seed_hash"` // SHA256(server_seed)
	Spins          []SpinVerificationData `json:"spins"`            // Each spin has its own client_seed
}

// VerifySessionRequest represents a request to verify a session
type VerifySessionRequest struct {
	ServerSeed string `json:"server_seed" validate:"required,len=64"` // Hex-encoded 256-bit seed
}

// VerifySessionResponse represents the result of session verification
type VerifySessionResponse struct {
	SessionID string `json:"session_id"`
	Valid     bool   `json:"valid"`
	Message   string `json:"message,omitempty"`
}

// PFSessionAuditResponse represents audit data for a completed session
type PFSessionAuditResponse struct {
	SessionID           string    `json:"session_id"`
	ServerSeedPlaintext string    `json:"server_seed_plaintext"`
	ServerSeedHash      string    `json:"server_seed_hash"`
	TotalSpins          int64     `json:"total_spins"`
	RevealedAt          time.Time `json:"revealed_at"`
}

// VerifySpinRequest represents a request to verify a single spin
type VerifySpinRequest struct {
	ServerSeed   string `json:"server_seed" validate:"required,len=64"`    // Hex-encoded 256-bit server seed
	ClientSeed   string `json:"client_seed" validate:"required"`           // Per-spin client seed
	Nonce        int64  `json:"nonce" validate:"required,min=1"`           // Spin nonce
	PrevSpinHash string `json:"prev_spin_hash" validate:"required,len=64"` // Previous spin hash (or server_seed_hash for first spin)
	SpinHash     string `json:"spin_hash" validate:"required,len=64"`      // The spin hash to verify
}

// VerifySpinResponse represents the result of single spin verification
type VerifySpinResponse struct {
	Valid              bool   `json:"valid"`
	ExpectedSpinHash   string `json:"expected_spin_hash,omitempty"`   // Calculated hash for comparison
	ProvidedSpinHash   string `json:"provided_spin_hash,omitempty"`   // The hash provided in request
	ServerSeedHashOK   bool   `json:"server_seed_hash_ok"`            // Whether server_seed hashes to expected commitment
	ServerSeedHash     string `json:"server_seed_hash,omitempty"`     // Calculated server seed hash
	Message            string `json:"message,omitempty"`
}

// VerifySpinWithReelRequest verifies spin and reel positions
type VerifySpinWithReelRequest struct {
	ServerSeed        string `json:"server_seed" validate:"required,len=64"`
	ClientSeed        string `json:"client_seed" validate:"required"`
	Nonce             int64  `json:"nonce" validate:"required,min=1"`
	PrevSpinHash      string `json:"prev_spin_hash" validate:"required,len=64"`
	SpinHash          string `json:"spin_hash" validate:"required,len=64"`
	ReelPositions     []int  `json:"reel_positions" validate:"required,len=5"`       // Expected reel positions
	ReelStripConfigID string `json:"reel_strip_config_id" validate:"required,uuid"`  // Config ID to lookup strip length
}

// VerifySpinWithReelResponse includes reel position verification
type VerifySpinWithReelResponse struct {
	Valid                    bool   `json:"valid"`
	SpinHashValid            bool   `json:"spin_hash_valid"`
	ReelPositionsValid       bool   `json:"reel_positions_valid"`
	ExpectedSpinHash         string `json:"expected_spin_hash,omitempty"`
	ExpectedReelPositions    []int  `json:"expected_reel_positions,omitempty"`
	ProvidedReelPositions    []int  `json:"provided_reel_positions,omitempty"`
	ServerSeedHash           string `json:"server_seed_hash,omitempty"`
	Message                  string `json:"message,omitempty"`
}

// VerifyActiveSpinRequest verifies a spin in an active session (server uses stored server_seed)
type VerifyActiveSpinRequest struct {
	ClientSeed        string `json:"client_seed" validate:"required"`
	Nonce             int64  `json:"nonce" validate:"required,min=1"`
	SpinHash          string `json:"spin_hash" validate:"required,len=64"`
	PrevSpinHash      string `json:"prev_spin_hash,omitempty"`          // Optional: if provided, use this instead of server-calculated value
	ReelPositions     []int  `json:"reel_positions,omitempty"`          // Optional: verify reel positions too
	ReelStripConfigID string `json:"reel_strip_config_id,omitempty"`    // Required if reel_positions provided
}

// VerifyActiveSpinResponse includes verification result without revealing server_seed
type VerifyActiveSpinResponse struct {
	Valid                 bool   `json:"valid"`
	SpinHashValid         bool   `json:"spin_hash_valid"`
	ReelPositionsValid    *bool  `json:"reel_positions_valid,omitempty"`    // nil if not checking reel positions
	ExpectedSpinHash      string `json:"expected_spin_hash"`
	ExpectedReelPositions []int  `json:"expected_reel_positions,omitempty"` // Only if reel_positions provided
	ServerSeedHash        string `json:"server_seed_hash"`                  // Hash only, not plaintext
	PrevSpinHash          string `json:"prev_spin_hash"`                    // The prev_spin_hash used for calculation (for debugging)
	Message               string `json:"message,omitempty"`
}
