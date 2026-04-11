package models

import (
	"time"

	"gorm.io/gorm"
)

// UserMediaProgress tracks viewing state for individual users on specific media items.
type UserMediaProgress struct {
	gorm.Model

	// --- RELATIONSHIPS (Foreign Keys) ---

	// Foreign Key: ID of the User who is watching
	UserID uint `gorm:"not null;uniqueIndex:idx_user_media"`
	User   User

	// Foreign Key: ID of the Media item being watched
	MediaID uint `gorm:"not null;uniqueIndex:idx_user_media"`
	Media   Media

	// --- PROGRESS DATA ---

	// CRITICAL: The timestamp (in seconds) where the user paused.
	PlayheadPositionSec float64 `gorm:"default:0"`

	// Last time this record was updated. Used to sort the 'Continue Watching' row.
	LastWatchedAt time.Time `gorm:"not null"`

	// Flag if the user has completed the item (e.g., progress > 95% of total duration).
	IsCompleted bool `gorm:"default:false"`
}
