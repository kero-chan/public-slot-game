package admin

import "errors"

var (
	// ErrAdminNotFound is returned when an admin is not found
	ErrAdminNotFound = errors.New("admin not found")

	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrAccountLocked is returned when trying to login to a locked account
	ErrAccountLocked = errors.New("account is locked due to too many failed login attempts")

	// ErrAccountInactive is returned when trying to login to an inactive account
	ErrAccountInactive = errors.New("account is inactive")

	// ErrAccountSuspended is returned when trying to login to a suspended account
	ErrAccountSuspended = errors.New("account is suspended")

	// ErrDuplicateUsername is returned when trying to create an admin with existing username
	ErrDuplicateUsername = errors.New("username already exists")

	// ErrDuplicateEmail is returned when trying to create an admin with existing email
	ErrDuplicateEmail = errors.New("email already exists")

	// ErrInvalidRole is returned when an invalid role is provided
	ErrInvalidRole = errors.New("invalid admin role")

	// ErrInvalidPassword is returned when password validation fails
	ErrInvalidPassword = errors.New("invalid password")

	// ErrWeakPassword is returned when password is too weak
	ErrWeakPassword = errors.New("password is too weak - must be at least 8 characters")

	// ErrCannotDeleteSelf is returned when admin tries to delete their own account
	ErrCannotDeleteSelf = errors.New("cannot delete your own account")

	// ErrInsufficientPermissions is returned when admin lacks required permissions
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// ErrCannotModifySuperAdmin is returned when trying to modify super admin without permission
	ErrCannotModifySuperAdmin = errors.New("cannot modify super admin account")
)
