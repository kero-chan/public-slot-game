package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	dfreespins "github.com/slotmachine/backend/domain/freespins"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/domain/spin"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/db"
	"github.com/slotmachine/backend/internal/game/cascade"
	"github.com/slotmachine/backend/internal/game/engine"
	"github.com/slotmachine/backend/internal/game/freespins"
	freespinsEngine "github.com/slotmachine/backend/internal/game/freespins"
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/rng"
	infraCache "github.com/slotmachine/backend/internal/infra/cache"
	"github.com/slotmachine/backend/internal/infra/repository"
	"github.com/slotmachine/backend/internal/pkg/cache"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/service"
)

// SimulationStats holds the statistics from simulation
type SimulationStats struct {
	TotalSpins             int     `json:"total_spins"`
	TotalWagered           float64 `json:"total_wagered"`
	TotalWon               float64 `json:"total_won"`
	RTP                    float64 `json:"rtp"`
	BaseGameWins           int     `json:"base_game_wins"`
	BaseGameTotalWon       float64 `json:"base_game_total_won"`
	BaseRTP                float64 `json:"base_rtp"`
	FreeSpinsTriggered     int     `json:"free_spins_triggered"`
	FreeSpinsTriggeredRate float64 `json:"free_spins_triggered_rate"`
	FreeSpinsTotalWon      float64 `json:"free_spins_total_won"`
	FreeRTP                float64 `json:"free_rtp"`
	FreeSpinsRetriggered   int     `json:"free_spins_retriggered"`
	MaxWin                 float64 `json:"max_win"`
	MaxWinSpin             int     `json:"max_win_spin"`

	// Hit frequency
	NoWinSpins int `json:"no_win_spins"`
	SmallWins  int `json:"small_wins"`  // < 5x bet
	MediumWins int `json:"medium_wins"` // 5x - 20x bet
	BigWins    int `json:"big_wins"`    // 20x - 100x bet
	MegaWins   int `json:"mega_wins"`   // > 100x bet

	// Cascade statistics
	TotalCascades     int     `json:"total_cascades"`
	MaxCascades       int     `json:"max_cascades"`
	AvgCascadesPerWin float64 `json:"avg_cascades_per_win"`

	// Free spins statistics
	TotalFreeSpins      int     `json:"total_free_spins"`
	AvgFreeSpinsAwarded float64 `json:"avg_free_spins_awarded"`
}

