package gateway

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type DiscordPlatform struct {
	token string
}

func NewDiscord(token string) *DiscordPlatform {
	return &DiscordPlatform{
		token: token,
	}
}

func (d *DiscordPlatform) Name() string {
	return "discord"
}

func (d *DiscordPlatform) Connect(ctx context.Context) error {
	if d.token == "" {
		return fmt.Errorf("discord token is empty")
	}

	if !strings.Contains(d.token, ".") {
		return fmt.Errorf("discord token must contain a dot")
	}

	return nil
}

func (d *DiscordPlatform) Disconnect(ctx context.Context) error {
	return nil
}

func (d *DiscordPlatform) SendMessage(ctx context.Context, channelID, text string) error {
	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "POST", url,
		"-H", "Authorization: Bot "+d.token,
		"-H", "Content-Type: application/json",
		"-d", fmt.Sprintf(`{"content":"%s"}`, text),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("discord send: %w: %s", err, string(output))
	}

	if strings.Contains(string(output), `"message"`) && strings.Contains(string(output), `"code"`) {
		return fmt.Errorf("discord API error: %s", string(output))
	}

	return nil
}

func (d *DiscordPlatform) Listen(ctx context.Context, handler func(Message)) error {
	return nil
}
