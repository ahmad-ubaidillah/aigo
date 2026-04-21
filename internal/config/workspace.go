package config

import (
	"encoding/json"
	"sync"
)

type Workspace struct {
	Name     string
	Config   map[string]string
	DataPath string
}

type WorkspaceConfig struct {
	data map[string]string
	mu   sync.RWMutex
}

type WorkspaceSwitcher struct {
	current  *Workspace
	wsList   map[string]*Workspace
	mu       sync.RWMutex
}

type ExportImport struct{}

func NewWorkspace(name string) *Workspace {
	return &Workspace{
		Name:     name,
		Config:   make(map[string]string),
		DataPath:  "",
	}
}

func NewWorkspaceConfig() *WorkspaceConfig {
	return &WorkspaceConfig{data: make(map[string]string)}
}

func (wc *WorkspaceConfig) Set(key, value string) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.data[key] = value
}

func (wc *WorkspaceConfig) Get(key string) string {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	return wc.data[key]
}

func NewWorkspaceSwitcher() *WorkspaceSwitcher {
	return &WorkspaceSwitcher{wsList: make(map[string]*Workspace)}
}

func (ws *WorkspaceSwitcher) Add(w *Workspace) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.wsList[w.Name] = w
}

func (ws *WorkspaceSwitcher) Switch(name string) *Workspace {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if w, ok := ws.wsList[name]; ok {
		ws.current = w
		return w
	}
	return nil
}

func (ws *WorkspaceSwitcher) Current() *Workspace {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.current
}

func NewExportImport() *ExportImport {
	return &ExportImport{}
}

func (ei *ExportImport) Export(profile *Profile) string {
	data, _ := json.Marshal(profile)
	return string(data)
}

func (ei *ExportImport) Import(data string) *Profile {
	var profile Profile
	json.Unmarshal([]byte(data), &profile)
	return &profile
}

type ConfigHotReload struct {
	watcher chan bool
}

func NewConfigHotReload() *ConfigHotReload {
	return &ConfigHotReload{make(chan bool)}
}

func (chr *ConfigHotReload) Reload() bool {
	select {
	case chr.watcher <- true:
		return true
	default:
		return false
	}
}