package models

import (
	"time"

	"gorm.io/gorm"
)

// UserMediaProgress tracks viewing state for individual users on specific media items.
type UserMediaProgress struct {
	gorm.Model

	// --- RELATIONSHIPS (Foreign Keys) ---

	UserID uint `gorm:"not null;uniqueIndex:idx_user_media" json:"user_id"`
	User   User `json:"-"`

	MediaID uint `gorm:"not null;uniqueIndex:idx_user_media" json:"media_id"`
	Media   Media `json:"media,omitempty"`

	PlayheadPositionSec float64 `gorm:"default:0" json:"playhead_position_sec"`
	LastWatchedAt time.Time `gorm:"not null" json:"last_watched_at"`
	IsCompleted bool `gorm:"default:false" json:"is_completed"`
}
