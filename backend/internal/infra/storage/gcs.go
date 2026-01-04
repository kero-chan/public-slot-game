package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	gcs "cloud.google.com/go/storage"
	"github.com/slotmachine/backend/internal/config"
	"google.golang.org/api/iterator"
)

// GCSStorage handles file storage operations with Google Cloud Storage
// Implements the Storage interface
type GCSStorage struct {
	client     *gcs.Client
	bucketName string
	publicURL  string
}

// Ensure GCSStorage implements Storage interface
var _ Storage = (*GCSStorage)(nil)

// NewGCSStorage creates a new GCS storage client
func NewGCSStorage(cfg *config.StorageConfig) (*GCSStorage, error) {
	ctx := context.Background()

	// Create GCS client (uses Application Default Credentials in production)
	client, err := gcs.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	storage := &GCSStorage{
		client:     client,
		bucketName: cfg.BucketName,
		publicURL:  cfg.PublicURL,
	}

	// Verify bucket exists
	bucket := client.Bucket(cfg.BucketName)
	_, err = bucket.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to access bucket %s: %w", cfg.BucketName, err)
	}

	return storage, nil
}

// UploadFile uploads a file to GCS with known size
func (s *GCSStorage) UploadFile(ctx context.Context, themeName, fileName string, reader io.Reader, size int64, contentType string) (string, error) {
	// Construct object path: {theme}/{filename}
	objectName := filepath.Join(themeName, fileName)

	// Determine content type if not provided
	if contentType == "" {
		contentType = getContentType(fileName)
	}

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(objectName)
	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType

	// Copy the file content
	if _, err := io.Copy(writer, reader); err != nil {
		writer.Close()
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Return public URL
	url := fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, objectName)
	return url, nil
}

// UploadFileStreaming uploads a file to GCS without knowing size upfront
// GCS Writer already supports streaming - it uses resumable uploads internally
func (s *GCSStorage) UploadFileStreaming(ctx context.Context, themeName, fileName string, reader io.Reader, contentType string) (string, error) {
	// Construct object path: {theme}/{filename}
	objectName := filepath.Join(themeName, fileName)

	// Determine content type if not provided
	if contentType == "" {
		contentType = getContentType(fileName)
	}

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(objectName)
	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType
	// GCS automatically uses resumable uploads for objects > 16MB
	// ChunkSize controls the upload buffer size (default 16MB)
	writer.ChunkSize = 16 * 1024 * 1024 // 16MB chunks for streaming

	// Stream the file content directly to GCS
	if _, err := io.Copy(writer, reader); err != nil {
		writer.Close()
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Return public URL
	url := fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, objectName)
	return url, nil
}

// DeleteFile deletes a file from GCS
func (s *GCSStorage) DeleteFile(ctx context.Context, themeName, fileName string) error {
	objectName := filepath.Join(themeName, fileName)
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(objectName)

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// ListFiles lists all files in a theme folder
func (s *GCSStorage) ListFiles(ctx context.Context, themeName string) ([]FileInfo, error) {
	prefix := themeName + "/"
	var files []FileInfo

	bucket := s.client.Bucket(s.bucketName)
	it := bucket.Objects(ctx, &gcs.Query{
		Prefix: prefix,
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		// Skip directories (objects ending with /)
		if strings.HasSuffix(attrs.Name, "/") {
			continue
		}

		files = append(files, FileInfo{
			Name:         strings.TrimPrefix(attrs.Name, prefix),
			Size:         attrs.Size,
			LastModified: attrs.Updated,
			URL:          fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, attrs.Name),
		})
	}

	return files, nil
}

// DeleteTheme deletes all files in a theme folder
func (s *GCSStorage) DeleteTheme(ctx context.Context, themeName string) error {
	prefix := themeName + "/"
	bucket := s.client.Bucket(s.bucketName)

	it := bucket.Objects(ctx, &gcs.Query{
		Prefix: prefix,
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete object %s: %w", attrs.Name, err)
		}
	}

	return nil
}

// GetPublicURL returns the public URL for a file
func (s *GCSStorage) GetPublicURL(themeName, fileName string) string {
	objectName := filepath.Join(themeName, fileName)
	return fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, objectName)
}

// GetBaseURL returns the base URL for a theme folder
func (s *GCSStorage) GetBaseURL(themeName string) string {
	return fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, themeName)
}

