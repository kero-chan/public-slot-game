#!/bin/bash

# Find the script's directory and navigate to backend root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT" || exit 1

echo "Building RTP Simulator..."
go build -o bin/rtp-tuning cmd/rtp-tuning/main.go

if [ $? -ne 0 ]; then
    echo "✗ Build failed"
    exit 1
fi

echo "✓ Build successful"
echo ""

# Run with default 1M spins or use first argument
SPINS=${1:-1000000}
BET=${2:-30.0}
PROGRESS=${3:-500000}
TARGET_RTP=${4:-96.5}
TOLERANCE=${9:-0.2}

BASE_SPIN_TARGET_RTP=${5:-63}
BASE_SPIN_RTP_TOLERANCE=${6:-1}
FREE_SPIN_TRIGGER_RATE=${7:-1.5}
FREE_SPIN_TRIGGER_RATE_TOLERANCE=${8:-0.05}

FREE_SPIN_TARGET_RTP=${6:-33.7}
FREE_SPIN_RTP_TOLERANCE=${7:-1}
FREE_SPIN_RETRIGGER_RATE=${7:-1.5}
FREE_SPIN_RETRIGGER_RATE_TOLERANCE=${8:-0.05}

MAX_ITER=${10:-1000}
SAVE_TO_DB=${12:-true}


echo "Running RTP simulation with $SPINS spins at $BET bet..."
echo ""

ARGS=(
  "-spins=$SPINS"
  "-bet=$BET"
  "-progress=$PROGRESS"
  "-save-to-db=$SAVE_TO_DB"  # true/false
  "-target-rtp=$TARGET_RTP"  # target rtp (base spin)
  "-rtp-tolerance=$TOLERANCE"  # rtp tolerance (total spin)
  "-base-spin-target-rtp=$BASE_SPIN_TARGET_RTP"  # base spin target rtp (base spin)
  "-base-spin-rtp-tolerance=$BASE_SPIN_RTP_TOLERANCE"  # base spin target rtp tolerance (base spin)
  "-free-spin-trigger-rate=$FREE_SPIN_TRIGGER_RATE"  # free spin trigger rate (base spin)
  "-free-spin-trigger-rate-tolerance=$FREE_SPIN_TRIGGER_RATE_TOLERANCE"  # free spin trigger rate tolerance (base spin)
  "-free-spin-target-rtp=$FREE_SPIN_TARGET_RTP"  # free spin target rtp (free spin)
  "-free-spin-rtp-tolerance=$FREE_SPIN_RTP_TOLERANCE"  # free spin target rtp tolerance (free spin)
  "-free-spin-retrigger-rate=$FREE_SPIN_RETRIGGER_RATE"  # free spin retrigger rate (free spin)
  "-free-spin-retrigger-rate-tolerance=$FREE_SPIN_RETRIGGER_RATE_TOLERANCE"  # free spin retrigger rate tolerance (free spin)
  "-max-iter=$MAX_ITER"  # max iterations
)

./bin/rtp-tuning "${ARGS[@]}"
