# Skills Development Guide

## What is a Skill?

A skill is a directory under `~/.aigo/skills/` that contains instructions and resources for the agent to perform specialized tasks.

## Structure

```
~/.aigo/skills/
└── my-skill/
    ├── SKILL.md          # Main instructions (required)
    ├── references/       # Supporting documentation
    ├── templates/        # Output templates
    └── scripts/          # Executable scripts
```

## SKILL.md Format

```markdown
# Skill Name

Brief description of what this skill does.

category: coding
version: 1.0

## Instructions

Step-by-step instructions for the agent.

## Examples

Example inputs and expected outputs.
```

## Adding a Skill

1. Create directory: `mkdir -p ~/.aigo/skills/my-skill`
2. Write SKILL.md with instructions
3. The agent auto-discovers skills on startup

## CLI Commands

- `aigo skill list` — List all skills
- `aigo skill view <name>` — View skill details
- `aigo skill create <name>` — Create a new skill template
- `aigo skill run <name>` — Execute a skill
