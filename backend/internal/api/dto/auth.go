package dto

import "time"

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string  `json:"username" validate:"required,min=3,max=50"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=8"`
	GameID   *string `json:"game_id,omitempty" validate:"omitempty,uuid"` // Optional game ID
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username    string  `json:"username" validate:"required"`
	Password    string  `json:"password" validate:"required"`
	GameID      *string `json:"game_id,omitempty" validate:"omitempty,uuid"` // Optional game ID
	ForceLogout bool    `json:"force_logout,omitempty"`                      // Force logout existing session
	DeviceInfo  *string `json:"device_info,omitempty"`                       // Device information (browser, OS, etc.)
}

// AuthResponse represents an authentication response (login)
type AuthResponse struct {
	SessionToken string        `json:"session_token"` // Session token for authentication
	ExpiresAt    int64         `json:"expires_at"`    // Unix timestamp when session expires
	Player       PlayerProfile `json:"player"`
}

// RegisterResponse represents a registration response
type RegisterResponse struct {
	Message string        `json:"message"`
	Player  PlayerProfile `json:"player"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	SessionToken string `json:"session_token,omitempty"` // If not provided, uses token from header
}

// PlayerProfile represents a player's profile in responses
type PlayerProfile struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Balance      float64    `json:"balance"`
	GameID       *string    `json:"game_id,omitempty"` // Game ID (nil = cross-game account)
	TotalSpins   int        `json:"total_spins"`
	TotalWagered float64    `json:"total_wagered"`
	TotalWon     float64    `json:"total_won"`
	IsActive     bool       `json:"is_active"`
	IsVerified   bool       `json:"is_verified"`
	CreatedAt    time.Time  `json:"created_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}


// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
