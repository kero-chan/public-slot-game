package main

import (
	tuningbasespin "github.com/slotmachine/backend/cmd/rtp-tuning/tuning-basespin"
	// tuningfreespin "github.com/slotmachine/backend/cmd/rtp-tuning/tuning-freespin"
)

func main() {
	tuningbasespin.ExecuteTuning()
	// tuningfreespin.ExecuteTuning()
}
