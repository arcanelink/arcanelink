package longpoll

import (
	"context"
	"sync"
	"time"

	"github.com/arcane/arcanelink/pkg/logger"
	"github.com/arcane/arcanelink/pkg/models"
	"go.uber.org/zap"
)

// Subscriber represents a long polling subscriber
type Subscriber struct {
	UserID  string
	Channel chan *Notification
	Timeout time.Time
}

// Notification represents a message notification
type Notification struct {
	DirectMessages []*models.DirectMessage
	RoomEvents     []*models.RoomEvent
}

// Manager manages long polling subscriptions
type Manager struct {
	subscribers sync.Map // map[string]*Subscriber
	mu          sync.RWMutex
}

// NewManager creates a new long polling manager
func NewManager() *Manager {
	m := &Manager{}
	// Start cleanup goroutine
	go m.cleanupExpired()
	return m
}

// Subscribe subscribes a user to long polling
func (m *Manager) Subscribe(ctx context.Context, userID string, timeout time.Duration) (*Notification, error) {
	// Create subscriber
	sub := &Subscriber{
		UserID:  userID,
		Channel: make(chan *Notification, 100),
		Timeout: time.Now().Add(timeout),
	}

	// Store subscriber
	m.subscribers.Store(userID, sub)
	defer m.subscribers.Delete(userID)

	logger.Debug("User subscribed to long poll", zap.String("user_id", userID), zap.Duration("timeout", timeout))

	// Wait for notification or timeout
	select {
	case notification := <-sub.Channel:
		logger.Debug("Long poll notification sent", zap.String("user_id", userID))
		return notification, nil
	case <-time.After(timeout):
		logger.Debug("Long poll timeout", zap.String("user_id", userID))
		return &Notification{
			DirectMessages: []*models.DirectMessage{},
			RoomEvents:     []*models.RoomEvent{},
		}, nil
	case <-ctx.Done():
		logger.Debug("Long poll cancelled", zap.String("user_id", userID))
		return nil, ctx.Err()
	}
}

// Notify notifies a user of new messages
func (m *Manager) Notify(userID string, notification *Notification) {
	if sub, ok := m.subscribers.Load(userID); ok {
		subscriber := sub.(*Subscriber)
		select {
		case subscriber.Channel <- notification:
			logger.Debug("Notification sent to subscriber", zap.String("user_id", userID))
		default:
			logger.Warn("Subscriber channel full, dropping notification", zap.String("user_id", userID))
		}
	} else {
		logger.Debug("No subscriber found for user", zap.String("user_id", userID))
	}
}

// NotifyDirectMessage notifies a user of a new direct message
func (m *Manager) NotifyDirectMessage(userID string, message *models.DirectMessage) {
	notification := &Notification{
		DirectMessages: []*models.DirectMessage{message},
		RoomEvents:     []*models.RoomEvent{},
	}
	m.Notify(userID, notification)
}

// NotifyRoomEvent notifies a user of a new room event
func (m *Manager) NotifyRoomEvent(userID string, event *models.RoomEvent) {
	notification := &Notification{
		DirectMessages: []*models.DirectMessage{},
		RoomEvents:     []*models.RoomEvent{event},
	}
	m.Notify(userID, notification)
}

// cleanupExpired removes expired subscribers
func (m *Manager) cleanupExpired() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.subscribers.Range(func(key, value interface{}) bool {
			sub := value.(*Subscriber)
			if now.After(sub.Timeout) {
				m.subscribers.Delete(key)
				logger.Debug("Removed expired subscriber", zap.String("user_id", sub.UserID))
			}
			return true
		})
	}
}

// GetSubscriberCount returns the number of active subscribers
func (m *Manager) GetSubscriberCount() int {
	count := 0
	m.subscribers.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// IsSubscribed checks if a user is currently subscribed
func (m *Manager) IsSubscribed(userID string) bool {
	_, ok := m.subscribers.Load(userID)
	return ok
}
