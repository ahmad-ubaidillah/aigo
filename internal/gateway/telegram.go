package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TelegramPlatform struct {
	token        string
	allowedChats map[int64]bool
	handler      func(Message)
	offset       int64
	client       *http.Client
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

type TelegramUpdate struct {
	UpdateID int64 `json:"update_id"`
	Message  struct {
		MessageID int64 `json:"message_id"`
		From      struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"chat"`
		Date    int64  `json:"date"`
		Text    string `json:"text"`
		Caption string `json:"caption"`
		Photo   []struct {
			FileID   string `json:"file_id"`
			FileSize int    `json:"file_size"`
		} `json:"photo"`
		Document *struct {
			FileID   string `json:"file_id"`
			FileName string `json:"file_name"`
			FileSize int    `json:"file_size"`
		} `json:"document"`
	} `json:"message"`
}

type TelegramResponse struct {
	OK     bool             `json:"ok"`
	Result []TelegramUpdate `json:"result"`
}

func NewTelegram(token string, allowedChats []int64) *TelegramPlatform {
	allowed := make(map[int64]bool)
	for _, chatID := range allowedChats {
		allowed[chatID] = true
	}

	return &TelegramPlatform{
		token:        token,
		allowedChats: allowed,
		client:       &http.Client{Timeout: 60 * time.Second},
		stopCh:       make(chan struct{}),
		offset:       0,
	}
}

func (t *TelegramPlatform) Name() string {
	return "telegram"
}

func (t *TelegramPlatform) Connect(ctx context.Context) error {
	if t.token == "" {
		return fmt.Errorf("telegram token is empty")
	}

	// Validate token by calling getMe
	resp, err := t.apiCall(ctx, "getMe", nil)
	if err != nil {
		return fmt.Errorf("telegram getMe failed: %w", err)
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			ID       int64  `json:"id"`
			IsBot    bool   `json:"is_bot"`
			Username string `json:"username"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return fmt.Errorf("parse getMe response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("telegram token invalid")
	}

	return nil
}

func (t *TelegramPlatform) Disconnect(ctx context.Context) error {
	close(t.stopCh)
	t.wg.Wait()
	return nil
}

func (t *TelegramPlatform) SendMessage(ctx context.Context, chatID, text string) error {
	params := url.Values{}
	params.Set("chat_id", chatID)
	params.Set("text", text)
	params.Set("parse_mode", "Markdown")

	_, err := t.apiCall(ctx, "sendMessage", params)
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}

	return nil
}

func (t *TelegramPlatform) SendTyping(ctx context.Context, chatID string) error {
	params := url.Values{}
	params.Set("chat_id", chatID)
	params.Set("action", "typing")

	_, err := t.apiCall(ctx, "sendChatAction", params)
	return err
}

func (t *TelegramPlatform) Listen(ctx context.Context, handler func(Message)) error {
	t.handler = handler
	t.wg.Add(1)
	defer t.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopCh:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			updates, err := t.getUpdates(ctx)
			if err != nil {
				continue
			}

			for _, update := range updates {
				t.offset = update.UpdateID + 1
				t.processUpdate(update)
			}
		}
	}
}

func (t *TelegramPlatform) getUpdates(ctx context.Context) ([]TelegramUpdate, error) {
	params := url.Values{}
	params.Set("offset", strconv.FormatInt(t.offset, 10))
	params.Set("timeout", "30")
	params.Set("allowed_updates", `["message"]`)

	resp, err := t.apiCall(ctx, "getUpdates", params)
	if err != nil {
		return nil, err
	}

	var result TelegramResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("parse updates: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("telegram getUpdates failed")
	}

	return result.Result, nil
}

func (t *TelegramPlatform) processUpdate(update TelegramUpdate) {
	if update.Message.Chat.ID == 0 {
		return
	}

	// Check if chat is allowed
	if len(t.allowedChats) > 0 && !t.allowedChats[update.Message.Chat.ID] {
		return
	}

	// Build user name
	userName := update.Message.From.FirstName
	if update.Message.From.LastName != "" {
		userName += " " + update.Message.From.LastName
	}
	if update.Message.From.Username != "" {
		userName += " (@" + update.Message.From.Username + ")"
	}

	// Build text (handle caption for media)
	text := update.Message.Text
	if text == "" {
		text = update.Message.Caption
	}

	// Build media URL if present
	mediaURL := ""
	if len(update.Message.Photo) > 0 {
		mediaURL = fmt.Sprintf("photo:%s", update.Message.Photo[len(update.Message.Photo)-1].FileID)
	}
	if update.Message.Document != nil {
		mediaURL = fmt.Sprintf("document:%s", update.Message.Document.FileID)
	}

	msg := Message{
		Platform:  "telegram",
		ChatID:    strconv.FormatInt(update.Message.Chat.ID, 10),
		UserID:    strconv.FormatInt(update.Message.From.ID, 10),
		UserName:  userName,
		Text:      text,
		MediaURL:  mediaURL,
		Timestamp: time.Unix(update.Message.Date, 0),
	}

	if t.handler != nil {
		t.handler(msg)
	}
}

func (t *TelegramPlatform) apiCall(ctx context.Context, method string, params url.Values) ([]byte, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/%s", t.token, method)

	var req *http.Request
	var err error

	if params != nil {
		req, err = http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return nil, err
		}
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return body, nil
}

// IsChatAllowed checks if a chat ID is in the allowed list
func (t *TelegramPlatform) IsChatAllowed(chatID int64) bool {
	if len(t.allowedChats) == 0 {
		return true // Allow all if no filter
	}
	return t.allowedChats[chatID]
}
