package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/arcane/arcanelink/internal/message/service"
	"github.com/arcane/arcanelink/pkg/models"
	pb "github.com/arcane/arcanelink/pkg/proto/message"
)

type GRPCHandler struct {
	pb.UnimplementedMessageServiceServer
	messageService *service.MessageService
}

func NewGRPCHandler(messageService *service.MessageService) *GRPCHandler {
	return &GRPCHandler{
		messageService: messageService,
	}
}

func (h *GRPCHandler) SendDirect(ctx context.Context, req *pb.SendDirectRequest) (*pb.SendDirectResponse, error) {
	content := &models.MessageContent{
		MsgType: models.MessageType(req.Content.Msgtype),
		Body:    req.Content.Body,
		URL:     req.Content.Url,
		GeoURI:  req.Content.GeoUri,
	}

	// Parse info JSON if provided
	if req.Content.InfoJson != "" {
		var info map[string]interface{}
		if err := json.Unmarshal([]byte(req.Content.InfoJson), &info); err == nil {
			content.Info = info
		}
	}

	msg, err := h.messageService.SendDirectMessage(req.Sender, req.Recipient, content)
	if err != nil {
		return nil, err
	}

	return &pb.SendDirectResponse{
		MsgId:     msg.MsgID,
		Timestamp: msg.Timestamp,
	}, nil
}

func (h *GRPCHandler) GetDirectHistory(ctx context.Context, req *pb.GetDirectHistoryRequest) (*pb.GetDirectHistoryResponse, error) {
	var beforeTimestamp int64
	if req.Before != "" {
		// Parse before token (format: t_timestamp)
		// For simplicity, we'll use 0 if parsing fails
		beforeTimestamp = 0
	}

	messages, hasMore, err := h.messageService.GetDirectHistory(
		req.UserId,
		req.Peer,
		int(req.Limit),
		beforeTimestamp,
	)
	if err != nil {
		return nil, err
	}

	pbMessages := make([]*pb.DirectMessage, len(messages))
	for i, msg := range messages {
		pbMessages[i] = convertToProtoMessage(msg)
	}

	prevToken := ""
	if hasMore && len(messages) > 0 {
		prevToken = "t_" + strconv.FormatInt(messages[len(messages)-1].Timestamp, 10)
	}

	return &pb.GetDirectHistoryResponse{
		Messages:  pbMessages,
		PrevToken: prevToken,
		HasMore:   hasMore,
	}, nil
}

func (h *GRPCHandler) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	var sinceTimestamp int64
	if req.Since != "" {
		// Parse since token
		sinceTimestamp = 0
	}

	timeout := time.Duration(req.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	syncResp, err := h.messageService.Sync(ctx, req.UserId, sinceTimestamp, timeout)
	if err != nil {
		return nil, err
	}

	pbMessages := make([]*pb.DirectMessage, len(syncResp.DirectMessages))
	for i, msg := range syncResp.DirectMessages {
		pbMessages[i] = convertToProtoMessage(msg)
	}

	pbEvents := make([]*pb.RoomEvent, len(syncResp.RoomEvents))
	for i, event := range syncResp.RoomEvents {
		contentJSON, _ := json.Marshal(event.Content)
		pbEvents[i] = &pb.RoomEvent{
			EventId:     event.EventID,
			RoomId:      event.RoomID,
			Sender:      event.Sender,
			EventType:   string(event.EventType),
			ContentJson: string(contentJSON),
			Timestamp:   event.Timestamp,
		}
	}

	return &pb.SyncResponse{
		NextToken:      syncResp.NextToken,
		DirectMessages: pbMessages,
		RoomEvents:     pbEvents,
	}, nil
}

func (h *GRPCHandler) ReceiveFederatedDirect(ctx context.Context, req *pb.ReceiveFederatedDirectRequest) (*pb.ReceiveFederatedDirectResponse, error) {
	content := &models.MessageContent{
		MsgType: models.MessageType(req.Content.Msgtype),
		Body:    req.Content.Body,
		URL:     req.Content.Url,
		GeoURI:  req.Content.GeoUri,
	}

	if req.Content.InfoJson != "" {
		var info map[string]interface{}
		if err := json.Unmarshal([]byte(req.Content.InfoJson), &info); err == nil {
			content.Info = info
		}
	}

	msg := &models.DirectMessage{
		MsgID:     req.MsgId,
		Sender:    req.Sender,
		Recipient: req.Recipient,
		Content:   content,
		Timestamp: req.Timestamp,
	}

	err := h.messageService.ReceiveFederatedMessage(msg)
	if err != nil {
		return &pb.ReceiveFederatedDirectResponse{
			Success:    false,
			ReceivedAt: time.Now().UnixMilli(),
		}, err
	}

	return &pb.ReceiveFederatedDirectResponse{
		Success:    true,
		ReceivedAt: time.Now().UnixMilli(),
	}, nil
}

func convertToProtoMessage(msg *models.DirectMessage) *pb.DirectMessage {
	infoJSON, _ := json.Marshal(msg.Content.Info)

	return &pb.DirectMessage{
		MsgId:     msg.MsgID,
		Sender:    msg.Sender,
		Recipient: msg.Recipient,
		Content: &pb.MessageContent{
			Msgtype:  string(msg.Content.MsgType),
			Body:     msg.Content.Body,
			Url:      msg.Content.URL,
			InfoJson: string(infoJSON),
			GeoUri:   msg.Content.GeoURI,
		},
		Timestamp: msg.Timestamp,
	}
}
