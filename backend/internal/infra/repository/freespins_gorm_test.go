package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/freespins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupFreeSpinsTestDB creates an in-memory SQLite database for testing free spins
func setupFreeSpinsTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to test database")

	// Create table manually with SQLite-compatible syntax
	err = db.Exec(`
		CREATE TABLE free_spins_sessions (
			id TEXT PRIMARY KEY,
			player_id TEXT NOT NULL,
			session_id TEXT NOT NULL,
			triggered_by_spin_id TEXT,
			scatter_count INTEGER NOT NULL,
			total_spins_awarded INTEGER NOT NULL,
			spins_completed INTEGER DEFAULT 0,
			remaining_spins INTEGER NOT NULL,
			locked_bet_amount REAL NOT NULL,
			total_won REAL DEFAULT 0.00,
			is_active INTEGER DEFAULT 1,
			is_completed INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			lock_version INTEGER DEFAULT 0,
			completed_at DATETIME
		)
	`).Error
	require.NoError(t, err, "Failed to create free_spins_sessions table")

	// Create indices
	db.Exec("CREATE INDEX idx_free_spins_player_id ON free_spins_sessions(player_id)")
	db.Exec("CREATE INDEX idx_free_spins_is_active ON free_spins_sessions(is_active)")
	db.Exec("CREATE INDEX idx_free_spins_created_at ON free_spins_sessions(created_at)")

	return db
}

// createTestFreeSpinsSession creates a test free spins session with default values
func createTestFreeSpinsSession(playerID uuid.UUID) *freespins.FreeSpinsSession {
	spinID := uuid.New()
	return &freespins.FreeSpinsSession{
		ID:                uuid.New(),
		PlayerID:          playerID,
		SessionID:         uuid.New(),
		TriggeredBySpinID: &spinID,
		ScatterCount:      3,
		TotalSpinsAwarded: 10,
		SpinsCompleted:    0,
		RemainingSpins:    10,
		LockedBetAmount:   100.0,
		TotalWon:          0.0,
		IsActive:          true,
		IsCompleted:       false,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
		LockVersion:       0,
		CompletedAt:       nil,
	}
}

// ============================================================================
// Create TESTS
// ============================================================================

func TestFreeSpinsGormRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("should create free spins session successfully", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)

		err := repo.Create(ctx, session)

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, session.ID)

		// Verify it was saved
		var saved freespins.FreeSpinsSession
		err = db.First(&saved, "id = ?", session.ID).Error
		require.NoError(t, err)
		assert.Equal(t, session.PlayerID, saved.PlayerID)
		assert.Equal(t, session.ScatterCount, saved.ScatterCount)
		assert.Equal(t, session.TotalSpinsAwarded, saved.TotalSpinsAwarded)
		assert.True(t, saved.IsActive)
		assert.False(t, saved.IsCompleted)
	})

	t.Run("should set default values", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)

		err := repo.Create(ctx, session)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, retrieved.SpinsCompleted)
		assert.Equal(t, 0.0, retrieved.TotalWon)
		assert.True(t, retrieved.IsActive)
		assert.False(t, retrieved.IsCompleted)
	})
}

// ============================================================================
// GetByID TESTS
// ============================================================================

func TestFreeSpinsGormRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get free spins session by ID successfully", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)
		err := repo.Create(ctx, session)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, session.ID)

		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.PlayerID, retrieved.PlayerID)
		assert.Equal(t, session.TotalSpinsAwarded, retrieved.TotalSpinsAwarded)
	})

	t.Run("should return error for non-existent ID", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		nonExistentID := uuid.New()

		retrieved, err := repo.GetByID(ctx, nonExistentID)

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, freespins.ErrFreeSpinsNotFound, err)
	})
}

// ============================================================================
// GetActiveByPlayer TESTS
// ============================================================================

