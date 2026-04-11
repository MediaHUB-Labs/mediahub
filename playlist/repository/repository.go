package repository

import (
	"context"
	"errors"
	"mediahub/models"

	"gorm.io/gorm"
)

type PlaylistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

// Create inserts a new playlist.
func (r *PlaylistRepository) Create(ctx context.Context, playlist *models.Playlist) error {
	return r.db.WithContext(ctx).Create(playlist).Error
}

// FindByID fetches a single playlist with its items.
func (r *PlaylistRepository) FindByID(ctx context.Context, id uint) (*models.Playlist, error) {
	var playlist models.Playlist
	result := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Items.Media").
		First(&playlist, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("playlist not found")
		}
		return nil, result.Error
	}

	return &playlist, nil
}

// FindByUser returns all playlists owned by a specific user.
func (r *PlaylistRepository) FindByUser(ctx context.Context, userID uint) ([]models.Playlist, error) {
	var playlists []models.Playlist
	result := r.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&playlists)

	if result.Error != nil {
		return nil, result.Error
	}

	return playlists, nil
}

// Update saves changes to a playlist.
func (r *PlaylistRepository) Update(ctx context.Context, playlist *models.Playlist) error {
	return r.db.WithContext(ctx).Save(playlist).Error
}

// Delete removes a playlist and all its items.
func (r *PlaylistRepository) Delete(ctx context.Context, id uint) error {
	// Delete playlist items first
	if err := r.db.WithContext(ctx).Where("playlist_id = ?", id).Delete(&models.PlaylistItem{}).Error; err != nil {
		return err
	}
	// Delete the playlist
	result := r.db.WithContext(ctx).Delete(&models.Playlist{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("playlist not found")
	}
	return nil
}

// AddItem adds a media item to a playlist.
func (r *PlaylistRepository) AddItem(ctx context.Context, playlistID, mediaID, position uint) (*models.PlaylistItem, error) {
	// If position is 0, auto-assign the next position
	if position == 0 {
		var maxPosition uint
		r.db.WithContext(ctx).
			Model(&models.PlaylistItem{}).
			Where("playlist_id = ?", playlistID).
			Select("COALESCE(MAX(position), 0)").
			Scan(&maxPosition)
		position = maxPosition + 1
	}

	item := &models.PlaylistItem{
		PlaylistID: playlistID,
		MediaID:    mediaID,
		Position:   position,
	}

	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}

	// Reload with media data
	r.db.WithContext(ctx).Preload("Media").First(item, item.ID)
	return item, nil
}

// RemoveItem removes a media item from a playlist.
func (r *PlaylistRepository) RemoveItem(ctx context.Context, playlistID, mediaID uint) error {
	result := r.db.WithContext(ctx).
		Where("playlist_id = ? AND media_id = ?", playlistID, mediaID).
		Delete(&models.PlaylistItem{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found in playlist")
	}
	return nil
}

// GetItemCount returns the number of items in a playlist.
func (r *PlaylistRepository) GetItemCount(ctx context.Context, playlistID uint) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&models.PlaylistItem{}).
		Where("playlist_id = ?", playlistID).
		Count(&count)
	return count, result.Error
}
