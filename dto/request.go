package dto

// Auth Structs
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

// Media Structs

type MediaCreateRequest struct {
	FilePath      string `json:"file_path" binding:"required"`
	MimeType      string `json:"mime_type" binding:"required"`
	Checksum      string `json:"checksum" binding:"required"`
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	Genres        string `json:"genres"`
	FileSizeKB    uint   `json:"file_size_kb" binding:"required"`
	DurationSec   uint   `json:"duration_sec"`
	Resolution    string `json:"resolution"`
	ThumbnailPath string `json:"thumbnail_path"`
}
