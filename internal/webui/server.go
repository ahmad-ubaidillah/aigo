// Package webui implements the Web UI — installer, dashboard, chat, and settings.
// Mobile-first responsive design with typing indicator.
package webui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hermes-v2/aigo/internal/skillhub"
)

//go:embed static/dashboard.html
var dashboardHTML string

//go:embed static/install.html
var installHTML string

// Server serves the Web UI with chat, settings, and installer.
type Server struct {
	port    int
	config  interface{}
	onSetup func(config map[string]interface{}) error
	onChat  func(message string) (string, error)

	// Streaming chat handler (optional)
	onChatStream func(message string, onChunk func(string)) (string, error)

	// Live stats
	mu        sync.RWMutex
	startTime time.Time
	msgCount  int
	channels  []string
	tools     int
	model     string
	provider  string

	// Skill hub
	skillHub *skillhub.OnlineHub

	// Security middleware
	auth      *AuthMiddleware
	rateLimit *RateLimitMiddleware
	cors      *CORSMiddleware
	apiKey    string
}

// New creates a new Web UI server.
func New(port int, config interface{}, onSetup func(map[string]interface{}) error) *Server {
	return &Server{
		port:      port,
		config:    config,
		onSetup:   onSetup,
		startTime: time.Now(),
	}
}

// SetSecurity configures auth, rate limiting, and CORS.
func (s *Server) SetSecurity(apiKey string, maxRequests int, rateWindow time.Duration, allowedOrigins []string) {
	s.apiKey = apiKey
	s.auth = NewAuthMiddleware(apiKey)
	s.rateLimit = NewRateLimitMiddleware(maxRequests, rateWindow)
	s.cors = NewCORSMiddleware(allowedOrigins)
}

func (s *Server) SetChatHandler(fn func(string) (string, error)) {
	s.onChat = fn
}

func (s *Server) SetChatStreamHandler(fn func(string, func(string)) (string, error)) {
	s.onChatStream = fn
}

func (s *Server) SetStats(channels []string, tools int, model, provider string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels = channels
	s.tools = tools
	s.model = model
	s.provider = provider
}

func (s *Server) SetSkillHub(hub *skillhub.OnlineHub) {
	s.skillHub = hub
}

func (s *Server) IncrMessages() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msgCount++
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Public wrapper: cors -> rateLimit -> handler (no auth)
	wrapPub := func(h http.HandlerFunc) http.HandlerFunc {
		if s.cors != nil {
			h = s.cors.Wrap(h)
		}
		if s.rateLimit != nil {
			h = s.rateLimit.Wrap(h)
		}
		return h
	}

	// API wrapper: cors -> rateLimit -> auth -> handler
	wrapAPI := func(h http.HandlerFunc) http.HandlerFunc {
		if s.cors != nil {
			h = s.cors.Wrap(h)
		}
		if s.rateLimit != nil {
			h = s.rateLimit.Wrap(h)
		}
		if s.auth != nil {
			h = s.auth.Wrap(h)
		}
		return h
	}

	// Public HTML pages
	mux.HandleFunc("/", wrapPub(s.handleDashboard))
	mux.HandleFunc("/install", wrapPub(s.handleInstall))
	mux.HandleFunc("/health", wrapPub(s.handleHealth))
	mux.HandleFunc("/ready", wrapPub(s.handleReady))

	// API endpoints (protected)
	mux.HandleFunc("/install/api", wrapAPI(s.handleInstallAPI))
	mux.HandleFunc("/api/status", wrapAPI(s.handleStatus))
	mux.HandleFunc("/api/config", wrapAPI(s.handleConfig))
	mux.HandleFunc("/api/config/save", wrapAPI(s.handleConfigSave))
	mux.HandleFunc("/api/chat", wrapAPI(s.handleChat))
	mux.HandleFunc("/api/chat/stream", wrapAPI(s.handleChatStream))
	mux.HandleFunc("/api/stats", wrapAPI(s.handleStats))

	// Skills API
	mux.HandleFunc("/api/skills/search", wrapAPI(s.handleSkillSearch))
	mux.HandleFunc("/api/skills/popular", wrapAPI(s.handleSkillPopular))
	mux.HandleFunc("/api/skills/install", wrapAPI(s.handleSkillInstall))
	mux.HandleFunc("/api/skills/sync", wrapAPI(s.handleSkillSync))
	mux.HandleFunc("/api/skills/sources", wrapAPI(s.handleSkillSources))
	mux.HandleFunc("/api/skills/list", wrapAPI(s.handleSkillList))
	mux.HandleFunc("/api/skills/stats", wrapAPI(s.handleSkillStats))
	mux.HandleFunc("/api/skills/detail", wrapAPI(s.handleSkillDetail))
	mux.HandleFunc("/api/skills/browse", wrapAPI(s.handleSkillBrowse))

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🌐 Web UI: http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := dashboardHTML
	if s.apiKey != "" {
		// Escape the API key for safe JavaScript string literal
		escaped := strings.ReplaceAll(s.apiKey, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `'`, `\'`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		escaped = strings.ReplaceAll(escaped, "\r", `\r`)
		html = strings.Replace(html, "/*INJECT_API_KEY*/", escaped, 1)
	} else {
		html = strings.Replace(html, "/*INJECT_API_KEY*/", "", 1)
	}
	fmt.Fprint(w, html)
}

