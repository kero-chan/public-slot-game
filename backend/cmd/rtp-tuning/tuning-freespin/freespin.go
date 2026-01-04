package tuningfreespin

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/cmd/rtp-tuning/tuning"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/db"
	"github.com/slotmachine/backend/internal/game/cascade"
	"github.com/slotmachine/backend/internal/game/freespins"
	freespinsEngine "github.com/slotmachine/backend/internal/game/freespins"
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/rng"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/slotmachine/backend/internal/infra/repository"
	"github.com/slotmachine/backend/internal/pkg/cache"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/service"
)

// ProgressTracker tracks progress across all workers using atomic operations
type ProgressTracker struct {
	totalSpins     int64
	processedSpins int64
	startTime      time.Time
	lastReport     time.Time
	interval       int64
	mu             sync.Mutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(totalSpins int, interval int) *ProgressTracker {
	return &ProgressTracker{
		totalSpins: int64(totalSpins),
		startTime:  time.Now(),
		lastReport: time.Now(),
		interval:   int64(interval),
	}
}

// Increment adds to the processed count and reports progress if needed
func (p *ProgressTracker) Increment(count int, prefix string) {
	newCount := atomic.AddInt64(&p.processedSpins, int64(count))

	// Check if we should report (using mutex to prevent concurrent reports)
	if newCount%p.interval < int64(count) {
		p.mu.Lock()
		defer p.mu.Unlock()

		// Double-check after acquiring lock
		now := time.Now()
		if now.Sub(p.lastReport) < time.Second {
			return
		}
		p.lastReport = now

		elapsed := time.Since(p.startTime)
		spinsPerSec := float64(newCount) / elapsed.Seconds()
		remaining := time.Duration(float64(p.totalSpins-newCount)/spinsPerSec) * time.Second

		fmt.Printf("%s Progress: %d/%d spins (%.1f%%) | %.0f spins/sec | ETA: %s\n",
			prefix, newCount, p.totalSpins,
			float64(newCount)/float64(p.totalSpins)*100,
			spinsPerSec, remaining.Round(time.Second))
	}
}

func ExecuteTuning() {
	tuningCfg := NewConfig()
	pgGenerator := tuning.NewPGReelGenerator(60, DefaultTopologies())
	symbolTargets := tuning.DefaultSymbolContributionTargets()

	fmt.Println("Using PG-style reel generator with:")
	fmt.Println("  - Reel topology (R1 activator, R3 core, R5 spike)")
	fmt.Println("  - Spacing rules (min distance between same symbols)")
	fmt.Println("  - Anti-cluster constraints (max consecutive same symbols)")
	fmt.Println("  - Symbol classification: Low(1-3x), Mid(5-6x), High(8-10x)")
	fmt.Println("  - AUTO-TUNING topology densities for symbol contribution targets:")
	fmt.Printf("      Low: %.0f%%, Mid: %.0f%%, High: %.0f%%\n",
		symbolTargets.LowPct, symbolTargets.MidPct, symbolTargets.HighPct)
	fmt.Printf("      Learning rate: %.2f\n", tuningCfg.TopologyLearningRate)
	fmt.Println()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         SLOT MACHINE RTP SIMULATOR v2.0                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Spins:                          %d\n", tuningCfg.TotalSpin)
	fmt.Printf("  Max Iter:                       %d\n", tuningCfg.MaxIter)
	fmt.Printf("  Bet Amount:                     %.2f\n", tuningCfg.BetAmount)
	fmt.Printf("  Target RTP:                     %.2f%%\n", tuningCfg.TargetRTP)
	fmt.Printf("  Target RTP Tolerance:           %.2f%%\n", tuningCfg.TargetRTPTolerance)
	fmt.Printf("  Target Bonus Trigger:           %.2f%%\n", tuningCfg.TargetBonusTriggerRate)
	fmt.Printf("  Target Bonus Trigger Tolerance: %.2f%%\n", tuningCfg.TargetBonusTriggerRateTolerance)
	fmt.Printf("  Target Hit Rate:                %.2f%%\n", tuningCfg.TargetHitRate)
	fmt.Printf("  Target Hit Rate Tolerance:      %.2f%%\n", tuningCfg.TargetHitRateTolerance)
	fmt.Printf("  Target High Symbol Win Rate:    %.2f%%\n", tuningCfg.TargetHighSymbolWinRate)
	fmt.Printf("  Reset Densities:                %v\n", tuningCfg.ResetDensities)
	fmt.Printf("  Topology Learning Rate:         %.2f\n", tuningCfg.TopologyLearningRate)
	fmt.Printf("  Save to DB:                     %v\n", tuningCfg.SaveToDB)
	fmt.Printf("  Parallel: %v (%d workers)\n", tuningCfg.ParallelCfg.Enabled, tuningCfg.ParallelCfg.NumWorkers)
	fmt.Println()

	fmt.Println()
	fmt.Println("Starting simulation...")
	fmt.Println()

	reelStrips, stats := executeTuning(&tuningCfg, pgGenerator, 100000)
	if tuningCfg.SaveToDB {
		fmt.Println("✓ Saved to database")
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}
		// Initialize logger
		log := logger.ProvideLogger(cfg)
		// Initialize database
		database, err := db.ProvideDatabase(cfg, log)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
			os.Exit(1)
		}

		cacheClient := cache.ProvideCache(cfg, log)

		// Initialize repositories and services
		reelStripRepo := repository.NewReelStripGormRepository(database, cacheClient)
		reelStripService := service.NewReelStripService(reelStripRepo, log)

		var stripIDs [5]uuid.UUID
		var allStrips []*reelstrip.ReelStrip
		for reelNum := 0; reelNum < 5; reelNum++ {
			stripData := reelStrips[reelNum]

			// Convert reels.ReelStrip ([]string) to regular []string
			stripSlice := []string(stripData)

			// Calculate checksum
			checksum := tuning.CalculateChecksum(stripSlice)

			strip := &reelstrip.ReelStrip{
				ID:          uuid.New(),
				GameMode:    tuningCfg.GameMode,
				ReelNumber:  reelNum,
				StripData:   stripSlice,
				Checksum:    checksum,
				StripLength: len(stripSlice),
				IsActive:    true,
			}

			allStrips = append(allStrips, strip)
			stripIDs[reelNum] = strip.ID
		}

		// Save all strips in batch
		if err := reelStripRepo.CreateBatch(context.Background(), allStrips); err != nil {
			fmt.Printf("failed to save reel strips: %s", err.Error())
		}

		timestamp := time.Now().Format("20060102-150405")
		configName := fmt.Sprintf("%s-%0.2f-%s", tuningCfg.GameMode, stats.RTP, timestamp)
		extraInfo := map[string]any{
			"stats":    stats,
			"paytable": symbols.Paytable,
		}

		extraInfoJSON, _ := json.Marshal(extraInfo)
		_, err = reelStripService.CreateConfig(
			context.Background(),
			configName,
			tuningCfg.GameMode,
			fmt.Sprintf("GameMode: %s\nTuning with %d spins\nRTP = %0.3f%%\nFreeSpinTriggerRate = %.3f%%\n", tuningCfg.GameMode, stats.TotalSpins, stats.RTP, stats.FreeSpinsTriggeredRate),
			stripIDs,
			stats.RTP,
			extraInfoJSON,
		)

		if err != nil {
			fmt.Printf("failed to create config: %s", err.Error())
		}
	}

	tuning.PrintResults(stats, tuningCfg.BetAmount, tuningCfg.TargetRTP)
}

