#!/bin/bash
set -e

# Get the script directory
SCRIPT_DIR="$(dirname "$0")"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Change to project root
cd "$PROJECT_ROOT"

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

echo "🎰 Seeding Reel Strips..."
echo ""

# Default parameters
MODE="${1:-both}"

echo "Parameters:"
echo "  Game Mode: $MODE"
echo ""

# Build and run the seed script (go run doesn't handle build tags well)
echo "Building seed script..."
(cd scripts/seed_reelstrips && go build -o seed_reelstrips)

echo "Running seed..."
./scripts/seed_reelstrips/seed_reelstrips \
    -mode="$MODE"

# Cleanup
rm -f ./scripts/seed_reelstrips/seed_reelstrips

echo ""
echo "✓ Seeding completed!"
