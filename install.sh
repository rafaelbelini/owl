#!/bin/bash
set -e

# Owl CLI Install Script
# Usage: curl -fsSL https://raw.githubusercontent.com/rafael/owl/main/install.sh | bash
# Or with specific version: curl -fsSL https://raw.githubusercontent.com/rafael/owl/v0.1.0/install.sh | bash

REPO="rafael/owl"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Parse version from script URL or use latest
get_version() {
    local url="$1"
    # Try to extract version from URL (e.g., /v0.1.0/install.sh)
    if [[ "$url" =~ /v?([0-9]+\.[0-9]+\.[0-9]+)/ ]]; then
        echo "v${BASH_REMATCH[1]}"
    else
        echo "latest"
    fi
}

VERSION="${VERSION:-$(get_version "$0")}"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux*)  OS="linux" ;;
  darwin*) OS="darwin" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)
    echo "[ERROR] Sistema operacional não suportado: $OS"
    exit 1
    ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "[ERROR] Arquitetura não suportada: $ARCH"
    exit 1
    ;;
esac

# Build binary name
BINARY="owl-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY="${BINARY}.exe"
fi

# Determine download URL
if [ "$VERSION" = "latest" ]; then
    TAG_URL="https://api.github.com/repos/${REPO}/releases/latest"
    VERSION_TAG=$(curl -fsSL "$TAG_URL" | grep '"tag_name"' | cut -d'"' -f4)
    if [ -z "$VERSION_TAG" ]; then
        echo "[ERROR] Não foi possível obter a versão mais recente"
        exit 1
    fi
    VERSION="$VERSION_TAG"
else
    VERSION_TAG="$VERSION"
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION_TAG}/${BINARY}"
SCRIPT_URL="https://raw.githubusercontent.com/${REPO}/${VERSION_TAG}/install.sh"

echo "Installing Owl CLI ${VERSION}..."
echo "  OS: $OS"
echo "  Arch: $ARCH"
echo "  Binary: $BINARY"

# Create temp directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Download binary
echo "Downloading..."
curl -fsSL "$DOWNLOAD_URL" -o "${TEMP_DIR}/owl" || {
    echo "[ERROR] Falha ao baixar o binário de $DOWNLOAD_URL"
    echo "[ERROR] Verifique se a versão ${VERSION_TAG} existe para ${OS}-${ARCH}"
    exit 1
}

# Download install script (for reference)
curl -fsSL "$SCRIPT_URL" -o "${TEMP_DIR}/install.sh" 2>/dev/null || true

# Install
mkdir -p "$INSTALL_DIR"
install -m 755 "${TEMP_DIR}/owl" "${INSTALL_DIR}/owl"

echo ""
echo "✓ Owl CLI installed successfully to ${INSTALL_DIR}/owl"
echo ""

# Verify installation
if "${INSTALL_DIR}/owl" --version &>/dev/null; then
    "${INSTALL_DIR}/owl" --version
else
    echo "⚠ Instalação concluída, mas 'owl' não está no PATH."
    echo "  Adicione ao PATH: export PATH=\"${INSTALL_DIR}:\$PATH\""
fi
