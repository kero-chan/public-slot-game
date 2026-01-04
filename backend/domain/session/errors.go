package session

import "errors"

var (
	// ErrNotFound is returned when a session is not found
	ErrNotFound = errors.New("session not found")

	// ErrSessionNotFound is an alias for ErrNotFound
	ErrSessionNotFound = ErrNotFound

	// ErrAlreadyEnded is returned when trying to use an ended session
	ErrAlreadyEnded = errors.New("session already ended")

	// ErrSessionAlreadyEnded is an alias for ErrAlreadyEnded
	ErrSessionAlreadyEnded = ErrAlreadyEnded

	// ErrActiveSessionExists is returned when trying to start a new session while one is active
	ErrActiveSessionExists = errors.New("active session already exists")

	// ErrInvalidBetAmount is returned when bet amount is invalid
	ErrInvalidBetAmount = errors.New("invalid bet amount")
)
