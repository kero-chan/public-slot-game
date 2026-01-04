#!/bin/bash
set -e

# Get the script directory
SCRIPT_DIR="$(dirname "$0")"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Change to backend directory
cd "$PROJECT_ROOT"

# Run go mod tidy
go mod tidy

echo "✓ Dependencies installed"
