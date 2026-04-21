package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hermes-v2/aigo/internal/gateway"
)

type Channel struct {
	accountSid string
	authToken  string
	fromNumber string
	onMessage  func(gateway.Message) error
	client     *http.Client
	running    bool
	mu         sync.Mutex

	connected  bool
	lastUpdate time.Time
	errorCount int
}

func New(accountSid, authToken, fromNumber string) *Channel {
	return &Channel{
		accountSid: accountSid,
		authToken:  authToken,
		fromNumber: fromNumber,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Channel) Name() string { return "whatsapp" }

func (c *Channel) Start(ctx context.Context, onMessage func(gateway.Message) error) error {
	c.onMessage = onMessage
	c.connected = true
	c.running = true
	log.Println("📡 WhatsApp (Twilio) channel connected")
	return nil
}

func (c *Channel) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.running = false
	c.connected = false
	log.Println("📡 WhatsApp channel stopped")
	return nil
}

func (c *Channel) Send(to, text string) error {
	if !c.connected {
		return fmt.Errorf("whatsapp: not connected")
	}

	form := url.Values{}
	form.Set("To", "whatsapp:"+to)
	form.Set("From", "whatsapp:"+c.fromNumber)
	form.Set("Body", text)

	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", c.accountSid)
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.accountSid, c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twilio error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Channel) SendMedia(to, text, mediaURL string) error {
	if !c.connected {
		return fmt.Errorf("whatsapp: not connected")
	}

	form := url.Values{}
	form.Set("To", "whatsapp:"+to)
	form.Set("From", "whatsapp:"+c.fromNumber)
	form.Set("Body", text)
	form.Set("MediaUrl", mediaURL)

	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", c.accountSid)
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.accountSid, c.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twilio error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Channel) HandleWebhook(body []byte) error {
	if c.onMessage == nil {
		return nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("parse webhook: %w", err)
	}

	bodyField, ok := payload["Body"].(string)
	fromField, ok := payload["From"].(string)
	if !ok || bodyField == "" || fromField == "" {
		return nil
	}

	fromNumber := strings.TrimPrefix(fromField, "whatsapp:")

	msg := gateway.Message{
		Channel:  "whatsapp",
		ChatID:   fromNumber,
		SenderID: fromNumber,
		Text:     bodyField,
	}

	return c.onMessage(msg)
}

type Config struct {
	AccountSid string `yaml:"account_sid"`
	AuthToken  string `yaml:"auth_token"`
	FromNumber string `yaml:"from_number"`
	Enabled    bool   `yaml:"enabled"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg struct {
		Channels struct {
			WhatsApp Config `yaml:"whatsapp"`
		} `yaml:"channels"`
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg.Channels.WhatsApp, nil
}

func NewFromConfig(cfg Config) *Channel {
	return New(cfg.AccountSid, cfg.AuthToken, cfg.FromNumber)
}