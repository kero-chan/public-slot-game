package handler

import (
	"github.com/gofiber/fiber/v2"
	adminDomain "github.com/slotmachine/backend/domain/admin"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// AdminAuthHandler handles admin authentication endpoints
type AdminAuthHandler struct {
	adminService adminDomain.Service
	logger       *logger.Logger
}

// NewAdminAuthHandler creates a new admin auth handler
func NewAdminAuthHandler(
	adminService adminDomain.Service,
	log *logger.Logger,
) *AdminAuthHandler {
	return &AdminAuthHandler{
		adminService: adminService,
		logger:       log,
	}
}

// Login authenticates an admin
// POST /admin/auth/login
func (h *AdminAuthHandler) Login(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	var req dto.AdminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Get client IP
	ip := c.Get("x-real-ip")
	if ip == "" {
		ip = c.IP()
	}

	// Authenticate admin
	admin, token, err := h.adminService.Login(c.Context(), req.Username, req.Password, ip)
	if err != nil {
		log.Warn().
			Err(err).
			Str("username", req.Username).
			Msg("Admin login failed")

		// Map errors to appropriate status codes
		switch err {
		case adminDomain.ErrInvalidCredentials:
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid username or password",
			})
		case adminDomain.ErrAccountLocked:
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error:   "account_locked",
				Message: "Account is locked due to too many failed login attempts",
			})
		case adminDomain.ErrAccountInactive:
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error:   "account_inactive",
				Message: "Account is inactive",
			})
		case adminDomain.ErrAccountSuspended:
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error:   "account_suspended",
				Message: "Account has been suspended",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "login_failed",
				Message: "Login failed. Please try again.",
			})
		}
	}

	log.Info().
		Str("admin_id", admin.ID.String()).
		Str("username", req.Username).
		Msg("Admin logged in successfully")

	response := dto.AdminLoginResponse{
		Token: token,
		Admin: mapAdminToDetail(admin),
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetProfile retrieves the current admin's profile
// GET /admin/auth/profile
func (h *AdminAuthHandler) GetProfile(c *fiber.Ctx) error {
	// Get admin from context (set by auth middleware)
	admin, ok := c.Locals("admin").(*adminDomain.Admin)
	if !ok || admin == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Admin authentication required",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    mapAdminToDetail(admin),
	})
}

// ChangePassword changes the current admin's password
// POST /admin/auth/change-password
func (h *AdminAuthHandler) ChangePassword(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get admin from context (set by auth middleware)
	admin, ok := c.Locals("admin").(*adminDomain.Admin)
	if !ok || admin == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Admin authentication required",
		})
	}

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		log.Warn().Err(err).Msg("Invalid request body")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Change password
	if err := h.adminService.ChangePassword(c.Context(), admin.ID, req.OldPassword, req.NewPassword); err != nil {
		switch err {
		case adminDomain.ErrInvalidPassword:
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_password",
				Message: "Current password is incorrect",
			})
		case adminDomain.ErrWeakPassword:
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "weak_password",
				Message: "Password must be at least 8 characters long",
			})
		default:
			log.Error().Err(err).Msg("Failed to change password")
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "failed_to_change_password",
				Message: "Failed to change password",
			})
		}
	}

	log.Info().Str("admin_id", admin.ID.String()).Msg("Password changed successfully")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Password changed successfully",
	})
}

// Helper function to map admin domain model to DTO
func mapAdminToDetail(admin *adminDomain.Admin) *dto.AdminDetail {
	return &dto.AdminDetail{
		ID:          admin.ID,
		Username:    admin.Username,
		Email:       admin.Email,
		FullName:    admin.FullName,
		Role:        string(admin.Role),
		Status:      string(admin.Status),
		Permissions: []string(admin.Permissions),
		LastLoginAt: admin.LastLoginAt,
		CreatedAt:   admin.CreatedAt,
	}
}
