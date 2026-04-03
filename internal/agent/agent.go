// Package agent contains the core agent loop, task router, and self-correcting logic.
package agent

import (
	"context"
	"fmt"

	aigoctx "github.com/ahmad-ubaidillah/aigo/internal/context"
	"github.com/ahmad-ubaidillah/aigo/internal/intent"
	"github.com/ahmad-ubaidillah/aigo/internal/memory"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

const maxConsecutiveErrors = 3

type Agent struct {
	classifier        *intent.Classifier
	router            *Router
	ctxEngine         *aigoctx.ContextEngine
	db                *memory.SessionDB
	cfg               types.Config
	sessionID         string
	consecutiveErrors int
	totalTurns        int
	progress          *types.ProgressState
}

func NewAgent(
	classifier *intent.Classifier,
	router *Router,
	ctxEngine *aigoctx.ContextEngine,
	db *memory.SessionDB,
	cfg types.Config,
	sessionID string,
) *Agent {
	return &Agent{
		classifier: classifier,
		router:     router,
		ctxEngine:  ctxEngine,
		db:         db,
		cfg:        cfg,
		sessionID:  sessionID,
		progress:   types.NewProgressState(sessionID),
	}
}

func (a *Agent) RunSession(ctx context.Context, sessionID, task string) (*types.ToolResult, error) {
	a.totalTurns++
	a.sessionID = sessionID

	a.progress.Task = task
	a.progress.Thinking("Analyzing task...")

	if a.totalTurns > a.cfg.OpenCode.MaxTurns {
		a.progress.Fail("Max turns reached")
		return nil, fmt.Errorf("max turns reached: %d", a.cfg.OpenCode.MaxTurns)
	}

	if err := a.ensureSession(sessionID); err != nil {
		a.progress.Fail("Session error")
		return nil, fmt.Errorf("ensure session: %w", err)
	}

	if err := a.db.AddMessage(sessionID, "user", task); err != nil {
		a.progress.Fail("Message error")
		return nil, fmt.Errorf("add user message: %w", err)
	}

	classification := a.classifier.Classify(task)
	a.ctxEngine.SetTaskGoal(task)
	a.ctxEngine.IncrementTurns()

	prompt := a.ctxEngine.BuildPrompt(task)
	if err := a.recordIntent(sessionID, fmt.Sprintf("%s | %s", classification.Intent, prompt)); err != nil {
		return nil, fmt.Errorf("record intent: %w", err)
	}

	result, err := a.executeTask(ctx, classification, task)
	if err != nil {
		a.consecutiveErrors++
		a.ctxEngine.RecordError(err.Error())
		if a.consecutiveErrors >= maxConsecutiveErrors {
			a.progress.Fail("Max consecutive errors")
			return &types.ToolResult{
				Success: false,
				Output:  "Escalation: " + err.Error(),
				Error:   "max consecutive errors reached, manual intervention required",
			}, nil
		}
		a.progress.Fail(err.Error())
		return nil, fmt.Errorf("execute task: %w", err)
	}

	a.consecutiveErrors = 0
	a.ctxEngine.AddL0(fmt.Sprintf("Turn %d: %s -> %s", a.totalTurns, classification.Intent, truncate(result.Output, 100)))

	if respErr := a.db.AddMessage(sessionID, "assistant", result.Output); respErr != nil {
		return result, fmt.Errorf("add assistant message: %w", respErr)
	}

	if updateErr := a.db.UpdateSessionActivity(sessionID); updateErr != nil {
		return result, fmt.Errorf("update session activity: %w", updateErr)
	}

	a.progress.Complete()
	return result, nil
}

func (a *Agent) GetProgress() *types.ProgressState {
	return a.progress
}

func (a *Agent) GetContextSummary() string {
	if a.ctxEngine != nil {
		return a.ctxEngine.GetContextSummary()
	}
	return ""
}

func (a *Agent) ensureSession(sessionID string) error {
	_, err := a.db.GetSession(sessionID)
	if err == nil {
		return nil
	}

	_, err = a.db.CreateSession(sessionID, "run", "")
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

func (a *Agent) recordIntent(sessionID, intent string) error {
	return a.db.AddMessage(sessionID, "system", fmt.Sprintf("Intent: %s", intent))
}

func (a *Agent) executeTask(
	ctx context.Context,
	classification intent.Classification,
	task string,
) (*types.ToolResult, error) {
	class := Classification{
		Intent:     classification.Intent,
		Confidence: classification.Confidence,
		Workspace:  classification.Workspace,
		SessionID:  a.sessionID,
	}

	return a.router.Route(class, task)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
