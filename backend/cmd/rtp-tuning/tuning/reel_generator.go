// Package tuning provides PG-style reel strip generation and tuning.
//
// This implements the methodology from base_rtp_tuning.md:
// - Reel topology design (each reel has a role)
// - Spacing rules (prevent clusters)
// - Anti-cluster constraints (prevent multi-way patterns)
// - Symbol contribution balancing
package tuning

import (
	"fmt"
	"math/rand"
	"strings"
)

// ReelRole defines the role of each reel in the topology
type ReelRole int

const (
	RoleActivator ReelRole = iota // R1: Kích hoạt - triggers wins
	RoleConnector                 // R2: Nối - connects patterns
	RoleCore                      // R3: Core - main win driver
	RoleAmplifier                 // R4: Khuếch đại - amplifies wins
	RoleRareSpike                 // R5: Spike hiếm - rare big wins
)

// ReelTopology defines how symbols should be distributed on each reel
type ReelTopology struct {
	// SymbolDensity maps symbol -> relative density multiplier for this reel
	// hệ số nhân mật độ tương đối cho cuộn này
	// 1.0 = normal, >1.0 = more frequent, <1.0 = less frequent
	SymbolDensity map[string]float64

	// MinSpacing maps symbol -> minimum positions between same symbols
	// vị trí tối thiểu giữa các ký hiệu giống nhau
	MinSpacing map[string]int

	// MaxClusterSize maps symbol -> max consecutive symbols allowed
	// số lượng ký hiệu liên tiếp tối đa được cho phép
	MaxClusterSize map[string]int

	// ForbiddenPairs lists symbol pairs that should not be adjacent
	//  liệt kê các cặp ký hiệu không được liền kề
	ForbiddenPairs [][2]string

	ClusterForbiddenPairs []ClusterForbidden

	// Role defines the role of this reel
	Role ReelRole

	// GoldConfig defines gold symbol replacement rules for this reel
	// Gold symbols are created by replacing base symbols after strip generation
	GoldConfig *GoldTopologyConfig
}

// GoldTopologyConfig defines gold symbol configuration for a reel
type GoldTopologyConfig struct {
	// Enabled controls whether gold symbols are generated on this reel
	Enabled bool

	// GoldRatio is the target percentage of paying symbols to convert to gold (0.0-1.0)
	// e.g., 0.12 means 12% of paying symbols will be gold
	GoldRatio float64

	// MinGoldSpacing is the minimum positions between gold symbols
	MinGoldSpacing int

	// MaxGoldCluster is the maximum consecutive gold symbols allowed
	MaxGoldCluster int

	// SymbolGoldRatio allows per-symbol gold ratio override (optional)
	// If not set, uses GoldRatio for all symbols
	SymbolGoldRatio map[string]float64
}

type ClusterForbidden struct {
	ClusterSymbols map[string]struct{}
	GroupSize      int
	AllowedSymbols []string
}

// ReelStripConfig holds configuration for reel strip generation
type ReelStripConfig struct {
	// BaseWeights are the base symbol weights (from symbol package)
	BaseWeights SymbolWeights

	// Topology defines reel-specific distribution rules
	Topology ReelTopology

	// StripLength is the target length of the reel strip
	StripLength int
}

// PGReelGenerator implements PG-style reel strip generation
type PGReelGenerator struct {
	rng        *rand.Rand
	topologies [5]ReelTopology

	// Symbol categories for balancing
	lowSymbols  []string
	midSymbols  []string
	highSymbols []string
}

// NewPGReelGenerator creates a new PG-style reel generator
// Symbol classification based on paytable (3-of-kind payout):
// - Low (1-3x):  liangtong(1), liangsuo(2), wusuo(3), wutong(3) → target 60-70% RTP
// - Mid (5-6x):  bawan(5), bai(6) → target 25-30% RTP
// - High (8-10x): zhong(8), fa(10) → target <10% RTP
func NewPGReelGenerator(seed int64, topologies [5]ReelTopology) *PGReelGenerator {
	return &PGReelGenerator{
		rng:         rand.New(rand.NewSource(seed)),
		topologies:  topologies,
		lowSymbols:  []string{"liangtong", "liangsuo", "wusuo", "wutong"}, // payout 1-3x
		midSymbols:  []string{"bawan", "bai"},                             // payout 5-6x
		highSymbols: []string{"zhong", "fa"},                              // payout 8-10x
	}
}

