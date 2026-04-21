package memory

import (
	"strconv"
	"strings"
	"sync"
)

type Versioning struct {
	data map[string]struct {
		Value   string
		Version int
	}
}

func NewVersioning() *Versioning {
	return &Versioning{data: make(map[string]struct {
		Value   string
		Version int
	})}
}

func (v *Versioning) Set(key, value string, version int) {
	v.data[key] = struct {
		Value   string
		Version int
	}{Value: value, Version: version}
}

func (v *Versioning) Get(key string) (string, int) {
	if val, ok := v.data[key]; ok {
		return val.Value, val.Version
	}
	return "", 0
}

type RelationshipNode struct {
	ID     string
	Type   string
	Edges  map[string]string
}

type RelationshipGraph struct {
	nodes map[string]*RelationshipNode
	mu    sync.RWMutex
}

func NewRelationshipGraph() *RelationshipGraph {
	return &RelationshipGraph{nodes: make(map[string]*RelationshipNode)}
}

func (g *RelationshipGraph) AddEdge(from, to, relType string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.nodes[from] == nil {
		g.nodes[from] = &RelationshipNode{ID: from, Edges: make(map[string]string)}
	}
	g.nodes[from].Edges[to] = relType
}

func (g *RelationshipGraph) HasEdge(from, to string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.nodes[from] == nil {
		return false
	}
	_, ok := g.nodes[from].Edges[to]
	return ok
}

type Profile struct {
	Preferences []string
	Dynamic    map[string]string
}

type ProfileGenerator struct{}

func NewProfileGenerator() *ProfileGenerator {
	return &ProfileGenerator{}
}

func (p *ProfileGenerator) Generate(prefs []string) *Profile {
	return &Profile{
		Preferences: prefs,
		Dynamic:     make(map[string]string),
	}
}

type ContainerSystem struct {
	itemToContainer map[string]string
}

func NewContainerSystem() *ContainerSystem {
	return &ContainerSystem{itemToContainer: make(map[string]string)}
}

func (c *ContainerSystem) Add(item, container string) {
	c.itemToContainer[item] = container
}

func (c *ContainerSystem) Has(item string) bool {
	_, ok := c.itemToContainer[item]
	return ok
}

type HotnessTracker struct {
	counts map[string]int
	mu     sync.RWMutex
}

func NewHotnessTracker() *HotnessTracker {
	return &HotnessTracker{counts: make(map[string]int)}
}

func (h *HotnessTracker) Track(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.counts[key]++
}

func (h *HotnessTracker) Get(key string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.counts[key]
}

type ContextLevels struct {
	l0Tokens int
	l1Tokens int
	l2Tokens int
}

func NewContextLevels() *ContextLevels {
	return &ContextLevels{l0Tokens: 50, l1Tokens: 500, l2Tokens: 5000}
}

func (c *ContextLevels) Summarize(items []string) string {
	if len(items) <= 3 {
		return strings.Join(items, ", ")
	}
	return items[0] + " + " + strconv.Itoa(len(items)-1) + " more"
}

type FactExtractor struct{}

func NewFactExtractor() *FactExtractor {
	return &FactExtractor{}
}

func (f *FactExtractor) Extract(text string) []string {
	facts := make([]string, 0)
	indicators := []string{" at ", " use ", " via "}
	for _, ind := range indicators {
		if strings.Contains(text, ind) {
			facts = append(facts, ind)
		}
	}
	return facts
}

type WisdomStore struct {
	lessons []string
	mu      sync.RWMutex
}

func NewWisdomStore() *WisdomStore {
	return &WisdomStore{lessons: make([]string, 0)}
}

func (w *WisdomStore) AddLesson(lesson string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, l := range w.lessons {
		if l == lesson {
			return
		}
	}
	w.lessons = append(w.lessons, lesson)
}

func (w *WisdomStore) HasLesson(lesson string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, l := range w.lessons {
		if l == lesson {
			return true
		}
	}
	return false
}

func (w *WisdomStore) GetLessons() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	result := make([]string, len(w.lessons))
	copy(result, w.lessons)
	return result
}