.PHONY: build run test clean tidy build-linux build-darwin-arm64 build-darwin-amd64 build-all

# Build the binary
build:
	@echo "Building aigo..."
	@go build -o dist/aigo ./cmd/aigo
	@echo "Done: dist/aigo"

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
