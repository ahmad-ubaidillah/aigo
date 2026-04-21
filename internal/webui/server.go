// Package webui implements the Web UI — installer, dashboard, chat, and settings.
// Mobile-first responsive design with typing indicator.
package webui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hermes-v2/aigo/internal/skillhub"
)

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

	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/install", s.handleInstall)
	mux.HandleFunc("/install/api", s.handleInstallAPI)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/config/save", s.handleConfigSave)
	mux.HandleFunc("/api/chat", s.handleChat)
	mux.HandleFunc("/api/chat/stream", s.handleChatStream)
	mux.HandleFunc("/api/stats", s.handleStats)

	// Skills API
	mux.HandleFunc("/api/skills/search", s.handleSkillSearch)
	mux.HandleFunc("/api/skills/popular", s.handleSkillPopular)
	mux.HandleFunc("/api/skills/install", s.handleSkillInstall)
	mux.HandleFunc("/api/skills/sync", s.handleSkillSync)
	mux.HandleFunc("/api/skills/sources", s.handleSkillSources)
	mux.HandleFunc("/api/skills/list", s.handleSkillList)
	mux.HandleFunc("/api/skills/stats", s.handleSkillStats)
	mux.HandleFunc("/api/skills/detail", s.handleSkillDetail)
	mux.HandleFunc("/api/skills/browse", s.handleSkillBrowse)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🌐 Web UI: http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, dashboardHTML)
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

// --- Responsive Dashboard HTML ---

