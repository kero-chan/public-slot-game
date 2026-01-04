package spin

import (
	"context"

	"github.com/google/uuid"
)

// Service defines the interface for spin business logic
type Service interface {
	// GenerateInitialGrid generates a demo grid for initial display
	// This ensures frontend has zero RNG - all symbol generation is backend-controlled
	GenerateInitialGrid(ctx context.Context) (Grid, error)

	// ExecuteSpin executes a regular spin
	// gameMode is optional: free_spin_trigger, wild_spin_trigger, bonus_spin_trigger
	// clientSeed is optional: for provably fair sessions, client provides their own seed per-spin
	// thetaSeed is optional: for Dual Commitment Protocol, revealed on first spin
	ExecuteSpin(ctx context.Context, playerID, sessionID uuid.UUID, betAmount float64, gameMode, clientSeed, thetaSeed string) (*SpinResult, error)

	// GetSpinDetails retrieves details of a specific spin
	GetSpinDetails(ctx context.Context, spinID uuid.UUID) (*Spin, error)

	// GetSpinHistory retrieves spin history for a player
	GetSpinHistory(ctx context.Context, playerID uuid.UUID, page, limit int) (*SpinHistoryResult, error)
}

// GridPosition represents a position on the grid (reel, row)
type GridPosition struct {
	Reel int `json:"reel"`
	Row  int `json:"row"`
}

// SpinResult represents the result of a spin execution
type SpinResult struct {
	SpinID                  uuid.UUID `json:"spin_id"`
	SessionID               uuid.UUID `json:"session_id"`
	BetAmount               float64   `json:"bet_amount"`
	BalanceBefore           float64   `json:"balance_before"`
	BalanceAfterBet         float64   `json:"balance_after_bet"`
	NewBalance              float64   `json:"new_balance"`
	Grid                    Grid      `json:"grid"`
	Cascades                Cascades  `json:"cascades"`
	SpinTotalWin            float64   `json:"spin_total_win"`
	ScatterCount            int       `json:"scatter_count"`
	IsFreeSpin              bool      `json:"is_free_spin"`
	FreeSpinsTriggered      bool      `json:"free_spins_triggered"`
	FreeSpinsRetriggered    bool      `json:"free_spins_retriggered,omitempty"`
	FreeSpinsAdditional     int       `json:"free_spins_additional,omitempty"`
	FreeSpinsSessionID      string    `json:"free_spins_session_id,omitempty"`
	FreeSpinsRemainingSpins int       `json:"free_spins_remaining_spins,omitempty"`
	FreeSessionTotalWin     float64   `json:"free_session_total_win,omitempty"`
	GameMode                string    `json:"game_mode,omitempty"`      // Game mode used: bonus_spin_trigger (guaranteed free spins)
	GameModeCost            float64   `json:"game_mode_cost,omitempty"` // Cost paid for game mode (1000)
	Timestamp               string    `json:"timestamp"`

	// Provably Fair data (only present if PF session is active)
	ProvablyFair *SpinProvablyFairData `json:"provably_fair,omitempty"`
}

// SpinProvablyFairData contains provably fair data for a spin
type SpinProvablyFairData struct {
	SpinIndex    int64  `json:"spin_index"`
	Nonce        int64  `json:"nonce"`
	SpinHash     string `json:"spin_hash"`
	PrevSpinHash string `json:"prev_spin_hash"`
}

// SpinHistoryResult represents paginated spin history
type SpinHistoryResult struct {
	Page  int     `json:"page"`
	Limit int     `json:"limit"`
	Total int64   `json:"total"`
	Spins []*Spin `json:"spins"`
}
