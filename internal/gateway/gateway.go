// Package gateway handles multi-channel message routing.
// Inspired by NekoClaw's service architecture and KrillClaw's channel pattern.
package gateway

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/hermes-v2/aigo/internal/agent"
)

// Channel defines the interface all messaging channels must implement.
type Channel interface {
	Name() string
	Start(ctx context.Context, onMessage func(Message) error) error
	Send(chatID string, text string) error
	Stop() error
}

// Message is a normalized message from any channel.
type Message struct {
	Channel  string // "telegram", "discord", "websocket"
	ChatID   string
	SenderID string
	Text     string
}

// Gateway routes messages between channels and the agent.
type Gateway struct {
	agent    *agent.Agent
	channels map[string]Channel
	mu       sync.RWMutex
}

// New creates a new gateway.
func New(a *agent.Agent) *Gateway {
	return &Gateway{
		agent:    a,
		channels: make(map[string]Channel),
	}
}

// Register adds a channel to the gateway.
func (g *Gateway) Register(ch Channel) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.channels[ch.Name()] = ch
	log.Printf("📡 Channel registered: %s", ch.Name())
}

// StartAll starts all registered channels.
func (g *Gateway) StartAll(ctx context.Context) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.channels) == 0 {
		log.Println("⚠️  No channels registered. Use 'aigo chat' for CLI mode.")
		return nil
	}

	errCh := make(chan error, len(g.channels))
	for name, ch := range g.channels {
		go func(name string, ch Channel) {
			log.Printf("🚀 Starting channel: %s", name)
			if err := ch.Start(ctx, g.handleMessage); err != nil {
				errCh <- fmt.Errorf("channel %s: %w", name, err)
			}
		}(name, ch)
	}

	// Wait for first error or context cancellation
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		g.StopAll()
		return ctx.Err()
	}
}

// handleMessage processes an incoming message from any channel.
func (g *Gateway) handleMessage(msg Message) error {
	log.Printf("📩 [%s] %s: %s", msg.Channel, msg.SenderID, truncate(msg.Text, 80))

	ctx := context.Background()
	result, err := g.agent.Run(ctx, msg.Text)
	if err != nil {
		return g.sendResponse(msg, fmt.Sprintf("Error: %v", err))
	}

	log.Printf("✅ [%s] responded in %d steps, %d tokens",
		msg.Channel, result.Steps, result.Usage.TotalTokens)

	return g.sendResponse(msg, result.Response)
}

// sendResponse routes a response back through the originating channel.
func (g *Gateway) sendResponse(msg Message, text string) error {
	g.mu.RLock()
	ch, ok := g.channels[msg.Channel]
	g.mu.RUnlock()

	if !ok {
		return fmt.Errorf("channel not found: %s", msg.Channel)
	}
	return ch.Send(msg.ChatID, text)
}

// StopAll stops all channels.
func (g *Gateway) StopAll() {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for name, ch := range g.channels {
		if err := ch.Stop(); err != nil {
			log.Printf("Error stopping channel %s: %v", name, err)
		}
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
