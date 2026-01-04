package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============= Admin Auth DTOs =============

// AdminLoginRequest represents an admin login request
type AdminLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AdminLoginResponse represents an admin login response
type AdminLoginResponse struct {
	Token string       `json:"token"`
	Admin *AdminDetail `json:"admin"`
}

// AdminDetail represents admin information in responses
type AdminDetail struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FullName    string     `json:"full_name,omitempty"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	Permissions []string   `json:"permissions,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateAdminRequest represents a request to create a new admin
type CreateAdminRequest struct {
	Username    string   `json:"username" validate:"required,min=3,max=50"`
	Email       string   `json:"email" validate:"required,email"`
	Password    string   `json:"password" validate:"required,min=8"`
	FullName    string   `json:"full_name,omitempty"`
	Role        string   `json:"role" validate:"required,oneof=super_admin admin operator"`
	Permissions []string `json:"permissions,omitempty"`
}

// UpdateAdminRequest represents a request to update an admin
type UpdateAdminRequest struct {
	Email       *string   `json:"email,omitempty" validate:"omitempty,email"`
	FullName    *string   `json:"full_name,omitempty"`
	Role        *string   `json:"role,omitempty" validate:"omitempty,oneof=super_admin admin operator"`
	Permissions *[]string `json:"permissions,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ResetPasswordRequest represents a password reset request (admin action)
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ListAdminsResponse represents a list of admins
type ListAdminsResponse struct {
	Admins []*AdminDetail `json:"admins"`
	Total  int64          `json:"total"`
	Page   int            `json:"page"`
	Limit  int            `json:"limit"`
}

// ============= Reel Strip Config DTOs =============

// CreateReelStripConfigRequest represents a request to create a new reel strip configuration
type CreateReelStripConfigRequest struct {
	Name        string      `json:"name" validate:"required,min=1,max=100"`
	GameMode    string      `json:"game_mode" validate:"required,oneof=base_game free_spins both"`
	Description string      `json:"description,omitempty"`
	TargetRTP   float64     `json:"target_rtp,omitempty" validate:"omitempty,gte=0,lte=100"`
	ReelStripIDs [5]uuid.UUID `json:"reel_strip_ids" validate:"required"`
	CreatedBy   string      `json:"created_by,omitempty"`
	Notes       string      `json:"notes,omitempty"`
}

// UpdateReelStripConfigRequest represents a request to update a reel strip configuration
type UpdateReelStripConfigRequest struct {
	Name        *string     `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string     `json:"description,omitempty"`
	TargetRTP   *float64    `json:"target_rtp,omitempty" validate:"omitempty,gte=0,lte=100"`
	Notes       *string     `json:"notes,omitempty"`
}

// SetDefaultConfigRequest represents a request to set a config as default
type SetDefaultConfigRequest struct {
	ConfigID uuid.UUID `json:"config_id" validate:"required"`
	GameMode string    `json:"game_mode" validate:"required,oneof=base_game free_spins both"`
}

// ReelStripConfigResponse represents a reel strip configuration response
type ReelStripConfigResponse struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	GameMode      string     `json:"game_mode"`
	Description   string     `json:"description,omitempty"`
	Reel0StripID  uuid.UUID  `json:"reel_0_strip_id"`
	Reel1StripID  uuid.UUID  `json:"reel_1_strip_id"`
	Reel2StripID  uuid.UUID  `json:"reel_2_strip_id"`
	Reel3StripID  uuid.UUID  `json:"reel_3_strip_id"`
	Reel4StripID  uuid.UUID  `json:"reel_4_strip_id"`
	TargetRTP     float64    `json:"target_rtp,omitempty"`
	IsActive      bool       `json:"is_active"`
	IsDefault     bool       `json:"is_default"`
	ActivatedAt   *time.Time `json:"activated_at,omitempty"`
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CreatedBy     string     `json:"created_by,omitempty"`
	Notes         string     `json:"notes,omitempty"`
}

// ListReelStripConfigsResponse represents a list of configurations
type ListReelStripConfigsResponse struct {
	Configs []ReelStripConfigResponse `json:"configs"`
	Total   int                       `json:"total"`
}

// ============= Player Reel Strip Assignment DTOs =============

// CreatePlayerAssignmentRequest represents a request to assign a config to a player
type CreatePlayerAssignmentRequest struct {
	PlayerID          uuid.UUID  `json:"player_id" validate:"required"`
	BaseGameConfigID  *uuid.UUID `json:"base_game_config_id,omitempty"`
	FreeSpinsConfigID *uuid.UUID `json:"free_spins_config_id,omitempty"`
	Reason            string     `json:"reason,omitempty" validate:"max=255"`
	AssignedBy        string     `json:"assigned_by,omitempty" validate:"max=100"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
}

// UpdatePlayerAssignmentRequest represents a request to update a player assignment
type UpdatePlayerAssignmentRequest struct {
	BaseGameConfigID  *uuid.UUID `json:"base_game_config_id,omitempty"`
	FreeSpinsConfigID *uuid.UUID `json:"free_spins_config_id,omitempty"`
	Reason            *string    `json:"reason,omitempty" validate:"omitempty,max=255"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	IsActive          *bool      `json:"is_active,omitempty"`
}

// PlayerAssignmentResponse represents a player reel strip assignment response
type PlayerAssignmentResponse struct {
	ID                uuid.UUID  `json:"id"`
	PlayerID          uuid.UUID  `json:"player_id"`
	BaseGameConfigID  *uuid.UUID `json:"base_game_config_id,omitempty"`
	FreeSpinsConfigID *uuid.UUID `json:"free_spins_config_id,omitempty"`
	AssignedAt        time.Time  `json:"assigned_at"`
	AssignedBy        string     `json:"assigned_by,omitempty"`
	Reason            string     `json:"reason,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	IsActive          bool       `json:"is_active"`
}

// ListPlayerAssignmentsResponse represents a list of player assignments
type ListPlayerAssignmentsResponse struct {
	Assignments []PlayerAssignmentResponse `json:"assignments"`
	Total       int                        `json:"total"`
}
