package repository

import (
	"context"
	"errors"
	"mediahub/models"
	"strings"

	"gorm.io/gorm"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

// Create inserts a new media record into the database.
func (r *MediaRepository) Create(ctx context.Context, media *models.Media) error {
	result := r.db.WithContext(ctx).Create(media)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID fetches a single media item by ID.
func (r *MediaRepository) FindByID(ctx context.Context, id uint) (*models.Media, error) {
	var media models.Media
	result := r.db.WithContext(ctx).First(&media, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("media not found")
		}
		return nil, result.Error
	}
	return &media, nil
}

// FindAll returns paginated media items with optional filters.
// Enforces visibility: public items are always included; private items only if they belong to userID.
func (r *MediaRepository) FindAll(ctx context.Context, limit, offset int, category, genre, mediaType string, userID uint) ([]models.Media, int64, error) {
	var items []models.Media
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&models.Media{})

	// Visibility filter: public items OR private items owned by this user
	query = query.Where("visibility = ? OR (visibility = ? AND uploaded_by_user_id = ?)", "public", "private", userID)

	// Apply optional filters
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if genre != "" {
		query = query.Where("genres LIKE ?", "%"+genre+"%")
	}
	if mediaType != "" {
		switch mediaType {
		case "video":
			query = query.Where("mime_type LIKE ?", "video/%")
		case "audio":
			query = query.Where("mime_type LIKE ?", "audio/%")
		case "image":
			query = query.Where("mime_type LIKE ?", "image/%")
		case "document":
			query = query.Where("mime_type LIKE ? OR mime_type LIKE ? OR mime_type = ?",
				"application/pdf", "application/msword%", "text/plain")
		}
	}

	// Count total before pagination
	query.Count(&totalCount)

	// Apply pagination and ordering
	result := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return items, totalCount, nil
}

// Search performs a text search on title and description with visibility enforcement.
func (r *MediaRepository) Search(ctx context.Context, queryStr string, userID uint, limit, offset int) ([]models.Media, int64, error) {
	var items []models.Media
	var totalCount int64

	searchTerm := "%" + strings.ToLower(queryStr) + "%"

	query := r.db.WithContext(ctx).Model(&models.Media{}).
		Where("(LOWER(title) LIKE ? OR LOWER(description) LIKE ?)", searchTerm, searchTerm).
		Where("visibility = ? OR (visibility = ? AND uploaded_by_user_id = ?)", "public", "private", userID)

	query.Count(&totalCount)

	result := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return items, totalCount, nil
}

// Update saves changes to an existing media record.
func (r *MediaRepository) Update(ctx context.Context, media *models.Media) error {
	return r.db.WithContext(ctx).Save(media).Error
}

// Delete soft-deletes a media record.
func (r *MediaRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Media{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("media not found or already deleted")
	}
	return nil
}

// GetCategories returns a list of distinct categories.
func (r *MediaRepository) GetCategories(ctx context.Context) ([]string, error) {
	var categories []string
	result := r.db.WithContext(ctx).Model(&models.Media{}).
		Where("category != ''").
		Distinct("category").
		Pluck("category", &categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

// GetByChecksum finds a media item by its file checksum (for duplicate detection).
func (r *MediaRepository) GetByChecksum(ctx context.Context, checksum string) (*models.Media, error) {
	var media models.Media
	result := r.db.WithContext(ctx).Where("checksum = ?", checksum).First(&media)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error for duplicate checks
		}
		return nil, result.Error
	}
	return &media, nil
}

// FindByUser returns media items uploaded by a specific user, optionally filtered by media type.
// This is used for the personal vault (photos, documents).
func (r *MediaRepository) FindByUser(ctx context.Context, userID uint, mediaType string, limit, offset int) ([]models.Media, int64, error) {
	var items []models.Media
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&models.Media{}).
		Where("uploaded_by_user_id = ?", userID)

	if mediaType != "" {
		switch mediaType {
		case "video":
			query = query.Where("mime_type LIKE ?", "video/%")
		case "audio":
			query = query.Where("mime_type LIKE ?", "audio/%")
		case "image":
			query = query.Where("mime_type LIKE ?", "image/%")
		case "document":
			query = query.Where("mime_type LIKE ? OR mime_type LIKE ? OR mime_type = ?",
				"application/pdf", "application/msword%", "text/plain")
		}
	}

	query.Count(&totalCount)

	result := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return items, totalCount, nil
}
