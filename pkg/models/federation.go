package models

// Federation API models

// FederationSendDirectRequest represents a federation direct message forward request
type FederationSendDirectRequest struct {
	Sender    string          `json:"sender" validate:"required"`
	Recipient string          `json:"recipient" validate:"required"`
	MsgID     string          `json:"msg_id" validate:"required"`
	Content   *MessageContent `json:"content" validate:"required"`
	Timestamp int64           `json:"timestamp" validate:"required"`
}

// FederationSendDirectResponse represents the response for federation direct message
type FederationSendDirectResponse struct {
	Success    bool  `json:"success"`
	ReceivedAt int64 `json:"received_at"`
}

// FederationSendRoomRequest represents a federation room event forward request
type FederationSendRoomRequest struct {
	RoomID    string                 `json:"room_id" validate:"required"`
	EventID   string                 `json:"event_id" validate:"required"`
	Sender    string                 `json:"sender" validate:"required"`
	EventType EventType              `json:"event_type" validate:"required"`
	Content   map[string]interface{} `json:"content" validate:"required"`
	Timestamp int64                  `json:"timestamp" validate:"required"`
}

// FederationSendRoomResponse represents the response for federation room event
type FederationSendRoomResponse struct {
	Success    bool  `json:"success"`
	ReceivedAt int64 `json:"received_at"`
}

// FederationQueryUserRequest represents a request to query user existence
type FederationQueryUserRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// FederationQueryUserResponse represents the response for user query
type FederationQueryUserResponse struct {
	UserID     string         `json:"user_id"`
	Exists     bool           `json:"exists"`
	Presence   PresenceStatus `json:"presence,omitempty"`
	LastActive int64          `json:"last_active,omitempty"`
}

// FederationInviteRequest represents a federation room invite request
type FederationInviteRequest struct {
	RoomID    string `json:"room_id" validate:"required"`
	Inviter   string `json:"inviter" validate:"required"`
	Invitee   string `json:"invitee" validate:"required"`
	RoomName  string `json:"room_name,omitempty"`
	RoomTopic string `json:"room_topic,omitempty"`
}

// FederationInviteResponse represents the response for federation invite
type FederationInviteResponse struct {
	Success bool `json:"success"`
}

// FederationJoinRoomRequest represents a federation join room request
type FederationJoinRoomRequest struct {
	RoomID string `json:"room_id" validate:"required"`
	UserID string `json:"user_id" validate:"required"`
}

// FederationJoinRoomResponse represents the response for federation join
type FederationJoinRoomResponse struct {
	Success   bool       `json:"success"`
	RoomState *RoomState `json:"room_state,omitempty"`
}

// FederationLeaveRoomRequest represents a federation leave room request
type FederationLeaveRoomRequest struct {
	RoomID string `json:"room_id" validate:"required"`
	UserID string `json:"user_id" validate:"required"`
}

// FederationLeaveRoomResponse represents the response for federation leave
type FederationLeaveRoomResponse struct {
	Success bool `json:"success"`
}

// FederationHealthResponse represents the health check response
type FederationHealthResponse struct {
	Status     string `json:"status"`
	Version    string `json:"version"`
	ServerName string `json:"server_name"`
	Timestamp  int64  `json:"timestamp"`
}

// ServerInfo represents discovered server information
type ServerInfo struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// WellKnownResponse represents .well-known/matrix/server response
type WellKnownResponse struct {
	Server string `json:"m.server"`
}
