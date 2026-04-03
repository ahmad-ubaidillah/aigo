package handlers

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/internal/gateway"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type GatewayHandler struct {
	manager *gateway.Manager
}

func NewGatewayHandler() *GatewayHandler {
	m := gateway.NewManager()

	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if telegramToken != "" {
		m.Register(gateway.NewTelegram(telegramToken, nil))
	}

	discordToken := os.Getenv("DISCORD_BOT_TOKEN")
	if discordToken != "" {
		m.Register(gateway.NewDiscord(discordToken))
	}

	slackToken := os.Getenv("SLACK_BOT_TOKEN")
	if slackToken != "" {
		m.Register(gateway.NewSlack(slackToken))
	}

	whatsappToken := os.Getenv("WHATSAPP_TOKEN")
	if whatsappToken != "" {
		m.Register(gateway.NewWhatsApp(whatsappToken))
	}

	return &GatewayHandler{manager: m}
}

func (h *GatewayHandler) CanHandle(intent string) bool {
	return intent == types.IntentGateway || intent == types.IntentMemory
}

func (h *GatewayHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	desc := strings.TrimSpace(task.Description)

	if strings.HasPrefix(desc, "send ") {
		parts := strings.SplitN(strings.TrimPrefix(desc, "send "), " ", 2)
		if len(parts) < 2 {
			return &types.ToolResult{
				Success: false,
				Error:   "usage: send <platform> <message>",
			}, nil
		}
		platform := parts[0]
		message := parts[1]
		err := h.manager.Send(platform, "", message)
		if err != nil {
			return &types.ToolResult{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
		return &types.ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Message sent via %s", platform),
		}, nil
	}

	if strings.HasPrefix(desc, "start ") {
		platform := strings.TrimPrefix(desc, "start ")
		err := h.manager.Start(ctx, platform)
		if err != nil {
			return &types.ToolResult{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
		return &types.ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Connected to %s", platform),
		}, nil
	}

	if strings.HasPrefix(desc, "stop ") {
		platform := strings.TrimPrefix(desc, "stop ")
		err := h.manager.Stop(platform)
		if err != nil {
			return &types.ToolResult{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
		return &types.ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Disconnected from %s", platform),
		}, nil
	}

	if desc == "status" || desc == "list" {
		status := h.manager.Status()
		var lines []string
		for platform, running := range status {
			statusIcon := "○"
			if running {
				statusIcon = "●"
			}
			lines = append(lines, fmt.Sprintf("%s %s", statusIcon, platform))
		}
		return &types.ToolResult{
			Success: true,
			Output:  "Gateway Status:\n" + strings.Join(lines, "\n"),
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output: `Gateway Commands:
  send <platform> <message>  - Send a message
  start <platform>           - Connect to platform
  stop <platform>            - Disconnect from platform
  status                      - Show connection status

Platforms: telegram, discord, slack, whatsapp
Set env vars: TELEGRAM_BOT_TOKEN, DISCORD_BOT_TOKEN, SLACK_BOT_TOKEN, WHATSAPP_TOKEN`,
	}, nil
}
