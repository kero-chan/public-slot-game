package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/reelstrip"
)

// CSVReelStrip represents a row from reel_strips_masterdata.csv
type CSVReelStrip struct {
	ID          string
	GameMode    string
	ReelNumber  int
	StripData   []string
	Checksum    string
	StripLength int
	CreatedAt   time.Time
	IsActive    bool
	Notes       string
}

// CSVReelStripConfig represents a row from reel_strip_configs_masterdata.csv
type CSVReelStripConfig struct {
	ID            string
	Name          string
	GameMode      string
	Description   string
	Reel0StripID  string
	Reel1StripID  string
	Reel2StripID  string
	Reel3StripID  string
	Reel4StripID  string
	TargetRTP     float64
	IsActive      bool
	IsDefault     bool
	ActivatedAt   *time.Time
	DeactivatedAt *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     string
	Notes         string
}

func main() {
	// Command-line flags
	gameModeFlag := flag.String("mode", "both", "Game mode: base_game, free_spins, or both")
	flag.Parse()

	// Initialize application with Wire
	application, err := InitializeSeedApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	log := application.Logger
	reelStripService := application.ReelStripService
	reelStripRepo := application.ReelStripRepository

	ctx := context.Background()

	log.Info().
		Str("mode", *gameModeFlag).
		Msg("Starting reel strip CSV import")

	startTime := time.Now()

	// Get script directory to locate CSV files
	scriptDir, err := getScriptDir()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get script directory")
	}

	reelStripsCSV := filepath.Join(scriptDir, "reel_strips_masterdata.csv")
	configsCSV := filepath.Join(scriptDir, "reel_strip_configs_masterdata.csv")

	// Import reel strips
	log.Info().Str("file", reelStripsCSV).Msg("Importing reel strips from CSV")
	reelStrips, err := importReelStripsFromCSV(reelStripsCSV)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to import reel strips from CSV")
	}
	log.Info().Int("count", len(reelStrips)).Msg("Parsed reel strips from CSV")

	// Import configs
	log.Info().Str("file", configsCSV).Msg("Importing configs from CSV")
	configs, err := importConfigsFromCSV(configsCSV)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to import configs from CSV")
	}
	log.Info().Int("count", len(configs)).Msg("Parsed configs from CSV")

	// Determine which game modes to import
	var gameModes []string
	switch *gameModeFlag {
	case "base_game":
		gameModes = []string{string(reelstrip.BaseGame)}
	case "free_spins":
		gameModes = []string{string(reelstrip.FreeSpins)}
	case "both":
		gameModes = []string{string(reelstrip.BaseGame), string(reelstrip.FreeSpins)}
	default:
		log.Fatal().Str("mode", *gameModeFlag).Msg("Invalid game mode. Use: base_game, free_spins, or both")
	}

	// Save reel strips to database
	for _, csvStrip := range reelStrips {
		// Skip if not in selected game modes
		if !contains(gameModes, csvStrip.GameMode) {
			continue
		}

		// Convert to domain model
		stripID, err := uuid.Parse(csvStrip.ID)
		if err != nil {
			log.Error().Err(err).Str("id", csvStrip.ID).Msg("Failed to parse strip ID")
			continue
		}

		strip := &reelstrip.ReelStrip{
			ID:          stripID,
			GameMode:    csvStrip.GameMode,
			ReelNumber:  csvStrip.ReelNumber,
			StripData:   csvStrip.StripData,
			Checksum:    csvStrip.Checksum,
			StripLength: csvStrip.StripLength,
			CreatedAt:   csvStrip.CreatedAt,
			IsActive:    csvStrip.IsActive,
			Notes:       csvStrip.Notes,
		}

		// Save to database
		if err := reelStripRepo.Create(ctx, strip); err != nil {
			log.Error().
				Err(err).
				Str("strip_id", strip.ID.String()).
				Str("game_mode", strip.GameMode).
				Int("reel_number", strip.ReelNumber).
				Msg("Failed to save reel strip")
		} else {
			log.Info().
				Str("strip_id", strip.ID.String()).
				Str("game_mode", strip.GameMode).
				Int("reel_number", strip.ReelNumber).
				Msg("Saved reel strip")
		}
	}

	// Save configs to database
	for _, csvConfig := range configs {
		// Skip if not in selected game modes
		if !contains(gameModes, csvConfig.GameMode) {
			continue
		}

		// Convert to domain model
		configID, err := uuid.Parse(csvConfig.ID)
		if err != nil {
			log.Error().Err(err).Str("id", csvConfig.ID).Msg("Failed to parse config ID")
			continue
		}

		reel0ID, _ := uuid.Parse(csvConfig.Reel0StripID)
		reel1ID, _ := uuid.Parse(csvConfig.Reel1StripID)
		reel2ID, _ := uuid.Parse(csvConfig.Reel2StripID)
		reel3ID, _ := uuid.Parse(csvConfig.Reel3StripID)
		reel4ID, _ := uuid.Parse(csvConfig.Reel4StripID)

		config := &reelstrip.ReelStripConfig{
			ID:            configID,
			Name:          csvConfig.Name,
			GameMode:      csvConfig.GameMode,
			Description:   csvConfig.Description,
			Reel0StripID:  reel0ID,
			Reel1StripID:  reel1ID,
			Reel2StripID:  reel2ID,
			Reel3StripID:  reel3ID,
			Reel4StripID:  reel4ID,
			TargetRTP:     csvConfig.TargetRTP,
			IsActive:      csvConfig.IsActive,
			IsDefault:     csvConfig.IsDefault,
			ActivatedAt:   csvConfig.ActivatedAt,
			DeactivatedAt: csvConfig.DeactivatedAt,
			CreatedAt:     csvConfig.CreatedAt,
			UpdatedAt:     csvConfig.UpdatedAt,
			CreatedBy:     csvConfig.CreatedBy,
			Notes:         csvConfig.Notes,
		}

		// Save to database
		if err := reelStripRepo.CreateConfig(ctx, config); err != nil {
			log.Error().
				Err(err).
				Str("config_id", config.ID.String()).
				Str("config_name", config.Name).
				Msg("Failed to save config")
		} else {
			log.Info().
				Str("config_id", config.ID.String()).
				Str("config_name", config.Name).
				Bool("is_default", config.IsDefault).
				Msg("Saved config")

			// Set as default if needed
			if config.IsDefault {
				if err := reelStripService.SetDefaultConfig(ctx, config.ID, config.GameMode); err != nil {
					log.Error().
						Err(err).
						Str("config_id", config.ID.String()).
						Msg("Failed to set default config")
				}
			}
		}
	}

	duration := time.Since(startTime)

	log.Info().
		Dur("duration", duration).
		Msg("Reel strip CSV import completed successfully")

	fmt.Println("\n=== Reel Strip CSV Import Summary ===")
	fmt.Printf("Game Modes: %v\n", gameModes)
	fmt.Printf("Total strips imported: %d\n", len(reelStrips))
	fmt.Printf("Total configs imported: %d\n", len(configs))
	fmt.Printf("Duration: %v\n", duration)
	fmt.Println("=====================================")
}

