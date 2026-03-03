package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arcane/arcanelink/pkg/database"
	"github.com/arcane/arcanelink/pkg/models"
)

type MessageRepository struct {
	db *database.DB
}

func NewMessageRepository(db *database.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// SaveDirectMessage saves a direct message to the database
func (r *MessageRepository) SaveDirectMessage(msg *models.DirectMessage) error {
	contentJSON, err := json.Marshal(msg.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	query := `
		INSERT INTO direct_messages (msg_id, sender, recipient, content, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = r.db.Exec(query,
		msg.MsgID,
		msg.Sender,
		msg.Recipient,
		contentJSON,
		msg.Timestamp,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to save direct message: %w", err)
	}
	return nil
}

// GetDirectMessage retrieves a direct message by ID
func (r *MessageRepository) GetDirectMessage(msgID string) (*models.DirectMessage, error) {
	query := `
		SELECT msg_id, sender, recipient, content, timestamp, created_at
		FROM direct_messages
		WHERE msg_id = $1
	`
	var msg models.DirectMessage
	var contentJSON []byte
	err := r.db.QueryRow(query, msgID).Scan(
		&msg.MsgID,
		&msg.Sender,
		&msg.Recipient,
		&contentJSON,
		&msg.Timestamp,
		&msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if err := json.Unmarshal(contentJSON, &msg.Content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal content: %w", err)
	}

	return &msg, nil
}

// GetDirectHistory retrieves direct message history between two users
func (r *MessageRepository) GetDirectHistory(userID, peerID string, limit int, beforeTimestamp int64) ([]*models.DirectMessage, error) {
	query := `
		SELECT msg_id, sender, recipient, content, timestamp, created_at
		FROM direct_messages
		WHERE ((sender = $1 AND recipient = $2) OR (sender = $2 AND recipient = $1))
		AND timestamp < $3
		ORDER BY timestamp DESC
		LIMIT $4
	`

	if beforeTimestamp == 0 {
		beforeTimestamp = time.Now().UnixMilli()
	}

	rows, err := r.db.Query(query, userID, peerID, beforeTimestamp, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query direct history: %w", err)
	}
	defer rows.Close()

	var messages []*models.DirectMessage
	for rows.Next() {
		var msg models.DirectMessage
		var contentJSON []byte
		if err := rows.Scan(
			&msg.MsgID,
			&msg.Sender,
			&msg.Recipient,
			&contentJSON,
			&msg.Timestamp,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if err := json.Unmarshal(contentJSON, &msg.Content); err != nil {
			return nil, fmt.Errorf("failed to unmarshal content: %w", err)
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}

// GetNewMessagesForUser retrieves new messages for a user since a timestamp
func (r *MessageRepository) GetNewMessagesForUser(userID string, sinceTimestamp int64) ([]*models.DirectMessage, error) {
	query := `
		SELECT msg_id, sender, recipient, content, timestamp, created_at
		FROM direct_messages
		WHERE recipient = $1 AND timestamp > $2
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Query(query, userID, sinceTimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to query new messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.DirectMessage
	for rows.Next() {
		var msg models.DirectMessage
		var contentJSON []byte
		if err := rows.Scan(
			&msg.MsgID,
			&msg.Sender,
			&msg.Recipient,
			&contentJSON,
			&msg.Timestamp,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if err := json.Unmarshal(contentJSON, &msg.Content); err != nil {
			return nil, fmt.Errorf("failed to unmarshal content: %w", err)
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}

// AddToMessageQueue adds a message to the delivery queue
func (r *MessageRepository) AddToMessageQueue(userID, messageType, messageID string) error {
	query := `
		INSERT INTO message_queue (user_id, message_type, message_id, created_at, delivered)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query, userID, messageType, messageID, time.Now(), false)
	if err != nil {
		return fmt.Errorf("failed to add to message queue: %w", err)
	}
	return nil
}

// GetUndeliveredMessages retrieves undelivered messages for a user
func (r *MessageRepository) GetUndeliveredMessages(userID string) ([]*models.MessageQueue, error) {
	query := `
		SELECT queue_id, user_id, message_type, message_id, created_at, delivered
		FROM message_queue
		WHERE user_id = $1 AND delivered = false
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query undelivered messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.MessageQueue
	for rows.Next() {
		var msg models.MessageQueue
		if err := rows.Scan(
			&msg.QueueID,
			&msg.UserID,
			&msg.MessageType,
			&msg.MessageID,
			&msg.CreatedAt,
			&msg.Delivered,
		); err != nil {
			return nil, fmt.Errorf("failed to scan queue message: %w", err)
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// MarkMessageDelivered marks a message as delivered
func (r *MessageRepository) MarkMessageDelivered(queueID int64) error {
	query := `UPDATE message_queue SET delivered = true WHERE queue_id = $1`
	_, err := r.db.Exec(query, queueID)
	if err != nil {
		return fmt.Errorf("failed to mark message delivered: %w", err)
	}
	return nil
}
