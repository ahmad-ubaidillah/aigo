package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

const defaultCDPTimeout = 15 * time.Second

// Inspect connects to Lightpanda CDP, navigates to pageURL, and discovers
// interactive elements (inputs, buttons, links, selects, etc.).
func Inspect(cdpURL, pageURL string) (*InspectResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCDPTimeout)
	defer cancel()

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, cdpURL)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var title string
	var rawJSON string

	inspectJS := `
	(function() {
		var results = [];
		var selectors = 'input, textarea, select, button, a, [role="button"]';
		var els = document.querySelectorAll(selectors);
		for (var i = 0; i < els.length; i++) {
			var el = els[i];
			var info = {
				tag: el.tagName.toLowerCase(),
				type: el.getAttribute('type') || '',
				name: el.getAttribute('name') || '',
				placeholder: el.getAttribute('placeholder') || '',
				text: (el.innerText || el.textContent || '').trim().substring(0, 200),
				href: el.getAttribute('href') || '',
				role: el.getAttribute('role') || '',
				id: el.id || ''
			};
			// Build a selector
			if (el.id) {
				info.selector = '#' + el.id;
			} else if (el.name) {
				info.selector = el.tagName.toLowerCase() + '[name="' + el.name + '"]';
			} else {
				info.selector = el.tagName.toLowerCase();
			}
			results.push(info);
		}
		return JSON.stringify(results);
	})()
	`

	tasks := chromedp.Tasks{
		chromedp.Navigate(pageURL),
		chromedp.WaitReady("body"),
		chromedp.Title(&title),
		chromedp.Evaluate(inspectJS, &rawJSON),
	}

	if err := chromedp.Run(browserCtx, tasks); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout connecting to Lightpanda at %s", cdpURL)
		}
		return nil, fmt.Errorf("inspect error: %w", err)
	}

	var elements []ElementInfo
	if err := json.Unmarshal([]byte(rawJSON), &elements); err != nil {
		return nil, fmt.Errorf("parse inspect results: %w", err)
	}

	return &InspectResult{
		URL:      pageURL,
		Title:    title,
		Elements: elements,
	}, nil
}
