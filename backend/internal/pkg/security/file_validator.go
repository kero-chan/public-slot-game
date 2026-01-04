package security

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

// File validation errors
var (
	ErrFileTypeNotAllowed     = errors.New("file type not allowed")
	ErrFileSizeTooLarge       = errors.New("file size exceeds limit")
	ErrInvalidMagicBytes      = errors.New("file content does not match extension")
	ErrPathTraversal          = errors.New("path traversal detected")
	ErrInvalidFilename        = errors.New("invalid filename")
	ErrZipBomb                = errors.New("zip bomb detected: extraction ratio too high")
	ErrTooManyFiles           = errors.New("too many files in archive")
	ErrNestedArchive          = errors.New("nested archives not allowed")
	ErrSymlinkNotAllowed      = errors.New("symbolic links not allowed in archive")
	ErrExtractionSizeTooLarge = errors.New("total extracted size exceeds limit")
)

// FileValidatorConfig holds configuration for file validation
type FileValidatorConfig struct {
	MaxFileSize         int64   // Maximum single file size in bytes (default: 50MB)
	MaxZipSize          int64   // Maximum ZIP file size in bytes (default: 500MB)
	MaxTotalExtractSize int64   // Maximum total extracted size (default: 500MB)
	MaxFilesInZip       int     // Maximum files in a ZIP (default: 1000)
	MaxCompressionRatio float64 // Maximum compression ratio (default: 100)
}

// DefaultConfig returns default configuration
func DefaultConfig() *FileValidatorConfig {
	return &FileValidatorConfig{
		MaxFileSize:         50 * 1024 * 1024,  // 50MB
		MaxZipSize:          500 * 1024 * 1024, // 100MB
		MaxTotalExtractSize: 500 * 1024 * 1024, // 500MB
		MaxFilesInZip:       1000,
		MaxCompressionRatio: 100, // Suspicious if compressed size is <1% of original
	}
}

// FileValidator provides secure file validation
type FileValidator struct {
	config *FileValidatorConfig
}

// NewFileValidator creates a new file validator
func NewFileValidator(config *FileValidatorConfig) *FileValidator {
	if config == nil {
		config = DefaultConfig()
	}
	return &FileValidator{config: config}
}

// AllowedExtensions maps extensions to their expected MIME magic bytes
var AllowedExtensions = map[string][]MagicSignature{
	".png":  {{Bytes: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}}},
	".jpg":  {{Bytes: []byte{0xFF, 0xD8, 0xFF}}},
	".jpeg": {{Bytes: []byte{0xFF, 0xD8, 0xFF}}},
	".gif":  {{Bytes: []byte{0x47, 0x49, 0x46, 0x38}}}, // GIF8
	".webp": {{Bytes: []byte{0x52, 0x49, 0x46, 0x46}, Offset: 0}, {Bytes: []byte{0x57, 0x45, 0x42, 0x50}, Offset: 8}},
	".svg":  {{Bytes: []byte("<?xml")}, {Bytes: []byte("<svg")}, {Bytes: []byte("<!DOCTYPE svg")}},
	".json": {{Bytes: []byte("{")}, {Bytes: []byte("[")}, {Bytes: []byte(" {")}, {Bytes: []byte(" [")}, {Bytes: []byte("\n{")}, {Bytes: []byte("\n[")}},
	".mp3":  {{Bytes: []byte{0xFF, 0xFB}}, {Bytes: []byte{0xFF, 0xFA}}, {Bytes: []byte{0xFF, 0xF3}}, {Bytes: []byte{0xFF, 0xF2}}, {Bytes: []byte("ID3")}},
	".wav":  {{Bytes: []byte{0x52, 0x49, 0x46, 0x46}}},                  // RIFF
	".ogg":  {{Bytes: []byte{0x4F, 0x67, 0x67, 0x53}}},                  // OggS
	".m4a":  {{Bytes: []byte("ftyp"), Offset: 4}},                       // ftyp at offset 4 (after 4-byte size)
	".aac":  {{Bytes: []byte{0xFF, 0xF1}}, {Bytes: []byte{0xFF, 0xF9}}}, // ADTS header
	".flac": {{Bytes: []byte("fLaC")}},                                  // FLAC magic
	".mp4":  {{Bytes: []byte("ftyp"), Offset: 4}},                       // ftyp at offset 4 (after 4-byte size)
	".webm": {{Bytes: []byte{0x1A, 0x45, 0xDF, 0xA3}}},                  // EBML header
	".zip":  {{Bytes: []byte{0x50, 0x4B, 0x03, 0x04}}, {Bytes: []byte{0x50, 0x4B, 0x05, 0x06}}, {Bytes: []byte{0x50, 0x4B, 0x07, 0x08}}},
}

// MagicSignature represents file magic bytes
type MagicSignature struct {
	Bytes  []byte
	Offset int
}

