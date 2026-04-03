//go:build desktop

package browser

import (
	"fmt"
	"image/png"
	"os"
	"strings"

	"github.com/go-vgo/robotgo"
	"github.com/vcaesar/bitmap"
)

func (c *Client) DesktopScreenshot(outputPath string) error {
	img := robotgo.CaptureScreen()
	defer robotgo.FreeBitmap(img)
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, robotgo.ToImage(img))
}

func (c *Client) DesktopScreenshotRegion(x, y, w, h int, outputPath string) error {
	img := robotgo.CaptureScreen(x, y, w, h)
	defer robotgo.FreeBitmap(img)
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, robotgo.ToImage(img))
}

func (c *Client) DesktopClick(x, y int, button string) error {
	robotgo.Move(x, y)
	if button == "" {
		button = "left"
	}
	robotgo.Click(button)
	return nil
}

func (c *Client) DesktopTypeString(text string) error {
	robotgo.TypeStr(text)
	return nil
}

func (c *Client) DesktopKeyTap(key string, modifiers ...string) error {
	args := make([]interface{}, len(modifiers))
	for i, m := range modifiers {
		args[i] = m
	}
	robotgo.KeyTap(key, args...)
	return nil
}

func (c *Client) DesktopGetPixelColor(x, y int) string {
	return robotgo.GetPixelColor(x, y)
}

func (c *Client) DesktopGetScreenSize() (int, int) {
	w, h := robotgo.GetScreenSize()
	return w, h
}

func (c *Client) DesktopGetMouseLocation() (int, int) {
	x, y := robotgo.Location()
	return x, y
}

func (c *Client) DesktopMoveMouse(x, y int) error {
	robotgo.Move(x, y)
	return nil
}

func (c *Client) DesktopFindImage(imagePath string) (int, int, error) {
	ref := bitmap.Open(imagePath)
	defer robotgo.FreeBitmap(ref)
	x, y := bitmap.Find(ref)
	if x < 0 || y < 0 {
		return -1, -1, fmt.Errorf("image %s not found", imagePath)
	}
	return x, y, nil
}

func (c *Client) DesktopActiveWindow() string {
	return robotgo.GetTitle()
}

func (c *Client) DesktopSetActiveWindow(title string) error {
	return robotgo.ActiveName(title)
}

func (c *Client) DesktopClipboardRead() (string, error) {
	return robotgo.ReadAll()
}

func (c *Client) DesktopClipboardWrite(text string) error {
	return robotgo.WriteAll(text)
}

func (c *Client) DesktopScroll(x, y int) error {
	robotgo.Scroll(x, y)
	return nil
}

func (c *Client) DesktopOpenApp(appName string) error {
	pids, err := robotgo.FindIds(appName)
	if err != nil {
		return fmt.Errorf("find app %s: %w", appName, err)
	}
	if len(pids) > 0 {
		return robotgo.ActivePid(pids[0])
	}
	return fmt.Errorf("app %s not found", appName)
}

func (c *Client) DesktopProcessList() ([]string, error) {
	procs, err := robotgo.Process()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, p := range procs {
		names = append(names, p.Name)
	}
	return names, nil
}

func (c *Client) DesktopKillProcess(pid int) error {
	return robotgo.Kill(pid)
}

func (c *Client) DesktopWaitForColor(x, y int, color string, timeout int) error {
	for i := 0; i < timeout*10; i++ {
		c := robotgo.GetPixelColor(x, y)
		if strings.EqualFold(c, color) {
			return nil
		}
		robotgo.MilliSleep(100)
	}
	return fmt.Errorf("timeout waiting for color %s at (%d,%d)", color, x, y)
}

func (c *Client) DesktopKeyToggle(key string, down bool, modifiers ...string) error {
	action := "down"
	if !down {
		action = "up"
	}
	args := []interface{}{action}
	for _, m := range modifiers {
		args = append(args, m)
	}
	robotgo.KeyToggle(key, args...)
	return nil
}

func (c *Client) DesktopDragSmooth(x, y int) error {
	robotgo.DragSmooth(x, y)
	return nil
}

func (c *Client) DesktopToggle(button, action string) error {
	robotgo.Toggle(button, action)
	return nil
}