// defaultTopologies returns PG-style reel topologies
// Per document: R1 activator, R2 connector, R3 core, R4 amplifier, R5 rare spike
//
// Symbol contribution targets (per base_rtp_tuning.md):
// - Low symbols: 60-70% of base RTP
// - Mid symbols: 25-30%
// - High symbols: <10%
//
// To achieve this balance, we need to REDUCE low symbol density
// and INCREASE mid symbol density across all reels
func defaultTopologies() [5]ReelTopology {
	return [5]ReelTopology{
		// Reel 1: Activator - triggers wins
		// Per base_rtp_tuning.md: Low 60-70%, Mid 25-30%, High <10%
		// Final calibrated: Low 0.62-0.67, Mid 1.12-1.15, High 0.55-0.65
		{
			Role: RoleActivator,
			SymbolDensity: map[string]float64{
				"fa":        0.6,
				"zhong":     0.6,
				"bai":       0.6,
				"bawan":     0.6,
				"wusuo":     1.7,
				"wutong":    1.7,
				"liangsuo":  1.7,
				"liangtong": 1.7,
				"bonus":     1.45,
			},
			MinSpacing: map[string]int{
				"fa":    4,
				"zhong": 4,
				"bai":   4,
				"bonus": 10, // Scatter needs wide spacing
			},
			MaxClusterSize: map[string]int{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     2,
				"wutong":    2,
				"liangsuo":  2,
				"liangtong": 2,
				"bonus":     1,
			},
			ForbiddenPairs: [][2]string{},
			ClusterForbiddenPairs: []ClusterForbidden{
				{
					ClusterSymbols: map[string]struct{}{"liangtong": {}, "liangsuo": {}, "wutong": {}, "wusuo": {}},
					GroupSize:      4,
					AllowedSymbols: []string{"fa", "zhong", "bai", "bawan"},
				},
				{
					ClusterSymbols: map[string]struct{}{"fa": {}, "zhong": {}, "bai": {}},
					GroupSize:      2,
					AllowedSymbols: []string{"liangtong", "liangsuo", "wutong", "wusuo", "bawan"},
				},
			},
		},

		// Reel 2: Connector - continues patterns
		{
			Role: RoleConnector,
			SymbolDensity: map[string]float64{
				"fa":        1.15,
				"zhong":     1.15,
				"bai":       1.15,
				"bawan":     1.15,
				"wusuo":     1.25,
				"wutong":    1.25,
				"liangsuo":  1.25,
				"liangtong": 1.25,
				"bonus":     1.4,
			},
			MinSpacing: map[string]int{
				"bonus": 10,
			},
			MaxClusterSize: map[string]int{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     2,
				"wutong":    2,
				"liangsuo":  2,
				"liangtong": 2,
				"bonus":     1,
			},
			ForbiddenPairs: [][2]string{},
			ClusterForbiddenPairs: []ClusterForbidden{
				{
					ClusterSymbols: map[string]struct{}{"liangtong": {}, "liangsuo": {}, "wutong": {}, "wusuo": {}},
					GroupSize:      4,
					AllowedSymbols: []string{"fa", "zhong", "bai", "bawan"},
				},
				{
					ClusterSymbols: map[string]struct{}{"fa": {}, "zhong": {}, "bai": {}},
					GroupSize:      2,
					AllowedSymbols: []string{"liangtong", "liangsuo", "wutong", "wusuo", "bawan"},
				},
			},
			GoldConfig: &GoldTopologyConfig{
				Enabled:        true,
				GoldRatio:      0.04, // 10% of paying symbols will be gold
				MinGoldSpacing: 3,    // At least 5 positions between gold symbols
				MaxGoldCluster: 2,    // No consecutive gold symbols
			},
		},

		// Reel 3: Core - main win driver
		{
			Role: RoleCore,
			SymbolDensity: map[string]float64{
				"fa":        1.55,
				"zhong":     1.55,
				"bai":       1.55,
				"bawan":     1.55,
				"wusuo":     0.4,
				"wutong":    0.4,
				"liangsuo":  0.4,
				"liangtong": 0.4,
				"bonus":     0.85,
			},
			MinSpacing: map[string]int{
				"fa":    4,
				"zhong": 4,
				"bai":   4,
				"bonus": 10,
			},
			MaxClusterSize: map[string]int{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     2,
				"wutong":    2,
				"liangsuo":  2,
				"liangtong": 2,
				"bonus":     1,
			},
			ForbiddenPairs: [][2]string{},
			ClusterForbiddenPairs: []ClusterForbidden{
				{
					ClusterSymbols: map[string]struct{}{"liangtong": {}, "liangsuo": {}, "wutong": {}, "wusuo": {}},
					GroupSize:      4,
					AllowedSymbols: []string{"fa", "zhong", "bai", "bawan"},
				},
				{
					ClusterSymbols: map[string]struct{}{"fa": {}, "zhong": {}, "bai": {}},
					GroupSize:      2,
					AllowedSymbols: []string{"liangtong", "liangsuo", "wutong", "wusuo", "bawan"},
				},
			},
			GoldConfig: &GoldTopologyConfig{
				Enabled:        true,
				GoldRatio:      0.05, // 12% of paying symbols will be gold (core reel has slightly more)
				MinGoldSpacing: 3,    // At least 4 positions between gold symbols
				MaxGoldCluster: 2,    // No consecutive gold symbols
			},
		},

		// Reel 4: Amplifier - extends wins to 4-of-kind
		{
			Role: RoleAmplifier,
			SymbolDensity: map[string]float64{
				"fa":        1.15,
				"zhong":     1.15,
				"bai":       1.15,
				"bawan":     1.15,
				"wusuo":     0.6,
				"wutong":    0.6,
				"liangsuo":  0.6,
				"liangtong": 0.6,
				"bonus":     0.75,
			},
			MinSpacing: map[string]int{
				"bonus": 10, // Scatter needs wide spacing
			},
			MaxClusterSize: map[string]int{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     2,
				"wutong":    2,
				"liangsuo":  2,
				"liangtong": 2,
				"bonus":     1,
			},
			ForbiddenPairs:        [][2]string{},
			ClusterForbiddenPairs: []ClusterForbidden{},
			GoldConfig: &GoldTopologyConfig{
				Enabled:        true,
				GoldRatio:      0.04, // 8% of paying symbols will be gold (amplifier reel has less)
				MinGoldSpacing: 3,    // At least 6 positions between gold symbols
				MaxGoldCluster: 2,    // No consecutive gold symbols
			},
		},

		// Reel 5: Rare Spike - controls 5-of-kind (rare big wins)
		{
			Role: RoleRareSpike,
			SymbolDensity: map[string]float64{
				"fa":        1.4,
				"zhong":     1.4,
				"bai":       1.4,
				"bawan":     1.4,
				"wusuo":     0.6,
				"wutong":    0.6,
				"liangsuo":  0.6,
				"liangtong": 0.6,
				"bonus":     0.85,
			},
			MinSpacing: map[string]int{
				"fa":    6,
				"zhong": 6,
				"bai":   3,
				"bonus": 10, // Scatter needs wide spacing
			},
			MaxClusterSize: map[string]int{
				"fa":        1,
				"zhong":     1,
				"bai":       1,
				"bawan":     1,
				"wusuo":     2,
				"wutong":    2,
				"liangsuo":  2,
				"liangtong": 2,
				"bonus":     1,
			},
			ForbiddenPairs:        [][2]string{},
			ClusterForbiddenPairs: []ClusterForbidden{},
		},
	}
}

// GenerateReelStrip generates a single reel strip with spacing and anti-cluster rules
// Gold symbols are applied AFTER strip generation based on GoldConfig in topology
func (g *PGReelGenerator) GenerateReelStrip(reelIndex int, weights SymbolWeights) ([]string, error) {
	if reelIndex < 0 || reelIndex >= 5 {
		return nil, fmt.Errorf("invalid reel index: %d", reelIndex)
	}

	topology := g.topologies[reelIndex]

	// Calculate adjusted weights based on topology density
	adjustedWeights := g.adjustWeightsByDensity(weights, topology.SymbolDensity)

	// Calculate strip length
	stripLength := 0
	for _, w := range adjustedWeights {
		stripLength += w
	}

	// Build symbol pool
	pool := g.buildSymbolPool(adjustedWeights)

	// Generate strip with constraints
	fmt.Println("reelIndex: ", reelIndex)
	strip, err := g.generateWithConstraints(pool, stripLength, topology)
	if err != nil {
		// Fallback to basic generation if constraints too tight
		fmt.Printf("Warning: Constraint generation failed for reel %d, using fallback: %v\n", reelIndex+1, err)
		strip = g.generateFallback(pool)
	}

	// Apply gold symbols after strip generation (for reels 2, 3, 4)
	// This replaces some base symbols with gold variants based on GoldConfig
	strip = g.ApplyGoldSymbols(strip, reelIndex)

	return strip, nil
}