func executeTuning(tuningCfg *tuning.TuningConfig, pgGenerator *tuning.PGReelGenerator, progressInterval int) ([]reels.ReelStrip, tuning.SimulationStats) {
	var reelStrips []reels.ReelStrip
	weightsSet := tuning.ConvertToReelWeightsSet(symbols.FreeSpinsWeights)
	for iter := 1; iter <= tuningCfg.MaxIter; iter++ {
		weightsSet = tuning.ConvertToReelWeightsSet(symbols.FreeSpinsWeights)

		// Generate reel strips using PG generator
		pgStrips, genErr := pgGenerator.GenerateAllReelStrips(weightsSet)
		if genErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate PG free spin reel strips: %v\n", genErr)
			os.Exit(1)
		}

		// Convert [][]string to []reels.ReelStrip
		reelStrips = tuning.ConvertToReelStrips(pgStrips)
		stats := runParallelSpinSimulation(
			reelStrips,
			tuningCfg,
			tuningCfg.ParallelCfg,
			progressInterval,
		)

		if stats.TotalWagered > 0 {
			stats.RTP = (stats.TotalWon / stats.TotalWagered) * 100
		}

		if stats.TotalCascades > 0 {
			stats.AvgCascadesPerWin = float64(stats.TotalCascades) / float64(stats.TotalSpins)
		}

		if stats.FreeSpinsTriggered > 0 {
			stats.AvgFreeSpinsAwarded /= float64(stats.FreeSpinsTriggered)
			stats.FreeSpinsTriggeredRate = float64(stats.FreeSpinsTriggered) / float64(stats.TotalSpins) * 100
		}

		var highSymbolWinsRate, lowSymbolWinsRate float64
		if stats.HighSymbolWins+stats.LowSymbolWins > 0 {
			highSymbolWinsRate = float64(stats.HighSymbolWins) / float64(stats.HighSymbolWins+stats.LowSymbolWins) * 100
			lowSymbolWinsRate = float64(stats.LowSymbolWins) / float64(stats.HighSymbolWins+stats.LowSymbolWins) * 100
		}

		if stats.TotalWon > 0 {
			// Low symbols contribution
			lowWin := stats.LiangtongWinAmount + stats.LiangsuoWinAmount + stats.WusuoWinAmount + stats.WutongWinAmount
			// Mid symbols contribution
			midWin := stats.BawanWinAmount + stats.BaiWinAmount
			// High symbols contribution
			highWin := stats.ZhongWinAmount + stats.FaWinAmount
			stats.LowSymbolRTPPct = (lowWin / stats.TotalWon) * 100
			stats.MidSymbolRTPPct = (midWin / stats.TotalWon) * 100
			stats.HighSymbolRTPPct = (highWin / stats.TotalWon) * 100
		}
		for _, s := range pgGenerator.GetAllReelDensities() {
			fmt.Printf("      R%d %-9s: Low=%.2f Mid=%.2f High=%.2f\n",
				s.ReelIndex, tuning.RoleName(s.Role), s.LowAvg, s.MidAvg, s.HighAvg)
		}

		// Calculate winning spins rate
		hitRate := float64(stats.TotalWinSpins) / float64(stats.TotalSpins) * 100

		// Calculate total wins and percentages
		totalWinCount := stats.Win3Count + stats.Win4Count + stats.Win5Count
		var win3Pct, win4Pct, win5Pct float64
		if totalWinCount > 0 {
			win3Pct = float64(stats.Win3Count) / float64(totalWinCount) * 100
			win4Pct = float64(stats.Win4Count) / float64(totalWinCount) * 100
			win5Pct = float64(stats.Win5Count) / float64(totalWinCount) * 100
		}

		nearHit3Rate := float64(stats.NearHit3Count) / float64(stats.TotalSpins) * 100
		fmt.Printf("=========================== REPORT ===============================\n\n")
		// Print iteration summary with topology and adjusted weights
		if pgGenerator != nil {
			tuning.PrintIterationSummary(iter, "BASE SPIN", pgGenerator, weightsSet)
		}
		fmt.Printf("Tuning FreeSpinRTP Iter %d: BonusTriggerRate:%.3f%% AvgFreeSpinsAwarded:%.2f%% RTP:%.3f%% HitRate:%.2f%% HighWinsRate:%0.2f%% LowWinsRate:%0.2f%%\n",
			iter, stats.FreeSpinsTriggeredRate, stats.AvgFreeSpinsAwarded, stats.RTP, hitRate, highSymbolWinsRate, lowSymbolWinsRate)
		fmt.Printf("  Symbol RTP Contribution: Low=%.2f%% (target 60-70%%) | Mid=%.2f%% (target 25-30%%) | High=%.2f%% (target <10%%)\n",
			stats.LowSymbolRTPPct, stats.MidSymbolRTPPct, stats.HighSymbolRTPPct)
		fmt.Printf("  Win Distribution: 3-kind=%.2f%% 4-kind=%.2f%% 5-kind=%.2f%%\n", win3Pct, win4Pct, win5Pct)
		// fmt.Printf("  Win Amounts: 3-kind=%.2f 4-kind=%.2f 5-kind=%.2f\n", stats.Win3Amount, stats.Win4Amount, stats.Win5Amount)
		fmt.Printf("  Near Hit (reel 1,2 but not 3): %.2f%%\n", nearHit3Rate)

		// Cascade depth distribution
		totalSpinsF := float64(stats.TotalSpins)
		// cascade0Pct := float64(stats.Cascade0Count) / totalSpinsF * 100
		cascade1Pct := float64(stats.Cascade1Count) / totalSpinsF * 100
		cascade2Pct := float64(stats.Cascade2Count) / totalSpinsF * 100
		cascade3Pct := float64(stats.Cascade3Count) / totalSpinsF * 100
		cascade4Pct := float64(stats.Cascade4Count) / totalSpinsF * 100
		cascade5PlusPct := float64(stats.Cascade5PlusCount) / totalSpinsF * 100
		fmt.Printf("  Cascade Depth: 1=%.2f%% 2=%.2f%% 3=%.2f%% 4=%.2f%% 5+=%.2f%%\n",
			cascade1Pct,
			cascade2Pct,
			cascade3Pct,
			cascade4Pct,
			cascade5PlusPct)
		fmt.Printf("=========================== END REPORT ===========================\n\n\n")

		matchRTP := math.Abs(stats.RTP-tuningCfg.TargetRTP) <= tuningCfg.TargetRTPTolerance
		matchFreeTriggerRate := math.Abs(stats.FreeSpinsTriggeredRate-tuningCfg.TargetBonusTriggerRate) <= tuningCfg.TargetBonusTriggerRateTolerance
		matchHighSymbolWinRate := math.Abs(highSymbolWinsRate-tuningCfg.TargetHighSymbolWinRate) <= tuningCfg.TargetHighSymbolWinRateTolerance

		if matchRTP && matchHighSymbolWinRate && matchFreeTriggerRate {
			return reelStrips, stats
		}

		if !matchFreeTriggerRate {
			fmt.Println("Adjusting bonus")
			adjustBonusWeight(tuningCfg, stats.FreeSpinsTriggeredRate, pgGenerator, iter)
		} else {
			adjustWeight(tuningCfg, highSymbolWinsRate, stats.RTP, pgGenerator)
		}
	}

	return []reels.ReelStrip{}, tuning.SimulationStats{}
}

