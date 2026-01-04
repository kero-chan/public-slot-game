package admin

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for admin data persistence
type Repository interface {
	// Create creates a new admin
	Create(ctx context.Context, admin *Admin) error

	// GetByID retrieves an admin by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Admin, error)

	// GetByUsername retrieves an admin by username
	GetByUsername(ctx context.Context, username string) (*Admin, error)

	// GetByEmail retrieves an admin by email
	GetByEmail(ctx context.Context, email string) (*Admin, error)

	// Update updates an admin
	Update(ctx context.Context, admin *Admin) error

	// Delete soft deletes an admin
	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves all admins with optional filters
	List(ctx context.Context, filters ListFilters) ([]*Admin, int64, error)

	// UpdateLastLogin updates the last login timestamp and IP
	UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error

	// IncrementFailedAttempts increments failed login attempts
	IncrementFailedAttempts(ctx context.Context, id uuid.UUID) error

	// ResetFailedAttempts resets failed login attempts
	ResetFailedAttempts(ctx context.Context, id uuid.UUID) error
}

// ListFilters represents filters for listing admins
type ListFilters struct {
	Role     *AdminRole
	Status   *AdminStatus
	Page     int
	PageSize int
}
