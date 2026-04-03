package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type eventBroadcaster struct {
	clients map[chan string]struct{}
}

func newEventBroadcaster() *eventBroadcaster {
	return &eventBroadcaster{
		clients: make(map[chan string]struct{}),
	}
}

func (eb *eventBroadcaster) add(ch chan string) {
	eb.clients[ch] = struct{}{}
}

func (eb *eventBroadcaster) remove(ch chan string) {
	delete(eb.clients, ch)
	close(ch)
}

func (eb *eventBroadcaster) broadcast(event string) {
	for ch := range eb.clients {
		select {
		case ch <- event:
		default:
		}
	}
}

var broadcaster = newEventBroadcaster()

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func readJSON(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listSessionsHandler(w, r)
	case http.MethodPost:
		s.createSessionHandler(w, r)
	}
}

func (s *Server) listSessionsHandler(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.db.ListSessions()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (s *Server) createSessionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Workspace string `json:"workspace"`
	}
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	sess, err := s.db.CreateSession(req.ID, req.Name, req.Workspace)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	broadcaster.broadcast(`{"type":"session_created","id":"` + sess.ID + `"}`)
	writeJSON(w, http.StatusCreated, sess)
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sess, err := s.db.GetSession(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		return
	}
	messages, err := s.db.GetMessages(id, 100)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"session":  sess,
		"messages": messages,
	})
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTasksHandler(w, r)
	case http.MethodPost:
		s.createTaskHandler(w, r)
	}
}

func (s *Server) listTasksHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	tasks, err := s.db.ListTasks(sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID   string `json:"session_id"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
	}
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.Priority == "" {
		req.Priority = types.PriorityMedium
	}
	task, err := s.db.AddTask(req.SessionID, req.Description, req.Priority)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	broadcaster.broadcast(`{"type":"task_created","id":` + strconv.FormatInt(task.ID, 10) + `}`)
	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) handleTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid task id"})
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	err = s.db.UpdateTaskStatus(id, types.TaskInProgress, "")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	broadcaster.broadcast(`{"type":"task_started","id":` + strconv.FormatInt(id, 10) + `}`)
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	agents := []map[string]any{
		{"name": "default", "status": "idle", "type": "general"},
		{"name": "coder", "status": "idle", "type": "coding"},
		{"name": "explorer", "status": "idle", "type": "research"},
	}
	writeJSON(w, http.StatusOK, agents)
}

func (s *Server) handleGateways(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getGatewaysHandler(w, r)
	case http.MethodPost:
		s.gatewayActionHandler(w, r)
	}
}

func (s *Server) getGatewaysHandler(w http.ResponseWriter, r *http.Request) {
	platforms := []map[string]any{
		{"name": "telegram", "connected": false, "status": "disconnected"},
		{"name": "discord", "connected": false, "status": "disconnected"},
		{"name": "slack", "connected": false, "status": "disconnected"},
		{"name": "whatsapp", "connected": false, "status": "disconnected"},
	}
	writeJSON(w, http.StatusOK, platforms)
}

func (s *Server) gatewayActionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Platform string `json:"platform"`
		Action   string `json:"action"`
		Token    string `json:"token,omitempty"`
	}
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	status := "connected"
	if req.Action == "disconnect" {
		status = "disconnected"
	}
	broadcaster.broadcast(`{"type":"gateway_` + req.Action + `","platform":"` + req.Platform + `"}`)
	writeJSON(w, http.StatusOK, map[string]string{"platform": req.Platform, "status": status})
}

func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.searchMemoryHandler(w, r)
	case http.MethodPost:
		s.addMemoryHandler(w, r)
	case http.MethodDelete:
		s.deleteMemoryHandler(w, r)
	}
}

func (s *Server) searchMemoryHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	var memories []types.Memory
	var err error
	if query != "" {
		memories, err = s.db.SearchMemory(query)
	} else {
		memories, err = s.db.ListMemories(category)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, memories)
}

func (s *Server) addMemoryHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content  string `json:"content"`
		Category string `json:"category"`
		Tags     string `json:"tags"`
	}
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	err := s.db.AddMemory(req.Content, req.Category, req.Tags)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	broadcaster.broadcast(`{"type":"memory_added"}`)
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (s *Server) deleteMemoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid memory id"})
		return
	}
	err = s.db.DeleteMemory(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.db.ListSessions()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	memories, err := s.db.ListMemories("")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sessions": len(sessions),
		"tasks":    0,
		"memories": len(memories),
		"agents":   3,
		"gateways": len(s.cfg.Gateway.Platforms),
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.cfg)
	case http.MethodPost:
		var req types.Config
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		s.cfg = req
		writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
	}
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan string, 10)
	broadcaster.add(ch)
	defer broadcaster.remove(ch)

	fmt.Fprint(w, "retry: 3000\n\n")
	flusher.Flush()

	for {
		select {
		case event := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", event)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleEventsWS(w http.ResponseWriter, r *http.Request) {
	ch := make(chan string, 10)
	broadcaster.add(ch)
	defer broadcaster.remove(ch)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	fmt.Fprint(w, "retry: 3000\n\n")
	flusher.Flush()

	for {
		select {
		case event := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", event)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.db.DeleteSessionHistory(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	broadcaster.broadcast(`{"type":"session_deleted","id":"` + id + `"}`)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleConfigUpdate(w http.ResponseWriter, r *http.Request) {
	var req types.Config
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.cfg = req
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	skillsDir := os.ExpandEnv("$HOME/.aigo/skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}

	var skills []map[string]any
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillPath := filepath.Join(skillsDir, entry.Name())
		skillFile := filepath.Join(skillPath, "SKILL.md")
		info, err := os.Stat(skillFile)
		enabled := err == nil && !info.IsDir()

		skills = append(skills, map[string]any{
			"name":        entry.Name(),
			"description": "",
			"category":    "general",
			"enabled":     enabled,
		})
	}

	if skills == nil {
		skills = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, skills)
}

func sanitizeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