func adjustBonusWeight(tuningCfg *tuning.TuningConfig, currentRate float64, pgGenerator *tuning.PGReelGenerator, iter int) {
	reelIndex := iter % 5
	if _, ok := pgGenerator.GetTopology(reelIndex).SymbolDensity[string(symbols.SymbolBonus)]; !ok {
		pgGenerator.GetTopology(reelIndex).SymbolDensity[string(symbols.SymbolBonus)] = 1.0
	}
	if currentRate < tuningCfg.TargetBonusTriggerRate+tuningCfg.TargetBonusTriggerRateTolerance {
		pgGenerator.GetTopology(reelIndex).SymbolDensity[string(symbols.SymbolBonus)] += tuningCfg.TopologyLearningRate
	} else {
		pgGenerator.GetTopology(reelIndex).SymbolDensity[string(symbols.SymbolBonus)] -= tuningCfg.TopologyLearningRate
	}

	// Normalize all topologies so that minimum density becomes 1.0
	normalizeAllTopologies(pgGenerator)
}

// giữ high và low mỗi mảng 4 phần tử từ lớn nhất đến bé nhất
var highSymbols = []string{
	"fa", "zhong", "bai", "bawan",
}

var lowSymbols = []string{
	"wusuo", "wutong", "liangsuo", "liangtong",
}

