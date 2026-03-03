package forwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/arcane/arcanelink/internal/federation/discovery"
	"github.com/arcane/arcanelink/pkg/logger"
	"github.com/arcane/arcanelink/pkg/models"
	"go.uber.org/zap"
)

type MessageForwarder struct {
	httpClient *http.Client
	resolver   *discovery.ServerResolver
}

func NewMessageForwarder(resolver *discovery.ServerResolver) *MessageForwarder {
	return &MessageForwarder{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		resolver: resolver,
	}
}

// ForwardDirectMessage forwards a direct message to a remote server
func (f *MessageForwarder) ForwardDirectMessage(msg *models.DirectMessage) error {
	recipientDomain := extractDomain(msg.Recipient)

	// Resolve server
	serverInfo, err := f.resolver.Resolve(recipientDomain)
	if err != nil {
		return fmt.Errorf("failed to resolve server: %w", err)
	}

	// Prepare request
	reqBody := &models.FederationSendDirectRequest{
		Sender:    msg.Sender,
		Recipient: msg.Recipient,
		MsgID:     msg.MsgID,
		Content:   msg.Content,
		Timestamp: msg.Timestamp,
	}

	// Send with retry
	return f.sendWithRetry(serverInfo, "/_fed/v1/send_direct", reqBody, 5)
}

// ForwardRoomEvent forwards a room event to remote servers
func (f *MessageForwarder) ForwardRoomEvent(event *models.RoomEvent, memberIDs []string) error {
	// Group members by domain
	domainMembers := make(map[string][]string)
	for _, memberID := range memberIDs {
		domain := extractDomain(memberID)
		domainMembers[domain] = append(domainMembers[domain], memberID)
	}

	// Forward to each domain
	var lastErr error
	successCount := 0

	for domain, members := range domainMembers {
		// Skip local domain
		if domain == "localhost" {
			continue
		}

		serverInfo, err := f.resolver.Resolve(domain)
		if err != nil {
			logger.Error("Failed to resolve server for room event",
				zap.String("domain", domain),
				zap.Error(err))
			lastErr = err
			continue
		}

		reqBody := &models.FederationSendRoomRequest{
			RoomID:    event.RoomID,
			EventID:   event.EventID,
			Sender:    event.Sender,
			EventType: event.EventType,
			Content:   event.Content,
			Timestamp: event.Timestamp,
		}

		if err := f.sendWithRetry(serverInfo, "/_fed/v1/send_room", reqBody, 3); err != nil {
			logger.Error("Failed to forward room event",
				zap.String("domain", domain),
				zap.Int("member_count", len(members)),
				zap.Error(err))
			lastErr = err
		} else {
			successCount++
		}
	}

	if successCount == 0 && lastErr != nil {
		return lastErr
	}

	return nil
}

// sendWithRetry sends a request with exponential backoff retry
func (f *MessageForwarder) sendWithRetry(serverInfo *models.ServerInfo, path string, body interface{}, maxRetries int) error {
	backoff := []time.Duration{0, 5 * time.Second, 30 * time.Second, 5 * time.Minute, 5 * time.Minute}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 && i < len(backoff) {
			logger.Debug("Retrying federation request",
				zap.Int("attempt", i+1),
				zap.Duration("backoff", backoff[i]))
			time.Sleep(backoff[i])
		}

		err := f.send(serverInfo, path, body)
		if err == nil {
			if i > 0 {
				logger.Info("Federation request succeeded after retry",
					zap.Int("attempts", i+1))
			}
			return nil
		}

		lastErr = err
		logger.Warn("Federation request failed",
			zap.Int("attempt", i+1),
			zap.Error(err))
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// send sends a single HTTP request
func (f *MessageForwarder) send(serverInfo *models.ServerInfo, path string, body interface{}) error {
	url := fmt.Sprintf("https://%s:%d%s", serverInfo.Host, serverInfo.Port, path)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	logger.Debug("Federation request successful",
		zap.String("url", url),
		zap.Int("status", resp.StatusCode))

	return nil
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
