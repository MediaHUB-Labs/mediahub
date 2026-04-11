package service

import (
	"context"
	"mediahub/dto"
	mediarepository "mediahub/media/repository"
	"mediahub/progress/repository"
	"mediahub/utils"
	"fmt"
)

type ProgressService struct {
	repo      *repository.ProgressRepository
	mediaRepo *mediarepository.MediaRepository
}

func NewProgressService(repo *repository.ProgressRepository, mediaRepo *mediarepository.MediaRepository) *ProgressService {
	return &ProgressService{
		repo:      repo,
		mediaRepo: mediaRepo,
	}
}

// SaveProgress upserts playback position and auto-detects completion (≥95% of duration).
func (s *ProgressService) SaveProgress(ctx context.Context, userID uint, req *dto.SaveProgressRequest) (*dto.ProgressResponse, error) {
	// Verify the media item exists
	media, err := s.mediaRepo.FindByID(ctx, req.MediaID)
	if err != nil {
		return nil, err
	}

	progress, err := s.repo.Upsert(ctx, userID, req.MediaID, req.PositionSec)
	if err != nil {
		return nil, err
	}

	// Auto-complete detection: if position ≥ 95% of total duration
	if media.DurationSec > 0 && req.PositionSec >= float64(media.DurationSec)*0.95 {
		progress.IsCompleted = true
		if err := s.repo.MarkCompleted(ctx, userID, req.MediaID); err != nil {
			utils.LogToFile(fmt.Sprintf("Failed to mark progress as completed: %v", err))
		}
	}

	return &dto.ProgressResponse{
		MediaID:             progress.MediaID,
		PlayheadPositionSec: progress.PlayheadPositionSec,
		IsCompleted:         progress.IsCompleted,
		LastWatchedAt:       progress.LastWatchedAt,
	}, nil
}

// GetContinueWatching returns items the user hasn't finished, sorted by most recent.
func (s *ProgressService) GetContinueWatching(ctx context.Context, userID uint, limit int) ([]dto.ContinueWatchingItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	records, err := s.repo.FindByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	var items []dto.ContinueWatchingItem
	for _, record := range records {
		// Skip if the media item was deleted (soft-deleted)
		if record.Media.ID == 0 {
			continue
		}

		items = append(items, dto.ContinueWatchingItem{
			Media: dto.MediaResponse{
				ID:            record.Media.ID,
				Title:         record.Media.Title,
				Description:   record.Media.Description,
				Category:      record.Media.Category,
				Genres:        record.Media.Genres,
				MimeType:      record.Media.MimeType,
				FileSizeKB:    record.Media.FileSizeKB,
				DurationSec:   record.Media.DurationSec,
				Resolution:    record.Media.Resolution,
				ThumbnailPath: record.Media.ThumbnailPath,
				IsTranscoded:  record.Media.IsTranscoded,
				IsNew:         record.Media.IsNew,
				Visibility:    record.Media.Visibility,
				CreatedAt:     record.Media.CreatedAt,
				UpdatedAt:     record.Media.UpdatedAt,
			},
			Progress: dto.ProgressResponse{
				MediaID:             record.MediaID,
				PlayheadPositionSec: record.PlayheadPositionSec,
				IsCompleted:         record.IsCompleted,
				LastWatchedAt:       record.LastWatchedAt,
			},
		})
	}

	return items, nil
}

// ClearProgress removes a user's progress for a specific media item.
func (s *ProgressService) ClearProgress(ctx context.Context, userID uint, mediaID uint) error {
	return s.repo.Delete(ctx, userID, mediaID)
}
