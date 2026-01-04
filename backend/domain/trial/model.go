package trial

import (
	"time"

	"github.com/google/uuid"
)

// Trial mode configuration constants
const (
	// TrialStartingBalance is the initial balance for trial players (same as registered)
	TrialStartingBalance = 100000.0

	// TrialSessionDuration is how long a trial session lasts (2 hours)
	TrialSessionDuration = 2 * time.Hour

	// TrialWinMultiplier increases all wins for trial mode (HUGE RTP)
	// 1.5x means 50% more wins
	TrialWinMultiplier = 1.5

	// TrialTokenPrefix identifies trial session tokens
	TrialTokenPrefix = "trial_"
)

// TrialSession represents an in-memory trial session stored in Redis
type TrialSession struct {
	ID             uuid.UUID  `json:"id"`
	SessionToken   string     `json:"session_token"`   // Format: "trial_<random-hex>"
	Balance        float64    `json:"balance"`         // Current balance
	GameID         *uuid.UUID `json:"game_id"`         // Associated game
	TotalSpins     int        `json:"total_spins"`     // Statistics
	TotalWagered   float64    `json:"total_wagered"`
	TotalWon       float64    `json:"total_won"`
	CreatedAt      time.Time  `json:"created_at"`
	LastActivityAt time.Time  `json:"last_activity_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
}

// TrialGameSession represents a game session for trial players (stored in Redis)
type TrialGameSession struct {
	ID              uuid.UUID  `json:"id"`
	TrialSessionID  uuid.UUID  `json:"trial_session_id"`
	BetAmount       float64    `json:"bet_amount"`
	StartingBalance float64    `json:"starting_balance"`
	TotalSpins      int        `json:"total_spins"`
	TotalWagered    float64    `json:"total_wagered"`
	TotalWon        float64    `json:"total_won"`
	CreatedAt       time.Time  `json:"created_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
}

// TrialFreeSpinsSession represents free spins session for trial players
type TrialFreeSpinsSession struct {
	ID              uuid.UUID `json:"id"`
	TrialSessionID  uuid.UUID `json:"trial_session_id"`
	GameSessionID   uuid.UUID `json:"game_session_id"`
	TotalSpins      int       `json:"total_spins"`
	RemainingSpins  int       `json:"remaining_spins"`
	CompletedSpins  int       `json:"completed_spins"`
	LockedBetAmount float64   `json:"locked_bet_amount"`
	TotalWon        float64   `json:"total_won"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

// NewTrialSession creates a new trial session
func NewTrialSession(gameID *uuid.UUID, sessionToken string) *TrialSession {
	now := time.Now().UTC()
	return &TrialSession{
		ID:             uuid.New(),
		SessionToken:   sessionToken,
		Balance:        TrialStartingBalance,
		GameID:         gameID,
		TotalSpins:     0,
		TotalWagered:   0,
		TotalWon:       0,
		CreatedAt:      now,
		LastActivityAt: now,
		ExpiresAt:      now.Add(TrialSessionDuration),
	}
}

// NewTrialGameSession creates a new trial game session
func NewTrialGameSession(trialSessionID uuid.UUID, betAmount, startingBalance float64) *TrialGameSession {
	return &TrialGameSession{
		ID:              uuid.New(),
		TrialSessionID:  trialSessionID,
		BetAmount:       betAmount,
		StartingBalance: startingBalance,
		TotalSpins:      0,
		TotalWagered:    0,
		TotalWon:        0,
		CreatedAt:       time.Now().UTC(),
	}
}

// NewTrialFreeSpinsSession creates a new trial free spins session
func NewTrialFreeSpinsSession(trialSessionID, gameSessionID uuid.UUID, spinsAwarded int, betAmount float64) *TrialFreeSpinsSession {
	return &TrialFreeSpinsSession{
		ID:              uuid.New(),
		TrialSessionID:  trialSessionID,
		GameSessionID:   gameSessionID,
		TotalSpins:      spinsAwarded,
		RemainingSpins:  spinsAwarded,
		CompletedSpins:  0,
		LockedBetAmount: betAmount,
		TotalWon:        0,
		IsActive:        true,
		CreatedAt:       time.Now().UTC(),
	}
}

// IsExpired checks if the trial session has expired
func (s *TrialSession) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// UpdateActivity updates the last activity timestamp
func (s *TrialSession) UpdateActivity() {
	s.LastActivityAt = time.Now().UTC()
}

// DeductBalance deducts bet amount from balance
func (s *TrialSession) DeductBalance(amount float64) bool {
	if s.Balance < amount {
		return false
	}
	s.Balance -= amount
	s.TotalWagered += amount
	return true
}

// CreditWin adds win amount to balance (with trial multiplier already applied)
func (s *TrialSession) CreditWin(amount float64) {
	s.Balance += amount
	s.TotalWon += amount
}

// IncrementSpins increments the spin counter
func (s *TrialSession) IncrementSpins() {
	s.TotalSpins++
}

// HasActiveGameSession checks if free spins session is active
func (fs *TrialFreeSpinsSession) HasRemainingSpins() bool {
	return fs.RemainingSpins > 0 && fs.IsActive
}

// DecrementRemainingSpins decrements remaining spins
func (fs *TrialFreeSpinsSession) DecrementRemainingSpins() {
	if fs.RemainingSpins > 0 {
		fs.RemainingSpins--
		fs.CompletedSpins++
	}
}

// AddRetriggerSpins adds more spins from retrigger
func (fs *TrialFreeSpinsSession) AddRetriggerSpins(spins int) {
	fs.TotalSpins += spins
	fs.RemainingSpins += spins
}

// Complete marks the free spins session as complete
func (fs *TrialFreeSpinsSession) Complete() {
	fs.IsActive = false
}
