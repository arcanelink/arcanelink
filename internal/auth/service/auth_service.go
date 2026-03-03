package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/arcane/arcanelink/internal/auth/repository"
	"github.com/arcane/arcanelink/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
	expiresIn int64
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, expiresIn int64) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		expiresIn: expiresIn,
	}
}

// CreateUser creates a new user account
func (s *AuthService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Check if username already exists
	exists, err := s.userRepo.UsernameExists(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if user ID already exists
	exists, err = s.userRepo.UserExists(req.UserID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("user ID already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		UserID:       req.UserID,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(userID, password string) (*models.LoginResponse, error) {
	// Get user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateToken(user.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		AccessToken: token,
		UserID:      user.UserID,
		ExpiresIn:   s.expiresIn,
	}, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("invalid token claims")
		}
		return userID, nil
	}

	return "", fmt.Errorf("invalid token")
}

// GetUserProfile retrieves user profile information
func (s *AuthService) GetUserProfile(userID string) (*models.UserProfile, error) {
	return s.userRepo.GetUserProfile(userID)
}

// UpdateUserProfile updates user profile information
func (s *AuthService) UpdateUserProfile(userID string, req *models.UpdateProfileRequest) error {
	// Check if user exists
	exists, err := s.userRepo.UserExists(userID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user not found")
	}

	return s.userRepo.UpdateUserProfile(userID, req)
}

// generateToken generates a JWT token for a user
func (s *AuthService) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Duration(s.expiresIn) * time.Second).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Logout handles user logout (currently just validates the user exists)
func (s *AuthService) Logout(userID string) error {
	exists, err := s.userRepo.UserExists(userID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user not found")
	}
	// In a real implementation, you might want to blacklist the token
	return nil
}