func main() {

	// Parse command line flags
	numSpins := flag.Int("spins", 1000000, "Number of spins to simulate")
	betAmount := flag.Float64("bet", 1.0, "Bet amount per spin")
	progressInterval := flag.Int("progress", 100000, "Progress report interval")
	isRealMode := flag.Bool("real", false, "Run real mode")
	targetRTP := flag.Float64("target-rtp", 96.7, "Target RTP")
	playerId := flag.String("player-id", "b76f37bc-8014-41eb-a710-d105a8ae6293", "Player ID")
	flag.Parse()

	// playerID := uuid.Nil
	playerID, err := uuid.Parse(*playerId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse player ID: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         SLOT MACHINE RTP SIMULATOR v2.0                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Spins:        %d\n", *numSpins)
	fmt.Printf("  Bet Amount:   %.2f\n", *betAmount)
	fmt.Printf("  Target RTP:   %.2f%%\n", *targetRTP)
	fmt.Printf("  Is Real   : %v\n", *isRealMode)
	fmt.Println()

	// Initialize game engine
	var gameEngine *engine.GameEngine
	var reelStripService reelstrip.Service

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

	// Initialize Redis client for PF session cache
	redisClient, err := infraCache.NewRedisClient(cfg, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Redis client: %v\n", err)
		os.Exit(1)
	}

	// Initialize repositories
	reelStripRepo := repository.NewReelStripGormRepository(database, cacheClient)
	sessionRepo := repository.NewSessionGormRepository(database)
	spinRepo := repository.NewSpinGormRepository(database)
	freespinsRepo := repository.NewFreeSpinsGormRepository(database)
	playerRepo := repository.NewPlayerGormRepository(database)
	pfRepo := repository.NewProvablyFairGormRepository(database)
	txManager := repository.NewTxManager(database)

	// Initialize services
	reelStripService = service.NewReelStripService(reelStripRepo, log)
	sessionService := service.NewSessionService(sessionRepo, playerRepo, log)
	gameEngine = engine.NewGameEngine(reelStripService, cacheClient, true)

	// Initialize provably fair service
	pfCache := infraCache.NewPFSessionCache(redisClient, log)
	pfService, err := service.NewProvablyFairService(pfRepo, pfCache, reelStripRepo, cfg, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create provably fair service: %v\n", err)
		os.Exit(1)
	}

	// Initialize spin and free spins services with pfService
	spinService := service.NewSpinService(spinRepo, playerRepo, sessionRepo, gameEngine, freespinsRepo, reelStripRepo, txManager, pfService.(*service.ProvablyFairService), log)
	freeSpinsService := service.NewFreeSpinsService(sessionRepo, freespinsRepo, spinRepo, playerRepo, gameEngine, pfService.(*service.ProvablyFairService), log)

	fmt.Println()
	fmt.Println("Starting simulation...")
	fmt.Println()

	if *isRealMode {
		// Real mode: uses actual services with provably fair sessions (like real client requests)
		stats := runRTPCheck(sessionService, spinService, freeSpinsService, pfService.(*service.ProvablyFairService), playerID, *betAmount, *numSpins, *progressInterval)
		printResults(stats, *betAmount, *targetRTP)
	} else {
		// Run simulation
		stats := runSimulation(gameEngine, playerID, *numSpins, *betAmount, *progressInterval)

		// Print results
		printResults(stats, *betAmount, *targetRTP)
	}
}

func runSimulation(gameEngine *engine.GameEngine, playerID uuid.UUID, numSpins int, betAmount float64, progressInterval int) SimulationStats {
	stats := SimulationStats{}
	cryptoRNG := rng.NewCryptoRNG()
	startTime := time.Now()
	for i := 0; i < numSpins; i++ {
		// Progress reporting
		if (i+1)%progressInterval == 0 {
			elapsed := time.Since(startTime)
			spinsPerSec := float64(i+1) / elapsed.Seconds()
			remaining := time.Duration(float64(numSpins-i-1)/spinsPerSec) * time.Second

			fmt.Printf("Progress: %d/%d spins (%.1f%%) | %.0f spins/sec | ETA: %s\n",
				i+1, numSpins, float64(i+1)/float64(numSpins)*100, spinsPerSec, remaining.Round(time.Second))
		}

		reelStrips, err := gameEngine.GetReelStripsForPlayer(context.Background(), playerID, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get reel strips: %v\n", err)
			os.Exit(1)
		}

		// Execute base spin (now uses DB-backed reel strips if enabled)
		// Use uuid.Nil for RTP simulation (will use default configuration)
		result, err := executeBaseSpin(reelStrips, cryptoRNG, betAmount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing spin %d: %v\n", i+1, err)
			continue
		}

		// Update statistics
		stats.TotalSpins++
		stats.TotalWagered += betAmount
		stats.TotalWon += result.TotalWin

		// Execute free spins if triggered
		if result.FreeSpinsTriggered {
			freeSpinStrips, err := gameEngine.GetReelStripsForPlayer(context.Background(), playerID, true)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get reel strips: %v\n", err)
				os.Exit(1)
			}

			freeSpinsTotalWin := executeFreeSpins(freeSpinStrips, cryptoRNG, result.ScatterCount, betAmount)
			stats.FreeSpinsTotalWon += freeSpinsTotalWin
			stats.TotalWon += freeSpinsTotalWin
		}

		// Track max win
		if result.TotalWin > stats.MaxWin {
			stats.MaxWin = result.TotalWin
			stats.MaxWinSpin = i + 1
		}

		// Categorize win size
		winMultiplier := result.TotalWin / betAmount
		if result.TotalWin == 0 {
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
		numCascades := len(result.Cascades)
		stats.TotalCascades += numCascades
		if numCascades > stats.MaxCascades {
			stats.MaxCascades = numCascades
		}
		if result.TotalWin > 0 {
			stats.BaseGameWins++
			stats.BaseGameTotalWon += result.TotalWin
		}

		// Track free spins trigger
		if result.FreeSpinsTriggered {
			stats.FreeSpinsTriggered++
			stats.AvgFreeSpinsAwarded += float64(result.FreeSpinsAwarded)
		}
	}

	// Calculate final statistics
	if stats.TotalWagered > 0 {
		stats.RTP = (stats.TotalWon / stats.TotalWagered) * 100
		stats.BaseRTP = (stats.BaseGameTotalWon / stats.TotalWagered) * 100
		stats.FreeRTP = (stats.FreeSpinsTotalWon / stats.TotalWagered) * 100
	}

	if stats.BaseGameWins > 0 {
		stats.AvgCascadesPerWin = float64(stats.TotalCascades) / float64(stats.BaseGameWins)
	}

	if stats.FreeSpinsTriggered > 0 {
		stats.AvgFreeSpinsAwarded /= float64(stats.FreeSpinsTriggered)
		stats.FreeSpinsTriggeredRate = float64(stats.FreeSpinsTriggered) / float64(stats.TotalSpins) * 100
	}

	return stats
}

func executeBaseSpin(reelStrips []reels.ReelStrip, cryptoRNG *rng.CryptoRNG, betAmount float64) (*engine.SpinResult, error) {
	spinID := uuid.New()
	isFreeSpin := false

	// Generate initial grid
	initialGrid, reelPositions, err := reels.GenerateGrid(reelStrips, cryptoRNG)
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
		cryptoRNG,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cascades: %w", err)
	}

	// Calculate total win
	totalWin := cascade.GetTotalWinFromCascades(cascadeResults, betAmount)

	// Check for free spins trigger
	triggerResult := freespins.CheckTrigger(finalGrid)

	result := &engine.SpinResult{
		SpinID:             spinID,
		Grid:               initialGrid,
		Cascades:           cascadeResults,
		TotalWin:           totalWin,
		ScatterCount:       triggerResult.ScatterCount,
		FreeSpinsTriggered: triggerResult.Triggered,
		FreeSpinsAwarded:   triggerResult.SpinsAwarded,
		ReelPositions:      reelPositions,
		Timestamp:          time.Now().UTC(),
	}

	return result, nil
}