var changeRate = map[string]float64{
	"fa":        0.6,
	"zhong":     0.7,
	"bai":       0.8,
	"bawan":     0.9,
	"wusuo":     0.7,
	"wutong":    0.7,
	"liangsuo":  1.0,
	"liangtong": 1.0,
}

// normalizeSymbolDensity normalizes the symbol density map so that
// the minimum value becomes 1.0 and other values are scaled proportionally.
// For example: fa=1.3, zhong=2.5 becomes fa=1.0, zhong=2.5/1.3≈1.923
func normalizeSymbolDensity(density map[string]float64) {
	if len(density) == 0 {
		return
	}

	// Find the minimum value (excluding bonus symbol)
	minVal := math.MaxFloat64
	for sym, val := range density {
		if sym == "bonus" {
			continue // Skip bonus symbol in normalization
		}
		if val < minVal && val > 0 {
			minVal = val
		}
	}

	// If no valid minimum found, return
	if minVal == math.MaxFloat64 || minVal <= 0 {
		return
	}

	// Normalize all values by dividing by the minimum (except bonus)
	for sym := range density {
		if sym == "bonus" {
			continue // Keep bonus symbol as is
		}
		density[sym] = density[sym] / minVal
	}
}

// normalizeAllTopologies normalizes symbol densities across all reels
func normalizeAllTopologies(pgGenerator *tuning.PGReelGenerator) {
	for r := 0; r < 5; r++ {
		topology := pgGenerator.GetTopology(r)
		normalizeSymbolDensity(topology.SymbolDensity)
	}
}

