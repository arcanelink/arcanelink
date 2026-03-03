package models

import "time"

// EventType represents the type of room event
type EventType string

const (
	EventRoomMessage EventType = "m.room.message"
	EventRoomCreate  EventType = "m.room.create"
	EventRoomName    EventType = "m.room.name"
	EventRoomTopic   EventType = "m.room.topic"
	EventRoomMember  EventType = "m.room.member"
)

// MembershipType represents room membership status
type MembershipType string

const (
	MembershipJoin   MembershipType = "join"
	MembershipLeave  MembershipType = "leave"
	MembershipInvite MembershipType = "invite"
	MembershipKick   MembershipType = "kick"
)

// Room represents a group chat room
type Room struct {
	RoomID    string    `json:"room_id" db:"room_id"`
	Creator   string    `json:"creator" db:"creator"`
	Name      string    `json:"name,omitempty" db:"name"`
	Topic     string    `json:"topic,omitempty" db:"topic"`
	AvatarURL string    `json:"avatar_url,omitempty" db:"avatar_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RoomMember represents a room membership
type RoomMember struct {
	RoomID   string    `json:"room_id" db:"room_id"`
	UserID   string    `json:"user_id" db:"user_id"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

// RoomEvent represents an event in a room
type RoomEvent struct {
	EventID   string                 `json:"event_id" db:"event_id"`
	RoomID    string                 `json:"room_id" db:"room_id"`
	Sender    string                 `json:"sender" db:"sender"`
	EventType EventType              `json:"event_type" db:"event_type"`
	Content   map[string]interface{} `json:"content" db:"content"`
	Timestamp int64                  `json:"timestamp" db:"timestamp"`
	CreatedAt time.Time              `json:"-" db:"created_at"`
}

// CreateRoomRequest represents a request to create a room
type CreateRoomRequest struct {
	Name   string   `json:"name" validate:"required,max=200"`
	Topic  string   `json:"topic,omitempty"`
	Invite []string `json:"invite,omitempty"`
}

// CreateRoomResponse represents the response after creating a room
type CreateRoomResponse struct {
	RoomID    string `json:"room_id"`
	CreatedAt int64  `json:"created_at"`
}

// JoinRoomRequest represents a request to join a room
type JoinRoomRequest struct {
	RoomID string `json:"room_id" validate:"required"`
}

// JoinRoomResponse represents the response after joining a room
type JoinRoomResponse struct {
	Success  bool  `json:"success"`
	JoinedAt int64 `json:"joined_at"`
}

// LeaveRoomRequest represents a request to leave a room
type LeaveRoomRequest struct {
	RoomID string `json:"room_id" validate:"required"`
}

// LeaveRoomResponse represents the response after leaving a room
type LeaveRoomResponse struct {
	Success bool `json:"success"`
}

// InviteUserRequest represents a request to invite a user to a room
type InviteUserRequest struct {
	RoomID string `json:"room_id" validate:"required"`
	UserID string `json:"user_id" validate:"required"`
}

// InviteUserResponse represents the response after inviting a user
type InviteUserResponse struct {
	Success bool `json:"success"`
}

// GetRoomMembersResponse represents the response for getting room members
type GetRoomMembersResponse struct {
	Members []*RoomMember `json:"members"`
}

// GetRoomsResponse represents the response for getting user's rooms
type GetRoomsResponse struct {
	Rooms []*RoomInfo `json:"rooms"`
}

// RoomInfo represents summary information about a room
type RoomInfo struct {
	RoomID       string `json:"room_id"`
	Name         string `json:"name,omitempty"`
	Topic        string `json:"topic,omitempty"`
	MemberCount  int    `json:"member_count"`
	LastActivity int64  `json:"last_activity"`
}

// SendRoomMessageRequest represents a request to send a room message
type SendRoomMessageRequest struct {
	RoomID  string          `json:"room_id" validate:"required"`
	Content *MessageContent `json:"content" validate:"required"`
}

// SendRoomMessageResponse represents the response after sending a room message
type SendRoomMessageResponse struct {
	EventID   string `json:"event_id"`
	Timestamp int64  `json:"timestamp"`
}

// GetRoomHistoryRequest represents a request for room history
type GetRoomHistoryRequest struct {
	RoomID string `json:"room_id" validate:"required"`
	Limit  int    `json:"limit,omitempty"`
	Before string `json:"before,omitempty"`
}

// GetRoomHistoryResponse represents the response for room history
type GetRoomHistoryResponse struct {
	Events    []*RoomEvent `json:"events"`
	PrevToken string       `json:"prev_token,omitempty"`
	HasMore   bool         `json:"has_more"`
}

// RoomState represents the state of a room
type RoomState struct {
	RoomID      string   `json:"room_id"`
	Name        string   `json:"name,omitempty"`
	Topic       string   `json:"topic,omitempty"`
	Creator     string   `json:"creator"`
	CreatedAt   int64    `json:"created_at"`
	MemberCount int      `json:"member_count"`
	Members     []string `json:"members,omitempty"`
}
