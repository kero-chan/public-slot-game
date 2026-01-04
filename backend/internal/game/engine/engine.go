package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/internal/game/cascade"
	"github.com/slotmachine/backend/internal/game/freespins"
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/rng"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/slotmachine/backend/internal/pkg/cache"
)

// GameEngine is an enhanced game engine that uses pre-generated reel strips from database
// This provides better performance by avoiding strip generation on every spin
type GameEngine struct {
	cryptoRNG          *rng.CryptoRNG
	reelStripService   reelstrip.Service
	useDBStrips        bool // Flag to enable/disable DB strips (for gradual rollout)
	fallbackToGenerate bool // If true, falls back to generation if DB strips not available
	cache              *cache.Cache
}

// GridPosition represents a position on the grid (reel, row)
type GridPosition struct {
	Reel int `json:"reel"`
	Row  int `json:"row"`
}

// SpinResult represents the complete result of a spin
type SpinResult struct {
	SpinID             uuid.UUID               `json:"spin_id"`
	Grid               reels.Grid              `json:"grid"`
	Cascades           []cascade.CascadeResult `json:"cascades"`
	TotalWin           float64                 `json:"total_win"`
	ScatterCount       int                     `json:"scatter_count"`
	FreeSpinsTriggered bool                    `json:"free_spins_triggered"`
	FreeSpinsAwarded   int                     `json:"free_spins_awarded,omitempty"`
	ReelPositions      []int                   `json:"reel_positions"`       // For provably fair
	ReelStripConfigID  *uuid.UUID              `json:"reel_strip_config_id"` // For provably fair verification
	Timestamp          time.Time               `json:"timestamp"`
}

// FreeSpinResult represents the result of a free spin
type FreeSpinResult struct {
	SpinID          uuid.UUID               `json:"spin_id"`
	Grid            reels.Grid              `json:"grid"`
	Cascades        []cascade.CascadeResult `json:"cascades"`
	TotalWin        float64                 `json:"total_win"`
	ScatterCount    int                     `json:"scatter_count"`
	Retriggered     bool                    `json:"retriggered"`
	AdditionalSpins int                     `json:"additional_spins,omitempty"`
	RemainingSpins  int                     `json:"remaining_spins"`
	SpinNumber      int                     `json:"spin_number"`
	ReelPositions   []int                   `json:"reel_positions"`
	Timestamp       time.Time               `json:"timestamp"`
}

// NewGameEngine creates a new game engine with database-backed reel strips
func NewGameEngine(reelStripService reelstrip.Service, cache *cache.Cache, useDBStrips bool) *GameEngine {
	return &GameEngine{
		cryptoRNG:          rng.NewCryptoRNG(),
		reelStripService:   reelStripService,
		useDBStrips:        useDBStrips,
		cache:              cache,
		fallbackToGenerate: true, // Always allow fallback for safety
	}
}

// GenerateInitialGrid generates a demo grid for initial display
// This ensures frontend has zero RNG - all symbol generation is backend-controlled
// For initial grid (before player context), use default configuration
func (e *GameEngine) GenerateInitialGrid(ctx context.Context) (reels.Grid, error) {
	isFreeSpin := false

	// Get reel strips using default config (no player context yet)
	reelStrips, err := e.GetReelStripsForPlayer(ctx, uuid.Nil, isFreeSpin)
	if err != nil {
		return nil, fmt.Errorf("failed to get reel strips: %w", err)
	}

	// Generate initial grid (10 rows total: 4 buffer + 6 visible)
	initialGrid, _, err := reels.GenerateGrid(reelStrips, e.cryptoRNG)
	if err != nil {
		return nil, fmt.Errorf("failed to generate grid: %w", err)
	}

	// Return all 10 rows for frontend display
	return initialGrid, nil
}

// Game mode constants - only bonus_spin_trigger (guaranteed free spins) is supported
const (
	GameModeBonusSpinTrigger = "bonus_spin_trigger"
)

