package admin

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// StringArray is a custom type for handling JSONB string arrays in PostgreSQL
type StringArray []string

// Scan implements the sql.Scanner interface for StringArray
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan StringArray: %v", value)
	}

	return json.Unmarshal(bytes, s)
}

// Value implements the driver.Valuer interface for StringArray
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

// AdminRole represents the role of an admin user
type AdminRole string

const (
	RoleSuperAdmin AdminRole = "super_admin" // Full access to everything
	RoleAdmin      AdminRole = "admin"       // Most administrative access
	RoleOperator   AdminRole = "operator"    // Limited operational access
)

// AdminStatus represents the status of an admin account
type AdminStatus string

const (
	StatusActive    AdminStatus = "active"
	StatusInactive  AdminStatus = "inactive"
	StatusSuspended AdminStatus = "suspended"
)

// Admin represents an administrative user
type Admin struct {
	ID           uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username     string      `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email        string      `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string      `gorm:"type:varchar(255);not null" json:"-"` // Never expose in JSON

	// Admin details
	FullName    string      `gorm:"type:varchar(100)" json:"full_name,omitempty"`
	Role        AdminRole   `gorm:"type:admin_role;not null;default:operator" json:"role"`
	Status      AdminStatus `gorm:"type:admin_status;not null;default:active" json:"status"`

	// Permissions - stored as JSONB array in database
	Permissions StringArray `gorm:"type:jsonb;default:'[]'" json:"permissions,omitempty"`

	// Security
	LastLoginAt         *time.Time `gorm:"type:timestamp with time zone" json:"last_login_at,omitempty"`
	LastLoginIP         string     `gorm:"type:varchar(45)" json:"last_login_ip,omitempty"`
	FailedLoginAttempts int        `gorm:"default:0" json:"-"`
	LockedUntil         *time.Time `gorm:"type:timestamp with time zone" json:"locked_until,omitempty"`

	// Audit
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	UpdatedBy *uuid.UUID `gorm:"type:uuid" json:"updated_by,omitempty"`
	DeletedAt *time.Time `gorm:"type:timestamp with time zone" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for GORM
func (Admin) TableName() string {
	return "admins"
}

// IsActive checks if the admin account is active
func (a *Admin) IsActive() bool {
	return a.Status == StatusActive && a.DeletedAt == nil
}

// IsLocked checks if the admin account is currently locked
func (a *Admin) IsLocked() bool {
	if a.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*a.LockedUntil)
}

// CanAccess checks if admin has a specific permission
func (a *Admin) CanAccess(permission string) bool {
	// Super admins have access to everything
	if a.Role == RoleSuperAdmin {
		return true
	}

	// Check specific permissions
	for _, p := range a.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}

	return false
}

// HasRole checks if admin has a specific role or higher
func (a *Admin) HasRole(role AdminRole) bool {
	if a.Role == RoleSuperAdmin {
		return true
	}
	if a.Role == RoleAdmin && (role == RoleAdmin || role == RoleOperator) {
		return true
	}
	return a.Role == role
}

// ResetFailedAttempts resets the failed login attempts counter
func (a *Admin) ResetFailedAttempts() {
	a.FailedLoginAttempts = 0
	a.LockedUntil = nil
}

// IncrementFailedAttempts increments failed login attempts and locks account if needed
func (a *Admin) IncrementFailedAttempts() {
	a.FailedLoginAttempts++

	// Lock account for 30 minutes after 5 failed attempts
	if a.FailedLoginAttempts >= 5 {
		lockUntil := time.Now().Add(30 * time.Minute)
		a.LockedUntil = &lockUntil
	}
}

// Sanitize removes sensitive information before returning to client
func (a *Admin) Sanitize() *Admin {
	a.PasswordHash = ""
	a.FailedLoginAttempts = 0
	return a
}
