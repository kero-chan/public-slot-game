package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupReelStripTestDB creates an in-memory SQLite database for testing reel strips
func setupReelStripTestDB(t *testing.T) (*gorm.DB, *cache.Cache) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to test database")

	// Create reel_strips table
	err = db.Exec(`
		CREATE TABLE reel_strips (
			id TEXT PRIMARY KEY,
			game_mode TEXT NOT NULL,
			reel_number INTEGER NOT NULL,
			strip_data TEXT NOT NULL,
			checksum TEXT NOT NULL UNIQUE,
			strip_length INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_active INTEGER DEFAULT 1,
			notes TEXT
		)
	`).Error
	require.NoError(t, err, "Failed to create reel_strips table")

	// Create reel_strip_configs table
	err = db.Exec(`
		CREATE TABLE reel_strip_configs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			game_mode TEXT NOT NULL,
			description TEXT,
			reel0_strip_id TEXT NOT NULL,
			reel1_strip_id TEXT NOT NULL,
			reel2_strip_id TEXT NOT NULL,
			reel3_strip_id TEXT NOT NULL,
			reel4_strip_id TEXT NOT NULL,
			target_rtp REAL,
			options TEXT,
			is_active INTEGER DEFAULT 1,
			is_default INTEGER DEFAULT 0,
			activated_at DATETIME,
			deactivated_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT,
			notes TEXT
		)
	`).Error
	require.NoError(t, err, "Failed to create reel_strip_configs table")

	// Create player_reel_strip_assignments table
	err = db.Exec(`
		CREATE TABLE player_reel_strip_assignments (
			id TEXT PRIMARY KEY,
			player_id TEXT NOT NULL,
			base_game_config_id TEXT,
			free_spins_config_id TEXT,
			assigned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			assigned_by TEXT,
			reason TEXT,
			expires_at DATETIME,
			is_active INTEGER DEFAULT 1
		)
	`).Error
	require.NoError(t, err, "Failed to create player_reel_strip_assignments table")

	// Create cache instance with minimal config for testing
	c := cache.NewCache(cache.NewCacheParams{
		Channel: "test",
		Config: &config.Config{
			App: config.AppConfig{
				Name: "test",
				Env:  "test",
			},
		},
	})

	return db, c
}

// createTestReelStrip creates a test reel strip
func createTestReelStrip(gameMode string, reelNumber int) *reelstrip.ReelStrip {
	stripData := []string{"A", "K", "Q", "J", "10", "9", "A", "K"}
	return &reelstrip.ReelStrip{
		ID:          uuid.New(),
		GameMode:    gameMode,
		ReelNumber:  reelNumber,
		StripData:   stripData,
		Checksum:    uuid.New().String(), // Unique checksum
		StripLength: len(stripData),
		IsActive:    true,
		Notes:       "Test strip",
	}
}

// ============================================================================
// ReelStrip CRUD TESTS
// ============================================================================

func TestReelStripGormRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("should create reel strip successfully", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		strip := createTestReelStrip("base_game", 0)

		err := repo.Create(ctx, strip)

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, strip.ID)

		// Verify it was saved
		retrieved, err := repo.GetByID(ctx, strip.ID)
		require.NoError(t, err)
		assert.Equal(t, strip.GameMode, retrieved.GameMode)
		assert.Equal(t, strip.ReelNumber, retrieved.ReelNumber)
		assert.Equal(t, len(strip.StripData), retrieved.StripLength)
	})

	t.Run("should create batch of strips", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		// Create 5 strips for a complete set
		strips := []*reelstrip.ReelStrip{}
		for i := 0; i < 5; i++ {
			strip := createTestReelStrip("base_game", i)
			strips = append(strips, strip)
		}

		err := repo.CreateBatch(ctx, strips)

		require.NoError(t, err)

		// Verify all were saved
		for i, strip := range strips {
			retrieved, err := repo.GetByID(ctx, strip.ID)
			require.NoError(t, err)
			assert.Equal(t, i, retrieved.ReelNumber)
		}
	})
}

func TestReelStripGormRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get reel strip by ID successfully", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		strip := createTestReelStrip("base_game", 0)
		err := repo.Create(ctx, strip)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, strip.ID)

		require.NoError(t, err)
		assert.Equal(t, strip.ID, retrieved.ID)
		assert.Equal(t, strip.GameMode, retrieved.GameMode)
	})

	t.Run("should return error for non-existent ID", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		nonExistentID := uuid.New()

		retrieved, err := repo.GetByID(ctx, nonExistentID)

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, reelstrip.ErrReelStripNotFound, err)
	})
}

func TestReelStripGormRepository_GetAllActive(t *testing.T) {
	ctx := context.Background()

	t.Run("should get all active strips for game mode", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		// Create strips for different game modes
		for i := 0; i < 5; i++ {
			strip := createTestReelStrip("base_game", i)
			err := repo.Create(ctx, strip)
			require.NoError(t, err)
		}

		for i := 0; i < 3; i++ {
			strip := createTestReelStrip("free_spins", i)
			err := repo.Create(ctx, strip)
			require.NoError(t, err)
		}

		strips, err := repo.GetAllActive(ctx, "base_game")

		require.NoError(t, err)
		assert.Len(t, strips, 5)
	})
}

func TestReelStripGormRepository_GetByGameModeAndReel(t *testing.T) {
	ctx := context.Background()

	t.Run("should get strips by game mode and reel number", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		strip := createTestReelStrip("base_game", 2)
		err := repo.Create(ctx, strip)
		require.NoError(t, err)

		strips, err := repo.GetByGameModeAndReel(ctx, "base_game", 2)

		require.NoError(t, err)
		assert.Len(t, strips, 1)
		assert.Equal(t, 2, strips[0].ReelNumber)
	})

	t.Run("should return error for invalid reel number", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		strips, err := repo.GetByGameModeAndReel(ctx, "base_game", 5)

		assert.Error(t, err)
		assert.Nil(t, strips)
		assert.Equal(t, reelstrip.ErrInvalidReelNumber, err)
	})
}

func TestReelStripGormRepository_CountActive(t *testing.T) {
	ctx := context.Background()

	t.Run("should count active strips per reel", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		// Create 2 strips for reel 0, 1 for reel 1
		for i := 0; i < 2; i++ {
			strip := createTestReelStrip("base_game", 0)
			err := repo.Create(ctx, strip)
			require.NoError(t, err)
		}

		strip := createTestReelStrip("base_game", 1)
		err := repo.Create(ctx, strip)
		require.NoError(t, err)

		counts, err := repo.CountActive(ctx, "base_game")

		require.NoError(t, err)
		assert.Equal(t, 2, counts[0])
		assert.Equal(t, 1, counts[1])
	})
}

func TestReelStripGormRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("should soft delete reel strip", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		strip := createTestReelStrip("base_game", 0)
		err := repo.Create(ctx, strip)
		require.NoError(t, err)

		err = repo.Delete(ctx, strip.ID)

		require.NoError(t, err)

		// Verify it's marked as inactive
		var saved reelstrip.ReelStrip
		err = db.First(&saved, "id = ?", strip.ID).Error
		require.NoError(t, err)
		assert.False(t, saved.IsActive)
	})
}

// ============================================================================
// ReelStripConfig TESTS
// ============================================================================

func TestReelStripGormRepository_CreateConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("should create config successfully", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		// Create 5 reel strips
		stripIDs := [5]uuid.UUID{}
		for i := 0; i < 5; i++ {
			strip := createTestReelStrip("base_game", i)
			err := repo.Create(ctx, strip)
			require.NoError(t, err)
			stripIDs[i] = strip.ID
		}

		config := &reelstrip.ReelStripConfig{
			ID:            uuid.New(),
			Name:          "test_config",
			GameMode:      "base_game",
			Reel0StripID:  stripIDs[0],
			Reel1StripID:  stripIDs[1],
			Reel2StripID:  stripIDs[2],
			Reel3StripID:  stripIDs[3],
			Reel4StripID:  stripIDs[4],
			IsDefault:     false,
			IsActive:      true,
			Description:   "Test configuration",
		}

		err := repo.CreateConfig(ctx, config)

		require.NoError(t, err)
	})
}