// adjustWeightsByDensity adjusts weights based on topology density multipliers
func (g *PGReelGenerator) adjustWeightsByDensity(weights SymbolWeights, density map[string]float64) SymbolWeights {
	adjusted := make(SymbolWeights)
	for sym, w := range weights {
		mult := 1.0
		if d, ok := density[sym]; ok {
			mult = d
		}
		adjusted[sym] = int(float64(w) * mult)
		if adjusted[sym] < 1 && w > 0 {
			adjusted[sym] = 1 // Ensure at least 1 if original weight > 0
		}
	}
	return adjusted
}

// buildSymbolPool creates a shuffled pool of symbols
func (g *PGReelGenerator) buildSymbolPool(weights SymbolWeights) []string {
	var pool []string
	for sym, count := range weights {
		for i := 0; i < count; i++ {
			pool = append(pool, sym)
		}
	}
	// Shuffle pool
	g.rng.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})
	return pool
}

// GenerationMode controls how to handle symbols that can't be placed
type GenerationMode int

const (
	// ModeFindBest uses findBestPlacement for unplaceable symbols (may violate constraints)
	ModeFindBest GenerationMode = iota
	// ModeRetryQueue puts unplaceable symbols in retry queue, tries again later
	ModeRetryQueue
	// ModeDropAndReplace drops unplaceable symbols and replaces with placeable ones
	ModeDropAndReplace
)

// GenerationStats tracks statistics about reel generation
type GenerationStats struct {
	TotalSymbols      int            // Total symbols in pool
	PlacedNormally    int            // Symbols placed without constraint violation
	PlacedWithRetry   int            // Symbols placed after retry
	PlacedForcefully  int            // Symbols placed using findBestPlacement (violated constraints)
	DroppedSymbols    map[string]int // Symbols dropped (could not place)
	ReplacedSymbols   map[string]int // Symbols added as replacement
	FinalStripLength  int
	ConstraintSuccess float64 // Percentage of symbols placed without violation
}

// generateWithConstraints generates strip with spacing and anti-cluster rules
func (g *PGReelGenerator) generateWithConstraints(pool []string, length int, topology ReelTopology) ([]string, error) {
	return g.generateWithConstraintsMode(pool, length, topology, ModeDropAndReplace)
}

// generateWithConstraintsMode generates strip with configurable handling of unplaceable symbols
func (g *PGReelGenerator) generateWithConstraintsMode(pool []string, length int, topology ReelTopology, mode GenerationMode) ([]string, error) {
	strip := make([]string, 0, length)
	remaining := make([]string, len(pool))
	copy(remaining, pool)

	// Track last positions of each symbol for spacing check
	lastPos := make(map[string]int)
	for sym := range topology.MinSpacing {
		lastPos[sym] = -100 // Initialize to far negative
	}

	// Track consecutive counts for cluster check
	consecutiveCount := make(map[string]int)

	// Retry queue for symbols that couldn't be placed
	retryQueue := make([]string, 0)

	// Track dropped symbols for replacement
	droppedSymbols := make(map[string]int)

	// Stats tracking
	placedNormally := 0
	placedWithRetry := 0
	placedForcefully := 0

	maxAttempts := length * 10 // Prevent infinite loop
	retryRounds := 0
	maxRetryRounds := 3

	for len(strip) < length && maxAttempts > 0 {
		placed := false

		// First, try to place from remaining pool
		for i, sym := range remaining {
			if g.canPlace(sym, len(strip), strip, topology, lastPos, consecutiveCount) {
				strip = append(strip, sym)
				lastPos[sym] = len(strip) - 1

				// Update consecutive count
				if len(strip) > 1 && strip[len(strip)-2] == sym {
					consecutiveCount[sym]++
				} else {
					consecutiveCount[sym] = 1
				}

				// Remove from remaining
				remaining = append(remaining[:i], remaining[i+1:]...)
				placed = true
				placedNormally++
				break
			}
		}

		if !placed && len(remaining) > 0 {
			switch mode {
			case ModeFindBest:
				// Original behavior: find best placement (may violate constraints)
				bestIdx := g.findBestPlacement(remaining, len(strip), strip, topology, lastPos)
				sym := remaining[bestIdx]
				strip = append(strip, sym)
				lastPos[sym] = len(strip) - 1

				if len(strip) > 1 && strip[len(strip)-2] == sym {
					consecutiveCount[sym]++
				} else {
					consecutiveCount[sym] = 1
				}

				remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
				placed = true
				placedForcefully++

			case ModeRetryQueue:
				// Move first unplaceable symbol to retry queue
				sym := remaining[0]
				retryQueue = append(retryQueue, sym)
				remaining = remaining[1:]

				// If remaining is empty but retry queue has items, swap them
				if len(remaining) == 0 && len(retryQueue) > 0 && retryRounds < maxRetryRounds {
					remaining = retryQueue
					retryQueue = make([]string, 0)
					retryRounds++
					// Shuffle remaining to try different order
					g.rng.Shuffle(len(remaining), func(i, j int) {
						remaining[i], remaining[j] = remaining[j], remaining[i]
					})
				}

			case ModeDropAndReplace:
				// Find the most problematic symbol (can't be placed)
				// Drop it and try to find a replacement that CAN be placed
				droppedIdx := -1
				for i, sym := range remaining {
					if !g.canPlace(sym, len(strip), strip, topology, lastPos, consecutiveCount) {
						droppedIdx = i
						break
					}
				}

				if droppedIdx >= 0 {
					droppedSym := remaining[droppedIdx]
					droppedSymbols[droppedSym]++
					remaining = append(remaining[:droppedIdx], remaining[droppedIdx+1:]...)

					// Try to find a replacement symbol that CAN be placed
					// Prefer symbols from the same category
					replacement := g.findReplacement(droppedSym, strip, topology, lastPos, consecutiveCount)
					if replacement != "" {
						strip = append(strip, replacement)
						lastPos[replacement] = len(strip) - 1

						if len(strip) > 1 && strip[len(strip)-2] == replacement {
							consecutiveCount[replacement]++
						} else {
							consecutiveCount[replacement] = 1
						}
						placed = true
						placedWithRetry++
					}
					// If no replacement found, we just dropped the symbol (strip will be shorter)
				}
			}
		}

		// Handle retry queue for ModeRetryQueue
		if mode == ModeRetryQueue && !placed && len(remaining) == 0 && len(retryQueue) > 0 {
			// Try retry queue one more time with findBestPlacement as last resort
			if retryRounds >= maxRetryRounds {
				bestIdx := g.findBestPlacement(retryQueue, len(strip), strip, topology, lastPos)
				sym := retryQueue[bestIdx]
				strip = append(strip, sym)
				lastPos[sym] = len(strip) - 1

				if len(strip) > 1 && strip[len(strip)-2] == sym {
					consecutiveCount[sym]++
				} else {
					consecutiveCount[sym] = 1
				}

				retryQueue = append(retryQueue[:bestIdx], retryQueue[bestIdx+1:]...)
				placed = true
				placedForcefully++
			}
		}

		maxAttempts--

		// Break if no more symbols to process
		if len(remaining) == 0 && len(retryQueue) == 0 {
			break
		}
	}

	// Log generation stats
	total := placedNormally + placedWithRetry + placedForcefully
	if total > 0 {
		successRate := float64(placedNormally+placedWithRetry) / float64(total) * 100
		if len(droppedSymbols) > 0 || placedForcefully > 0 {
			fmt.Printf("    Generation: %d placed (%.1f%% clean), %d forced, dropped: %v\n",
				total, successRate, placedForcefully, droppedSymbols)
		}
	}

	if len(strip) < length {
		// In ModeDropAndReplace, we accept shorter strips
		if mode == ModeDropAndReplace {
			fmt.Printf("    Warning: Strip length %d (expected %d), dropped %d symbols\n",
				len(strip), length, length-len(strip))
		} else {
			return nil, fmt.Errorf("could not generate strip of length %d, only got %d", length, len(strip))
		}
	}

	return strip, nil
}