func adjustWeight(tuningCfg *tuning.TuningConfig, highSymbolWinsRate, statsRTP float64, pgGenerator *tuning.PGReelGenerator) {

	if highSymbolWinsRate > tuningCfg.TargetHighSymbolWinRate+tuningCfg.TargetHighSymbolWinRateTolerance {
		fmt.Println("Tăng low symbol win rate")
		reels := []int{1, 3}
		for _, r := range reels {
			for _, sym := range lowSymbols {
				if _, ok := pgGenerator.GetTopology(r).SymbolDensity[string(sym)]; !ok {
					pgGenerator.GetTopology(r).SymbolDensity[string(sym)] = 1.0
				}
				pgGenerator.GetTopology(r).SymbolDensity[string(sym)] += tuningCfg.TopologyLearningRate * changeRate[string(sym)]

			}
		}

	} else if highSymbolWinsRate >= tuningCfg.TargetHighSymbolWinRate-tuningCfg.TargetHighSymbolWinRateTolerance {
		if statsRTP > tuningCfg.TargetRTP {
			fmt.Println("reduce RTP")
			lowReels := []int{2, 4}
			for _, r := range lowReels {
				for _, sym := range lowSymbols {
					if _, ok := pgGenerator.GetTopology(r).SymbolDensity[string(sym)]; !ok {
						pgGenerator.GetTopology(r).SymbolDensity[string(sym)] = 1.0
					}
					pgGenerator.GetTopology(r).SymbolDensity[string(sym)] += tuningCfg.TopologyLearningRate * changeRate[string(sym)]

				}
			}

			highReels := []int{1, 3}
			for _, r := range highReels {
				for _, sym := range highSymbols {
					if _, ok := pgGenerator.GetTopology(r).SymbolDensity[string(sym)]; !ok {
						pgGenerator.GetTopology(r).SymbolDensity[string(sym)] = 1.0
					}
					pgGenerator.GetTopology(r).SymbolDensity[string(sym)] += tuningCfg.TopologyLearningRate * changeRate[string(sym)]

				}
			}
		} else {
			fmt.Println("increase RTP")
			lowReels := []int{1, 3}
			for _, r := range lowReels {
				for _, sym := range lowSymbols {
					if _, ok := pgGenerator.GetTopology(r).SymbolDensity[string(sym)]; !ok {
						pgGenerator.GetTopology(r).SymbolDensity[string(sym)] = 1.0
					}
					pgGenerator.GetTopology(r).SymbolDensity[string(sym)] += tuningCfg.TopologyLearningRate * changeRate[string(sym)]

				}
			}

			highReels := []int{2, 4}
			for _, r := range highReels {
				for _, sym := range highSymbols {
					if _, ok := pgGenerator.GetTopology(r).SymbolDensity[string(sym)]; !ok {
						pgGenerator.GetTopology(r).SymbolDensity[string(sym)] = 1.0
					}
					pgGenerator.GetTopology(r).SymbolDensity[string(sym)] += tuningCfg.TopologyLearningRate * changeRate[string(sym)]

				}
			}
		}
	} else {
		fmt.Println("Tăng high symbol win rate")
		reels := []int{2, 4}
		for _, r := range reels {
			for _, sym := range highSymbols {
				if _, ok := pgGenerator.GetTopology(r).SymbolDensity[string(sym)]; !ok {
					pgGenerator.GetTopology(r).SymbolDensity[string(sym)] = 1.0
				}
				pgGenerator.GetTopology(r).SymbolDensity[string(sym)] += tuningCfg.TopologyLearningRate * changeRate[string(sym)]

			}
		}
	}

	// Normalize all topologies so that minimum density becomes 1.0
	normalizeAllTopologies(pgGenerator)
}
func runParallelSpinSimulation(
	reelStips []reels.ReelStrip,
	tuningCfg *tuning.TuningConfig,
	parallelCfg tuning.ParallelConfig,
	progressInterval int,
) tuning.SimulationStats {
	numWorkers := parallelCfg.NumWorkers
	spinsPerWorker := tuningCfg.TotalSpin / numWorkers
	remainder := tuningCfg.TotalSpin % numWorkers

	results := make([]tuning.WorkerResult, numWorkers)
	var wg sync.WaitGroup

	// Progress tracking
	tracker := NewProgressTracker(tuningCfg.TotalSpin, progressInterval)

	fmt.Printf("Running parallel free spin simulation with %d workers\n", numWorkers)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		workerSpins := spinsPerWorker
		if w < remainder {
			workerSpins++
		}

		go func(workerID int, spins int) {
			defer wg.Done()

			workerRNG := rng.NewFastRNG()
			for i := 0; i < workerID*1000; i++ {
				workerRNG.Intn(100)
			}

			stats := tuning.SimulationStats{}
			batchSize := 1000

			for i := 0; i < spins; i++ {
				scatterCount := 5
				rNumber, _ := workerRNG.Int(100)
				if rNumber < 92 {
					scatterCount = 3
				} else if rNumber < 98 {
					scatterCount = 4
				}

				executeFreeSpins(&stats, reelStips, workerRNG, scatterCount, tuningCfg.BetAmount)
				stats.TotalWagered += tuningCfg.BuyCost

				if (i+1)%batchSize == 0 {
					tracker.Increment(batchSize, "Tuning FreeSpinRTP")
				}
			}

			remainingSpins := spins % batchSize
			if remainingSpins > 0 {
				tracker.Increment(remainingSpins, "Tuning FreeSpinRTP")
			}

			results[workerID] = tuning.WorkerResult{
				Stats:          stats,
				SpinsProcessed: spins,
			}
		}(w, workerSpins)
	}

	wg.Wait()

	// Merge results and sum free spin totals
	merged := tuning.MergeStats(results)

	return merged
}

