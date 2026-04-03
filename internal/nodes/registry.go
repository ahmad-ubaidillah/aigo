package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Registry struct {
	mu    sync.RWMutex
	nodes map[string]Node
}

type Node interface {
	ID() string
	Type() string
	Name() string
	Description() string
	Execute(ctx context.Context, input interface{}) (interface{}, error)
	Validate() error
}

type BaseNode struct {
	id          string
	nodeType    string
	name        string
	description string
	timeout     time.Duration
}

func (n *BaseNode) ID() string             { return n.id }
func (n *BaseNode) Type() string           { return n.nodeType }
func (n *BaseNode) Name() string           { return n.name }
func (n *BaseNode) Description() string    { return n.description }
func (n *BaseNode) Timeout() time.Duration { return n.timeout }

type NodeConfig map[string]interface{}

func NewRegistry() *Registry {
	r := &Registry{
		nodes: make(map[string]Node),
	}
	r.registerBuiltins()
	return r
}

func (r *Registry) registerBuiltins() {
	r.Register(&HTTPNode{BaseNode: BaseNode{
		id: "http", nodeType: "http", name: "HTTP Request",
		description: "Make HTTP requests to APIs",
	}})
	r.Register(&BrowserNode{BaseNode: BaseNode{
		id: "browser", nodeType: "browser", name: "Browser",
		description: "Interact with web browser",
	}})
	r.Register(&CodeNode{BaseNode: BaseNode{
		id: "code", nodeType: "code", name: "Code Executor",
		description: "Execute code in Python or Go",
	}})
	r.Register(&SearchNode{BaseNode: BaseNode{
		id: "search", nodeType: "search", name: "Web Search",
		description: "Search the web for information",
	}})
	r.Register(&LLMNode{BaseNode: BaseNode{
		id: "llm", nodeType: "llm", name: "LLM Prompt",
		description: "Call language model for text generation",
	}})
	r.Register(&ConditionNode{BaseNode: BaseNode{
		id: "condition", nodeType: "condition", name: "Condition",
		description: "Branch workflow based on condition",
	}})
	r.Register(&LoopNode{BaseNode: BaseNode{
		id: "loop", nodeType: "loop", name: "Loop",
		description: "Iterate over items",
	}})
	r.Register(&WaitNode{BaseNode: BaseNode{
		id: "wait", nodeType: "wait", name: "Wait",
		description: "Wait for specified duration",
	}})
	r.Register(&TransformNode{BaseNode: BaseNode{
		id: "transform", nodeType: "transform", name: "Transform",
		description: "Transform data with custom logic",
	}})
}

func (r *Registry) Register(node Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node.ID() == "" {
		return fmt.Errorf("node id cannot be empty")
	}
	if _, exists := r.nodes[node.ID()]; exists {
		return fmt.Errorf("node already registered: %s", node.ID())
	}

	r.nodes[node.ID()] = node
	return nil
}

func (r *Registry) Get(id string) (Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, ok := r.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", id)
	}
	return node, nil
}

func (r *Registry) List() []Node {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		result = append(result, node)
	}
	return result
}

func (r *Registry) ListByType(nodeType string) []Node {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Node
	for _, node := range r.nodes {
		if node.Type() == nodeType {
			result = append(result, node)
		}
	}
	return result
}

func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[id]; !exists {
		return fmt.Errorf("node not found: %s", id)
	}
	delete(r.nodes, id)
	return nil
}

type HTTPNode struct {
	BaseNode
	config NodeConfig
}

func (n *HTTPNode) Config() NodeConfig     { return n.config }
func (n *HTTPNode) SetConfig(c NodeConfig) { n.config = c }

func (n *HTTPNode) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	cfg := n.config
	if cfg == nil {
		cfg = make(NodeConfig)
	}

	cfg["input"] = input
	return cfg, nil
}

func (n *HTTPNode) Validate() error {
	if n.config == nil {
		return fmt.Errorf("config required")
	}
	return nil
}

type BrowserNode struct {
	BaseNode
	config NodeConfig
}

func (n *BrowserNode) Config() NodeConfig     { return n.config }
func (n *BrowserNode) SetConfig(c NodeConfig) { n.config = c }

func (n *BrowserNode) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	cfg := n.config
	if cfg == nil {
		cfg = make(NodeConfig)
	}
	cfg["input"] = input
	return cfg, nil
}

func (n *BrowserNode) Validate() error { return nil }

type CodeNode struct {
	BaseNode
	config NodeConfig
}

func (n *CodeNode) Config() NodeConfig     { return n.config }
func (n *CodeNode) SetConfig(c NodeConfig) { n.config = c }

func (n *CodeNode) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	cfg := n.config
	if cfg == nil {
		cfg = make(NodeConfig)
	}
	cfg["input"] = input
	return cfg, nil
}

func (n *CodeNode) Validate() error { return nil }

type SearchNode struct {
	BaseNode
	config NodeConfig
}

func (n *SearchNode) Config() NodeConfig     { return n.config }
func (n *SearchNode) SetConfig(c NodeConfig) { n.config = c }

func (n *SearchNode) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	cfg := n.config
	if cfg == nil {
		cfg = make(NodeConfig)
	}
	cfg["input"] = input
	return cfg, nil
}

func (n *SearchNode) Validate() error { return nil }

type LLMNode struct {
	BaseNode
	config NodeConfig
}

func (n *LLMNode) Config() NodeConfig     { return n.config }
func (n *LLMNode) SetConfig(c NodeConfig) { n.config = c }

