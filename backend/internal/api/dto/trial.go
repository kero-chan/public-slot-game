package dto

// StartTrialRequest represents the request to start a trial session
type StartTrialRequest struct {
	GameID string `json:"game_id,omitempty"` // Optional game ID
}

// TrialProfile represents trial player profile info
type TrialProfile struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	Balance      float64 `json:"balance"`
	TotalSpins   int     `json:"total_spins"`
	TotalWagered float64 `json:"total_wagered"`
	TotalWon     float64 `json:"total_won"`
	IsTrial      bool    `json:"is_trial"`
}

// TrialResponse represents the response when starting a trial session
type TrialResponse struct {
	SessionToken string       `json:"session_token"`
	ExpiresAt    int64        `json:"expires_at"` // Unix timestamp
	Player       TrialProfile `json:"player"`
}

// TrialBalanceResponse represents the response for balance query
type TrialBalanceResponse struct {
	Balance float64 `json:"balance"`
}
