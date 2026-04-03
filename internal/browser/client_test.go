//go:build desktop

package browser

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	t.Parallel()
	c := NewClient()
	if c == nil {
		t.Error("expected client")
	}
}

func TestClient_CloseNilBrowser(t *testing.T) {
	t.Parallel()
	c := NewClient()
	err := c.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_WebMethodsNilBrowser(t *testing.T) {
	t.Parallel()
	c := NewClient()
	ctx := context.Background()

	if _, err := c.WebGetText(ctx, "http://test", "#id"); err == nil {
		t.Error("expected error for nil browser")
	}
	if err := c.WebClick(ctx, "http://test", "#id"); err == nil {
		t.Error("expected error for nil browser")
	}
	if err := c.WebFill(ctx, "http://test", "#id", "text"); err == nil {
		t.Error("expected error for nil browser")
	}
	if _, err := c.WebGetHTML(ctx, "http://test"); err == nil {
		t.Error("expected error for nil browser")
	}
	if _, err := c.WebSearch(ctx, "query", 5); err == nil {
		t.Error("expected error for nil browser")
	}
	if err := c.WebNavigate(ctx, "http://test"); err == nil {
		t.Error("expected error for nil browser")
	}
	if _, err := c.WebGetTitle(ctx); err == nil {
		t.Error("expected error for nil browser")
	}
	if err := c.WebType(ctx, "http://test", "#id", "text"); err == nil {
		t.Error("expected error for nil browser")
	}
	if err := c.WebSelect(ctx, "http://test", "#id", "opt"); err == nil {
		t.Error("expected error for nil browser")
	}
	if err := c.WebWaitForElement(ctx, "http://test", "#id", 1); err == nil {
		t.Error("expected error for nil browser")
	}
	if _, err := c.WebEvaluate(ctx, "http://test", "1+1"); err == nil {
		t.Error("expected error for nil browser")
	}
}

func TestClient_DesktopGetScreenSize(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopGetMouseLocation(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopGetPixelColor(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopMoveMouse(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopTypeString(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopKeyTap(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopScroll(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopClipboard(t *testing.T) {
	t.Parallel()
	t.Skip("requires display and clipboard utilities")
}

func TestClient_DesktopActiveWindow(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopProcessList(t *testing.T) {
	t.Parallel()
	c := NewClient()
	procs, err := c.DesktopProcessList()
	if err != nil {
		t.Fatal(err)
	}
	if len(procs) == 0 {
		t.Error("expected at least one process")
	}
}

func TestClient_DesktopOpenAppNotFound(t *testing.T) {
	t.Parallel()
	c := NewClient()
	err := c.DesktopOpenApp("nonexistent_app_xyz_12345")
	if err == nil {
		t.Error("expected error for nonexistent app")
	}
}

func TestClient_DesktopFindImageNotFound(t *testing.T) {
	t.Parallel()
	c := NewClient()
	_, _, err := c.DesktopFindImage("/nonexistent/image.png")
	if err == nil {
		t.Error("expected error for nonexistent image")
	}
}

func TestClient_DesktopScreenshot(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopScreenshotRegion(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopWaitForColorTimeout(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopKeyToggle(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopDragSmooth(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}

func TestClient_DesktopToggle(t *testing.T) {
	t.Parallel()
	t.Skip("requires display")
}
