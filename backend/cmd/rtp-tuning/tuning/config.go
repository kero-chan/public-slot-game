package tuning

import "runtime"

type SimulationStats struct {
	TotalSpins             int     `json:"total_spins"`
	TotalWinSpins          int     `json:"total_win_spins"`
	AvgFreeSpinsAwarded    float64 `json:"avg_free_spins_awarded"`
	TotalWagered           float64 `json:"total_wagered"`
	TotalWon               float64 `json:"total_won"`
	RTP                    float64 `json:"rtp"`
	FreeSpinsTriggered     int     `json:"free_spins_triggered"`
	FreeSpinsTriggeredRate float64 `json:"free_spins_triggered_rate"`
	MaxWin                 float64 `json:"max_win"`
	MaxWinSpin             int     `json:"max_win_spin"`

	// Hit frequency
	NoWinSpins     int `json:"no_win_spins"`
	SmallWins      int `json:"small_wins"`  // < 5x bet
	MediumWins     int `json:"medium_wins"` // 5x - 20x bet
	BigWins        int `json:"big_wins"`    // 20x - 100x bet
	MegaWins       int `json:"mega_wins"`   // > 100x bet
	LowSymbolWins  int `json:"low_symbol_wins"`
	HighSymbolWins int `json:"high_symbol_wins"`

	// Cascade statistics
	TotalCascades     int     `json:"total_cascades"`
	MaxCascades       int     `json:"max_cascades"`
	AvgCascadesPerWin float64 `json:"avg_cascades_per_win"`

	// Cascade depth distribution (per spin)
	Cascade0Count     int `json:"cascade_0_count"`     // Spins with no win (0 cascades)
	Cascade1Count     int `json:"cascade_1_count"`     // Spins with exactly 1 cascade
	Cascade2Count     int `json:"cascade_2_count"`     // Spins with exactly 2 cascades
	Cascade3Count     int `json:"cascade_3_count"`     // Spins with exactly 3 cascades
	Cascade4Count     int `json:"cascade_4_count"`     // Spins with exactly 4 cascades
	Cascade5PlusCount int `json:"cascade_5plus_count"` // Spins with 5+ cascades

	// Win by symbol count (3-of-kind, 4-of-kind, 5-of-kind)
	Win3Count  int     `json:"win3_count"`  // Number of 3-of-kind wins
	Win4Count  int     `json:"win4_count"`  // Number of 4-of-kind wins
	Win5Count  int     `json:"win5_count"`  // Number of 5-of-kind wins
	Win3Amount float64 `json:"win3_amount"` // Total amount from 3-of-kind
	Win4Amount float64 `json:"win4_amount"` // Total amount from 4-of-kind
	Win5Amount float64 `json:"win5_amount"` // Total amount from 5-of-kind

	// Near hit statistics (almost winning - symbol on reel 1,2 but not reel 3)
	NearHit3Count int `json:"near_hit3_count"` // Symbol on reel 1,2 but not reel 3 (gần trúng 3-of-kind)

	FaWinAmount        float64 `json:"fa_win_amount"`
	FaWinCount         float64 `json:"fa_win_count"`
	ZhongWinAmount     float64 `json:"zhong_win_amount"`
	ZhongWinCount      float64 `json:"zhong_win_count"`
	BaiWinAmount       float64 `json:"bai_win_amount"`
	BaiWinCount        float64 `json:"bai_win_count"`
	BawanWinAmount     float64 `json:"bawan_win_amount"`
	BawanWinCount      float64 `json:"bawan_win_count"`
	WusuoWinAmount     float64 `json:"wusuo_win_amount"`
	WusuoWinCount      float64 `json:"wusuo_win_count"`
	WutongWinAmount    float64 `json:"wutong_win_amount"`
	WutongWinCount     float64 `json:"wutong_win_count"`
	LiangsuoWinAmount  float64 `json:"liangsuo_win_amount"`
	LiangsuoWinCount   float64 `json:"liangsuo_win_count"`
	LiangtongWinAmount float64 `json:"liangtong_win_amount"`
	LiangtongWinCount  float64 `json:"liangtong_win_count"`

	// Symbol RTP contribution by category (per base_rtp_tuning.md STEP 5)
	// Target: Low 60-70%, Mid 25-30%, High <10%
	LowSymbolRTPPct  float64 `json:"low_symbol_rtp_pct"`  // Low symbols: liangtong, liangsuo, wusuo, wutong
	MidSymbolRTPPct  float64 `json:"mid_symbol_rtp_pct"`  // Mid symbols: bawan, bai
	HighSymbolRTPPct float64 `json:"high_symbol_rtp_pct"` // High symbols: zhong, fa
}

