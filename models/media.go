package models

import "gorm.io/gorm"

// Media holds data for all types: Movies, Videos, Images, Documents, Audio.
type Media struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt

	// --- CORE IDENTIFICATION & TYPE ---

	FilePath string `gorm:"uniqueIndex;not null" json:"file_path"`
	MimeType string `gorm:"not null" json:"mime_type"`
	MediaType string `gorm:"index" json:"media_type"`

	// --- OWNERSHIP & ACCESS CONTROL ---

	UploadedByUserID uint `gorm:"index;default:0" json:"uploaded_by_user_id"`
	Visibility string `gorm:"default:'public';index" json:"visibility"`

	// --- DISPLAY & ORGANIZATION ---

	Title       string `gorm:"not null" json:"title"`
	Description string `json:"description"`
	Category string `gorm:"index" json:"category"`
	Genres string `json:"genres"`

	// --- FILE & VIDEO METRICS ---

	FileSizeKB uint `gorm:"not null" json:"file_size_kb"`
	Checksum string `gorm:"uniqueIndex" json:"checksum"`
	DurationSec uint `json:"duration_sec"`
	Resolution string `json:"resolution"`

	// --- ASSETS & STATE ---

	ThumbnailPath string `json:"thumbnail_path"`
	IsTranscoded bool `gorm:"default:false" json:"is_transcoded"`
	TranscodedPath string `json:"transcoded_path"`
	IsNew bool `gorm:"default:true" json:"is_new"`

	// Relation: Links to all progress records for this media item (used for lookups).
	ProgressRecords []UserMediaProgress `json:"progress_records,omitempty"`
}
