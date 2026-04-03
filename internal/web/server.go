package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/ahmad-ubaidillah/aigo/internal/memory"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	addr    string
	db      *memory.SessionDB
	cfg     types.Config
	mux     *http.ServeMux
	embedFS fs.FS
}

func NewServer(addr string, db *memory.SessionDB, cfg types.Config) *Server {
	s := &Server{
		addr: addr,
		db:   db,
		cfg:  cfg,
		mux:  http.NewServeMux(),
	}
	s.embedFS, _ = fs.Sub(staticFiles, "static")
	s.setupRoutes()
	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.addr, s.mux)
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/sessions", s.handleSessions)
	s.mux.HandleFunc("POST /api/sessions", s.handleSessions)
	s.mux.HandleFunc("GET /api/sessions/{id}", s.handleSession)
	s.mux.HandleFunc("GET /api/tasks", s.handleTasks)
	s.mux.HandleFunc("POST /api/tasks", s.handleTasks)
	s.mux.HandleFunc("POST /api/tasks/{id}/execute", s.handleTask)
	s.mux.HandleFunc("GET /api/agents", s.handleAgents)
	s.mux.HandleFunc("GET /api/gateways", s.handleGateways)
	s.mux.HandleFunc("POST /api/gateways", s.handleGateways)
	s.mux.HandleFunc("GET /api/memory", s.handleMemory)
	s.mux.HandleFunc("POST /api/memory", s.handleMemory)
	s.mux.HandleFunc("DELETE /api/memory/{id}", s.handleMemory)
	s.mux.HandleFunc("GET /api/stats", s.handleStats)
	s.mux.HandleFunc("GET /api/events", s.handleEvents)
	s.mux.HandleFunc("GET /api/events/ws", s.handleEventsWS)
	s.mux.HandleFunc("GET /api/config", s.handleConfig)
	s.mux.HandleFunc("PUT /api/config", s.handleConfigUpdate)
	s.mux.HandleFunc("GET /api/skills", s.handleSkills)

	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(s.embedFS))))

	s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, staticFiles, "static/index.html")
	})
}
