#!/bin/bash

# Find the script's directory and navigate to backend root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT" || exit 1

echo "Running tests..."
echo ""

# Run all tests with coverage
go test -v -cover ./...

echo ""
echo "✓ Tests completed"
