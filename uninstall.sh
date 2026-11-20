#!/bin/bash
set -e

# Owl CLI Uninstall Script
# Usage: curl -fsSL https://raw.githubusercontent.com/rafael/owl/main/uninstall.sh | bash

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
OWL_BIN="${INSTALL_DIR}/owl"

echo "Uninstalling Owl CLI..."

# Check if owl is installed
if [ ! -f "$OWL_BIN" ]; then
    # Try to find it in common locations
    for path in /usr/local/bin/owl ~/bin/owl ~/.local/bin/owl; do
        if [ -f "$path" ]; then
            OWL_BIN="$path"
            break
        fi
    done
fi

if [ ! -f "$OWL_BIN" ]; then
    echo "Owl CLI is not installed."
    exit 0
fi

# Get current version for confirmation
CURRENT_VERSION=$("$OWL_BIN" --version 2>/dev/null | awk '{print $NF}' || echo "unknown")

echo "  Found: $OWL_BIN (version: $CURRENT_VERSION)"

# Confirm uninstallation (skip with --yes flag)
if [ "$1" != "--yes" ] && [ "$1" != "-y" ]; then
    echo ""
    read -p "Are you sure you want to uninstall Owl CLI? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Uninstall cancelled."
        exit 0
    fi
fi

# Remove the binary
rm -f "$OWL_BIN"

if [ -f "$OWL_BIN" ]; then
    echo "Failed to remove $OWL_BIN. Check permissions."
    exit 1
fi

echo ""
echo "✓ Owl CLI uninstalled successfully"
echo ""
echo "If you installed with a custom INSTALL_DIR, you may need to manually remove:"
echo "  $OWL_BIN"
