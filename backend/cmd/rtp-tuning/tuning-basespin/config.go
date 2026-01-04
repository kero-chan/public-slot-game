package tuningbasespin

import "github.com/slotmachine/backend/cmd/rtp-tuning/tuning"

func NewConfig() tuning.TuningConfig {
	return tuning.TuningConfig{
		TotalSpin:                        1_000_000,
		MaxIter:                          1_000,
		BetAmount:                        10.0,
		TargetRTP:                        63.5,
		TargetRTPTolerance:               0.5,
		TargetBonusTriggerRate:           1.5,
		TargetBonusTriggerRateTolerance:  0.05,
		TargetHitRate:                    35.0,
		TargetHitRateTolerance:           0.5,
		TargetHighSymbolWinRate:          30.0,
		TargetHighSymbolWinRateTolerance: 1,
		ParallelCfg:                      tuning.DefaultParallelConfig(),
		ResetDensities:                   true,
		TopologyLearningRate:             0.02,
		SaveToDB:                         true,
		GameMode:                         "base_game",
	}
}
