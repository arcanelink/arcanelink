package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/arcane/arcanelink/pkg/database"
	"github.com/arcane/arcanelink/pkg/models"
)

type PresenceRepository struct {
	db *database.DB
}

func NewPresenceRepository(db *database.DB) *PresenceRepository {
	return &PresenceRepository{db: db}
}

// UpdatePresence updates or creates user presence
func (r *PresenceRepository) UpdatePresence(userID string, status models.PresenceStatus, statusMsg string) error {
	query := `
		INSERT INTO presence (user_id, status, last_active, status_msg)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET status = $2, last_active = $3, status_msg = $4
	`
	_, err := r.db.Exec(query, userID, status, time.Now(), statusMsg)
	if err != nil {
		return fmt.Errorf("failed to update presence: %w", err)
	}
	return nil
}

// GetPresence retrieves user presence
func (r *PresenceRepository) GetPresence(userID string) (*models.Presence, error) {
	query := `
		SELECT user_id, status, last_active, status_msg
		FROM presence
		WHERE user_id = $1
	`
	var presence models.Presence
	err := r.db.QueryRow(query, userID).Scan(
		&presence.UserID,
		&presence.Status,
		&presence.LastActive,
		&presence.StatusMsg,
	)
	if err == sql.ErrNoRows {
		// Return offline status if not found
		return &models.Presence{
			UserID:     userID,
			Status:     models.PresenceOffline,
			LastActive: time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get presence: %w", err)
	}
	return &presence, nil
}

// GetMultiplePresence retrieves presence for multiple users
func (r *PresenceRepository) GetMultiplePresence(userIDs []string) ([]*models.Presence, error) {
	if len(userIDs) == 0 {
		return []*models.Presence{}, nil
	}

	query := `
		SELECT user_id, status, last_active, status_msg
		FROM presence
		WHERE user_id = ANY($1)
	`
	rows, err := r.db.Query(query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query multiple presence: %w", err)
	}
	defer rows.Close()

	presenceMap := make(map[string]*models.Presence)
	for rows.Next() {
		var presence models.Presence
		if err := rows.Scan(
			&presence.UserID,
			&presence.Status,
			&presence.LastActive,
			&presence.StatusMsg,
		); err != nil {
			return nil, fmt.Errorf("failed to scan presence: %w", err)
		}
		presenceMap[presence.UserID] = &presence
	}

	// Fill in missing users with offline status
	result := make([]*models.Presence, len(userIDs))
	for i, userID := range userIDs {
		if presence, ok := presenceMap[userID]; ok {
			result[i] = presence
		} else {
			result[i] = &models.Presence{
				UserID:     userID,
				Status:     models.PresenceOffline,
				LastActive: time.Now(),
			}
		}
	}

	return result, nil
}

// MarkOfflineIfInactive marks users as offline if they haven't been active
func (r *PresenceRepository) MarkOfflineIfInactive(timeout time.Duration) (int, error) {
	query := `
		UPDATE presence
		SET status = 'offline'
		WHERE status != 'offline'
		AND last_active < $1
	`
	result, err := r.db.Exec(query, time.Now().Add(-timeout))
	if err != nil {
		return 0, fmt.Errorf("failed to mark offline: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// GetOnlineCount returns the number of online users
func (r *PresenceRepository) GetOnlineCount() (int, error) {
	query := `SELECT COUNT(*) FROM presence WHERE status = 'online'`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get online count: %w", err)
	}
	return count, nil
}

// DeleteOldPresence deletes presence records older than specified duration
func (r *PresenceRepository) DeleteOldPresence(olderThan time.Duration) (int, error) {
	query := `
		DELETE FROM presence
		WHERE status = 'offline'
		AND last_active < $1
	`
	result, err := r.db.Exec(query, time.Now().Add(-olderThan))
	if err != nil {
		return 0, fmt.Errorf("failed to delete old presence: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}
