package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/internal/game/reels"
	"github.com/slotmachine/backend/internal/game/rng"
	"github.com/slotmachine/backend/internal/game/symbols"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// ReelStripService implements reelstrip.Service
type ReelStripService struct {
	repo   reelstrip.Repository
	logger *logger.Logger
	rng    *rng.CryptoRNG
}

// NewReelStripService creates a new reel strip service
func NewReelStripService(repo reelstrip.Repository, log *logger.Logger) reelstrip.Service {
	return &ReelStripService{
		repo:   repo,
		logger: log,
		rng:    rng.NewCryptoRNG(),
	}
}

// GetRandomReelSet retrieves a random set of reel strips for a spin (deprecated - use config-based approach)
func (s *ReelStripService) GetRandomReelSet(ctx context.Context, gameMode string) (*reelstrip.ReelStripSet, error) {
	log := s.logger.WithTraceContext(ctx)

	// Validate game mode
	if err := s.validateGameMode(gameMode); err != nil {
		return nil, err
	}

	set := &reelstrip.ReelStripSet{
		GameMode: reelstrip.GameMode(gameMode),
	}

	// For each reel (0-4), get a random active strip
	for reelNum := 0; reelNum < 5; reelNum++ {
		strips, err := s.repo.GetByGameModeAndReel(ctx, gameMode, reelNum)
		if err != nil {
			log.Error().
				Err(err).
				Str("game_mode", gameMode).
				Int("reel_number", reelNum).
				Msg("Failed to get strips for reel")
			return nil, fmt.Errorf("failed to get strips for reel %d: %w", reelNum, err)
		}

		if len(strips) == 0 {
			return nil, fmt.Errorf("no active strip found for reel %d: %w", reelNum, reelstrip.ErrIncompleteSet)
		}

		// Select random strip from available strips for this reel
		randomIndex, err := s.rng.IntRange(0, len(strips)-1)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random index for reel %d: %w", reelNum, err)
		}
		strip := strips[randomIndex]

		// Validate integrity
		if err := s.ValidateStripIntegrity(strip); err != nil {
			log.Warn().
				Err(err).
				Str("strip_id", strip.ID.String()).
				Int("reel_number", reelNum).
				Msg("Strip integrity validation failed")
			// Continue with warning, don't fail the request
		}

		set.Strips[reelNum] = strip
	}

	if !set.IsComplete() {
		return nil, reelstrip.ErrIncompleteSet
	}

	return set, nil
}

// GenerateAndSaveStrips generates new reel strips and saves them to database
func (s *ReelStripService) GenerateAndSaveStrips(ctx context.Context, gameMode string, count int, version int) error {
	if err := s.validateGameMode(gameMode); err != nil {
		return err
	}

	if count <= 0 {
		return fmt.Errorf("count must be greater than 0")
	}

	s.logger.Info().
		Str("game_mode", gameMode).
		Int("count", count).
		Int("version", version).
		Msg("Generating reel strips")

	var allStrips []*reelstrip.ReelStrip
	isFreeSpin := gameMode == string(reelstrip.FreeSpins)

	// Generate 'count' sets of complete reel strips (5 reels per set)
	for setIdx := 0; setIdx < count; setIdx++ {
		for reelNum := 0; reelNum < 5; reelNum++ {
			// Get appropriate weights for this reel
			var weights symbols.ReelWeights
			if isFreeSpin {
				weights = symbols.GetFreeSpinsWeights(reelNum)
			} else {
				weights = symbols.GetBaseGameWeights(reelNum)
			}

			// Generate reel strip using existing game logic
			stripData, err := reels.GenerateReelStrip(weights, s.rng)
			if err != nil {
				return fmt.Errorf("failed to generate strip for reel %d: %w", reelNum, err)
			}

			// Convert reels.ReelStrip ([]string) to regular []string
			stripSlice := []string(stripData)

			// Calculate checksum
			checksum := s.calculateChecksum(stripSlice)

			strip := &reelstrip.ReelStrip{
				GameMode:    gameMode,
				ReelNumber:  reelNum,
				StripData:   stripSlice,
				Checksum:    checksum,
				StripLength: len(stripSlice),
				IsActive:    true,
			}

			allStrips = append(allStrips, strip)
		}

		s.logger.Debug().
			Int("set", setIdx+1).
			Int("total_sets", count).
			Msg("Generated reel strip set")
	}

	// Save all strips in batch
	if err := s.repo.CreateBatch(ctx, allStrips); err != nil {
		return fmt.Errorf("failed to save reel strips: %w", err)
	}

	s.logger.Info().
		Int("total_strips", len(allStrips)).
		Str("game_mode", gameMode).
		Msg("Successfully saved reel strips")

	return nil
}

