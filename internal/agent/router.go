package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/ahmad-ubaidillah/aigo/internal/handlers"
	"github.com/ahmad-ubaidillah/aigo/internal/memory"
	"github.com/ahmad-ubaidillah/aigo/internal/opencode"
	"github.com/ahmad-ubaidillah/aigo/internal/research"
	"github.com/ahmad-ubaidillah/aigo/internal/skills"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type Classification struct {
	Intent     string
	Confidence float64
	Workspace  string
	SessionID  string
}

type Handler interface {
	CanHandle(intent string) bool
	Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error)
}

type Router struct {
	handlers     map[string]Handler
	db           *memory.SessionDB
	cfg          types.Config
	ocClient     *opencode.Client
	registry     *skills.Registry
	searchClient *research.SearchClient
}

func NewRouter(db *memory.SessionDB, cfg types.Config, ocClient *opencode.Client) *Router {
	r := &Router{
		handlers:     make(map[string]Handler),
		db:           db,
		cfg:          cfg,
		ocClient:     ocClient,
		registry:     skills.NewRegistry(),
		searchClient: research.NewSearchClient(),
	}

	r.registerDefaults()
	return r
}

func (r *Router) RegisterHandler(intent string, h Handler) {
	r.handlers[intent] = h
}

func (r *Router) Route(classification Classification, taskDesc string) (*types.ToolResult, error) {
	handler, ok := r.handlers[classification.Intent]
	if !ok {
		return nil, fmt.Errorf("no handler for intent: %s", classification.Intent)
	}

	task := &types.Task{
		SessionID:   classification.SessionID,
		Description: taskDesc,
		Status:      types.TaskPending,
		Priority:    types.PriorityMedium,
		Workspace:   classification.Workspace,
		CreatedAt:   time.Now(),
	}

	return handler.Execute(context.Background(), task, classification.Workspace)
}

func (r *Router) registerDefaults() {
	r.RegisterHandler(types.IntentCoding, &codingHandler{client: r.ocClient})
	r.RegisterHandler(types.IntentWeb, &handlers.WebHandler{})
	r.RegisterHandler(types.IntentFile, &handlers.FileHandler{})
	r.RegisterHandler(types.IntentGateway, &handlers.GatewayHandler{})
	r.RegisterHandler(types.IntentMemory, &handlers.GatewayHandler{})
	r.RegisterHandler(types.IntentAutomation, &handlers.AutomationHandler{})
	r.RegisterHandler(types.IntentGeneral, &handlers.GeneralHandler{})
	r.RegisterHandler(types.IntentSkill, handlers.NewSkillHandler(r.registry))
	r.RegisterHandler(types.IntentResearch, handlers.NewResearchHandler(r.searchClient))
	r.RegisterHandler(types.IntentHTTPCall, handlers.NewHTTPHandler())
	r.RegisterHandler(types.IntentBrowser, handlers.NewBrowserHandler())
	r.RegisterHandler(types.IntentPython, handlers.NewPythonHandler())
}

type codingHandler struct {
	client *opencode.Client
}

func (h *codingHandler) CanHandle(intent string) bool {
	return intent == types.IntentCoding
}

func (h *codingHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	if h.client == nil {
		return &types.ToolResult{
			Success: false,
			Error:   "OpenCode client not configured",
		}, nil
	}
	return h.client.Run(ctx, task.Description, task.SessionID)
}
