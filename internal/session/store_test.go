package session

import (
	"os"
	"testing"
)

func TestSessionStore_SaveAndLoad(t *testing.T) {
	tmpFile := "/tmp/test_sessions.db"
	os.Remove(tmpFile)

	store, err := NewSessionStore(tmpFile)
	if err != nil {
		t.Fatalf("NewSessionStore failed: %v", err)
	}
	defer store.Close()
	defer os.Remove(tmpFile)

	state := []byte(`{"key": "value"}`)
	err = store.SaveSession("test-session", state)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	loaded, err := store.LoadSession("test-session")
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}

	if string(loaded) != string(state) {
		t.Errorf("Expected %s, got %s", state, loaded)
	}
}

func TestSessionStore_Delete(t *testing.T) {
	tmpFile := "/tmp/test_sessions_del.db"
	os.Remove(tmpFile)

	store, err := NewSessionStore(tmpFile)
	if err != nil {
		t.Fatalf("NewSessionStore failed: %v", err)
	}
	defer store.Close()
	defer os.Remove(tmpFile)

	store.SaveSession("to-delete", []byte("state"))

	err = store.DeleteSession("to-delete")
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	_, err = store.LoadSession("to-delete")
	if err == nil {
		t.Error("Should fail after delete")
	}
}

func TestSessionStore_AddMessage(t *testing.T) {
	tmpFile := "/tmp/test_sessions_msg.db"
	os.Remove(tmpFile)

	store, err := NewSessionStore(tmpFile)
	if err != nil {
		t.Fatalf("NewSessionStore failed: %v", err)
	}
	defer store.Close()
	defer os.Remove(tmpFile)

	store.SaveSession("msg-session", []byte("{}"))

	err = store.AddMessage("msg-session", "user", "Hello", nil)
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	messages, err := store.GetMessages("msg-session", 10)
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Content != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", messages[0].Content)
	}
}

func TestSessionStore_ListSessions(t *testing.T) {
	tmpFile := "/tmp/test_sessions_list.db"
	os.Remove(tmpFile)

	store, err := NewSessionStore(tmpFile)
	if err != nil {
		t.Fatalf("NewSessionStore failed: %v", err)
	}
	defer store.Close()
	defer os.Remove(tmpFile)

	store.SaveSession("session-1", []byte("{}"))
	store.SaveSession("session-2", []byte("{}"))

	sessions, err := store.ListSessions(10)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
}

func TestMarshalState(t *testing.T) {
	data := map[string]string{"key": "value"}
	bytes, err := MarshalState(data)
	if err != nil {
		t.Fatalf("MarshalState failed: %v", err)
	}

	var result map[string]string
	err = UnmarshalState(bytes, &result)
	if err != nil {
		t.Fatalf("UnmarshalState failed: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("Expected 'value', got '%s'", result["key"])
	}
}