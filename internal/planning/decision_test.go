package planning

import (
	"testing"
)

func TestDecision_Track(t *testing.T) {
	d := NewDecision("add auth", "use JWT")
	if d == nil {
		t.Fatal("NewDecision() returned nil")
	}

	if d.Key != "add auth" {
		t.Errorf("expected key 'add auth', got '%s'", d.Key)
	}
}

func TestDecisionStore_Add(t *testing.T) {
	store := NewDecisionStore()
	store.Add("key1", "value1")
	store.Add("key2", "value2")

	if store.Len() != 2 {
		t.Errorf("expected 2 decisions, got %d", store.Len())
	}
}

func TestDecisionStore_Get(t *testing.T) {
	store := NewDecisionStore()
	store.Add("key1", "value1")

	val := store.Get("key1")
	if val == nil || val.Value != "value1" {
		t.Errorf("expected 'value1', got '%v'", val)
	}
}