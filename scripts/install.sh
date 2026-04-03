#!/bin/bash
set -e

REPO="ahmad-ubaidillah/aigo"
INSTALL_DIR="${AIGO_INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="aigo"

PURPLE='\033[0;35m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${PURPLE}"
cat << 'EOF'
db         88    ,ad8888ba,     ,ad8888ba,    
 d88b        88   d8"'    `"8b   d8"'    `"8b   
d8'`8b       88  d8'            d8'        `8b  
d8'  `8b      88  88             88          88  
d8YaaaaY8b     88  88      88888  88          88  
d8""""""""8b    88  Y8,        88  Y8,        ,8P 
d8'        `8b   88   Y8a.    .a88   Y8a.    .a8P  
d8'          `8b  88    `"Y88888P"     `"Y8888Y"'  
EOF
echo -e "${NC}"

echo -e "${GREEN}⚡ Aigo Installer${NC} - your buddy aigo"
echo ""

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

check_prerequisites() {
    echo -e "${YELLOW}▸ Checking prerequisites...${NC}"
    
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}✗ curl is required but not installed${NC}"
        exit 1
    fi
    
    if ! command -v git &> /dev/null; then
        echo -e "${YELLOW}! git not found, will use go install${NC}"
    fi
    
    echo -e "${GREEN}✓ Prerequisites OK${NC}"
}

get_latest_version() {
    VERSION=$(curl -sf "https://api.github.com/repos/${REPO}/releases/latest" | grep -oP '"tag_name":\s*"\K[^"]+' || echo "")
    echo "$VERSION"
}

download_binary() {
    OS=$(detect_os)
    ARCH=$(detect_arch)
    
    if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
        echo -e "${RED}✗ Unsupported platform: $(uname -s)/$(uname -m)${NC}" >&2
        return 1
    fi
    
    echo -e "${YELLOW}▸ Fetching latest release...${NC}"
    VERSION=$(get_latest_version)
    
    if [ -z "$VERSION" ]; then
        echo -e "${YELLOW}! No release found, will build from source${NC}"
        return 1
    fi
    
    echo -e "${YELLOW}▸ Downloading Aizen ${VERSION} for ${OS}/${ARCH}...${NC}"
    
    BINARY_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${OS}-${ARCH}"
    
    mkdir -p "$INSTALL_DIR"
    
    if curl -fsSL --connect-timeout 30 --max-time 120 -o "$INSTALL_DIR/$BINARY_NAME" "$BINARY_URL" 2>/dev/null; then
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
        echo -e "${GREEN}✓ Download complete${NC}"
        return 0
    else
        rm -f "$INSTALL_DIR/$BINARY_NAME"
        echo -e "${YELLOW}! Download failed, will try build from source${NC}"
        return 1
    fi
}

fallback_build() {
    if ! command -v go &> /dev/null; then
        echo -e "${RED}✗ Go is not installed. Install from: https://go.dev/dl/${NC}" >&2
        echo ""
        echo "Or use manual installation:"
        echo "  git clone https://github.com/${REPO}.git"
        echo "  cd aigo"
        echo "  go install ./cmd/aigo"
        exit 1
    fi
    
    echo -e "${YELLOW}▸ Building from source with Go...${NC}"
    
    TEMP_DIR=$(mktemp -d)
    git clone --depth 1 --branch main "https://github.com/${REPO}.git" "$TEMP_DIR" 2>/dev/null || {
        echo -e "${RED}✗ Failed to clone repository${NC}" >&2
        rm -rf "$TEMP_DIR"
        exit 1
    }
    
    cd "$TEMP_DIR"
    mkdir -p "$INSTALL_DIR"
    go build -o "$INSTALL_DIR/$BINARY_NAME" ./cmd/aigo
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    rm -rf "$TEMP_DIR"
    
    echo -e "${GREEN}✓ Build complete${NC}"
}

add_to_path() {
    SHELL_RC=""
    if [ -f "$HOME/.bashrc" ]; then
        SHELL_RC="$HOME/.bashrc"
    elif [ -f "$HOME/.zshrc" ]; then
        SHELL_RC="$HOME/.zshrc"
    fi
    
    if [ -n "$SHELL_RC" ]; then
        if ! grep -q "\.local/bin" "$SHELL_RC" 2>/dev/null; then
            echo ""
            echo -e "${YELLOW}▸ Adding to PATH...${NC}"
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_RC"
            echo -e "${GREEN}✓ Added to ${SHELL_RC}${NC}"
            echo "  Run: source ${SHELL_RC}"
        fi
    fi
}

main() {
    check_prerequisites
    
    echo -e "${YELLOW}▸ Installing Aizen...${NC}"
    echo ""
    
    if download_binary; then
        echo -e "${GREEN}✓ Installed: ${INSTALL_DIR}/${BINARY_NAME}${NC}"
    else
        echo -e "${YELLOW}▸ Building from source...${NC}"
        fallback_build
        echo -e "${GREEN}✓ Installed: ${INSTALL_DIR}/${BINARY_NAME}${NC}"
    fi
    
    add_to_path
    
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}✓ Installation complete!${NC}"
    echo ""
    echo "  Run 'aigo' to start"
    echo "  Run 'aigo setup' for first-time configuration"
    echo "  Run 'aigo tui' for interactive mode"
    echo ""
    echo -e "${PURPLE}Execute with Zen ⚕${NC}"
}

main "$@"