func TestFreeSpinsGormRepository_GetActiveByPlayer(t *testing.T) {
	ctx := context.Background()

	t.Run("should get active free spins session successfully", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)
		err := repo.Create(ctx, session)
		require.NoError(t, err)

		retrieved, err := repo.GetActiveByPlayer(ctx, playerID)

		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.True(t, retrieved.IsActive)
	})

	t.Run("should return error when no active session exists", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()

		retrieved, err := repo.GetActiveByPlayer(ctx, playerID)

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, freespins.ErrFreeSpinsNotFound, err)
	})

	t.Run("should not return completed sessions", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)
		err := repo.Create(ctx, session)
		require.NoError(t, err)

		// Complete the session
		err = repo.CompleteSession(ctx, session.ID)
		require.NoError(t, err)

		// Try to get active session
		retrieved, err := repo.GetActiveByPlayer(ctx, playerID)

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, freespins.ErrFreeSpinsNotFound, err)
	})

	t.Run("should return most recent active session", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()

		// Create first session and complete it
		session1 := createTestFreeSpinsSession(playerID)
		err := repo.Create(ctx, session1)
		require.NoError(t, err)
		err = repo.CompleteSession(ctx, session1.ID)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		// Create second session (active)
		session2 := createTestFreeSpinsSession(playerID)
		err = repo.Create(ctx, session2)
		require.NoError(t, err)

		retrieved, err := repo.GetActiveByPlayer(ctx, playerID)

		require.NoError(t, err)
		assert.Equal(t, session2.ID, retrieved.ID)
	})
}

// ============================================================================
// Update TESTS
// ============================================================================

func TestFreeSpinsGormRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("should update free spins session successfully", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)
		err := repo.Create(ctx, session)
		require.NoError(t, err)

		// Update fields
		session.SpinsCompleted = 5
		session.RemainingSpins = 5
		session.TotalWon = 1000.0

		err = repo.Update(ctx, session)

		require.NoError(t, err)

		// Verify updates
		updated, err := repo.GetByID(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, 5, updated.SpinsCompleted)
		assert.Equal(t, 5, updated.RemainingSpins)
		assert.Equal(t, 1000.0, updated.TotalWon)
	})
}

// ============================================================================
// UpdateSpins TESTS
// ============================================================================

func TestFreeSpinsGormRepository_UpdateSpins(t *testing.T) {
	ctx := context.Background()

	t.Run("should update spins successfully", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)
		session.TotalSpinsAwarded = 10
		session.SpinsCompleted = 0
		session.RemainingSpins = 10
		err := repo.Create(ctx, session)
		require.NoError(t, err)

		err = repo.UpdateSpins(ctx, session.ID, 3, 7)

		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, updated.SpinsCompleted)
		assert.Equal(t, 7, updated.RemainingSpins)
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		nonExistentID := uuid.New()

		err := repo.UpdateSpins(ctx, nonExistentID, 5, 5)

		assert.Error(t, err)
		assert.Equal(t, freespins.ErrFreeSpinsNotFound, err)
	})
}

// ============================================================================
// AddTotalWon TESTS
// ============================================================================

func TestFreeSpinsGormRepository_AddTotalWon(t *testing.T) {
	ctx := context.Background()

	t.Run("should add to total won successfully", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)
		session.TotalWon = 1000.0
		err := repo.Create(ctx, session)
		require.NoError(t, err)

		err = repo.AddTotalWon(ctx, session.ID, 2500.0)

		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, 3500.0, updated.TotalWon) // 1000.0 + 2500.0
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		nonExistentID := uuid.New()

		err := repo.AddTotalWon(ctx, nonExistentID, 1000.0)

		assert.Error(t, err)
		assert.Equal(t, freespins.ErrFreeSpinsNotFound, err)
	})
}

// ============================================================================
// CompleteSession TESTS
// ============================================================================

func TestFreeSpinsGormRepository_CompleteSession(t *testing.T) {
	ctx := context.Background()

	t.Run("should complete session successfully", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)
		err := repo.Create(ctx, session)
		require.NoError(t, err)

		err = repo.CompleteSession(ctx, session.ID)

		require.NoError(t, err)

		completed, err := repo.GetByID(ctx, session.ID)
		require.NoError(t, err)
		assert.False(t, completed.IsActive)
		assert.True(t, completed.IsCompleted)
		assert.NotNil(t, completed.CompletedAt)
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		nonExistentID := uuid.New()

		err := repo.CompleteSession(ctx, nonExistentID)

		assert.Error(t, err)
		assert.Equal(t, freespins.ErrFreeSpinsNotFound, err)
	})
}

