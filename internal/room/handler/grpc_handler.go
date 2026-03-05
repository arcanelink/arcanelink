package handler

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/arcane/arcanelink/internal/room/service"
	"github.com/arcane/arcanelink/pkg/models"
	pb "github.com/arcane/arcanelink/pkg/proto/room"
)

type GRPCHandler struct {
	pb.UnimplementedRoomServiceServer
	roomService *service.RoomService
}

func NewGRPCHandler(roomService *service.RoomService) *GRPCHandler {
	return &GRPCHandler{
		roomService: roomService,
	}
}

func (h *GRPCHandler) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	room, err := h.roomService.CreateRoom(req.Creator, req.Name, req.Topic, req.Invite)
	if err != nil {
		return nil, err
	}

	return &pb.CreateRoomResponse{
		RoomId:    room.RoomID,
		CreatedAt: room.CreatedAt.Unix(),
	}, nil
}

func (h *GRPCHandler) JoinRoom(ctx context.Context, req *pb.JoinRoomRequest) (*pb.JoinRoomResponse, error) {
	err := h.roomService.JoinRoom(req.RoomId, req.UserId)
	if err != nil {
		return &pb.JoinRoomResponse{Success: false}, err
	}

	return &pb.JoinRoomResponse{
		Success:  true,
		JoinedAt: ctx.Value("joined_at").(int64),
	}, nil
}

func (h *GRPCHandler) LeaveRoom(ctx context.Context, req *pb.LeaveRoomRequest) (*pb.LeaveRoomResponse, error) {
	err := h.roomService.LeaveRoom(req.RoomId, req.UserId)
	if err != nil {
		return &pb.LeaveRoomResponse{Success: false}, err
	}

	return &pb.LeaveRoomResponse{Success: true}, nil
}

func (h *GRPCHandler) InviteUser(ctx context.Context, req *pb.InviteUserRequest) (*pb.InviteUserResponse, error) {
	err := h.roomService.InviteUser(req.RoomId, req.Inviter, req.Invitee)
	if err != nil {
		return &pb.InviteUserResponse{Success: false}, err
	}

	return &pb.InviteUserResponse{Success: true}, nil
}

func (h *GRPCHandler) SendRoomMessage(ctx context.Context, req *pb.SendRoomMessageRequest) (*pb.SendRoomMessageResponse, error) {
	var info map[string]interface{}
	if req.InfoJson != "" {
		json.Unmarshal([]byte(req.InfoJson), &info)
	}

	content := &models.MessageContent{
		MsgType: models.MessageType(req.Msgtype),
		Body:    req.Body,
		URL:     req.Url,
		Info:    info,
	}

	event, err := h.roomService.SendRoomMessage(req.RoomId, req.Sender, content)
	if err != nil {
		return nil, err
	}

	return &pb.SendRoomMessageResponse{
		EventId:   event.EventID,
		Timestamp: event.Timestamp,
	}, nil
}

func (h *GRPCHandler) GetRoomMembers(ctx context.Context, req *pb.GetRoomMembersRequest) (*pb.GetRoomMembersResponse, error) {
	members, err := h.roomService.GetRoomMembers(req.RoomId)
	if err != nil {
		return nil, err
	}

	pbMembers := make([]*pb.RoomMember, len(members))
	for i, m := range members {
		pbMembers[i] = &pb.RoomMember{
			UserId:   m.UserID,
			JoinedAt: m.JoinedAt.Unix(),
		}
	}

	return &pb.GetRoomMembersResponse{Members: pbMembers}, nil
}

func (h *GRPCHandler) GetRooms(ctx context.Context, req *pb.GetRoomsRequest) (*pb.GetRoomsResponse, error) {
	rooms, err := h.roomService.GetUserRooms(req.UserId)
	if err != nil {
		return nil, err
	}

	pbRooms := make([]*pb.RoomInfo, len(rooms))
	for i, r := range rooms {
		pbRooms[i] = &pb.RoomInfo{
			RoomId:       r.RoomID,
			Name:         r.Name,
			Topic:        r.Topic,
			MemberCount:  int32(r.MemberCount),
			LastActivity: r.LastActivity,
		}
	}

	return &pb.GetRoomsResponse{Rooms: pbRooms}, nil
}

func (h *GRPCHandler) GetRoomHistory(ctx context.Context, req *pb.GetRoomHistoryRequest) (*pb.GetRoomHistoryResponse, error) {
	var beforeTimestamp int64
	// Parse before token if needed

	events, hasMore, err := h.roomService.GetRoomHistory(
		req.RoomId,
		req.UserId,
		int(req.Limit),
		beforeTimestamp,
	)
	if err != nil {
		return nil, err
	}

	pbEvents := make([]*pb.RoomEvent, len(events))
	for i, e := range events {
		contentJSON, _ := json.Marshal(e.Content)
		pbEvents[i] = &pb.RoomEvent{
			EventId:     e.EventID,
			RoomId:      e.RoomID,
			Sender:      e.Sender,
			EventType:   string(e.EventType),
			ContentJson: string(contentJSON),
			Timestamp:   e.Timestamp,
		}
	}

	prevToken := ""
	if hasMore && len(events) > 0 {
		prevToken = "t_" + strconv.FormatInt(events[len(events)-1].Timestamp, 10)
	}

	return &pb.GetRoomHistoryResponse{
		Events:    pbEvents,
		PrevToken: prevToken,
		HasMore:   hasMore,
	}, nil
}

func (h *GRPCHandler) GetRoomState(ctx context.Context, req *pb.GetRoomStateRequest) (*pb.GetRoomStateResponse, error) {
	state, err := h.roomService.GetRoomState(req.RoomId)
	if err != nil {
		return nil, err
	}

	return &pb.GetRoomStateResponse{
		RoomId:      state.RoomID,
		Name:        state.Name,
		Topic:       state.Topic,
		Creator:     state.Creator,
		CreatedAt:   state.CreatedAt,
		MemberCount: int32(state.MemberCount),
	}, nil
}

func (h *GRPCHandler) DeleteRoom(ctx context.Context, req *pb.DeleteRoomRequest) (*pb.DeleteRoomResponse, error) {
	err := h.roomService.DeleteRoom(req.RoomId, req.UserId)
	if err != nil {
		return &pb.DeleteRoomResponse{Success: false}, err
	}

	return &pb.DeleteRoomResponse{Success: true}, nil
}