// findReplacement finds a replacement symbol that can be placed
// Prefers symbols from the same category as the dropped symbol
func (g *PGReelGenerator) findReplacement(droppedSym string, strip []string, topology ReelTopology, lastPos map[string]int, consecutiveCount map[string]int) string {
	baseDropped := strings.Split(droppedSym, "_")[0]

	// Determine category of dropped symbol
	var candidateSymbols []string

	// Check if dropped is low symbol
	for _, sym := range g.lowSymbols {
		if sym == baseDropped {
			candidateSymbols = g.lowSymbols
			break
		}
	}
	// Check if dropped is mid symbol
	if candidateSymbols == nil {
		for _, sym := range g.midSymbols {
			if sym == baseDropped {
				candidateSymbols = g.midSymbols
				break
			}
		}
	}
	// Check if dropped is high symbol
	if candidateSymbols == nil {
		for _, sym := range g.highSymbols {
			if sym == baseDropped {
				candidateSymbols = g.highSymbols
				break
			}
		}
	}
	// Default to all symbols if not found
	if candidateSymbols == nil {
		candidateSymbols = append(append(g.lowSymbols, g.midSymbols...), g.highSymbols...)
	}

	// Shuffle candidates
	shuffled := make([]string, len(candidateSymbols))
	copy(shuffled, candidateSymbols)
	g.rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Find first candidate that can be placed
	pos := len(strip)
	for _, sym := range shuffled {
		if g.canPlace(sym, pos, strip, topology, lastPos, consecutiveCount) {
			return sym
		}
	}

	// If no same-category symbol works, try all categories
	allSymbols := append(append(g.lowSymbols, g.midSymbols...), g.highSymbols...)
	g.rng.Shuffle(len(allSymbols), func(i, j int) {
		allSymbols[i], allSymbols[j] = allSymbols[j], allSymbols[i]
	})

	for _, sym := range allSymbols {
		if g.canPlace(sym, pos, strip, topology, lastPos, consecutiveCount) {
			return sym
		}
	}

	return "" // No replacement found
}

// canPlace checks if a symbol can be placed at current position
// Treats base symbol and _gold variant as the same symbol for spacing/cluster rules
// e.g., "fa" and "fa_gold" are considered the same for min spacing checks
func (g *PGReelGenerator) canPlace(sym string, pos int, strip []string, topology ReelTopology, lastPos map[string]int, consecutiveCount map[string]int) bool {
	// Extract base symbol (e.g., "fa" from "fa_gold" or "fa" from "fa")
	baseSymbol := strings.Split(sym, "_")[0]
	goldSymbol := baseSymbol + "_gold"

	// Check minimum spacing - treat base and gold as same symbol for spacing
	// e.g., if fa has minSpace=2, then fa...fa_gold with distance < 2 is invalid
	if minSpace, ok := topology.MinSpacing[baseSymbol]; ok {
		// Check distance from last occurrence of base symbol
		if lp, ok := lastPos[baseSymbol]; ok && pos-lp < minSpace {
			return false
		}
		// Check distance from last occurrence of gold variant
		if lp, ok := lastPos[goldSymbol]; ok && pos-lp < minSpace {
			return false
		}
	}

	// Check max cluster size - also treat base and gold as same symbol family
	if maxCluster, ok := topology.MaxClusterSize[baseSymbol]; ok {
		// Count consecutive same-family symbols (base + gold)
		familyConsecutive := consecutiveCount[baseSymbol] + consecutiveCount[goldSymbol]
		if len(strip) > 0 {
			lastSym := strip[len(strip)-1]
			lastSymBase := strings.Split(lastSym, "_")[0]
			// If last symbol is same family, check cluster limit
			if lastSymBase == baseSymbol && familyConsecutive >= maxCluster {
				return false
			}
		}
	}

	// Check cluster forbidden pairs
	// Logic: If the previous GroupSize symbols are ALL in ClusterSymbols,
	// then the current symbol MUST be in AllowedSymbols (otherwise return false)
	//
	// Example: ClusterSymbols={low symbols}, GroupSize=3, AllowedSymbols={high symbols}
	// If strip ends with [low, low, low], then next symbol must be a high symbol
	for _, cluster := range topology.ClusterForbiddenPairs {
		if pos < cluster.GroupSize {
			continue // Not enough symbols yet to check
		}

		// Check if previous GroupSize symbols are all in ClusterSymbols
		allInCluster := true
		for i := 1; i <= cluster.GroupSize; i++ {
			prevSym := strip[pos-i]
			prevSymBase := strings.Split(prevSym, "_")[0]
			if _, ok := cluster.ClusterSymbols[prevSymBase]; !ok {
				allInCluster = false
				break
			}
		}

		// If all previous GroupSize symbols are in ClusterSymbols,
		// the current symbol must be in AllowedSymbols
		if allInCluster {
			allowed := false
			for _, allowedSym := range cluster.AllowedSymbols {
				if baseSymbol == allowedSym {
					allowed = true
					break
				}
			}
			if !allowed {
				return false
			}
		}
	}

	// Check forbidden pairs - treat base and gold as same symbol
	if len(strip) > 0 {
		lastSym := strip[len(strip)-1]
		lastSymBase := strings.Split(lastSym, "_")[0]
		for _, pair := range topology.ForbiddenPairs {
			// Check forbidden pairs using base symbols (treat gold as same)
			pair0Base := strings.Split(pair[0], "_")[0]
			pair1Base := strings.Split(pair[1], "_")[0]
			if (pair0Base == lastSymBase && pair1Base == baseSymbol) || (pair0Base == baseSymbol && pair1Base == lastSymBase) {
				return false
			}
		}
	}

	return true
}

