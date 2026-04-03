package orchestration

import (
	"sync"
	"time"
)

// Message represents a message sent between agents.
type Message struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`      // sender agent ID
	To        string    `json:"to"`        // receiver agent ID (empty for broadcast)
	Type      string    `json:"type"`      // message type (e.g., "task", "result", "error", "sync")
	Payload   any       `json:"payload"`   // message content
	Timestamp time.Time `json:"timestamp"` // when the message was created
	Priority  int       `json:"priority"`  // higher priority messages are processed first
}

// MessageBus handles routing messages between agents.
type MessageBus struct {
	mu         sync.RWMutex
	handlers   map[string]func(Message) // agentID -> handler
	queue      []Message
	maxQueue   int
	bufferSize int
}

// NewMessageBus creates a new MessageBus for inter-agent communication.
func NewMessageBus(bufferSize int) *MessageBus {
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	return &MessageBus{
		handlers:   make(map[string]func(Message)),
		queue:      make([]Message, 0),
		maxQueue:   bufferSize,
		bufferSize: bufferSize,
	}
}

// Send delivers a message to a specific agent.
func (mb *MessageBus) Send(msg Message) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	// Set timestamp if not set
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Generate ID if not set
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}

	handler, exists := mb.handlers[msg.To]
	if !exists {
		// Agent not subscribed, queue the message
		if len(mb.queue) >= mb.maxQueue {
			// Remove oldest message
			mb.queue = mb.queue[1:]
		}
		mb.queue = append(mb.queue, msg)
		return nil
	}

	// Deliver directly
	go handler(msg)
	return nil
}

// Broadcast sends a message to all subscribed agents.
func (mb *MessageBus) Broadcast(msg Message) error {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	// Set timestamp if not set
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Generate ID if not set
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}

	// Deliver to all handlers
	for _, handler := range mb.handlers {
		go handler(msg)
	}
	return nil
}

// Subscribe registers a handler for an agent to receive messages.
func (mb *MessageBus) Subscribe(agentID string, handler func(Message)) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.handlers[agentID] = handler
}

// Unsubscribe removes an agent's message handler.
func (mb *MessageBus) Unsubscribe(agentID string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	delete(mb.handlers, agentID)
}

// GetPendingMessages returns all queued messages for a specific agent.
func (mb *MessageBus) GetPendingMessages(agentID string) []Message {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	var result []Message
	var remaining []Message
	for _, msg := range mb.queue {
		if msg.To == agentID {
			result = append(result, msg)
		} else {
			remaining = append(remaining, msg)
		}
	}
	mb.queue = remaining
	return result
}

// Clear removes all queued messages.
func (mb *MessageBus) Clear() {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.queue = make([]Message, 0)
}

// SubscribedAgents returns a list of all subscribed agent IDs.
func (mb *MessageBus) SubscribedAgents() []string {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	agents := make([]string, 0, len(mb.handlers))
	for id := range mb.handlers {
		agents = append(agents, id)
	}
	return agents
}

// generateMessageID creates a unique message identifier.
func generateMessageID() string {
	return time.Now().Format("20060102150405.999999999")
}