func TestReelStripGormRepository_GetConfigByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get config by ID", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		// Create strips and config
		stripIDs := [5]uuid.UUID{}
		for i := 0; i < 5; i++ {
			strip := createTestReelStrip("base_game", i)
			err := repo.Create(ctx, strip)
			require.NoError(t, err)
			stripIDs[i] = strip.ID
		}

		config := &reelstrip.ReelStripConfig{
			ID:           uuid.New(),
			Name:         "test_config",
			GameMode:     "base_game",
			Reel0StripID: stripIDs[0],
			Reel1StripID: stripIDs[1],
			Reel2StripID: stripIDs[2],
			Reel3StripID: stripIDs[3],
			Reel4StripID: stripIDs[4],
			IsDefault:    false,
			IsActive:     true,
		}
		err := repo.CreateConfig(ctx, config)
		require.NoError(t, err)

		retrieved, err := repo.GetConfigByID(ctx, config.ID)

		require.NoError(t, err)
		assert.Equal(t, config.Name, retrieved.Name)
	})
}

func TestReelStripGormRepository_SetDefaultConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("should set config as default", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		// Create strips
		stripIDs := [5]uuid.UUID{}
		for i := 0; i < 5; i++ {
			strip := createTestReelStrip("base_game", i)
			err := repo.Create(ctx, strip)
			require.NoError(t, err)
			stripIDs[i] = strip.ID
		}

		// Create first config as default
		config1 := &reelstrip.ReelStripConfig{
			ID:           uuid.New(),
			Name:         "config1",
			GameMode:     "base_game",
			Reel0StripID: stripIDs[0],
			Reel1StripID: stripIDs[1],
			Reel2StripID: stripIDs[2],
			Reel3StripID: stripIDs[3],
			Reel4StripID: stripIDs[4],
			IsDefault:    true,
			IsActive:     true,
		}
		err := repo.CreateConfig(ctx, config1)
		require.NoError(t, err)

		// Create second config
		config2 := &reelstrip.ReelStripConfig{
			ID:           uuid.New(),
			Name:         "config2",
			GameMode:     "base_game",
			Reel0StripID: stripIDs[0],
			Reel1StripID: stripIDs[1],
			Reel2StripID: stripIDs[2],
			Reel3StripID: stripIDs[3],
			Reel4StripID: stripIDs[4],
			IsDefault:    false,
			IsActive:     true,
		}
		err = repo.CreateConfig(ctx, config2)
		require.NoError(t, err)

		// Set config2 as default
		err = repo.SetDefaultConfig(ctx, config2.ID, "base_game")

		require.NoError(t, err)

		// Verify config2 is now default
		retrieved, err := repo.GetConfigByID(ctx, config2.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.IsDefault)

		// Verify config1 is no longer default
		retrieved1, err := repo.GetConfigByID(ctx, config1.ID)
		require.NoError(t, err)
		assert.False(t, retrieved1.IsDefault)
	})
}

func TestReelStripGormRepository_GetDefaultConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("should get default config for game mode", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		// Create strips
		stripIDs := [5]uuid.UUID{}
		for i := 0; i < 5; i++ {
			strip := createTestReelStrip("base_game", i)
			err := repo.Create(ctx, strip)
			require.NoError(t, err)
			stripIDs[i] = strip.ID
		}

		config := &reelstrip.ReelStripConfig{
			ID:           uuid.New(),
			Name:         "default_config",
			GameMode:     "base_game",
			Reel0StripID: stripIDs[0],
			Reel1StripID: stripIDs[1],
			Reel2StripID: stripIDs[2],
			Reel3StripID: stripIDs[3],
			Reel4StripID: stripIDs[4],
			IsDefault:    true,
			IsActive:     true,
		}
		err := repo.CreateConfig(ctx, config)
		require.NoError(t, err)

		retrieved, err := repo.GetDefaultConfig(ctx, "base_game")

		require.NoError(t, err)
		assert.Equal(t, config.Name, retrieved.Name)
		assert.True(t, retrieved.IsDefault)
	})

	t.Run("should return error when no default config exists", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		retrieved, err := repo.GetDefaultConfig(ctx, "base_game")

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, reelstrip.ErrNoDefaultConfig, err)
	})
}

