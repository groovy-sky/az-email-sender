#!/bin/bash
# Installation script for azemailsender-cli

set -e

APP_NAME="azemailsender-cli"
GITHUB_REPO="groovy-sky/azemailsender"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}azemailsender-cli installer${NC}"
echo ""

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case $OS in
    Darwin)
        OS_LOWER="darwin"
        ;;
    Linux)
        OS_LOWER="linux"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        OS_LOWER="windows"
        ;;
    *)
        echo -e "${RED}Unsupported operating system: $OS${NC}"
        exit 1
        ;;
esac

case $ARCH in
    x86_64|amd64)
        ARCH_LOWER="amd64"
        ;;
    arm64|aarch64)
        ARCH_LOWER="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

BINARY_NAME="${APP_NAME}-${OS_LOWER}-${ARCH_LOWER}"
if [ "$OS_LOWER" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

echo "Detected platform: ${OS_LOWER}/${ARCH_LOWER}"
echo ""

# Check if we're installing from local build or downloading
if [ -d "dist" ] && [ -f "dist/${BINARY_NAME}" ]; then
    echo "Installing from local build..."
    BINARY_PATH="dist/${BINARY_NAME}"
else
    echo "Downloading latest release..."
    
    # Get latest release info
    LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$LATEST_RELEASE" ]; then
        echo -e "${RED}Failed to get latest release information${NC}"
        exit 1
    fi
    
    echo "Latest release: $LATEST_RELEASE"
    
    # Download binary
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_RELEASE}/${BINARY_NAME}"
    
    echo "Downloading from: $DOWNLOAD_URL"
    
    TEMP_FILE="/tmp/${BINARY_NAME}"
    if ! curl -L -o "$TEMP_FILE" "$DOWNLOAD_URL"; then
        echo -e "${RED}Failed to download ${BINARY_NAME}${NC}"
        exit 1
    fi
    
    BINARY_PATH="$TEMP_FILE"
fi

# Check if install directory exists and is writable
if [ ! -d "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Creating install directory: $INSTALL_DIR${NC}"
    sudo mkdir -p "$INSTALL_DIR"
fi

if [ ! -w "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Installing to $INSTALL_DIR requires sudo permissions${NC}"
    SUDO="sudo"
else
    SUDO=""
fi

# Install binary
echo "Installing ${APP_NAME} to ${INSTALL_DIR}..."

if [ "$OS_LOWER" = "windows" ]; then
    INSTALL_PATH="${INSTALL_DIR}/${APP_NAME}.exe"
else
    INSTALL_PATH="${INSTALL_DIR}/${APP_NAME}"
fi

$SUDO cp "$BINARY_PATH" "$INSTALL_PATH"
$SUDO chmod +x "$INSTALL_PATH"

# Clean up temporary file
if [ -f "/tmp/${BINARY_NAME}" ]; then
    rm "/tmp/${BINARY_NAME}"
fi

echo -e "${GREEN}✓ Installation complete!${NC}"
echo ""
echo "Run '${APP_NAME} --help' to get started."
echo ""

# Verify installation
if command -v "$APP_NAME" >/dev/null 2>&1; then
    echo "Installed version:"
    "$APP_NAME" version
else
    echo -e "${YELLOW}Note: You may need to add $INSTALL_DIR to your PATH${NC}"
    echo "Add this line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\""
fi