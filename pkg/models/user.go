package models

import "time"

// User represents a user account
type User struct {
	UserID      string    `json:"user_id" db:"user_id"`
	Username    string    `json:"username" db:"username"`
	PasswordHash string   `json:"-" db:"password_hash"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	DisplayName string    `json:"display_name,omitempty" db:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty" db:"avatar_url"`
	StatusMsg   string    `json:"status_msg,omitempty" db:"status_msg"`
}

// UserProfile represents public user profile information
type UserProfile struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	StatusMsg   string `json:"status_msg,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
	ExpiresIn   int64  `json:"expires_in"`
}

// CreateUserRequest represents a user registration request
type CreateUserRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	StatusMsg   string `json:"status_msg,omitempty"`
}
