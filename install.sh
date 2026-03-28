#!/bin/bash
# Install script for AIDev CLI
# Usage: curl -sSL https://install.aidev.sh | sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Determine OS and architecture
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
  Linux*)     OS_NAME="linux";;
  Darwin*)    OS_NAME="darwin";;
  MINGW*)     OS_NAME="windows";;
  MSYS*)      OS_NAME="windows";;
  *)          OS_NAME="unknown";;
esac

case "$ARCH" in
  x86_64)     ARCH_NAME="amd64";;
  aarch64)    ARCH_NAME="arm64";;
  arm64)      ARCH_NAME="arm64";;
  *)          ARCH_NAME="unknown";;
esac

if [ "$OS_NAME" = "unknown" ] || [ "$ARCH_NAME" = "unknown" ]; then
  echo -e "${RED}Error: Unsupported OS ($OS) or architecture ($ARCH)${NC}"
  exit 1
fi

# Get latest version from GitHub
echo -e "${YELLOW}Fetching latest version...${NC}"
LATEST_VERSION=$(curl -s https://api.github.com/repos/aidev/aidev-cli/releases/latest | grep -Po '"tag_name": "\K[^"]*' || echo "")

if [ -z "$LATEST_VERSION" ]; then
  echo -e "${RED}Error: Could not fetch latest version from GitHub${NC}"
  exit 1
fi

VERSION="${LATEST_VERSION#v}" # Remove 'v' prefix if present

echo -e "${GREEN}Latest version: $VERSION${NC}"

# Build download URL
DOWNLOAD_URL="https://github.com/aidev/aidev-cli/releases/download/v${VERSION}/aidev_${VERSION}_${OS_NAME}_${ARCH_NAME}"

if [ "$OS_NAME" = "windows" ]; then
  DOWNLOAD_URL="${DOWNLOAD_URL}.zip"
else
  DOWNLOAD_URL="${DOWNLOAD_URL}.tar.gz"
fi

echo -e "${YELLOW}Downloading from: $DOWNLOAD_URL${NC}"

# Create temporary directory
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Download binary
if ! curl -sSL "$DOWNLOAD_URL" -o "$TMPDIR/aidev.tar.gz"; then
  echo -e "${RED}Error: Failed to download aidev binary${NC}"
  exit 1
fi

# Extract
echo -e "${YELLOW}Extracting...${NC}"
if [ "$OS_NAME" = "windows" ]; then
  unzip -q "$TMPDIR/aidev.tar.gz" -d "$TMPDIR"
else
  tar -xzf "$TMPDIR/aidev.tar.gz" -C "$TMPDIR"
fi

# Find binary (it might be in a subdirectory)
if [ -f "$TMPDIR/aidev/aidev" ]; then
  BINARY_PATH="$TMPDIR/aidev/aidev"
elif [ -f "$TMPDIR/aidev" ]; then
  BINARY_PATH="$TMPDIR/aidev"
else
  echo -e "${RED}Error: Could not find aidev binary in archive${NC}"
  exit 1
fi

# Determine install location
if [ -d "$HOME/.local/bin" ]; then
  INSTALL_DIR="$HOME/.local/bin"
elif [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
elif [ -w "$HOME/bin" ]; then
  INSTALL_DIR="$HOME/bin"
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

# Install
echo -e "${YELLOW}Installing to $INSTALL_DIR...${NC}"
chmod +x "$BINARY_PATH"
cp "$BINARY_PATH" "$INSTALL_DIR/aidev"

# Verify installation
if command -v aidev &> /dev/null; then
  INSTALLED_VERSION=$(aidev --version 2>&1 | grep -oP 'version \K[^ ]*' || echo "unknown")
  echo -e "${GREEN}✓ Successfully installed aidev $INSTALLED_VERSION${NC}"
  echo ""
  echo "Quick start:"
  echo "  aidev login          # Login with email/password or API key"
  echo "  aidev                # Launch the TUI"
  echo "  aidev ssh <name>     # SSH directly to an instance"
  echo "  aidev --help         # Show help"
  echo ""
  echo "Next steps: Run 'aidev login' to get started!"
else
  echo -e "${RED}Error: Installation verification failed${NC}"
  echo "Please ensure $INSTALL_DIR is in your PATH"
  exit 1
fi
