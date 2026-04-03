package hooks

import (
	"fmt"
	"testing"
)

func TestHookRegistry_RegisterAndFire(t *testing.T) {
	t.Parallel()

	r := NewHookRegistry()
	var called bool
	r.Register(HookSessionStart, HookFunc(func(event HookEvent) error {
		called = true
		return nil
	}))

	r.Fire(HookSessionStart, map[string]string{"session": "1"})
	if !called {
		t.Error("handler not called")
	}
}

func TestHookRegistry_FireNoHandlers(t *testing.T) {
	t.Parallel()

	r := NewHookRegistry()
	errs := r.Fire(HookSessionStart, nil)
	if errs != nil {
		t.Errorf("expected nil, got %v", errs)
	}
}

func TestHookRegistry_FireError(t *testing.T) {
	t.Parallel()

	r := NewHookRegistry()
	r.Register(HookSessionStart, HookFunc(func(event HookEvent) error {
		return fmt.Errorf("handler error")
	}))

	errs := r.Fire(HookSessionStart, nil)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestHookRegistry_ListHooks(t *testing.T) {
	t.Parallel()

	r := NewHookRegistry()
	r.Register(HookSessionStart, HookFunc(func(event HookEvent) error { return nil }))
	r.Register(HookSessionStart, HookFunc(func(event HookEvent) error { return nil }))

	count := r.ListHooks(HookSessionStart)
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestHookRegistry_Clear(t *testing.T) {
	t.Parallel()

	r := NewHookRegistry()
	r.Register(HookSessionStart, HookFunc(func(event HookEvent) error { return nil }))
	r.Clear()
	if r.ListHooks(HookSessionStart) != 0 {
		t.Error("expected 0 hooks after clear")
	}
}
