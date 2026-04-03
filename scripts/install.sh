#!/bin/bash
set -e

REPO="ahmad-ubaidillah/aigo"
INSTALL_DIR="${AIGO_INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="aigo"

detect_os() {
    case "$(uname -s)" in
        Darwin*) echo "darwin" ;;
        Linux*) echo "linux" ;;
        *) echo "unknown" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        arm64|aarch64) echo "arm64" ;;
        x86_64) echo "amd64" ;;
        *) echo "unknown" ;;
    esac
}

get_latest_version() {
    curl -sf "https://api.github.com/repos/${REPO}/releases/latest" | grep -oP '"tag_name":\s*"\K[^"]+'
}

download_binary() {
    OS=$(detect_os)
    ARCH=$(detect_arch)
    
    if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
        echo "Unsupported platform" >&2
        return 1
    fi
    
    VERSION=$(get_latest_version)
    BINARY_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${OS}-${ARCH}"
    
    mkdir -p "$INSTALL_DIR"
    curl -fsSL -o "$INSTALL_DIR/$BINARY_NAME" "$BINARY_URL"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
}

fallback_build() {
    if ! command -v go &> /dev/null; then
        echo "Error: Go is not installed" >&2
        exit 1
    fi
    
    TEMP_DIR=$(mktemp -d)
    git clone --depth 1 --branch main "https://github.com/${REPO}.git" "$TEMP_DIR" 2>/dev/null || {
        echo "Error: Could not clone repository" >&2
        exit 1
    }
    
    cd "$TEMP_DIR"
    mkdir -p "$INSTALL_DIR"
    go build -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/aigo
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    rm -rf "$TEMP_DIR"
}

main() {
    echo "Installing Aigo..."
    
    if download_binary; then
        echo "Installed: $INSTALL_DIR/$BINARY_NAME"
    else
        echo "Release not found. Building from source..."
        fallback_build
        echo "Installed: $INSTALL_DIR/$BINARY_NAME"
    fi
    
    echo ""
    echo "Add to PATH if needed:"
    echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
    echo "  source ~/.bashrc"
    echo ""
    echo "Run: $BINARY_NAME"
}

main "$@"
