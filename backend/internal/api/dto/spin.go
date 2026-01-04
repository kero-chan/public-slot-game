package dto

import "time"

// InitialGridResponse represents an initial/demo grid for display
type InitialGridResponse struct {
	Grid [][]int `json:"grid"`
}

// ExecuteSpinRequest represents a spin execution request
type ExecuteSpinRequest struct {
	SessionID  string  `json:"session_id" validate:"omitempty,uuid"`
	BetAmount  float64 `json:"bet_amount" validate:"required,gt=0"`
	GameMode   string  `json:"game_mode,omitempty"`   // Optional: bonus_spin_trigger (guaranteed free spins)
	ClientSeed string  `json:"client_seed,omitempty"` // Optional: for provably fair, client provides per-spin seed
	// Dual Commitment Protocol: theta_seed is revealed on first spin
	ThetaSeed string `json:"theta_seed,omitempty"` // Required on first spin if theta_commitment was provided
}

// SpinProvablyFairData contains provably fair data for a spin response
type SpinProvablyFairData struct {
	SpinHash     string `json:"spin_hash"`      // Current spin's hash (for client tracking)
	PrevSpinHash string `json:"prev_spin_hash"` // Previous spin's hash (or server_seed_hash for first spin)
	Nonce        int64  `json:"nonce"`          // Spin nonce in session
}

// SpinResponse represents a spin result
type SpinResponse struct {
	SpinID                  string                `json:"spin_id"`
	SessionID               string                `json:"session_id"`
	BetAmount               float64               `json:"bet_amount"`
	BalanceBefore           float64               `json:"balance_before"`
	BalanceAfterBet         float64               `json:"balance_after_bet"`
	NewBalance              float64               `json:"new_balance"`
	Grid                    [][]int               `json:"grid"`
	Cascades                []CascadeInfo         `json:"cascades"`
	SpinTotalWin            float64               `json:"spin_total_win"`
	ScatterCount            int                   `json:"scatter_count"`
	IsFreeSpin              bool                  `json:"is_free_spin"`
	FreeSpinsTriggered      bool                  `json:"free_spins_triggered"`
	FreeSpinsRetriggered    bool                  `json:"free_spins_retriggered"`
	FreeSpinsAdditional     int                   `json:"free_spins_additional,omitempty"`
	FreeSpinsSessionID      string                `json:"free_spins_session_id,omitempty"`
	FreeSpinsRemainingSpins int                   `json:"free_spins_remaining_spins"`
	FreeSessionTotalWin     float64               `json:"free_session_total_win"`
	GameMode                string                `json:"game_mode,omitempty"`      // Game mode used: bonus_spin_trigger (guaranteed free spins)
	GameModeCost            float64               `json:"game_mode_cost,omitempty"` // Cost paid for game mode (1000)
	Timestamp               string                `json:"timestamp"`
	ProvablyFair            *SpinProvablyFairData `json:"provably_fair,omitempty"` // Present if PF session is active
}

// CascadeInfo represents cascade information
type CascadeInfo struct {
	CascadeNumber   int       `json:"cascade_number"`
	GridAfter       [][]int   `json:"grid_after"` // Grid state after this cascade
	Multiplier      int       `json:"multiplier"`
	Wins            []WinInfo `json:"wins"`
	TotalCascadeWin float64   `json:"total_cascade_win"`
	WinningTileKind string    `json:"winning_tile_kind,omitempty"` // Highest priority winning symbol (fa > zhong > bai > bawan)
}

// Position represents a grid position [reel, row]
type Position struct {
	Reel          int  `json:"reel"`
	Row           int  `json:"row"`
	IsGoldToWild  bool `json:"is_gold_to_wild,omitempty"` // True if this gold tile transforms to wild
}

// WinInfo represents a win in a cascade
type WinInfo struct {
	Symbol       int        `json:"symbol"`        // Symbol ID (same as grid values)
	Count        int        `json:"count"`
	Ways         int        `json:"ways"`
	Payout       float64    `json:"payout"`
	WinAmount    float64    `json:"win_amount"`
	Positions    []Position `json:"positions"`     // Grid positions that form this win
	WinIntensity string     `json:"win_intensity"` // Visual intensity: small, medium, big, mega
}

// SpinHistoryResponse represents paginated spin history
type SpinHistoryResponse struct {
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Total int64         `json:"total"`
	Spins []SpinSummary `json:"spins"`
}

// SpinSummary represents a summary of a spin for history
type SpinSummary struct {
	SpinID             string    `json:"spin_id"`
	SessionID          string    `json:"session_id"`
	BetAmount          float64   `json:"bet_amount"`
	TotalWin           float64   `json:"total_win"`
	ScatterCount       int       `json:"scatter_count"`
	IsFreeSpin         bool      `json:"is_free_spin"`
	FreeSpinsTriggered bool      `json:"free_spins_triggered"`
	CreatedAt          time.Time `json:"created_at"`
}
