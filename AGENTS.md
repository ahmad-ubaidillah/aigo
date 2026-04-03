# Aigo — Coding Standards

> Generated from `rule.json`. All 30 rules enforced. Every line of code must follow these.

## Core Principles

### 1. KISS — Keep It Simple, Stupid
Avoid unnecessary complexity. Choose the simplest solution that works.

### 2. DRY — Don't Repeat Yourself
Avoid code duplication by creating reusable functions or modules.

### 3. YAGNI — You Aren't Gonna Need It
Do not implement features until they are actually needed.

---

## Architecture & Design

### 4. Separation of Concerns
Separate responsibilities across layers (UI, business logic, data access).

### 5. Single Responsibility Principle
A class or function should have only one reason to change.

### 6. Open/Closed Principle
Code should be open for extension but closed for modification.

### 7. Liskov Substitution Principle
Subclasses must be replaceable for their base classes without breaking behavior.

### 8. Interface Segregation Principle
Use small, specific interfaces instead of large general ones.

### 9. Dependency Inversion Principle
Depend on abstractions, not on concrete implementations.

### 10. Encapsulation
Hide internal implementation details and expose only what is necessary.

### 11. Abstraction
Focus on what a component does, not how it does it.

### 12. Law of Demeter
Objects should only interact with their immediate dependencies.

### 13. Composition over Inheritance
Prefer composition over inheritance for better flexibility.

---

## Data & State

### 14. Immutability
Avoid mutating state directly; prefer immutable data structures.

---

## Code Structure

### 15. Small Functions
Functions should be small, focused, and easy to understand.

### 16. Single Level of Abstraction
Maintain a consistent level of abstraction within a function.

### 17. Shallow Nesting
Avoid deep nesting. Use early returns or function extraction to simplify control flow.

### 18. Low Complexity
Keep cyclomatic and cognitive complexity low by simplifying logic and reducing nesting.

### 19. No God Object
Avoid classes or modules that handle too many responsibilities.

---

## Naming & Style

### 20. Meaningful Naming
Use clear, descriptive, and intention-revealing names for variables, functions, and classes.

### 21. Naming Consistency
Follow consistent naming conventions (camelCase, PascalCase, etc.) across the codebase.

### 22. Max Line Length
Each line of code must not exceed 150 characters. Split long lines into multiple lines for readability.

### 23. File Size Limit
Avoid excessively large files. Split files that become too long into smaller modules.

### 24. No Magic Numbers
Replace hard-coded numbers with named constants.

### 25. Consistency
Maintain consistent style, structure, and patterns throughout the project.

---

## Quality & Safety

### 26. Error Handling
Handle errors explicitly and provide meaningful error messages.

### 27. Readability Over Cleverness
Prioritize readability over clever or overly concise code.

### 28. Testing Friendly
Write code that is easy to test (low coupling, dependency injection, pure functions).

### 29. Limit Parameters
Limit the number of function parameters. Use objects if there are too many.

### 30. Avoid Side Effects
Minimize hidden side effects within functions.

### 31. Convention Over Configuration
Follow established conventions instead of creating custom configurations.

---

## Go-Specific Conventions

### Package Structure
```
cmd/aigo/          — entry point
internal/          — private packages (not importable externally)
  agent/           — agent core, loop, router
  context/         — L0/L1/L2 context engine
  gateway/         — messaging platform adapters
  handlers/        — native task handlers
  intent/          — intent classification
  memory/          — SQLite session + memory
  opencode/        — OpenCode delegation
  tui/             — terminal UI
  web/             — web GUI server
  cli/             — cobra commands
pkg/types/         — shared types (importable)
```

### Naming Conventions
- **Packages**: lowercase, single word (`agent`, `gateway`, `memory`)
- **Types**: PascalCase (`SessionDB`, `ContextEngine`, `IntentGate`)
- **Interfaces**: `-er` suffix or descriptive (`Handler`, `Platform`, `Classifier`)
- **Functions**: PascalCase (exported), camelCase (unexported)
- **Variables**: camelCase, descriptive (`sessionID`, `maxTurns`, `consecutiveErrors`)
- **Constants**: PascalCase or UPPER_SNAKE_CASE for config values
- **Errors**: prefix with `Err` (`ErrSessionNotFound`, `ErrMaxTurnsReached`)

### Error Handling
```go
// GOOD: explicit error handling
result, err := db.GetSession(id)
if err != nil {
    return fmt.Errorf("get session %s: %w", id, err)
}

// BAD: ignored errors
result, _ := db.GetSession(id)
```

### Interface Design
```go
// GOOD: small, specific interfaces
type Handler interface {
    Execute(ctx context.Context, task *Task) (*Result, error)
}

// BAD: god interface with too many methods
type Manager interface {
    Execute()
    Validate()
    Format()
    Save()
    Load()
    Delete()
    // ... 20 more methods
}
```

### Function Size
- Max 50 lines per function (soft limit)
- Max 3 levels of nesting
- Max 4 parameters (use struct if more needed)

### Line Length
- Max 150 characters per line
- Break long lines at logical boundaries

### File Size
- Max 300 lines per file (soft limit)
- Split large files into focused modules

---

## Enforcement

These rules are enforced by:
1. **Code review** — every PR checked against these rules
2. **linter** — `golangci-lint` with custom rules
3. **CI** — build fails on violations
4. **AI agents** — all code generation follows these rules
