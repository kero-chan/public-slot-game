package handler

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/infra/storage"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/pkg/security"
)

// AdminUploadHandler handles file upload endpoints
type AdminUploadHandler struct {
	storage   storage.Storage
	logger    *logger.Logger
	validator *security.FileValidator
}

// NewAdminUploadHandler creates a new admin upload handler
func NewAdminUploadHandler(
	s storage.Storage,
	log *logger.Logger,
) *AdminUploadHandler {
	return &AdminUploadHandler{
		storage:   s,
		logger:    log,
		validator: security.NewFileValidator(nil),
	}
}

// validateThemeName validates theme name to prevent path traversal and invalid characters
func (h *AdminUploadHandler) validateThemeName(themeName string) error {
	if themeName == "" {
		return fmt.Errorf("theme name is required")
	}
	// Prevent path traversal
	if strings.Contains(themeName, "..") || strings.Contains(themeName, "/") || strings.Contains(themeName, "\\") {
		return fmt.Errorf("invalid theme name: path traversal not allowed")
	}
	// Only allow alphanumeric, dash, underscore
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(themeName) {
		return fmt.Errorf("invalid theme name: only alphanumeric, dash, and underscore allowed")
	}
	// Length limit
	if len(themeName) > 100 {
		return fmt.Errorf("theme name too long (max 100 characters)")
	}
	return nil
}

// UploadFile handles single file upload
// POST /admin/upload/:theme
func (h *AdminUploadHandler) UploadFile(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	if err := h.validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_file",
			Message: "File is required",
		})
	}

	// Optional: get custom path from form
	customPath := c.FormValue("path", "")

	// Reject ZIP files - use folder upload instead
	if h.validator.IsZipFile(file.Filename) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "zip_not_supported",
			Message: "ZIP files are not supported. Please use folder upload via direct-upload endpoint instead.",
		})
	}

	// Regular file upload
	return h.handleSingleFileUpload(c, themeName, file, customPath, log)
}

// handleSingleFileUpload processes a single non-ZIP file
func (h *AdminUploadHandler) handleSingleFileUpload(
	c *fiber.Ctx,
	themeName string,
	file *multipart.FileHeader,
	customPath string,
	log *logger.Logger,
) error {
	fileName := file.Filename
	if customPath != "" {
		fileName = filepath.Join(customPath, file.Filename)
	}

	// Validate file type
	if !h.validator.IsAllowedExtension(fileName) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_file_type",
			Message: "File type not allowed. Allowed types: png, jpg, jpeg, gif, webp, svg, json, mp3, wav, ogg, mp4, webm",
		})
	}

	// Validate file size
	if err := h.validator.ValidateFileSize(file.Size, false); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "file_too_large",
			Message: err.Error(),
		})
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open uploaded file")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "file_open_failed",
			Message: "Failed to process uploaded file",
		})
	}
	defer src.Close()

	// Read only first 512 bytes for magic bytes validation (memory efficient)
	magicBytes := make([]byte, 512)
	n, err := src.Read(magicBytes)
	if err != nil && err != io.EOF {
		log.Error().Err(err).Msg("Failed to read file header")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "file_read_failed",
			Message: "Failed to read file",
		})
	}

	// Validate magic bytes
	if err := h.validator.ValidateMagicBytes(fileName, bytes.NewReader(magicBytes[:n])); err != nil {
		log.Warn().Err(err).Str("filename", fileName).Msg("Magic bytes validation failed")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_file_content",
			Message: "File content does not match file type",
		})
	}

	// Reset file position to beginning for upload
	if seeker, ok := src.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			log.Error().Err(err).Msg("Failed to reset file position")
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error:   "file_read_failed",
				Message: "Failed to process file",
			})
		}
	}

	// Upload to storage (streaming - no full file in memory)
	url, err := h.storage.UploadFile(
		c.Context(),
		themeName,
		fileName,
		src,
		file.Size,
		security.GetContentType(fileName),
	)
	if err != nil {
		log.Error().Err(err).Str("theme", themeName).Str("file", fileName).Msg("Failed to upload file to storage")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "upload_failed",
			Message: "Failed to upload file to storage",
		})
	}

	log.Info().Str("theme", themeName).Str("file", fileName).Str("url", url).Msg("File uploaded")

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"url":      url,
			"filename": fileName,
			"size":     file.Size,
		},
	})
}

