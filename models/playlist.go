package models

import "gorm.io/gorm"

// Playlist represents a user-created playlist (primarily for music).
type Playlist struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	UserID      uint `gorm:"not null;index"` // Owner of the playlist
	User        User
	IsPublic    bool           `gorm:"default:false"` // Whether other users can see this playlist
	Items       []PlaylistItem `gorm:"foreignKey:PlaylistID"`
}

// PlaylistItem represents a media item within a playlist with ordering.
type PlaylistItem struct {
	gorm.Model
	PlaylistID uint `gorm:"not null;uniqueIndex:idx_playlist_media"`
	MediaID    uint `gorm:"not null;uniqueIndex:idx_playlist_media"`
	Position   uint `gorm:"not null;default:0"` // Ordering within the playlist
	Media      Media
}
