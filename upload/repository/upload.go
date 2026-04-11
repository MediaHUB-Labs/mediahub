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

type UploadRepository struct {
	db *gorm.DB
}

func NewUploadRepository(db *gorm.DB) *UploadRepository {
	return &UploadRepository{db: db}
}
func (r *UploadRepository) Upload(ctx context.Context, file *multipart.FileHeader) (*dto.UploadResult, error) {

	utils.LogToFile("Upload started for file: " + file.Filename)

	basePath := "uploads"

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

	// Detect MIME type
	mimeType := file.Header.Get("Content-Type")

	// Decide sub-folder based on type
	subDir := "others"
	switch {
	case strings.HasPrefix(mimeType, "video/"):
		subDir = "videos"
	case strings.HasPrefix(mimeType, "image/"):
		subDir = "images"
	case strings.HasPrefix(mimeType, "audio/"):
		subDir = "audio"
	case mimeType == "application/pdf":
		subDir = "documents"
	}

	// Final base upload path
	baseUploadPath := filepath.Join("uploads", subDir)

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
