package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/arcane/arcanelink/internal/api-gateway/middleware"
	"github.com/arcane/arcanelink/pkg/logger"
	"github.com/arcane/arcanelink/pkg/models"
	authpb "github.com/arcane/arcanelink/pkg/proto/auth"
	messagepb "github.com/arcane/arcanelink/pkg/proto/message"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type APIHandler struct {
	authClient    authpb.AuthServiceClient
	messageClient messagepb.MessageServiceClient
}

func NewAPIHandler(authAddr, messageAddr string) (*APIHandler, error) {
	authConn, err := grpc.Dial(authAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	messageConn, err := grpc.Dial(messageAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &APIHandler{
		authClient:    authpb.NewAuthServiceClient(authConn),
		messageClient: messagepb.NewMessageServiceClient(messageConn),
	}, nil
}

// Auth handlers

func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	resp, err := h.authClient.Login(context.Background(), &authpb.LoginRequest{
		UserId:   req.UserID,
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