// ============================================================================
// AddSpins TESTS
// ============================================================================

func TestFreeSpinsGormRepository_AddSpins(t *testing.T) {
	ctx := context.Background()

	t.Run("should add spins successfully (retrigger)", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)
		session.TotalSpinsAwarded = 10
		session.RemainingSpins = 5
		session.SpinsCompleted = 5
		err := repo.Create(ctx, session)
		require.NoError(t, err)

		// Add 5 more spins (retrigger)
		err = repo.AddSpins(ctx, session.ID, 5)

		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, 15, updated.TotalSpinsAwarded) // 10 + 5
		assert.Equal(t, 10, updated.RemainingSpins)    // 5 + 5
		assert.Equal(t, 5, updated.SpinsCompleted)     // Unchanged
	})

	t.Run("should handle multiple retriggers", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()
		session := createTestFreeSpinsSession(playerID)
		session.TotalSpinsAwarded = 10
		session.RemainingSpins = 3
		err := repo.Create(ctx, session)
		require.NoError(t, err)

		// First retrigger: add 5 spins
		err = repo.AddSpins(ctx, session.ID, 5)
		require.NoError(t, err)

		// Second retrigger: add 3 more spins
		err = repo.AddSpins(ctx, session.ID, 3)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, session.ID)
		require.NoError(t, err)
		assert.Equal(t, 18, updated.TotalSpinsAwarded) // 10 + 5 + 3
		assert.Equal(t, 11, updated.RemainingSpins)    // 3 + 5 + 3
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		nonExistentID := uuid.New()

		err := repo.AddSpins(ctx, nonExistentID, 5)

		assert.Error(t, err)
		assert.Equal(t, freespins.ErrFreeSpinsNotFound, err)
	})
}

// ============================================================================
// GetByPlayer TESTS
// ============================================================================

func TestFreeSpinsGormRepository_GetByPlayer(t *testing.T) {
	ctx := context.Background()

	t.Run("should get sessions by player successfully", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()

		// Create multiple sessions
		session1 := createTestFreeSpinsSession(playerID)
		err := repo.Create(ctx, session1)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		session2 := createTestFreeSpinsSession(playerID)
		err = repo.Create(ctx, session2)
		require.NoError(t, err)

		sessions, err := repo.GetByPlayer(ctx, playerID, 10, 0)

		require.NoError(t, err)
		assert.Len(t, sessions, 2)
		// Should be ordered by created_at DESC
		assert.Equal(t, session2.ID, sessions[0].ID)
		assert.Equal(t, session1.ID, sessions[1].ID)
	})

	t.Run("should paginate results", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()

		// Create 5 sessions
		for i := 0; i < 5; i++ {
			session := createTestFreeSpinsSession(playerID)
			err := repo.Create(ctx, session)
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

	t.Run("should return empty list when no sessions", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		playerID := uuid.New()

		sessions, err := repo.GetByPlayer(ctx, playerID, 10, 0)

		require.NoError(t, err)
		assert.NotNil(t, sessions)
		assert.Len(t, sessions, 0)
	})

	t.Run("should only return sessions for specified player", func(t *testing.T) {
		db := setupFreeSpinsTestDB(t)
		repo := NewFreeSpinsGormRepository(db)

		player1ID := uuid.New()
		player2ID := uuid.New()

		// Create session for player 1
		session1 := createTestFreeSpinsSession(player1ID)
		err := repo.Create(ctx, session1)
		require.NoError(t, err)

		// Create session for player 2
		session2 := createTestFreeSpinsSession(player2ID)
		err = repo.Create(ctx, session2)
		require.NoError(t, err)

		sessions, err := repo.GetByPlayer(ctx, player1ID, 10, 0)

		require.NoError(t, err)
		assert.Len(t, sessions, 1)
		assert.Equal(t, session1.ID, sessions[0].ID)
	})
}
