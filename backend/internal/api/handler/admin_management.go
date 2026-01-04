package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	adminDomain "github.com/slotmachine/backend/domain/admin"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// AdminManagementHandler handles admin management endpoints
type AdminManagementHandler struct {
	adminService adminDomain.Service
	logger       *logger.Logger
}

// NewAdminManagementHandler creates a new admin management handler
func NewAdminManagementHandler(
	adminService adminDomain.Service,
	log *logger.Logger,
) *AdminManagementHandler {
	return &AdminManagementHandler{
		adminService: adminService,
		logger:       log,
	}
}

// CreateAdmin creates a new admin
// POST /admin/management/admins
func (h *AdminManagementHandler) CreateAdmin(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get current admin from context (set by auth middleware)
	currentAdmin := getAdminFromContext(c)
	var adminID uuid.UUID
	if currentAdmin != nil {
		adminID = currentAdmin.ID
	}

	var req dto.CreateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Create admin
	createReq := adminDomain.CreateAdminRequest{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		FullName:    req.FullName,
		Role:        adminDomain.AdminRole(req.Role),
		Permissions: req.Permissions,
	}

	admin, err := h.adminService.CreateAdmin(c.Context(), createReq, adminID)
	if err != nil {
		switch err {
		case adminDomain.ErrDuplicateUsername:
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "duplicate_username",
				Message: "Username already exists",
			})
		case adminDomain.ErrDuplicateEmail:
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "duplicate_email",
				Message: "Email already exists",
			})
		case adminDomain.ErrWeakPassword:
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "weak_password",
				Message: "Password must be at least 8 characters long",
			})
		default:
			log.Error().Err(err).Msg("Failed to create admin")
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "failed_to_create_admin",
				Message: "Failed to create admin",
			})
		}
	}

	log.Info().
		Str("admin_id", admin.ID.String()).
		Str("created_by", adminID.String()).
		Msg("Admin created successfully")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    mapAdminToDetail(admin),
	})
}

// GetAdmin retrieves an admin by ID
// GET /admin/management/admins/:id
func (h *AdminManagementHandler) GetAdmin(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid admin ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_admin_id",
			Message: "Invalid admin ID",
		})
	}

	admin, err := h.adminService.GetAdmin(c.Context(), adminID)
	if err != nil {
		if err == adminDomain.ErrAdminNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "admin_not_found",
				Message: "Admin not found",
			})
		}
		log.Error().Err(err).Msg("Failed to get admin")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_admin",
			Message: "Failed to retrieve admin",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    mapAdminToDetail(admin),
	})
}

// ListAdmins retrieves all admins
// GET /admin/management/admins
func (h *AdminManagementHandler) ListAdmins(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	filters := adminDomain.ListFilters{
		Page:     page,
		PageSize: limit,
	}

	// Optional filters
	if role := c.Query("role"); role != "" {
		r := adminDomain.AdminRole(role)
		filters.Role = &r
	}
	if status := c.Query("status"); status != "" {
		s := adminDomain.AdminStatus(status)
		filters.Status = &s
	}

	admins, total, err := h.adminService.ListAdmins(c.Context(), filters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list admins")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_list_admins",
			Message: "Failed to retrieve admins",
		})
	}

	// Map to DTOs
	adminDetails := make([]*dto.AdminDetail, len(admins))
	for i, admin := range admins {
		adminDetails[i] = mapAdminToDetail(admin)
	}

	response := dto.ListAdminsResponse{
		Admins: adminDetails,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// UpdateAdmin updates an admin
// PUT /admin/management/admins/:id
func (h *AdminManagementHandler) UpdateAdmin(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get current admin from context (set by auth middleware)
	currentAdmin := getAdminFromContext(c)
	var currentAdminID uuid.UUID
	if currentAdmin != nil {
		currentAdminID = currentAdmin.ID
	}

	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid admin ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_admin_id",
			Message: "Invalid admin ID",
		})
	}

	var req dto.UpdateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Build update request
	updateReq := adminDomain.UpdateAdminRequest{
		Email:       req.Email,
		FullName:    req.FullName,
		Permissions: req.Permissions,
	}
	if req.Role != nil {
		r := adminDomain.AdminRole(*req.Role)
		updateReq.Role = &r
	}

	admin, err := h.adminService.UpdateAdmin(c.Context(), adminID, updateReq, currentAdminID)
	if err != nil {
		switch err {
		case adminDomain.ErrAdminNotFound:
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "admin_not_found",
				Message: "Admin not found",
			})
		case adminDomain.ErrDuplicateEmail:
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "duplicate_email",
				Message: "Email already exists",
			})
		default:
			log.Error().Err(err).Msg("Failed to update admin")
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "failed_to_update_admin",
				Message: "Failed to update admin",
			})
		}
	}

	log.Info().
		Str("admin_id", adminID.String()).
		Str("updated_by", currentAdminID.String()).
		Msg("Admin updated successfully")

	return c.JSON(fiber.Map{
		"success": true,
		"data":    mapAdminToDetail(admin),
	})
}