type TuningConfig struct {
	TotalSpin                        int
	MaxIter                          int
	BetAmount                        float64
	BuyCost                          float64
	TargetRTP                        float64
	TargetRTPTolerance               float64
	TargetBonusTriggerRate           float64
	TargetBonusTriggerRateTolerance  float64
	TargetHitRate                    float64
	TargetHitRateTolerance           float64
	TargetHighSymbolWinRate          float64
	TargetHighSymbolWinRateTolerance float64

	// Parallel simulation config
	ParallelCfg          ParallelConfig
	ResetDensities       bool
	TopologyLearningRate float64
	SaveToDB             bool
	GameMode             string
}

// ParallelConfig holds configuration for parallel simulation
type ParallelConfig struct {
	NumWorkers int  // Number of goroutines to use
	Enabled    bool // Enable parallel simulation
}

// DefaultParallelConfig returns default parallel configuration
func DefaultParallelConfig() ParallelConfig {
	numCPU := runtime.NumCPU()
	return ParallelConfig{
		NumWorkers: numCPU,
		Enabled:    true,
	}
}

type WorkerResult struct {
	Stats              SimulationStats
	SpinsProcessed     int
	FreeSpinSpins      int // Total free spin spins processed
	FreeSpinWinSpins   int // Winning spins in free spin mode
	FreeSpinNoWinSpins int // Non-winning spins in free spin mode
}

// mergeStats merges multiple worker results into a single SimulationStats
func MergeStats(results []WorkerResult) SimulationStats {
	merged := SimulationStats{}

	for _, r := range results {
		s := r.Stats

		// Basic counts
		merged.TotalSpins += s.TotalSpins
		merged.TotalWinSpins += s.TotalWinSpins
		merged.AvgFreeSpinsAwarded += s.AvgFreeSpinsAwarded
		merged.TotalWagered += s.TotalWagered
		merged.TotalWon += s.TotalWon
		merged.FreeSpinsTriggered += s.FreeSpinsTriggered

		// Max win tracking
		if s.MaxWin > merged.MaxWin {
			merged.MaxWin = s.MaxWin
			merged.MaxWinSpin = s.MaxWinSpin
		}

		// Win categories
		merged.NoWinSpins += s.NoWinSpins
		merged.SmallWins += s.SmallWins
		merged.MediumWins += s.MediumWins
		merged.BigWins += s.BigWins
		merged.MegaWins += s.MegaWins
		merged.LowSymbolWins += s.LowSymbolWins
		merged.HighSymbolWins += s.HighSymbolWins

		// Cascade statistics
		merged.TotalCascades += s.TotalCascades
		if s.MaxCascades > merged.MaxCascades {
			merged.MaxCascades = s.MaxCascades
		}

		// Cascade depth distribution
		merged.Cascade0Count += s.Cascade0Count
		merged.Cascade1Count += s.Cascade1Count
		merged.Cascade2Count += s.Cascade2Count
		merged.Cascade3Count += s.Cascade3Count
		merged.Cascade4Count += s.Cascade4Count
		merged.Cascade5PlusCount += s.Cascade5PlusCount

		// Win by symbol count
		merged.Win3Count += s.Win3Count
		merged.Win4Count += s.Win4Count
		merged.Win5Count += s.Win5Count
		merged.Win3Amount += s.Win3Amount
		merged.Win4Amount += s.Win4Amount
		merged.Win5Amount += s.Win5Amount

		// Near hit
		merged.NearHit3Count += s.NearHit3Count

		// Free spins

		// Symbol win amounts
		merged.FaWinAmount += s.FaWinAmount
		merged.FaWinCount += s.FaWinCount
		merged.ZhongWinAmount += s.ZhongWinAmount
		merged.ZhongWinCount += s.ZhongWinCount
		merged.BaiWinAmount += s.BaiWinAmount
		merged.BaiWinCount += s.BaiWinCount
		merged.BawanWinAmount += s.BawanWinAmount
		merged.BawanWinCount += s.BawanWinCount
		merged.WusuoWinAmount += s.WusuoWinAmount
		merged.WusuoWinCount += s.WusuoWinCount
		merged.WutongWinAmount += s.WutongWinAmount
		merged.WutongWinCount += s.WutongWinCount
		merged.LiangsuoWinAmount += s.LiangsuoWinAmount
		merged.LiangsuoWinCount += s.LiangsuoWinCount
		merged.LiangtongWinAmount += s.LiangtongWinAmount
		merged.LiangtongWinCount += s.LiangtongWinCount
	}

	return merged
}
