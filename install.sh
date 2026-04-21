#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Aigo Installer${NC}"

TMPDIR=$(mktemp -d)
cd "$TMPDIR"

echo -e "${YELLOW}Cloning...${NC}"
git clone --depth 1 https://github.com/ahmad-ubaidillah/aigo.git aigo-src
cd aigo-src

echo -e "${YELLOW}Building...${NC}"
go build -ldflags="-s -w" -o aigo ./cmd/aigo/

if command -v upx &> /dev/null; then
    upx -9 aigo
fi

INSTALL_DIR="${1:-/usr/local/bin}"

echo -e "${YELLOW}Installing...${NC}"
sudo mkdir -p "$INSTALL_DIR"
sudo mv aigo "$INSTALL_DIR/aigo"
sudo chmod +x "$INSTALL_DIR/aigo"

cd /
rm -rf "$TMPDIR"

if "$INSTALL_DIR/aigo" version &> /dev/null; then
    echo -e "${GREEN}Done! Run: aigo --help${NC}"
else
    echo -e "${RED}Failed${NC}"
    exit 1
fi