// executeFreeSpins executes all free spins in a session and returns total win
func executeFreeSpins(reelStrips []reels.ReelStrip, cryptoRNG *rng.CryptoRNG, scatterCount int, betAmount float64) float64 {
	isFreeSpin := true
	// Create a free spins session
	session := freespinsEngine.NewSession(uuid.Nil, scatterCount, betAmount, nil)
	totalWin := 0.0
	spinNumber := 1

	// Execute all free spins in the session
	for !session.IsComplete() {
		// Generate initial grid
		initialGrid, reelPositions, err := reels.GenerateGrid(reelStrips, cryptoRNG)
		if err != nil {
			fmt.Printf("failed to generate grid: %s", err.Error())
			break
		}

		// Execute cascades with free spin multipliers
		cascadeResults, finalGrid, err := cascade.ExecuteCascades(
			initialGrid,
			reelStrips,
			reelPositions,
			betAmount,
			isFreeSpin,
			cryptoRNG,
		)
		if err != nil {
			fmt.Printf("failed to execute cascades: %s", err.Error())
			break
		}

		// Calculate total win
		totalCascadeWin := cascade.GetTotalWinFromCascades(cascadeResults, betAmount)

		// Check for retrigger
		retriggerResult := freespins.CheckRetrigger(finalGrid, session.RemainingSpins-1)

		result := &engine.FreeSpinResult{
			TotalWin:        totalCascadeWin,
			Retriggered:     retriggerResult.Retriggered,
			AdditionalSpins: retriggerResult.AdditionalSpins,
		}

		totalWin += result.TotalWin
		session.ExecuteSpin(result.TotalWin)

		// Handle retrigger
		if result.Retriggered {
			session.AddRetriggerSpins(result.AdditionalSpins)
		}

		spinNumber++
	}

	return totalWin
}