// generateBonusSpinTriggerGrid generates a grid guaranteed to trigger free spins
// Creates excitement by building anticipation: 2 bonus in first 3 reels, then 1+ in last 2 reels
// This mimics the anticipation experience but guarantees the payoff
func (e *GameEngine) generateBonusSpinTriggerGrid(reelStrips []reels.ReelStrip) (reels.Grid, []int, error) {
	// Visible rows for win checking (rows 5-8 are fully visible)
	const visibleStartRow = reels.WinCheckStartRow         // 5
	const visibleEndRow = reels.WinCheckEndRow             // 8
	visibleRowCount := visibleEndRow - visibleStartRow + 1 // 4 rows

	// Generate base grid using RNG
	grid, positions, err := reels.GenerateGrid(reelStrips, e.cryptoRNG)
	if err != nil {
		return nil, nil, err
	}

	// Remove bonus symbols from ALL reels to control placement
	for reel := 0; reel < reels.ReelCount; reel++ {
		for row := 0; row < len(grid[reel]); row++ {
			if grid[reel][row] == string(symbols.Bonus) {
				nonBonusSymbols := symbols.NonBonusSymbols()
				randomIdx, err := e.cryptoRNG.Intn(len(nonBonusSymbols))
				if err != nil {
					return nil, nil, fmt.Errorf("failed to get random index: %w", err)
				}
				grid[reel][row] = string(nonBonusSymbols[randomIdx])
			}
		}
	}

	// Step 1: Place 2 bonus symbols in first 3 reels (like anticipation mode)
	// Pick 2 different reels from first 3
	firstThreeReels := []int{0, 1, 2}
	if err := e.cryptoRNG.Shuffle(len(firstThreeReels), func(i, j int) {
		firstThreeReels[i], firstThreeReels[j] = firstThreeReels[j], firstThreeReels[i]
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to shuffle reels: %w", err)
	}
	selectedFirstReels := firstThreeReels[:2]

	for _, reelIdx := range selectedFirstReels {
		randomOffset, err := e.cryptoRNG.Intn(visibleRowCount)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get random row: %w", err)
		}
		randomRow := visibleStartRow + randomOffset
		grid[reelIdx][randomRow] = string(symbols.Bonus)
	}

	// Step 2: Place at least 1 bonus in last 2 reels (reels 4 and 5, indices 3 and 4)
	// Randomly decide 1 or 2 bonus in last reels for variety
	numLastReelBonus := 1
	extraBonus, err := e.cryptoRNG.Intn(2) // 0 or 1
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get random extra bonus: %w", err)
	}
	if extraBonus == 1 {
		numLastReelBonus = 2 // Sometimes place 2 for bigger wins (4 total bonus)
	}

	lastTwoReels := []int{3, 4}
	if err := e.cryptoRNG.Shuffle(len(lastTwoReels), func(i, j int) {
		lastTwoReels[i], lastTwoReels[j] = lastTwoReels[j], lastTwoReels[i]
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to shuffle last reels: %w", err)
	}

	for i := 0; i < numLastReelBonus; i++ {
		reelIdx := lastTwoReels[i]
		randomOffset, err := e.cryptoRNG.Intn(visibleRowCount)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get random row: %w", err)
		}
		randomRow := visibleStartRow + randomOffset
		grid[reelIdx][randomRow] = string(symbols.Bonus)
	}

	// Step 3: Remove winning ways - check reel1 & reel2, fix reel3
	// NOTE: This step must NOT replace bonus symbols placed above

	// Collect paying symbols from reel 0 (index 0)
	reel0Symbols := make(map[symbols.Symbol]bool)
	for row := visibleStartRow; row <= visibleEndRow; row++ {
		sym := symbols.GetBaseSymbol(grid[0][row])
		if symbols.IsPayingSymbol(sym) {
			reel0Symbols[sym] = true
		}
		if sym == symbols.SymbolWild {
			// Wild can match all paying symbols
			for _, ps := range symbols.PayingSymbols() {
				reel0Symbols[ps] = true
			}
		}
	}

	// Find symbols in both reel 0 and reel 1 (potential wins)
	potentialWins := make(map[symbols.Symbol]bool)
	for row := visibleStartRow; row <= visibleEndRow; row++ {
		sym := symbols.GetBaseSymbol(grid[1][row])
		if reel0Symbols[sym] {
			potentialWins[sym] = true
		}
		if sym == symbols.SymbolWild {
			for s := range reel0Symbols {
				potentialWins[s] = true
			}
		}
	}

	// Fix reel 2: replace symbols that would form wins
	// IMPORTANT: Never replace Bonus symbols - they must remain for free spins trigger
	for row := visibleStartRow; row <= visibleEndRow; row++ {
		baseSym := symbols.GetBaseSymbol(grid[2][row])

		// Skip bonus symbols - they must not be replaced
		if baseSym == symbols.SymbolBonus {
			continue
		}

		needReplace := potentialWins[baseSym] || (baseSym == symbols.SymbolWild && len(potentialWins) > 0)

		if needReplace {
			// Find safe replacement (not in potentialWins, not wild, not bonus)
			for _, ps := range symbols.PayingSymbols() {
				if !potentialWins[ps] && ps != symbols.SymbolBonus {
					grid[2][row] = string(ps)
					break
				}
			}
		}
	}

	return grid, positions, nil
}