// UploadMultipleFiles handles multiple file uploads
// POST /admin/upload/:theme/batch
func (h *AdminUploadHandler) UploadMultipleFiles(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	if err := h.validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	// Get multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_form",
			Message: "Invalid multipart form",
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "no_files",
			Message: "No files provided",
		})
	}

	// Optional: get base path from form
	basePath := c.FormValue("path", "")

	var uploadedFiles []fiber.Map
	var errors []string
	config := h.validator.GetConfig()

	// Calculate total size to check memory limits
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}

	// Reject if total batch size exceeds limit (prevent memory exhaustion)
	maxBatchSize := config.MaxTotalExtractSize // Reuse this limit for batch uploads
	if totalSize > maxBatchSize {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "batch_too_large",
			Message: fmt.Sprintf("Total batch size %d bytes exceeds limit of %d bytes", totalSize, maxBatchSize),
		})
	}

	// Process files one at a time with streaming to minimize memory usage
	for _, file := range files {
		// Reject ZIP files
		if h.validator.IsZipFile(file.Filename) {
			errors = append(errors, fmt.Sprintf("%s: ZIP files are not supported. Use folder upload instead.", file.Filename))
			continue
		}

		// Validate size before reading
		if err := h.validator.ValidateFileSize(file.Size, false); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", file.Filename, err.Error()))
			continue
		}

		// Stream directly to storage
		uploaded, fileErr := h.processSingleFileStreaming(c, themeName, file, basePath, log)
		if fileErr != "" {
			errors = append(errors, fileErr)
		} else if uploaded != nil {
			uploadedFiles = append(uploadedFiles, uploaded)
		}
	}

	return c.JSON(fiber.Map{
		"success": len(errors) == 0,
		"data": fiber.Map{
			"uploaded": uploadedFiles,
			"errors":   errors,
			"total":    len(files),
			"success":  len(uploadedFiles),
			"failed":   len(errors),
		},
	})
}

// processSingleFileStreaming handles a single file with streaming (no full file in memory)
func (h *AdminUploadHandler) processSingleFileStreaming(
	c *fiber.Ctx,
	themeName string,
	file *multipart.FileHeader,
	basePath string,
	log *logger.Logger,
) (fiber.Map, string) {
	fileName := file.Filename
	if basePath != "" {
		fileName = filepath.Join(basePath, file.Filename)
	}

	// Validate file type
	if !h.validator.IsAllowedExtension(fileName) {
		return nil, fmt.Sprintf("%s: file type not allowed", file.Filename)
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Sprintf("%s: failed to open file", file.Filename)
	}
	defer src.Close()

	// Read only first 512 bytes for magic validation
	magicBytes := make([]byte, 512)
	n, err := src.Read(magicBytes)
	if err != nil && err != io.EOF {
		return nil, fmt.Sprintf("%s: failed to read file", file.Filename)
	}

	// Validate magic bytes
	if err := h.validator.ValidateMagicBytes(fileName, bytes.NewReader(magicBytes[:n])); err != nil {
		log.Warn().Err(err).Str("filename", fileName).Msg("Magic bytes validation failed")
		return nil, fmt.Sprintf("%s: content does not match file type", file.Filename)
	}

	// Reset file position
	if seeker, ok := src.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Sprintf("%s: failed to reset file position", file.Filename)
		}
	}

	// Upload to storage with streaming
	url, err := h.storage.UploadFile(
		c.Context(),
		themeName,
		fileName,
		src,
		file.Size,
		security.GetContentType(fileName),
	)
	if err != nil {
		log.Error().Err(err).Str("theme", themeName).Str("file", fileName).Msg("Failed to upload file")
		return nil, fmt.Sprintf("%s: upload failed", file.Filename)
	}

	log.Info().Str("theme", themeName).Str("file", fileName).Msg("File uploaded")

	return fiber.Map{
		"url":      url,
		"filename": fileName,
		"size":     file.Size,
	}, ""
}

// ListFiles lists all files in a theme
// GET /admin/upload/:theme/files
func (h *AdminUploadHandler) ListFiles(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	if err := h.validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	files, err := h.storage.ListFiles(c.Context(), themeName)
	if err != nil {
		log.Error().Err(err).Str("theme", themeName).Msg("Failed to list files")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "list_failed",
			Message: "Failed to list files",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"theme": themeName,
			"files": files,
			"count": len(files),
		},
	})
}

// DeleteFile deletes a file from storage
// DELETE /admin/upload/:theme/:filename
func (h *AdminUploadHandler) DeleteFile(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	if err := h.validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	fileName := c.Params("*") // Capture rest of path as filename
	if fileName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_params",
			Message: "File name is required",
		})
	}

	err := h.storage.DeleteFile(c.Context(), themeName, fileName)
	if err != nil {
		log.Error().Err(err).Str("theme", themeName).Str("file", fileName).Msg("Failed to delete file")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete file",
		})
	}

	log.Info().Str("theme", themeName).Str("file", fileName).Msg("File deleted")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "File deleted",
	})
}

// DeleteTheme deletes all files in a theme
// DELETE /admin/upload/:theme
func (h *AdminUploadHandler) DeleteTheme(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	if err := h.validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	err := h.storage.DeleteTheme(c.Context(), themeName)
	if err != nil {
		log.Error().Err(err).Str("theme", themeName).Msg("Failed to delete theme files")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete theme files",
		})
	}

	log.Info().Str("theme", themeName).Msg("Theme files deleted")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "All theme files deleted",
	})
}
