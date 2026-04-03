package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ahmad-ubaidillah/aigo/internal/browser"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type BrowserHandler struct {
	browser *browser.Client
}

func NewBrowserHandler() *BrowserHandler {
	return &BrowserHandler{
		browser: browser.NewClient(),
	}
}

func (h *BrowserHandler) CanHandle(intent string) bool {
	return intent == types.IntentBrowser
}

type BrowserAction struct {
	Action   string `json:"action"`
	URL      string `json:"url,omitempty"`
	Selector string `json:"selector,omitempty"`
	Value    string `json:"value,omitempty"`
	Script   string `json:"script,omitempty"`
	Query    string `json:"query,omitempty"`
	MaxRes   int    `json:"max_results,omitempty"`
}

func (h *BrowserHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	var action BrowserAction
	if err := json.Unmarshal([]byte(task.Description), &action); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("parse action: %v", err),
		}, nil
	}

	if err := h.browser.Launch(ctx); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("launch browser: %v", err),
		}, nil
	}
	defer h.browser.Close()

	result, err := h.executeAction(ctx, action)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("execute: %v", err),
		}, nil
	}

	return &types.ToolResult{
		Success: true,
		Output:  result,
	}, nil
}

func (h *BrowserHandler) executeAction(ctx context.Context, action BrowserAction) (string, error) {
	switch action.Action {
	case "navigate":
		if action.URL == "" {
			return "", fmt.Errorf("URL required for navigate")
		}
		if err := h.browser.WebNavigate(ctx, action.URL); err != nil {
			return "", err
		}
		return fmt.Sprintf("Navigated to %s", action.URL), nil

	case "screenshot":
		if action.URL == "" {
			return "", fmt.Errorf("URL required for screenshot")
		}
		if err := h.browser.WebScreenshot(ctx, action.URL, "/tmp/screenshot.png"); err != nil {
			return "", err
		}
		return "Screenshot captured", nil

	case "click":
		if action.URL == "" || action.Selector == "" {
			return "", fmt.Errorf("URL and selector required for click")
		}
		if err := h.browser.WebClick(ctx, action.URL, action.Selector); err != nil {
			return "", err
		}
		return fmt.Sprintf("Clicked %s", action.Selector), nil

	case "fill":
		if action.URL == "" || action.Selector == "" || action.Value == "" {
			return "", fmt.Errorf("URL, selector, and value required for fill")
		}
		if err := h.browser.WebFill(ctx, action.URL, action.Selector, action.Value); err != nil {
			return "", err
		}
		return fmt.Sprintf("Filled %s", action.Selector), nil

	case "get_text":
		if action.URL == "" || action.Selector == "" {
			return "", fmt.Errorf("URL and selector required for get_text")
		}
		text, err := h.browser.WebGetText(ctx, action.URL, action.Selector)
		if err != nil {
			return "", err
		}
		return text, nil

	case "get_html":
		if action.URL == "" {
			return "", fmt.Errorf("URL required for get_html")
		}
		html, err := h.browser.WebGetHTML(ctx, action.URL)
		if err != nil {
			return "", err
		}
		return html, nil

	case "search":
		if action.Query == "" {
			return "", fmt.Errorf("query required for search")
		}
		maxRes := action.MaxRes
		if maxRes == 0 {
			maxRes = 5
		}
		titles, err := h.browser.WebSearch(ctx, action.Query, maxRes)
		if err != nil {
			return "", err
		}
		data, _ := json.Marshal(titles)
		return string(data), nil

	case "evaluate":
		if action.URL == "" || action.Script == "" {
			return "", fmt.Errorf("URL and script required for evaluate")
		}
		result, err := h.browser.WebEvaluate(ctx, action.URL, action.Script)
		if err != nil {
			return "", err
		}
		return result, nil

	case "desktop_screenshot":
		if err := h.browser.DesktopScreenshot("/tmp/desktop_screenshot.png"); err != nil {
			return "", err
		}
		return "Desktop screenshot captured", nil

	case "desktop_click":
		x := 0
		y := 0
		if action.Selector != "" {
			fmt.Sscanf(action.Selector, "%d,%d", &x, &y)
		}
		if err := h.browser.DesktopClick(x, y, "left"); err != nil {
			return "", err
		}
		return fmt.Sprintf("Clicked at %d,%d", x, y), nil

	case "desktop_type":
		if action.Value == "" {
			return "", fmt.Errorf("value required for desktop_type")
		}
		if err := h.browser.DesktopTypeString(action.Value); err != nil {
			return "", err
		}
		return fmt.Sprintf("Typed: %s", action.Value), nil

	case "desktop_clipboard_read":
		text, err := h.browser.DesktopClipboardRead()
		if err != nil {
			return "", err
		}
		return text, nil

	case "desktop_clipboard_write":
		if action.Value == "" {
			return "", fmt.Errorf("value required for desktop_clipboard_write")
		}
		if err := h.browser.DesktopClipboardWrite(action.Value); err != nil {
			return "", err
		}
		return "Clipboard updated", nil

	default:
		return "", fmt.Errorf("unknown action: %s", action.Action)
	}
}

type BatchBrowserAction struct {
	Actions []BrowserAction `json:"actions"`
}

func (h *BrowserHandler) ExecuteBatch(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	var batch BatchBrowserAction
	if err := json.Unmarshal([]byte(task.Description), &batch); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("parse batch: %v", err),
		}, nil
	}

	if err := h.browser.Launch(ctx); err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}
	defer h.browser.Close()

	results := make([]string, len(batch.Actions))
	for i, action := range batch.Actions {
		result, err := h.executeAction(ctx, action)
		if err != nil {
			results[i] = fmt.Sprintf("ERROR: %v", err)
		} else {
			results[i] = result
		}
	}

	output, _ := json.Marshal(results)
	return &types.ToolResult{
		Success: true,
		Output:  string(output),
	}, nil
}