const dashboardHTML = `<!DOCTYPE html>
<html><head><title>Aigo</title>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=no">
<meta name="theme-color" content="#0f172a">
<style>
*{margin:0;padding:0;box-sizing:border-box}
:root{--bg:#0f172a;--card:#1e293b;--border:#334155;--text:#e2e8f0;--muted:#94a3b8;--blue:#3b82f6;--cyan:#38bdf8;--pink:#f472b6;--green:#22c55e;--red:#ef4444;--yellow:#facc15;--purple:#a78bfa;--radius:10px}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:var(--bg);color:var(--text);height:100vh;height:100dvh;display:flex;flex-direction:column;overflow:hidden}
.header{background:var(--card);padding:.75rem 1rem;display:flex;align-items:center;justify-content:space-between;border-bottom:1px solid var(--border);flex-shrink:0}
.header h1{font-size:1.2rem;color:var(--cyan)}.header h1 span{color:var(--pink)}
.nav{display:flex;gap:.4rem}
.nav button{background:none;border:1px solid var(--border);color:var(--muted);padding:.45rem .85rem;border-radius:8px;cursor:pointer;font-size:.85rem;transition:all .15s}
.nav button.active,.nav button:hover{background:var(--border);color:var(--text)}
.main{flex:1;display:flex;flex-direction:column;overflow:hidden;position:relative}
.sidebar{display:none;background:var(--card);border-bottom:1px solid var(--border);padding:.75rem;flex-shrink:0}
.sidebar.show{display:flex;gap:.75rem;overflow-x:auto;-webkit-overflow-scrolling:touch}
.stat{background:var(--bg);border:1px solid var(--border);border-radius:var(--radius);padding:.6rem .8rem;min-width:120px;flex-shrink:0}
.stat .label{font-size:.7rem;color:var(--muted);text-transform:uppercase;letter-spacing:.5px}
.stat .value{font-size:1.1rem;font-weight:700;margin-top:.2rem;white-space:nowrap}
.dot{width:8px;height:8px;border-radius:50%;display:inline-block;margin-right:4px}
.dot.on{background:var(--green)}.dot.off{background:var(--red)}

/* Chat */
.chat{flex:1;display:flex;flex-direction:column;overflow:hidden}
.messages{flex:1;overflow-y:auto;padding:1rem;-webkit-overflow-scrolling:touch}
.msg{margin-bottom:.75rem;max-width:85%;animation:fadeIn .3s ease}
@keyframes fadeIn{from{opacity:0;transform:translateY(8px)}to{opacity:1;transform:translateY(0)}}
.msg.user{margin-left:auto;background:var(--blue);border-radius:16px 16px 4px 16px;padding:.6rem .9rem}
.msg.bot{background:var(--card);border:1px solid var(--border);border-radius:16px 16px 16px 4px;padding:.6rem .9rem}
.msg .meta{font-size:.65rem;color:var(--muted);margin-bottom:.2rem}
.msg .text{font-size:.9rem;line-height:1.5;word-break:break-word}
.msg .text code{background:var(--bg);padding:.1rem .3rem;border-radius:4px;font-size:.8rem}
.msg .actions{margin-top:.3rem;display:flex;gap:.5rem}
.msg .actions button{background:none;border:1px solid var(--border);color:var(--muted);padding:.2rem .5rem;border-radius:4px;cursor:pointer;font-size:.7rem}
.msg .actions button:hover{color:var(--text);border-color:var(--cyan)}
.cursor{animation:blink 1s infinite;color:var(--cyan)}
@keyframes blink{0%,100%{opacity:1}50%{opacity:0}}
.typing{display:none;padding:.5rem .9rem;margin-bottom:.75rem}
.typing.show{display:flex;align-items:center;gap:.5rem}
.typing .label{font-size:.8rem;color:var(--muted)}
.typing .dots{display:flex;gap:3px}
.typing .dot{width:6px;height:6px;border-radius:50%;background:var(--cyan);animation:bounce 1.4s infinite}
.typing .dot:nth-child(2){animation-delay:.2s}
.typing .dot:nth-child(3){animation-delay:.4s}
@keyframes bounce{0%,80%,100%{transform:translateY(0);opacity:.3}40%{transform:translateY(-6px);opacity:1}}
.input-bar{padding:.75rem;background:var(--card);border-top:1px solid var(--border);display:flex;gap:.5rem;flex-shrink:0}
.input-bar input{flex:1;padding:.7rem .9rem;border:1px solid var(--border);border-radius:20px;background:var(--bg);color:var(--text);font-size:.95rem;outline:none;-webkit-appearance:none}
.input-bar input:focus{border-color:var(--cyan)}
.input-bar button{width:42px;height:42px;border-radius:50%;background:var(--blue);color:#fff;border:none;cursor:pointer;display:flex;align-items:center;justify-content:center;flex-shrink:0}
.input-bar button:disabled{opacity:.4}
.input-bar button svg{width:18px;height:18px}

/* Panels */
.panel{display:none;flex:1;flex-direction:column;overflow:hidden}
.panel.show{display:flex}

/* Skills */
.skills-header{display:flex;align-items:center;justify-content:space-between;padding:.75rem 1rem;flex-shrink:0}
.skills-header h2{font-size:1.1rem;color:var(--cyan)}
.skills-actions{display:flex;align-items:center;gap:.5rem}
.skills-actions button{background:var(--blue);color:#fff;border:none;padding:.4rem .8rem;border-radius:6px;cursor:pointer;font-size:.85rem}
.skills-actions button:hover{opacity:.9}
.skills-actions button:disabled{opacity:.4}
.skill-stat-badge{font-size:.75rem;color:var(--muted);background:var(--bg);padding:.3rem .6rem;border-radius:12px}
.skills-search{display:flex;gap:.5rem;padding:0 1rem;flex-shrink:0}
.skills-search input{flex:1;padding:.6rem .8rem;border:1px solid var(--border);border-radius:8px;background:var(--bg);color:var(--text);font-size:.9rem;outline:none;-webkit-appearance:none}
.skills-search input:focus{border-color:var(--cyan)}
.skills-search button{width:38px;height:38px;border-radius:8px;background:var(--card);border:1px solid var(--border);cursor:pointer;font-size:1rem;display:flex;align-items:center;justify-content:center}
.skills-tabs{display:flex;gap:.4rem;padding:.75rem 1rem;flex-shrink:0;overflow-x:auto}
.skills-tabs button{background:none;border:1px solid var(--border);color:var(--muted);padding:.4rem .75rem;border-radius:6px;cursor:pointer;font-size:.8rem;white-space:nowrap}
.skills-tabs button.active,.skills-tabs button:hover{background:var(--border);color:var(--text)}
.skills-list{flex:1;overflow-y:auto;padding:0 1rem 1rem;display:flex;flex-direction:column;gap:.5rem}
.skill-card{background:var(--card);border:1px solid var(--border);border-radius:var(--radius);padding:.85rem;cursor:pointer;transition:border-color .2s}
.skill-card:hover{border-color:var(--cyan)}
.skill-card .name{font-weight:600;font-size:.9rem;margin-bottom:.3rem;display:flex;align-items:center;gap:.4rem}
.skill-card .name .trust{font-size:.75rem}
.skill-card .desc{font-size:.8rem;color:var(--muted);line-height:1.4;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden}
.skill-card .meta{font-size:.7rem;color:var(--muted);margin-top:.5rem;display:flex;gap:.5rem;flex-wrap:wrap;align-items:center}
.skill-card .meta .tag{background:var(--bg);padding:.15rem .45rem;border-radius:4px}
.skill-card .meta .installs{color:var(--green)}
.skill-card .meta .source{color:var(--purple)}
.skill-card .actions{margin-top:.6rem;display:flex;gap:.5rem}
.btn-install{background:var(--blue);color:#fff;border:none;padding:.35rem .75rem;border-radius:6px;cursor:pointer;font-size:.78rem;transition:all .15s}
.btn-install:hover{opacity:.9}
.btn-install.done{background:var(--green);cursor:default}
.btn-install.err{background:var(--red)}
.btn-detail{background:none;border:1px solid var(--border);color:var(--muted);padding:.35rem .75rem;border-radius:6px;cursor:pointer;font-size:.78rem}
.btn-detail:hover{color:var(--text);border-color:var(--cyan)}
.loading{text-align:center;color:var(--muted);padding:2rem}
.skeleton{background:linear-gradient(90deg,var(--card) 25%,#253045 50%,var(--card) 75%);background-size:200% 100%;animation:shimmer 1.5s infinite;border-radius:var(--radius);height:80px;margin-bottom:.5rem}
@keyframes shimmer{0%{background-position:200% 0}100%{background-position:-200% 0}}
.empty{text-align:center;color:var(--muted);padding:2rem;font-size:.9rem}

/* Skill Detail */
.skill-detail{display:none;flex:1;overflow-y:auto;padding:1rem}
.skill-detail.show{display:block}
.back-btn{background:none;border:1px solid var(--border);color:var(--muted);padding:.4rem .75rem;border-radius:6px;cursor:pointer;font-size:.85rem;margin-bottom:.75rem}
.back-btn:hover{color:var(--text);border-color:var(--cyan)}
.detail-name{font-size:1.3rem;font-weight:700;margin-bottom:.5rem}
.detail-desc{color:var(--muted);font-size:.9rem;line-height:1.6;margin-bottom:1rem}
.detail-meta{display:flex;gap:1rem;flex-wrap:wrap;margin-bottom:1rem}
.detail-meta .chip{background:var(--bg);border:1px solid var(--border);padding:.3rem .7rem;border-radius:20px;font-size:.8rem}
.detail-tags{display:flex;gap:.4rem;flex-wrap:wrap;margin-bottom:1rem}
.detail-tags .tag{background:var(--bg);padding:.2rem .5rem;border-radius:4px;font-size:.75rem;color:var(--muted)}
.detail-identifier{background:var(--bg);padding:.6rem .8rem;border-radius:8px;font-size:.8rem;color:var(--muted);margin-bottom:1rem;word-break:break-all}
.detail-identifier code{color:var(--cyan)}

/* Settings */
.settings{display:none;flex:1;overflow-y:auto;padding:1rem}
.settings.show{display:block}
.settings h2{color:var(--cyan);font-size:1.1rem;margin-bottom:1rem}
.settings-section{background:var(--card);border:1px solid var(--border);border-radius:var(--radius);padding:1rem;margin-bottom:.75rem}
.settings-section h3{font-size:.95rem;margin-bottom:.75rem;color:var(--text);display:flex;align-items:center;gap:.4rem}
.field{margin-bottom:.75rem}
.field:last-child{margin-bottom:0}
.field label{display:block;font-size:.8rem;color:var(--muted);margin-bottom:.3rem}
.field .hint{font-size:.72rem;color:var(--muted);margin-top:.2rem;opacity:.8}
.field input,.field select,.field textarea{width:100%;padding:.55rem .75rem;border:1px solid var(--border);border-radius:8px;background:var(--bg);color:var(--text);font-size:.9rem;outline:none;-webkit-appearance:none}
.field input:focus,.field select:focus,.field textarea:focus{border-color:var(--cyan)}
.field textarea{min-height:80px;resize:vertical;font-family:monospace}
.field input[type="password"]{font-family:monospace}
.settings-actions{display:flex;gap:.5rem;margin-top:1rem}
.btn-save{background:var(--green);color:#fff;border:none;padding:.6rem 1.5rem;border-radius:8px;cursor:pointer;font-size:.9rem;font-weight:600}
.btn-save:hover{opacity:.9}
.btn-reset{background:none;border:1px solid var(--border);color:var(--muted);padding:.6rem 1.5rem;border-radius:8px;cursor:pointer;font-size:.9rem}
.btn-reset:hover{color:var(--text)}

/* Toast */
.toast{position:fixed;bottom:80px;left:50%;transform:translateX(-50%) translateY(20px);background:var(--card);border:1px solid var(--border);padding:.6rem 1.2rem;border-radius:8px;font-size:.85rem;z-index:100;opacity:0;transition:all .3s;pointer-events:none;max-width:90%}
.toast.show{opacity:1;transform:translateX(-50%) translateY(0)}
.toast.success{border-color:var(--green);color:var(--green)}
.toast.error{border-color:var(--red);color:var(--red)}

/* Responsive */
@media(min-width:768px){
  .sidebar{display:flex !important;flex-wrap:wrap;gap:.75rem;border-bottom:none;border-right:1px solid var(--border);padding:1rem;flex-direction:column;width:260px;flex-shrink:0}
  .main{flex-direction:row}
  .sidebar.show{flex-direction:column;overflow-y:auto}
  .stat{min-width:auto}
  .msg{max-width:70%}
}
</style>
</head>
<body>
<div class="header">
  <h1>Aigo <span>v0.3</span></h1>
  <div class="nav" id="nav">
    <button class="active" data-tab="chat">Chat</button>
    <button data-tab="skills">Skills</button>
    <button data-tab="settings">Settings</button>
    <button data-tab="stats" id="statsBtn">Stats</button>
  </div>
</div>
<div class="main">
  <div class="sidebar" id="sidebar">
    <div class="stat"><div class="label">Status</div><div class="value"><span class="dot on" id="sDot"></span><span id="sStatus">Running</span></div></div>
    <div class="stat"><div class="label">Uptime</div><div class="value" id="sUptime">-</div></div>
    <div class="stat"><div class="label">Messages</div><div class="value" id="sMsgs">0</div></div>
    <div class="stat"><div class="label">Model</div><div class="value" id="sModel" style="font-size:.8rem">-</div></div>
    <div class="stat"><div class="label">Channels</div><div class="value" id="sChannels" style="font-size:.8rem">-</div></div>
    <div class="stat"><div class="label">Tools</div><div class="value" id="sTools">-</div></div>
  </div>

  <!-- Chat Panel -->
  <div class="chat panel show" id="chatView">
    <div class="messages" id="messages">
      <div class="msg bot"><div class="meta">Aigo</div><div class="text">Hello! I am Aigo. How can I help?</div></div>
    </div>
    <div class="typing" id="typing">
      <div class="label">Aigo</div>
      <div class="dots"><div class="dot"></div><div class="dot"></div><div class="dot"></div></div>
    </div>
    <div class="input-bar">
      <input type="text" id="msgInput" placeholder="Type a message..." autocomplete="off"/>
      <button id="sendBtn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 2L11 13"/><path d="M22 2L15 22L11 13L2 9L22 2Z"/></svg></button>
    </div>
  </div>

  <!-- Skills Panel -->
  <div class="panel" id="skillsView">
    <div class="skills-header">
      <h2>Skill Marketplace</h2>
      <div class="skills-actions">
        <button id="syncBtn">Sync</button>
        <span id="skillStats" class="skill-stat-badge"></span>
      </div>
    </div>
    <div class="skills-search">
      <input type="text" id="skillSearchInput" placeholder="Search skills..." autocomplete="off"/>
      <button id="skillSearchBtn">Search</button>
    </div>
    <div class="skills-tabs" id="skillsTabs">
      <button class="active" data-stab="popular">Popular</button>
      <button data-stab="all">Browse</button>
      <button data-stab="installed">Installed</button>
      <button data-stab="sources">Sources</button>
    </div>
    <div class="skills-list" id="skillsList"></div>
    <div class="skill-detail" id="skillDetail">
      <button class="back-btn" id="skillBackBtn">Back</button>
      <div id="skillDetailContent"></div>
    </div>
  </div>

  <!-- Settings Panel -->
  <div class="settings" id="settingsView">
    <h2>Settings</h2>
    <div id="settingsForm"></div>
  </div>
</div>

<div class="toast" id="toast"></div>

<script>
(function(){
"use strict";

// --- Helpers ---
var $ = function(id){ return document.getElementById(id); };
var sidebarOpen = false;
var currentSkillTab = "popular";

function esc(s){ return (s||"").replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;"); }
function formatNum(n){ if(n>=1e6) return (n/1e6).toFixed(1)+"M"; if(n>=1e3) return (n/1e3).toFixed(1)+"K"; return String(n); }

function toast(msg, type){
  var t = $("toast");
  t.textContent = msg;
  t.className = "toast show " + (type||"");
  clearTimeout(t._timer);
  t._timer = setTimeout(function(){ t.className = "toast"; }, 3000);
}

function showTyping(show){ $("typing").classList.toggle("show", show); }
function scrollBottom(){ var m = $("messages"); m.scrollTop = m.scrollHeight; }

// --- Tab Navigation ---
$("nav").addEventListener("click", function(e){
  var btn = e.target.closest("button[data-tab]");
  if(!btn) return;
  var tab = btn.dataset.tab;
  document.querySelectorAll(".nav button").forEach(function(b){ b.classList.remove("active"); });
  btn.classList.add("active");

  $("chatView").classList.toggle("show", tab === "chat");
  $("skillsView").classList.toggle("show", tab === "skills");
  $("settingsView").classList.toggle("show", tab === "settings");

  if(tab === "stats"){
    sidebarOpen = !sidebarOpen;
    $("sidebar").classList.toggle("show", sidebarOpen);
    $("statsBtn").classList.toggle("active", sidebarOpen);
  }
  if(tab === "chat") $("msgInput").focus();
  if(tab === "skills"){ loadSkillStats(); loadSkills("popular"); }
  if(tab === "settings") loadSettings();
});

// --- Chat ---
function addMsg(cls, who, text){
  var d = document.createElement("div");
  d.className = "msg " + cls;
  var safe = esc(text).replace(/\n/g, "<br>");
  d.innerHTML = '<div class="meta">'+esc(who)+'</div><div class="text">'+safe+'</div><div class="actions"><button class="copy-btn">Copy</button></div>';
  var btn = d.querySelector(".copy-btn");
  var raw = text;
  btn.addEventListener("click", function(){
    navigator.clipboard.writeText(raw).then(function(){
      btn.textContent = "Copied!";
      setTimeout(function(){ btn.textContent = "Copy"; }, 2000);
    });
  });
  $("messages").appendChild(d);
  scrollBottom();
}

$("sendBtn").addEventListener("click", sendMsg);
$("msgInput").addEventListener("keydown", function(e){ if(e.key === "Enter") sendMsg(); });

function sendMsg(){
  var inp = $("msgInput"), text = inp.value.trim();
  if(!text) return;
  inp.value = "";
  $("sendBtn").disabled = true;
  addMsg("user", "You", text);
  showTyping(true);

  var d = document.createElement("div");
  d.className = "msg bot";
  d.innerHTML = '<div class="meta">Aigo</div><div class="text" id="streamText"><span class="cursor">...</span></div><div class="actions"><button class="copy-btn">Copy</button></div>';
  $("messages").appendChild(d);
  scrollBottom();
  var streamEl = $("streamText");
  var copyBtn = d.querySelector(".copy-btn");
  var fullText = "";

  fetch("/api/chat/stream", {
    method: "POST",
    headers: {"Content-Type":"application/json"},
    body: JSON.stringify({message: text})
  }).then(function(res){
    if(!res.ok){
      return fetch("/api/chat", {method:"POST", headers:{"Content-Type":"application/json"}, body:JSON.stringify({message:text})})
        .then(function(r){ return r.json(); })
        .then(function(d2){
          showTyping(false);
          streamEl.innerHTML = d2.ok ? esc(d2.response).replace(/\n/g,"<br>") : "Error: "+esc(d2.error||"unknown");
          fullText = d2.response || "";
        });
    }
    showTyping(false);
    var reader = res.body.getReader();
    var decoder = new TextDecoder();
    var buffer = "";
    function read(){
      reader.read().then(function(result){
        if(result.done){
          streamEl.innerHTML = esc(fullText).replace(/\n/g,"<br>");
          return;
        }
        buffer += decoder.decode(result.value, {stream:true});
        var lines = buffer.split("\n");
        buffer = lines.pop() || "";
        for(var i=0; i<lines.length; i++){
          var line = lines[i];
          if(line.indexOf("data: ") !== 0) continue;
          try {
            var data = JSON.parse(line.slice(6));
            if(data.chunk){ fullText += data.chunk; streamEl.innerHTML = esc(fullText).replace(/\n/g,"<br>") + '<span class="cursor">...</span>'; scrollBottom(); }
            if(data.done){ streamEl.innerHTML = esc(fullText).replace(/\n/g,"<br>"); }
            if(data.error){ streamEl.innerHTML = "Error: "+esc(data.error); }
          } catch(ex){}
        }
        read();
      });
    }
    read();
  }).catch(function(e){
    showTyping(false);
    streamEl.innerHTML = "Error: "+esc(e.message);
  }).finally(function(){
    $("sendBtn").disabled = false;
    copyBtn.addEventListener("click", function(){
      navigator.clipboard.writeText(fullText).then(function(){
        copyBtn.textContent = "Copied!";
        setTimeout(function(){ copyBtn.textContent = "Copy"; }, 2000);
      });
    });
    inp.focus();
  });
}

// --- Stats ---
function refreshStats(){
  fetch("/api/stats").then(function(r){ return r.json(); }).then(function(d){
    $("sUptime").textContent = d.uptime;
    $("sMsgs").textContent = d.messages;
    $("sChannels").textContent = (d.channels||[]).join(", ") || "None";
    $("sModel").textContent = d.provider + " / " + d.model;
    $("sTools").textContent = d.tools;
  }).catch(function(){});
}
setInterval(refreshStats, 5000);
refreshStats();

// --- Settings ---
var configFields = [
  {section: "AI Provider", icon: "AI", fields: [
    {key:"provider", label:"Provider", type:"select", options:["nous","openai","anthropic","ollama","openrouter","openai-compatible"], hint:"Choose your AI model provider"},
    {key:"api_key", label:"API Key", type:"password", hint:"Your provider API key (kept local, never shared)"},
    {key:"base_url", label:"Base URL", type:"text", hint:"API endpoint URL. Leave default for most providers", placeholder:"https://api.openai.com/v1"},
    {key:"model", label:"Model", type:"text", hint:"Model name/ID to use for chat", placeholder:"xiaomi/mimo-v2-pro"}
  ]},
  {section: "Channels", icon: "CH", fields: [
    {key:"telegram_token", label:"Telegram Bot Token", type:"password", hint:"Get from @BotFather on Telegram", placeholder:"123456:ABC-DEF..."},
    {key:"telegram_allowed_users", label:"Allowed Telegram User IDs", type:"text", hint:"Comma-separated user IDs. Leave empty to allow all", placeholder:"123456789"},
    {key:"discord_token", label:"Discord Bot Token", type:"password", hint:"From Discord Developer Portal"},
    {key:"slack_app_token", label:"Slack App Token", type:"password", hint:"xapp-... token from Slack"},
    {key:"slack_bot_token", label:"Slack Bot Token", type:"password", hint:"xoxb-... token from Slack"}
  ]},
  {section: "Memory", icon: "MM", fields: [
    {key:"memory_enabled", label:"Enable Memory", type:"select", options:["true","false"], hint:"Allow Aigo to remember conversations"},
    {key:"use_fts5", label:"Full-Text Search (FTS5)", type:"select", options:["true","false"], hint:"Use SQLite FTS5 for fast semantic search across memories"}
  ]},
  {section: "Behavior", icon: "BH", fields: [
    {key:"max_tokens", label:"Max Response Tokens", type:"number", hint:"Maximum tokens in AI responses. Higher = longer responses", placeholder:"4096"},
    {key:"temperature", label:"Temperature", type:"number", hint:"Creativity level (0.0 = precise, 1.0 = creative)", placeholder:"0.7"},
    {key:"system_prompt", label:"System Prompt", type:"textarea", hint:"Custom instructions for the AI. Defines personality and behavior", placeholder:"You are Aigo, a helpful AI assistant..."},
    {key:"auto_reply", label:"Auto-Reply on Channels", type:"select", options:["true","false"], hint:"Automatically respond to messages on connected channels"}
  ]}
];

function loadSettings(){
  fetch("/api/config").then(function(r){ return r.json(); }).then(function(cfg){
    renderSettings(cfg);
  }).catch(function(){
    $("settingsForm").innerHTML = '<div class="empty">Error loading config</div>';
  });
}

function renderSettings(cfg){
  var html = "";
  for(var s=0; s<configFields.length; s++){
    var section = configFields[s];
    html += '<div class="settings-section"><h3>'+section.icon+" "+esc(section.section)+'</h3>';
    for(var f=0; f<section.fields.length; f++){
      var field = section.fields[f];
      var val = cfg[field.key] !== undefined ? String(cfg[field.key]) : "";
      html += '<div class="field"><label>'+esc(field.label)+'</label>';
      if(field.type === "select"){
        html += '<select data-key="'+field.key+'">';
        for(var o=0; o<field.options.length; o++){
          var opt = field.options[o];
          html += '<option value="'+esc(opt)+'"'+(val===opt?" selected":"")+'>'+esc(opt)+'</option>';
        }
        html += '</select>';
      } else if(field.type === "textarea"){
        html += '<textarea data-key="'+field.key+'" placeholder="'+esc(field.placeholder||"")+'">'+esc(val)+'</textarea>';
      } else {
        html += '<input type="'+field.type+'" data-key="'+field.key+'" value="'+esc(val)+'" placeholder="'+esc(field.placeholder||"")+'"/>';
      }
      html += '<div class="hint">'+esc(field.hint)+'</div></div>';
    }
    html += '</div>';
  }
  html += '<div class="settings-actions"><button class="btn-save" id="saveConfigBtn">Save Settings</button><button class="btn-reset" id="resetConfigBtn">Reset</button></div>';
  // Also show raw JSON toggle
  html += '<details style="margin-top:1rem"><summary style="cursor:pointer;color:var(--muted);font-size:.8rem;margin-top:.5rem">Show raw JSON config</summary><pre style="background:var(--card);padding:.75rem;border-radius:8px;font-size:.75rem;color:var(--muted);margin-top:.5rem;overflow-x:auto;max-height:300px">'+esc(JSON.stringify(cfg,null,2))+'</pre></details>';
  $("settingsForm").innerHTML = html;

  $("saveConfigBtn").addEventListener("click", saveSettings);
  $("resetConfigBtn").addEventListener("click", function(){ loadSettings(); toast("Reset to current values","success"); });
}

function saveSettings(){
  var newConfig = {};
  var inputs = $("settingsForm").querySelectorAll("[data-key]");
  for(var i=0; i<inputs.length; i++){
    var el = inputs[i];
    var key = el.dataset.key;
    var val = el.value;
    // Type coercion
    if(val === "true") val = true;
    else if(val === "false") val = false;
    else if(val !== "" && !isNaN(val) && el.type === "number") val = Number(val);
    if(val !== "") newConfig[key] = val;
  }
  fetch("/api/config/save", {
    method: "POST",
    headers: {"Content-Type":"application/json"},
    body: JSON.stringify(newConfig)
  }).then(function(r){ return r.json(); }).then(function(d){
    if(d.ok) toast("Settings saved! Restart may be needed.", "success");
    else toast("Error: "+esc(d.error), "error");
  }).catch(function(e){ toast("Error: "+esc(e.message), "error"); });
}

// --- Skills ---
$("skillsTabs").addEventListener("click", function(e){
  var btn = e.target.closest("button[data-stab]");
  if(!btn) return;
  currentSkillTab = btn.dataset.stab;
  document.querySelectorAll("#skillsTabs button").forEach(function(b){ b.classList.remove("active"); });
  btn.classList.add("active");
  hideSkillDetail();
  loadSkills(currentSkillTab);
});

$("skillSearchBtn").addEventListener("click", searchSkills);
$("skillSearchInput").addEventListener("keydown", function(e){ if(e.key === "Enter") searchSkills(); });
$("syncBtn").addEventListener("click", syncSkills);
$("skillBackBtn").addEventListener("click", hideSkillDetail);

// Event delegation for skill cards
$("skillsList").addEventListener("click", function(e){
  var installBtn = e.target.closest(".btn-install");
  if(installBtn && !installBtn.classList.contains("done")){
    e.stopPropagation();
    installSkill(installBtn.dataset.id, installBtn);
    return;
  }
  var card = e.target.closest(".skill-card");
  if(card && card.dataset.name){
    viewSkill(card.dataset.name, card.dataset.source || "", card.dataset.id || "");
  }
});

function loadSkillStats(){
  fetch("/api/skills/stats").then(function(r){ return r.json(); }).then(function(d){
    if(d.ok) $("skillStats").textContent = (d.stats.total_indexed||0) + " indexed";
  }).catch(function(){});
}

function loadSkills(tab){
  var list = $("skillsList");
  list.innerHTML = '<div class="skeleton"></div><div class="skeleton"></div><div class="skeleton"></div>';
  var url;
  if(tab === "popular") url = "/api/skills/popular?limit=50";
  else if(tab === "all") url = "/api/skills/browse?limit=100";
  else if(tab === "installed") url = "/api/skills/list";
  else if(tab === "sources") url = "/api/skills/sources";
  else { list.innerHTML = '<div class="empty">Unknown tab</div>'; return; }

  fetch(url).then(function(r){ return r.json(); }).then(function(d){
    if(!d.ok){ list.innerHTML = '<div class="empty">Error: '+esc(d.error)+'</div>'; return; }
    if(tab === "sources"){
      renderSources(d.sources || [], list);
      return;
    }
    var skills = d.skills || [];
    if(skills.length === 0){ list.innerHTML = '<div class="empty">No skills found. Try syncing first.</div>'; return; }
    list.innerHTML = "";
    for(var i=0; i<skills.length; i++){
      list.appendChild(renderSkillCard(skills[i], tab !== "installed"));
    }
  }).catch(function(e){ list.innerHTML = '<div class="empty">Error: '+esc(e.message)+'</div>'; });
}

function renderSources(sources, list){
  list.innerHTML = "";
  for(var i=0; i<sources.length; i++){
    var s = sources[i];
    var d = document.createElement("div");
    d.className = "skill-card";
    var status = s.enabled ? "Active" : "Disabled";
    var sync = s.last_sync ? s.last_sync.substring(0,16) : "never";
    d.innerHTML = '<div class="name">'+status+" "+esc(s.name)+'</div>'+
      '<div class="desc">'+esc(s.url)+'</div>'+
      '<div class="meta"><span class="source">'+esc(s.type)+'</span><span>Synced: '+esc(sync)+'</span></div>';
    list.appendChild(d);
  }
  if(sources.length === 0) list.innerHTML = '<div class="empty">No sources configured</div>';
}

function renderSkillCard(s, showInstall){
  var card = document.createElement("div");
  card.className = "skill-card";
  card.dataset.name = s.name || "";
  card.dataset.source = s.source || "";
  card.dataset.id = s.identifier || "";

  var trustIcon = s.trust_level === "builtin" ? "Builtin" : s.trust_level === "trusted" ? "Trusted" : "Community";
  var installs = s.installs > 0 ? '<span class="installs">'+formatNum(s.installs)+' installs</span>' : "";
  var tags = "";
  if(s.tags && s.tags.length){
    var max = Math.min(s.tags.length, 3);
    for(var t=0; t<max; t++) tags += '<span class="tag">'+esc(s.tags[t])+'</span>';
  }

  var actionsHtml = "";
  if(showInstall){
    actionsHtml = '<div class="actions"><button class="btn-install" data-id="'+esc(s.identifier)+'">Install</button><button class="btn-detail">Details</button></div>';
  }

  card.innerHTML = '<div class="name"><span class="trust">'+trustIcon+'</span> '+esc(s.name)+'</div>'+
    '<div class="desc">'+esc((s.description||"").substring(0,180))+'</div>'+
    '<div class="meta"><span class="source">'+esc(s.source||"")+'</span>'+installs+tags+'</div>'+
    actionsHtml;
  return card;
}

function searchSkills(){
  var q = $("skillSearchInput").value.trim();
  if(!q) return;
  var list = $("skillsList");
  list.innerHTML = '<div class="skeleton"></div><div class="skeleton"></div>';
  fetch("/api/skills/search?q="+encodeURIComponent(q)).then(function(r){ return r.json(); }).then(function(d){
    if(!d.ok){ list.innerHTML = '<div class="empty">Error: '+esc(d.error)+'</div>'; return; }
    var skills = d.skills || [];
    if(skills.length === 0){ list.innerHTML = '<div class="empty">No skills found for "'+esc(q)+'"</div>'; return; }
    list.innerHTML = "";
    for(var i=0; i<skills.length; i++){
      list.appendChild(renderSkillCard(skills[i], true));
    }
  }).catch(function(e){ list.innerHTML = '<div class="empty">Error: '+esc(e.message)+'</div>'; });
}

function installSkill(id, btn){
  btn.disabled = true;
  btn.textContent = "Installing...";
  fetch("/api/skills/install", {
    method: "POST",
    headers: {"Content-Type":"application/json"},
    body: JSON.stringify({identifier: id})
  }).then(function(r){ return r.json(); }).then(function(d){
    if(d.ok){ btn.textContent = "Installed"; btn.classList.add("done"); toast("Installed: "+id, "success"); }
    else { btn.textContent = "Error"; btn.classList.add("err"); btn.disabled = false; toast(d.error, "error"); }
  }).catch(function(e){ btn.textContent = "Error"; btn.classList.add("err"); btn.disabled = false; toast(e.message, "error"); });
}

function syncSkills(){
  var btn = $("syncBtn");
  btn.disabled = true;
  btn.textContent = "Syncing...";
  fetch("/api/skills/sync", {method:"POST"}).then(function(r){ return r.json(); }).then(function(d){
    if(d.ok){ btn.textContent = d.total + " new"; toast("Synced "+d.total+" new skills", "success"); loadSkillStats(); loadSkills(currentSkillTab); }
    else { btn.textContent = "Error"; toast(d.error, "error"); }
  }).catch(function(e){ btn.textContent = "Error"; toast(e.message, "error"); });
  setTimeout(function(){ btn.textContent = "Sync"; btn.disabled = false; }, 4000);
}

function viewSkill(name, source, identifier){
  $("skillsList").style.display = "none";
  $("skillDetail").classList.add("show");
  $("skillDetailContent").innerHTML = '<div class="skeleton" style="height:200px"></div>';
  var url = "/api/skills/detail?name="+encodeURIComponent(name);
  if(identifier) url += "&id="+encodeURIComponent(identifier);
  fetch(url).then(function(r){ return r.json(); }).then(function(d){
    if(!d.ok){ $("skillDetailContent").innerHTML = '<div class="empty">Error: '+esc(d.error)+'</div>'; return; }
    var s = d.skill;
    var trustText = s.trust_level === "builtin" ? "Builtin" : s.trust_level === "trusted" ? "Trusted" : "Community";
    var installs = s.installs > 0 ? '<span class="chip">'+formatNum(s.installs)+' installs</span>' : "";
    var tagsHtml = "";
    if(s.tags && s.tags.length){
      tagsHtml = '<div class="detail-tags">';
      for(var t=0; t<s.tags.length; t++) tagsHtml += '<span class="tag">'+esc(s.tags[t])+'</span>';
      tagsHtml += '</div>';
    }
    var repoHtml = s.repo ? '<div class="detail-meta"><span class="chip">Repo: '+esc(s.repo)+'</span></div>' : "";
    var urlHtml = s.detail_url ? '<div><a href="'+esc(s.detail_url)+'" target="_blank" style="color:var(--cyan);font-size:.85rem">View Source</a></div>' : "";
    var installBtnHtml = "";
    if(!s.installed){
      installBtnHtml = '<button class="btn-install" data-id="'+esc(s.identifier)+'" style="padding:.5rem 1.2rem;font-size:.9rem">Install Skill</button>';
    } else {
      installBtnHtml = '<button class="btn-install done" disabled>Already Installed</button>';
    }
    $("skillDetailContent").innerHTML = '<div class="detail-name">'+esc(s.name)+'</div>'+
      '<div class="detail-desc">'+esc(s.description || "No description")+'</div>'+
      '<div class="detail-meta"><span class="chip">'+esc(trustText)+'</span><span class="chip">'+esc(s.source||"")+'</span>'+installs+'</div>'+
      repoHtml+urlHtml+tagsHtml+
      '<div class="detail-identifier">ID: <code>'+esc(s.identifier)+'</code></div>'+
      '<div style="margin-top:.5rem">'+installBtnHtml+'</div>';
  }).catch(function(e){ $("skillDetailContent").innerHTML = '<div class="empty">Error: '+esc(e.message)+'</div>'; });
}

function hideSkillDetail(){
  $("skillsList").style.display = "flex";
  $("skillDetail").classList.remove("show");
}

// --- Init ---
$("msgInput").focus();

})();
</script>
</body></html>
`

