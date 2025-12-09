package dto

import "time"

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
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type HealthReponse struct {
	Message string `json:"message"`
}

type UploadResult struct {
	FilePath   string
	MimeType   string
	FileSizeKB uint
	Checksum   string
}