func (n *LLMNode) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	cfg := n.config
	if cfg == nil {
		cfg = make(NodeConfig)
	}
	cfg["input"] = input
	return cfg, nil
}

func (n *LLMNode) Validate() error { return nil }

type ConditionNode struct {
	BaseNode
	config    NodeConfig
	condition string
}

func (n *ConditionNode) Config() NodeConfig     { return n.config }
func (n *ConditionNode) SetConfig(c NodeConfig) { n.config = c }

func (n *ConditionNode) SetCondition(cond string) { n.condition = cond }

func (n *ConditionNode) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	cfg := n.config
	if cfg == nil {
		cfg = make(NodeConfig)
	}
	cfg["input"] = input
	cfg["condition"] = n.condition
	return cfg, nil
}

func (n *ConditionNode) Validate() error { return nil }

type LoopNode struct {
	BaseNode
	config NodeConfig
	items  []interface{}
}

func (n *LoopNode) Config() NodeConfig           { return n.config }
func (n *LoopNode) SetConfig(c NodeConfig)       { n.config = c }
func (n *LoopNode) SetItems(items []interface{}) { n.items = items }

func (n *LoopNode) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	cfg := n.config
	if cfg == nil {
		cfg = make(NodeConfig)
	}
	cfg["input"] = input
	cfg["items"] = n.items
	return cfg, nil
}

func (n *LoopNode) Validate() error { return nil }

type WaitNode struct {
	BaseNode
	duration time.Duration
}

func (n *WaitNode) Duration() time.Duration     { return n.duration }
func (n *WaitNode) SetDuration(d time.Duration) { n.duration = d }

func (n *WaitNode) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(n.duration):
		return input, nil
	}
}

func (n *WaitNode) Validate() error { return nil }

type TransformNode struct {
	BaseNode
	config    NodeConfig
	transform func(interface{}) (interface{}, error)
}

func (n *TransformNode) Config() NodeConfig                                     { return n.config }
func (n *TransformNode) SetConfig(c NodeConfig)                                 { n.config = c }
func (n *TransformNode) SetTransform(fn func(interface{}) (interface{}, error)) { n.transform = fn }

func (n *TransformNode) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	if n.transform == nil {
		return input, nil
	}
	return n.transform(input)
}

func (n *TransformNode) Validate() error { return nil }

type Workflow struct {
	id          string
	name        string
	description string
	nodes       []WorkflowNode
	edges       []WorkflowEdge
	mu          sync.RWMutex
	createdAt   time.Time
	updatedAt   time.Time
}

type WorkflowNode struct {
	ID       string     `json:"id"`
	Type     string     `json:"type"`
	Config   NodeConfig `json:"config"`
	Position Position   `json:"position"`
}

type WorkflowEdge struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition,omitempty"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func NewWorkflow(id, name, description string) *Workflow {
	return &Workflow{
		id:          id,
		name:        name,
		description: description,
		nodes:       make([]WorkflowNode, 0),
		edges:       make([]WorkflowEdge, 0),
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}
}

func (w *Workflow) AddNode(node WorkflowNode) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.nodes = append(w.nodes, node)
	w.updatedAt = time.Now()
}

func (w *Workflow) AddEdge(edge WorkflowEdge) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.edges = append(w.edges, edge)
	w.updatedAt = time.Now()
}

func (w *Workflow) Execute(ctx context.Context, initialInput interface{}) (interface{}, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	state := map[string]interface{}{
		"_start": initialInput,
	}

	for _, edge := range w.edges {
		if edge.Condition != "" {
			continue
		}

		var input interface{}
		if edge.From == "_start" {
			input = initialInput
		} else {
			input = state[edge.From]
		}

		result := input
		state[edge.To] = result
	}

	return state["_end"], nil
}

func (w *Workflow) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID          string         `json:"id"`
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Nodes       []WorkflowNode `json:"nodes"`
		Edges       []WorkflowEdge `json:"edges"`
		CreatedAt   time.Time      `json:"created_at"`
		UpdatedAt   time.Time      `json:"updated_at"`
	}{
		ID:          w.id,
		Name:        w.name,
		Description: w.description,
		Nodes:       w.nodes,
		Edges:       w.edges,
		CreatedAt:   w.createdAt,
		UpdatedAt:   w.updatedAt,
	})
}

type WorkflowExecutor struct {
	registry *Registry
	mu       sync.RWMutex
}

func NewWorkflowExecutor(reg *Registry) *WorkflowExecutor {
	return &WorkflowExecutor{registry: reg}
}

func (e *WorkflowExecutor) ExecuteNode(ctx context.Context, nodeID string, input interface{}) (interface{}, error) {
	node, err := e.registry.Get(nodeID)
	if err != nil {
		return nil, err
	}
	return node.Execute(ctx, input)
}

func (e *WorkflowExecutor) ValidateWorkflow(wf *Workflow) error {
	nodeIDs := make(map[string]bool)
	for _, node := range wf.nodes {
		nodeIDs[node.ID] = true
		_, err := e.registry.Get(node.Type)
		if err != nil {
			return fmt.Errorf("node type not found: %s", node.Type)
		}
	}

	for _, edge := range wf.edges {
		if edge.From != "_start" && !nodeIDs[edge.From] {
			return fmt.Errorf("edge references non-existent node: %s", edge.From)
		}
		if !nodeIDs[edge.To] {
			return fmt.Errorf("edge references non-existent node: %s", edge.To)
		}
	}
	return nil
}
