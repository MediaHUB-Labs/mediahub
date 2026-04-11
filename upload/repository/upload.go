package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mediahub/dto"
	"mediahub/utils"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Allowed MIME types for upload
var allowedMIMETypes = map[string]bool{
	// Video
	"video/mp4":        true,
	"video/x-matroska": true,
	"video/webm":       true,
	"video/avi":        true,
	"video/x-msvideo":  true,
	"video/quicktime":  true,
	"video/x-flv":      true,
	"video/mpeg":       true,
	// Audio
	"audio/mpeg":    true,
	"audio/mp3":     true,
	"audio/wav":     true,
	"audio/x-wav":   true,
	"audio/ogg":     true,
	"audio/flac":    true,
	"audio/aac":     true,
	"audio/x-m4a":   true,
	"audio/mp4":     true,
	"audio/webm":    true,
	"audio/x-flac":  true,
	// Images
	"image/jpeg":    true,
	"image/png":     true,
	"image/gif":     true,
	"image/webp":    true,
	"image/svg+xml": true,
	"image/bmp":     true,
	"image/tiff":    true,
	// Documents
	"application/pdf": true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"text/plain": true,
}

type UploadRepository struct {
	db *gorm.DB
}

func NewUploadRepository(db *gorm.DB) *UploadRepository {
	return &UploadRepository{db: db}
}

// getBasePath returns the configured upload base path.
// Uses UPLOAD_PATH env if set, otherwise resolves "uploads" relative to the executable.
func getBasePath() string {
	envPath := os.Getenv("UPLOAD_PATH")
	if envPath != "" {
		return envPath
	}

	// Resolve relative to executable directory
	exePath, err := os.Executable()
	if err != nil {
		// Fallback to relative path
		return "uploads"
	}
	return filepath.Join(filepath.Dir(exePath), "uploads")
}

func (r *UploadRepository) Upload(ctx context.Context, file *multipart.FileHeader) (*dto.UploadResult, error) {

	utils.LogToFile("Upload started for file: " + file.Filename)

	// Validate MIME type
	mimeType := file.Header.Get("Content-Type")
	if !allowedMIMETypes[mimeType] {
		utils.LogToFile("Rejected file with unsupported MIME type: " + mimeType)
		return nil, fmt.Errorf("unsupported file type: %s", mimeType)
	}

	basePath := getBasePath()

	// Ensure uploads directory exists
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		utils.LogToFile("Failed to create uploads directory: " + err.Error())
		return nil, err
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		utils.LogToFile("Failed to open uploaded file: " + err.Error())
		return nil, err
	}
	defer src.Close()

	// Decide sub-folder based on type
	subDir := "others"
	switch {
	case strings.HasPrefix(mimeType, "video/"):
		subDir = "videos"
	case strings.HasPrefix(mimeType, "image/"):
		subDir = "images"
	case strings.HasPrefix(mimeType, "audio/"):
		subDir = "audio"
	case mimeType == "application/pdf",
		mimeType == "application/msword",
		mimeType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		mimeType == "text/plain":
		subDir = "documents"
	}

	// Final base upload path
	baseUploadPath := filepath.Join(basePath, subDir)

	// Create directory
	if err := os.MkdirAll(baseUploadPath, os.ModePerm); err != nil {
		utils.LogToFile("Failed to create upload directory: " + err.Error())
		return nil, err
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	newName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	fullPath := filepath.Join(baseUploadPath, newName)

	utils.LogToFile("Saving file to: " + fullPath)

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		utils.LogToFile("Failed to create destination file: " + err.Error())
		return nil, err
	}
	defer dst.Close()

	// Copy file + calculate checksum at same time
	hasher := sha256.New()
	writer := io.MultiWriter(dst, hasher)

	size, err := io.Copy(writer, src)
	if err != nil {
		utils.LogToFile("Failed while copying file to disk: " + err.Error())
		return nil, err
	}

	// Final values
	checksum := hex.EncodeToString(hasher.Sum(nil))
	fileSizeKB := uint(size / 1024)

	utils.LogToFile(fmt.Sprintf(
		"Upload success | Path: %s | SizeKB: %d | MIME: %s | Checksum: %s",
		fullPath, fileSizeKB, mimeType, checksum,
	))

	result := &dto.UploadResult{
		FilePath:   fullPath,
		MimeType:   mimeType,
		FileSizeKB: fileSizeKB,
		Checksum:   checksum,
	}

	return result, nil
}