// IsAllowedExtension checks if the extension is in the allowed list
func (v *FileValidator) IsAllowedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, ok := AllowedExtensions[ext]
	return ok
}

// IsZipFile checks if the extension indicates a ZIP file
func (v *FileValidator) IsZipFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".zip"
}

// ValidateFileSize checks if file size is within limits
func (v *FileValidator) ValidateFileSize(size int64, isZip bool) error {
	maxSize := v.config.MaxFileSize
	if isZip {
		maxSize = v.config.MaxZipSize
	}
	if size > maxSize {
		return fmt.Errorf("%w: %d bytes exceeds limit of %d bytes", ErrFileSizeTooLarge, size, maxSize)
	}
	return nil
}

// ValidateMagicBytes validates file content matches its extension
func (v *FileValidator) ValidateMagicBytes(filename string, reader io.Reader) error {
	ext := strings.ToLower(filepath.Ext(filename))
	signatures, ok := AllowedExtensions[ext]
	if !ok {
		return ErrFileTypeNotAllowed
	}

	// Read enough bytes to check all signatures
	maxOffset := 0
	maxLen := 0
	for _, sig := range signatures {
		if sig.Offset+len(sig.Bytes) > maxOffset+maxLen {
			maxOffset = sig.Offset
			maxLen = len(sig.Bytes)
		}
	}
	headerSize := maxOffset + maxLen
	if headerSize < 32 {
		headerSize = 32 // Read at least 32 bytes
	}

	header := make([]byte, headerSize)
	n, err := io.ReadFull(reader, header)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("failed to read file header: %w", err)
	}
	header = header[:n]

	// Check if any signature matches
	for _, sig := range signatures {
		if sig.Offset+len(sig.Bytes) <= len(header) {
			if bytes.HasPrefix(header[sig.Offset:], sig.Bytes) {
				return nil
			}
		}
	}

	// Special handling for text-based files (SVG, JSON)
	if ext == ".svg" || ext == ".json" {
		// Allow whitespace/BOM at the beginning
		trimmed := bytes.TrimLeft(header, " \t\r\n\xef\xbb\xbf")
		for _, sig := range signatures {
			if bytes.HasPrefix(trimmed, sig.Bytes) {
				return nil
			}
		}
	}

	return fmt.Errorf("%w: %s", ErrInvalidMagicBytes, ext)
}

// SanitizeFilename removes dangerous characters and prevents path traversal
func (v *FileValidator) SanitizeFilename(filename string) (string, error) {
	// Check for path traversal
	if strings.Contains(filename, "..") {
		return "", ErrPathTraversal
	}

	// Normalize path separators
	filename = filepath.ToSlash(filename)

	// Remove leading slashes
	filename = strings.TrimLeft(filename, "/")

	// Check for absolute paths (Windows)
	if len(filename) >= 2 && filename[1] == ':' {
		return "", ErrPathTraversal
	}

	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Validate filename characters (allow alphanumeric, dash, underscore, dot, slash)
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-./]+$`)
	if !validPattern.MatchString(filename) {
		// Try to sanitize by replacing invalid characters
		sanitized := regexp.MustCompile(`[^a-zA-Z0-9_\-./]`).ReplaceAllString(filename, "_")
		if sanitized == "" || sanitized == "_" {
			return "", ErrInvalidFilename
		}
		filename = sanitized
	}

	// Ensure filename is not empty after sanitization
	if filename == "" || filename == "." {
		return "", ErrInvalidFilename
	}

	return filename, nil
}

// ValidateZipEntry validates a single ZIP entry for security issues
func (v *FileValidator) ValidateZipEntry(name string, uncompressedSize uint64, isSymlink bool) error {
	// Check for symlinks
	if isSymlink {
		return ErrSymlinkNotAllowed
	}

	// Check for path traversal
	if _, err := v.SanitizeFilename(name); err != nil {
		return err
	}

	// Check for nested archives
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".zip" || ext == ".tar" || ext == ".gz" || ext == ".rar" || ext == ".7z" {
		return ErrNestedArchive
	}

	return nil
}

// CheckCompressionRatio checks for potential zip bombs
func (v *FileValidator) CheckCompressionRatio(compressedSize, uncompressedSize int64) error {
	if compressedSize == 0 {
		return nil // Empty file
	}
	ratio := float64(uncompressedSize) / float64(compressedSize)
	if ratio > v.config.MaxCompressionRatio {
		return fmt.Errorf("%w: ratio %.2f exceeds limit %.2f", ErrZipBomb, ratio, v.config.MaxCompressionRatio)
	}
	return nil
}

// GetConfig returns the current configuration
func (v *FileValidator) GetConfig() *FileValidatorConfig {
	return v.config
}

// GetContentType returns the appropriate content type for a file extension
func GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	contentTypes := map[string]string{
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".json": "application/json",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".m4a":  "audio/mp4",
		".aac":  "audio/aac",
		".flac": "audio/flac",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".zip":  "application/zip",
	}
	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}
