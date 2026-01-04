package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slotmachine/backend/internal/api/dto"
	infraCache "github.com/slotmachine/backend/internal/infra/cache"
	"github.com/slotmachine/backend/internal/infra/storage"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/pkg/security"
)

// ChunkedUploadSession tracks an ongoing chunked upload
type ChunkedUploadSession struct {
	UploadID       string            `json:"upload_id"`
	ThemeName      string            `json:"theme_name"`
	FileName       string            `json:"file_name"`
	TotalSize      int64             `json:"total_size"`
	ChunkSize      int64             `json:"chunk_size"`
	TotalChunks    int               `json:"total_chunks"`
	UploadedChunks map[int]bool      `json:"-"`
	ChunkChecksums map[int]string    `json:"-"` // SHA256 checksums per chunk
	TempDir        string            `json:"-"`
	CreatedAt      time.Time         `json:"created_at"`
	ExpiresAt      time.Time         `json:"expires_at"`
	CustomPath     string            `json:"custom_path,omitempty"`
	FileChecksum   string            `json:"file_checksum,omitempty"` // Optional SHA256 of entire file
	mu             sync.Mutex
}

// ProcessingStatus tracks the status of background file processing
type ProcessingStatus struct {
	UploadID    string          `json:"upload_id"`
	Status      string          `json:"status"` // "processing", "completed", "failed"
	Progress    int             `json:"progress"` // 0-100
	Message     string          `json:"message,omitempty"`
	Error       string          `json:"error,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
	StartedAt   time.Time       `json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

// Redis key prefix for processing status
const (
	ProcessingStatusKeyPrefix = "chunked_upload_status:"
	ProcessingStatusTTL       = 1 * time.Hour // TTL for processing status in Redis
)

// MaxConcurrentUploads limits the number of concurrent upload sessions to prevent memory exhaustion
const MaxConcurrentUploads = 100

// AdminChunkedUploadHandler handles chunked file upload endpoints
type AdminChunkedUploadHandler struct {
	storage      storage.Storage
	logger       *logger.Logger
	validator    *security.FileValidator
	redis        *infraCache.RedisClient
	sessions     map[string]*ChunkedUploadSession
	sessionMu    sync.RWMutex
	processing   map[string]*ProcessingStatus // In-memory fallback when Redis unavailable
	processingMu sync.RWMutex
	tempBase     string
}

// NewAdminChunkedUploadHandler creates a new chunked upload handler
func NewAdminChunkedUploadHandler(
	s storage.Storage,
	log *logger.Logger,
	redis *infraCache.RedisClient,
) *AdminChunkedUploadHandler {
	handler := &AdminChunkedUploadHandler{
		storage:    s,
		logger:     log,
		validator:  security.NewFileValidator(nil),
		redis:      redis,
		sessions:   make(map[string]*ChunkedUploadSession),
		processing: make(map[string]*ProcessingStatus),
		tempBase:   os.TempDir(),
	}

	// Start cleanup goroutines
	go handler.cleanupExpiredSessions()
	go handler.cleanupCompletedProcessing()

	return handler
}

// cleanupCompletedProcessing periodically removes completed processing statuses from in-memory fallback
func (h *AdminChunkedUploadHandler) cleanupCompletedProcessing() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.processingMu.Lock()
		now := time.Now()
		for id, status := range h.processing {
			// Remove completed/failed statuses older than 30 minutes
			if status.CompletedAt != nil && now.Sub(*status.CompletedAt) > 30*time.Minute {
				delete(h.processing, id)
			}
		}
		h.processingMu.Unlock()
	}
}

// cleanupExpiredSessions periodically removes expired upload sessions
func (h *AdminChunkedUploadHandler) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.sessionMu.Lock()
		now := time.Now()
		for id, session := range h.sessions {
			if now.After(session.ExpiresAt) {
				// Clean up temp directory
				os.RemoveAll(session.TempDir)
				delete(h.sessions, id)
				h.logger.Info().Str("upload_id", id).Msg("Cleaned up expired upload session")
			}
		}
		h.sessionMu.Unlock()
	}
}

