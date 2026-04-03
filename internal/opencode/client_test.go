package opencode

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewClient_WithValidBinary(t *testing.T) {
	t.Parallel()

	client, err := NewClient("/bin/sh", 30, "/tmp")
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, "/bin/sh", client.binary)
	require.Equal(t, 30*time.Second, client.timeout)
	require.Equal(t, "/tmp", client.workdir)
}

func TestNewClient_WithEmptyBinaryPath(t *testing.T) {
	t.Parallel()

	// When binary is empty, NewClient auto-detects "opencode" from PATH.
	// If opencode is not installed, it should return an error.
	// We test this by temporarily ensuring no opencode is available.
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "/nonexistent/path")
	defer os.Setenv("PATH", origPath)

	_, err := NewClient("", 30, "/tmp")
	require.Error(t, err)
}

func TestClient_buildArgs_Basic(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "opencode",
		timeout: 30 * time.Second,
		workdir: "/tmp",
	}

	args := client.buildArgs("session123", "hello world", nil)
	require.Len(t, args, 3)
	require.Equal(t, "run", args[0])
	require.Equal(t, "--session=session123", args[1])
	require.Equal(t, "--prompt=hello world", args[2])
}

func TestClient_buildArgs_WithFiles(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "opencode",
		timeout: 30 * time.Second,
		workdir: "/tmp",
	}

	args := client.buildArgs("session123", "hello world", []string{"file1.go", "file2.go"})
	require.Len(t, args, 4)
	require.Equal(t, "run", args[0])
	require.Equal(t, "--session=session123", args[1])
	require.Equal(t, "--prompt=hello world", args[2])
	require.Equal(t, "--files=file1.go,file2.go", args[3])
}

func TestClient_buildArgs_WithEmptyFilesSlice(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "opencode",
		timeout: 30 * time.Second,
		workdir: "/tmp",
	}

	args := client.buildArgs("session123", "hello", []string{})
	require.Len(t, args, 3)
	require.Equal(t, "run", args[0])
	require.Equal(t, "--session=session123", args[1])
	require.Equal(t, "--prompt=hello", args[2])
}

func TestClient_Run_Success(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "echo",
		timeout: 5 * time.Second,
		workdir: "/tmp",
	}

	result, err := client.Run(context.Background(), "test", "session123")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Contains(t, result.Output, "test")
}

func TestClient_RunWithFiles_Success(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "echo",
		timeout: 5 * time.Second,
		workdir: "/tmp",
	}

	result, err := client.RunWithFiles(context.Background(), "test", "session123", []string{"file1.go"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Contains(t, result.Output, "test")
}

func TestClient_StreamOutput_Success(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "echo",
		timeout: 5 * time.Second,
		workdir: "/tmp",
	}

	var lines []string
	callback := func(line string) {
		lines = append(lines, line)
	}

	err := client.StreamOutput(context.Background(), "hello", "session123", callback)
	require.NoError(t, err)
	require.Len(t, lines, 1)
	require.Contains(t, lines[0], "hello")
}

func TestNewClient_InvalidBinaryPath(t *testing.T) {
	t.Parallel()

	client, err := NewClient("/this/does/not/exist", 30, "/tmp")
	require.Error(t, err)
	require.Nil(t, client)
}

func TestNewClient_NonExecutableFile(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "test")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	os.Chmod(f.Name(), 0644)

	_, err = NewClient(f.Name(), 30, "/tmp")
	require.Error(t, err)
}

func TestClient_timeoutIsCorrectlySet(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "opencode",
		timeout: 42 * time.Second,
		workdir: "/tmp",
	}

	require.Equal(t, 42*time.Second, client.timeout)
}

func TestClient_workdirIsCorrectlySet(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "opencode",
		timeout: 30 * time.Second,
		workdir: "/custom/workdir",
	}

	require.Equal(t, "/custom/workdir", client.workdir)
}

func TestClient_CheckVersion_ReturnsValue(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "echo",
		timeout: 5 * time.Second,
		workdir: "/tmp",
	}

	result, err := client.CheckVersion()
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestClient_HealthCheck_ReturnsErrorForInvalidCommand(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "/nonexistent/binary",
		timeout: 5 * time.Second,
		workdir: "/tmp",
	}

	result, err := client.HealthCheck()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Success)
	require.Contains(t, result.Error, "version check failed")
}

func TestClient_execCommandError_ReturnsError(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "/nonexistent/binary",
		timeout: 5 * time.Second,
		workdir: "/tmp",
	}

	result, err := client.execCommand(context.Background(), []string{"arg1"})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Success)
	require.NotEmpty(t, result.Error)
}

func TestClient_Timeout_Exceeded(t *testing.T) {
	t.Parallel()

	client := &Client{
		binary:  "sleep",
		timeout: 1 * time.Second,
		workdir: "/tmp",
	}

	result, err := client.Run(context.Background(), "2", "session123")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Success)
	require.NotEmpty(t, result.Error)
}
