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

echo "🎨 Seeding Game Assets..."
echo ""

# Build and run the seed script
echo "Building seed script..."
(cd scripts/seed_assets && go build -o seed_assets)

echo "Running seed..."
./scripts/seed_assets/seed_assets

# Cleanup
rm -f ./scripts/seed_assets/seed_assets

echo ""
echo "✓ Asset seeding completed!"