// CreateFolder creates an empty folder/prefix in storage
// In GCS, folders are virtual - we create a placeholder object
func (s *GCSStorage) CreateFolder(ctx context.Context, themeName string) error {
	objectName := themeName + "/.folder"
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(objectName)
	writer := obj.NewWriter(ctx)
	writer.ContentType = "application/x-directory"

	if _, err := writer.Write([]byte{}); err != nil {
		writer.Close()
		return fmt.Errorf("failed to create folder: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}

// RenameFolder renames a folder by copying all files to new location and deleting old ones
func (s *GCSStorage) RenameFolder(ctx context.Context, oldThemeName, newThemeName string) error {
	oldPrefix := oldThemeName + "/"
	newPrefix := newThemeName + "/"

	bucket := s.client.Bucket(s.bucketName)
	it := bucket.Objects(ctx, &gcs.Query{
		Prefix: oldPrefix,
	})

	var objectsToDelete []string

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		// Calculate new object name
		relativePath := strings.TrimPrefix(attrs.Name, oldPrefix)
		newObjectName := newPrefix + relativePath

		// Copy object to new location
		src := bucket.Object(attrs.Name)
		dst := bucket.Object(newObjectName)
		if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
			return fmt.Errorf("failed to copy object %s to %s: %w", attrs.Name, newObjectName, err)
		}

		objectsToDelete = append(objectsToDelete, attrs.Name)
	}

	// Delete old objects
	for _, objectName := range objectsToDelete {
		if err := bucket.Object(objectName).Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete old object %s: %w", objectName, err)
		}
	}

	return nil
}

// FolderExists checks if a folder/prefix exists in storage
func (s *GCSStorage) FolderExists(ctx context.Context, themeName string) (bool, error) {
	prefix := themeName + "/"

	bucket := s.client.Bucket(s.bucketName)
	it := bucket.Objects(ctx, &gcs.Query{
		Prefix: prefix,
	})

	_, err := it.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to list objects: %w", err)
	}

	return true, nil
}

// Close closes the GCS client
func (s *GCSStorage) Close() error {
	return s.client.Close()
}

// GeneratePresignedUploadURL generates a signed URL for direct client upload to GCS
func (s *GCSStorage) GeneratePresignedUploadURL(ctx context.Context, themeName, fileName, contentType string, expiryMinutes int) (*PresignedUploadInfo, error) {
	objectName := filepath.Join(themeName, fileName)

	// Set default expiry if not specified
	if expiryMinutes <= 0 {
		expiryMinutes = 15 // Default 15 minutes
	}
	expiry := time.Duration(expiryMinutes) * time.Minute

	// Determine content type if not provided
	if contentType == "" {
		contentType = getContentType(fileName)
	}

	// Build headers
	headers := []string{"Content-Type:" + contentType}

	// Add Cache-Control header for non-JSON files (1 year cache)
	// JSON files should not be cached as they may contain dynamic config
	ext := strings.ToLower(filepath.Ext(fileName))
	cacheControl := ""
	if ext != ".json" {
		cacheControl = "public, max-age=31536000, immutable"
		headers = append(headers, "Cache-Control:"+cacheControl)
	}

	// Generate signed URL for PUT operation
	opts := &gcs.SignedURLOptions{
		Method:      "PUT",
		Expires:     time.Now().Add(expiry),
		ContentType: contentType,
		Headers:     headers,
	}

	bucket := s.client.Bucket(s.bucketName)
	signedURL, err := bucket.SignedURL(objectName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed URL: %w", err)
	}

	// Build response headers map
	responseHeaders := map[string]string{"Content-Type": contentType}
	if cacheControl != "" {
		responseHeaders["Cache-Control"] = cacheControl
	}

	return &PresignedUploadInfo{
		UploadURL:   signedURL,
		PublicURL:   fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, objectName),
		Method:      "PUT",
		Headers:     responseHeaders,
		ExpiresAt:   time.Now().Add(expiry),
		ContentType: contentType,
	}, nil
}

// FileExists checks if a specific file exists in GCS
func (s *GCSStorage) FileExists(ctx context.Context, themeName, fileName string) (bool, error) {
	objectName := filepath.Join(themeName, fileName)

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(objectName)

	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == gcs.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}