// ReelStripsResult contains reel strips and their config ID for provably fair verification
type ReelStripsResult struct {
	Strips   []reels.ReelStrip
	ConfigID *uuid.UUID // nil if using generated strips (not from DB config)
}

// getReelStripsForPlayer retrieves reel strips for a specific player
// Uses player-specific configuration if assigned, otherwise uses default config
func (e *GameEngine) GetReelStripsForPlayer(ctx context.Context, playerID uuid.UUID, isFreeSpin bool) ([]reels.ReelStrip, error) {
	result, err := e.GetReelStripsWithConfigID(ctx, playerID, isFreeSpin)
	if err != nil {
		return nil, err
	}
	return result.Strips, nil
}

// GetReelStripsWithConfigID retrieves reel strips and their config ID for a specific player
// Returns the config ID for provably fair verification
func (e *GameEngine) GetReelStripsWithConfigID(ctx context.Context, playerID uuid.UUID, isFreeSpin bool) (*ReelStripsResult, error) {
	// If DB strips are disabled, generate on-the-fly
	if !e.useDBStrips || e.reelStripService == nil {
		strips, err := reels.GenerateAllReelStrips(isFreeSpin, e.cryptoRNG)
		return &ReelStripsResult{Strips: strips, ConfigID: nil}, err
	}

	gameMode := string(reelstrip.BaseGame)
	if isFreeSpin {
		gameMode = string(reelstrip.FreeSpins)
	}

	// If playerID is nil (e.g., for initial grid), use default config
	if playerID == uuid.Nil {
		configSet, err := e.reelStripService.GetDefaultReelSet(ctx, gameMode)
		if err == nil && configSet != nil && configSet.IsComplete() {
			strips := e.convertConfigSetToReelStrips(configSet)
			configID := configSet.Config.ID
			return &ReelStripsResult{Strips: strips, ConfigID: &configID}, nil
		}
		// Fall through to fallback
	} else {
		// Get player-specific reel strip configuration
		configSet, err := e.reelStripService.GetReelSetForPlayer(ctx, playerID, gameMode)
		if err == nil && configSet != nil && configSet.IsComplete() {
			strips := e.convertConfigSetToReelStrips(configSet)
			configID := configSet.Config.ID
			return &ReelStripsResult{Strips: strips, ConfigID: &configID}, nil
		}
		// Fall through to fallback
	}

	// Fallback: Generate strips on-the-fly if DB lookup fails
	if e.fallbackToGenerate {
		strips, err := reels.GenerateAllReelStrips(isFreeSpin, e.cryptoRNG)
		return &ReelStripsResult{Strips: strips, ConfigID: nil}, err
	}

	return nil, fmt.Errorf("failed to get strips for player and fallback is disabled")
}

