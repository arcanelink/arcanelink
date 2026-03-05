package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arcane/arcanelink/internal/room/repository"
	"github.com/arcane/arcanelink/pkg/logger"
	"github.com/arcane/arcanelink/pkg/models"
	"go.uber.org/zap"
)

type RoomService struct {
	repo      *repository.RoomRepository
	fedClient FederationClient
}

// FederationClient interface for forwarding room events
type FederationClient interface {
	ForwardRoomEvent(event *models.RoomEvent, members []string) error
}

func NewRoomService(repo *repository.RoomRepository) *RoomService {
	return &RoomService{
		repo: repo,
	}
}

// SetFederationClient sets the federation client
func (s *RoomService) SetFederationClient(client FederationClient) {
	s.fedClient = client
}

// CreateRoom creates a new room
func (s *RoomService) CreateRoom(creator, name, topic string, inviteList []string) (*models.Room, error) {
	// Generate room ID
	roomID := fmt.Sprintf("!%s:%s", uuid.New().String()[:8], extractDomain(creator))

	room := &models.Room{
		RoomID:    roomID,
		Creator:   creator,
		Name:      name,
		Topic:     topic,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create room
	if err := s.repo.CreateRoom(room); err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	// Add creator as member
	if err := s.repo.AddMember(roomID, creator); err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	// Create room creation event
	event := &models.RoomEvent{
		EventID:   fmt.Sprintf("evt_%s", uuid.New().String()),
		RoomID:    roomID,
		Sender:    creator,
		EventType: models.EventRoomCreate,
		Content: map[string]interface{}{
			"creator":      creator,
			"room_version": "1",
		},
		Timestamp: time.Now().UnixMilli(),
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveRoomEvent(event); err != nil {
		logger.Error("Failed to save room creation event", zap.Error(err))
	}

	// Invite users
	for _, userID := range inviteList {
		if err := s.InviteUser(roomID, creator, userID); err != nil {
			logger.Error("Failed to invite user", zap.String("user_id", userID), zap.Error(err))
		}
	}

	logger.Info("Room created", zap.String("room_id", roomID), zap.String("creator", creator))
	return room, nil
}

// JoinRoom adds a user to a room
func (s *RoomService) JoinRoom(roomID, userID string) error {
	// Check if room exists
	exists, err := s.repo.RoomExists(roomID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("room not found")
	}

	// Add member
	if err := s.repo.AddMember(roomID, userID); err != nil {
		return err
	}

	// Create join event
	event := &models.RoomEvent{
		EventID:   fmt.Sprintf("evt_%s", uuid.New().String()),
		RoomID:    roomID,
		Sender:    userID,
		EventType: models.EventRoomMember,
		Content: map[string]interface{}{
			"membership": "join",
			"user_id":    userID,
		},
		Timestamp: time.Now().UnixMilli(),
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveRoomEvent(event); err != nil {
		logger.Error("Failed to save join event", zap.Error(err))
	}

	// Notify other members
	members, _ := s.repo.GetMembers(roomID)
	if s.fedClient != nil {
		memberIDs := make([]string, len(members))
		for i, m := range members {
			memberIDs[i] = m.UserID
		}
		s.fedClient.ForwardRoomEvent(event, memberIDs)
	}

	logger.Info("User joined room", zap.String("room_id", roomID), zap.String("user_id", userID))
	return nil
}

// LeaveRoom removes a user from a room
func (s *RoomService) LeaveRoom(roomID, userID string) error {
	// Check if user is member
	isMember, err := s.repo.IsMember(roomID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("user is not a member of this room")
	}

	// Remove member
	if err := s.repo.RemoveMember(roomID, userID); err != nil {
		return err
	}

	// Create leave event
	event := &models.RoomEvent{
		EventID:   fmt.Sprintf("evt_%s", uuid.New().String()),
		RoomID:    roomID,
		Sender:    userID,
		EventType: models.EventRoomMember,
		Content: map[string]interface{}{
			"membership": "leave",
			"user_id":    userID,
		},
		Timestamp: time.Now().UnixMilli(),
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveRoomEvent(event); err != nil {
		logger.Error("Failed to save leave event", zap.Error(err))
	}

	logger.Info("User left room", zap.String("room_id", roomID), zap.String("user_id", userID))
	return nil
}

// InviteUser invites a user to a room
func (s *RoomService) InviteUser(roomID, inviter, invitee string) error {
	// Check if inviter is member
	isMember, err := s.repo.IsMember(roomID, inviter)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("inviter is not a member of this room")
	}

	// Create invite event
	event := &models.RoomEvent{
		EventID:   fmt.Sprintf("evt_%s", uuid.New().String()),
		RoomID:    roomID,
		Sender:    inviter,
		EventType: models.EventRoomMember,
		Content: map[string]interface{}{
			"membership": "invite",
			"user_id":    invitee,
			"inviter":    inviter,
		},
		Timestamp: time.Now().UnixMilli(),
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveRoomEvent(event); err != nil {
		logger.Error("Failed to save invite event", zap.Error(err))
	}

	// For simplicity, auto-join the invited user
	// In a real implementation, the invitee would need to accept
	if err := s.repo.AddMember(roomID, invitee); err != nil {
		return err
	}

	logger.Info("User invited to room", zap.String("room_id", roomID), zap.String("invitee", invitee))
	return nil
}

// SendRoomMessage sends a message to a room
func (s *RoomService) SendRoomMessage(roomID, sender string, content *models.MessageContent) (*models.RoomEvent, error) {
	// Check if sender is member
	isMember, err := s.repo.IsMember(roomID, sender)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("sender is not a member of this room")
	}

	// Create message event
	event := &models.RoomEvent{
		EventID:   fmt.Sprintf("evt_%s", uuid.New().String()),
		RoomID:    roomID,
		Sender:    sender,
		EventType: models.EventRoomMessage,
		Content: map[string]interface{}{
			"msgtype": content.MsgType,
			"body":    content.Body,
			"url":     content.URL,
			"info":    content.Info,
			"geo_uri": content.GeoURI,
		},
		Timestamp: time.Now().UnixMilli(),
		CreatedAt: time.Now(),
	}

	// Save event
	if err := s.repo.SaveRoomEvent(event); err != nil {
		return nil, fmt.Errorf("failed to save room message: %w", err)
	}

	// Notify members via federation
	members, _ := s.repo.GetMembers(roomID)
	if s.fedClient != nil && len(members) > 0 {
		memberIDs := make([]string, len(members))
		for i, m := range members {
			memberIDs[i] = m.UserID
		}
		if err := s.fedClient.ForwardRoomEvent(event, memberIDs); err != nil {
			logger.Error("Failed to forward room event", zap.Error(err))
		}
	}

	logger.Info("Room message sent", zap.String("room_id", roomID), zap.String("event_id", event.EventID))
	return event, nil
}

// GetRoomMembers retrieves all members of a room
func (s *RoomService) GetRoomMembers(roomID string) ([]*models.RoomMember, error) {
	return s.repo.GetMembers(roomID)
}

// GetUserRooms retrieves all rooms a user is a member of
func (s *RoomService) GetUserRooms(userID string) ([]*models.RoomInfo, error) {
	return s.repo.GetUserRooms(userID)
}

// GetRoomHistory retrieves room event history
func (s *RoomService) GetRoomHistory(roomID, userID string, limit int, beforeTimestamp int64) ([]*models.RoomEvent, bool, error) {
	// Check if user is member
	isMember, err := s.repo.IsMember(roomID, userID)
	if err != nil {
		return nil, false, err
	}
	if !isMember {
		return nil, false, fmt.Errorf("user is not a member of this room")
	}

	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	events, err := s.repo.GetRoomHistory(roomID, limit+1, beforeTimestamp)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(events) > limit
	if hasMore {
		events = events[:limit]
	}

	return events, hasMore, nil
}

// GetRoomState retrieves room state information
func (s *RoomService) GetRoomState(roomID string) (*models.RoomState, error) {
	room, err := s.repo.GetRoom(roomID)
	if err != nil {
		return nil, err
	}

	memberCount, err := s.repo.GetMemberCount(roomID)
	if err != nil {
		return nil, err
	}

	members, err := s.repo.GetMembers(roomID)
	if err != nil {
		return nil, err
	}

	memberIDs := make([]string, len(members))
	for i, m := range members {
		memberIDs[i] = m.UserID
	}

	return &models.RoomState{
		RoomID:      room.RoomID,
		Name:        room.Name,
		Topic:       room.Topic,
		Creator:     room.Creator,
		CreatedAt:   room.CreatedAt.Unix(),
		MemberCount: memberCount,
		Members:     memberIDs,
	}, nil
}

// extractDomain extracts domain from user ID
func extractDomain(userID string) string {
	if len(userID) > 0 && userID[0] == '@' {
		parts := []rune(userID[1:])
		for i, r := range parts {
			if r == ':' && i+1 < len(parts) {
				return string(parts[i+1:])
			}
		}
	}
	return "localhost"
}

// DeleteRoom deletes a room
func (s *RoomService) DeleteRoom(roomID, userID string) error {
	// Check if room exists
	room, err := s.repo.GetRoom(roomID)
	if err != nil {
		return fmt.Errorf("room not found: %w", err)
	}

	// Only the creator can delete the room
	if room.Creator != userID {
		return fmt.Errorf("only the room creator can delete the room")
	}

	// Delete the room
	if err := s.repo.DeleteRoom(roomID); err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	logger.Info("Room deleted", zap.String("room_id", roomID), zap.String("user_id", userID))
	return nil
}
