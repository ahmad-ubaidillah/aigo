// Package discord implements Discord channel for Aigo.
// Uses discordgo library for bot integration.
package discord

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/hermes-v2/aigo/internal/gateway"
)

// Channel implements the gateway.Channel interface for Discord.
type Channel struct {
	token     string
	session   *discordgo.Session
	onMessage func(gateway.Message) error
}

// New creates a new Discord channel.
func New(token string) *Channel {
	return &Channel{token: token}
}

func (c *Channel) Name() string { return "discord" }

func (c *Channel) Start(ctx context.Context, onMessage func(gateway.Message) error) error {
	c.onMessage = onMessage

	dg, err := discordgo.New("Bot " + c.token)
	if err != nil {
		return fmt.Errorf("create discord session: %w", err)
	}
	c.session = dg

	// Handler for messages
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore own messages
		if m.Author.ID == s.State.User.ID {
			return
		}

		// Only respond to mentions or DMs
		isDM := m.GuildID == ""
		isMention := false
		for _, mention := range m.Mentions {
			if mention.ID == s.State.User.ID {
				isMention = true
				break
			}
		}

		if !isDM && !isMention {
			return
		}

		// Strip mention from text
		text := m.Content
		if isMention {
			text = strings.ReplaceAll(text, fmt.Sprintf("<@%s>", s.State.User.ID), "")
			text = strings.TrimSpace(text)
		}

		if text == "" {
			return
		}

		gwMsg := gateway.Message{
			Channel:  "discord",
			ChatID:   m.ChannelID,
			SenderID: m.Author.ID,
			Text:     text,
		}

		if c.onMessage != nil {
			go func(msg gateway.Message) {
				if err := c.onMessage(msg); err != nil {
					log.Printf("Discord handler error: %v", err)
				}
			}(gwMsg)
		}
	})

	// Open connection
	if err := dg.Open(); err != nil {
		return fmt.Errorf("discord open: %w", err)
	}

	log.Printf("Discord bot started: %s", dg.State.User.Username)

	// Wait for context cancellation
	<-ctx.Done()
	return c.Stop()
}

func (c *Channel) Send(channelID string, text string) error {
	if c.session == nil {
		return fmt.Errorf("discord session not initialized")
	}

	// Discord message limit is 2000 chars
	if len(text) <= 2000 {
		_, err := c.session.ChannelMessageSend(channelID, text)
		return err
	}

	// Split long messages
	for i := 0; i < len(text); i += 2000 {
		end := i + 2000
		if end > len(text) {
			end = len(text)
		}
		if _, err := c.session.ChannelMessageSend(channelID, text[i:end]); err != nil {
			return err
		}
	}
	return nil
}

func (c *Channel) Stop() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}
