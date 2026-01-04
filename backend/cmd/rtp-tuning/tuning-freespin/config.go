package tuningfreespin

import "github.com/slotmachine/backend/cmd/rtp-tuning/tuning"

// func NewConfig() tuning.TuningConfig {
// 	return tuning.TuningConfig{
// 		TotalSpin:                        10_000,
// 		MaxIter:                          5_000,
// 		BetAmount:                        10.0,
// 		BuyCost:                          10.0,
// 		TargetRTP:                        2150,
// 		TargetRTPTolerance:               10,
// 		TargetBonusTriggerRate:           1.0,
// 		TargetBonusTriggerRateTolerance:  0.05,
// 		TargetHitRate:                    35.0,
// 		TargetHitRateTolerance:           0.5,
// 		TargetHighSymbolWinRate:          30.0,
// 		TargetHighSymbolWinRateTolerance: 1,
// 		ParallelCfg:                      tuning.DefaultParallelConfig(),
// 		ResetDensities:                   true,
// 		TopologyLearningRate:             0.02,
// 		SaveToDB:                         true,
// 		GameMode:                         "free_spins",
// 	}
// }

// Config for bonus spin trigger
func NewConfig() tuning.TuningConfig {
	return tuning.TuningConfig{
		TotalSpin:                        10_000,
		MaxIter:                          5_000,
		BetAmount:                        20.0,
		BuyCost:                          750.0,
		TargetRTP:                        96.5,
		TargetRTPTolerance:               0.5,
		TargetBonusTriggerRate:           1.0,
		TargetBonusTriggerRateTolerance:  0.05,
		TargetHitRate:                    35.0,
		TargetHitRateTolerance:           0.5,
		TargetHighSymbolWinRate:          30.0,
		TargetHighSymbolWinRateTolerance: 1,
		ParallelCfg:                      tuning.DefaultParallelConfig(),
		ResetDensities:                   true,
		TopologyLearningRate:             0.02,
		SaveToDB:                         true,
		GameMode:                         "bonus_spin_trigger",
	}
}
