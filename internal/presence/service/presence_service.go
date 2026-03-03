package service

import (
	"fmt"

	"github.com/arcane/arcanelink/internal/presence/repository"
	"github.com/arcane/arcanelink/pkg/logger"
	"github.com/arcane/arcanelink/pkg/models"
	"go.uber.org/zap"
)

type PresenceService struct {
	repo *repository.PresenceRepository
}

func NewPresenceService(repo *repository.PresenceRepository) *PresenceService {
	return &PresenceService{
		repo: repo,
	}
}

// UpdatePresence updates user presence status
func (s *PresenceService) UpdatePresence(userID string, status models.PresenceStatus, statusMsg string) error {
	// Validate status
	validStatuses := map[models.PresenceStatus]bool{
		models.PresenceOnline:  true,
		models.PresenceOffline: true,
		models.PresenceAway:    true,
		models.PresenceBusy:    true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid presence status: %s", status)
	}

	if err := s.repo.UpdatePresence(userID, status, statusMsg); err != nil {
		return fmt.Errorf("failed to update presence: %w", err)
	}

	logger.Debug("Presence updated",
		zap.String("user_id", userID),
		zap.String("status", string(status)))

	return nil
}

// GetPresence retrieves user presence
func (s *PresenceService) GetPresence(userID string) (*models.Presence, error) {
	return s.repo.GetPresence(userID)
}

// GetMultiplePresence retrieves presence for multiple users
func (s *PresenceService) GetMultiplePresence(userIDs []string) ([]*models.Presence, error) {
	if len(userIDs) == 0 {
		return []*models.Presence{}, nil
	}

	return s.repo.GetMultiplePresence(userIDs)
}

// MarkOnline marks a user as online (called during sync)
func (s *PresenceService) MarkOnline(userID string) error {
	return s.UpdatePresence(userID, models.PresenceOnline, "")
}

// MarkOffline marks a user as offline
func (s *PresenceService) MarkOffline(userID string) error {
	return s.UpdatePresence(userID, models.PresenceOffline, "")
}
