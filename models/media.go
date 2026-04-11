package models

import "gorm.io/gorm"

// Media holds data for all types: Movies, Videos, Images, Documents, Audio.
type Media struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt

	// --- CORE IDENTIFICATION & TYPE ---

	// CRITICAL: The unique path to the physical file on the server.
	FilePath string `gorm:"uniqueIndex;not null"`

	// The file's MIME type (e.g., 'video/mp4', 'image/jpeg', 'application/pdf').
	MimeType string `gorm:"not null"`

	// --- OWNERSHIP & ACCESS CONTROL ---

	// The user who uploaded this media item. 0 = system/unknown (legacy data).
	UploadedByUserID uint `gorm:"index;default:0"`

	// Visibility: "public" (movies, music — all users) or "private" (photos, docs — owner only).
	Visibility string `gorm:"default:'public';index"`

	// --- DISPLAY & ORGANIZATION ---

	Title       string `gorm:"not null"`
	Description string

	// Category could be 'Movie', 'Home Video', 'Music', 'Photo', 'Document', etc.
	Category string `gorm:"index"`

	// Genres or Tags (store as a comma-separated string)
	Genres string // e.g., "Action,Sci-Fi"

	// --- FILE & VIDEO METRICS ---

	FileSizeKB uint `gorm:"not null"` // Stored in Kilobytes

	// Checksum of the source file to detect if the file has changed on disk.
	// Also useful for checking data duplication/corruption/partial
	Checksum string `gorm:"uniqueIndex"`

	// Playback duration in seconds (0 for images/documents).
	DurationSec uint

	// Resolution (e.g., '1920x1080', '4K').
	Resolution string

	// --- ASSETS & STATE ---

	// Path to the generated thumbnail image file.
	ThumbnailPath string

	// Flag if the Transcoder Service has processed the file for streaming (HLS/DASH).
	IsTranscoded bool `gorm:"default:false"`

	// Path to the HLS manifest file (.m3u8) if transcoded.
	TranscodedPath string

	// Flag for new/recently added items.
	IsNew bool `gorm:"default:true"`

	// Relation: Links to all progress records for this media item (used for lookups).
	ProgressRecords []UserMediaProgress
}