// DeleteAdmin soft deletes an admin
// DELETE /admin/management/admins/:id
func (h *AdminManagementHandler) DeleteAdmin(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get current admin from context (set by auth middleware)
	currentAdmin := getAdminFromContext(c)
	var currentAdminID uuid.UUID
	if currentAdmin != nil {
		currentAdminID = currentAdmin.ID
	}

	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid admin ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_admin_id",
			Message: "Invalid admin ID",
		})
	}

	if err := h.adminService.DeleteAdmin(c.Context(), adminID, currentAdminID); err != nil {
		switch err {
		case adminDomain.ErrAdminNotFound:
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "admin_not_found",
				Message: "Admin not found",
			})
		case adminDomain.ErrCannotDeleteSelf:
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error:   "cannot_delete_self",
				Message: "Cannot delete your own account",
			})
		case adminDomain.ErrCannotModifySuperAdmin:
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error:   "cannot_modify_super_admin",
				Message: "Insufficient permissions to modify super admin",
			})
		default:
			log.Error().Err(err).Msg("Failed to delete admin")
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "failed_to_delete_admin",
				Message: "Failed to delete admin",
			})
		}
	}

	log.Info().
		Str("admin_id", adminID.String()).
		Str("deleted_by", currentAdminID.String()).
		Msg("Admin deleted successfully")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Admin deleted successfully",
	})
}

// ResetPassword resets an admin's password (admin action)
// POST /admin/management/admins/:id/reset-password
func (h *AdminManagementHandler) ResetPassword(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get current admin from context (set by auth middleware)
	currentAdmin := getAdminFromContext(c)
	var currentAdminID uuid.UUID
	if currentAdmin != nil {
		currentAdminID = currentAdmin.ID
	}

	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid admin ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_admin_id",
			Message: "Invalid admin ID",
		})
	}

	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.adminService.ResetPassword(c.Context(), adminID, req.NewPassword, currentAdminID); err != nil {
		switch err {
		case adminDomain.ErrAdminNotFound:
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "admin_not_found",
				Message: "Admin not found",
			})
		case adminDomain.ErrWeakPassword:
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "weak_password",
				Message: "Password must be at least 8 characters long",
			})
		default:
			log.Error().Err(err).Msg("Failed to reset password")
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "failed_to_reset_password",
				Message: "Failed to reset password",
			})
		}
	}

	log.Info().
		Str("admin_id", adminID.String()).
		Str("reset_by", currentAdminID.String()).
		Msg("Password reset successfully")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Password reset successfully",
	})
}

// ActivateAdmin activates an admin account
// POST /admin/management/admins/:id/activate
func (h *AdminManagementHandler) ActivateAdmin(c *fiber.Ctx) error {
	return h.updateAdminStatus(c, "activate")
}

// DeactivateAdmin deactivates an admin account
// POST /admin/management/admins/:id/deactivate
func (h *AdminManagementHandler) DeactivateAdmin(c *fiber.Ctx) error {
	return h.updateAdminStatus(c, "deactivate")
}

// SuspendAdmin suspends an admin account
// POST /admin/management/admins/:id/suspend
func (h *AdminManagementHandler) SuspendAdmin(c *fiber.Ctx) error {
	return h.updateAdminStatus(c, "suspend")
}

// Helper function to update admin status
func (h *AdminManagementHandler) updateAdminStatus(c *fiber.Ctx, action string) error {
	log := h.logger.WithTrace(c)

	// Get current admin from context (set by auth middleware)
	currentAdmin := getAdminFromContext(c)
	var currentAdminID uuid.UUID
	if currentAdmin != nil {
		currentAdminID = currentAdmin.ID
	}

	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Warn().Err(err).Msg("Invalid admin ID")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_admin_id",
			Message: "Invalid admin ID",
		})
	}

	var serviceErr error
	switch action {
	case "activate":
		serviceErr = h.adminService.ActivateAdmin(c.Context(), adminID, currentAdminID)
	case "deactivate":
		serviceErr = h.adminService.DeactivateAdmin(c.Context(), adminID, currentAdminID)
	case "suspend":
		serviceErr = h.adminService.SuspendAdmin(c.Context(), adminID, currentAdminID)
	}

	if serviceErr != nil {
		if serviceErr == adminDomain.ErrAdminNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "admin_not_found",
				Message: "Admin not found",
			})
		}
		log.Error().Err(serviceErr).Msg("Failed to " + action + " admin")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_" + action,
			Message: "Failed to " + action + " admin",
		})
	}

	log.Info().
		Str("admin_id", adminID.String()).
		Str("updated_by", currentAdminID.String()).
		Msg("Admin " + action + "d successfully")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Admin " + action + "d successfully",
	})
}
