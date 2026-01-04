package handler

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/infra/storage"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/pkg/security"
)

// AdminDirectUploadHandler handles direct upload endpoints (presigned URL based)
type AdminDirectUploadHandler struct {
	storage   storage.Storage
	logger    *logger.Logger
	validator *security.FileValidator
}

// NewAdminDirectUploadHandler creates a new direct upload handler
func NewAdminDirectUploadHandler(
	s storage.Storage,
	log *logger.Logger,
) *AdminDirectUploadHandler {
	return &AdminDirectUploadHandler{
		storage:   s,
		logger:    log,
		validator: security.NewFileValidator(nil),
	}
}

// validateThemeName validates theme name to prevent path traversal
func (h *AdminDirectUploadHandler) validateThemeName(themeName string) error {
	if themeName == "" {
		return fmt.Errorf("theme name is required")
	}
	if strings.Contains(themeName, "..") || strings.Contains(themeName, "/") || strings.Contains(themeName, "\\") {
		return fmt.Errorf("invalid theme name: path traversal not allowed")
	}
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(themeName) {
		return fmt.Errorf("invalid theme name: only alphanumeric, dash, and underscore allowed")
	}
	if len(themeName) > 100 {
		return fmt.Errorf("theme name too long (max 100 characters)")
	}
	return nil
}

// PresignedURLRequest represents a request for a single presigned URL
type PresignedURLRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type,omitempty"`
	FilePath    string `json:"file_path,omitempty"` // Optional subfolder path
}

// BatchPresignedURLRequest represents a request for multiple presigned URLs
type BatchPresignedURLRequest struct {
	Files         []PresignedURLRequest `json:"files"`
	ExpiryMinutes int                   `json:"expiry_minutes,omitempty"` // Default 15 minutes
}

// PresignedURLResponse represents the response for a presigned URL request
type PresignedURLResponse struct {
	FileName    string            `json:"file_name"`
	FilePath    string            `json:"file_path"`
	UploadURL   string            `json:"upload_url"`
	PublicURL   string            `json:"public_url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	ExpiresAt   string            `json:"expires_at"`
	ContentType string            `json:"content_type"`
}

// GeneratePresignedURL generates a presigned URL for direct upload
// POST /admin/direct-upload/:theme/presign
func (h *AdminDirectUploadHandler) GeneratePresignedURL(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	if err := h.validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	var req PresignedURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.FileName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "file_name is required",
		})
	}

	// Validate file type (no ZIP files allowed - use folder upload instead)
	if h.validator.IsZipFile(req.FileName) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "zip_not_supported",
			Message: "ZIP files are not supported. Please use folder upload instead.",
		})
	}

	if !h.validator.IsAllowedExtension(req.FileName) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_file_type",
			Message: "File type not allowed",
		})
	}

	// Build full file path
	fileName := req.FileName
	if req.FilePath != "" {
		fileName = filepath.Join(req.FilePath, req.FileName)
	}

	// Generate presigned URL
	info, err := h.storage.GeneratePresignedUploadURL(c.Context(), themeName, fileName, req.ContentType, 15)
	if err != nil {
		log.Error().Err(err).Str("theme", themeName).Str("file", fileName).Msg("Failed to generate presigned URL")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "presign_failed",
			Message: "Failed to generate upload URL",
		})
	}

	log.Info().Str("theme", themeName).Str("file", fileName).Msg("Generated presigned URL for direct upload")

	return c.JSON(fiber.Map{
		"success": true,
		"data": PresignedURLResponse{
			FileName:    req.FileName,
			FilePath:    fileName,
			UploadURL:   info.UploadURL,
			PublicURL:   info.PublicURL,
			Method:      info.Method,
			Headers:     info.Headers,
			ExpiresAt:   info.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			ContentType: info.ContentType,
		},
	})
}

// GenerateBatchPresignedURLs generates presigned URLs for multiple files (folder upload)
// POST /admin/direct-upload/:theme/presign/batch
func (h *AdminDirectUploadHandler) GenerateBatchPresignedURLs(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	if err := h.validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	var req BatchPresignedURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if len(req.Files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "files array is required and must not be empty",
		})
	}

	// Limit batch size
	if len(req.Files) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "batch_too_large",
			Message: "Maximum 100 files per batch",
		})
	}

	expiryMinutes := req.ExpiryMinutes
	if expiryMinutes <= 0 {
		expiryMinutes = 15
	}

	var results []PresignedURLResponse
	var errors []fiber.Map

	for _, file := range req.Files {
		if file.FileName == "" {
			errors = append(errors, fiber.Map{
				"file_name": file.FileName,
				"error":     "file_name is required",
			})
			continue
		}

		// Validate file type (no ZIP files allowed - use folder upload instead)
		if h.validator.IsZipFile(file.FileName) {
			errors = append(errors, fiber.Map{
				"file_name": file.FileName,
				"error":     "ZIP files not supported - use folder upload instead",
			})
			continue
		}

		if !h.validator.IsAllowedExtension(file.FileName) {
			errors = append(errors, fiber.Map{
				"file_name": file.FileName,
				"error":     "file type not allowed",
			})
			continue
		}

		// Build full file path
		fileName := file.FileName
		if file.FilePath != "" {
			fileName = filepath.Join(file.FilePath, file.FileName)
		}

		// Generate presigned URL
		info, err := h.storage.GeneratePresignedUploadURL(c.Context(), themeName, fileName, file.ContentType, expiryMinutes)
		if err != nil {
			log.Error().Err(err).Str("file", fileName).Msg("Failed to generate presigned URL")
			errors = append(errors, fiber.Map{
				"file_name": file.FileName,
				"error":     "failed to generate upload URL",
			})
			continue
		}

		results = append(results, PresignedURLResponse{
			FileName:    file.FileName,
			FilePath:    fileName,
			UploadURL:   info.UploadURL,
			PublicURL:   info.PublicURL,
			Method:      info.Method,
			Headers:     info.Headers,
			ExpiresAt:   info.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			ContentType: info.ContentType,
		})
	}

	log.Info().
		Str("theme", themeName).
		Int("requested", len(req.Files)).
		Int("success", len(results)).
		Int("errors", len(errors)).
		Msg("Generated batch presigned URLs")

	return c.JSON(fiber.Map{
		"success": len(errors) == 0,
		"data": fiber.Map{
			"uploads":       results,
			"errors":        errors,
			"total":         len(req.Files),
			"success_count": len(results),
			"error_count":   len(errors),
		},
	})
}

// ConfirmUploadRequest represents a request to confirm upload completion
type ConfirmUploadRequest struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path,omitempty"`
}

