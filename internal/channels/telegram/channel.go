// Package telegram implements Telegram channel for Aigo.
// Features: long polling, auto-reconnect, heartbeat, exponential backoff.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hermes-v2/aigo/internal/gateway"
)

// Channel implements the gateway.Channel interface for Telegram.
type Channel struct {
	token     string
	baseURL   string
	offset    int
	onMessage func(gateway.Message) error
	client    *http.Client
	running   bool
	mu        sync.Mutex

	// Stats
	connected    bool
	lastUpdate   time.Time
	errorCount   int
	reconnectDelay time.Duration
}

// New creates a new Telegram channel.
func New(token string) *Channel {
	return &Channel{
		token:   token,
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
		client:  &http.Client{Timeout: 90 * time.Second}, // Longer than polling timeout
	}
}

func (c *Channel) Name() string { return "telegram" }

func (c *Channel) Start(ctx context.Context, onMessage func(gateway.Message) error) error {
	c.onMessage = onMessage

	// Verify token with retry
	if err := c.verifyToken(ctx); err != nil {
		return fmt.Errorf("telegram verify: %w", err)
	}

	c.connected = true
	c.reconnectDelay = 1 * time.Second
	log.Println("📡 Telegram bot connected")

	// Start heartbeat goroutine
	go c.heartbeat(ctx)

	// Main polling loop with auto-reconnect
	c.running = true
	for c.running {
		select {
		case <-ctx.Done():
			c.running = false
			c.connected = false
			return nil
		default:
		}

		if err := c.pollUpdates(ctx); err != nil {
			c.errorCount++
			log.Printf("Telegram poll error (#%d): %v", c.errorCount, err)
			c.connected = false

			// Exponential backoff on errors
			log.Printf("Reconnecting in %v...", c.reconnectDelay)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(c.reconnectDelay):
			}

			// Exponential backoff: 1s → 2s → 4s → 8s → max 30s
			c.reconnectDelay *= 2
			if c.reconnectDelay > 30*time.Second {
				c.reconnectDelay = 30 * time.Second
			}

			// Verify token on reconnect
			if err := c.verifyToken(ctx); err != nil {
				log.Printf("Reconnect verify failed: %v", err)
				continue
			}
			c.connected = true
			c.errorCount = 0
			c.reconnectDelay = 1 * time.Second
			log.Println("Telegram reconnected ✓")
		}
	}
	return nil
}

// verifyToken checks if the bot token is valid.
func (c *Channel) verifyToken(ctx context.Context) error {
	var result map[string]interface{}
	if err := c.apiCall(ctx, "getMe", nil, &result); err != nil {
		return err
	}
	if ok, _ := result["ok"].(bool); !ok {
		return fmt.Errorf("invalid bot token")
	}
	return nil
}

// heartbeat sends periodic "typing" status to keep connection alive and detect dead connections.
func (c *Channel) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !c.connected {
				continue
			}
			// Simple health check — just verify getMe still works
			var result map[string]interface{}
			if err := c.apiCall(ctx, "getMe", nil, &result); err != nil {
				log.Printf("Heartbeat failed: %v", err)
				c.connected = false
			}
		}
	}
}

// pollUpdates fetches updates using long polling.
func (c *Channel) pollUpdates(ctx context.Context) error {
	payload := map[string]interface{}{
		"timeout": 30, // Telegram long-poll: wait up to 30s for new messages
		"allowed_updates": []string{"message", "edited_message", "callback_query"},
	}
	if c.offset > 0 {
		payload["offset"] = c.offset
	}

	var result map[string]interface{}
	if err := c.apiCall(ctx, "getUpdates", payload, &result); err != nil {
		return err
	}

	if ok, _ := result["ok"].(bool); !ok {
		return fmt.Errorf("api returned ok=false")
	}

	updates, _ := result["result"].([]interface{})
	for _, u := range updates {
		m, ok := u.(map[string]interface{})
		if !ok {
			continue
		}

		updateID, _ := m["update_id"].(float64)
		c.offset = int(updateID) + 1
		c.lastUpdate = time.Now()

		// Handle regular messages
		if msg, ok := m["message"].(map[string]interface{}); ok {
			c.handleMessage(msg)
		}

		// Handle edited messages
		if msg, ok := m["edited_message"].(map[string]interface{}); ok {
			c.handleMessage(msg)
		}

		// Handle callback queries (inline keyboard buttons)
		if cb, ok := m["callback_query"].(map[string]interface{}); ok {
			c.handleCallbackQuery(cb)
		}
	}

	return nil
}

func (c *Channel) handleMessage(msg map[string]interface{}) {
	text, _ := msg["text"].(string)
	if text == "" {
		// Check for caption (photos/documents)
		text, _ = msg["caption"].(string)
	}
	if text == "" {
		return
	}

	chat, _ := msg["chat"].(map[string]interface{})
	chatID, _ := chat["id"].(float64)
	from, _ := msg["from"].(map[string]interface{})
	userID, _ := from["id"].(float64)

	gwMsg := gateway.Message{
		Channel:  "telegram",
		ChatID:   strconv.FormatFloat(chatID, 'f', 0, 64),
		SenderID: strconv.FormatFloat(userID, 'f', 0, 64),
		Text:     text,
	}

	if c.onMessage != nil {
		go func(m gateway.Message) {
			if err := c.onMessage(m); err != nil {
				log.Printf("Telegram handler error: %v", err)
			}
		}(gwMsg)
	}
}

func (c *Channel) handleCallbackQuery(cb map[string]interface{}) {
	data, _ := cb["data"].(string)
	if data == "" {
		return
	}
	// Acknowledge callback query
	go func() {
		var result map[string]interface{}
		payload := map[string]interface{}{
			"callback_query_id": cb["id"],
		}
		c.apiCall(context.Background(), "answerCallbackQuery", payload, &result)
	}()
}

func (c *Channel) Send(chatID string, text string) error {
	if len(text) > 4096 {
		text = text[:4093] + "..."
	}

	// Try markdown first, fall back to plain text
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	var result map[string]interface{}
	err := c.apiCall(context.Background(), "sendMessage", payload, &result)

	// If markdown fails, try without parse_mode
	if err != nil || (result != nil && result["ok"] == false) {
		payload = map[string]interface{}{
			"chat_id": chatID,
			"text":    text,
		}
		return c.apiCall(context.Background(), "sendMessage", payload, &result)
	}
	return nil
}

// SendTyping sends a "typing..." chat action.
func (c *Channel) SendTyping(chatID string) {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"action":  "typing",
	}
	var result map[string]interface{}
	c.apiCall(context.Background(), "sendChatAction", payload, &result)
}

// IsConnected returns whether the Telegram connection is healthy.
func (c *Channel) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// Stats returns connection statistics.
func (c *Channel) Stats() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return map[string]interface{}{
		"connected":   c.connected,
		"last_update": c.lastUpdate.Format(time.RFC3339),
		"error_count": c.errorCount,
		"offset":      c.offset,
	}
}

func (c *Channel) Stop() error {
	c.running = false
	c.connected = false
	return nil
}

func (c *Channel) apiCall(ctx context.Context, method string, payload interface{}, result interface{}) error {
	apiURL := fmt.Sprintf("%s/%s", c.baseURL, method)

	var bodyReader io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bodyReader)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, result)
}
