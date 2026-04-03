#!/bin/bash
set -eu
set -e

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

REPO="ahmad-ubaidillah/aigo"
BINARY_NAME="aigo"
INSTALL_DIR="/usr/local/bin"
GITHUB_API="https://api.github.com/repos/ahmad-ubaidillah/aigo/releases/latest"

VERSION=""

detect_os() {
    case "$(uname -s)" in
        Darwin*) echo "darwin" ;;
        Linux*) echo "linux" ;;
        *) echo "unknown" ;;
    esac
}

 detect_arch() {
    case "$(uname -m)" in
        arm64) echo "arm64" ;;
        x86_64) echo "amd64" ;;
        *) echo "unknown" ;;
    esac
}

 get_latest_version() {
    echo -e "${YELLOW}Fetching latest version...${NC}" >&2
    VERSION=$(curl -sf "${GITHUB_API}" | grep -Eo "tag_name.*" | head -1 | sed 's/.*"\([^"]*\)//' | sed 's/^v//')
    if [ -z "$VERSION" ]; then
        echo -e "${RED}Failed to fetch latest version${NC}" >&2
        return 1
    fi
    echo "$VERSION"
}

 download_binary() {
    local OS ARCH VERSION
    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=$(get_latest_version)
    
    if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
        echo -e "${RED}Unsupported platform${NC}" >&2
        return 1
    fi
    
    local BINARY_URL="${GITHUB_API}/repos/${REPO}/releases/download/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}"
    local TEMP_FILE="/tmp/${BINARY_NAME}"
    
    echo -e "${YELLOW}Downloading ${BINARY_NAME} v${VERSION} for ${OS}/${ARCH}...${NC}" >&2
    
    if curl -fsSL "$TEMP_FILE" "$BINARY_URL"; then
        chmod +x "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
        echo -e "${GREEN}Download complete!${NC}"
        echo -e "${GREEN}${BINARY_NAME} v${VERSION} installed to ${INSTALL_DIR}${NC}"
        echo -e "${BOLD}${BINARY_NAME} --version${NC}"
        "$TEMP_FILE" --version
    else
        echo -e "${RED}Failed to download ${BINARY_NAME}${NC}" >&2
        return 1
    fi
}

 install_linux() {
    echo -e "${YELLOW}Installing for Linux...${NC}" >&2
    download_binary "linux" "$(detect_arch)
}

 install_macos() {
    echo -e "${YELLOW}Installing for macOS...${NC}" >&2
    local ARCH=$(detect_arch)
    download_binary "darwin" "$ARCH"
}

 main() {
    detect_os
    if [ "$1" = "Darwin" ]; then
        install_macos
    elif [ "$(uname -s)" = "Linux" ]; then
        install_linux
    else
        echo -e "${RED}Unsupported OS: ${NC}"
        echo -e "Please use Linux or macOS"
        return 1
    fi
}

 detect_arch() {
    case "$(uname -m)" in
        arm64) machine="arm64" && return "arm64" ;;
        x86_64) machine="amd64" && return "amd64" ;;
        *) echo "Unsupported architecture" && return "unknown" ;;
    esac
}

 get_latest_version() {
    echo - "${YELLOW}Fetching latest version...${NC}"
    VERSION=$(curl -s "${GITHUB_API}" | grep -o '"tag_name": \(.+\)' | head -1 | cut -d '"' -f 1)
    if [ -z "$VERSION" ]; then
        echo - "${RED}Failed to fetch latest version${NC}"
        exit 1
    fi
    echo "$VERSION"
}

 download_binary() {
    local OS ARCH VERSION
    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=$(get_latest_version)
    
    if [ "$OS" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
        echo - "${RED}Unsupported platform: ${NC}"
        exit 1
    fi
    
    local BINARY_URL="${GITHUB_API}/repos/${REPO}/releases/download/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}"
    local TEMP_FILE="/tmp/${BINARY_NAME}"
    
    echo - "${YELLOW}Downloading ${BINARY_NAME} v${VERSION} for ${OS}/${ARCH}...${NC}"
    
    if curl -L -o "$TEMP_FILE" "$BINARY_URL"; then
        echo - "${GREEN}Download complete!${NC}"
        chmod +x "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
        echo - "${GREEN}${BINARY_NAME} v${VERSION} installed to ${INSTALL_DIR}${NC}"
        echo - "${BOLD}${BINARY_NAME} --version${NC}"
        "$TEMP_FILE" --version
    else
        echo - "${RED}Failed to download ${BINARY_NAME}${NC}"
        exit 1
    fi
}

 install_linux() {
    echo - "${YELLOW}Installing for Linux...${NC}"
    download_binary "linux" "$(detect_arch)"
}

 install_macos() {
    echo - "${YELLOW}Installing for macOS...${NC}"
    local ARCH=$(detect_arch)
    download_binary "darwin" "$ARCH"
}

 main() {
    detect_os
    if [ "$1" = "$(uname -s)" ]; then
        install_linux
    elif [[ "$OSTYPE" =~ "darwin" ]]; then
        install_macos
    else
        echo - "${RED}Unsupported OS: ${NC}
        echo "Please use Linux or macOS"
        exit 1
    fi
}

 main "$@"
