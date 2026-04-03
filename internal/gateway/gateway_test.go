package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test the Manager functionality
func TestNewManager(t *testing.T) {
	t.Parallel()

	m := NewManager()
	require.NotNil(t, m)
	require.NotNil(t, m.platforms)
	require.NotNil(t, m.running)
	require.Len(t, m.platforms, 0)
	require.Len(t, m.running, 0)
}

func TestManager_RegisterAndGetPlatform(t *testing.T) {
	t.Parallel()

	m := NewManager()
	platform := &dummyPlatform{name: "test"}
	m.Register(platform)

	require.Len(t, m.platforms, 1)
	require.Equal(t, platform, m.platforms["test"])
}

func TestManager_StartPlatformNotFound(t *testing.T) {
	t.Parallel()

	m := NewManager()
	err := m.Start(context.Background(), "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "platform nonexistent not found")
}

func TestManager_StartSuccess(t *testing.T) {
	t.Parallel()

	m := NewManager()
	platform := &dummyPlatform{name: "test", connectErr: nil}
	m.Register(platform)

	err := m.Start(context.Background(), "test")
	require.NoError(t, err)
	require.True(t, m.IsRunning("test"))
}

func TestManager_StartConnectFailure(t *testing.T) {
	t.Parallel()

	m := NewManager()
	platform := &dummyPlatform{name: "test", connectErr: errors.New("connection failed")}
	m.Register(platform)

	err := m.Start(context.Background(), "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "connect test:")
	require.False(t, m.IsRunning("test"))
}

func TestManager_StopPlatformNotFound(t *testing.T) {
	t.Parallel()

	m := NewManager()
	err := m.Stop("nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "platform nonexistent not found")
}

func TestManager_StopSuccess(t *testing.T) {
	t.Parallel()

	m := NewManager()
	platform := &dummyPlatform{name: "test"}
	m.Register(platform)
	// Manually set as running to test stop
	m.mu.Lock()
	m.running["test"] = true
	m.mu.Unlock()

	err := m.Stop("test")
	require.NoError(t, err)
	require.False(t, m.IsRunning("test"))
}

func TestManager_SendMessagePlatformNotFound(t *testing.T) {
	t.Parallel()

	m := NewManager()
	err := m.Send("nonexistent", "chat123", "hello")
	require.Error(t, err)
	require.Contains(t, err.Error(), "platform nonexistent not found")
}

func TestManager_Status(t *testing.T) {
	t.Parallel()

	m := NewManager()
	p1 := &dummyPlatform{name: "platform1"}
	p2 := &dummyPlatform{name: "platform2"}
	m.Register(p1)
	m.Register(p2)

	// Set one as running
	m.mu.Lock()
	m.running["platform1"] = true
	m.mu.Unlock()

	status := m.Status()
	require.Len(t, status, 2)
	require.True(t, status["platform1"])
	require.False(t, status["platform2"])
}

// Test Telegram platform
func TestTelegramPlatform_Name(t *testing.T) {
	p := NewTelegramToken("test-token")
	require.Equal(t, "telegram", p.Name())
}

func TestTelegramPlatform_Connect_EmptyToken(t *testing.T) {
	t.Parallel()

	p := NewTelegramToken("")
	err := p.Connect(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "telegram token is empty")
}

func TestTelegramPlatform_Connect_NoDigits(t *testing.T) {
	t.Parallel()
	t.Skip("requires network")

	p := NewTelegramToken("abcdef")
	err := p.Connect(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "telegram token must contain digits")
}

func TestTelegramPlatform_Connect_ValidToken(t *testing.T) {
	t.Parallel()
	t.Skip("requires network")

	p := NewTelegramToken("12345:ABCDEF")
	err := p.Connect(context.Background())
	require.NoError(t, err)
}

func TestTelegramPlatform_Disconnect(t *testing.T) {
	t.Parallel()
	t.Skip("requires network")

	p := NewTelegramToken("12345:ABCDEF")
	err := p.Disconnect(context.Background())
	require.NoError(t, err)
}

// Helper types for testing
type dummyPlatform struct {
	name       string
	connectErr error
}

func (d *dummyPlatform) Name() string {
	return d.name
}

func (d *dummyPlatform) Connect(ctx context.Context) error {
	return d.connectErr
}

func (d *dummyPlatform) Disconnect(ctx context.Context) error {
	return nil
}

func (d *dummyPlatform) SendMessage(ctx context.Context, chatID, text string) error {
	return nil
}

func (d *dummyPlatform) Listen(ctx context.Context, handler func(Message)) error {
	return nil
}

// Helper to create telegram platform for testing (avoiding direct access to unexported struct)
func NewTelegramToken(token string) *TelegramPlatform {
	return &TelegramPlatform{
		token: token,
	}
}

// Helper method to check if platform is running (accessing unexported field for testing)
func (m *Manager) IsRunning(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running[name]
}
