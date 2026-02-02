#!/bin/sh
set -e

# Repository configuration
REPO_OWNER="gmsakibursabbir"
REPO_NAME="tinitui"
BINARY_NAME="tinitui"

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Construct binary name matching release assets
# Format: tinytui-{os}-{arch} (e.g. tinytui-linux-amd64)
ASSET_NAME="tinytui-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    ASSET_NAME="${ASSET_NAME}.exe"
fi

echo "Installing TiniTUI for ${OS}/${ARCH}..."

# Find latest release tag
LATEST_RELEASE_URL="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
TAG_NAME=$(curl -sL "$LATEST_RELEASE_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$TAG_NAME" ]; then
    echo "Error: Could not find latest release."
    exit 1
fi

echo "Latest version: ${TAG_NAME}"

# Construct download URL
DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${TAG_NAME}/${ASSET_NAME}"

# Determine install directory
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    SUDO=""
    
    # Check PATH
    case ":$PATH:" in
        *":$INSTALL_DIR:"*) ;;
        *)
            # Try to auto-add to PATH
            SHELL_CONFIG=""
            if [ -f "$HOME/.zshrc" ]; then
                SHELL_CONFIG="$HOME/.zshrc"
            elif [ -f "$HOME/.bashrc" ]; then
                SHELL_CONFIG="$HOME/.bashrc"
            elif [ -f "$HOME/.profile" ]; then
                SHELL_CONFIG="$HOME/.profile"
            fi
            
            if [ -n "$SHELL_CONFIG" ]; then
                echo "Adding $INSTALL_DIR to PATH in $SHELL_CONFIG..."
                echo "" >> "$SHELL_CONFIG"
                echo "# TiniTUI" >> "$SHELL_CONFIG"
                echo "export PATH=\"\$HOME/.local/bin:\$PATH\"" >> "$SHELL_CONFIG"
                echo "Fixed! Please restart your shell or run: source $SHELL_CONFIG"
            else
                echo "Warning: $INSTALL_DIR is not in your PATH. Please add it manually."
            fi
            ;;
    esac


# Download
TMP_FILE=$(mktemp)
echo "Downloading from $DOWNLOAD_URL..."
HTTP_STATUS=$(curl -sL -w "%{http_code}" -o "$TMP_FILE" "$DOWNLOAD_URL")

if [ "$HTTP_STATUS" != "200" ]; then
    echo "Error: Download failed with status $HTTP_STATUS"
    rm "$TMP_FILE"
    exit 1
fi

# Install
chmod +x "$TMP_FILE"
echo "Installing to $INSTALL_DIR/$BINARY_NAME..."
mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"

echo "Success! Run '$BINARY_NAME' to start."