// InitChunkedUpload initializes a new chunked upload session
// POST /admin/upload/:theme/chunked/init
func (h *AdminChunkedUploadHandler) InitChunkedUpload(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Check concurrent upload limit to prevent memory exhaustion
	h.sessionMu.RLock()
	currentSessions := len(h.sessions)
	h.sessionMu.RUnlock()

	if currentSessions >= MaxConcurrentUploads {
		log.Warn().Int("current_sessions", currentSessions).Msg("Max concurrent uploads reached")
		return c.Status(fiber.StatusTooManyRequests).JSON(dto.ErrorResponse{
			Error:   "too_many_uploads",
			Message: fmt.Sprintf("Maximum concurrent uploads (%d) reached. Please try again later.", MaxConcurrentUploads),
		})
	}

	themeName := c.Params("theme")
	if err := validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	// Parse request body
	var req struct {
		FileName     string `json:"file_name"`
		TotalSize    int64  `json:"total_size"`
		ChunkSize    int64  `json:"chunk_size"`
		CustomPath   string `json:"path,omitempty"`
		FileChecksum string `json:"file_checksum,omitempty"` // Optional SHA256 of entire file for integrity verification
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate file name
	if req.FileName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_file_name",
			Message: "File name is required",
		})
	}

	// Validate file type
	if !h.validator.IsAllowedExtension(req.FileName) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_file_type",
			Message: "File type not allowed",
		})
	}

	// Reject ZIP files - use folder upload via direct-upload endpoint instead
	if h.validator.IsZipFile(req.FileName) {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "zip_not_supported",
			Message: "ZIP files are not supported. Please use folder upload via direct-upload endpoint instead.",
		})
	}

	// Validate total size (allow up to 1GB for chunked uploads)
	maxChunkedSize := int64(1024 * 1024 * 1024) // 1GB
	if req.TotalSize <= 0 || req.TotalSize > maxChunkedSize {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_size",
			Message: fmt.Sprintf("File size must be between 1 byte and %d bytes", maxChunkedSize),
		})
	}

	// Validate chunk size (5MB to 50MB)
	minChunkSize := int64(5 * 1024 * 1024)   // 5MB
	maxChunkSize := int64(50 * 1024 * 1024)  // 50MB
	if req.ChunkSize < minChunkSize || req.ChunkSize > maxChunkSize {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_chunk_size",
			Message: fmt.Sprintf("Chunk size must be between %d and %d bytes", minChunkSize, maxChunkSize),
		})
	}

	// Calculate total chunks
	totalChunks := int((req.TotalSize + req.ChunkSize - 1) / req.ChunkSize)

	// Generate unique upload ID
	uploadID := generateUploadID(themeName, req.FileName)

	// Create temp directory for chunks
	tempDir := filepath.Join(h.tempBase, "chunked_uploads", uploadID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create temp directory")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "init_failed",
			Message: "Failed to initialize upload",
		})
	}

	// Create session
	session := &ChunkedUploadSession{
		UploadID:       uploadID,
		ThemeName:      themeName,
		FileName:       req.FileName,
		TotalSize:      req.TotalSize,
		ChunkSize:      req.ChunkSize,
		TotalChunks:    totalChunks,
		UploadedChunks: make(map[int]bool),
		ChunkChecksums: make(map[int]string),
		TempDir:        tempDir,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(2 * time.Hour), // 2 hour expiry
		CustomPath:     req.CustomPath,
		FileChecksum:   req.FileChecksum,
	}

	h.sessionMu.Lock()
	h.sessions[uploadID] = session
	h.sessionMu.Unlock()

	log.Info().
		Str("upload_id", uploadID).
		Str("theme", themeName).
		Str("file", req.FileName).
		Int64("total_size", req.TotalSize).
		Int("total_chunks", totalChunks).
		Msg("Chunked upload initialized")

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"upload_id":    uploadID,
			"chunk_size":   req.ChunkSize,
			"total_chunks": totalChunks,
			"expires_at":   session.ExpiresAt,
		},
	})
}

