package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents an application error code
type ErrorCode string

// Error codes
const (
	ErrInvalidRequest      ErrorCode = "INVALID_REQUEST"
	ErrUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrForbidden           ErrorCode = "FORBIDDEN"
	ErrNotFound            ErrorCode = "NOT_FOUND"
	ErrInsufficientBalance ErrorCode = "INSUFFICIENT_BALANCE"
	ErrInvalidBetAmount    ErrorCode = "INVALID_BET_AMOUNT"
	ErrSessionNotFound     ErrorCode = "SESSION_NOT_FOUND"
	ErrSessionExpired      ErrorCode = "SESSION_EXPIRED"
	ErrFreeSpinsNotActive  ErrorCode = "FREE_SPINS_NOT_ACTIVE"
	ErrInternalError       ErrorCode = "INTERNAL_ERROR"
	ErrRNGError            ErrorCode = "RNG_ERROR"
	ErrDatabaseError       ErrorCode = "DATABASE_ERROR"
	ErrRateLimitExceeded   ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrInvalidCredentials  ErrorCode = "INVALID_CREDENTIALS"
	ErrUserAlreadyExists   ErrorCode = "USER_ALREADY_EXISTS"
	ErrInvalidToken        ErrorCode = "INVALID_TOKEN"
	ErrTokenExpired        ErrorCode = "TOKEN_EXPIRED"
)

// HTTPError represents an HTTP error with code and details
type HTTPError struct {
	StatusCode int
	Code       ErrorCode
	Message    string
	Details    interface{}
	Err        error
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// New creates a new HTTPError
func New(statusCode int, code ErrorCode, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// NewWithDetails creates a new HTTPError with details
func NewWithDetails(statusCode int, code ErrorCode, message string, details interface{}) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

// Wrap wraps an existing error
func Wrap(statusCode int, code ErrorCode, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}

// Common error constructors

func BadRequest(message string) *HTTPError {
	return New(http.StatusBadRequest, ErrInvalidRequest, message)
}

func Unauthorized(message string) *HTTPError {
	return New(http.StatusUnauthorized, ErrUnauthorized, message)
}

func Forbidden(message string) *HTTPError {
	return New(http.StatusForbidden, ErrForbidden, message)
}

func NotFound(message string) *HTTPError {
	return New(http.StatusNotFound, ErrNotFound, message)
}

func InternalError(message string, err error) *HTTPError {
	return Wrap(http.StatusInternalServerError, ErrInternalError, message, err)
}

func InsufficientBalance(required, available float64) *HTTPError {
	return NewWithDetails(
		http.StatusBadRequest,
		ErrInsufficientBalance,
		"Insufficient balance for bet amount",
		map[string]interface{}{
			"required":  required,
			"available": available,
		},
	)
}

func InvalidBetAmount(message string) *HTTPError {
	return New(http.StatusBadRequest, ErrInvalidBetAmount, message)
}

func SessionNotFound(sessionID string) *HTTPError {
	return NewWithDetails(
		http.StatusNotFound,
		ErrSessionNotFound,
		"Game session not found",
		map[string]string{"session_id": sessionID},
	)
}

func RateLimitExceeded(retryAfter int) *HTTPError {
	return NewWithDetails(
		http.StatusTooManyRequests,
		ErrRateLimitExceeded,
		"Too many requests. Please try again later.",
		map[string]int{"retry_after": retryAfter},
	)
}

func TooManyRequests(message string) *HTTPError {
	return New(http.StatusTooManyRequests, ErrRateLimitExceeded, message)
}

func ServiceUnavailable(message string) *HTTPError {
	return New(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}
