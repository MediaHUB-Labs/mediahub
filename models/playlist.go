package models

import "gorm.io/gorm"

// Playlist represents a user-created playlist (primarily for music).
type Playlist struct {
	gorm.Model
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	UserID      uint `gorm:"not null;index" json:"user_id"`
	User        User `json:"-"`
	IsPublic    bool           `gorm:"default:false" json:"is_public"`
	Items       []PlaylistItem `gorm:"foreignKey:PlaylistID" json:"items,omitempty"`
}

// PlaylistItem represents a media item within a playlist with ordering.
type PlaylistItem struct {
	gorm.Model
	PlaylistID uint `gorm:"not null;uniqueIndex:idx_playlist_media" json:"playlist_id"`
	MediaID    uint `gorm:"not null;uniqueIndex:idx_playlist_media" json:"media_id"`
	Position   uint `gorm:"not null;default:0" json:"position"`
	Media      Media `json:"media,omitempty"`
}