// convertConfigSetToReelStrips converts domain ReelStripConfigSet to game engine ReelStrip format
func (e *GameEngine) convertConfigSetToReelStrips(configSet *reelstrip.ReelStripConfigSet) []reels.ReelStrip {
	return []reels.ReelStrip{
		reels.ReelStrip(configSet.Strips[0].StripData),
		reels.ReelStrip(configSet.Strips[1].StripData),
		reels.ReelStrip(configSet.Strips[2].StripData),
		reels.ReelStrip(configSet.Strips[3].StripData),
		reels.ReelStrip(configSet.Strips[4].StripData),
	}
}

// GetReelStripsForFreeSpinSession retrieves reel strips for a free spin session
// Priority: session.ReelStripConfigID > player assignment > default config > fallback
func (e *GameEngine) GetReelStripsForFreeSpinSession(
	ctx context.Context,
	playerID uuid.UUID,
	session *freespins.Session,
) ([]reels.ReelStrip, error) {
	// First, try to get reel strips from session's ReelStripConfigID if set
	if session.ReelStripConfigID != nil && e.reelStripService != nil {
		configSet, err := e.reelStripService.GetReelSetByConfig(ctx, *session.ReelStripConfigID)
		if err == nil && configSet != nil && configSet.IsComplete() {
			return e.convertConfigSetToReelStrips(configSet), nil
		}
		// Log warning but continue to fallback
		// Failed to get config from ReelStripConfigID, will use default logic
	}

	// Fallback to existing logic: get reel strips for player (free spins mode)
	return e.GetReelStripsForPlayer(ctx, playerID, true)
}

// ValidateBetAmount validates that bet amount is within allowed range
func (e *GameEngine) ValidateBetAmount(betAmount, minBet, maxBet float64) error {
	if betAmount < minBet {
		return fmt.Errorf("bet amount %.2f is below minimum %.2f", betAmount, minBet)
	}
	if betAmount > maxBet {
		return fmt.Errorf("bet amount %.2f exceeds maximum %.2f", betAmount, maxBet)
	}
	return nil
}

// GetRNG returns the crypto RNG instance
func (e *GameEngine) GetRNG() *rng.CryptoRNG {
	return e.cryptoRNG
}

// ExecuteBaseSpinWithRNG executes a base game spin using a provided RNG
// This is used for provably fair gaming where the RNG is derived from a hash chain
func (e *GameEngine) ExecuteBaseSpinWithRNG(
	ctx context.Context,
	playerID uuid.UUID,
	betAmount float64,
	gameMode string,
	customRNG rng.RNG,
) (*SpinResult, error) {
	spinID := uuid.New()
	isFreeSpin := false

	// Get reel strips with config ID for this specific player
	reelStripsResult, err := e.GetReelStripsWithConfigID(ctx, playerID, isFreeSpin)
	if err != nil {
		return nil, fmt.Errorf("failed to get reel strips: %w", err)
	}
	reelStrips := reelStripsResult.Strips

	var initialGrid reels.Grid
	var reelPositions []int

	// Generate grid based on game mode using the provided RNG
	if gameMode == GameModeBonusSpinTrigger {
		// For bonus spin trigger, use the custom RNG
		initialGrid, reelPositions, err = e.generateBonusSpinTriggerGridWithRNG(reelStrips, customRNG)
	} else {
		// Normal spin with custom RNG
		initialGrid, reelPositions, err = reels.GenerateGrid(reelStrips, customRNG)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate grid: %w", err)
	}

	// Execute cascades with custom RNG
	cascadeResults, finalGrid, err := cascade.ExecuteCascades(
		initialGrid,
		reelStrips,
		reelPositions,
		betAmount,
		isFreeSpin,
		customRNG,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cascades: %w", err)
	}

	// Calculate total win
	totalWin := cascade.GetTotalWinFromCascades(cascadeResults, betAmount)

	// Check for free spins trigger
	triggerResult := freespins.CheckTrigger(finalGrid)

	result := &SpinResult{
		SpinID:             spinID,
		Grid:               initialGrid,
		Cascades:           cascadeResults,
		TotalWin:           totalWin,
		ScatterCount:       triggerResult.ScatterCount,
		FreeSpinsTriggered: triggerResult.Triggered,
		FreeSpinsAwarded:   triggerResult.SpinsAwarded,
		ReelPositions:      reelPositions,
		ReelStripConfigID:  reelStripsResult.ConfigID,
		Timestamp:          time.Now().UTC(),
	}

	return result, nil
}

