package tuning

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/symbols"
)

// printIterationSummary prints topology and adjusted weights for current iteration
func PrintIterationSummary(iter int, mode string, gen *PGReelGenerator, baseWeights *ReelWeightsSet) {
	fmt.Println()
	fmt.Printf("╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  %s - Iteration %d Summary                                    \n", mode, iter)
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n")
	printTopology(gen)
	printAdjustedWeights(gen, baseWeights)
}

// getDensity returns density value or 1.0 if not set
func getDensity(m map[string]float64, sym string) float64 {
	if v, ok := m[sym]; ok {
		return v
	}
	return 1.0
}

// formatSpacing returns spacing value as string or "-" if not set
func formatSpacing(m map[string]int, sym string) string {
	if v, ok := m[sym]; ok {
		return fmt.Sprintf("%d", v)
	}
	return "-"
}

// printTopology prints the current PG topology configuration in matrix format
func printTopology(gen *PGReelGenerator) {
	if gen == nil {
		fmt.Println("  (PG Generator not initialized)")
		return
	}

	// Get all topologies
	topologies := make([]ReelTopology, 5)
	for i := 0; i < 5; i++ {
		topologies[i] = gen.GetTopology(i)
	}

	// Print Reel Roles
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    REEL TOPOLOGY CONFIG                         │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	fmt.Println("=== REEL ROLES ===")
	fmt.Println("               Reel1       Reel2       Reel3       Reel4       Reel5")
	fmt.Printf("Role           %-11s %-11s %-11s %-11s %-11s\n",
		RoleName(topologies[0].Role),
		RoleName(topologies[1].Role),
		RoleName(topologies[2].Role),
		RoleName(topologies[3].Role),
		RoleName(topologies[4].Role),
	)

	// Print Symbol Density
	fmt.Println("=== SYMBOL DENSITY (multiplier) ===")
	fmt.Println("               Reel1   Reel2   Reel3   Reel4   Reel5")

	allSymbols := []string{"fa", "zhong", "bai", "bawan", "wusuo", "wutong", "liangsuo", "liangtong", "bonus"}
	for _, sym := range allSymbols {
		fmt.Printf("%-14s %5.2f   %5.2f   %5.2f   %5.2f   %5.2f\n",
			sym,
			getDensity(topologies[0].SymbolDensity, sym),
			getDensity(topologies[1].SymbolDensity, sym),
			getDensity(topologies[2].SymbolDensity, sym),
			getDensity(topologies[3].SymbolDensity, sym),
			getDensity(topologies[4].SymbolDensity, sym),
		)
	}

	// Print Min Spacing
	fmt.Println("=== MIN SPACING (positions) ===")
	fmt.Println("               Reel1   Reel2   Reel3   Reel4   Reel5")

	spacingSymbols := []string{"fa", "zhong", "bai", "bawan", "bonus"}
	for _, sym := range spacingSymbols {
		fmt.Printf("%-14s %5s   %5s   %5s   %5s   %5s\n",
			sym,
			formatSpacing(topologies[0].MinSpacing, sym),
			formatSpacing(topologies[1].MinSpacing, sym),
			formatSpacing(topologies[2].MinSpacing, sym),
			formatSpacing(topologies[3].MinSpacing, sym),
			formatSpacing(topologies[4].MinSpacing, sym),
		)
	}

	// Print Forbidden Pairs
	fmt.Println("=== FORBIDDEN PAIRS ===")
	for i := 0; i < 5; i++ {
		if len(topologies[i].ForbiddenPairs) > 0 {
			pairs := ""
			for j, pair := range topologies[i].ForbiddenPairs {
				if j > 0 {
					pairs += ", "
				}
				pairs += fmt.Sprintf("[%s-%s]", pair[0], pair[1])
			}
			fmt.Printf("  Reel%d: %s\n", i+1, pairs)
		}
	}
	fmt.Println()
}