// importReelStripsFromCSV reads and parses the reel strips CSV file
func importReelStripsFromCSV(filename string) ([]*CSVReelStrip, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	// Skip header row
	var strips []*CSVReelStrip
	for i, record := range records[1:] {
		if len(record) < 9 {
			return nil, fmt.Errorf("invalid CSV row %d: expected 9 columns, got %d", i+2, len(record))
		}

		// Parse strip data JSON array
		var stripData []string
		if err := json.Unmarshal([]byte(record[3]), &stripData); err != nil {
			return nil, fmt.Errorf("failed to parse strip_data at row %d: %w", i+2, err)
		}

		// Parse reel number
		reelNumber, err := strconv.Atoi(record[2])
		if err != nil {
			return nil, fmt.Errorf("failed to parse reel_number at row %d: %w", i+2, err)
		}

		// Parse strip length
		stripLength, err := strconv.Atoi(record[5])
		if err != nil {
			return nil, fmt.Errorf("failed to parse strip_length at row %d: %w", i+2, err)
		}

		// Parse created_at
		createdAt, err := time.Parse("2006-01-02 15:04:05.999 -0700", record[6])
		if err != nil {
			// Try alternative format
			createdAt, err = time.Parse("2006-01-02 15:04:05 -0700", record[6])
			if err != nil {
				return nil, fmt.Errorf("failed to parse created_at at row %d: %w", i+2, err)
			}
		}

		// Parse is_active
		isActive := record[7] == "true" || record[7] == "t" || record[7] == "1"

		strip := &CSVReelStrip{
			ID:          record[0],
			GameMode:    record[1],
			ReelNumber:  reelNumber,
			StripData:   stripData,
			Checksum:    record[4],
			StripLength: stripLength,
			CreatedAt:   createdAt,
			IsActive:    isActive,
			Notes:       record[8],
		}

		strips = append(strips, strip)
	}

	return strips, nil
}

