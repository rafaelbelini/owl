#!/bin/bash

set -e

# Install script for Owl CLI
# This script builds and optionally installs the owl binary to PATH

echo "Installing Owl CLI..."

cd "$(dirname "$0")"

# Build the CLI
echo "Building owl binary..."
go build -o owl .

# Default install path
INSTALL_DIR="/usr/local/bin"
INSTALL_OWL="${INSTALL_DIR}/owl"

# Check for custom install path
if [ -n "$1" ]; then
    INSTALL_DIR="$1"
    INSTALL_OWL="${INSTALL_DIR}/owl"
fi

# Create install directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Install the binary
echo "Installing owl to ${INSTALL_OWL}..."
cp owl "$INSTALL_OWL"
chmod +x "$INSTALL_OWL"

echo ""
echo "Owl CLI installed successfully!"
echo ""
echo "Usage: owl run <test-file-or-directory>"
echo ""
echo "If ${INSTALL_DIR} is not in your PATH, add it:"
echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
