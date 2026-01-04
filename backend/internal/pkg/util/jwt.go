package util

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT claims (used for admin auth)
type Claims struct {
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	GameID   *string `json:"game_id,omitempty"` // nil = cross-game account
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token
// gameID can be nil for cross-game accounts
func GenerateJWT(userID, username string, gameID *uuid.UUID, secret string, expirationHours int) (string, error) {
	var gameIDStr *string
	if gameID != nil {
		s := gameID.String()
		gameIDStr = &s
	}

	claims := &Claims{
		UserID:   userID,
		Username: username,
		GameID:   gameIDStr,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expirationHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT validates a JWT token and returns claims
func ValidateJWT(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
