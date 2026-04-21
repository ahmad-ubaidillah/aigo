// Package slack implements Slack channel for Aigo.
// Uses slack-go with Socket Mode — no public URL or ngrok needed.
package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	"github.com/hermes-v2/aigo/internal/gateway"
)

// Channel implements the gateway.Channel interface for Slack.
type Channel struct {
	appToken  string // xapp-... (Socket Mode)
	botToken  string // xoxb-... (Bot)
	client    *slack.Client
	socket    *socketmode.Client
	onMessage func(gateway.Message) error
	userID    string
}

// New creates a new Slack channel.
func New(appToken, botToken string) *Channel {
	return &Channel{
		appToken: appToken,
		botToken: botToken,
	}
}

func (c *Channel) Name() string { return "slack" }

func (c *Channel) Start(ctx context.Context, onMessage func(gateway.Message) error) error {
	c.onMessage = onMessage

	c.client = slack.New(c.botToken, slack.OptionAppLevelToken(c.appToken))

	// Verify auth
	authResp, err := c.client.AuthTestContext(ctx)
	if err != nil {
		return fmt.Errorf("slack auth test: %w", err)
	}
	c.userID = authResp.UserID
	log.Printf("Slack bot: @%s (team: %s)", authResp.User, authResp.Team)

	// Socket Mode
	c.socket = socketmode.New(c.client, socketmode.OptionDebug(false))

	go c.eventLoop(ctx)

	go func() {
		if err := c.socket.RunContext(ctx); err != nil && ctx.Err() == nil {
			log.Printf("Slack socket error: %v", err)
		}
	}()

	log.Println("📡 Slack channel started (Socket Mode)")
	<-ctx.Done()
	return c.Stop()
}

func (c *Channel) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-c.socket.Events:
			c.handleEvent(evt)
		}
	}
}

func (c *Channel) handleEvent(evt socketmode.Event) {
	switch evt.Type {
	case socketmode.EventTypeEventsAPI:
		c.socket.Ack(*evt.Request)
		c.handleEventsAPI(evt)

	case socketmode.EventTypeInteractive:
		c.socket.Ack(*evt.Request)

	case socketmode.EventTypeSlashCommand:
		c.socket.Ack(*evt.Request)
	}
}

// eventsAPIPayload is the minimal structure we need from the Events API envelope.
type eventsAPIPayload struct {
	Type      string          `json:"type"`
	Event     json.RawMessage `json:"event"`
	EventID   string          `json:"event_id"`
	EventTime int64           `json:"event_time"`
}

// slackEvent is the inner event structure (common fields).
type slackEvent struct {
	Type        string `json:"type"`
	User        string `json:"user"`
	Text        string `json:"text"`
	Channel     string `json:"channel"`
	ChannelType string `json:"channel_type"`
	BotID       string `json:"bot_id"`
	SubType     string `json:"subtype"`
}

func (c *Channel) handleEventsAPI(evt socketmode.Event) {
	// The raw Events API envelope is available as JSON
	rawBytes, err := json.Marshal(evt.Data)
	if err != nil {
		return
	}

	var payload eventsAPIPayload
	if err := json.Unmarshal(rawBytes, &payload); err != nil {
		return
	}

	// Check if this is a callback event
	if payload.Type != "event_callback" && payload.Type != "" {
		// In Socket Mode, type is "events_api"
	}

	var inner slackEvent
	if err := json.Unmarshal(payload.Event, &inner); err != nil {
		return
	}

	switch inner.Type {
	case "message":
		c.processMessage(inner)
	case "app_mention":
		c.processMention(inner)
	}
}

func (c *Channel) processMessage(ev slackEvent) {
	if ev.BotID != "" || ev.SubType == "bot_message" {
		return
	}

	isDM := ev.ChannelType == "im"
	isMention := strings.Contains(ev.Text, fmt.Sprintf("<@%s>", c.userID))

	if !isDM && !isMention {
		return
	}

	text := ev.Text
	if isMention {
		text = strings.ReplaceAll(text, fmt.Sprintf("<@%s>", c.userID), "")
		text = strings.TrimSpace(text)
	}

	if text == "" {
		return
	}

	c.dispatch(gateway.Message{
		Channel:  "slack",
		ChatID:   ev.Channel,
		SenderID: ev.User,
		Text:     text,
	})
}

func (c *Channel) processMention(ev slackEvent) {
	if ev.BotID != "" {
		return
	}

	text := strings.ReplaceAll(ev.Text, fmt.Sprintf("<@%s>", c.userID), "")
	text = strings.TrimSpace(text)

	if text == "" {
		return
	}

	c.dispatch(gateway.Message{
		Channel:  "slack",
		ChatID:   ev.Channel,
		SenderID: ev.User,
		Text:     text,
	})
}

func (c *Channel) dispatch(msg gateway.Message) {
	if c.onMessage != nil {
		go func(m gateway.Message) {
			if err := c.onMessage(m); err != nil {
				log.Printf("Slack handler error: %v", err)
			}
		}(msg)
	}
}

func (c *Channel) Send(channelID string, text string) error {
	if c.client == nil {
		return fmt.Errorf("slack client not initialized")
	}

	if len(text) <= 4000 {
		_, _, err := c.client.PostMessage(channelID, slack.MsgOptionText(text, false))
		return err
	}

	for i := 0; i < len(text); i += 4000 {
		end := i + 4000
		if end > len(text) {
			end = len(text)
		}
		if _, _, err := c.client.PostMessage(channelID, slack.MsgOptionText(text[i:end], false)); err != nil {
			return err
		}
	}
	return nil
}

func (c *Channel) Stop() error {
	log.Println("Slack channel stopped")
	return nil
}
