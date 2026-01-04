#!/bin/bash

# Find the script's directory and navigate to backend root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT" || exit 1

echo "Building RTP Simulator..."
go build -o bin/rtp-simulator cmd/rtp-simulator/main.go

if [ $? -ne 0 ]; then
    echo "✗ Build failed"
    exit 1
fi

echo "✓ Build successful"
echo ""

# Run with default 1M spins or use first argument
SPINS=${1:-20000000}
BET=${2:-30.0}
PROGRESS=${3:-500000}
TARGET_RTP=${4:-97.1}
IS_REAL_MODE=${5:-false}
PLAYER_ID=${6:-"b76f37bc-8014-41eb-a710-d105a8ae6293"}


echo "Running RTP simulation with $SPINS spins at $BET bet..."
echo ""

ARGS=(
  "-spins=$SPINS"
  "-bet=$BET"
  "-progress=$PROGRESS"
  "-real=$IS_REAL_MODE"
  "-target-rtp=$TARGET_RTP"
  "-player-id=$PLAYER_ID"
)

./bin/rtp-simulator "${ARGS[@]}"