// executeFreeSpins executes all free spins in a session and returns total win
func executeFreeSpins(stats *tuning.SimulationStats, reelStrips []reels.ReelStrip, rngInstance rng.RNG, scatterCount int, betAmount float64) {
	isFreeSpin := true
	// Create a free spins session
	session := freespinsEngine.NewSession(uuid.Nil, scatterCount, betAmount, nil)
	totalWin := 0.0
	spinNumber := 1

	// Execute all free spins in the session
	for !session.IsComplete() {
		// Generate initial grid
		initialGrid, reelPositions, err := reels.GenerateGrid(reelStrips, rngInstance)
		if err != nil {
			fmt.Printf("failed to generate grid: %s", err.Error())
			os.Exit(1)
		}

		// Execute cascades with free spin multipliers
		cascadeResults, finalGrid, err := cascade.ExecuteCascades(
			initialGrid,
			reelStrips,
			reelPositions,
			betAmount,
			isFreeSpin,
			rngInstance,
		)
		if err != nil {
			fmt.Printf("failed to execute cascades: %s", err.Error())
			break
		}

		// Calculate total win
		totalCascadeWin := cascade.GetTotalWinFromCascades(cascadeResults, betAmount)
		// Check for retrigger
		retriggerResult := freespins.CheckRetrigger(finalGrid, session.RemainingSpins-1)
		hasHighSymbolWin := false
		hasLowSymbolWin := false
		for _, cascade := range cascadeResults {
			for _, winCascade := range cascade.Wins {
				switch winCascade.Count {
				case 3:
					stats.Win3Count++
					stats.Win3Amount += winCascade.WinAmount
				case 4:
					stats.Win4Count++
					stats.Win4Amount += winCascade.WinAmount
				case 5:
					stats.Win5Count++
					stats.Win5Amount += winCascade.WinAmount
				}

				switch winCascade.Symbol {
				case symbols.SymbolFa:
					stats.FaWinAmount += winCascade.WinAmount
					stats.FaWinCount++
					hasHighSymbolWin = true
				case symbols.SymbolZhong:
					stats.ZhongWinAmount += winCascade.WinAmount
					stats.ZhongWinCount++
					hasHighSymbolWin = true
				case symbols.SymbolBai:
					stats.BaiWinAmount += winCascade.WinAmount
					stats.BaiWinCount++
					hasHighSymbolWin = true
				case symbols.SymbolBawan:
					stats.BawanWinAmount += winCascade.WinAmount
					stats.BawanWinCount++
					hasHighSymbolWin = true
				case symbols.SymbolWusuo:
					stats.WusuoWinAmount += winCascade.WinAmount
					stats.WusuoWinCount++
					hasLowSymbolWin = true
				case symbols.SymbolWutong:
					stats.WutongWinAmount += winCascade.WinAmount
					stats.WutongWinCount++
					hasLowSymbolWin = true
				case symbols.SymbolLiangsuo:
					stats.LiangsuoWinAmount += winCascade.WinAmount
					stats.LiangsuoWinCount++
					hasLowSymbolWin = true
				case symbols.SymbolLiangtong:
					stats.LiangtongWinAmount += winCascade.WinAmount
					stats.LiangtongWinCount++
					hasLowSymbolWin = true
				}
			}

		}

		if !hasHighSymbolWin && !hasLowSymbolWin {
			// Track near hits on initial grid (before cascades) - only count once per spin
			if tuning.IsNearHit(initialGrid) {
				stats.NearHit3Count++
			}
		}

		if hasHighSymbolWin {
			stats.HighSymbolWins++
		}
		if hasLowSymbolWin {
			stats.LowSymbolWins++
		}

		stats.TotalWon += totalCascadeWin
		stats.TotalSpins++

		totalWin += totalCascadeWin
		session.ExecuteSpin(totalCascadeWin)

		// Handle retrigger
		if retriggerResult.Retriggered {
			stats.FreeSpinsTriggered++
			stats.AvgFreeSpinsAwarded += float64(retriggerResult.AdditionalSpins)
			session.AddRetriggerSpins(retriggerResult.AdditionalSpins)
		}

		// Categorize win size
		winMultiplier := totalCascadeWin / betAmount
		if totalCascadeWin == 0 {
			stats.NoWinSpins++
		} else if winMultiplier < 5 {
			stats.SmallWins++
		} else if winMultiplier < 20 {
			stats.MediumWins++
		} else if winMultiplier < 100 {
			stats.BigWins++
		} else {
			stats.MegaWins++
		}

		// Track cascades
		numCascades := len(cascadeResults)
		stats.TotalCascades += numCascades
		if numCascades > stats.MaxCascades {
			stats.MaxCascades = numCascades
		}

		if totalCascadeWin > stats.MaxWin {
			stats.MaxWin = totalCascadeWin
		}

		// Track cascade depth distribution
		switch numCascades {
		case 0:
			stats.Cascade0Count++
		case 1:
			stats.Cascade1Count++
		case 2:
			stats.Cascade2Count++
		case 3:
			stats.Cascade3Count++
		case 4:
			stats.Cascade4Count++
		default:
			stats.Cascade5PlusCount++
		}

		// Track free spin wins (not base game)
		if totalCascadeWin > 0 {
			stats.TotalWinSpins++
		}

		spinNumber++
	}
}