// UploadChunk handles uploading a single chunk
// POST /admin/upload/:theme/chunked/:uploadId/chunk
func (h *AdminChunkedUploadHandler) UploadChunk(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	uploadID := c.Params("uploadId")

	if err := validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	// Get session
	h.sessionMu.RLock()
	session, exists := h.sessions[uploadID]
	h.sessionMu.RUnlock()

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "session_not_found",
			Message: "Upload session not found or expired",
		})
	}

	// Verify theme matches
	if session.ThemeName != themeName {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "theme_mismatch",
			Message: "Theme name does not match upload session",
		})
	}

	// Check if session expired
	if time.Now().After(session.ExpiresAt) {
		h.sessionMu.Lock()
		delete(h.sessions, uploadID)
		h.sessionMu.Unlock()
		os.RemoveAll(session.TempDir)
		return c.Status(fiber.StatusGone).JSON(dto.ErrorResponse{
			Error:   "session_expired",
			Message: "Upload session has expired",
		})
	}

	// Get chunk index from form
	chunkIndexStr := c.FormValue("chunk_index")
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil || chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_chunk_index",
			Message: fmt.Sprintf("Chunk index must be between 0 and %d", session.TotalChunks-1),
		})
	}

	// Get optional chunk checksum for integrity verification
	chunkChecksum := c.FormValue("chunk_checksum")

	// Get chunk file
	file, err := c.FormFile("chunk")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_chunk",
			Message: "Chunk data is required",
		})
	}

	// Validate chunk size
	expectedSize := session.ChunkSize
	if chunkIndex == session.TotalChunks-1 {
		// Last chunk may be smaller
		expectedSize = session.TotalSize - int64(chunkIndex)*session.ChunkSize
	}

	if file.Size > expectedSize {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "chunk_too_large",
			Message: "Chunk size exceeds expected size",
		})
	}

	// Open chunk file from multipart
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "chunk_read_failed",
			Message: "Failed to read chunk data",
		})
	}
	defer src.Close()

	// Create temp file for chunk - stream directly to disk to avoid memory buildup
	chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%d", chunkIndex))
	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		log.Error().Err(err).Int("chunk_index", chunkIndex).Msg("Failed to create chunk file")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "chunk_write_failed",
			Message: "Failed to save chunk",
		})
	}

	// Stream from upload to file while calculating checksum
	// This avoids loading the entire chunk into memory
	hashWriter := sha256.New()
	multiWriter := io.MultiWriter(chunkFile, hashWriter)

	written, err := io.Copy(multiWriter, src)
	chunkFile.Close() // Close immediately after writing

	if err != nil {
		os.Remove(chunkPath) // Clean up on error
		log.Error().Err(err).Int("chunk_index", chunkIndex).Msg("Failed to write chunk")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "chunk_write_failed",
			Message: "Failed to save chunk",
		})
	}

	// Verify written size matches
	if written != file.Size {
		os.Remove(chunkPath)
		log.Error().Int64("expected", file.Size).Int64("written", written).Msg("Chunk size mismatch")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "chunk_write_failed",
			Message: "Chunk data was corrupted during transfer",
		})
	}

	calculatedChecksum := hex.EncodeToString(hashWriter.Sum(nil))

	// Validate chunk checksum if provided
	if chunkChecksum != "" && chunkChecksum != calculatedChecksum {
		os.Remove(chunkPath) // Clean up invalid chunk
		log.Warn().
			Str("upload_id", uploadID).
			Int("chunk_index", chunkIndex).
			Str("expected", chunkChecksum).
			Str("calculated", calculatedChecksum).
			Msg("Chunk checksum mismatch")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "checksum_mismatch",
			Message: "Chunk data integrity check failed",
		})
	}

	// Mark chunk as uploaded and store checksum
	session.mu.Lock()
	session.UploadedChunks[chunkIndex] = true
	session.ChunkChecksums[chunkIndex] = calculatedChecksum
	uploadedCount := len(session.UploadedChunks)
	session.mu.Unlock()

	log.Debug().
		Str("upload_id", uploadID).
		Int("chunk_index", chunkIndex).
		Int("uploaded", uploadedCount).
		Int("total", session.TotalChunks).
		Str("checksum", calculatedChecksum[:16]+"..."). // Log first 16 chars
		Msg("Chunk uploaded")

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"chunk_index":     chunkIndex,
			"uploaded_chunks": uploadedCount,
			"total_chunks":    session.TotalChunks,
			"complete":        uploadedCount == session.TotalChunks,
			"chunk_checksum":  calculatedChecksum,
		},
	})
}

