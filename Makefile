.PHONY: build run test clean tidy build-linux build-darwin-arm64 build-darwin-amd64 build-all install install-system uninstall go-install

# Build the binary
build:
	@echo "Building aigo..."
	@go build -o dist/aigo ./cmd/aigo
	@echo "Done: dist/aigo"

# Install using go install (native Go)
go-install:
	@echo "Installing aigo via go install..."
	@go install ./cmd/aigo
	@echo "Installed: $$(which aigo)"

# Install to ~/.local/bin (shell script alternative)
install:
	@echo "Installing aigo to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@go build -o ~/.local/bin/aigo ./cmd/aigo
	@echo "Installed: ~/.local/bin/aigo"
	@echo "Make sure ~/.local/bin is in your PATH:"
	@echo "  export PATH=\"$$HOME/.local/bin:$$PATH\""
	@echo "Add this to your ~/.bashrc or ~/.zshrc to persist."

# Install to /usr/local/bin (requires sudo)
install-system:
	@echo "Installing aigo to /usr/local/bin (requires sudo)..."
	@go build -o dist/aigo ./cmd/aigo
	@sudo cp dist/aigo /usr/local/bin/aigo
	@echo "Installed: /usr/local/bin/aigo"

# Uninstall from ~/.local/bin
uninstall:
	@echo "Uninstalling aigo from ~/.local/bin..."
	@rm -f ~/.local/bin/aigo
	@echo "Uninstalled."

# Cross-compile: linux-amd64
build-linux:
	@GOOS=linux GOARCH=amd64 go build -o dist/aigo-linux-amd64 ./cmd/aigo
# Cross-compile: darwin-arm64
build-darwin-arm64:
	@GOOS=darwin GOARCH=arm64 go build -o dist/aigo-darwin-arm64 ./cmd/aigo
# Cross-compile: darwin-amd64
build-darwin-amd64:
	@GOOS=darwin GOARCH=amd64 go build -o dist/aigo-darwin-amd64 ./cmd/aigo
# Build all platforms
build-all: build-linux build-darwin-arm64 build-darwin-amd64

# Run the CLI
run:
	@go run ./cmd/aigo

# Run tests
test:
	@go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@go tool cover -html=coverage.out

# Clean build artifacts
clean:
	@rm -rf dist/
	@rm -f coverage.out
	@go clean -cache

# Download and tidy dependencies
tidy:
	@go mod tidy
	@go mod download

# Install dependencies
deps:
	@go mod download

# Lint
lint:
	@golangci-lint run ./...

# Format code
fmt:
	@go fmt ./...

# Full CI pipeline
ci: fmt lint test build
	@echo "CI passed!"

# First run setup
setup:
	@go run ./cmd/aigo setup

.DEFAULT_GOAL := build