// BatchConfirmUploadRequest represents a request to confirm multiple uploads
type BatchConfirmUploadRequest struct {
	Files []ConfirmUploadRequest `json:"files"`
}

// ConfirmUpload confirms that a file was successfully uploaded via presigned URL
// POST /admin/direct-upload/:theme/confirm
func (h *AdminDirectUploadHandler) ConfirmUpload(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	if err := h.validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	var req ConfirmUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if req.FileName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "file_name is required",
		})
	}

	// Build full file path
	fileName := req.FileName
	if req.FilePath != "" {
		fileName = filepath.Join(req.FilePath, req.FileName)
	}

	// Verify the file exists in storage
	exists, err := h.storage.FileExists(c.Context(), themeName, fileName)
	if err != nil {
		log.Error().Err(err).Str("theme", themeName).Str("file", fileName).Msg("Failed to check file existence")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "verification_failed",
			Message: "Failed to verify upload",
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "file_not_found",
			Message: "File was not uploaded or upload failed",
		})
	}

	publicURL := h.storage.GetPublicURL(themeName, fileName)

	log.Info().Str("theme", themeName).Str("file", fileName).Msg("Confirmed direct upload")

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"file_name":  req.FileName,
			"file_path":  fileName,
			"public_url": publicURL,
			"verified":   true,
		},
	})
}

// ConfirmBatchUpload confirms multiple files were successfully uploaded
// POST /admin/direct-upload/:theme/confirm/batch
func (h *AdminDirectUploadHandler) ConfirmBatchUpload(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	if err := h.validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	var req BatchConfirmUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if len(req.Files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "files array is required",
		})
	}

	var confirmed []fiber.Map
	var notFound []fiber.Map
	var errors []fiber.Map

	for _, file := range req.Files {
		if file.FileName == "" {
			errors = append(errors, fiber.Map{
				"file_name": file.FileName,
				"error":     "file_name is required",
			})
			continue
		}

		// Build full file path
		fileName := file.FileName
		if file.FilePath != "" {
			fileName = filepath.Join(file.FilePath, file.FileName)
		}

		// Verify the file exists in storage
		exists, err := h.storage.FileExists(c.Context(), themeName, fileName)
		if err != nil {
			log.Error().Err(err).Str("file", fileName).Msg("Failed to check file existence")
			errors = append(errors, fiber.Map{
				"file_name": file.FileName,
				"error":     "verification failed",
			})
			continue
		}

		if !exists {
			notFound = append(notFound, fiber.Map{
				"file_name": file.FileName,
				"file_path": fileName,
			})
			continue
		}

		publicURL := h.storage.GetPublicURL(themeName, fileName)
		confirmed = append(confirmed, fiber.Map{
			"file_name":  file.FileName,
			"file_path":  fileName,
			"public_url": publicURL,
		})
	}

	log.Info().
		Str("theme", themeName).
		Int("requested", len(req.Files)).
		Int("confirmed", len(confirmed)).
		Int("not_found", len(notFound)).
		Int("errors", len(errors)).
		Msg("Confirmed batch direct uploads")

	return c.JSON(fiber.Map{
		"success": len(notFound) == 0 && len(errors) == 0,
		"data": fiber.Map{
			"confirmed":       confirmed,
			"not_found":       notFound,
			"errors":          errors,
			"total":           len(req.Files),
			"confirmed_count": len(confirmed),
			"not_found_count": len(notFound),
			"error_count":     len(errors),
		},
	})
}
