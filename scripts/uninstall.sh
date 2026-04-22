#!/bin/bash
set -e

INSTALL_DIR="${AIGO_INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="aigo"
DATA_DIR="$HOME/.aigo"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

FORCE=false
if [ "$1" = "--yes" ] || [ "$1" = "-y" ]; then
    FORCE=true
fi

echo -e "${YELLOW}⚡ Aigo Uninstaller${NC}"
echo ""

# Detect binary location
BINARY_PATH=""
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"
elif command -v aigo &> /dev/null; then
    BINARY_PATH=$(command -v aigo)
fi

echo "Binary:  ${BINARY_PATH:-not found}"
echo "Data:    $DATA_DIR"
echo ""

if [ "$FORCE" = false ]; then
    echo -n "Remove Aigo binary and all data? [y/N] "
    read -r CONFIRM
    if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
        echo "Cancelled."
        exit 0
    fi
fi

# Remove binary
if [ -n "$BINARY_PATH" ] && [ -f "$BINARY_PATH" ]; then
    rm -f "$BINARY_PATH"
    echo -e "${GREEN}✓ Binary removed${NC}"
else
    echo -e "${YELLOW}! Binary not found${NC}"
fi

# Remove data directory
if [ -d "$DATA_DIR" ]; then
    rm -rf "$DATA_DIR"
    echo -e "${GREEN}✓ Data removed ($DATA_DIR)${NC}"
else
    echo -e "${YELLOW}! Data directory not found${NC}"
fi

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Aigo has been uninstalled.${NC}"
echo ""
echo "  To reinstall:"
echo "    curl -fsSL https://raw.githubusercontent.com/ahmad-ubaidillah/aigo/main/scripts/install.sh | bash"
echo ""
