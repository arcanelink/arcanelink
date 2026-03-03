package handler

import (
	"context"

	"github.com/arcane/arcanelink/internal/presence/service"
	"github.com/arcane/arcanelink/pkg/models"
	pb "github.com/arcane/arcanelink/pkg/proto/presence"
)

type GRPCHandler struct {
	pb.UnimplementedPresenceServiceServer
	presenceService *service.PresenceService
}

func NewGRPCHandler(presenceService *service.PresenceService) *GRPCHandler {
	return &GRPCHandler{
		presenceService: presenceService,
	}
}

func (h *GRPCHandler) UpdatePresence(ctx context.Context, req *pb.UpdatePresenceRequest) (*pb.UpdatePresenceResponse, error) {
	status := models.PresenceStatus(req.Status)
	err := h.presenceService.UpdatePresence(req.UserId, status, req.StatusMsg)
	if err != nil {
		return &pb.UpdatePresenceResponse{Success: false}, err
	}

	return &pb.UpdatePresenceResponse{Success: true}, nil
}

func (h *GRPCHandler) GetPresence(ctx context.Context, req *pb.GetPresenceRequest) (*pb.GetPresenceResponse, error) {
	presence, err := h.presenceService.GetPresence(req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetPresenceResponse{
		UserId:     presence.UserID,
		Presence:   string(presence.Status),
		LastActive: presence.LastActive.Unix(),
		StatusMsg:  presence.StatusMsg,
	}, nil
}

func (h *GRPCHandler) GetMultiplePresence(ctx context.Context, req *pb.GetMultiplePresenceRequest) (*pb.GetMultiplePresenceResponse, error) {
	presences, err := h.presenceService.GetMultiplePresence(req.UserIds)
	if err != nil {
		return nil, err
	}

	pbPresences := make([]*pb.PresenceInfo, len(presences))
	for i, p := range presences {
		pbPresences[i] = &pb.PresenceInfo{
			UserId:     p.UserID,
			Presence:   string(p.Status),
			LastActive: p.LastActive.Unix(),
		}
	}

	return &pb.GetMultiplePresenceResponse{Presences: pbPresences}, nil
}
