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
	Title       string `form:"title"`
	Description string `form:"description"`
	Category    string `form:"category"`
	Genres      string `form:"genres"`
	DurationSec uint   `form:"duration_sec"`
	Resolution  string `form:"resolution"`
}
