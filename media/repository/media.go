package repository

import (
	"context"
	"mediahub/models"

	"gorm.io/gorm"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Create(ctx context.Context, media *models.Media) error {
	result := r.db.WithContext(ctx).Create(media)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