// --- Installer HTML ---

const installHTML = `<!DOCTYPE html>
<html><head><title>Aigo Setup</title>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=no">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#0f172a;color:#e2e8f0;min-height:100vh;padding:1.5rem}
.wizard{max-width:500px;margin:0 auto}
h1{font-size:1.5rem;margin-bottom:.5rem;color:#38bdf8}h1 span{color:#f472b6}
.sub{color:#94a3b8;margin-bottom:1.5rem;font-size:.9rem}
.step{background:#1e293b;border-radius:10px;padding:1.25rem;margin-bottom:.75rem}
.step h2{font-size:1rem;margin-bottom:.75rem;color:#f1f5f9}
label{display:block;margin:.4rem 0 .2rem;color:#94a3b8;font-size:.85rem}
input,select{width:100%;padding:.65rem;border:1px solid #334155;border-radius:8px;background:#0f172a;color:#e2e8f0;font-size:.95rem;margin-bottom:.75rem;-webkit-appearance:none}
input:focus,select:focus{outline:none;border-color:#38bdf8}
button{background:linear-gradient(135deg,#38bdf8,#818cf8);color:#fff;border:none;padding:.75rem;border-radius:8px;font-size:1rem;cursor:pointer;width:100%;margin-top:.5rem}
.status{padding:.75rem;border-radius:8px;margin-top:.75rem;font-size:.9rem}
.ok{background:#064e3b;color:#6ee7b7}.err{background:#7f1d1d;color:#fca5a5}
</style></head><body><div class="wizard">
<h1>🦞 <span>Aigo</span> Setup</h1>
<p class="sub">Configure your AI agent</p>
<form id="f">
<div class="step"><h2>1. Provider</h2>
<label>Provider</label><select name="provider"><option value="nous">Nous Research</option><option value="openai">OpenAI</option><option value="anthropic">Anthropic</option><option value="ollama">Ollama (Local)</option></select>
<label>API Key</label><input type="password" name="api_key" placeholder="sk-..."/>
<label>Base URL</label><input type="text" name="base_url" placeholder="https://api.openai.com/v1"/>
<label>Model</label><input type="text" name="model" placeholder="xiaomi/mimo-v2-pro" value="xiaomi/mimo-v2-pro"/></div>
<div class="step"><h2>2. Channels</h2>
<label>Telegram Token</label><input type="password" name="telegram_token" placeholder="123456:ABC..."/>
<label>Discord Token</label><input type="password" name="discord_token" placeholder="MTIz..."/>
<label>Slack App Token</label><input type="password" name="slack_app_token" placeholder="xapp-..."/>
<label>Slack Bot Token</label><input type="password" name="slack_bot_token" placeholder="xoxb-..."/></div>
<div class="step"><h2>3. Memory</h2>
<label>FTS5 Search</label><select name="use_fts5"><option value="true">Enabled (SQLite)</option><option value="false">Disabled (Flat Files)</option></select></div>
<button type="submit">Start Aigo</button></form>
<div id="s"></div></div>
<script>
document.getElementById('f').addEventListener('submit',async e=>{e.preventDefault();
const f=new FormData(e.target),cfg=Object.fromEntries(f.entries());
cfg.use_fts5=cfg.use_fts5==='true';cfg.memory_enabled=true;
const s=document.getElementById('s');s.innerHTML='<div class="status">Saving...</div>';
try{const r=await fetch('/install/api',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(cfg)});
const d=await r.json();s.innerHTML=d.ok?'<div class="status ok">Saved! Redirecting...</div>':'<div class="status err">'+d.error+'</div>';
if(d.ok)setTimeout(()=>location.href='/',2000)}catch(e){s.innerHTML='<div class="status err">'+e.message+'</div>'}});
</script></body></html>`