// CompleteChunkedUpload assembles all chunks and uploads the final file
// POST /admin/upload/:theme/chunked/:uploadId/complete
func (h *AdminChunkedUploadHandler) CompleteChunkedUpload(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	uploadID := c.Params("uploadId")

	if err := validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	// Get and remove session
	h.sessionMu.Lock()
	session, exists := h.sessions[uploadID]
	if exists {
		delete(h.sessions, uploadID)
	}
	h.sessionMu.Unlock()

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "session_not_found",
			Message: "Upload session not found or expired",
		})
	}

	// Verify theme matches
	if session.ThemeName != themeName {
		// Clean up temp dir since we're rejecting
		os.RemoveAll(session.TempDir)
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "theme_mismatch",
			Message: "Theme name does not match upload session",
		})
	}

	// Verify all chunks are uploaded
	session.mu.Lock()
	uploadedCount := len(session.UploadedChunks)
	session.mu.Unlock()

	if uploadedCount != session.TotalChunks {
		// Clean up temp dir since we're rejecting
		os.RemoveAll(session.TempDir)
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "incomplete_upload",
			Message: fmt.Sprintf("Only %d of %d chunks uploaded", uploadedCount, session.TotalChunks),
		})
	}

	// Create processing status and save to Redis
	status := &ProcessingStatus{
		UploadID:  uploadID,
		Status:    "processing",
		Progress:  0,
		Message:   "Assembling file chunks...",
		StartedAt: time.Now(),
	}

	if err := h.saveProcessingStatus(c.Context(), status); err != nil {
		log.Error().Err(err).Msg("Failed to save processing status to Redis")
		// Continue anyway - processing will still work, just status polling might fail
	}

	// Start background processing
	go h.processUploadInBackground(session, uploadID, log)

	// Return immediately with processing status
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"upload_id": uploadID,
			"status":    "processing",
			"message":   "Upload is being processed in background. Poll status endpoint for updates.",
		},
	})
}