// printAdjustedWeights prints the adjusted weights after applying topology density
func printAdjustedWeights(gen *PGReelGenerator, baseWeights *ReelWeightsSet) {
	if gen == nil || baseWeights == nil {
		return
	}

	adjustedWeights := gen.GetAdjustedWeights(baseWeights)
	if adjustedWeights == nil {
		return
	}

	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│              ADJUSTED WEIGHTS (After Topology Density)          │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")
	fmt.Println("               Reel1  Reel2  Reel3  Reel4  Reel5")

	// Base paying symbols
	baseSymbols := []string{"fa", "zhong", "bai", "bawan", "wusuo", "wutong", "liangsuo", "liangtong"}
	for _, sym := range baseSymbols {
		fmt.Printf("%-14s %4d   %4d   %4d   %4d   %4d\n",
			sym,
			adjustedWeights.Reel1[sym],
			adjustedWeights.Reel2[sym],
			adjustedWeights.Reel3[sym],
			adjustedWeights.Reel4[sym],
			adjustedWeights.Reel5[sym],
		)
	}

	fmt.Println("  --- Special Symbols ---")
	// Special symbols
	specialSymbols := []string{"bonus"}
	for _, sym := range specialSymbols {
		fmt.Printf("%-14s %4d   %4d   %4d   %4d   %4d\n",
			sym,
			adjustedWeights.Reel1[sym],
			adjustedWeights.Reel2[sym],
			adjustedWeights.Reel3[sym],
			adjustedWeights.Reel4[sym],
			adjustedWeights.Reel5[sym],
		)
	}

	// Calculate and print totals
	totals := []int{0, 0, 0, 0, 0}
	reels := []*SymbolWeights{
		&adjustedWeights.Reel1,
		&adjustedWeights.Reel2,
		&adjustedWeights.Reel3,
		&adjustedWeights.Reel4,
		&adjustedWeights.Reel5,
	}
	for i, reel := range reels {
		for _, w := range *reel {
			totals[i] += w
		}
	}
	fmt.Println("  -----------------")
	fmt.Printf("%-14s %4d   %4d   %4d   %4d   %4d\n",
		"TOTAL",
		totals[0], totals[1], totals[2], totals[3], totals[4],
	)
	fmt.Println()

	// Print Gold Configuration (gold symbols are applied post-generation, not via weights)
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│              GOLD SYMBOL CONFIGURATION (Post-Generation)        │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")
	fmt.Println("               Reel1      Reel2      Reel3      Reel4      Reel5")

	// Get gold config for each reel
	goldConfigs := make([]*GoldTopologyConfig, 5)
	for i := 0; i < 5; i++ {
		topology := gen.GetTopology(i)
		goldConfigs[i] = topology.GoldConfig
	}

	// Print Enabled status
	fmt.Printf("%-14s ", "Enabled")
	for i := 0; i < 5; i++ {
		if goldConfigs[i] != nil && goldConfigs[i].Enabled {
			fmt.Printf("%-10s ", "Yes")
		} else {
			fmt.Printf("%-10s ", "No")
		}
	}
	fmt.Println()

	// Print GoldRatio
	fmt.Printf("%-14s ", "GoldRatio")
	for i := 0; i < 5; i++ {
		if goldConfigs[i] != nil && goldConfigs[i].Enabled {
			fmt.Printf("%-10.2f ", goldConfigs[i].GoldRatio*100)
		} else {
			fmt.Printf("%-10s ", "-")
		}
	}
	fmt.Println("(%)")

	// Print MinGoldSpacing
	fmt.Printf("%-14s ", "MinSpacing")
	for i := 0; i < 5; i++ {
		if goldConfigs[i] != nil && goldConfigs[i].Enabled {
			fmt.Printf("%-10d ", goldConfigs[i].MinGoldSpacing)
		} else {
			fmt.Printf("%-10s ", "-")
		}
	}
	fmt.Println()

	// Print MaxGoldCluster
	fmt.Printf("%-14s ", "MaxCluster")
	for i := 0; i < 5; i++ {
		if goldConfigs[i] != nil && goldConfigs[i].Enabled {
			fmt.Printf("%-10d ", goldConfigs[i].MaxGoldCluster)
		} else {
			fmt.Printf("%-10s ", "-")
		}
	}
	fmt.Println()
	fmt.Println()
}

