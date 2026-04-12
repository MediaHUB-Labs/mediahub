package dto

// ═══════════════════════════════════════════════════════
// Auth Request Structs
// ═══════════════════════════════════════════════════════

type SignupRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	ID        uint   `json:"id" binding:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UserRequest struct {
	ID uint `json:"id" binding:"required"`
}

// ═══════════════════════════════════════════════════════
// Media Request Structs
// ═══════════════════════════════════════════════════════

type MediaCreateRequest struct {
	Title       string `form:"title"`
	Description string `form:"description"`
	Category    string `form:"category"`
	Type        string `form:"type"`
	Genres      string `form:"genres"`
	DurationSec uint   `form:"duration_sec"`
	Resolution  string `form:"resolution"`
}

type MediaIDRequest struct {
	ID uint `json:"id" binding:"required"`
}

type MediaListRequest struct {
	Limit    int    `form:"limit"`    // default 20
	Offset   int    `form:"offset"`   // default 0
	Category string `form:"category"` // filter by category
	Genres   string `form:"genres"`   // filter by genre
	Type     string `form:"type"`     // filter by media type: video, audio, image, document
}

type MediaSearchRequest struct {
	Query  string `form:"q" binding:"required"`
	Limit  int    `form:"limit"`  // default 20
	Offset int    `form:"offset"` // default 0
}

type MediaUpdateRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Genres      string `json:"genres"`
}

// ═══════════════════════════════════════════════════════
// Progress Request Structs
// ═══════════════════════════════════════════════════════

type SaveProgressRequest struct {
	MediaID     uint    `json:"media_id" binding:"required"`
	PositionSec float64 `json:"position_sec" binding:"required"`
}

type ClearProgressRequest struct {
	MediaID uint `json:"media_id" binding:"required"`
}

// ═══════════════════════════════════════════════════════
// Playlist Request Structs
// ═══════════════════════════════════════════════════════

type CreatePlaylistRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

type UpdatePlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    *bool  `json:"is_public"` // pointer to distinguish false from unset
}

type PlaylistItemRequest struct {
	MediaID uint `json:"media_id" binding:"required"`
}
