# Installation Guide

This guide covers all installation methods for Aigo.

## Prerequisites

- **Go 1.26+** - For building from source
- **API Key** - From your preferred AI provider

## Method 1: Build from Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/ahmad-ubaidillah/aigo.git
cd aigo

# Build the binary
go build -ldflags="-s -w" -o aigo ./cmd/aigo/

# Make it executable
chmod +x aigo

# Test it
./aigo version
```

## Method 2: Compressed Binary (Smaller Size)

```bash
# Build
go build -ldflags="-s -w" -o aigo ./cmd/aigo/

# Install UPX if not present
# macOS: brew install upx
# Linux: apt install upx

# Compress
upx -9 aigo

# Size comparison
ls -lh aigo
```

Expected sizes:
- Normal: ~24MB
- Stripped: ~16MB  
- Compressed: ~6-7MB

## Method 3: Install to PATH

```bash
# Build
go build -ldflags="-s -w" -o aigo ./cmd/aigo/

# Move to PATH
sudo mv aigo /usr/local/bin/

# Now use from anywhere
aigo "Hello"
```

## Method 4: Docker

```dockerfile
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache upx
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o aigo ./cmd/aigo/ && upx -9 aigo

FROM alpine:latest
COPY --from=builder /app/aigo /usr/local/bin/aigo
ENTRYPOINT ["aigo"]
```

Build and run:

```bash
docker build -t aigo .
docker run -it -e OPENAI_API_KEY=sk-... aigo
```

## Configuration

### Minimal Config

Create `~/.aigo/config.yaml`:

```yaml
llm:
  provider: "openai"
  model: "gpt-4o-mini"
  api_key: "${OPENAI_API_KEY}"
```

### Full Config Example

```yaml
llm:
  provider: "openai"
  model: "gpt-4o"
  api_key: "${OPENAI_API_KEY}"
  base_url: ""  # Custom endpoint (optional)

agent:
  max_iterations: 50
  max_tokens: 8000

gateway:
  enabled: false
  host: "127.0.0.1"
  port: 8080
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI key | - |
| `ANTHROPIC_API_KEY` | Anthropic key | - |
| `GOOGLE_API_KEY` | Google key | - |
| `DEEPSEEK_API_KEY` | DeepSeek key | - |
| `AIGO_MODEL` | Default model | gpt-4o-mini |
| `AIGO_PROVIDER` | Default provider | openai |
| `AIGO_BASE_URL` | Custom endpoint | - |

## Platform-Specific

### macOS

```bash
# Install Go
brew install go

# Build
go build -ldflags="-s -w" -o aigo ./cmd/aigo/
upx -9 aigo
```

### Linux (Ubuntu/Debian)

```bash
# Install Go
sudo apt install golang-go upx

# Build
go build -ldflags="-s -w" -o aigo ./cmd/aigo/
upx -9 aigo
```

### Raspberry Pi

For Pi Zero 2 W (32-bit):

```bash
GOOS=linux GOARCH=arm GOMIPS=hardfloat go build -ldflags="-s -w" -o aigo ./cmd/aigo/
upx -9 aigo
```

For Pi 4 (64-bit):

```bash
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o aigo ./cmd/aigo/
upx -9 aigo
```

### Windows

```powershell
# Install Go from https://go.dev/dl/
go build -ldflags="-s -w" -o aigo.exe .\cmd\aigo\
```

## Verification

```bash
# Check it works
./aigo version

# Quick test
./aigo "What is 1+1?"

# Interactive mode
./aigo
```

## Updating

```bash
# Pull latest
git pull origin main

# Rebuild
go build -ldflags="-s -w" -o aigo ./cmd/aigo/
upx -9 aigo
```

## Uninstall

```bash
# If installed to PATH
sudo rm /usr/local/bin/aigo

# Remove config
rm -rf ~/.aigo
```

## Troubleshooting

### "go: command not found"

Install Go from https://go.dev/dl/

### "permission denied"

```bash
chmod +x aigo
```

### "API key not found"

Set your API key:

```bash
export OPENAI_API_KEY="sk-..."
# or use config.yaml
```

### Build errors

Make sure Go version is 1.26+:

```bash
go version
```

## Quick Reference

```bash
# Minimal install
git clone https://github.com/ahmad-ubaidillah/aigo.git
cd aigo
go build -ldflags="-s -w" -o aigo ./cmd/aigo/
upx -9 aigo
./aigo "Hello"
```

---

For more details, see:
- [Getting Started](./getting-started.md)
- [Providers](./providers.md)
- [Tools](./tools.md)