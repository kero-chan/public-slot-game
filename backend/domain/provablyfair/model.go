package provablyfair

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PFSession represents a Provably Fair gaming session
// One session = one server_seed = N spins linked by hash chain
// Client seed is provided per-spin (not per-session) to ensure server cannot predict outcomes
//
// Dual Commitment Protocol:
// 1. Client generates theta_seed, sends theta_commitment = SHA256(theta_seed) BEFORE session starts
// 2. Server generates server_seed ONLY AFTER receiving theta_commitment
// 3. On first spin, client reveals theta_seed and server verifies it matches theta_commitment
// 4. This ensures neither party can bias the random outcome
type PFSession struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PlayerID            uuid.UUID  `gorm:"type:uuid;not null;index"`
	GameSessionID       uuid.UUID  `gorm:"type:uuid;not null;index"`                         // Links to game_sessions
	ServerSeedHash      string     `gorm:"type:varchar(64);not null"`                        // SHA256(server_seed) - public from start
	EncryptedServerSeed string     `gorm:"type:text;not null"`                               // AES-256-GCM encrypted server seed for recovery
	NonceStart          int64      `gorm:"not null;default:1"`                               // Starting nonce (always 1)
	LastNonce           int64      `gorm:"not null;default:0"`                               // Last used nonce
	LastSpinHash        string     `gorm:"type:varchar(64)"`                                 // Hash of the last spin
	Status              string     `gorm:"type:varchar(20);not null;default:'active';index"` // active, ended
	CreatedAt           time.Time  `gorm:"not null;default:now()"`
	EndedAt             *time.Time `gorm:"index"`
	// Dual Commitment Protocol fields
	ThetaCommitment string `gorm:"type:varchar(64)"` // SHA256(theta_seed) - client's commitment sent BEFORE seeing server_seed
	ThetaSeed       string `gorm:"type:varchar(64)"` // Client's session seed - revealed on first spin
	ThetaVerified   bool   `gorm:"not null;default:false"` // True after theta_seed is verified on first spin
}

// TableName specifies the table name for GORM
func (PFSession) TableName() string {
	return "pf_sessions"
}

// Session status constants
const (
	SessionStatusActive = "active"
	SessionStatusEnded  = "ended"
)

// SpinLog represents an append-only audit log for each spin
// CRITICAL: This table is append-only - NO UPDATE, NO DELETE
// Client seed is stored per-spin to ensure server cannot predict outcomes
type SpinLog struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PFSessionID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	SpinID            uuid.UUID  `gorm:"type:uuid;not null;index"`  // Links to spins table
	SpinIndex         int64      `gorm:"not null;index"`            // Sequential index within session
	Nonce             int64      `gorm:"not null"`
	ClientSeed        string     `gorm:"type:varchar(64);not null"` // Client-provided seed for this spin
	SpinHash          string     `gorm:"type:varchar(64);not null"` // SHA256(prev_spin_hash + server_seed + client_seed + nonce)
	PrevSpinHash      string     `gorm:"type:varchar(64);not null"` // Previous spin hash (server_seed_hash for first spin)
	ReelPositions     IntSlice   `gorm:"type:jsonb;not null"`       // Array of 5 reel positions from RNG
	ReelStripConfigID *uuid.UUID `gorm:"type:uuid;index"`           // Reference to reel_strip_configs for verification
	GameMode          *string    `gorm:"type:varchar(32)"`          // Game mode: nil for normal, or bonus_spin_trigger etc.
	IsFreeSpin        bool       `gorm:"not null;default:false"`    // Whether this was a free spin
	CreatedAt         time.Time  `gorm:"not null;default:now();index"`
}

// TableName specifies the table name for GORM
func (SpinLog) TableName() string {
	return "spin_logs"
}

// SessionAudit stores revealed server_seed after session ends
// Used for client verification
type SessionAudit struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PFSessionID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	ServerSeedPlaintext string    `gorm:"type:varchar(64);not null"` // Revealed after session ends
	ServerSeedHash      string    `gorm:"type:varchar(64);not null"` // Must match PFSession.ServerSeedHash
	TotalSpins          int64     `gorm:"not null"`
	RevealedAt          time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for GORM
func (SessionAudit) TableName() string {
	return "session_audits"
}

// PFSessionState represents the runtime state stored in Redis
// This is the source of truth during active sessions
// Client seed is provided per-spin, not stored in session state
type PFSessionState struct {
	SessionID      uuid.UUID `json:"session_id"`
	PlayerID       uuid.UUID `json:"player_id"`
	GameSessionID  uuid.UUID `json:"game_session_id"`
	ServerSeed     string    `json:"server_seed"`      // Plaintext - ONLY in Redis, never in DB until reveal
	ServerSeedHash string    `json:"server_seed_hash"` // SHA256(server_seed) - commitment shown to player
	LastSpinHash   string    `json:"last_spin_hash"`   // Hash of last spin, or server_seed_hash for first spin
	Nonce          int64     `json:"nonce"`
	Status         string    `json:"status"`
	UpdatedAt      time.Time `json:"updated_at"`
	// Dual Commitment Protocol fields
	ThetaCommitment string `json:"theta_commitment,omitempty"` // SHA256(theta_seed) - client's commitment
	ThetaSeed       string `json:"theta_seed,omitempty"`       // Client's session seed - revealed on first spin
	ThetaVerified   bool   `json:"theta_verified"`             // True after theta_seed is verified
}

// MarshalBinary implements encoding.BinaryMarshaler for Redis
func (s *PFSessionState) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Redis
func (s *PFSessionState) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

// SpinHashData contains all data needed to calculate a spin hash
type SpinHashData struct {
	PrevSpinHash string `json:"prev_spin_hash"`
	ServerSeed   string `json:"server_seed"`
	ClientSeed   string `json:"client_seed"`
	Nonce        int64  `json:"nonce"`
}

// VerificationData contains all data needed for client verification
// Client seed is stored per-spin in SpinVerification
type VerificationData struct {
	SessionID      uuid.UUID          `json:"session_id"`
	ServerSeed     string             `json:"server_seed"`
	ServerSeedHash string             `json:"server_seed_hash"`
	Spins          []SpinVerification `json:"spins"`
}

// SpinVerification contains data for verifying a single spin
type SpinVerification struct {
	SpinIndex         int64      `json:"spin_index"`
	Nonce             int64      `json:"nonce"`
	ClientSeed        string     `json:"client_seed"`          // Client-provided seed for this spin
	SpinHash          string     `json:"spin_hash"`
	PrevSpinHash      string     `json:"prev_spin_hash"`
	ReelPositions     []int      `json:"reel_positions"`       // Array of 5 reel positions
	ReelStripConfigID *uuid.UUID `json:"reel_strip_config_id"` // Which reel strip config was used
	GameMode          *string    `json:"game_mode"`            // Game mode if any
	IsFreeSpin        bool       `json:"is_free_spin"`
}

// StringSlice is a helper type for storing string slices in JSONB
type StringSlice []string

// Scan implements the sql.Scanner interface
func (s *StringSlice) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value implements the driver.Valuer interface
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// IntSlice is a helper type for storing int slices in JSONB (for reel positions)
type IntSlice []int

// Scan implements the sql.Scanner interface
func (s *IntSlice) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value implements the driver.Valuer interface
func (s IntSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}