// generateBonusSpinTriggerGridWithRNG generates a bonus spin trigger grid using custom RNG
func (e *GameEngine) generateBonusSpinTriggerGridWithRNG(reelStrips []reels.ReelStrip, customRNG rng.RNG) (reels.Grid, []int, error) {
	const visibleStartRow = reels.WinCheckStartRow
	const visibleEndRow = reels.WinCheckEndRow
	visibleRowCount := visibleEndRow - visibleStartRow + 1

	// Generate base grid using custom RNG
	grid, positions, err := reels.GenerateGrid(reelStrips, customRNG)
	if err != nil {
		return nil, nil, err
	}

	// Remove bonus symbols from ALL reels to control placement
	for reel := 0; reel < reels.ReelCount; reel++ {
		for row := 0; row < len(grid[reel]); row++ {
			if grid[reel][row] == string(symbols.Bonus) {
				nonBonusSymbols := symbols.NonBonusSymbols()
				randomIdx, err := customRNG.Intn(len(nonBonusSymbols))
				if err != nil {
					return nil, nil, fmt.Errorf("failed to get random index: %w", err)
				}
				grid[reel][row] = string(nonBonusSymbols[randomIdx])
			}
		}
	}

	// Place 2 bonus symbols in first 3 reels
	firstThreeReels := []int{0, 1, 2}
	if err := customRNG.Shuffle(len(firstThreeReels), func(i, j int) {
		firstThreeReels[i], firstThreeReels[j] = firstThreeReels[j], firstThreeReels[i]
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to shuffle reels: %w", err)
	}
	selectedFirstReels := firstThreeReels[:2]

	for _, reelIdx := range selectedFirstReels {
		randomOffset, err := customRNG.Intn(visibleRowCount)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get random row: %w", err)
		}
		randomRow := visibleStartRow + randomOffset
		grid[reelIdx][randomRow] = string(symbols.Bonus)
	}

	// Place at least 1 bonus in last 2 reels
	numLastReelBonus := 1
	extraBonus, err := customRNG.Intn(2)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get random extra bonus: %w", err)
	}
	if extraBonus == 1 {
		numLastReelBonus = 2
	}

	lastTwoReels := []int{3, 4}
	if err := customRNG.Shuffle(len(lastTwoReels), func(i, j int) {
		lastTwoReels[i], lastTwoReels[j] = lastTwoReels[j], lastTwoReels[i]
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to shuffle last reels: %w", err)
	}

	for i := 0; i < numLastReelBonus; i++ {
		reelIdx := lastTwoReels[i]
		randomOffset, err := customRNG.Intn(visibleRowCount)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get random row: %w", err)
		}
		randomRow := visibleStartRow + randomOffset
		grid[reelIdx][randomRow] = string(symbols.Bonus)
	}

	// Step 3: Remove winning ways - check reel1 & reel2, fix reel3
	// Collect paying symbols from reel 0 (index 0)
	reel0Symbols := make(map[symbols.Symbol]bool)
	for row := visibleStartRow; row <= visibleEndRow; row++ {
		sym := symbols.GetBaseSymbol(grid[0][row])
		if symbols.IsPayingSymbol(sym) {
			reel0Symbols[sym] = true
		}
		if sym == symbols.SymbolWild {
			// Wild can match all paying symbols
			for _, ps := range symbols.PayingSymbols() {
				reel0Symbols[ps] = true
			}
		}
	}

	// Find symbols in both reel 0 and reel 1 (potential wins)
	potentialWins := make(map[symbols.Symbol]bool)
	for row := visibleStartRow; row <= visibleEndRow; row++ {
		sym := symbols.GetBaseSymbol(grid[1][row])
		if reel0Symbols[sym] {
			potentialWins[sym] = true
		}
		if sym == symbols.SymbolWild {
			for s := range reel0Symbols {
				potentialWins[s] = true
			}
		}
	}

	// Fix reel 2: replace symbols that would form wins
	// IMPORTANT: Never replace Bonus symbols - they must remain for free spins trigger
	for row := visibleStartRow; row <= visibleEndRow; row++ {
		baseSym := symbols.GetBaseSymbol(grid[2][row])

		// Skip bonus symbols - they must not be replaced
		if baseSym == symbols.SymbolBonus {
			continue
		}

		needReplace := potentialWins[baseSym] || (baseSym == symbols.SymbolWild && len(potentialWins) > 0)

		if needReplace {
			// Find safe replacement (not in potentialWins, not wild, not bonus)
			for _, ps := range symbols.PayingSymbols() {
				if !potentialWins[ps] && ps != symbols.SymbolBonus {
					grid[2][row] = string(ps)
					break
				}
			}
		}
	}

	return grid, positions, nil
}

