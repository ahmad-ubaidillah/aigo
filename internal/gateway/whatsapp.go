package gateway

import (
	"context"
	"fmt"
)

type WhatsAppPlatform struct {
	sessionID string
}

func NewWhatsApp(sessionID string) *WhatsAppPlatform {
	return &WhatsAppPlatform{
		sessionID: sessionID,
	}
}

func (w *WhatsAppPlatform) Name() string {
	return "whatsapp"
}

func (w *WhatsAppPlatform) Connect(ctx context.Context) error {
	if w.sessionID == "" {
		return fmt.Errorf("whatsapp session ID is empty")
	}

	return nil
}

func (w *WhatsAppPlatform) Disconnect(ctx context.Context) error {
	return nil
}

func (w *WhatsAppPlatform) SendMessage(ctx context.Context, phone, text string) error {
	return nil
}

func (w *WhatsAppPlatform) Listen(ctx context.Context, handler func(Message)) error {
	return nil
}
