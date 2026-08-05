#!/bin/bash

set -e

# Build script for Owl CLI

echo "Building Owl CLI..."

cd "$(dirname "$0")/.."

# Build the CLI
echo "Building owl binary..."
go build -o owl .

echo "Build complete: ./owl"
echo ""
echo "Usage: ./owl run <test-file-or-directory>"
