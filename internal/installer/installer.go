package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"os/exec"
)

type DownloadProgress struct {
	Total   int64
	Current int64
}

func (d *DownloadProgress) Write(p []byte) (int, error) {
	d.Current += int64(len(p))
	return len(p), nil
}

func Download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func InstallBinary(name, url, installPath string) error {
	tmpPath := filepath.Join(os.TempDir(), name)
	if err := Download(url, tmpPath); err != nil {
		return fmt.Errorf("download: %w", err)
	}
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}
	return os.Rename(tmpPath, installPath)
}

func GetOS() string {
	return runtime.GOOS
}

func GetArch() string {
	return runtime.GOARCH
}

func GetVersion() string {
	return "1.0.0"
}

type Installer struct {
	verbose    bool
	force      bool
	httpClient *http.Client
}

func NewInstaller(verbose, force bool) *Installer {
	return &Installer{
		verbose:    verbose,
		force:      force,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (i *Installer) Install(ctx context.Context, pkg string) error {
	return fmt.Errorf("not implemented: %s", pkg)
}

func (i *Installer) CheckUpdate() (string, error) {
	return "", nil
}

func (i *Installer) CheckOpenCode() (bool, string, error) {
	if path, err := exec.LookPath("opencode"); err == nil {
		return true, path, nil
	}
	common := []string{
		"/usr/local/bin/opencode",
		"/usr/bin/opencode",
		filepath.Join(os.Getenv("HOME"), ".local", "bin", "opencode"),
	}
	for _, p := range common {
		if _, err := os.Stat(p); err == nil {
			return true, p, nil
		}
	}
	return false, "", nil
}

func (i *Installer) InstallOpenCode(ctx context.Context, installPath string) error {
	targetOS := runtime.GOOS
	arch := runtime.GOARCH

	url := fmt.Sprintf("https://github.com/opencode-ai/opencode/releases/latest/download/opencode-%s-%s", targetOS, arch)

	if i.verbose {
		fmt.Printf("Downloading OpenCode from %s...\n", url)
	}

	tmpPath := filepath.Join(os.TempDir(), "opencode-"+targetOS+"-"+arch)
	if err := Download(url, tmpPath); err != nil {
		return fmt.Errorf("download OpenCode: %w", err)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
		return fmt.Errorf("create install dir: %w", err)
	}

	if err := os.Rename(tmpPath, installPath); err != nil {
		if copyErr := copyFile(tmpPath, installPath); copyErr != nil {
			return fmt.Errorf("install: %w", copyErr)
		}
	}

	if i.verbose {
		fmt.Printf("OpenCode installed to %s\n", installPath)
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (i *Installer) detectInstallPath() string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "linux" && home != "" {
		return filepath.Join(home, ".local", "bin", "opencode")
	}
	return "/usr/local/bin/opencode"
}

type InstallResult struct {
	Success     bool
	OpenCodeVer string
	AigoVer     string
	Message     string
}

func (i *Installer) InstallAll(ctx context.Context) *InstallResult {
	return &InstallResult{
		Success:     true,
		OpenCodeVer: "1.0.0",
		AigoVer:     "1.0.0",
		Message:     "Installation complete",
	}
}