func (s *Server) handleInstall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, installHTML)
}

func (s *Server) handleInstallAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	if s.onSetup != nil {
		if err := s.onSetup(config); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	if s.onChat == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Chat not configured"})
		return
	}
	s.IncrMessages()
	response, err := s.onChat(req.Message)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "response": response})
}

// handleChatStream handles streaming chat via SSE.
func (s *Server) handleChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if s.onChatStream == nil && s.onChat == nil {
		http.Error(w, "Chat not configured", 500)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", 500)
		return
	}

	s.IncrMessages()

	// Send chunk function
	sendChunk := func(text string) {
		// Escape newlines for SSE format
		escaped := strings.ReplaceAll(text, "\n", "\\n")
		fmt.Fprintf(w, "data: {\"chunk\":\"%s\"}\n\n", escaped)
		flusher.Flush()
	}

	// Use streaming handler if available, otherwise fall back to non-streaming
	var fullResponse string
	var err error

	if s.onChatStream != nil {
		fullResponse, err = s.onChatStream(req.Message, func(chunk string) {
			sendChunk(chunk)
		})
	} else {
		// Non-streaming fallback — send full response at once
		fullResponse, err = s.onChat(req.Message)
		if err == nil {
			sendChunk(fullResponse)
		}
	}

	if err != nil {
		fmt.Fprintf(w, "data: {\"error\":\"%s\"}\n\n", strings.ReplaceAll(err.Error(), "\"", "\\\""))
		flusher.Flush()
		return
	}

	// Send done signal
	fmt.Fprintf(w, "data: {\"done\":true}\n\n")
	flusher.Flush()
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "running",
		"uptime":  time.Since(s.startTime).Round(time.Second).String(),
		"version": "0.3.0",
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "healthy"})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ready",
		"uptime": time.Since(s.startTime).Round(time.Second).String(),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": s.msgCount,
		"channels": s.channels,
		"tools":    s.tools,
		"model":    s.model,
		"provider": s.provider,
		"uptime":   time.Since(s.startTime).Round(time.Second).String(),
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}

func (s *Server) handleConfigSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var newConfig map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	if s.onSetup != nil {
		if err := s.onSetup(newConfig); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
			return
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "message": "Config saved. Restart required."})
}

// --- Skills API Handlers ---

func (s *Server) handleSkillSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.skillHub == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Skill hub not initialized"})
		return
	}
	query := r.URL.Query().Get("q")
	if query == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Missing query parameter 'q'"})
		return
	}
	results, err := s.skillHub.Search(query, 50)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "skills": results, "count": len(results)})
}

func (s *Server) handleSkillPopular(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.skillHub == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Skill hub not initialized"})
		return
	}
	results, err := s.skillHub.PopularSkills(50)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "skills": results, "count": len(results)})
}

func (s *Server) handleSkillInstall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "POST required"})
		return
	}
	if s.skillHub == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Skill hub not initialized"})
		return
	}
	var req struct {
		Identifier string `json:"identifier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	if err := s.skillHub.Install(req.Identifier); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "message": "Installed: " + req.Identifier})
}

func (s *Server) handleSkillSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "POST required"})
		return
	}
	if s.skillHub == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Skill hub not initialized"})
		return
	}
	result, err := s.skillHub.SyncIndex()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":       true,
		"total":    result.TotalNew,
		"synced":   result.Synced,
		"errors":   result.Errors,
		"message":  result.String(),
	})
}

func (s *Server) handleSkillSources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.skillHub == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Skill hub not initialized"})
		return
	}
	sources := s.skillHub.ListSources()
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "sources": sources})
}

func (s *Server) handleSkillList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.skillHub == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Skill hub not initialized"})
		return
	}
	skills, err := s.skillHub.ListInstalled()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "skills": skills, "count": len(skills)})
}

func (s *Server) handleSkillStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.skillHub == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Skill hub not initialized"})
		return
	}
	stats := s.skillHub.Stats()
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "stats": stats})
}

func (s *Server) handleSkillDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.skillHub == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Skill hub not initialized"})
		return
	}
	identifier := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	if identifier == "" && name == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Missing 'id' or 'name' parameter"})
		return
	}

	var skill *skillhub.Skill
	var err error
	if identifier != "" {
		skill, err = s.skillHub.FindByIdentifier(identifier)
	}
	if skill == nil && name != "" {
		skill, err = s.skillHub.FindByName(name)
	}
	if err != nil || skill == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Skill not found"})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "skill": skill})
}

func (s *Server) handleSkillBrowse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.skillHub == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "Skill hub not initialized"})
		return
	}
	source := r.URL.Query().Get("source")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
		if limit <= 0 || limit > 200 {
			limit = 50
		}
	}

	skills, err := s.skillHub.BrowseOnline(source, limit)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "skills": skills, "count": len(skills), "source": source})
}