// isNearHit checks if the grid has a near hit (symbol on reel 1,2 but not reel 3)
// Returns true if ANY paying symbol has a near hit pattern (only count once per spin)
func IsNearHit(grid reels.Grid) bool {
	startRow := reels.WinCheckStartRow
	endRow := reels.WinCheckEndRow

	payingSymbols := []symbols.Symbol{
		symbols.SymbolFa, symbols.SymbolZhong, symbols.SymbolBai, symbols.SymbolBawan,
		symbols.SymbolWusuo, symbols.SymbolWutong, symbols.SymbolLiangsuo, symbols.SymbolLiangtong,
	}

	for _, sym := range payingSymbols {
		symStr := string(sym)

		// Check if symbol exists on reel 1, 2, 3
		hasOnReel1 := false
		hasOnReel2 := false
		hasOnReel3 := false

		for row := startRow; row <= endRow; row++ {
			if grid[0][row] == symStr || grid[0][row] == symStr+"_gold" {
				hasOnReel1 = true
			}
			if grid[1][row] == symStr || grid[1][row] == symStr+"_gold" {
				hasOnReel2 = true
			}
			if grid[2][row] == symStr || grid[2][row] == symStr+"_gold" {
				hasOnReel3 = true
			}
		}

		// Near hit: symbol on reel 1 & 2 but NOT on reel 3 (gần trúng 3-of-kind)
		if hasOnReel1 && hasOnReel2 && !hasOnReel3 {
			return true // Found a near hit, return immediately
		}
	}
	return false
}

// convertToReelWeightsSet converts symbols.BaseGameWeights to ReelWeightsSet
func ConvertToReelWeightsSet(weights struct {
	Reel1 symbols.ReelWeights
	Reel2 symbols.ReelWeights
	Reel3 symbols.ReelWeights
	Reel4 symbols.ReelWeights
	Reel5 symbols.ReelWeights
}) *ReelWeightsSet {
	return &ReelWeightsSet{
		Reel1: convertReelWeights(weights.Reel1),
		Reel2: convertReelWeights(weights.Reel2),
		Reel3: convertReelWeights(weights.Reel3),
		Reel4: convertReelWeights(weights.Reel4),
		Reel5: convertReelWeights(weights.Reel5),
	}
}

// convertReelWeights converts symbols.ReelWeights to SymbolWeights
func convertReelWeights(rw symbols.ReelWeights) SymbolWeights {
	sw := make(SymbolWeights)
	for k, v := range rw {
		sw[k] = v
	}
	return sw
}

// convertToReelStrips converts [][]string to []reels.ReelStrip
func ConvertToReelStrips(pgStrips [][]string) []reels.ReelStrip {
	reelStrips := make([]reels.ReelStrip, len(pgStrips))
	for i, strip := range pgStrips {
		reelStrips[i] = reels.ReelStrip(strip)
	}
	return reelStrips
}

