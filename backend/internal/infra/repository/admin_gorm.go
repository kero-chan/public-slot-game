package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/admin"
	"gorm.io/gorm"
)

// AdminGormRepository implements admin.Repository using GORM
type AdminGormRepository struct {
	db *gorm.DB
}

// NewAdminGormRepository creates a new GORM admin repository
func NewAdminGormRepository(db *gorm.DB) admin.Repository {
	return &AdminGormRepository{
		db: db,
	}
}

// Create creates a new admin
func (r *AdminGormRepository) Create(ctx context.Context, adm *admin.Admin) error {
	if err := r.db.WithContext(ctx).Create(adm).Error; err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}
	return nil
}

// GetByID retrieves an admin by ID
func (r *AdminGormRepository) GetByID(ctx context.Context, id uuid.UUID) (*admin.Admin, error) {
	var adm admin.Admin
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&adm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, admin.ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to get admin by ID: %w", err)
	}
	return &adm, nil
}

// GetByUsername retrieves an admin by username
func (r *AdminGormRepository) GetByUsername(ctx context.Context, username string) (*admin.Admin, error) {
	var adm admin.Admin
	if err := r.db.WithContext(ctx).
		Where("username = ? AND deleted_at IS NULL", username).
		First(&adm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, admin.ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to get admin by username: %w", err)
	}
	return &adm, nil
}

// GetByEmail retrieves an admin by email
func (r *AdminGormRepository) GetByEmail(ctx context.Context, email string) (*admin.Admin, error) {
	var adm admin.Admin
	if err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&adm).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, admin.ErrAdminNotFound
		}
		return nil, fmt.Errorf("failed to get admin by email: %w", err)
	}
	return &adm, nil
}

// Update updates an admin
func (r *AdminGormRepository) Update(ctx context.Context, adm *admin.Admin) error {
	adm.UpdatedAt = time.Now()
	result := r.db.WithContext(ctx).Save(adm)
	if result.Error != nil {
		return fmt.Errorf("failed to update admin: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return admin.ErrAdminNotFound
	}
	return nil
}

// Delete soft deletes an admin
func (r *AdminGormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&admin.Admin{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to delete admin: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return admin.ErrAdminNotFound
	}
	return nil
}

// List retrieves all admins with optional filters
func (r *AdminGormRepository) List(ctx context.Context, filters admin.ListFilters) ([]*admin.Admin, int64, error) {
	var admins []*admin.Admin
	var total int64

	query := r.db.WithContext(ctx).
		Model(&admin.Admin{}).
		Where("deleted_at IS NULL")

	// Apply filters
	if filters.Role != nil {
		query = query.Where("role = ?", *filters.Role)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count admins: %w", err)
	}

	// Apply pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	// Get admins
	if err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&admins).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list admins: %w", err)
	}

	return admins, total, nil
}

// UpdateLastLogin updates the last login timestamp and IP
func (r *AdminGormRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&admin.Admin{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}
	return nil
}

// IncrementFailedAttempts increments failed login attempts
func (r *AdminGormRepository) IncrementFailedAttempts(ctx context.Context, id uuid.UUID) error {
	// Get current admin
	var adm admin.Admin
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&adm).Error; err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}

	// Increment attempts
	adm.IncrementFailedAttempts()

	// Update
	result := r.db.WithContext(ctx).
		Model(&admin.Admin{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"failed_login_attempts": adm.FailedLoginAttempts,
			"locked_until":          adm.LockedUntil,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to increment failed attempts: %w", result.Error)
	}
	return nil
}

// ResetFailedAttempts resets failed login attempts
func (r *AdminGormRepository) ResetFailedAttempts(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&admin.Admin{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"failed_login_attempts": 0,
			"locked_until":          nil,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to reset failed attempts: %w", result.Error)
	}
	return nil
}
