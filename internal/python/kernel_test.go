package python

import (
	"testing"
	"time"
)

func TestNewKernel(t *testing.T) {
	t.Parallel()
	k := NewKernel(&KernelOptions{WorkDir: t.TempDir()})
	if k == nil {
		t.Error("expected kernel")
	}
	if k.timeout != 60*time.Second {
		t.Errorf("expected 60s timeout, got %v", k.timeout)
	}
}

func TestNewKernelCustomTimeout(t *testing.T) {
	t.Parallel()
	k := NewKernel(&KernelOptions{Timeout: 30 * time.Second, WorkDir: t.TempDir()})
	if k.timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", k.timeout)
	}
}

func TestKernelOptions(t *testing.T) {
	t.Parallel()
	opts := KernelOptions{
		Timeout:  10 * time.Second,
		WorkDir:  "/tmp",
		Packages: []string{"numpy", "pandas"},
	}
	if opts.Timeout != 10*time.Second {
		t.Errorf("expected 10s, got %v", opts.Timeout)
	}
	if len(opts.Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(opts.Packages))
	}
}

func TestExecResult(t *testing.T) {
	t.Parallel()
	r := ExecResult{
		Success:     true,
		Output:      "hello",
		Error:       "",
		ReturnValue: "42",
		Duration:    time.Second,
	}
	if !r.Success {
		t.Error("expected success")
	}
	if r.Output != "hello" {
		t.Errorf("expected hello, got %s", r.Output)
	}
}

func TestKernel_PackageManagement(t *testing.T) {
	t.Parallel()
	k := NewKernel(&KernelOptions{WorkDir: t.TempDir()})
	if k.packages == nil {
		t.Error("expected packages map")
	}
}

func TestKernel_Timeout(t *testing.T) {
	t.Parallel()
	k := NewKernel(&KernelOptions{Timeout: 10 * time.Second, WorkDir: t.TempDir()})
	if k.timeout != 10*time.Second {
		t.Errorf("expected 10s, got %v", k.timeout)
	}
}

func TestKernel_WorkDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	k := NewKernel(&KernelOptions{WorkDir: dir})
	if k.workDir != dir {
		t.Errorf("expected %s, got %s", dir, k.workDir)
	}
}

func TestKernel_IsRunning(t *testing.T) {
	t.Parallel()
	k := NewKernel(&KernelOptions{WorkDir: t.TempDir()})
	if k.isRunning {
		t.Error("expected not running")
	}
}

func TestKernel_PackagesMap(t *testing.T) {
	t.Parallel()
	k := NewKernel(&KernelOptions{WorkDir: t.TempDir()})
	k.packages["numpy"] = true
	if !k.packages["numpy"] {
		t.Error("expected numpy to be installed")
	}
}
