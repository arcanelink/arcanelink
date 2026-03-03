package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/arcane/arcanelink/pkg/database"
	"github.com/arcane/arcanelink/pkg/models"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (user_id, username, password_hash, display_name, avatar_url, status_msg, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(query,
		user.UserID,
		user.Username,
		user.PasswordHash,
		user.DisplayName,
		user.AvatarURL,
		user.StatusMsg,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByID retrieves a user by user ID
func (r *UserRepository) GetUserByID(userID string) (*models.User, error) {
	query := `
		SELECT user_id, username, password_hash, created_at, updated_at, display_name, avatar_url, status_msg
		FROM users
		WHERE user_id = $1
	`
	var user models.User
	err := r.db.QueryRow(query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DisplayName,
		&user.AvatarURL,
		&user.StatusMsg,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT user_id, username, password_hash, created_at, updated_at, display_name, avatar_url, status_msg
		FROM users
		WHERE username = $1
	`
	var user models.User
	err := r.db.QueryRow(query, username).Scan(
		&user.UserID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DisplayName,
		&user.AvatarURL,
		&user.StatusMsg,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// UpdateUserProfile updates user profile information
func (r *UserRepository) UpdateUserProfile(userID string, profile *models.UpdateProfileRequest) error {
	query := `
		UPDATE users
		SET display_name = $1, avatar_url = $2, status_msg = $3, updated_at = $4
		WHERE user_id = $5
	`
	_, err := r.db.Exec(query,
		profile.DisplayName,
		profile.AvatarURL,
		profile.StatusMsg,
		time.Now(),
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}
	return nil
}

// UserExists checks if a user exists
func (r *UserRepository) UserExists(userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`
	var exists bool
	err := r.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// UsernameExists checks if a username is already taken
func (r *UserRepository) UsernameExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return exists, nil
}

// GetUserProfile retrieves public user profile
func (r *UserRepository) GetUserProfile(userID string) (*models.UserProfile, error) {
	query := `
		SELECT user_id, display_name, avatar_url, status_msg
		FROM users
		WHERE user_id = $1
	`
	var profile models.UserProfile
	err := r.db.QueryRow(query, userID).Scan(
		&profile.UserID,
		&profile.DisplayName,
		&profile.AvatarURL,
		&profile.StatusMsg,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return &profile, nil
}

// Helper function to handle NULL values
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func scanNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
