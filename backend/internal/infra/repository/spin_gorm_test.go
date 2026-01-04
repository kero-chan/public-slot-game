package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/spin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupSpinTestDB creates an in-memory SQLite database for testing spins
func setupSpinTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to test database")

	// Create table manually with SQLite-compatible syntax
	err = db.Exec(`
		CREATE TABLE spins (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			player_id TEXT NOT NULL,
			bet_amount REAL NOT NULL,
			balance_before REAL NOT NULL,
			balance_after REAL NOT NULL,
			grid TEXT NOT NULL,
			cascades TEXT,
			total_win REAL DEFAULT 0.00,
			scatter_count INTEGER DEFAULT 0,
			reel_positions TEXT NOT NULL,
			is_free_spin INTEGER DEFAULT 0,
			free_spins_session_id TEXT,
			free_spins_triggered INTEGER DEFAULT 0,
			game_mode TEXT DEFAULT NULL,
			game_mode_cost REAL DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	require.NoError(t, err, "Failed to create spins table")

	// Create indices
	db.Exec("CREATE INDEX idx_spins_session_id ON spins(session_id)")
	db.Exec("CREATE INDEX idx_spins_player_id ON spins(player_id)")
	db.Exec("CREATE INDEX idx_spins_is_free_spin ON spins(is_free_spin)")
	db.Exec("CREATE INDEX idx_spins_free_spins_session_id ON spins(free_spins_session_id)")
	db.Exec("CREATE INDEX idx_spins_created_at ON spins(created_at)")

	return db
}

// createTestSpin creates a test spin with default values
func createTestSpin(playerID, sessionID uuid.UUID) *spin.Spin {
	grid := spin.Grid{
		{"A", "K", "Q", "J", "10", "9"},
		{"K", "Q", "J", "10", "9", "A"},
		{"Q", "J", "10", "9", "A", "K"},
		{"J", "10", "9", "A", "K", "Q"},
		{"10", "9", "A", "K", "Q", "J"},
	}

	return &spin.Spin{
		ID:                 uuid.New(),
		SessionID:          sessionID,
		PlayerID:           playerID,
		BetAmount:          100.0,
		BalanceBefore:      10000.0,
		BalanceAfter:       9900.0,
		Grid:               grid,
		Cascades:           spin.Cascades{},
		TotalWin:           0.0,
		ScatterCount:       0,
		ReelPositions:      []int{0, 1, 2, 3, 4},
		IsFreeSpin:         false,
		FreeSpinsSessionID: nil,
		FreeSpinsTriggered: false,
		CreatedAt:          time.Now().UTC(),
	}
}

// ============================================================================
// Create TESTS
// ============================================================================

func TestSpinGormRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("should create spin successfully", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()
		s := createTestSpin(playerID, sessionID)

		err := repo.Create(ctx, s)

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, s.ID)

		// Verify it was saved
		var saved spin.Spin
		err = db.First(&saved, "id = ?", s.ID).Error
		require.NoError(t, err)
		assert.Equal(t, s.PlayerID, saved.PlayerID)
		assert.Equal(t, s.SessionID, saved.SessionID)
		assert.Equal(t, s.BetAmount, saved.BetAmount)
		assert.Equal(t, s.TotalWin, saved.TotalWin)
	})

	t.Run("should store grid and cascades as JSON", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()
		s := createTestSpin(playerID, sessionID)

		// Add cascades
		s.Cascades = spin.Cascades{
			{
				CascadeNumber:   1,
				Multiplier:      1,
				TotalCascadeWin: 100.0,
				Wins: []spin.CascadeWin{
					{
						Symbol:    "A",
						Count:     3,
						Ways:      1,
						Payout:    1.0,
						WinAmount: 100.0,
					},
				},
			},
		}

		err := repo.Create(ctx, s)
		require.NoError(t, err)

		// Verify JSON fields were saved
		retrieved, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.Len(t, retrieved.Grid, 5)
		assert.Len(t, retrieved.Cascades, 1)
		assert.Equal(t, 1, retrieved.Cascades[0].CascadeNumber)
	})

	t.Run("should handle free spin attributes", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()
		freeSpinsSessionID := uuid.New()
		s := createTestSpin(playerID, sessionID)
		s.IsFreeSpin = true
		s.FreeSpinsSessionID = &freeSpinsSessionID
		s.FreeSpinsTriggered = false

		err := repo.Create(ctx, s)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.IsFreeSpin)
		assert.NotNil(t, retrieved.FreeSpinsSessionID)
		assert.Equal(t, freeSpinsSessionID, *retrieved.FreeSpinsSessionID)
	})
}

// ============================================================================
// GetByID TESTS
// ============================================================================

func TestSpinGormRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get spin by ID successfully", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()
		s := createTestSpin(playerID, sessionID)
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, s.ID)

		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, s.ID, retrieved.ID)
		assert.Equal(t, s.PlayerID, retrieved.PlayerID)
		assert.Equal(t, s.SessionID, retrieved.SessionID)
	})

	t.Run("should return error for non-existent ID", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		nonExistentID := uuid.New()

		retrieved, err := repo.GetByID(ctx, nonExistentID)

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, spin.ErrSpinNotFound, err)
	})
}

// ============================================================================
// GetBySession TESTS
// ============================================================================

func TestSpinGormRepository_GetBySession(t *testing.T) {
	ctx := context.Background()

	t.Run("should get spins by session successfully", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()

		// Create multiple spins for the session
		s1 := createTestSpin(playerID, sessionID)
		err := repo.Create(ctx, s1)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		s2 := createTestSpin(playerID, sessionID)
		err = repo.Create(ctx, s2)
		require.NoError(t, err)

		spins, err := repo.GetBySession(ctx, sessionID)

		require.NoError(t, err)
		assert.Len(t, spins, 2)
		// Should be ordered by created_at ASC
		assert.Equal(t, s1.ID, spins[0].ID)
		assert.Equal(t, s2.ID, spins[1].ID)
	})

	t.Run("should return empty list when no spins", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		sessionID := uuid.New()

		spins, err := repo.GetBySession(ctx, sessionID)

		require.NoError(t, err)
		assert.NotNil(t, spins)
		assert.Len(t, spins, 0)
	})

	t.Run("should only return spins for specified session", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		session1ID := uuid.New()
		session2ID := uuid.New()

		// Create spin for session 1
		s1 := createTestSpin(playerID, session1ID)
		err := repo.Create(ctx, s1)
		require.NoError(t, err)

		// Create spin for session 2
		s2 := createTestSpin(playerID, session2ID)
		err = repo.Create(ctx, s2)
		require.NoError(t, err)

		spins, err := repo.GetBySession(ctx, session1ID)

		require.NoError(t, err)
		assert.Len(t, spins, 1)
		assert.Equal(t, s1.ID, spins[0].ID)
	})
}

// ============================================================================
// GetByPlayer TESTS
// ============================================================================

func TestSpinGormRepository_GetByPlayer(t *testing.T) {
	ctx := context.Background()

	t.Run("should get spins by player successfully", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()

		// Create multiple spins
		s1 := createTestSpin(playerID, sessionID)
		err := repo.Create(ctx, s1)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		s2 := createTestSpin(playerID, sessionID)
		err = repo.Create(ctx, s2)
		require.NoError(t, err)

		spins, err := repo.GetByPlayer(ctx, playerID, 10, 0)

		require.NoError(t, err)
		assert.Len(t, spins, 2)
		// Should be ordered by created_at DESC
		assert.Equal(t, s2.ID, spins[0].ID)
		assert.Equal(t, s1.ID, spins[1].ID)
	})

	t.Run("should paginate results", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()

		// Create 5 spins
		for i := 0; i < 5; i++ {
			s := createTestSpin(playerID, sessionID)
			err := repo.Create(ctx, s)
			require.NoError(t, err)
			time.Sleep(5 * time.Millisecond)
		}

		// Get first page (2 items)
		page1, err := repo.GetByPlayer(ctx, playerID, 2, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 2)

		// Get second page
		page2, err := repo.GetByPlayer(ctx, playerID, 2, 2)
		require.NoError(t, err)
		assert.Len(t, page2, 2)

		// Verify no overlap
		assert.NotEqual(t, page1[0].ID, page2[0].ID)
	})
}

// ============================================================================
// GetByPlayerInTimeRange TESTS
// ============================================================================

func TestSpinGormRepository_GetByPlayerInTimeRange(t *testing.T) {
	ctx := context.Background()

	t.Run("should get spins in time range successfully", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()

		now := time.Now().UTC()
		start := now.Add(-1 * time.Hour)
		end := now.Add(1 * time.Hour)

		// Create spin within range
		s := createTestSpin(playerID, sessionID)
		s.CreatedAt = now
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		spins, err := repo.GetByPlayerInTimeRange(ctx, playerID, start, end, 10, 0)

		require.NoError(t, err)
		assert.Len(t, spins, 1)
		assert.Equal(t, s.ID, spins[0].ID)
	})

	t.Run("should exclude spins outside time range", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()

		now := time.Now().UTC()
		start := now.Add(-2 * time.Hour)
		end := now.Add(-1 * time.Hour)

		// Create spin outside range
		s := createTestSpin(playerID, sessionID)
		s.CreatedAt = now
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		spins, err := repo.GetByPlayerInTimeRange(ctx, playerID, start, end, 10, 0)

		require.NoError(t, err)
		assert.Len(t, spins, 0)
	})
}

// ============================================================================
// GetByFreeSpinsSession TESTS
// ============================================================================

func TestSpinGormRepository_GetByFreeSpinsSession(t *testing.T) {
	ctx := context.Background()

	t.Run("should get spins by free spins session successfully", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()
		freeSpinsSessionID := uuid.New()

		// Create free spins
		s1 := createTestSpin(playerID, sessionID)
		s1.IsFreeSpin = true
		s1.FreeSpinsSessionID = &freeSpinsSessionID
		err := repo.Create(ctx, s1)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		s2 := createTestSpin(playerID, sessionID)
		s2.IsFreeSpin = true
		s2.FreeSpinsSessionID = &freeSpinsSessionID
		err = repo.Create(ctx, s2)
		require.NoError(t, err)

		spins, err := repo.GetByFreeSpinsSession(ctx, freeSpinsSessionID)

		require.NoError(t, err)
		assert.Len(t, spins, 2)
		// Should be ordered by created_at ASC
		assert.Equal(t, s1.ID, spins[0].ID)
		assert.Equal(t, s2.ID, spins[1].ID)
	})

	t.Run("should return empty list when no free spins", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		freeSpinsSessionID := uuid.New()

		spins, err := repo.GetByFreeSpinsSession(ctx, freeSpinsSessionID)

		require.NoError(t, err)
		assert.NotNil(t, spins)
		assert.Len(t, spins, 0)
	})
}

// ============================================================================
// Count TESTS
// ============================================================================

func TestSpinGormRepository_Count(t *testing.T) {
	ctx := context.Background()

	t.Run("should count spins for player successfully", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()

		// Create multiple spins
		for i := 0; i < 5; i++ {
			s := createTestSpin(playerID, sessionID)
			err := repo.Create(ctx, s)
			require.NoError(t, err)
		}

		count, err := repo.Count(ctx, playerID)

		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})

	t.Run("should return zero for player with no spins", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()

		count, err := repo.Count(ctx, playerID)

		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("should only count spins for specified player", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		player1ID := uuid.New()
		player2ID := uuid.New()
		sessionID := uuid.New()

		// Create spins for player 1
		for i := 0; i < 3; i++ {
			s := createTestSpin(player1ID, sessionID)
			err := repo.Create(ctx, s)
			require.NoError(t, err)
		}

		// Create spins for player 2
		for i := 0; i < 2; i++ {
			s := createTestSpin(player2ID, sessionID)
			err := repo.Create(ctx, s)
			require.NoError(t, err)
		}

		count, err := repo.Count(ctx, player1ID)

		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}

// ============================================================================
// CountInTimeRange TESTS
// ============================================================================

func TestSpinGormRepository_CountInTimeRange(t *testing.T) {
	ctx := context.Background()

	t.Run("should count spins in time range successfully", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()

		now := time.Now().UTC()
		start := now.Add(-1 * time.Hour)
		end := now.Add(1 * time.Hour)

		// Create spins within range
		for i := 0; i < 3; i++ {
			s := createTestSpin(playerID, sessionID)
			s.CreatedAt = now
			err := repo.Create(ctx, s)
			require.NoError(t, err)
		}

		count, err := repo.CountInTimeRange(ctx, playerID, start, end)

		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	t.Run("should exclude spins outside time range", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()

		now := time.Now().UTC()
		start := now.Add(-2 * time.Hour)
		end := now.Add(-1 * time.Hour)

		// Create spins outside range
		for i := 0; i < 3; i++ {
			s := createTestSpin(playerID, sessionID)
			s.CreatedAt = now // This is outside the range
			err := repo.Create(ctx, s)
			require.NoError(t, err)
		}

		count, err := repo.CountInTimeRange(ctx, playerID, start, end)

		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("should handle boundary conditions", func(t *testing.T) {
		db := setupSpinTestDB(t)
		repo := NewSpinGormRepository(db)

		playerID := uuid.New()
		sessionID := uuid.New()

		now := time.Now().UTC()

		// Create spin exactly at start time
		s1 := createTestSpin(playerID, sessionID)
		s1.CreatedAt = now
		err := repo.Create(ctx, s1)
		require.NoError(t, err)

		// Create spin exactly at end time
		s2 := createTestSpin(playerID, sessionID)
		s2.CreatedAt = now.Add(1 * time.Hour)
		err = repo.Create(ctx, s2)
		require.NoError(t, err)

		count, err := repo.CountInTimeRange(ctx, playerID, now, now.Add(1*time.Hour))

		require.NoError(t, err)
		assert.Equal(t, int64(2), count) // Both should be included
	})
}
