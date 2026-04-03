# Hooks System

## Overview

Hooks allow you to run custom logic at specific lifecycle events in Aigo.

## Hook Types

| Hook | Trigger |
|---|---|
| `gateway:startup` | Gateway starts |
| `session:start` | New session begins |
| `session:end` | Session completes |
| `agent:start` | Agent begins task |
| `agent:step` | After each tool use |
| `agent:end` | Agent finishes |

## Hook Structure

```
~/.aigo/hooks/
├── HOOK.yaml           # Hook configuration
└── handler.go          # Hook handler (optional)
```

## HOOK.yaml Schema

```yaml
name: my-hook
type: session:start
enabled: true
command: echo "Session started"
```

## Programmatic Usage

```go
registry := hooks.NewHookRegistry()
registry.Register(hooks.HookSessionStart, hooks.HookFunc(func(event hooks.HookEvent) error {
    fmt.Println("Session started:", event.Payload["session_id"])
    return nil
}))
registry.Fire(hooks.HookSessionStart, map[string]string{"session_id": "abc123"})
```
