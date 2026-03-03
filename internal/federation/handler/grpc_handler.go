package handler

import (
	"context"

	"github.com/arcane/arcanelink/internal/federation/service"
	"github.com/arcane/arcanelink/pkg/models"
	pb "github.com/arcane/arcanelink/pkg/proto/federation"
)

type GRPCHandler struct {
	pb.UnimplementedFederationServiceServer
	federationService *service.FederationService
}

func NewGRPCHandler(federationService *service.FederationService) *GRPCHandler {
	return &GRPCHandler{
		federationService: federationService,
	}
}

func (h *GRPCHandler) ForwardDirectMessage(ctx context.Context, req *pb.ForwardDirectMessageRequest) (*pb.ForwardDirectMessageResponse, error) {
	msg := &models.DirectMessage{
		MsgID:     req.MsgId,
		Sender:    req.Sender,
		Recipient: req.Recipient,
		Content: &models.MessageContent{
			MsgType: models.MessageType(req.Msgtype),
			Body:    req.Body,
			URL:     req.Url,
		},
		Timestamp: req.Timestamp,
	}

	err := h.federationService.ForwardDirectMessage(msg)
	if err != nil {
		return &pb.ForwardDirectMessageResponse{Success: false}, err
	}

	return &pb.ForwardDirectMessageResponse{Success: true}, nil
}

func (h *GRPCHandler) ForwardRoomEvent(ctx context.Context, req *pb.ForwardRoomEventRequest) (*pb.ForwardRoomEventResponse, error) {
	event := &models.RoomEvent{
		EventID:   req.EventId,
		RoomID:    req.RoomId,
		Sender:    req.Sender,
		EventType: models.EventType(req.EventType),
		Timestamp: req.Timestamp,
	}

	err := h.federationService.ForwardRoomEvent(event, req.MemberIds)
	if err != nil {
		return &pb.ForwardRoomEventResponse{
			Success:        false,
			DeliveredCount: 0,
		}, err
	}

	return &pb.ForwardRoomEventResponse{
		Success:        true,
		DeliveredCount: int32(len(req.MemberIds)),
	}, nil
}

func (h *GRPCHandler) QueryUser(ctx context.Context, req *pb.QueryUserRequest) (*pb.QueryUserResponse, error) {
	// For now, return a simple response
	// In a real implementation, this would query the user's homeserver
	return &pb.QueryUserResponse{
		UserId:  req.UserId,
		Exists:  true,
		Presence: "online",
	}, nil
}