// importConfigsFromCSV reads and parses the configs CSV file
func importConfigsFromCSV(filename string) ([]*CSVReelStripConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	// Skip header row
	var configs []*CSVReelStripConfig
	for i, record := range records[1:] {
		if len(record) < 18 {
			return nil, fmt.Errorf("invalid CSV row %d: expected 18 columns, got %d", i+2, len(record))
		}

		// Parse target_rtp
		targetRTP, err := strconv.ParseFloat(record[9], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target_rtp at row %d: %w", i+2, err)
		}

		// Parse is_active
		isActive := record[10] == "true" || record[10] == "t" || record[10] == "1"

		// Parse is_default
		isDefault := record[11] == "true" || record[11] == "t" || record[11] == "1"

		// Parse timestamps
		var activatedAt, deactivatedAt *time.Time
		if record[12] != "" {
			t, err := time.Parse("2006-01-02 15:04:05.999 -0700", record[12])
			if err == nil {
				activatedAt = &t
			}
		}
		if record[13] != "" {
			t, err := time.Parse("2006-01-02 15:04:05.999 -0700", record[13])
			if err == nil {
				deactivatedAt = &t
			}
		}

		createdAt, err := time.Parse("2006-01-02 15:04:05.999 -0700", record[14])
		if err != nil {
			createdAt, err = time.Parse("2006-01-02 15:04:05 -0700", record[14])
			if err != nil {
				return nil, fmt.Errorf("failed to parse created_at at row %d: %w", i+2, err)
			}
		}

		updatedAt, err := time.Parse("2006-01-02 15:04:05.999 -0700", record[15])
		if err != nil {
			updatedAt, err = time.Parse("2006-01-02 15:04:05 -0700", record[15])
			if err != nil {
				return nil, fmt.Errorf("failed to parse updated_at at row %d: %w", i+2, err)
			}
		}

		config := &CSVReelStripConfig{
			ID:            record[0],
			Name:          record[1],
			GameMode:      record[2],
			Description:   record[3],
			Reel0StripID:  record[4],
			Reel1StripID:  record[5],
			Reel2StripID:  record[6],
			Reel3StripID:  record[7],
			Reel4StripID:  record[8],
			TargetRTP:     targetRTP,
			IsActive:      isActive,
			IsDefault:     isDefault,
			ActivatedAt:   activatedAt,
			DeactivatedAt: deactivatedAt,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
			CreatedBy:     record[16],
			Notes:         record[17],
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// getScriptDir returns the directory where this script is located
func getScriptDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		// Fallback to current working directory
		return os.Getwd()
	}
	return filepath.Dir(ex), nil
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
