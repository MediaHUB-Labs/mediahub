package service

import (
	"context"
	"errors"
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

func (s *MediaService) CreateMedia(ctx context.Context, file *multipart.FileHeader, req *dto.MediaCreateRequest) (*models.Media, error) {

	data, err := s.uploadRepo.Upload(ctx, file)
	if err != nil {
		return nil, err
	}

	media := &models.Media{
		FilePath:     data.FilePath,
		MimeType:     data.MimeType,
		Title:        req.Title,
		Description:  req.Description,
		Category:     req.Category,
		Genres:       req.Genres,
		FileSizeKB:   data.FileSizeKB,
		Checksum:     data.Checksum,
		DurationSec:  req.DurationSec,
		Resolution:   req.Resolution,
		IsTranscoded: false,
		IsNew:        true,
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

	return media, nil
}