// ExecuteFreeSpinWithRNG executes a free spin using a provided RNG
// This is used for provably fair gaming
func (e *GameEngine) ExecuteFreeSpinWithRNG(
	ctx context.Context,
	playerID uuid.UUID,
	session *freespins.Session,
	spinNumber int,
	customRNG rng.RNG,
) (*FreeSpinResult, error) {
	spinID := uuid.New()
	betAmount := session.LockedBetAmount

	// Get reel strips for this free spin session
	// Priority: session.ReelStripConfigID > player assignment > default config > fallback
	reelStrips, err := e.GetReelStripsForFreeSpinSession(ctx, playerID, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get reel strips: %w", err)
	}

	// Generate initial grid with custom RNG
	initialGrid, reelPositions, err := reels.GenerateGrid(reelStrips, customRNG)
	if err != nil {
		return nil, fmt.Errorf("failed to generate grid: %w", err)
	}

	// Execute cascades with custom RNG
	cascadeResults, finalGrid, err := cascade.ExecuteCascades(
		initialGrid,
		reelStrips,
		reelPositions,
		betAmount,
		true, // isFreeSpin
		customRNG,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cascades: %w", err)
	}

	// Calculate total win
	totalWin := cascade.GetTotalWinFromCascades(cascadeResults, betAmount)

	// Check for retrigger
	retriggerResult := freespins.CheckRetrigger(finalGrid, session.RemainingSpins-1)

	result := &FreeSpinResult{
		SpinID:          spinID,
		Grid:            initialGrid,
		Cascades:        cascadeResults,
		TotalWin:        totalWin,
		ScatterCount:    retriggerResult.ScatterCount,
		Retriggered:     retriggerResult.Retriggered,
		AdditionalSpins: retriggerResult.AdditionalSpins,
		RemainingSpins:  retriggerResult.NewTotalRemaining,
		SpinNumber:      spinNumber,
		ReelPositions:   reelPositions,
		Timestamp:       time.Now().UTC(),
	}

	return result, nil
}