func printResults(stats SimulationStats, betAmount float64, targetRTP float64) {
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
	fmt.Println()

	// Base game statistics
	fmt.Println("═══ BASE GAME ═══")
	fmt.Printf("Base Game Wins:        %d (%.2f%%)\n", stats.BaseGameWins,
		float64(stats.BaseGameWins)/float64(stats.TotalSpins)*100)
	fmt.Printf("Base Game RTP:         %.4f%%\n", stats.BaseRTP)
	fmt.Printf("Avg Cascades/Win:      %.2f\n", stats.AvgCascadesPerWin)
	fmt.Printf("Max Cascades:          %d\n", stats.MaxCascades)
	fmt.Println()

	// Free spins statistics
	fmt.Println("═══ FREE SPINS ═══")
	fmt.Printf("Triggered:             %d times (%.4f%%)\n", stats.FreeSpinsTriggered, stats.FreeSpinsTriggeredRate)
	fmt.Printf("Avg Spins Awarded:     %.2f\n", stats.AvgFreeSpinsAwarded)
	fmt.Printf("Free Spins RTP:        %.4f%%\n", stats.FreeRTP)
	fmt.Printf("Avg Trigger Frequency: 1 in %.0f spins\n", float64(stats.TotalSpins)/float64(stats.FreeSpinsTriggered))
	fmt.Println()

	// Max win
	fmt.Println("═══ MAX WIN ═══")
	maxWinMultiplier := stats.MaxWin / betAmount
	fmt.Printf("Max Win:               %.2f (%.1fx bet)\n", stats.MaxWin, maxWinMultiplier)
	fmt.Printf("Occurred at Spin:      %d\n", stats.MaxWinSpin)
	fmt.Println()

	// Volatility indicator
	fmt.Println("═══ VOLATILITY INDICATORS ═══")
	avgWin := stats.TotalWon / float64(totalWinSpins)
	fmt.Printf("Average Win:           %.2f (%.2fx bet)\n", avgWin, avgWin/betAmount)
	fmt.Printf("Max/Avg Win Ratio:     %.1fx\n", stats.MaxWin/avgWin)

	volatility := "MEDIUM"
	if maxWinMultiplier > 500 {
		volatility = "HIGH"
	} else if maxWinMultiplier < 100 {
		volatility = "LOW"
	}
	fmt.Printf("Volatility:            %s\n", volatility)
	fmt.Println()
}

