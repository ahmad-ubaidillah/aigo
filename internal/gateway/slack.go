package gateway

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type SlackPlatform struct {
	token string
}

func NewSlack(token string) *SlackPlatform {
	return &SlackPlatform{
		token: token,
	}
}

func (s *SlackPlatform) Name() string {
	return "slack"
}

func (s *SlackPlatform) Connect(ctx context.Context) error {
	if s.token == "" {
		return fmt.Errorf("slack token is empty")
	}

	if !strings.HasPrefix(s.token, "xoxb-") && !strings.HasPrefix(s.token, "xapp-") {
		return fmt.Errorf("slack token must start with xoxb- or xapp-")
	}

	return nil
}

func (s *SlackPlatform) Disconnect(ctx context.Context) error {
	return nil
}

func (s *SlackPlatform) SendMessage(ctx context.Context, channelID, text string) error {
	url := "https://slack.com/api/chat.postMessage"
	body := fmt.Sprintf(`{"channel":"%s","text":"%s"}`, channelID, text)
	cmd := exec.CommandContext(ctx, "curl", "-s", "-X", "POST", url,
		"-H", "Authorization: Bearer "+s.token,
		"-H", "Content-Type: application/json",
		"-d", body,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("slack send: %w: %s", err, string(output))
	}

	if strings.Contains(string(output), `"ok":false`) {
		return fmt.Errorf("slack API error: %s", string(output))
	}

	return nil
}

func (s *SlackPlatform) Listen(ctx context.Context, handler func(Message)) error {
	return nil
}
