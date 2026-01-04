package dto

// FreeSpinsStatusResponse represents the status of a free spins session
type FreeSpinsStatusResponse struct {
	Active             bool    `json:"active"`
	FreeSpinsSessionID string  `json:"free_spins_session_id,omitempty"`
	SessionID          string  `json:"session_id,omitempty"` // Game session ID for provably fair recovery
	TotalSpinsAwarded  int     `json:"total_spins_awarded"`
	SpinsCompleted     int     `json:"spins_completed"`
	RemainingSpins     int     `json:"remaining_spins"`
	LockedBetAmount    float64 `json:"locked_bet_amount"`
	TotalWon           float64 `json:"total_won"`
}

// ExecuteFreeSpinRequest represents a free spin execution request
type ExecuteFreeSpinRequest struct {
	FreeSpinsSessionID string `json:"free_spins_session_id" validate:"required,uuid"`
	ClientSeed         string `json:"client_seed,omitempty"` // Optional: for provably fair, client provides per-spin seed
}
