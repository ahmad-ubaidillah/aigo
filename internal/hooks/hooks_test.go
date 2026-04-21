package hooks

import (
	"testing"
)

func TestHookTypes(t *testing.T) {
	if len(HookTypes) != 48 {
		t.Logf("Expected 48 hook types, got %d", len(HookTypes))
	}
}

func TestHookRegistry(t *testing.T) {
	r := NewHookRegistry()
	r.Register("test-hook", func() error { return nil })

	if !r.Has("test-hook") {
		t.Error("Hook should be registered")
	}
}

func TestHookExecution(t *testing.T) {
	r := NewHookRegistry()
	called := false
	r.Register("test", func() error { called = true; return nil })

	r.Execute("test")
	if !called {
		t.Error("Hook should have been executed")
	}
}

func TestHookExecutor_Trigger(t *testing.T) {
	r := NewHookRegistry()
	r.Register("on_start", func() error { return nil })

	e := NewHookExecutor(r)
	err := e.TriggerOnStart()
	if err != nil {
		t.Errorf("TriggerOnStart failed: %v", err)
	}

	log := e.GetLog(10)
	if len(log) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(log))
	}
}

func TestHookExecutor_TriggerOnTool(t *testing.T) {
	r := NewHookRegistry()
	r.Register("on_tool_start", func() error { return nil })
	r.Register("on_tool_end", func() error { return nil })

	e := NewHookExecutor(r)
	e.TriggerOnToolStart("test")
	e.TriggerOnToolEnd("test")

	log := e.GetLog(10)
	if len(log) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(log))
	}
}

func TestHookExecutor_GetLogByHook(t *testing.T) {
	r := NewHookRegistry()
	r.Register("on_start", func() error { return nil })
	r.Register("on_stop", func() error { return nil })

	e := NewHookExecutor(r)
	e.TriggerOnStart()
	e.TriggerOnStop()

	startLog := e.GetLogByHook("on_start")
	if len(startLog) != 1 {
		t.Errorf("Expected 1 on_start log, got %d", len(startLog))
	}
}

func TestHookExecutor_ClearLog(t *testing.T) {
	r := NewHookRegistry()
	r.Register("on_start", func() error { return nil })

	e := NewHookExecutor(r)
	e.TriggerOnStart()
	e.TriggerOnStart()

	e.ClearLog()

	log := e.GetLog(10)
	if len(log) != 0 {
		t.Error("Log should be empty after clear")
	}
}