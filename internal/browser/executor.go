package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

// Executor runs workflow actions against a Lightpanda CDP server.
type Executor struct {
	cdpURL string
}

// NewExecutor creates an executor targeting the given CDP websocket URL.
func NewExecutor(cdpURL string) *Executor {
	return &Executor{cdpURL: cdpURL}
}

// Run executes the named action from the workflow.
func (e *Executor) Run(wf *Workflow, actionName string) (*ActionResult, error) {
	action, ok := wf.Actions[actionName]
	if !ok {
		return nil, fmt.Errorf("action %q not found in workflow", actionName)
	}

	result := &ActionResult{Action: actionName}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, e.cdpURL)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// Navigate to the action's starting URL if specified
	if action.URL != "" {
		if err := chromedp.Run(browserCtx, chromedp.Navigate(action.URL), chromedp.WaitReady("body")); err != nil {
			result.Error = fmt.Sprintf("navigate to %s: %v", action.URL, err)
			return result, nil
		}
	}

	// Execute each step
	for i, step := range action.Steps {
		if err := e.executeStep(browserCtx, wf, &step); err != nil {
			result.Error = fmt.Sprintf("step %d failed: %v", i+1, err)
			result.StepsRun = i
			return result, nil
		}
	}

	result.Success = true
	result.StepsRun = len(action.Steps)
	return result, nil
}

func (e *Executor) executeStep(ctx context.Context, wf *Workflow, step *Step) error {
	switch {
	case step.Navigate != "":
		return e.doNavigate(ctx, step.Navigate)
	case step.Click != nil:
		return e.doClick(ctx, step.Click)
	case step.Fill != nil:
		return e.doFill(ctx, wf, step.Fill)
	case step.Wait != nil:
		return e.doWait(ctx, step.Wait)
	case step.Eval != "":
		return e.doEval(ctx, step.Eval)
	case step.Screenshot != "":
		return e.doScreenshot(ctx, step.Screenshot)
	default:
		return fmt.Errorf("empty step")
	}
}

func (e *Executor) doNavigate(ctx context.Context, url string) error {
	return chromedp.Run(ctx, chromedp.Navigate(url), chromedp.WaitReady("body"))
}

func (e *Executor) doClick(ctx context.Context, cs *ClickStep) error {
	return chromedp.Run(ctx,
		chromedp.WaitVisible(cs.Selector, chromedp.ByQuery),
		chromedp.Click(cs.Selector, chromedp.ByQuery),
	)
}

func (e *Executor) doFill(ctx context.Context, wf *Workflow, fs *FillStep) error {
	val := fs.Value
	if wf != nil {
		val = interpolateString(val, wf.Env)
	}
	return chromedp.Run(ctx,
		chromedp.WaitVisible(fs.Selector, chromedp.ByQuery),
		chromedp.SetValue(fs.Selector, val, chromedp.ByQuery),
	)
}

func (e *Executor) doWait(ctx context.Context, ws *WaitStep) error {
	if ws.Selector != "" {
		timeout := 10 * time.Second
		if ws.Timeout != "" {
			if d, err := time.ParseDuration(ws.Timeout); err == nil {
				timeout = d
			}
		}
		waitCtx, waitCancel := context.WithTimeout(ctx, timeout)
		defer waitCancel()
		return chromedp.Run(waitCtx, chromedp.WaitReady(ws.Selector, chromedp.ByQuery))
	}
	// No selector — just wait a beat
	if ws.Timeout != "" {
		if d, err := time.ParseDuration(ws.Timeout); err == nil {
			time.Sleep(d)
			return nil
		}
	}
	time.Sleep(1 * time.Second)
	return nil
}

func (e *Executor) doEval(ctx context.Context, js string) error {
	var result string
	return chromedp.Run(ctx, chromedp.Evaluate(js, &result))
}

func (e *Executor) doScreenshot(ctx context.Context, path string) error {
	var buf []byte
	if err := chromedp.Run(ctx, chromedp.FullScreenshot(&buf, 90)); err != nil {
		return fmt.Errorf("screenshot: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(path, buf, 0644); err != nil {
		return fmt.Errorf("write screenshot: %w", err)
	}
	return nil
}

// toJSON is a helper for returning JSON from eval steps.
func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
