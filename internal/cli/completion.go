package cli

import (
	"fmt"
	"runtime"
)

func GenerateShellCompletion(shell string) string {
	switch shell {
	case "bash":
		return bashCompletion()
	case "zsh":
		return zshCompletion()
	case "fish":
		return fishCompletion()
	default:
		return fmt.Sprintf("unsupported shell: %s", shell)
	}
}

func bashCompletion() string {
	return `_aigo() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="run doctor completion version help"

    case "${prev}" in
        doctor)
            COMPREPLY=()
            return 0
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "${cur}"))
            return 0
            ;;
    esac

    COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
    return 0
}
complete -F _aigo aigo
`
}

func zshCompletion() string {
	return `#compdef aigo

_aigo() {
    local -a commands
    commands=(
        'run:Start Aigo agent'
        'doctor:Run diagnostics'
        'completion:Generate shell completion'
        'version:Show version'
        'help:Show help'
    )

    _describe 'command' commands
}

_aigo "$@"
`
}

func fishCompletion() string {
	return `complete -c aigo -f
complete -c aigo -n "__fish_use_subcommand" -a run -d "Start Aigo agent"
complete -c aigo -n "__fish_use_subcommand" -a doctor -d "Run diagnostics"
complete -c aigo -n "__fish_use_subcommand" -a completion -d "Generate shell completion"
complete -c aigo -n "__fish_use_subcommand" -a version -d "Show version"
complete -c aigo -n "__fish_use_subcommand" -a help -d "Show help"
complete -c aigo -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
`
}

func GenerateExampleConfig() string {
	return `# Aigo Configuration
# Save as ~/.aigo/config.yaml

model:
  default: "opencode/qwen3.6-plus-free"
  coding: "auto"
  intent: "gpt-4o-mini"

opencode:
  binary: ""
  timeout: 300
  max_turns: 50

gateway:
  enabled: false
  platforms: []
  telegram:
    token: ""
  discord:
    token: ""

memory:
  max_l0_items: 20
  max_l1_items: 50
  auto_compress: true
  token_budget: 8000
  smart_prune: true

web:
  enabled: false
  port: ":8080"
  auth:
    enabled: false
    username: ""
    password: ""
`
}

func PrintDoctorResults(results []DoctorResult) {
	for _, r := range results {
		icon := "✓"
		switch r.Status {
		case "warn":
			icon = "⚠"
		case "info":
			icon = "ℹ"
		}
		fmt.Printf("  %s %-12s %s\n", icon, r.Name, r.Detail)
	}
}

func PrintVersion() {
	fmt.Printf("aigo version 0.1.0 (%s/%s)\n", runtime.GOOS, runtime.GOARCH)
}
