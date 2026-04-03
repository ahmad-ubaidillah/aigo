package browser

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Client struct {
	browser *rod.Browser
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Launch(ctx context.Context) error {
	path, ok := launcher.LookPath()
	if !ok {
		return fmt.Errorf("browser not found")
	}
	url := launcher.New().Bin(path).Headless(true).NoSandbox(true).MustLaunch()
	c.browser = rod.New().ControlURL(url).MustConnect()
	return nil
}

func (c *Client) Close() error {
	if c.browser != nil {
		c.browser.MustClose()
	}
	return nil
}

func (c *Client) WebScreenshot(ctx context.Context, url, outputPath string) error {
	if c.browser == nil {
		return fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage(url)
	defer page.MustClose()
	data := page.MustScreenshot()
	return os.WriteFile(outputPath, data, 0644)
}

func (c *Client) WebGetText(ctx context.Context, url, selector string) (string, error) {
	if c.browser == nil {
		return "", fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage(url)
	defer page.MustClose()
	el, err := page.Element(selector)
	if err != nil {
		return "", fmt.Errorf("find element %s: %w", selector, err)
	}
	return el.Text()
}

func (c *Client) WebClick(ctx context.Context, url, selector string) error {
	if c.browser == nil {
		return fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage(url)
	defer page.MustClose()
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("find element %s: %w", selector, err)
	}
	return el.Click(proto.InputMouseButtonLeft, 1)
}

func (c *Client) WebFill(ctx context.Context, url, selector, text string) error {
	if c.browser == nil {
		return fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage(url)
	defer page.MustClose()
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("find element %s: %w", selector, err)
	}
	return el.Input(text)
}

func (c *Client) WebGetHTML(ctx context.Context, url string) (string, error) {
	if c.browser == nil {
		return "", fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage(url)
	defer page.MustClose()
	return page.HTML()
}

func (c *Client) WebSearch(ctx context.Context, query string, maxResults int) ([]string, error) {
	if c.browser == nil {
		return nil, fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage("https://www.google.com/search?q=" + query)
	defer page.MustClose()
	elements, err := page.Elements("h3")
	if err != nil {
		return nil, fmt.Errorf("find results: %w", err)
	}
	var titles []string
	for i, el := range elements {
		if i >= maxResults {
			break
		}
		t, _ := el.Text()
		if t != "" {
			titles = append(titles, t)
		}
	}
	return titles, nil
}

func (c *Client) WebNavigate(ctx context.Context, url string) error {
	if c.browser == nil {
		return fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage(url)
	defer page.MustClose()
	return page.WaitLoad()
}

func (c *Client) WebGetTitle(ctx context.Context) (string, error) {
	if c.browser == nil {
		return "", fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage("")
	defer page.MustClose()
	res, err := page.Eval("document.title")
	if err != nil {
		return "", err
	}
	return res.Value.String(), nil
}

func (c *Client) WebType(ctx context.Context, url, selector, text string) error {
	if c.browser == nil {
		return fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage(url)
	defer page.MustClose()
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("find element %s: %w", selector, err)
	}
	return el.Input(text)
}

func (c *Client) WebSelect(ctx context.Context, url, selector, optionText string) error {
	if c.browser == nil {
		return fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage(url)
	defer page.MustClose()
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("find element %s: %w", selector, err)
	}
	return el.Select([]string{optionText}, true, rod.SelectorTypeText)
}

func (c *Client) WebWaitForElement(ctx context.Context, url, selector string, timeoutSec int) error {
	if c.browser == nil {
		return fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage(url)
	defer page.MustClose()
	_, err := page.Timeout(time.Duration(timeoutSec) * time.Second).Element(selector)
	return err
}

func (c *Client) WebEvaluate(ctx context.Context, url, js string) (string, error) {
	if c.browser == nil {
		return "", fmt.Errorf("browser not launched")
	}
	page := c.browser.MustPage(url)
	defer page.MustClose()
	res, err := page.Eval(js)
	if err != nil {
		return "", err
	}
	return res.Value.String(), nil
}