// runRTPCheck runs RTP simulation using real services with provably fair sessions
// This simulates actual client requests with full PF session management
func runRTPCheck(
	sessionService session.Service,
	spinService spin.Service,
	freeSpinsService dfreespins.Service,
	pfService *service.ProvablyFairService,
	playerID uuid.UUID,
	betAmount float64,
	numSpins int,
	progressInterval int,
) SimulationStats {
	ctx := context.Background()
	stats := SimulationStats{}

	// Start a game session
	gameSession, err := sessionService.StartSession(ctx, playerID, betAmount)
	if err != nil {
		fmt.Printf("❌ Failed to start session: %s\n", err.Error())
		return stats
	}
	defer sessionService.EndSession(context.Background(), gameSession.ID)

	// Start a provably fair session for this game session
	_, err = pfService.StartSession(ctx, playerID, gameSession.ID, "")
	if err != nil {
		fmt.Printf("❌ Failed to start PF session: %s\n", err.Error())
		return stats
	}
	defer pfService.EndSession(context.Background(), gameSession.ID)

	// Generate a client seed for RTP simulation
	clientSeed := fmt.Sprintf("rtp-sim-%d", time.Now().UnixNano())

	startTime := time.Now()
	for i := 0; i < numSpins; i++ {
		// Progress reporting
		if (i+1)%progressInterval == 0 {
			elapsed := time.Since(startTime)
			spinsPerSec := float64(i+1) / elapsed.Seconds()
			remaining := time.Duration(float64(numSpins-i-1)/spinsPerSec) * time.Second

			fmt.Printf("Progress: %d/%d spins (%.1f%%) | %.0f spins/sec | ETA: %s\n",
				i+1, numSpins, float64(i+1)/float64(numSpins)*100, spinsPerSec, remaining.Round(time.Second))
		}

		// Execute spin with client seed (no theta seed for simulation)
		result, err := spinService.ExecuteSpin(ctx, playerID, gameSession.ID, betAmount, "", clientSeed, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing spin %d: %v\n", i+1, err)
			continue
		}

		// Update statistics
		stats.TotalSpins++
		stats.TotalWagered += betAmount
		stats.TotalWon += result.SpinTotalWin

		// Execute free spins if triggered
		if result.FreeSpinsTriggered {
			stats.FreeSpinsTriggered++
			remainingSpins := result.FreeSpinsRemainingSpins

			// Execute all free spins in the session
			for remainingSpins > 0 {
				freeSpinSessionID, err := uuid.Parse(result.FreeSpinsSessionID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error parsing free spins session ID: %v\n", err)
					break
				}

				// Execute free spin with client seed
				freeResult, err := freeSpinsService.ExecuteFreeSpin(ctx, freeSpinSessionID, clientSeed)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error executing free spin: %v\n", err)
					break
				}

				stats.FreeSpinsTotalWon += freeResult.SpinTotalWin
				stats.TotalWon += freeResult.SpinTotalWin
				stats.TotalFreeSpins++
				remainingSpins = freeResult.FreeSpinsRemainingSpins

				// Check for retrigger
				if freeResult.FreeSpinsRetriggered {
					stats.FreeSpinsRetriggered++
				}
			}
		}

		// Track max win
		if result.SpinTotalWin > stats.MaxWin {
			stats.MaxWin = result.SpinTotalWin
			stats.MaxWinSpin = i + 1
		}

		// Categorize win size
		winMultiplier := result.SpinTotalWin / betAmount
		if result.SpinTotalWin == 0 {
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
		numCascades := len(result.Cascades)
		stats.TotalCascades += numCascades
		if numCascades > stats.MaxCascades {
			stats.MaxCascades = numCascades
		}
		if result.SpinTotalWin > 0 {
			stats.BaseGameWins++
			stats.BaseGameTotalWon += result.SpinTotalWin
		}
	}

	// Calculate final statistics
	if stats.TotalWagered > 0 {
		stats.RTP = (stats.TotalWon / stats.TotalWagered) * 100
		stats.BaseRTP = (stats.BaseGameTotalWon / stats.TotalWagered) * 100
		stats.FreeRTP = (stats.FreeSpinsTotalWon / stats.TotalWagered) * 100
	}

	if stats.BaseGameWins > 0 {
		stats.AvgCascadesPerWin = float64(stats.TotalCascades) / float64(stats.BaseGameWins)
	}

	if stats.FreeSpinsTriggered > 0 {
		stats.AvgFreeSpinsAwarded = float64(stats.TotalFreeSpins) / float64(stats.FreeSpinsTriggered)
		stats.FreeSpinsTriggeredRate = float64(stats.FreeSpinsTriggered) / float64(stats.TotalSpins) * 100
	}

	return stats
}
