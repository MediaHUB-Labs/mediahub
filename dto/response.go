package dto

import "time"

// ═══════════════════════════════════════════════════════
// Auth Response Structs
// ═══════════════════════════════════════════════════════

type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	Token string `json:"token,omitempty"`
	User  UserResponse
}

// ═══════════════════════════════════════════════════════
// Generic Response
// ═══════════════════════════════════════════════════════

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type HealthReponse struct {
	Message string `json:"message"`
}

// ═══════════════════════════════════════════════════════
// Upload Response
// ═══════════════════════════════════════════════════════

type UploadResult struct {
	FilePath   string
	MimeType   string
	FileSizeKB uint
	Checksum   string
}

// ═══════════════════════════════════════════════════════
// Media Response Structs
// ═══════════════════════════════════════════════════════

type MediaResponse struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Category         string    `json:"category"`
	Genres           string    `json:"genres"`
	MimeType         string    `json:"mime_type"`
	FileSizeKB       uint      `json:"file_size_kb"`
	DurationSec      uint      `json:"duration_sec"`
	Resolution       string    `json:"resolution"`
	ThumbnailPath    string    `json:"thumbnail_path,omitempty"`
	IsTranscoded     bool      `json:"is_transcoded"`
	IsNew            bool      `json:"is_new"`
	Visibility       string    `json:"visibility"`
	UploadedByUserID uint      `json:"uploaded_by_user_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	// Optional: include user's progress if available
	Progress *ProgressResponse `json:"progress,omitempty"`
}

type MediaListResponse struct {
	Items      []MediaResponse `json:"items"`
	TotalCount int64           `json:"total_count"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
}

type CategoryResponse struct {
	Categories []string `json:"categories"`
}

// ═══════════════════════════════════════════════════════
// Progress Response Structs
// ═══════════════════════════════════════════════════════

type ProgressResponse struct {
	MediaID             uint      `json:"media_id"`
	PlayheadPositionSec float64   `json:"playhead_position_sec"`
	IsCompleted         bool      `json:"is_completed"`
	LastWatchedAt       time.Time `json:"last_watched_at"`
}

type ContinueWatchingItem struct {
	Media    MediaResponse    `json:"media"`
	Progress ProgressResponse `json:"progress"`
}

// ═══════════════════════════════════════════════════════
// Playlist Response Structs
// ═══════════════════════════════════════════════════════

type PlaylistResponse struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	IsPublic    bool                   `json:"is_public"`
	UserID      uint                   `json:"user_id"`
	ItemCount   int                    `json:"item_count"`
	Items       []PlaylistItemResponse `json:"items,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type PlaylistItemResponse struct {
	ID       uint          `json:"id"`
	MediaID  uint          `json:"media_id"`
	Position uint          `json:"position"`
	Media    MediaResponse `json:"media"`
}

// ═══════════════════════════════════════════════════════
// FFmpeg / Media Info
// ═══════════════════════════════════════════════════════

type MediaInfo struct {
	DurationSec uint   `json:"duration_sec"`
	Resolution  string `json:"resolution"`
	Codec       string `json:"codec"`
	Bitrate     string `json:"bitrate"`
}
