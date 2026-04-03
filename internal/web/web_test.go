package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

func TestNewServer(t *testing.T) {
	t.Parallel()
	s := NewServer(":8080", nil, types.Config{})
	if s == nil {
		t.Error("expected server")
	}
}

func TestServer_HandleHealth(t *testing.T) {
	t.Parallel()
	s := NewServer(":8080", nil, types.Config{})
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	s.handleHealth(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestServer_HandleConfig(t *testing.T) {
	t.Parallel()
	s := NewServer(":8080", nil, types.Config{})
	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()
	s.handleConfig(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestServer_HandleSkills(t *testing.T) {
	t.Parallel()
	s := NewServer(":8080", nil, types.Config{})
	req := httptest.NewRequest("GET", "/api/skills", nil)
	w := httptest.NewRecorder()
	s.handleSkills(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
