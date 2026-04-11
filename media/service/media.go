package service

import (
	"context"
	"errors"
	"fmt"
	"mediahub/dto"
	"mediahub/media/repository"
	"mediahub/models"
	uploadrepository "mediahub/upload/repository"
	"mediahub/utils"
	"mime/multipart"
	"os"
	"strings"
)

type MediaService struct {
	repo       *repository.MediaRepository
	uploadRepo *uploadrepository.UploadRepository
}

func NewMediaService(repo *repository.MediaRepository, uploadRepo *uploadrepository.UploadRepository) *MediaService {
	return &MediaService{
		repo:       repo,
		uploadRepo: uploadRepo,
	}
}

// CreateMedia handles file upload, metadata extraction, and DB insert with visibility rules.
// Movies/music → public, Photos/documents → private (vault per user).
func (s *MediaService) CreateMedia(ctx context.Context, file *multipart.FileHeader, req *dto.MediaCreateRequest, userID uint) (*models.Media, error) {

	data, err := s.uploadRepo.Upload(ctx, file)
	if err != nil {
		return nil, err
	}

	// Determine visibility based on MIME type
	visibility := "public" // Default for videos and audio
	if strings.HasPrefix(data.MimeType, "image/") {
		visibility = "private" // Photos are vault items
	} else if !strings.HasPrefix(data.MimeType, "video/") && !strings.HasPrefix(data.MimeType, "audio/") {
		visibility = "private" // Documents are vault items
	}

	media := &models.Media{
		FilePath:         data.FilePath,
		MimeType:         data.MimeType,
		Title:            req.Title,
		Description:      req.Description,
		Category:         req.Category,
		Genres:           req.Genres,
		FileSizeKB:       data.FileSizeKB,
		Checksum:         data.Checksum,
		DurationSec:      req.DurationSec,
		Resolution:       req.Resolution,
		IsTranscoded:     false,
		IsNew:            true,
		UploadedByUserID: userID,
		Visibility:       visibility,
	}

	// Try to probe media info with FFmpeg (if available)
	if strings.HasPrefix(data.MimeType, "video/") || strings.HasPrefix(data.MimeType, "audio/") {
		info, probeErr := utils.ProbeMediaInfo(data.FilePath)
		if probeErr != nil {
			utils.LogToFile("FFprobe error (non-fatal): " + probeErr.Error())
		} else if info != nil {
			// Use probed values if not manually provided
			if media.DurationSec == 0 {
				media.DurationSec = info.DurationSec
			}
			if media.Resolution == "" {
				media.Resolution = info.Resolution
			}
		}
	}

	// Try DB insert
	if err := s.repo.Create(ctx, media); err != nil {
		// ROLLBACK → delete uploaded file
		_ = os.Remove(data.FilePath)
		// Detect duplicate checksum
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "media.checksum") {
			utils.LogToFile("Uploaded file already exist. Upload Reverted.")
			return nil, errors.New("file already exist.")
		}
		return nil, err
	}

	// Generate thumbnail for video files (async, non-blocking)
	if strings.HasPrefix(data.MimeType, "video/") {
		go func() {
			basePath := os.Getenv("UPLOAD_PATH")
			if basePath == "" {
				exePath, err := os.Executable()
				if err == nil {
					basePath = fmt.Sprintf("%s/uploads", exePath[:len(exePath)-len("/mediahub-server")])
				} else {
					basePath = "uploads"
				}
			}
			thumbnailPath := utils.GetThumbnailPath(basePath, media.ID)
			if err := utils.GenerateThumbnail(data.FilePath, thumbnailPath); err != nil {
				utils.LogToFile("Thumbnail generation failed (non-fatal): " + err.Error())
				return
			}
			// Update media record with thumbnail path
			media.ThumbnailPath = thumbnailPath
			_ = s.repo.Update(context.Background(), media)
		}()
	}

	return media, nil
}

// GetMedia returns a single media item, enforcing visibility rules.
func (s *MediaService) GetMedia(ctx context.Context, id uint, userID uint) (*models.Media, error) {
	media, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Enforce visibility: private items only accessible by owner
	if media.Visibility == "private" && media.UploadedByUserID != userID {
		return nil, errors.New("access denied")
	}

	return media, nil
}

// ListMedia returns paginated media items with visibility enforcement.
func (s *MediaService) ListMedia(ctx context.Context, req *dto.MediaListRequest, userID uint) ([]models.Media, int64, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	return s.repo.FindAll(ctx, limit, offset, req.Category, req.Genres, req.Type, userID)
}

// SearchMedia searches for media with visibility enforcement.
func (s *MediaService) SearchMedia(ctx context.Context, query string, userID uint, limit, offset int) ([]models.Media, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.Search(ctx, query, userID, limit, offset)
}

// GetCategories returns all distinct media categories.
func (s *MediaService) GetCategories(ctx context.Context) ([]string, error) {
	return s.repo.GetCategories(ctx)
}

// UpdateMedia updates metadata for a media item. Private items can only be updated by the owner.
func (s *MediaService) UpdateMedia(ctx context.Context, req *dto.MediaUpdateRequest, userID uint) (*models.Media, error) {
	media, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Only owner can update private items; anyone can update public items they uploaded
	if media.UploadedByUserID != userID {
		return nil, errors.New("access denied: only the uploader can update this media")
	}

	if req.Title != "" {
		media.Title = req.Title
	}
	if req.Description != "" {
		media.Description = req.Description
	}
	if req.Category != "" {
		media.Category = req.Category
	}
	if req.Genres != "" {
		media.Genres = req.Genres
	}

	if err := s.repo.Update(ctx, media); err != nil {
		return nil, err
	}

	return media, nil
}

// DeleteMedia soft-deletes a media item. Only the uploader can delete.
func (s *MediaService) DeleteMedia(ctx context.Context, id uint, userID uint) error {
	media, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Only owner can delete
	if media.UploadedByUserID != userID {
		return errors.New("access denied: only the uploader can delete this media")
	}

	// Optionally remove file from disk
	if media.FilePath != "" {
		_ = os.Remove(media.FilePath)
	}
	// Remove thumbnail if exists
	if media.ThumbnailPath != "" {
		_ = os.Remove(media.ThumbnailPath)
	}

	return s.repo.Delete(ctx, id)
}

// GetUserVault returns private media items for a specific user (photos, documents).
func (s *MediaService) GetUserVault(ctx context.Context, userID uint, mediaType string, limit, offset int) ([]models.Media, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.FindByUser(ctx, userID, mediaType, limit, offset)
}

// MediaToResponse converts a Media model to a MediaResponse DTO.
func MediaToResponse(media *models.Media) dto.MediaResponse {
	return dto.MediaResponse{
		ID:               media.ID,
		Title:            media.Title,
		Description:      media.Description,
		Category:         media.Category,
		Genres:           media.Genres,
		MimeType:         media.MimeType,
		FileSizeKB:       media.FileSizeKB,
		DurationSec:      media.DurationSec,
		Resolution:       media.Resolution,
		ThumbnailPath:    media.ThumbnailPath,
		IsTranscoded:     media.IsTranscoded,
		IsNew:            media.IsNew,
		Visibility:       media.Visibility,
		UploadedByUserID: media.UploadedByUserID,
		CreatedAt:        media.CreatedAt,
		UpdatedAt:        media.UpdatedAt,
	}
}
