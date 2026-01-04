package dto

// GetBalanceResponse represents the balance response
type GetBalanceResponse struct {
	Balance float64 `json:"balance"`
}

// UpdateBalanceRequest represents a balance update request (admin only)
type UpdateBalanceRequest struct {
	NewBalance float64 `json:"new_balance" validate:"required,min=0"`
}