// GenerateAndSaveStripSet generates one complete set of 5 reel strips and returns their IDs
func (s *ReelStripService) GenerateAndSaveStripSet(ctx context.Context, gameMode string) ([5]uuid.UUID, error) {
	log := s.logger.WithTraceContext(ctx)
	var stripIDs [5]uuid.UUID

	if err := s.validateGameMode(gameMode); err != nil {
		return stripIDs, err
	}

	isFreeSpin := gameMode == string(reelstrip.FreeSpins)

	// Generate one complete set (5 strips, one per reel)
	generatedStrips, err := reels.GenerateAllReelStrips(isFreeSpin, s.rng)
	if err != nil {
		return stripIDs, fmt.Errorf("failed to generate reel strips: %w", err)
	}

	// Create strip models for each reel
	var allStrips []*reelstrip.ReelStrip
	for reelNum := 0; reelNum < 5; reelNum++ {
		stripData := generatedStrips[reelNum]

		// Convert reels.ReelStrip ([]string) to regular []string
		stripSlice := []string(stripData)

		// Calculate checksum
		checksum := s.calculateChecksum(stripSlice)

		strip := &reelstrip.ReelStrip{
			ID:          uuid.New(),
			GameMode:    gameMode,
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
	if err := s.repo.CreateBatch(ctx, allStrips); err != nil {
		return stripIDs, fmt.Errorf("failed to save reel strips: %w", err)
	}

	log.Info().
		Int("strips_count", len(allStrips)).
		Str("game_mode", gameMode).
		Msg("Successfully generated and saved reel strip set")

	return stripIDs, nil
}

// GetStripByID retrieves a specific reel strip by ID
func (s *ReelStripService) GetStripByID(ctx context.Context, id uuid.UUID) (*reelstrip.ReelStrip, error) {
	return s.repo.GetByID(ctx, id)
}

// GetActiveStripsCount returns the count of active strips per reel for a game mode
func (s *ReelStripService) GetActiveStripsCount(ctx context.Context, gameMode string) (map[int]int, error) {
	if err := s.validateGameMode(gameMode); err != nil {
		return nil, err
	}

	return s.repo.CountActive(ctx, gameMode)
}

// RotateStrips is deprecated - use version management at config level instead
// Kept for backward compatibility but now warns about deprecation
func (s *ReelStripService) RotateStrips(ctx context.Context, gameMode string, newVersion int, count int) error {
	s.logger.Warn().Msg("RotateStrips is deprecated - use config-based version management")
	return fmt.Errorf("RotateStrips is deprecated - use CreateConfig and SetDefaultConfig instead")
}

// ValidateStripIntegrity validates the checksum of a reel strip
func (s *ReelStripService) ValidateStripIntegrity(strip *reelstrip.ReelStrip) error {
	if strip == nil {
		return fmt.Errorf("strip is nil")
	}

	expectedChecksum := s.calculateChecksum(strip.StripData)
	if expectedChecksum != strip.Checksum {
		return reelstrip.ErrChecksumMismatch
	}

	// Validate strip length
	if len(strip.StripData) != strip.StripLength {
		return reelstrip.ErrInvalidStripLength
	}

	return nil
}

// Helper functions

func (s *ReelStripService) validateGameMode(gameMode string) error {
	if gameMode != string(reelstrip.BaseGame) && gameMode != string(reelstrip.FreeSpins) && gameMode != string(reelstrip.Both) && gameMode != string(reelstrip.BonusSpinTrigger) {
		return reelstrip.ErrInvalidGameMode
	}
	return nil
}

func (s *ReelStripService) calculateChecksum(stripData []string) string {
	// Convert to JSON for consistent hashing
	jsonData, _ := json.Marshal(stripData)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// ===== New Config-Based Methods =====

// GetReelSetForPlayer retrieves the reel strip set for a specific player
// This is the main method - it handles player assignments, defaults, and fallbacks
func (s *ReelStripService) GetReelSetForPlayer(ctx context.Context, playerID uuid.UUID, gameMode string) (*reelstrip.ReelStripConfigSet, error) {
	log := s.logger.WithTraceContext(ctx)

	if err := s.validateGameMode(gameMode); err != nil {
		return nil, err
	}

	// Priority 1: Check player assignment table
	assignment, err := s.repo.GetPlayerAssignment(ctx, playerID)
	if err == nil && assignment != nil {
		var configID *uuid.UUID
		if gameMode == string(reelstrip.BaseGame) && assignment.BaseGameConfigID != nil {
			configID = assignment.BaseGameConfigID
		} else if gameMode == string(reelstrip.FreeSpins) && assignment.FreeSpinsConfigID != nil {
			configID = assignment.FreeSpinsConfigID
		}

		if configID != nil {
			configSet, err := s.repo.GetSetByConfigID(ctx, *configID)
			if err == nil {
				if assignment.ExpiresAt != nil {
					ttl := time.Until(*assignment.ExpiresAt)
					configSet.TTL = &ttl
				}
				return configSet, nil
			}
			log.Warn().Err(err).Str("config_id", configID.String()).Msg("Failed to load assigned config, falling back")
		}
	}

	// Priority 2: Check default configuration
	defaultConfig, err := s.repo.GetDefaultConfig(ctx, gameMode)
	if err == nil && defaultConfig != nil {
		configSet, err := s.repo.GetSetByConfigID(ctx, defaultConfig.ID)
		if err == nil {
			return configSet, nil
		}
		log.Warn().Err(err).Msg("Failed to load default config, falling back to legacy")
	}

	// Priority 3: Fallback to legacy random selection (deprecated)
	log.Warn().Msg("No config found, using legacy random selection (deprecated)")
	legacySet, err := s.GetRandomReelSet(ctx, gameMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get reel set for player: %w", err)
	}

	// Convert legacy set to config set
	return &reelstrip.ReelStripConfigSet{
		Config: nil, // No config in legacy mode
		Strips: legacySet.Strips,
	}, nil
}

// GetReelSetByConfig retrieves a reel strip set by configuration ID
func (s *ReelStripService) GetReelSetByConfig(ctx context.Context, configID uuid.UUID) (*reelstrip.ReelStripConfigSet, error) {
	log := s.logger.WithTraceContext(ctx)

	configSet, err := s.repo.GetSetByConfigID(ctx, configID)
	if err != nil {
		log.Error().Err(err).Str("config_id", configID.String()).Msg("Failed to get reel set by config")
		return nil, fmt.Errorf("failed to get reel set by config: %w", err)
	}

	// Validate integrity
	for i, strip := range configSet.Strips {
		if err := s.ValidateStripIntegrity(strip); err != nil {
			log.Warn().
				Err(err).
				Str("strip_id", strip.ID.String()).
				Int("reel_number", i).
				Msg("Strip integrity validation failed")
		}
	}

	return configSet, nil
}

// GetDefaultReelSet retrieves the default reel strip set for a game mode
func (s *ReelStripService) GetDefaultReelSet(ctx context.Context, gameMode string) (*reelstrip.ReelStripConfigSet, error) {
	if err := s.validateGameMode(gameMode); err != nil {
		return nil, err
	}

	defaultConfig, err := s.repo.GetDefaultConfig(ctx, gameMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get default config: %w", err)
	}

	return s.GetReelSetByConfig(ctx, defaultConfig.ID)
}

// CreateConfig creates a new reel strip configuration
func (s *ReelStripService) CreateConfig(ctx context.Context, name, gameMode, description string, reelStripIDs [5]uuid.UUID, targetRTP float64, extraInfoJSON []byte) (*reelstrip.ReelStripConfig, error) {
	log := s.logger.WithTraceContext(ctx)

	if err := s.validateGameMode(gameMode); err != nil {
		return nil, err
	}

	// Validate that all reel strip IDs exist
	strips, err := s.repo.GetByIDs(ctx, reelStripIDs[:])
	if err != nil {
		return nil, fmt.Errorf("failed to validate reel strips: %w", err)
	}
	if len(strips) != 5 {
		return nil, reelstrip.ErrIncompleteSet
	}

	config := &reelstrip.ReelStripConfig{
		Name:         name,
		GameMode:     gameMode,
		Description:  description,
		Reel0StripID: reelStripIDs[0],
		Reel1StripID: reelStripIDs[1],
		Reel2StripID: reelStripIDs[2],
		Reel3StripID: reelStripIDs[3],
		Reel4StripID: reelStripIDs[4],
		TargetRTP:    targetRTP,
		IsActive:     true,
		IsDefault:    false,
		Options:      extraInfoJSON,
	}

	if err := s.repo.CreateConfig(ctx, config); err != nil {
		log.Error().Err(err).Str("name", name).Msg("Failed to create config")
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	log.Info().
		Str("config_id", config.ID.String()).
		Str("name", name).
		Str("game_mode", gameMode).
		Msg("Created reel strip config")

	return config, nil
}

// GetConfigByID retrieves a configuration by ID
func (s *ReelStripService) GetConfigByID(ctx context.Context, id uuid.UUID) (*reelstrip.ReelStripConfig, error) {
	return s.repo.GetConfigByID(ctx, id)
}

// GetConfigByName retrieves a configuration by name
func (s *ReelStripService) GetConfigByName(ctx context.Context, name string) (*reelstrip.ReelStripConfig, error) {
	return s.repo.GetConfigByName(ctx, name)
}

// ListConfigs retrieves reel strip configurations with filtering and pagination
func (s *ReelStripService) ListConfigs(ctx context.Context, filters *reelstrip.ConfigListFilters) ([]*reelstrip.ReelStripConfig, int64, error) {
	// Validate game mode if provided
	if filters.GameMode != nil && *filters.GameMode != "" {
		if err := s.validateGameMode(*filters.GameMode); err != nil {
			return nil, 0, err
		}
	}

	return s.repo.ListConfigs(ctx, filters)
}

// SetDefaultConfig sets a configuration as the default for its game mode
func (s *ReelStripService) SetDefaultConfig(ctx context.Context, configID uuid.UUID, gameMode string) error {
	log := s.logger.WithTraceContext(ctx)

	if err := s.validateGameMode(gameMode); err != nil {
		return err
	}

	if err := s.repo.SetDefaultConfig(ctx, configID, gameMode); err != nil {
		log.Error().Err(err).Str("config_id", configID.String()).Msg("Failed to set default config")
		return fmt.Errorf("failed to set default config: %w", err)
	}

	log.Info().
		Str("config_id", configID.String()).
		Str("game_mode", gameMode).
		Msg("Set default config")

	return nil
}

// ActivateConfig activates a configuration
func (s *ReelStripService) ActivateConfig(ctx context.Context, configID uuid.UUID) error {
	config, err := s.repo.GetConfigByID(ctx, configID)
	if err != nil {
		return err
	}

	config.IsActive = true
	return s.repo.UpdateConfig(ctx, config)
}

// DeactivateConfig deactivates a configuration
func (s *ReelStripService) DeactivateConfig(ctx context.Context, configID uuid.UUID) error {
	config, err := s.repo.GetConfigByID(ctx, configID)
	if err != nil {
		return err
	}

	config.IsActive = false
	return s.repo.UpdateConfig(ctx, config)
}

// AssignConfigToPlayer assigns a configuration to a player
func (s *ReelStripService) AssignConfigToPlayer(ctx context.Context, playerID, configID uuid.UUID, gameMode, reason, assignedBy string, expiresAt *time.Time) error {
	log := s.logger.WithTraceContext(ctx)

	if err := s.validateGameMode(gameMode); err != nil {
		return err
	}

	// Verify config exists
	config, err := s.repo.GetConfigByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("config not found: %w", err)
	}

	// Verify config game mode matches
	if config.GameMode != gameMode {
		return fmt.Errorf("config game mode %s does not match requested game mode %s", config.GameMode, gameMode)
	}

	// Check if assignment already exists
	existing, err := s.repo.GetPlayerAssignment(ctx, playerID)
	if err == nil && existing != nil {
		// Update existing assignment
		if gameMode == string(reelstrip.BaseGame) {
			existing.BaseGameConfigID = &configID
		} else {
			existing.FreeSpinsConfigID = &configID
		}
		existing.Reason = reason
		if assignedBy != "" {
			existing.AssignedBy = assignedBy
		}
		if expiresAt != nil {
			existing.ExpiresAt = expiresAt
		}

		if err := s.repo.UpdateAssignment(ctx, existing); err != nil {
			return fmt.Errorf("failed to update assignment: %w", err)
		}
	} else {
		// Create new assignment
		assignment := &reelstrip.PlayerReelStripAssignment{
			PlayerID:   playerID,
			Reason:     reason,
			AssignedBy: assignedBy,
			ExpiresAt:  expiresAt,
			IsActive:   true,
		}

		if gameMode == string(reelstrip.BaseGame) {
			assignment.BaseGameConfigID = &configID
		} else {
			assignment.FreeSpinsConfigID = &configID
		}

		if err := s.repo.CreateAssignment(ctx, assignment); err != nil {
			return fmt.Errorf("failed to create assignment: %w", err)
		}
	}

	log.Info().
		Str("player_id", playerID.String()).
		Str("config_id", configID.String()).
		Str("game_mode", gameMode).
		Str("reason", reason).
		Msg("Assigned config to player")

	return nil
}

// GetPlayerAssignment retrieves a player's assignment
func (s *ReelStripService) GetPlayerAssignment(ctx context.Context, playerID uuid.UUID) (*reelstrip.PlayerReelStripAssignment, error) {
	return s.repo.GetPlayerAssignment(ctx, playerID)
}

// RemovePlayerAssignment removes a player's assignment
func (s *ReelStripService) RemovePlayerAssignment(ctx context.Context, playerID uuid.UUID) error {
	log := s.logger.WithTraceContext(ctx)

	assignment, err := s.repo.GetPlayerAssignment(ctx, playerID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteAssignment(ctx, assignment.ID); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to remove assignment")
		return fmt.Errorf("failed to remove assignment: %w", err)
	}

	log.Info().
		Str("player_id", playerID.String()).
		Msg("Removed player assignment")

	return nil
}
