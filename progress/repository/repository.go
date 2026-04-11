package repository

import (
	"context"
	"errors"
	"mediahub/models"
	"time"

	"gorm.io/gorm"
)

type ProgressRepository struct {
	db *gorm.DB
}

func NewProgressRepository(db *gorm.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

// Upsert creates or updates a user's playback progress for a media item.
func (r *ProgressRepository) Upsert(ctx context.Context, userID, mediaID uint, positionSec float64) (*models.UserMediaProgress, error) {
	var progress models.UserMediaProgress

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND media_id = ?", userID, mediaID).
		First(&progress)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create new progress record
		progress = models.UserMediaProgress{
			UserID:              userID,
			MediaID:             mediaID,
			PlayheadPositionSec: positionSec,
			LastWatchedAt:       time.Now(),
		}
		if err := r.db.WithContext(ctx).Create(&progress).Error; err != nil {
			return nil, err
		}
		return &progress, nil
	} else if result.Error != nil {
		return nil, result.Error
	}

	// Update existing
	progress.PlayheadPositionSec = positionSec
	progress.LastWatchedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(&progress).Error; err != nil {
		return nil, err
	}

	return &progress, nil
}

// FindByUser returns progress records for a user, sorted by last_watched_at desc.
// Excludes completed items to power the "Continue Watching" feature.
func (r *ProgressRepository) FindByUser(ctx context.Context, userID uint, limit int) ([]models.UserMediaProgress, error) {
	var records []models.UserMediaProgress

	result := r.db.WithContext(ctx).
		Preload("Media").
		Where("user_id = ? AND is_completed = ?", userID, false).
		Order("last_watched_at DESC").
		Limit(limit).
		Find(&records)

	if result.Error != nil {
		return nil, result.Error
	}

	return records, nil
}

// FindByUserAndMedia returns a specific progress record.
func (r *ProgressRepository) FindByUserAndMedia(ctx context.Context, userID, mediaID uint) (*models.UserMediaProgress, error) {
	var progress models.UserMediaProgress

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND media_id = ?", userID, mediaID).
		First(&progress)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil // No progress yet
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return &progress, nil
}

// Delete removes a progress record for a specific user and media item.
func (r *ProgressRepository) Delete(ctx context.Context, userID, mediaID uint) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND media_id = ?", userID, mediaID).
		Delete(&models.UserMediaProgress{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no progress record found")
	}
	return nil
}

// MarkCompleted sets the is_completed flag for a progress record.
func (r *ProgressRepository) MarkCompleted(ctx context.Context, userID, mediaID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.UserMediaProgress{}).
		Where("user_id = ? AND media_id = ?", userID, mediaID).
		Update("is_completed", true).Error
}
