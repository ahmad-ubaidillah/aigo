package handlers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
	"github.com/stretchr/testify/require"
)

// Test FileHandler
func TestFileHandler_CanHandle(t *testing.T) {
	t.Parallel()

	h := &FileHandler{}
	require.True(t, h.CanHandle(types.IntentFile))
	require.False(t, h.CanHandle(types.IntentWeb))
	require.False(t, h.CanHandle(types.IntentGeneral))
}

func TestFileHandler_Read(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "testfile")
	require.NoError(t, err)
	testContent := "Hello, World!\nThis is a test file."
	_, err = f.WriteString(testContent)
	require.NoError(t, err)
	f.Close()

	h := &FileHandler{}
	result, err := h.Read(f.Name())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Equal(t, testContent, result.Output)

	os.Remove(f.Name())
}

func TestFileHandler_Read_NonExistentFile(t *testing.T) {
	t.Parallel()

	h := &FileHandler{}
	_, err := h.Read("/this/file/does/not/exist")
	require.Error(t, err)
}

func TestFileHandler_Write(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "testdir")
	require.NoError(t, err)
	f.Close()
	os.Remove(f.Name())

	testPath := f.Name() + "/testfile.txt"
	testContent := "Hello, World!\nThis is a test file."

	h := &FileHandler{}
	result, err := h.Write(testPath, testContent)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)

	content, err := os.ReadFile(testPath)
	require.NoError(t, err)
	require.Equal(t, testContent, string(content))

	os.RemoveAll(f.Name())
}

func TestFileHandler_List(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	f1, err := os.CreateTemp(dir, "file1")
	require.NoError(t, err)
	f1.Close()
	f2, err := os.CreateTemp(dir, "file2")
	require.NoError(t, err)
	f2.Close()

	subdir := filepath.Join(dir, "subdir")
	err = os.Mkdir(subdir, 0755)
	require.NoError(t, err)

	h := &FileHandler{}
	result, err := h.List(dir)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Contains(t, result.Output, "[file] ")
	require.Contains(t, result.Output, "[dir] subdir")
}

func TestFileHandler_Execute_UnknownCommand(t *testing.T) {
	t.Parallel()

	h := &FileHandler{}
	task := &types.Task{
		Description: "unknown",
	}
	result, err := h.Execute(context.Background(), task, "/tmp")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Success)
	require.Contains(t, result.Error, "unknown file command")
}

func TestFileHandler_Execute_ReadCommand(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "testfile")
	require.NoError(t, err)
	testContent := "Hello, World!"
	_, err = f.WriteString(testContent)
	require.NoError(t, err)
	f.Close()

	h := &FileHandler{}
	task := &types.Task{
		Description: "read " + f.Name(),
	}
	result, err := h.Execute(context.Background(), task, "/tmp")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)
	require.Equal(t, testContent, result.Output)

	os.Remove(f.Name())
}

func TestFileHandler_Execute_WriteCommand(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "testdir")
	require.NoError(t, err)
	f.Close()
	os.Remove(f.Name())

	h := &FileHandler{}
	task := &types.Task{
		Description: "write " + f.Name() + "/test.txt Hello World",
	}
	result, err := h.Execute(context.Background(), task, "/tmp")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)

	content, err := os.ReadFile(f.Name() + "/test.txt")
	require.NoError(t, err)
	require.Equal(t, "Hello World", string(content))

	os.RemoveAll(f.Name())
}

// Test other handlers briefly
func TestGeneralHandler_CanHandle(t *testing.T) {
	t.Parallel()

	h := &GeneralHandler{}
	require.True(t, h.CanHandle(types.IntentGeneral))
	require.False(t, h.CanHandle(types.IntentWeb))
	require.False(t, h.CanHandle(types.IntentFile))
}

func TestAutomationHandler_CanHandle(t *testing.T) {
	t.Parallel()

	h := &AutomationHandler{}
	require.True(t, h.CanHandle(types.IntentAutomation))
	require.False(t, h.CanHandle(types.IntentWeb))
	require.False(t, h.CanHandle(types.IntentFile))
}

// Test gateway handler
func TestGatewayHandler_CanHandle(t *testing.T) {
	t.Parallel()

	h := &GatewayHandler{}
	require.True(t, h.CanHandle(types.IntentGateway))
	require.False(t, h.CanHandle(types.IntentWeb))
	require.False(t, h.CanHandle(types.IntentFile))
}
