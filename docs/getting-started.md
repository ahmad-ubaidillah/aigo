# Getting Started with Aigo

Welcome to Aigo! This guide will help you get up and running quickly.

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/ahmad-ubaidillah/aigo.git
cd aigo

# 2. Build the binary
go build -ldflags="-s -w" -o aigo ./cmd/aigo/

# 3. (Optional) Compress with UPX for smaller size
upx -9 aigo

# 4. Run!
./aigo "Hello, help me write a Go function"
```

## Usage Modes

### Interactive Chat

```bash
./aigo
```

This opens an interactive terminal interface. Type your messages and press Enter.

### One-Shot Query

```bash
./aigo "What is 2+2?"
```

### Gateway Mode (Multi-channel)

```bash
./aigo start
```

Starts the gateway server for Telegram, Discord, Slack, WhatsApp integration.

## Your First Conversation

```
$ ./aigo
aigo> help me create a simple HTTP server in Go
```

Aigo will:
1. Understand your intent
2. Create the code
3. Ask for confirmation
4. Write the file

## Configuration

Create `~/.aigo/config.yaml`:

```yaml
llm:
  provider: "openai"
  model: "gpt-4o-mini"
  api_key: "${OPENAI_API_KEY}"

agent:
  max_iterations: 50
  max_tokens: 8000
```

You can also use environment variables:

```bash
export OPENAI_API_KEY="sk-..."
./aigo "your query"
```

## Common Commands

| Command | Description |
|---------|-------------|
| `./aigo` | Start interactive chat |
| `./aigo "query"` | One-shot query |
| `./aigo start` | Start gateway |
| `./aigo version` | Show version |
| `./aigo help` | Show help |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `AIGO_API_KEY` | Default API key |
| `AIGO_BASE_URL` | Custom endpoint |
| `AIGO_MODEL` | Default model |
| `AIGO_PROVIDER` | Default provider |

## Next Steps

- [Installation Guide](./installation.md) - Detailed installation
- [Features](./features.md) - Explore all features
- [Providers](./providers.md) - See available AI providers
- [Tools](./tools.md) - Available tools reference

## Troubleshooting

### "API key not found"

Set your API key:

```bash
export OPENAI_API_KEY="sk-..."
# or
export ANTHROPIC_API_KEY="sk-ant-..."
```

### "Permission denied"

Make the binary executable:

```bash
chmod +x aigo
```

### Something not working

Run with verbose logging:

```bash
AIGO_LOG_LEVEL=debug ./aigo "query"
```

## Examples

### Simple Query

```bash
./aigo "What is Go?"
```

### Code Generation

```bash
./aigo "Write a function to reverse a string in Go"
```

### File Editing

```bash
./aigo "Add error handling to main.go"
```

### Project Work

```bash
./aigo "Create a REST API with Gin"
```

---

**Need more help?** Check the [documentation](./installation.md) or open an issue on GitHub.