package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arcane/arcanelink/internal/message/longpoll"
	"github.com/arcane/arcanelink/internal/message/repository"
	"github.com/arcane/arcanelink/pkg/logger"
	"github.com/arcane/arcanelink/pkg/models"
	"go.uber.org/zap"
)

type MessageService struct {
	repo        *repository.MessageRepository
	longPoll    *longpoll.Manager
	fedClient   FederationClient // Interface for federation service
}

// FederationClient interface for forwarding messages to federation service
type FederationClient interface {
	ForwardDirectMessage(msg *models.DirectMessage) error
}

func NewMessageService(repo *repository.MessageRepository, longPoll *longpoll.Manager) *MessageService {
	return &MessageService{
		repo:     repo,
		longPoll: longPoll,
	}
}

// SetFederationClient sets the federation client
func (s *MessageService) SetFederationClient(client FederationClient) {
	s.fedClient = client
}

// SendDirectMessage sends a direct message
func (s *MessageService) SendDirectMessage(sender, recipient string, content *models.MessageContent) (*models.DirectMessage, error) {
	// Generate message ID
	msgID := fmt.Sprintf("msg_%s", uuid.New().String())
	timestamp := time.Now().UnixMilli()

	msg := &models.DirectMessage{
		MsgID:     msgID,
		Sender:    sender,
		Recipient: recipient,
		Content:   content,
		Timestamp: timestamp,
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := s.repo.SaveDirectMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Add to message queue
	if err := s.repo.AddToMessageQueue(recipient, "direct", msgID); err != nil {
		logger.Error("Failed to add message to queue", zap.Error(err))
	}

	// Check if recipient is on the same server
	senderDomain := extractDomain(sender)
	recipientDomain := extractDomain(recipient)

	if senderDomain == recipientDomain {
		// Local delivery - notify via long poll
		s.longPoll.NotifyDirectMessage(recipient, msg)
		logger.Info("Message delivered locally", zap.String("msg_id", msgID))
	} else {
		// Remote delivery - forward to federation service
		if s.fedClient != nil {
			if err := s.fedClient.ForwardDirectMessage(msg); err != nil {
				logger.Error("Failed to forward message to federation", zap.Error(err))
				// Message is still saved locally, federation will retry
			}
		}
		logger.Info("Message forwarded to federation", zap.String("msg_id", msgID))
	}

	return msg, nil
}

// ReceiveFederatedMessage receives a message from federation service
func (s *MessageService) ReceiveFederatedMessage(msg *models.DirectMessage) error {
	// Save to database
	if err := s.repo.SaveDirectMessage(msg); err != nil {
		return fmt.Errorf("failed to save federated message: %w", err)
	}

	// Add to message queue
	if err := s.repo.AddToMessageQueue(msg.Recipient, "direct", msg.MsgID); err != nil {
		logger.Error("Failed to add federated message to queue", zap.Error(err))
	}

	// Notify recipient via long poll
	s.longPoll.NotifyDirectMessage(msg.Recipient, msg)

	logger.Info("Federated message received", zap.String("msg_id", msg.MsgID))
	return nil
}

// GetDirectHistory retrieves direct message history
func (s *MessageService) GetDirectHistory(userID, peerID string, limit int, beforeTimestamp int64) ([]*models.DirectMessage, bool, error) {
	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	messages, err := s.repo.GetDirectHistory(userID, peerID, limit+1, beforeTimestamp)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	return messages, hasMore, nil
}

// Sync performs a long polling sync for a user
func (s *MessageService) Sync(ctx context.Context, userID string, sinceTimestamp int64, timeout time.Duration) (*models.SyncResponse, error) {
	// First check for existing messages
	messages, err := s.repo.GetNewMessagesForUser(userID, sinceTimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to get new messages: %w", err)
	}

	if len(messages) > 0 {
		// Return immediately if there are new messages
		nextToken := fmt.Sprintf("t_%d", time.Now().UnixMilli())
		return &models.SyncResponse{
			NextToken:      nextToken,
			DirectMessages: messages,
			RoomEvents:     []*models.RoomEvent{},
		}, nil
	}

	// No new messages, subscribe to long poll with 1 second timeout
	// This reduces server load while still being responsive
	shortTimeout := 1 * time.Second
	logger.Debug("User subscribing to long poll", zap.String("user_id", userID))
	notification, err := s.longPoll.Subscribe(ctx, userID, shortTimeout)
	if err != nil {
		return nil, fmt.Errorf("long poll error: %w", err)
	}

	nextToken := fmt.Sprintf("t_%d", time.Now().UnixMilli())
	return &models.SyncResponse{
		NextToken:      nextToken,
		DirectMessages: notification.DirectMessages,
		RoomEvents:     notification.RoomEvents,
	}, nil
}

// extractDomain extracts domain from user ID (@user:domain.com -> domain.com)
func extractDomain(userID string) string {
	// Simple extraction, assumes format @username:domain.com
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