func CalculateChecksum(stripData []string) string {
	// Convert to JSON for consistent hashing
	jsonData, _ := json.Marshal(stripData)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

func PrintResults(stats SimulationStats, betAmount float64, targetRTP float64) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    SIMULATION RESULTS                      ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Overall RTP
	fmt.Println("═══ OVERALL STATISTICS ═══")
	fmt.Printf("Total Spins:           %d\n", stats.TotalSpins)
	fmt.Printf("Total Wagered:         %.2f\n", stats.TotalWagered)
	fmt.Printf("Total Won:             %.2f\n", stats.TotalWon)
	fmt.Printf("RTP:                   %.4f%% ", stats.RTP)

	// Color-code RTP result
	diff := stats.RTP - targetRTP
	if diff > -0.3 && diff < 0.3 {
		fmt.Printf("✓ (target: %.2f%%)\n", targetRTP)
	} else if diff > -1.0 && diff < 1.0 {
		fmt.Printf("⚠ (target: %.2f%%, diff: %+.2f%%)\n", targetRTP, diff)
	} else {
		fmt.Printf("✗ (target: %.2f%%, diff: %+.2f%%)\n", targetRTP, diff)
	}
	fmt.Println()

	// Hit frequency
	fmt.Println("═══ HIT FREQUENCY ═══")
	totalWinSpins := stats.SmallWins + stats.MediumWins + stats.BigWins + stats.MegaWins
	hitFrequency := float64(totalWinSpins) / float64(stats.TotalSpins) * 100

	fmt.Printf("Winning Spins:         %d (%.2f%%)\n", totalWinSpins, hitFrequency)
	fmt.Printf("No Win:                %d (%.2f%%)\n", stats.NoWinSpins,
		float64(stats.NoWinSpins)/float64(stats.TotalSpins)*100)
	fmt.Printf("Small Wins (<5x):      %d (%.2f%%)\n", stats.SmallWins,
		float64(stats.SmallWins)/float64(stats.TotalSpins)*100)
	fmt.Printf("Medium Wins (5-20x):   %d (%.2f%%)\n", stats.MediumWins,
		float64(stats.MediumWins)/float64(stats.TotalSpins)*100)
	fmt.Printf("Big Wins (20-100x):    %d (%.2f%%)\n", stats.BigWins,
		float64(stats.BigWins)/float64(stats.TotalSpins)*100)
	fmt.Printf("Mega Wins (>100x):     %d (%.2f%%)\n", stats.MegaWins,
		float64(stats.MegaWins)/float64(stats.TotalSpins)*100)
	fmt.Printf("Avg Cascades/Win:      %.2f\n", stats.AvgCascadesPerWin)
	fmt.Printf("Max Cascades:          %d\n", stats.MaxCascades)
	fmt.Printf("Triggered:             %d times (%.4f%%)\n", stats.FreeSpinsTriggered, stats.FreeSpinsTriggeredRate)
	fmt.Printf("Avg Spins Awarded:     %.2f\n", stats.AvgFreeSpinsAwarded)
	fmt.Printf("Avg Trigger Frequency: 1 in %.0f spins\n", float64(stats.TotalSpins)/float64(stats.FreeSpinsTriggered))
	fmt.Println()

	// Max win
	fmt.Println("═══ MAX WIN ═══")
	maxWinMultiplier := stats.MaxWin / betAmount
	fmt.Printf("Max Win:               %.2f (%.1fx bet)\n", stats.MaxWin, maxWinMultiplier)
	fmt.Println()

	// Volatility indicator
	fmt.Println("═══ VOLATILITY INDICATORS ═══")
	avgWin := stats.TotalWon / float64(totalWinSpins)
	fmt.Printf("Average Win:           %.2f (%.2fx bet)\n", avgWin, avgWin/betAmount)
	fmt.Printf("Max/Avg Win Ratio:     %.1fx\n", stats.MaxWin/avgWin)
	if stats.LowSymbolWins+stats.HighSymbolWins > 0 {
		fmt.Printf("LowSymbol Win Ratio:   %0.2f\n", float64(stats.LowSymbolWins)*100/float64(stats.LowSymbolWins+stats.HighSymbolWins))
		fmt.Printf("HighSymbol Win Ratio:   %0.2f\n", float64(stats.HighSymbolWins)*100/float64(stats.LowSymbolWins+stats.HighSymbolWins))
	}

	volatility := "MEDIUM"
	if maxWinMultiplier > 500 {
		volatility = "HIGH"
	} else if maxWinMultiplier < 100 {
		volatility = "LOW"
	}
	fmt.Printf("Volatility:            %s\n", volatility)
	fmt.Println()

	totalSymbolWinCount := stats.FaWinCount + stats.ZhongWinCount + stats.BaiWinCount + stats.BawanWinCount +
		stats.WusuoWinCount + stats.WutongWinCount + stats.LiangsuoWinCount + stats.LiangtongWinCount

	// Win distribution
	fmt.Println("═══ WIN DISTRIBUTION ═══")
	fmt.Printf("FA:                    %.2f (%.2f%%)\n", stats.FaWinAmount, stats.FaWinAmount/stats.TotalWon*100)
	fmt.Printf("FA Count:              %.2f (%.2f%%)\n", stats.FaWinCount, stats.FaWinCount/float64(totalSymbolWinCount)*100)
	fmt.Printf("ZHONG:                 %.2f (%.2f%%)\n", stats.ZhongWinAmount, stats.ZhongWinAmount/stats.TotalWon*100)
	fmt.Printf("ZHONG Count:           %.2f (%.2f%%)\n", stats.ZhongWinCount, stats.ZhongWinCount/float64(totalSymbolWinCount)*100)
	fmt.Printf("BAI:                   %.2f (%.2f%%)\n", stats.BaiWinAmount, stats.BaiWinAmount/stats.TotalWon*100)
	fmt.Printf("BAI Count:             %.2f (%.2f%%)\n", stats.BaiWinCount, stats.BaiWinCount/float64(totalSymbolWinCount)*100)
	fmt.Printf("BAWAN:                 %.2f (%.2f%%)\n", stats.BawanWinAmount, stats.BawanWinAmount/stats.TotalWon*100)
	fmt.Printf("BAWAN Count:           %.2f (%.2f%%)\n", stats.BawanWinCount, stats.BawanWinCount/float64(totalSymbolWinCount)*100)
	fmt.Printf("WUSUO:                 %.2f (%.2f%%)\n", stats.WusuoWinAmount, stats.WusuoWinAmount/stats.TotalWon*100)
	fmt.Printf("WUSUO Count:           %.2f (%.2f%%)\n", stats.WusuoWinCount, stats.WusuoWinCount/float64(totalSymbolWinCount)*100)
	fmt.Printf("WUTONG:                %.2f (%.2f%%)\n", stats.WutongWinAmount, stats.WutongWinAmount/stats.TotalWon*100)
	fmt.Printf("WUTONG Count:          %.2f (%.2f%%)\n", stats.WutongWinCount, stats.WutongWinCount/float64(totalSymbolWinCount)*100)
	fmt.Printf("LIANGSUO:              %.2f (%.2f%%)\n", stats.LiangsuoWinAmount, stats.LiangsuoWinAmount/stats.TotalWon*100)
	fmt.Printf("LIANGSUO Count:        %.2f (%.2f%%)\n", stats.LiangsuoWinCount, stats.LiangsuoWinCount/float64(totalSymbolWinCount)*100)
	fmt.Printf("LIANGTONG:             %.2f (%.2f%%)\n", stats.LiangtongWinAmount, stats.LiangtongWinAmount/stats.TotalWon*100)
	fmt.Printf("LIANGTONG Count:       %.2f (%.2f%%)\n", stats.LiangtongWinCount, stats.LiangtongWinCount/float64(totalSymbolWinCount)*100)
	fmt.Println()

	// Symbol RTP contribution by category (per base_rtp_md)
	fmt.Println("═══ SYMBOL RTP CONTRIBUTION ═══")
	fmt.Printf("Low Symbols (liangtong, liangsuo, wusuo, wutong):  %.2f%% (target: 60-70%%)\n", stats.LowSymbolRTPPct)
	fmt.Printf("Mid Symbols (bawan, bai):                          %.2f%% (target: 25-30%%)\n", stats.MidSymbolRTPPct)
	fmt.Printf("High Symbols (zhong, fa):                          %.2f%% (target: <10%%)\n", stats.HighSymbolRTPPct)

	fmt.Println()
}