// GetProcessingStatus returns the current status of a background upload processing
// GET /admin/upload/status/:uploadId
func (h *AdminChunkedUploadHandler) GetProcessingStatus(c *fiber.Ctx) error {
	uploadID := c.Params("uploadId")

	status, err := h.getProcessingStatus(c.Context(), uploadID)
	if err != nil {
		h.logger.Error().Err(err).Str("upload_id", uploadID).Msg("Failed to get processing status from Redis")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "status_error",
			Message: "Failed to retrieve processing status",
		})
	}

	if status == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "status_not_found",
			Message: "Processing status not found. Upload may have completed and been cleaned up.",
		})
	}

	response := fiber.Map{
		"upload_id":  status.UploadID,
		"status":     status.Status,
		"progress":   status.Progress,
		"message":    status.Message,
		"started_at": status.StartedAt,
	}
	if status.Error != "" {
		response["error"] = status.Error
	}
	if status.Result != nil {
		response["result"] = json.RawMessage(status.Result)
	}
	if status.CompletedAt != nil {
		response["completed_at"] = status.CompletedAt
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// saveProcessingStatus saves processing status to Redis (with in-memory fallback)
func (h *AdminChunkedUploadHandler) saveProcessingStatus(ctx context.Context, status *ProcessingStatus) error {
	// Try Redis first
	if h.redis != nil {
		data, err := json.Marshal(status)
		if err != nil {
			return fmt.Errorf("failed to marshal status: %w", err)
		}
		key := ProcessingStatusKeyPrefix + status.UploadID
		if err := h.redis.Set(ctx, key, string(data), ProcessingStatusTTL); err == nil {
			return nil
		}
		// Fall through to in-memory if Redis fails
	}

	// Fallback to in-memory storage
	h.processingMu.Lock()
	// Deep copy the status to avoid race conditions
	statusCopy := *status
	h.processing[status.UploadID] = &statusCopy
	h.processingMu.Unlock()
	return nil
}

// getProcessingStatus retrieves processing status from Redis (with in-memory fallback)
func (h *AdminChunkedUploadHandler) getProcessingStatus(ctx context.Context, uploadID string) (*ProcessingStatus, error) {
	// Try Redis first
	if h.redis != nil {
		key := ProcessingStatusKeyPrefix + uploadID
		data, err := h.redis.Get(ctx, key)
		if err == nil && data != "" {
			var status ProcessingStatus
			if err := json.Unmarshal([]byte(data), &status); err == nil {
				return &status, nil
			}
		}
		// Fall through to in-memory if Redis fails or not found
	}

	// Fallback to in-memory storage
	h.processingMu.RLock()
	status, exists := h.processing[uploadID]
	h.processingMu.RUnlock()

	if !exists {
		return nil, nil
	}

	// Return a copy to avoid race conditions
	statusCopy := *status
	return &statusCopy, nil
}

// updateProcessingStatus updates processing status in Redis (with in-memory fallback)
func (h *AdminChunkedUploadHandler) updateProcessingStatus(ctx context.Context, uploadID string, updateFn func(*ProcessingStatus)) error {
	status, err := h.getProcessingStatus(ctx, uploadID)
	if err != nil {
		return err
	}
	if status == nil {
		status = &ProcessingStatus{
			UploadID:  uploadID,
			Status:    "processing",
			StartedAt: time.Now(),
		}
	}

	updateFn(status)

	return h.saveProcessingStatus(ctx, status)
}

// processUploadInBackground handles the actual file assembly and upload
func (h *AdminChunkedUploadHandler) processUploadInBackground(
	session *ChunkedUploadSession,
	uploadID string,
	log *logger.Logger,
) {
	// Use background context since the HTTP request context is already done
	ctx := context.Background()

	// Ensure temp dir is cleaned up after completion
	defer os.RemoveAll(session.TempDir)

	updateStatus := func(progress int, message string) {
		if err := h.updateProcessingStatus(ctx, uploadID, func(s *ProcessingStatus) {
			s.Progress = progress
			s.Message = message
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to update processing status")
		}
	}

	failWithError := func(errMsg string) {
		now := time.Now()
		if err := h.updateProcessingStatus(ctx, uploadID, func(s *ProcessingStatus) {
			s.Status = "failed"
			s.Error = errMsg
			s.CompletedAt = &now
		}); err != nil {
			log.Error().Err(err).Msg("Failed to update processing status on failure")
		}
	}

	completeWithResult := func(result interface{}) {
		now := time.Now()
		resultJSON, err := json.Marshal(result)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal result")
			resultJSON = []byte("{}")
		}
		if err := h.updateProcessingStatus(ctx, uploadID, func(s *ProcessingStatus) {
			s.Status = "completed"
			s.Progress = 100
			s.Message = "Upload completed successfully"
			s.Result = resultJSON
			s.CompletedAt = &now
		}); err != nil {
			log.Error().Err(err).Msg("Failed to update processing status on completion")
		}
	}

	// Assemble chunks into a temporary file (streaming, not in memory)
	updateStatus(10, "Assembling file chunks...")
	assembledFilePath := filepath.Join(session.TempDir, "assembled_file")
	assembledFile, err := os.Create(assembledFilePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create assembled file")
		failWithError("Failed to create assembled file")
		return
	}

	// Use a hash writer to calculate checksum while writing
	hashWriter := sha256.New()
	multiWriter := io.MultiWriter(assembledFile, hashWriter)

	var totalWritten int64
	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			assembledFile.Close()
			log.Error().Err(err).Int("chunk_index", i).Msg("Failed to open chunk file")
			failWithError("Failed to assemble file chunks")
			return
		}

		written, err := io.Copy(multiWriter, chunkFile)
		chunkFile.Close()
		if err != nil {
			assembledFile.Close()
			log.Error().Err(err).Int("chunk_index", i).Msg("Failed to copy chunk data")
			failWithError("Failed to assemble file chunks")
			return
		}
		totalWritten += written

		// Update progress (10-40% for assembly)
		progress := 10 + (30 * (i + 1) / session.TotalChunks)
		updateStatus(progress, fmt.Sprintf("Assembling chunks (%d/%d)...", i+1, session.TotalChunks))
	}
	assembledFile.Close()

	// Verify final size
	if totalWritten != session.TotalSize {
		failWithError(fmt.Sprintf("Assembled size %d does not match expected %d", totalWritten, session.TotalSize))
		return
	}

	updateStatus(45, "Verifying file integrity...")

	// Calculate file checksum
	calculatedFileChecksum := hex.EncodeToString(hashWriter.Sum(nil))

	// Verify file checksum if provided during init
	if session.FileChecksum != "" && session.FileChecksum != calculatedFileChecksum {
		log.Warn().
			Str("upload_id", session.UploadID).
			Str("expected", session.FileChecksum).
			Str("calculated", calculatedFileChecksum).
			Msg("File checksum mismatch")
		failWithError("Assembled file integrity check failed")
		return
	}

	// Open assembled file for validation
	assembledFile, err = os.Open(assembledFilePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open assembled file for validation")
		failWithError("Failed to open assembled file")
		return
	}

	// Validate magic bytes (only read first 512 bytes)
	magicBytes := make([]byte, 512)
	n, _ := assembledFile.Read(magicBytes)
	if err := h.validator.ValidateMagicBytes(session.FileName, bytes.NewReader(magicBytes[:n])); err != nil {
		assembledFile.Close()
		log.Warn().Err(err).Str("filename", session.FileName).Msg("Magic bytes validation failed")
		failWithError("File content does not match file type")
		return
	}
	assembledFile.Close()

	updateStatus(50, "Uploading to cloud storage...")

	// Upload file - stream from file
	result, err := h.processFileUploadBackground(session, assembledFilePath, calculatedFileChecksum, updateStatus, log)
	if err != nil {
		failWithError(err.Error())
		return
	}
	completeWithResult(result)
}

