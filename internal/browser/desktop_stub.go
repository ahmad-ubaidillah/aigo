//go:build !desktop

package browser

import "fmt"

func (c *Client) DesktopScreenshot(outputPath string) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopScreenshotRegion(x, y, w, h int, outputPath string) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopClick(x, y int, button string) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopTypeString(text string) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopKeyTap(key string, modifiers ...string) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopGetPixelColor(x, y int) string {
	return ""
}

func (c *Client) DesktopGetScreenSize() (int, int) {
	return 0, 0
}

func (c *Client) DesktopGetMouseLocation() (int, int) {
	return 0, 0
}

func (c *Client) DesktopMoveMouse(x, y int) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopFindImage(imagePath string) (int, int, error) {
	return -1, -1, fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopActiveWindow() string {
	return ""
}

func (c *Client) DesktopSetActiveWindow(title string) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopClipboardRead() (string, error) {
	return "", fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopClipboardWrite(text string) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopScroll(x, y int) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopOpenApp(appName string) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopProcessList() ([]string, error) {
	return nil, fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopKillProcess(pid int) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopWaitForColor(x, y int, color string, timeout int) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopKeyToggle(key string, down bool, modifiers ...string) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopDragSmooth(x, y int) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}

func (c *Client) DesktopToggle(button, action string) error {
	return fmt.Errorf("desktop support not available (build with -tags desktop)")
}
