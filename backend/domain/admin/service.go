package admin

import (
	"context"

	"github.com/google/uuid"
)

// Service defines the business logic interface for admin management
type Service interface {
	// Authentication
	Login(ctx context.Context, username, password, ip string) (*Admin, string, error)
	ValidateToken(ctx context.Context, token string) (*Admin, error)

	// Admin Management
	CreateAdmin(ctx context.Context, req CreateAdminRequest, createdBy uuid.UUID) (*Admin, error)
	GetAdmin(ctx context.Context, id uuid.UUID) (*Admin, error)
	UpdateAdmin(ctx context.Context, id uuid.UUID, req UpdateAdminRequest, updatedBy uuid.UUID) (*Admin, error)
	DeleteAdmin(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	ListAdmins(ctx context.Context, filters ListFilters) ([]*Admin, int64, error)

	// Password Management
	ChangePassword(ctx context.Context, adminID uuid.UUID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, adminID uuid.UUID, newPassword string, resetBy uuid.UUID) error

	// Status Management
	ActivateAdmin(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error
	DeactivateAdmin(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error
	SuspendAdmin(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error

	// Player Management
	CreatePlayer(ctx context.Context, req CreatePlayerRequest, createdBy uuid.UUID) (interface{}, error)
	GetPlayer(ctx context.Context, id uuid.UUID) (interface{}, error)
	ListPlayers(ctx context.Context, filters PlayerListFilters) (interface{}, int64, error)
	ActivatePlayer(ctx context.Context, playerID uuid.UUID, updatedBy uuid.UUID) error
	DeactivatePlayer(ctx context.Context, playerID uuid.UUID, updatedBy uuid.UUID) error
	ForceLogoutPlayer(ctx context.Context, playerID uuid.UUID, adminID uuid.UUID) error
}

// CreateAdminRequest represents a request to create an admin
type CreateAdminRequest struct {
	Username    string
	Email       string
	Password    string
	FullName    string
	Role        AdminRole
	Permissions []string
}

// UpdateAdminRequest represents a request to update an admin
type UpdateAdminRequest struct {
	Email       *string
	FullName    *string
	Role        *AdminRole
	Permissions *[]string
}

// CreatePlayerRequest represents a request to create a player
type CreatePlayerRequest struct {
	Username string
	Email    string
	Password string
	Balance  float64
	GameID   *uuid.UUID
}

// PlayerListFilters represents filters for listing players
type PlayerListFilters struct {
	Username string
	Email    string
	GameID   *uuid.UUID
	IsActive *bool
	Page     int
	Limit    int
	SortBy   string
	SortDesc bool
}
