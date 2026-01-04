package freespins

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a free spins session (lightweight version for game engine)
type Session struct {
	ID                uuid.UUID
	PlayerID          uuid.UUID
	ReelStripConfigID *uuid.UUID
	TotalSpinsAwarded int
	SpinsCompleted    int
	RemainingSpins    int
	LockedBetAmount   float64
	TotalWon          float64
	IsActive          bool
	CreatedAt         time.Time
}

// NewSession creates a new free spins session
func NewSession(playerID uuid.UUID, scatterCount int, betAmount float64, reelStripConfigID *uuid.UUID) *Session {
	spinsAwarded := CalculateFreeSpinsAward(scatterCount)

	return &Session{
		ID:                uuid.New(),
		PlayerID:          playerID,
		ReelStripConfigID: reelStripConfigID,
		TotalSpinsAwarded: spinsAwarded,
		SpinsCompleted:    0,
		RemainingSpins:    spinsAwarded,
		LockedBetAmount:   betAmount,
		TotalWon:          0.0,
		IsActive:          true,
		CreatedAt:         time.Now().UTC(),
	}
}

// ExecuteSpin executes one free spin
func (s *Session) ExecuteSpin(winAmount float64) {
	s.SpinsCompleted++
	s.RemainingSpins--
	s.TotalWon += winAmount

	// Check if session is complete
	if s.RemainingSpins <= 0 {
		s.IsActive = false
	}
}

// AddRetriggerSpins adds spins from a retrigger
func (s *Session) AddRetriggerSpins(additionalSpins int) {
	s.TotalSpinsAwarded += additionalSpins
	s.RemainingSpins += additionalSpins
}

// IsComplete checks if the session is complete
func (s *Session) IsComplete() bool {
	return s.RemainingSpins <= 0 || !s.IsActive
}

// GetProgress returns the progress percentage (0-100)
func (s *Session) GetProgress() float64 {
	if s.TotalSpinsAwarded == 0 {
		return 100.0
	}
	return float64(s.SpinsCompleted) / float64(s.TotalSpinsAwarded) * 100.0
}