func TestReelStripGormRepository_GetSetByConfigID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get complete reel strip set", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		// Create 5 reel strips
		stripIDs := [5]uuid.UUID{}
		for i := 0; i < 5; i++ {
			strip := createTestReelStrip("base_game", i)
			err := repo.Create(ctx, strip)
			require.NoError(t, err)
			stripIDs[i] = strip.ID
		}

		// Create config
		config := &reelstrip.ReelStripConfig{
			ID:           uuid.New(),
			Name:         "test_set",
			GameMode:     "base_game",
			Reel0StripID: stripIDs[0],
			Reel1StripID: stripIDs[1],
			Reel2StripID: stripIDs[2],
			Reel3StripID: stripIDs[3],
			Reel4StripID: stripIDs[4],
			IsDefault:    false,
			IsActive:     true,
		}
		err := repo.CreateConfig(ctx, config)
		require.NoError(t, err)

		// Get the complete set
		set, err := repo.GetSetByConfigID(ctx, config.ID)

		require.NoError(t, err)
		assert.NotNil(t, set)
		assert.True(t, set.IsComplete())
		for i := 0; i < 5; i++ {
			assert.NotNil(t, set.Strips[i])
			assert.Equal(t, i, set.Strips[i].ReelNumber)
		}
	})
}

// ============================================================================
// PlayerReelStripAssignment TESTS
// ============================================================================

func TestReelStripGormRepository_CreateAssignment(t *testing.T) {
	ctx := context.Background()

	t.Run("should create player assignment successfully", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		playerID := uuid.New()
		baseConfigID := uuid.New()

		assignment := &reelstrip.PlayerReelStripAssignment{
			ID:               uuid.New(),
			PlayerID:         playerID,
			BaseGameConfigID: &baseConfigID,
			IsActive:         true,
		}

		err := repo.CreateAssignment(ctx, assignment)

		require.NoError(t, err)
	})
}

func TestReelStripGormRepository_GetPlayerAssignment(t *testing.T) {
	ctx := context.Background()

	t.Run("should get player assignment", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		playerID := uuid.New()
		baseConfigID := uuid.New()

		assignment := &reelstrip.PlayerReelStripAssignment{
			ID:               uuid.New(),
			PlayerID:         playerID,
			BaseGameConfigID: &baseConfigID,
			IsActive:         true,
		}
		err := repo.CreateAssignment(ctx, assignment)
		require.NoError(t, err)

		retrieved, err := repo.GetPlayerAssignment(ctx, playerID)

		require.NoError(t, err)
		assert.Equal(t, playerID, retrieved.PlayerID)
		assert.Equal(t, baseConfigID, *retrieved.BaseGameConfigID)
	})

	t.Run("should return empty assignment for player with no assignment", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		playerID := uuid.New()

		retrieved, err := repo.GetPlayerAssignment(ctx, playerID)

		require.NoError(t, err)
		assert.Equal(t, playerID, retrieved.PlayerID)
		assert.Nil(t, retrieved.BaseGameConfigID)
	})
}

func TestReelStripGormRepository_DeleteAssignment(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete assignment", func(t *testing.T) {
		db, c := setupReelStripTestDB(t)
		repo := NewReelStripGormRepository(db, c)

		playerID := uuid.New()
		baseConfigID := uuid.New()

		assignment := &reelstrip.PlayerReelStripAssignment{
			ID:               uuid.New(),
			PlayerID:         playerID,
			BaseGameConfigID: &baseConfigID,
			IsActive:         true,
		}
		err := repo.CreateAssignment(ctx, assignment)
		require.NoError(t, err)

		err = repo.DeleteAssignment(ctx, assignment.ID)

		require.NoError(t, err)

		// Verify the record is deleted
		var saved reelstrip.PlayerReelStripAssignment
		err = db.First(&saved, "id = ?", assignment.ID).Error
		assert.Error(t, err) // Should return "record not found"
	})
}
