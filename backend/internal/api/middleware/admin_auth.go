package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	adminDomain "github.com/slotmachine/backend/domain/admin"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/pkg/util"
)

// AdminAuthMiddleware validates admin JWT tokens
func AdminAuthMiddleware(cfg *config.Config, log *logger.Logger, adminService adminDomain.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Warn().Msg("Missing authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing authorization token",
			})
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn().Msg("Invalid authorization header format")
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
		}

		token := parts[1]

		// Check if it's a service token (never expires, for service-to-service auth)
		if cfg.JWT.ServiceToken != "" && token == cfg.JWT.ServiceToken {
			log.Debug().Msg("Service token authentication")
			// Set service account context
			c.Locals("user_id", "service")
			c.Locals("username", "internal-service")
			c.Locals("admin_role", adminDomain.RoleSuperAdmin)
			c.Locals("is_service", true)
			return c.Next()
		}

		// Parse and validate JWT
		claims, err := util.ValidateJWT(token, cfg.JWT.Secret)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid token")
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
		}

		// Validate admin exists and is active
		admin, err := adminService.ValidateToken(c.Context(), token)
		if err != nil {
			log.Warn().Err(err).Str("user_id", claims.UserID).Msg("Admin validation failed")
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or inactive admin account",
			})
		}

		// Store admin info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("admin_role", admin.Role)
		c.Locals("admin", admin)

		return c.Next()
	}
}

// AdminRoleMiddleware checks if admin has required role
func AdminRoleMiddleware(requiredRole adminDomain.AdminRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		admin, ok := c.Locals("admin").(*adminDomain.Admin)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Admin authentication required",
			})
		}

		if !admin.HasRole(requiredRole) {
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error:   "forbidden",
				Message: "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

// AdminPermissionMiddleware checks if admin has specific permission
func AdminPermissionMiddleware(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		admin, ok := c.Locals("admin").(*adminDomain.Admin)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   "unauthorized",
				Message: "Admin authentication required",
			})
		}

		if !admin.CanAccess(permission) {
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error:   "forbidden",
				Message: "Insufficient permissions for this action",
			})
		}

		return c.Next()
	}
}
