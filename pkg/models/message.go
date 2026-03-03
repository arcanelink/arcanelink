package models

import "time"

// MessageType represents the type of message content
type MessageType string

const (
	MsgTypeText     MessageType = "m.text"
	MsgTypeImage    MessageType = "m.image"
	MsgTypeFile     MessageType = "m.file"
	MsgTypeAudio    MessageType = "m.audio"
	MsgTypeVideo    MessageType = "m.video"
	MsgTypeLocation MessageType = "m.location"
	MsgTypeReaction MessageType = "m.reaction"
)

// MessageContent represents the content of a message
type MessageContent struct {
	MsgType   MessageType            `json:"msgtype"`
	Body      string                 `json:"body"`
	URL       string                 `json:"url,omitempty"`
	Info      map[string]interface{} `json:"info,omitempty"`
	GeoURI    string                 `json:"geo_uri,omitempty"`
	RelatesTo *RelatesTo             `json:"m.relates_to,omitempty"`
}

// RelatesTo represents message relationships (reply, edit, delete, reaction)
type RelatesTo struct {
	RelType  string `json:"rel_type"`
	EventID  string `json:"event_id"`
	Key      string `json:"key,omitempty"` // For reactions
}

// DirectMessage represents a private chat message
type DirectMessage struct {
	MsgID     string          `json:"msg_id" db:"msg_id"`
	Sender    string          `json:"sender" db:"sender"`
	Recipient string          `json:"recipient" db:"recipient"`
	Content   *MessageContent `json:"content" db:"content"`
	Timestamp int64           `json:"timestamp" db:"timestamp"`
	CreatedAt time.Time       `json:"-" db:"created_at"`
}

// SendDirectRequest represents a request to send a direct message
type SendDirectRequest struct {
	Recipient string          `json:"recipient" validate:"required"`
	Content   *MessageContent `json:"content" validate:"required"`
}

// SendDirectResponse represents the response after sending a direct message
type SendDirectResponse struct {
	MsgID     string `json:"msg_id"`
	Timestamp int64  `json:"timestamp"`
}

// DirectHistoryRequest represents a request for direct message history
type DirectHistoryRequest struct {
	Peer   string `json:"peer" validate:"required"`
	Limit  int    `json:"limit,omitempty"`
	Before string `json:"before,omitempty"`
}

// DirectHistoryResponse represents the response for direct message history
type DirectHistoryResponse struct {
	Messages  []*DirectMessage `json:"messages"`
	PrevToken string           `json:"prev_token,omitempty"`
	HasMore   bool             `json:"has_more"`
}

// SyncRequest represents a sync request (long polling)
type SyncRequest struct {
	Since   string `json:"since,omitempty"`
	Timeout int    `json:"timeout,omitempty"` // milliseconds
}

// SyncResponse represents a sync response
type SyncResponse struct {
	NextToken       string            `json:"next_token"`
	DirectMessages  []*DirectMessage  `json:"direct_messages,omitempty"`
	RoomEvents      []*RoomEvent      `json:"room_events,omitempty"`
	PresenceUpdates []*PresenceUpdate `json:"presence_updates,omitempty"`
}

// MessageQueue represents a message in the delivery queue
type MessageQueue struct {
	QueueID     int64     `json:"queue_id" db:"queue_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	MessageType string    `json:"message_type" db:"message_type"` // "direct" or "room"
	MessageID   string    `json:"message_id" db:"message_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	Delivered   bool      `json:"delivered" db:"delivered"`
}
