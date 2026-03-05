package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/arcane/arcanelink/internal/api-gateway/middleware"
	"github.com/arcane/arcanelink/pkg/logger"
	"github.com/arcane/arcanelink/pkg/models"
	authpb "github.com/arcane/arcanelink/pkg/proto/auth"
	messagepb "github.com/arcane/arcanelink/pkg/proto/message"
	roompb "github.com/arcane/arcanelink/pkg/proto/room"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type APIHandler struct {
	authClient    authpb.AuthServiceClient
	messageClient messagepb.MessageServiceClient
	roomClient    roompb.RoomServiceClient
	serverDomain  string
}

func NewAPIHandler(authAddr, messageAddr, roomAddr, serverDomain string) (*APIHandler, error) {
	authConn, err := grpc.Dial(authAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	messageConn, err := grpc.Dial(messageAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	roomConn, err := grpc.Dial(roomAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &APIHandler{
		authClient:    authpb.NewAuthServiceClient(authConn),
		messageClient: messagepb.NewMessageServiceClient(messageConn),
		roomClient:    roompb.NewRoomServiceClient(roomConn),
		serverDomain:  serverDomain,
	}, nil
}

// Auth handlers

func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	// Construct user_id from username and server domain
	userID := "@" + req.Username + ":" + h.serverDomain

	resp, err := h.authClient.Login(context.Background(), &authpb.LoginRequest{
		UserId:   userID,
		Password: req.Password,
	})
	if err != nil {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials")
		return
	}

	respondJSON(w, http.StatusOK, models.LoginResponse{
		AccessToken: resp.AccessToken,
		UserID:      resp.UserId,
		ExpiresIn:   resp.ExpiresIn,
	})
}

func (h *APIHandler) Register(w http.ResponseWriter, r *http.Request) {
	logger.Info("Register function called")

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Domain   string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	logger.Info("Register request decoded", zap.String("username", req.Username), zap.String("domain", req.Domain))

	// Validate input
	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Username and password are required")
		return
	}

	// Use server domain if not provided
	if req.Domain == "" {
		req.Domain = h.serverDomain
		logger.Info("Using default server domain", zap.String("domain", req.Domain))
	}

	// Construct user_id from username and domain
	userID := "@" + req.Username + ":" + req.Domain

	logger.Info("Attempting to register user", zap.String("user_id", userID))

	resp, err := h.authClient.CreateUser(context.Background(), &authpb.CreateUserRequest{
		UserId:   userID,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err), zap.String("user_id", userID))
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", fmt.Sprintf("Registration failed: %v", err))
		return
	}

	logger.Info("User created successfully", zap.String("user_id", userID))

	// After creating user, login to get access token
	loginResp, err := h.authClient.Login(context.Background(), &authpb.LoginRequest{
		UserId:   userID,
		Password: req.Password,
	})
	if err != nil {
		logger.Error("Failed to login after registration", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Registration succeeded but login failed")
		return
	}

	respondJSON(w, http.StatusOK, models.LoginResponse{
		AccessToken: loginResp.AccessToken,
		UserID:      resp.UserId,
		ExpiresIn:   loginResp.ExpiresIn,
	})
}

func (h *APIHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	_, err := h.authClient.Logout(context.Background(), &authpb.LogoutRequest{
		UserId: userID,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Logout failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Message handlers

func (h *APIHandler) SendDirect(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req models.SendDirectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	infoJSON, _ := json.Marshal(req.Content.Info)

	resp, err := h.messageClient.SendDirect(context.Background(), &messagepb.SendDirectRequest{
		Sender:    userID,
		Recipient: req.Recipient,
		Content: &messagepb.MessageContent{
			Msgtype:  string(req.Content.MsgType),
			Body:     req.Content.Body,
			Url:      req.Content.URL,
			InfoJson: string(infoJSON),
			GeoUri:   req.Content.GeoURI,
		},
	})
	if err != nil {
		logger.Error("Failed to send direct message", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Failed to send message")
		return
	}

	respondJSON(w, http.StatusOK, models.SendDirectResponse{
		MsgID:     resp.MsgId,
		Timestamp: resp.Timestamp,
	})
}

func (h *APIHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	since := r.URL.Query().Get("since")
	timeoutStr := r.URL.Query().Get("timeout")

	timeout := 30000 // default 30 seconds
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = t
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+5000)*time.Millisecond)
	defer cancel()

	resp, err := h.messageClient.Sync(ctx, &messagepb.SyncRequest{
		UserId:  userID,
		Since:   since,
		Timeout: int32(timeout),
	})
	if err != nil {
		logger.Error("Sync failed", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Sync failed")
		return
	}

	// Convert proto messages to models
	directMessages := make([]*models.DirectMessage, len(resp.DirectMessages))
	for i, msg := range resp.DirectMessages {
		var info map[string]interface{}
		if msg.Content.InfoJson != "" {
			json.Unmarshal([]byte(msg.Content.InfoJson), &info)
		}

		directMessages[i] = &models.DirectMessage{
			MsgID:     msg.MsgId,
			Sender:    msg.Sender,
			Recipient: msg.Recipient,
			Content: &models.MessageContent{
				MsgType: models.MessageType(msg.Content.Msgtype),
				Body:    msg.Content.Body,
				URL:     msg.Content.Url,
				Info:    info,
				GeoURI:  msg.Content.GeoUri,
			},
			Timestamp: msg.Timestamp,
		}
	}

	roomEvents := make([]*models.RoomEvent, len(resp.RoomEvents))
	for i, event := range resp.RoomEvents {
		var content map[string]interface{}
		if event.ContentJson != "" {
			json.Unmarshal([]byte(event.ContentJson), &content)
		}

		roomEvents[i] = &models.RoomEvent{
			EventID:   event.EventId,
			RoomID:    event.RoomId,
			Sender:    event.Sender,
			EventType: models.EventType(event.EventType),
			Content:   content,
			Timestamp: event.Timestamp,
		}
	}

	respondJSON(w, http.StatusOK, models.SyncResponse{
		NextToken:      resp.NextToken,
		DirectMessages: directMessages,
		RoomEvents:     roomEvents,
	})
}

func (h *APIHandler) GetDirectHistory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	peer := r.URL.Query().Get("peer")
	limitStr := r.URL.Query().Get("limit")
	before := r.URL.Query().Get("before")

	if peer == "" {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Missing peer parameter")
		return
	}

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	resp, err := h.messageClient.GetDirectHistory(context.Background(), &messagepb.GetDirectHistoryRequest{
		UserId: userID,
		Peer:   peer,
		Limit:  int32(limit),
		Before: before,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Failed to get history")
		return
	}

	messages := make([]*models.DirectMessage, len(resp.Messages))
	for i, msg := range resp.Messages {
		var info map[string]interface{}
		if msg.Content.InfoJson != "" {
			json.Unmarshal([]byte(msg.Content.InfoJson), &info)
		}

		messages[i] = &models.DirectMessage{
			MsgID:     msg.MsgId,
			Sender:    msg.Sender,
			Recipient: msg.Recipient,
			Content: &models.MessageContent{
				MsgType: models.MessageType(msg.Content.Msgtype),
				Body:    msg.Content.Body,
				URL:     msg.Content.Url,
				Info:    info,
				GeoURI:  msg.Content.GeoUri,
			},
			Timestamp: msg.Timestamp,
		}
	}

	respondJSON(w, http.StatusOK, models.DirectHistoryResponse{
		Messages:  messages,
		PrevToken: resp.PrevToken,
		HasMore:   resp.HasMore,
	})
}

func (h *APIHandler) SendRoomMessage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req models.SendRoomMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	infoJSON, _ := json.Marshal(req.Content.Info)

	resp, err := h.roomClient.SendRoomMessage(context.Background(), &roompb.SendRoomMessageRequest{
		RoomId: req.RoomID,
		Sender: userID,
		Msgtype: string(req.Content.MsgType),
		Body:    req.Content.Body,
		Url:     req.Content.URL,
		InfoJson: string(infoJSON),
	})
	if err != nil {
		logger.Error("Failed to send room message", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Failed to send message")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"event_id": resp.EventId,
		"timestamp": resp.Timestamp,
	})
}

// Room handlers

func (h *APIHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req struct {
		Name   string   `json:"name"`
		Invite []string `json:"invite,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	resp, err := h.roomClient.CreateRoom(context.Background(), &roompb.CreateRoomRequest{
		Creator: userID,
		Name:    req.Name,
		Invite:  req.Invite,
	})
	if err != nil {
		logger.Error("Failed to create room", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Failed to create room")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"room_id": resp.RoomId,
	})
}

func (h *APIHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req struct {
		RoomID string `json:"room_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	_, err := h.roomClient.JoinRoom(context.Background(), &roompb.JoinRoomRequest{
		RoomId: req.RoomID,
		UserId: userID,
	})
	if err != nil {
		logger.Error("Failed to join room", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Failed to join room")
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *APIHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req struct {
		RoomID string `json:"room_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	_, err := h.roomClient.LeaveRoom(context.Background(), &roompb.LeaveRoomRequest{
		RoomId: req.RoomID,
		UserId: userID,
	})
	if err != nil {
		logger.Error("Failed to leave room", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Failed to leave room")
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *APIHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	resp, err := h.roomClient.GetRooms(context.Background(), &roompb.GetRoomsRequest{
		UserId: userID,
	})
	if err != nil {
		logger.Error("Failed to get rooms", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Failed to get rooms")
		return
	}

	rooms := make([]map[string]interface{}, len(resp.Rooms))
	for i, room := range resp.Rooms {
		rooms[i] = map[string]interface{}{
			"room_id":      room.RoomId,
			"name":         room.Name,
			"topic":        room.Topic,
			"member_count": room.MemberCount,
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"rooms": rooms})
}

func (h *APIHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req struct {
		RoomID string `json:"room_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	_, err := h.roomClient.DeleteRoom(context.Background(), &roompb.DeleteRoomRequest{
		RoomId: req.RoomID,
		UserId: userID,
	})
	if err != nil {
		logger.Error("Failed to delete room", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", "Failed to delete room")
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   errorCode,
		"message": message,
	})
}
