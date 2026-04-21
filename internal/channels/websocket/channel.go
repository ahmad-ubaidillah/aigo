// Package websocket implements WebSocket channel for Aigo.
// Wire protocol: client sends {"type":"message","text":"..."}
//                server sends {"type":"text","text":"..."} then {"type":"done"}
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/hermes-v2/aigo/internal/gateway"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Channel implements the gateway.Channel interface for WebSocket.
type Channel struct {
	port        int
	authToken   string
	connections map[*websocket.Conn]string // ws → chatID
	mu          sync.RWMutex
	onMessage   func(gateway.Message) error
	server      *http.Server
}

// New creates a new WebSocket channel.
func New(port int, authToken string) *Channel {
	return &Channel{
		port:        port,
		authToken:   authToken,
		connections: make(map[*websocket.Conn]string),
	}
}

func (c *Channel) Name() string { return "websocket" }

func (c *Channel) Start(ctx context.Context, onMessage func(gateway.Message) error) error {
	c.onMessage = onMessage

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", c.handleWS)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		connCount := len(c.connections)
		c.mu.RUnlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "ok",
			"platform":    "websocket",
			"connections": connCount,
		})
	})

	c.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.port),
		Handler: mux,
	}

	log.Printf("WebSocket server starting on :%d", c.port)

	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	return c.Stop()
}

func (c *Channel) handleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	chatID := fmt.Sprintf("ws:%d", len(c.connections)+1)
	c.mu.Lock()
	c.connections[ws] = chatID
	c.mu.Unlock()

	log.Printf("WebSocket client connected: %s", chatID)

	defer func() {
		c.mu.Lock()
		delete(c.connections, ws)
		c.mu.Unlock()
		log.Printf("WebSocket client disconnected: %s", chatID)
	}()

	for {
		_, msgBytes, err := ws.ReadMessage()
		if err != nil {
			break
		}

		var msg struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			ws.WriteJSON(map[string]string{"type": "error", "text": "Invalid JSON"})
			continue
		}

		switch msg.Type {
		case "message":
			if msg.Text == "" {
				ws.WriteJSON(map[string]string{"type": "error", "text": "Empty message"})
				continue
			}
			gwMsg := gateway.Message{
				Channel:  "websocket",
				ChatID:   chatID,
				SenderID: chatID,
				Text:     msg.Text,
			}
			if c.onMessage != nil {
				if err := c.onMessage(gwMsg); err != nil {
					ws.WriteJSON(map[string]string{"type": "error", "text": err.Error()})
				}
			}
			ws.WriteJSON(map[string]string{"type": "done"})

		case "ping":
			ws.WriteJSON(map[string]string{"type": "pong"})
		}
	}
}

func (c *Channel) Send(chatID string, text string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for ws, id := range c.connections {
		if id == chatID {
			// Split long messages
			if len(text) > 65536 {
				for i := 0; i < len(text); i += 65536 {
					end := i + 65536
					if end > len(text) {
						end = len(text)
					}
					if err := ws.WriteJSON(map[string]string{"type": "text", "text": text[i:end]}); err != nil {
						return err
					}
				}
				return nil
			}
			return ws.WriteJSON(map[string]string{"type": "text", "text": text})
		}
	}
	return fmt.Errorf("connection not found for chat_id: %s", chatID)
}

func (c *Channel) Stop() error {
	if c.server != nil {
		return c.server.Close()
	}
	return nil
}
