package storage

import (
	"context"
	"io"
	"time"
)

// FileInfo represents information about a stored file
type FileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	URL          string    `json:"url"`
}

// Storage defines the interface for file storage operations
type Storage interface {
	// UploadFile uploads a file to storage with known size
	UploadFile(ctx context.Context, themeName, fileName string, reader io.Reader, size int64, contentType string) (string, error)
	// UploadFileStreaming uploads a file to storage without knowing size upfront
	// Uses multipart upload internally for efficient streaming of large files
	UploadFileStreaming(ctx context.Context, themeName, fileName string, reader io.Reader, contentType string) (string, error)
	// DeleteFile deletes a file from storage
	DeleteFile(ctx context.Context, themeName, fileName string) error
	// ListFiles lists all files in a theme folder
	ListFiles(ctx context.Context, themeName string) ([]FileInfo, error)
	// DeleteTheme deletes all files in a theme folder
	DeleteTheme(ctx context.Context, themeName string) error
	// GetPublicURL returns the public URL for a file
	GetPublicURL(themeName, fileName string) string
	// GetBaseURL returns the base URL for a theme folder (publicURL/bucket/themeName)
	GetBaseURL(themeName string) string
	// CreateFolder creates an empty folder/prefix in storage
	CreateFolder(ctx context.Context, themeName string) error
	// RenameFolder renames a folder by copying all files to new location and deleting old ones
	RenameFolder(ctx context.Context, oldThemeName, newThemeName string) error
	// FolderExists checks if a folder/prefix exists in storage
	FolderExists(ctx context.Context, themeName string) (bool, error)
	// GeneratePresignedUploadURL generates a presigned URL for direct client upload
	// Returns the upload URL and the final public URL after upload
	GeneratePresignedUploadURL(ctx context.Context, themeName, fileName, contentType string, expiryMinutes int) (*PresignedUploadInfo, error)
	// FileExists checks if a specific file exists in storage
	FileExists(ctx context.Context, themeName, fileName string) (bool, error)
}

// PresignedUploadInfo contains information for direct client upload
type PresignedUploadInfo struct {
	UploadURL   string            `json:"upload_url"`   // Presigned URL for PUT request
	PublicURL   string            `json:"public_url"`   // Final public URL after upload
	Method      string            `json:"method"`       // HTTP method to use (PUT)
	Headers     map[string]string `json:"headers"`      // Required headers for upload
	ExpiresAt   time.Time         `json:"expires_at"`   // When the presigned URL expires
	ContentType string            `json:"content_type"` // Content type for upload
}