// ExecuteTrialSpin executes a spin for trial mode using HUGE RTP weights
// Trial spins use generated strips with higher winning rates, not DB-backed strips
func (e *GameEngine) ExecuteTrialSpin(ctx context.Context, betAmount float64, gameMode string) (*SpinResult, error) {
	spinID := uuid.New()
	isFreeSpin := false

	// Generate trial reel strips with HUGE RTP
	reelStrips, err := reels.GenerateTrialReelStrips(isFreeSpin, e.cryptoRNG)
	if err != nil {
		return nil, fmt.Errorf("failed to generate trial reel strips: %w", err)
	}

	var initialGrid reels.Grid
	var reelPositions []int

	// Generate grid based on game mode
	if gameMode == GameModeBonusSpinTrigger {
		// Guaranteed free spins for trial users too
		initialGrid, reelPositions, err = e.generateBonusSpinTriggerGrid(reelStrips)
	} else {
		initialGrid, reelPositions, err = reels.GenerateGrid(reelStrips, e.cryptoRNG)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate grid: %w", err)
	}

	// Execute cascades
	cascadeResults, finalGrid, err := cascade.ExecuteCascades(
		initialGrid,
		reelStrips,
		reelPositions,
		betAmount,
		isFreeSpin,
		e.cryptoRNG,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cascades: %w", err)
	}

	// Calculate total win
	totalWin := cascade.GetTotalWinFromCascades(cascadeResults, betAmount)

	// Check for free spins trigger
	triggerResult := freespins.CheckTrigger(finalGrid)

	result := &SpinResult{
		SpinID:             spinID,
		Grid:               initialGrid,
		Cascades:           cascadeResults,
		TotalWin:           totalWin,
		ScatterCount:       triggerResult.ScatterCount,
		FreeSpinsTriggered: triggerResult.Triggered,
		FreeSpinsAwarded:   triggerResult.SpinsAwarded,
		ReelPositions:      reelPositions,
		ReelStripConfigID:  nil, // Trial uses generated strips, not DB config
		Timestamp:          time.Now().UTC(),
	}

	return result, nil
}

// ExecuteTrialFreeSpin executes a free spin for trial mode using HUGE RTP weights
func (e *GameEngine) ExecuteTrialFreeSpin(
	betAmount float64,
	remainingSpins int,
	spinNumber int,
) (*FreeSpinResult, error) {
	spinID := uuid.New()
	isFreeSpin := true

	// Generate trial free spin strips with HUGE RTP
	reelStrips, err := reels.GenerateTrialReelStrips(isFreeSpin, e.cryptoRNG)
	if err != nil {
		return nil, fmt.Errorf("failed to generate trial reel strips: %w", err)
	}

	// Generate initial grid
	initialGrid, reelPositions, err := reels.GenerateGrid(reelStrips, e.cryptoRNG)
	if err != nil {
		return nil, fmt.Errorf("failed to generate grid: %w", err)
	}

	// Execute cascades with free spin multipliers
	cascadeResults, finalGrid, err := cascade.ExecuteCascades(
		initialGrid,
		reelStrips,
		reelPositions,
		betAmount,
		isFreeSpin,
		e.cryptoRNG,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cascades: %w", err)
	}

	// Calculate total win
	totalWin := cascade.GetTotalWinFromCascades(cascadeResults, betAmount)

	// Check for retrigger
	retriggerResult := freespins.CheckRetrigger(finalGrid, remainingSpins-1)

	result := &FreeSpinResult{
		SpinID:          spinID,
		Grid:            initialGrid,
		Cascades:        cascadeResults,
		TotalWin:        totalWin,
		ScatterCount:    retriggerResult.ScatterCount,
		Retriggered:     retriggerResult.Retriggered,
		AdditionalSpins: retriggerResult.AdditionalSpins,
		RemainingSpins:  retriggerResult.NewTotalRemaining,
		SpinNumber:      spinNumber,
		ReelPositions:   reelPositions,
		Timestamp:       time.Now().UTC(),
	}

	return result, nil
}

// SetUseDBStrips enables or disables using DB strips (for feature toggle)
func (e *GameEngine) SetUseDBStrips(enabled bool) {
	e.useDBStrips = enabled
}

// IsUsingDBStrips returns whether DB strips are currently enabled
func (e *GameEngine) IsUsingDBStrips() bool {
	return e.useDBStrips
}

// CountSymbol counts a specific symbol in a grid
func CountSymbolWithDB(grid reels.Grid, symbol symbols.Symbol) int {
	return grid.CountSymbol(string(symbol))
}
