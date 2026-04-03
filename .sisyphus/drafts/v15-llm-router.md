# Aigo V1.5 — LLM Router, Auto-Install, Token Budget, Progressive Enhancement

## Context
- **Goal**: Make Aigo "never die" — orchestration never stops even if one provider fails
- **Current State**: 5 LLM clients exist but disconnected, OpenCode = CLI stub, installer = stub, token budget = scattered
- **Target**: LLM Router with fallback, Auto-Install OpenCode, Token Budget Manager, Progressive Enhancement chain

## Key Findings

### LLM Providers (internal/llm/)
- ✅ OpenAIClient, AnthropicClient, OpenRouterClient, GLMClient, LocalClient — ALL exist
- ❌ NO factory function — manual instantiation per client
- ❌ NO fallback/rotation logic
- ❌ Config only supports SINGLE provider (LLMConfig.Provider string)

### OpenCode Integration
- ✅ CLI wrapper exists (internal/opencode/client.go) with HealthCheck()
- ❌ Installer.CheckOpenCode() = stub (returns true, nil)
- ❌ Installer.InstallOpenCode() = stub (returns nil)
- ❌ No preflight health check before delegation

### Token Tracking
- ✅ ContextEngine tracks tokenCount, has TokenBudget (default 8000)
- ✅ AgentLoop has tokenBudget (default 100000)
- ✅ Distiller + ToonFormatter exist for compression
- ❌ NO centralized budget manager
- ❌ NO alerting when budget approaches limit

### Setup Wizard
- ✅ Asks CLI/Web mode + API key
- ❌ NO provider selection (only one API key)
- ❌ NO OpenCode installation
- ❌ NO local LLM detection

## Decisions Needed
1. Provider priority order for fallback chain
2. OpenCode install method (download binary vs build from source)
3. Token budget alerting mechanism
4. Multi-provider config structure

## Architecture Vision
```
Aigo (Otak) → LLM Router → [OpenAI → Anthropic → OpenRouter → Ollama]
              ↓
         Atlas Orchestrator
              ↓
    [OpenCode Go (TANGAN) OR Native Tools fallback]
```
