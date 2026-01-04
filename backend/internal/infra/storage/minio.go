package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/slotmachine/backend/internal/config"
)

// MinIOStorage handles file storage operations with MinIO/S3
// Implements the Storage interface
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	publicURL  string
}

// Ensure MinIOStorage implements Storage interface
var _ Storage = (*MinIOStorage)(nil)

// NewMinIOStorage creates a new MinIO storage client
func NewMinIOStorage(cfg *config.StorageConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	storage := &MinIOStorage{
		client:     client,
		bucketName: cfg.BucketName,
		publicURL:  cfg.PublicURL,
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		// Set bucket policy to allow public read
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, cfg.BucketName)

		err = client.SetBucketPolicy(ctx, cfg.BucketName, policy)
		if err != nil {
			return nil, fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}

	return storage, nil
}

// UploadFile uploads a file to the storage with known size
func (s *MinIOStorage) UploadFile(ctx context.Context, themeName, fileName string, reader io.Reader, size int64, contentType string) (string, error) {
	// Construct object path: {theme}/{filename}
	objectName := filepath.Join(themeName, fileName)

	// Determine content type if not provided
	if contentType == "" {
		contentType = getContentType(fileName)
	}

	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return public URL
	url := fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, objectName)
	return url, nil
}

// UploadFileStreaming uploads a file to storage without knowing size upfront
// Uses multipart upload internally (MinIO automatically uses multipart when size is -1)
func (s *MinIOStorage) UploadFileStreaming(ctx context.Context, themeName, fileName string, reader io.Reader, contentType string) (string, error) {
	// Construct object path: {theme}/{filename}
	objectName := filepath.Join(themeName, fileName)

	// Determine content type if not provided
	if contentType == "" {
		contentType = getContentType(fileName)
	}

	// Use size = -1 to enable automatic multipart upload
	// MinIO client will automatically chunk the upload into parts
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
		// PartSize will default to 16MB, which is good for streaming
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return public URL
	url := fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, objectName)
	return url, nil
}

// DeleteFile deletes a file from storage
func (s *MinIOStorage) DeleteFile(ctx context.Context, themeName, fileName string) error {
	objectName := filepath.Join(themeName, fileName)
	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// ListFiles lists all files in a theme folder
func (s *MinIOStorage) ListFiles(ctx context.Context, themeName string) ([]FileInfo, error) {
	prefix := themeName + "/"
	var files []FileInfo

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		// Skip directories
		if strings.HasSuffix(object.Key, "/") {
			continue
		}

		files = append(files, FileInfo{
			Name:         strings.TrimPrefix(object.Key, prefix),
			Size:         object.Size,
			LastModified: object.LastModified,
			URL:          fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, object.Key),
		})
	}

	return files, nil
}

// DeleteTheme deletes all files in a theme folder
func (s *MinIOStorage) DeleteTheme(ctx context.Context, themeName string) error {
	prefix := themeName + "/"

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return fmt.Errorf("failed to list objects: %w", object.Err)
		}

		err := s.client.RemoveObject(ctx, s.bucketName, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete object %s: %w", object.Key, err)
		}
	}

	return nil
}

// GetPublicURL returns the public URL for a file
func (s *MinIOStorage) GetPublicURL(themeName, fileName string) string {
	objectName := filepath.Join(themeName, fileName)
	return fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, objectName)
}

// GetBaseURL returns the base URL for a theme folder
func (s *MinIOStorage) GetBaseURL(themeName string) string {
	return fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, themeName)
}

// CreateFolder creates an empty folder/prefix in storage
// In S3/MinIO, folders are virtual - we create a placeholder object
func (s *MinIOStorage) CreateFolder(ctx context.Context, themeName string) error {
	// Create a placeholder file to ensure the "folder" exists
	// MinIO/S3 doesn't have real folders, but we can create a .folder marker
	objectName := themeName + "/.folder"
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, strings.NewReader(""), 0, minio.PutObjectOptions{
		ContentType: "application/x-directory",
	})
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}

// RenameFolder renames a folder by copying all files to new location and deleting old ones
func (s *MinIOStorage) RenameFolder(ctx context.Context, oldThemeName, newThemeName string) error {
	oldPrefix := oldThemeName + "/"
	newPrefix := newThemeName + "/"

	// List all objects with old prefix
	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    oldPrefix,
		Recursive: true,
	})

	var objectsToDelete []string

	for object := range objectCh {
		if object.Err != nil {
			return fmt.Errorf("failed to list objects: %w", object.Err)
		}

		// Calculate new object name
		relativePath := strings.TrimPrefix(object.Key, oldPrefix)
		newObjectName := newPrefix + relativePath

		// Copy object to new location
		src := minio.CopySrcOptions{
			Bucket: s.bucketName,
			Object: object.Key,
		}
		dst := minio.CopyDestOptions{
			Bucket: s.bucketName,
			Object: newObjectName,
		}

		_, err := s.client.CopyObject(ctx, dst, src)
		if err != nil {
			return fmt.Errorf("failed to copy object %s to %s: %w", object.Key, newObjectName, err)
		}

		objectsToDelete = append(objectsToDelete, object.Key)
	}

	// Delete old objects
	for _, objectName := range objectsToDelete {
		err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete old object %s: %w", objectName, err)
		}
	}

	return nil
}

// FolderExists checks if a folder/prefix exists in storage
func (s *MinIOStorage) FolderExists(ctx context.Context, themeName string) (bool, error) {
	prefix := themeName + "/"

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:  prefix,
		MaxKeys: 1,
	})

	for object := range objectCh {
		if object.Err != nil {
			return false, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		// Found at least one object with this prefix
		return true, nil
	}

	return false, nil
}

// GeneratePresignedUploadURL generates a presigned URL for direct client upload to MinIO
func (s *MinIOStorage) GeneratePresignedUploadURL(ctx context.Context, themeName, fileName, contentType string, expiryMinutes int) (*PresignedUploadInfo, error) {
	objectName := filepath.Join(themeName, fileName)

	// Set default expiry if not specified
	if expiryMinutes <= 0 {
		expiryMinutes = 15 // Default 15 minutes
	}
	expiry := time.Duration(expiryMinutes) * time.Minute

	// Generate presigned PUT URL
	presignedURL, err := s.client.PresignedPutObject(ctx, s.bucketName, objectName, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Determine content type if not provided
	if contentType == "" {
		contentType = getContentType(fileName)
	}

	// Build headers
	headers := map[string]string{"Content-Type": contentType}

	// Add Cache-Control header for non-JSON files (1 year cache)
	// JSON files should not be cached as they may contain dynamic config
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".json" {
		headers["Cache-Control"] = "public, max-age=31536000, immutable"
	}

	return &PresignedUploadInfo{
		UploadURL:   presignedURL.String(),
		PublicURL:   fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucketName, objectName),
		Method:      "PUT",
		Headers:     headers,
		ExpiresAt:   time.Now().Add(expiry),
		ContentType: contentType,
	}, nil
}

// FileExists checks if a specific file exists in storage
func (s *MinIOStorage) FileExists(ctx context.Context, themeName, fileName string) (bool, error) {
	objectName := filepath.Join(themeName, fileName)

	_, err := s.client.StatObject(ctx, s.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		// Check if it's a "not found" error
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

