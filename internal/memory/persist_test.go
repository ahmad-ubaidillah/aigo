package memory

import (
	"os"
	"testing"
)

func TestSQLiteStore_SaveAndGet(t *testing.T) {
	tmpFile := "/tmp/test_memory.db"
	os.Remove(tmpFile)

	store, err := NewSQLiteStore(tmpFile)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()
	defer os.Remove(tmpFile)

	err = store.SaveMemory("test content", "daily", nil)
	if err != nil {
		t.Fatalf("SaveMemory failed: %v", err)
	}

	record, err := store.GetMemory(1)
	if err != nil {
		t.Fatalf("GetMemory failed: %v", err)
	}

	if record.Content != "test content" {
		t.Errorf("Expected 'test content', got '%s'", record.Content)
	}
}

func TestSQLiteStore_Search(t *testing.T) {
	tmpFile := "/tmp/test_memory_search.db"
	os.Remove(tmpFile)

	store, err := NewSQLiteStore(tmpFile)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()
	defer os.Remove(tmpFile)

	store.SaveMemory("hello world", "daily", nil)
	store.SaveMemory("foo bar", "daily", nil)
	store.SaveMemory("test content", "longterm", nil)

	results, err := store.SearchMemories("hello", 10)
	if err != nil {
		t.Fatalf("SearchMemories failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestSQLiteStore_Update(t *testing.T) {
	tmpFile := "/tmp/test_memory_update.db"
	os.Remove(tmpFile)

	store, err := NewSQLiteStore(tmpFile)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()
	defer os.Remove(tmpFile)

	store.SaveMemory("original", "daily", nil)

	err = store.UpdateMemory(1, "updated", nil)
	if err != nil {
		t.Fatalf("UpdateMemory failed: %v", err)
	}

	record, _ := store.GetMemory(1)
	if record.Content != "updated" {
		t.Errorf("Expected 'updated', got '%s'", record.Content)
	}
}

func TestSQLiteStore_Delete(t *testing.T) {
	tmpFile := "/tmp/test_memory_delete.db"
	os.Remove(tmpFile)

	store, err := NewSQLiteStore(tmpFile)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()
	defer os.Remove(tmpFile)

	store.SaveMemory("to delete", "daily", nil)

	err = store.DeleteMemory(1)
	if err != nil {
		t.Fatalf("DeleteMemory failed: %v", err)
	}

	_, err = store.GetMemory(1)
	if err == nil {
		t.Error("Should fail after delete")
	}
}

func TestSQLiteStore_ListByCategory(t *testing.T) {
	tmpFile := "/tmp/test_memory_list.db"
	os.Remove(tmpFile)

	store, err := NewSQLiteStore(tmpFile)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()
	defer os.Remove(tmpFile)

	store.SaveMemory("daily1", "daily", nil)
	store.SaveMemory("daily2", "daily", nil)
	store.SaveMemory("longterm1", "longterm", nil)

	results, err := store.ListMemories("daily", 10)
	if err != nil {
		t.Fatalf("ListMemories failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 daily, got %d", len(results))
	}
}