package models

import "time"

// PresenceStatus represents user presence status
type PresenceStatus string

const (
	PresenceOnline  PresenceStatus = "online"
	PresenceOffline PresenceStatus = "offline"
	PresenceAway    PresenceStatus = "away"
	PresenceBusy    PresenceStatus = "busy"
)

// Presence represents user presence information
type Presence struct {
	UserID     string         `json:"user_id" db:"user_id"`
	Status     PresenceStatus `json:"presence" db:"status"`
	LastActive time.Time      `json:"last_active" db:"last_active"`
	StatusMsg  string         `json:"status_msg,omitempty" db:"status_msg"`
}

// PresenceUpdate represents a presence update notification
type PresenceUpdate struct {
	UserID     string         `json:"user_id"`
	Presence   PresenceStatus `json:"presence"`
	LastActive int64          `json:"last_active"`
}

// GetPresenceResponse represents the response for getting presence
type GetPresenceResponse struct {
	UserID     string         `json:"user_id"`
	Presence   PresenceStatus `json:"presence"`
	LastActive int64          `json:"last_active"`
}
