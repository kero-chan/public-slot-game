package freespins

import "errors"

var (
	// ErrNotFound is returned when a free spins session is not found
	ErrNotFound = errors.New("free spins session not found")

	// ErrFreeSpinsNotFound is an alias for ErrNotFound
	ErrFreeSpinsNotFound = ErrNotFound

	// ErrNotActive is returned when trying to use an inactive session
	ErrNotActive = errors.New("free spins session not active")

	// ErrFreeSpinsNotActive is an alias for ErrNotActive
	ErrFreeSpinsNotActive = ErrNotActive

	// ErrActiveFreeSpinsExists is returned when player already has active free spins
	ErrActiveFreeSpinsExists = errors.New("player already has an active free spins session")

	// ErrNoRemainingSpins is returned when no spins are remaining
	ErrNoRemainingSpins = errors.New("no remaining free spins")

	// ErrInvalidScatterCount is returned when scatter count is invalid
	ErrInvalidScatterCount = errors.New("invalid scatter count")

	// ErrAlreadyCompleted is returned when session is already completed
	ErrAlreadyCompleted = errors.New("free spins session already completed")

	ErrNotFoundOrLockChanged = errors.New("free spins session not found or updated by another session")
)
