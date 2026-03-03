package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arcane/arcanelink/pkg/database"
	"github.com/arcane/arcanelink/pkg/models"
)

type RoomRepository struct {
	db *database.DB
}

func NewRoomRepository(db *database.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// CreateRoom creates a new room
func (r *RoomRepository) CreateRoom(room *models.Room) error {
	query := `
		INSERT INTO rooms (room_id, creator, name, topic, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(query,
		room.RoomID,
		room.Creator,
		room.Name,
		room.Topic,
		room.AvatarURL,
		room.CreatedAt,
		room.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}
	return nil
}

// GetRoom retrieves a room by ID
func (r *RoomRepository) GetRoom(roomID string) (*models.Room, error) {
	query := `
		SELECT room_id, creator, name, topic, avatar_url, created_at, updated_at
		FROM rooms
		WHERE room_id = $1
	`
	var room models.Room
	err := r.db.QueryRow(query, roomID).Scan(
		&room.RoomID,
		&room.Creator,
		&room.Name,
		&room.Topic,
		&room.AvatarURL,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("room not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	return &room, nil
}

// AddMember adds a member to a room
func (r *RoomRepository) AddMember(roomID, userID string) error {
	query := `
		INSERT INTO room_members (room_id, user_id, joined_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (room_id, user_id) DO NOTHING
	`
	_, err := r.db.Exec(query, roomID, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

// RemoveMember removes a member from a room
func (r *RoomRepository) RemoveMember(roomID, userID string) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}

// IsMember checks if a user is a member of a room
func (r *RoomRepository) IsMember(roomID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`
	var exists bool
	err := r.db.QueryRow(query, roomID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}
	return exists, nil
}

// GetMembers retrieves all members of a room
func (r *RoomRepository) GetMembers(roomID string) ([]*models.RoomMember, error) {
	query := `
		SELECT room_id, user_id, joined_at
		FROM room_members
		WHERE room_id = $1
		ORDER BY joined_at ASC
	`
	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to query members: %w", err)
	}
	defer rows.Close()

	var members []*models.RoomMember
	for rows.Next() {
		var member models.RoomMember
		if err := rows.Scan(&member.RoomID, &member.UserID, &member.JoinedAt); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, &member)
	}
	return members, nil
}

// GetUserRooms retrieves all rooms a user is a member of
func (r *RoomRepository) GetUserRooms(userID string) ([]*models.RoomInfo, error) {
	query := `
		SELECT r.room_id, r.name, r.topic,
		       COUNT(rm.user_id) as member_count,
		       COALESCE(MAX(re.timestamp), EXTRACT(EPOCH FROM r.created_at) * 1000) as last_activity
		FROM rooms r
		INNER JOIN room_members rm ON r.room_id = rm.room_id
		LEFT JOIN room_events re ON r.room_id = re.room_id
		WHERE r.room_id IN (
			SELECT room_id FROM room_members WHERE user_id = $1
		)
		GROUP BY r.room_id, r.name, r.topic, r.created_at
		ORDER BY last_activity DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*models.RoomInfo
	for rows.Next() {
		var room models.RoomInfo
		if err := rows.Scan(
			&room.RoomID,
			&room.Name,
			&room.Topic,
			&room.MemberCount,
			&room.LastActivity,
		); err != nil {
			return nil, fmt.Errorf("failed to scan room info: %w", err)
		}
		rooms = append(rooms, &room)
	}
	return rooms, nil
}

// SaveRoomEvent saves a room event
func (r *RoomRepository) SaveRoomEvent(event *models.RoomEvent) error {
	contentJSON, err := json.Marshal(event.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	query := `
		INSERT INTO room_events (event_id, room_id, sender, event_type, content, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.db.Exec(query,
		event.EventID,
		event.RoomID,
		event.Sender,
		event.EventType,
		contentJSON,
		event.Timestamp,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to save room event: %w", err)
	}
	return nil
}

// GetRoomHistory retrieves room event history
func (r *RoomRepository) GetRoomHistory(roomID string, limit int, beforeTimestamp int64) ([]*models.RoomEvent, error) {
	query := `
		SELECT event_id, room_id, sender, event_type, content, timestamp, created_at
		FROM room_events
		WHERE room_id = $1 AND timestamp < $2
		ORDER BY timestamp DESC
		LIMIT $3
	`

	if beforeTimestamp == 0 {
		beforeTimestamp = time.Now().UnixMilli()
	}

	rows, err := r.db.Query(query, roomID, beforeTimestamp, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query room history: %w", err)
	}
	defer rows.Close()

	var events []*models.RoomEvent
	for rows.Next() {
		var event models.RoomEvent
		var contentJSON []byte
		if err := rows.Scan(
			&event.EventID,
			&event.RoomID,
			&event.Sender,
			&event.EventType,
			&contentJSON,
			&event.Timestamp,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if err := json.Unmarshal(contentJSON, &event.Content); err != nil {
			return nil, fmt.Errorf("failed to unmarshal content: %w", err)
		}

		events = append(events, &event)
	}
	return events, nil
}

// GetMemberCount returns the number of members in a room
func (r *RoomRepository) GetMemberCount(roomID string) (int, error) {
	query := `SELECT COUNT(*) FROM room_members WHERE room_id = $1`
	var count int
	err := r.db.QueryRow(query, roomID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get member count: %w", err)
	}
	return count, nil
}

// RoomExists checks if a room exists
func (r *RoomRepository) RoomExists(roomID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM rooms WHERE room_id = $1)`
	var exists bool
	err := r.db.QueryRow(query, roomID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check room existence: %w", err)
	}
	return exists, nil
}