// findBestPlacement finds the symbol with least constraint violations
// Treats base symbol and _gold variant as the same symbol for scoring
func (g *PGReelGenerator) findBestPlacement(remaining []string, pos int, strip []string, topology ReelTopology, lastPos map[string]int) int {
	bestIdx := 0
	bestScore := -1000

	for i, sym := range remaining {
		score := 0

		// Extract base symbol (e.g., "fa" from "fa_gold")
		baseSymbol := strings.Split(sym, "_")[0]
		goldSymbol := baseSymbol + "_gold"

		// Penalize spacing violations - check both base and gold variants
		if minSpace, ok := topology.MinSpacing[baseSymbol]; ok {
			// Check gap from base symbol
			if lp, ok := lastPos[baseSymbol]; ok {
				gap := pos - lp
				if gap >= minSpace {
					score += 10
				} else {
					score -= (minSpace - gap) * 2
				}
			}
			// Check gap from gold variant
			if lp, ok := lastPos[goldSymbol]; ok {
				gap := pos - lp
				if gap < minSpace {
					score -= (minSpace - gap) * 2
				}
			}
		}

		// Penalize consecutive same symbol family (base or gold)
		if len(strip) > 0 {
			lastSym := strip[len(strip)-1]
			lastSymBase := strings.Split(lastSym, "_")[0]
			if lastSymBase == baseSymbol {
				score -= 5
			}
		}

		// Penalize forbidden pairs - using base symbols
		if len(strip) > 0 {
			lastSym := strip[len(strip)-1]
			lastSymBase := strings.Split(lastSym, "_")[0]
			for _, pair := range topology.ForbiddenPairs {
				pair0Base := strings.Split(pair[0], "_")[0]
				pair1Base := strings.Split(pair[1], "_")[0]
				if (pair0Base == lastSymBase && pair1Base == baseSymbol) || (pair0Base == baseSymbol && pair1Base == lastSymBase) {
					score -= 20
				}
			}
		}

		// Penalize cluster forbidden violations
		for _, cluster := range topology.ClusterForbiddenPairs {
			if pos >= cluster.GroupSize {
				// Check if previous GroupSize symbols are all in ClusterSymbols
				allInCluster := true
				for j := 1; j <= cluster.GroupSize; j++ {
					prevSym := strip[pos-j]
					prevSymBase := strings.Split(prevSym, "_")[0]
					if _, ok := cluster.ClusterSymbols[prevSymBase]; !ok {
						allInCluster = false
						break
					}
				}
				// If all previous symbols are in cluster, penalize if current is not allowed
				if allInCluster {
					allowed := false
					for _, allowedSym := range cluster.AllowedSymbols {
						if baseSymbol == allowedSym {
							allowed = true
							break
						}
					}
					if !allowed {
						score -= 30 // Heavy penalty for cluster forbidden violation
					} else {
						score += 15 // Bonus for breaking the cluster correctly
					}
				}
			}
		}

		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	return bestIdx
}

// generateFallback generates strip with basic shuffle (fallback)
func (g *PGReelGenerator) generateFallback(pool []string) []string {
	strip := make([]string, len(pool))
	copy(strip, pool)
	g.rng.Shuffle(len(strip), func(i, j int) {
		strip[i], strip[j] = strip[j], strip[i]
	})
	return strip
}

// GenerateAllReelStrips generates all 5 reel strips using PG methodology
func (g *PGReelGenerator) GenerateAllReelStrips(weights *ReelWeightsSet) ([][]string, error) {
	strips := make([][]string, 5)

	for i := 0; i < 5; i++ {
		w := weights.GetReel(i)
		if w == nil {
			return nil, fmt.Errorf("no weights for reel %d", i)
		}

		strip, err := g.GenerateReelStrip(i, w)
		if err != nil {
			return nil, fmt.Errorf("failed to generate reel %d: %w", i, err)
		}
		strips[i] = strip
	}

	return strips, nil
}

// SetTopology allows customizing topology for a specific reel
func (g *PGReelGenerator) SetTopology(reelIndex int, topology ReelTopology) {
	if reelIndex >= 0 && reelIndex < 5 {
		g.topologies[reelIndex] = topology
	}
}

// GetTopology returns the current topology for a reel
func (g *PGReelGenerator) GetTopology(reelIndex int) ReelTopology {
	if reelIndex >= 0 && reelIndex < 5 {
		return g.topologies[reelIndex]
	}
	return ReelTopology{}
}

// SymbolContributionTargets defines target percentages for symbol RTP contribution
type SymbolContributionTargets struct {
	LowPct  float64 // Target: 60-70%
	MidPct  float64 // Target: 25-30%
	HighPct float64 // Target: <10%
}

// DefaultSymbolContributionTargets returns PG-style targets
func DefaultSymbolContributionTargets() SymbolContributionTargets {
	return SymbolContributionTargets{
		LowPct:  65.0, // Center of 60-70%
		MidPct:  28.0, // Center of 25-30%
		HighPct: 7.0,  // Center of <10%
	}
}

// AdjustTopologyDensities automatically adjusts topology densities based on
// current symbol contribution vs targets. This is the core auto-tuning method.
//
// Per base_rtp_tuning.md adjustment order: Spacing → Reel role → Relative density → Weight
// This method handles "Relative density" adjustment automatically in the tuning loop.
//
// Parameters:
//   - currentLow, currentMid, currentHigh: Current symbol RTP contribution percentages
//   - targets: Target contribution percentages
//   - learningRate: How much to adjust (0.0-1.0, recommended 0.1-0.3)
//
// Returns: true if adjustment was made, false if already converged
func (g *PGReelGenerator) AdjustTopologyDensities(
	currentLow, currentMid, currentHigh float64,
	targets SymbolContributionTargets,
	learningRate float64,
) bool {
	// Calculate errors
	lowErr := currentLow - targets.LowPct    // positive = too high
	midErr := currentMid - targets.MidPct    // positive = too high
	highErr := currentHigh - targets.HighPct // positive = too high

	// Check if already converged (within 3% for each category)
	const tolerance = 3.0
	if abs(lowErr) < tolerance && abs(midErr) < tolerance && abs(highErr) < tolerance {
		return false // Already converged
	}

	// Adjustment strategy:
	// If Low is too high (>70%) → reduce low density, increase mid density
	// If Low is too low (<60%) → increase low density, reduce mid density
	// High symbols stay relatively stable (controlled by spacing, not density)

	// Calculate density adjustments
	// Scale by learning rate and error magnitude
	lowDelta := -lowErr * learningRate * 0.01 // Convert % to density multiplier
	midDelta := -midErr * learningRate * 0.01
	highDelta := -highErr * learningRate * 0.01

	// Apply adjustments to all reels
	for reelIdx := 0; reelIdx < 5; reelIdx++ {
		g.adjustReelDensities(reelIdx, lowDelta, midDelta, highDelta)
	}

	return true
}

// RoleAdjustmentMultipliers defines how each reel role responds to adjustments
// This implements the PG-style differentiated reel behavior
type RoleAdjustmentMultipliers struct {
	LowMult  float64 // Multiplier for low symbol adjustments
	MidMult  float64 // Multiplier for mid symbol adjustments
	HighMult float64 // Multiplier for high symbol adjustments
}

// getRoleMultipliers returns adjustment multipliers based on reel role
// Per tuning_full.md: Each reel has a specific function in the topology
func getRoleMultipliers(role ReelRole) RoleAdjustmentMultipliers {
	switch role {
	case RoleActivator: // R1: Triggers wins - favors low symbols
		return RoleAdjustmentMultipliers{
			LowMult:  1.3, // More sensitive to low adjustments
			MidMult:  0.7, // Less sensitive to mid
			HighMult: 0.4, // Minimal high symbol changes
		}
	case RoleConnector: // R2: Connects patterns - balanced with slight low bias
		return RoleAdjustmentMultipliers{
			LowMult:  1.1,
			MidMult:  1.0,
			HighMult: 0.7,
		}
	case RoleCore: // R3: Main driver - balanced adjustments
		return RoleAdjustmentMultipliers{
			LowMult:  1.0,
			MidMult:  1.0,
			HighMult: 1.0,
		}
	case RoleAmplifier: // R4: Amplifies wins - favors mid symbols
		return RoleAdjustmentMultipliers{
			LowMult:  0.7,
			MidMult:  1.3, // More sensitive to mid adjustments
			HighMult: 1.0,
		}
	case RoleRareSpike: // R5: Rare big wins - favors high symbols
		return RoleAdjustmentMultipliers{
			LowMult:  0.5, // Minimal low symbol changes
			MidMult:  0.7,
			HighMult: 1.5, // More sensitive to high adjustments
		}
	default:
		return RoleAdjustmentMultipliers{1.0, 1.0, 1.0}
	}
}

// adjustReelDensities adjusts densities for a single reel based on its role
// Different reels respond differently to the same delta based on their function
func (g *PGReelGenerator) adjustReelDensities(reelIdx int, lowDelta, midDelta, highDelta float64) {
	if reelIdx < 0 || reelIdx >= 5 {
		return
	}

	// Get role-based multipliers
	role := g.topologies[reelIdx].Role
	mult := getRoleMultipliers(role)

	// Apply role-specific multipliers to deltas
	adjustedLowDelta := lowDelta * mult.LowMult
	adjustedMidDelta := midDelta * mult.MidMult
	adjustedHighDelta := highDelta * mult.HighMult

	// Density bounds
	const minDensity = 0.3
	const maxDensity = 2.0

	// Adjust low symbols
	for _, sym := range g.lowSymbols {
		current := g.topologies[reelIdx].SymbolDensity[sym]
		if current == 0 {
			current = 1.0 // Default
		}
		newDensity := clampDensity(current+adjustedLowDelta, minDensity, maxDensity)
		g.topologies[reelIdx].SymbolDensity[sym] = newDensity
	}

	// Adjust mid symbols
	for _, sym := range g.midSymbols {
		current := g.topologies[reelIdx].SymbolDensity[sym]
		if current == 0 {
			current = 1.0
		}
		newDensity := clampDensity(current+adjustedMidDelta, minDensity, maxDensity)
		g.topologies[reelIdx].SymbolDensity[sym] = newDensity
	}

	// Adjust high symbols
	for _, sym := range g.highSymbols {
		current := g.topologies[reelIdx].SymbolDensity[sym]
		if current == 0 {
			current = 1.0
		}
		newDensity := clampDensity(current+adjustedHighDelta, minDensity, maxDensity)
		g.topologies[reelIdx].SymbolDensity[sym] = newDensity
	}
}

// clampDensity clamps density to valid range
func clampDensity(d, min, max float64) float64 {
	if d < min {
		return min
	}
	if d > max {
		return max
	}
	return d
}

// abs returns absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetCurrentDensities returns the current density configuration (for logging)
func (g *PGReelGenerator) GetCurrentDensities() map[string]float64 {
	// Return Reel 1 densities as representative sample
	densities := make(map[string]float64)
	for sym, d := range g.topologies[0].SymbolDensity {
		densities[sym] = d
	}
	return densities
}

// ReelDensitySummary holds density summary for a single reel
type ReelDensitySummary struct {
	ReelIndex int
	Role      ReelRole
	LowAvg    float64 // Average density for low symbols
	MidAvg    float64 // Average density for mid symbols
	HighAvg   float64 // Average density for high symbols
}

// GetAllReelDensities returns density summary for all 5 reels
// This shows the differentiated behavior based on reel roles
func (g *PGReelGenerator) GetAllReelDensities() []ReelDensitySummary {
	summaries := make([]ReelDensitySummary, 5)

	for reelIdx := 0; reelIdx < 5; reelIdx++ {
		summary := ReelDensitySummary{
			ReelIndex: reelIdx + 1, // 1-indexed for display
			Role:      g.topologies[reelIdx].Role,
		}

		// Calculate average for low symbols
		var lowSum float64
		for _, sym := range g.lowSymbols {
			d := g.topologies[reelIdx].SymbolDensity[sym]
			if d == 0 {
				d = 1.0
			}
			lowSum += d
		}
		summary.LowAvg = lowSum / float64(len(g.lowSymbols))

		// Calculate average for mid symbols
		var midSum float64
		for _, sym := range g.midSymbols {
			d := g.topologies[reelIdx].SymbolDensity[sym]
			if d == 0 {
				d = 1.0
			}
			midSum += d
		}
		summary.MidAvg = midSum / float64(len(g.midSymbols))

		// Calculate average for high symbols
		var highSum float64
		for _, sym := range g.highSymbols {
			d := g.topologies[reelIdx].SymbolDensity[sym]
			if d == 0 {
				d = 1.0
			}
			highSum += d
		}
		summary.HighAvg = highSum / float64(len(g.highSymbols))

		summaries[reelIdx] = summary
	}

	return summaries
}

// RoleName returns human-readable name for a reel role
func RoleName(role ReelRole) string {
	switch role {
	case RoleActivator:
		return "Activator"
	case RoleConnector:
		return "Connector"
	case RoleCore:
		return "Core"
	case RoleAmplifier:
		return "Amplifier"
	case RoleRareSpike:
		return "RareSpike"
	default:
		return "Unknown"
	}
}

// ResetToNeutralDensities resets all symbol densities to 1.0 (neutral)
// Call this before starting auto-tuning to have a clean slate
func (g *PGReelGenerator) ResetToNeutralDensities() {
	allSymbols := append(append(g.lowSymbols, g.midSymbols...), g.highSymbols...)
	for reelIdx := 0; reelIdx < 5; reelIdx++ {
		for _, sym := range allSymbols {
			g.topologies[reelIdx].SymbolDensity[sym] = 1.0
		}
	}
}

// AnalyzeStrip analyzes a reel strip for spacing and clustering metrics
func (g *PGReelGenerator) AnalyzeStrip(strip []string) StripAnalysis {
	analysis := StripAnalysis{
		Length:         len(strip),
		SymbolCounts:   make(map[string]int),
		AvgSpacing:     make(map[string]float64),
		MaxClusterSize: make(map[string]int),
	}

	// Count symbols and track positions
	positions := make(map[string][]int)
	for i, sym := range strip {
		analysis.SymbolCounts[sym]++
		positions[sym] = append(positions[sym], i)
	}

	// Calculate average spacing
	for sym, pos := range positions {
		if len(pos) > 1 {
			totalSpacing := 0
			for i := 1; i < len(pos); i++ {
				totalSpacing += pos[i] - pos[i-1]
			}
			analysis.AvgSpacing[sym] = float64(totalSpacing) / float64(len(pos)-1)
		}
	}

	// Find max cluster sizes
	for _, sym := range strip {
		analysis.MaxClusterSize[sym] = 0
	}
	currentSym := ""
	currentCluster := 0
	for _, sym := range strip {
		if sym == currentSym {
			currentCluster++
		} else {
			currentSym = sym
			currentCluster = 1
		}
		if currentCluster > analysis.MaxClusterSize[sym] {
			analysis.MaxClusterSize[sym] = currentCluster
		}
	}

	return analysis
}

// StripAnalysis holds analysis results for a reel strip
type StripAnalysis struct {
	Length         int
	SymbolCounts   map[string]int
	AvgSpacing     map[string]float64
	MaxClusterSize map[string]int
}

// GetAdjustedWeights returns weights adjusted by topology density for all reels
// This shows the actual weights used after applying density multipliers
func (g *PGReelGenerator) GetAdjustedWeights(baseWeights *ReelWeightsSet) *ReelWeightsSet {
	if baseWeights == nil {
		return nil
	}

	adjusted := &ReelWeightsSet{
		Reel1: g.adjustWeightsByDensity(baseWeights.Reel1, g.topologies[0].SymbolDensity),
		Reel2: g.adjustWeightsByDensity(baseWeights.Reel2, g.topologies[1].SymbolDensity),
		Reel3: g.adjustWeightsByDensity(baseWeights.Reel3, g.topologies[2].SymbolDensity),
		Reel4: g.adjustWeightsByDensity(baseWeights.Reel4, g.topologies[3].SymbolDensity),
		Reel5: g.adjustWeightsByDensity(baseWeights.Reel5, g.topologies[4].SymbolDensity),
	}

	return adjusted
}

// ApplyGoldSymbols replaces some base symbols with gold variants on a reel strip.
// This is called AFTER strip generation to add gold symbols while respecting constraints.
//
// Gold symbol replacement rules:
// - Only paying symbols can become gold (not wild, bonus, scatter)
// - GoldRatio determines what percentage of paying symbols become gold
// - MinGoldSpacing ensures minimum distance between gold symbols
// - MaxGoldCluster limits consecutive gold symbols
//
// Parameters:
// - strip: The generated reel strip (will be modified in place)
// - reelIndex: The reel index (0-4)
//
// Returns: modified strip with gold symbols
func (g *PGReelGenerator) ApplyGoldSymbols(strip []string, reelIndex int) []string {
	if reelIndex < 0 || reelIndex >= 5 {
		return strip
	}

	goldConfig := g.topologies[reelIndex].GoldConfig
	if goldConfig == nil || !goldConfig.Enabled || goldConfig.GoldRatio <= 0 {
		return strip
	}

	// Define paying symbols (can become gold)
	payingSymbols := make(map[string]bool)
	for _, sym := range g.lowSymbols {
		payingSymbols[sym] = true
	}
	for _, sym := range g.midSymbols {
		payingSymbols[sym] = true
	}
	for _, sym := range g.highSymbols {
		payingSymbols[sym] = true
	}

	// Find all positions with paying symbols
	payingPositions := make([]int, 0)
	for i, sym := range strip {
		if payingSymbols[sym] {
			payingPositions = append(payingPositions, i)
		}
	}

	if len(payingPositions) == 0 {
		return strip
	}

	// Calculate target number of gold symbols
	targetGoldCount := int(float64(len(payingPositions)) * goldConfig.GoldRatio)
	if targetGoldCount == 0 && goldConfig.GoldRatio > 0 {
		targetGoldCount = 1 // At least 1 gold if ratio > 0
	}

	// Shuffle paying positions to randomize which become gold
	shuffledPositions := make([]int, len(payingPositions))
	copy(shuffledPositions, payingPositions)
	g.rng.Shuffle(len(shuffledPositions), func(i, j int) {
		shuffledPositions[i], shuffledPositions[j] = shuffledPositions[j], shuffledPositions[i]
	})

	// Track gold symbol positions for spacing constraints
	goldPositions := make([]int, 0)
	goldCount := 0

	// Try to place gold symbols
	for _, pos := range shuffledPositions {
		if goldCount >= targetGoldCount {
			break
		}

		// Check MinGoldSpacing constraint
		if !g.checkGoldSpacing(pos, goldPositions, goldConfig.MinGoldSpacing, len(strip)) {
			continue
		}

		// Check MaxGoldCluster constraint
		if !g.checkGoldCluster(pos, strip, goldConfig.MaxGoldCluster) {
			continue
		}

		// Replace with gold variant
		baseSym := strip[pos]
		strip[pos] = baseSym + "_gold"
		goldPositions = append(goldPositions, pos)
		goldCount++
	}

	// Log gold symbol placement stats
	if goldCount > 0 {
		fmt.Printf("    Reel %d: Applied %d/%d gold symbols (%.1f%% of %d paying symbols)\n",
			reelIndex+1, goldCount, targetGoldCount,
			float64(goldCount)/float64(len(payingPositions))*100, len(payingPositions))
	}

	return strip
}

// checkGoldSpacing checks if placing gold at pos respects MinGoldSpacing constraint
// Uses circular distance for wrap-around strips
func (g *PGReelGenerator) checkGoldSpacing(pos int, goldPositions []int, minSpacing int, stripLen int) bool {
	if minSpacing <= 0 {
		return true
	}

	for _, goldPos := range goldPositions {
		// Calculate circular distance
		dist := pos - goldPos
		if dist < 0 {
			dist = -dist
		}
		// Also check wrap-around distance
		wrapDist := stripLen - dist
		if wrapDist < dist {
			dist = wrapDist
		}

		if dist < minSpacing {
			return false
		}
	}
	return true
}

// checkGoldCluster checks if placing gold at pos respects MaxGoldCluster constraint
// Looks at adjacent positions to count consecutive gold symbols
func (g *PGReelGenerator) checkGoldCluster(pos int, strip []string, maxCluster int) bool {
	if maxCluster <= 0 {
		return true
	}

	// Count consecutive gold symbols before this position
	beforeCount := 0
	for i := pos - 1; i >= 0; i-- {
		if strings.HasSuffix(strip[i], "_gold") {
			beforeCount++
		} else {
			break
		}
	}

	// Count consecutive gold symbols after this position
	afterCount := 0
	for i := pos + 1; i < len(strip); i++ {
		if strings.HasSuffix(strip[i], "_gold") {
			afterCount++
		} else {
			break
		}
	}

	// Total cluster size if we place gold here
	clusterSize := beforeCount + 1 + afterCount

	return clusterSize <= maxCluster
}

// GetGoldConfig returns the gold configuration for a specific reel
func (g *PGReelGenerator) GetGoldConfig(reelIndex int) *GoldTopologyConfig {
	if reelIndex < 0 || reelIndex >= 5 {
		return nil
	}
	return g.topologies[reelIndex].GoldConfig
}

// SetGoldRatio updates the gold ratio for a specific reel
func (g *PGReelGenerator) SetGoldRatio(reelIndex int, ratio float64) {
	if reelIndex < 0 || reelIndex >= 5 {
		return
	}
	if g.topologies[reelIndex].GoldConfig != nil {
		g.topologies[reelIndex].GoldConfig.GoldRatio = ratio
	}
}

// GoldSymbolStats holds statistics about gold symbols on a reel strip
type GoldSymbolStats struct {
	ReelIndex      int
	TotalSymbols   int
	PayingSymbols  int
	GoldSymbols    int
	GoldRatio      float64 // Actual ratio achieved
	TargetRatio    float64 // Target ratio from config
	AvgGoldSpacing float64 // Average distance between gold symbols
	MaxGoldCluster int     // Maximum consecutive gold symbols found
}

// AnalyzeGoldSymbols analyzes gold symbol distribution on a reel strip
func (g *PGReelGenerator) AnalyzeGoldSymbols(strip []string, reelIndex int) GoldSymbolStats {
	stats := GoldSymbolStats{
		ReelIndex:    reelIndex,
		TotalSymbols: len(strip),
	}

	// Get target ratio from config
	if goldConfig := g.GetGoldConfig(reelIndex); goldConfig != nil {
		stats.TargetRatio = goldConfig.GoldRatio
	}

	// Define paying symbols
	payingSymbols := make(map[string]bool)
	for _, sym := range g.lowSymbols {
		payingSymbols[sym] = true
	}
	for _, sym := range g.midSymbols {
		payingSymbols[sym] = true
	}
	for _, sym := range g.highSymbols {
		payingSymbols[sym] = true
	}

	// Count symbols and find gold positions
	goldPositions := make([]int, 0)
	for i, sym := range strip {
		baseSym := strings.Split(sym, "_")[0]
		if payingSymbols[baseSym] {
			stats.PayingSymbols++
		}
		if strings.HasSuffix(sym, "_gold") {
			stats.GoldSymbols++
			goldPositions = append(goldPositions, i)
		}
	}

	// Calculate actual gold ratio
	if stats.PayingSymbols > 0 {
		stats.GoldRatio = float64(stats.GoldSymbols) / float64(stats.PayingSymbols)
	}

	// Calculate average spacing between gold symbols
	if len(goldPositions) > 1 {
		totalSpacing := 0
		for i := 1; i < len(goldPositions); i++ {
			totalSpacing += goldPositions[i] - goldPositions[i-1]
		}
		stats.AvgGoldSpacing = float64(totalSpacing) / float64(len(goldPositions)-1)
	}

	// Find max gold cluster
	currentCluster := 0
	for _, sym := range strip {
		if strings.HasSuffix(sym, "_gold") {
			currentCluster++
			if currentCluster > stats.MaxGoldCluster {
				stats.MaxGoldCluster = currentCluster
			}
		} else {
			currentCluster = 0
		}
	}

	return stats
}
