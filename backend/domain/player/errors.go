package player

import "errors"

var (
	// ErrNotFound is returned when a player is not found
	ErrNotFound = errors.New("player not found")

	// ErrPlayerNotFound is an alias for ErrNotFound
	ErrPlayerNotFound = ErrNotFound

	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUsernameExists is returned when a username already exists
	ErrUsernameExists = errors.New("username already exists")

	// ErrEmailExists is returned when an email already exists
	ErrEmailExists = errors.New("email already exists")

	// ErrPlayerAlreadyExists is returned when player (username or email) already exists
	ErrPlayerAlreadyExists = errors.New("player already exists")

	// ErrInsufficientBalance is returned when balance is insufficient for operation
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	ErrNotFoundOrLockChanged = errors.New("player not found or updated by another session")

	// ErrGameAccessDenied is returned when player tries to access a game they're not bound to
	ErrGameAccessDenied = errors.New("player not authorized for this game")

	// ErrGameNotFound is returned when specified game doesn't exist
	ErrGameNotFound = errors.New("game not found")

	// ErrGameIDRequired is returned when game_id is missing for player registration
	ErrGameIDRequired = errors.New("game_id is required for registration")
)
