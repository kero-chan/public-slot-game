package reelstrip

import "errors"

var (
	// ReelStrip errors
	ErrReelStripNotFound  = errors.New("reel strip not found")
	ErrInvalidGameMode    = errors.New("invalid game mode")
	ErrInvalidReelNumber  = errors.New("reel number must be between 0 and 4")
	ErrIncompleteSet      = errors.New("incomplete reel strip set")
	ErrInvalidStripLength = errors.New("invalid strip length")
	ErrChecksumMismatch   = errors.New("checksum mismatch")
	ErrNoActiveStrips     = errors.New("no active reel strips available")

	// ReelStripConfig errors
	ErrConfigNotFound  = errors.New("reel strip config not found")
	ErrNoDefaultConfig = errors.New("no default config found for game mode")
	ErrInvalidConfig   = errors.New("invalid reel strip config")

	// PlayerReelStripAssignment errors
	ErrAssignmentNotFound = errors.New("player reel strip assignment not found")
	ErrInvalidAssignment  = errors.New("invalid player assignment")
)
