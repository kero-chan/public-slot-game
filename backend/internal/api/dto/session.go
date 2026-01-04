package dto

import "time"

// StartSessionRequest represents a request to start a game session
type StartSessionRequest struct {
	BetAmount float64 `json:"bet_amount" validate:"required,gt=0"`

	// Dual Commitment Protocol: Client sends theta_commitment BEFORE seeing server_seed
	// This is SHA256(theta_seed) where theta_seed will be revealed on first spin
	ThetaCommitment string `json:"theta_commitment,omitempty"`
}

// SessionResponse represents a game session
type SessionResponse struct {
	ID              string     `json:"id"`
	PlayerID        string     `json:"player_id"`
	BetAmount       float64    `json:"bet_amount"`
	StartingBalance float64    `json:"starting_balance"`
	EndingBalance   *float64   `json:"ending_balance,omitempty"`
	TotalSpins      int        `json:"total_spins"`
	TotalWagered    float64    `json:"total_wagered"`
	TotalWon        float64    `json:"total_won"`
	NetChange       float64    `json:"net_change"`
	CreatedAt       time.Time  `json:"created_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`

	// Provably Fair data (only present if PF is enabled)
	ProvablyFair *SessionProvablyFairData `json:"provably_fair,omitempty"`
}

// SessionProvablyFairData contains provably fair data for a session
// On start: shows server_seed_hash (commitment)
// On end: reveals server_seed and all spin data for verification
type SessionProvablyFairData struct {
	SessionID      string                 `json:"session_id"`
	ServerSeedHash string                 `json:"server_seed_hash"` // Always present
	NonceStart     int64                  `json:"nonce_start,omitempty"`
	ServerSeed     string                 `json:"server_seed,omitempty"` // Only present on end (revealed)
	TotalSpins     int64                  `json:"total_spins,omitempty"`
	Spins          []SpinVerificationData `json:"spins,omitempty"` // Only present on end
}

// SessionHistoryResponse represents paginated session history
type SessionHistoryResponse struct {
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
	Sessions []SessionResponse `json:"sessions"`
}
