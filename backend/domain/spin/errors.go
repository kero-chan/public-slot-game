package spin

import "errors"

var (
	// ErrNotFound is returned when a spin is not found
	ErrNotFound = errors.New("spin not found")

	// ErrSpinNotFound is an alias for ErrNotFound
	ErrSpinNotFound = ErrNotFound

	// ErrInvalidSession is returned when session is invalid
	ErrInvalidSession = errors.New("invalid session")

	// ErrInsufficientBalance is returned when player has insufficient balance
	ErrInsufficientBalance = errors.New("insufficient balance for bet")

	// ErrInvalidBetAmount is returned when bet amount is invalid
	ErrInvalidBetAmount = errors.New("invalid bet amount")

	// ErrGameEngineFailure is returned when game engine fails
	ErrGameEngineFailure = errors.New("game engine failure")
)