// processFileUploadBackground handles uploading a regular file in background
func (h *AdminChunkedUploadHandler) processFileUploadBackground(
	session *ChunkedUploadSession,
	assembledFilePath string,
	calculatedFileChecksum string,
	updateStatus func(int, string),
	log *logger.Logger,
) (interface{}, error) {
	// Open file for upload
	file, err := os.Open(assembledFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for upload")
	}
	defer file.Close()

	// Determine upload path
	var objectName string
	if session.CustomPath != "" {
		objectName = session.CustomPath
	} else {
		objectName = session.FileName
	}

	updateStatus(60, "Uploading file to storage...")

	// Upload to storage
	url, err := h.storage.UploadFile(
		context.Background(),
		session.ThemeName,
		objectName,
		file,
		session.TotalSize,
		"",
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload file to storage")
		return nil, fmt.Errorf("failed to upload file to storage")
	}

	updateStatus(100, "Upload completed")

	log.Info().
		Str("upload_id", session.UploadID).
		Str("url", url).
		Int64("size", session.TotalSize).
		Str("checksum", calculatedFileChecksum[:16]+"...").
		Msg("Chunked upload completed successfully")

	return fiber.Map{
		"url":           url,
		"filename":      session.FileName,
		"size":          session.TotalSize,
		"file_checksum": calculatedFileChecksum,
	}, nil
}

// GetUploadStatus returns the status of a chunked upload
// GET /admin/upload/:theme/chunked/:uploadId/status
func (h *AdminChunkedUploadHandler) GetUploadStatus(c *fiber.Ctx) error {
	themeName := c.Params("theme")
	uploadID := c.Params("uploadId")

	if err := validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	h.sessionMu.RLock()
	session, exists := h.sessions[uploadID]
	h.sessionMu.RUnlock()

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "session_not_found",
			Message: "Upload session not found or expired",
		})
	}

	session.mu.Lock()
	uploadedChunks := make([]int, 0, len(session.UploadedChunks))
	for idx := range session.UploadedChunks {
		uploadedChunks = append(uploadedChunks, idx)
	}
	sort.Ints(uploadedChunks)
	session.mu.Unlock()

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"upload_id":       uploadID,
			"file_name":       session.FileName,
			"total_size":      session.TotalSize,
			"chunk_size":      session.ChunkSize,
			"total_chunks":    session.TotalChunks,
			"uploaded_chunks": uploadedChunks,
			"progress":        float64(len(uploadedChunks)) / float64(session.TotalChunks) * 100,
			"expires_at":      session.ExpiresAt,
		},
	})
}

// AbortChunkedUpload cancels an ongoing chunked upload
// DELETE /admin/upload/:theme/chunked/:uploadId
func (h *AdminChunkedUploadHandler) AbortChunkedUpload(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	themeName := c.Params("theme")
	uploadID := c.Params("uploadId")

	if err := validateThemeName(themeName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_theme",
			Message: err.Error(),
		})
	}

	h.sessionMu.Lock()
	session, exists := h.sessions[uploadID]
	if exists {
		delete(h.sessions, uploadID)
	}
	h.sessionMu.Unlock()

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "session_not_found",
			Message: "Upload session not found or expired",
		})
	}

	// Clean up temp directory
	os.RemoveAll(session.TempDir)

	log.Info().Str("upload_id", uploadID).Msg("Chunked upload aborted")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Upload aborted",
	})
}

// generateUploadID creates a unique upload ID
func generateUploadID(themeName, fileName string) string {
	data := fmt.Sprintf("%s_%s_%d_%d", themeName, fileName, time.Now().UnixNano(), time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// validateThemeName validates the theme name (shared with regular upload handler)
func validateThemeName(themeName string) error {
	if themeName == "" {
		return fmt.Errorf("theme name is required")
	}
	if strings.Contains(themeName, "..") || strings.Contains(themeName, "/") || strings.Contains(themeName, "\\") {
		return fmt.Errorf("invalid theme name: path traversal not allowed")
	}
	for _, r := range themeName {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return fmt.Errorf("invalid theme name: only alphanumeric, dash, and underscore allowed")
		}
	}
	if len(themeName) > 100 {
		return fmt.Errorf("theme name too long (max 100 characters)")
	}
	return nil
}
