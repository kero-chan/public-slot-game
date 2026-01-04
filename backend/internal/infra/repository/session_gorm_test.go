package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupSessionTestDB creates an in-memory SQLite database for testing sessions
func setupSessionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to test database")

	// Create table manually with SQLite-compatible syntax
	err = db.Exec(`
		CREATE TABLE game_sessions (
			id TEXT PRIMARY KEY,
			player_id TEXT NOT NULL,
			bet_amount REAL NOT NULL,
			starting_balance REAL NOT NULL,
			ending_balance REAL,
			total_spins INTEGER DEFAULT 0,
			total_wagered REAL DEFAULT 0.00,
			total_won REAL DEFAULT 0.00,
			net_change REAL DEFAULT 0.00,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			ended_at DATETIME
		)
	`).Error
	require.NoError(t, err, "Failed to create game_sessions table")

	// Create index on player_id for better query performance
	err = db.Exec("CREATE INDEX idx_sessions_player_id ON game_sessions(player_id)").Error
	require.NoError(t, err, "Failed to create player_id index")

	// Create index on ended_at for active session queries
	err = db.Exec("CREATE INDEX idx_sessions_ended_at ON game_sessions(ended_at)").Error
	require.NoError(t, err, "Failed to create ended_at index")

	return db
}

// createTestSession creates a test session with default values
func createTestSession(playerID uuid.UUID) *session.GameSession {
	return &session.GameSession{
		ID:              uuid.New(),
		PlayerID:        playerID,
		BetAmount:       100.0,
		StartingBalance: 10000.0,
		EndingBalance:   nil,
		TotalSpins:      0,
		TotalWagered:    0.0,
		TotalWon:        0.0,
		NetChange:       0.0,
		CreatedAt:       time.Now().UTC(),
		EndedAt:         nil,
	}
}

// ============================================================================
// Create TESTS
// ============================================================================

func TestSessionGormRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("should create session successfully", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)

		err := repo.Create(ctx, s)

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, s.ID)

		// Verify it was saved
		var saved session.GameSession
		err = db.First(&saved, "id = ?", s.ID).Error
		require.NoError(t, err)
		assert.Equal(t, s.PlayerID, saved.PlayerID)
		assert.Equal(t, s.BetAmount, saved.BetAmount)
		assert.Equal(t, s.StartingBalance, saved.StartingBalance)
		assert.Nil(t, saved.EndedAt)
	})

	t.Run("should set default values for statistics", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)

		err := repo.Create(ctx, s)
		require.NoError(t, err)

		// Verify default statistics
		retrieved, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, retrieved.TotalSpins)
		assert.Equal(t, 0.0, retrieved.TotalWagered)
		assert.Equal(t, 0.0, retrieved.TotalWon)
		assert.Equal(t, 0.0, retrieved.NetChange)
	})
}

// ============================================================================
// GetByID TESTS
// ============================================================================

func TestSessionGormRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get session by ID successfully", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, s.ID)

		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, s.ID, retrieved.ID)
		assert.Equal(t, s.PlayerID, retrieved.PlayerID)
		assert.Equal(t, s.BetAmount, retrieved.BetAmount)
		assert.Equal(t, s.StartingBalance, retrieved.StartingBalance)
	})

	t.Run("should return error for non-existent ID", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		nonExistentID := uuid.New()

		retrieved, err := repo.GetByID(ctx, nonExistentID)

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, session.ErrSessionNotFound, err)
	})
}

// ============================================================================
// GetActiveSessionByPlayer TESTS
// ============================================================================

func TestSessionGormRepository_GetActiveSessionByPlayer(t *testing.T) {
	ctx := context.Background()

	t.Run("should get active session successfully", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		retrieved, err := repo.GetActiveSessionByPlayer(ctx, playerID)

		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, s.ID, retrieved.ID)
		assert.Equal(t, playerID, retrieved.PlayerID)
		assert.Nil(t, retrieved.EndedAt)
	})

	t.Run("should return error when no active session exists", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()

		retrieved, err := repo.GetActiveSessionByPlayer(ctx, playerID)

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, session.ErrSessionNotFound, err)
	})

	t.Run("should not return ended sessions", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		// End the session
		err = repo.EndSession(ctx, s.ID, 9500.0)
		require.NoError(t, err)

		// Try to get active session
		retrieved, err := repo.GetActiveSessionByPlayer(ctx, playerID)

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, session.ErrSessionNotFound, err)
	})

	t.Run("should return most recent active session", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()

		// Create first session and end it
		s1 := createTestSession(playerID)
		err := repo.Create(ctx, s1)
		require.NoError(t, err)
		err = repo.EndSession(ctx, s1.ID, 9500.0)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		// Create second session (active)
		s2 := createTestSession(playerID)
		err = repo.Create(ctx, s2)
		require.NoError(t, err)

		retrieved, err := repo.GetActiveSessionByPlayer(ctx, playerID)

		require.NoError(t, err)
		assert.Equal(t, s2.ID, retrieved.ID)
	})
}

// ============================================================================
// Update TESTS
// ============================================================================

func TestSessionGormRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("should update session successfully", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		// Update fields
		s.TotalSpins = 10
		s.TotalWagered = 1000.0
		s.TotalWon = 500.0

		err = repo.Update(ctx, s)

		require.NoError(t, err)

		// Verify updates
		updated, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.Equal(t, 10, updated.TotalSpins)
		assert.Equal(t, 1000.0, updated.TotalWagered)
		assert.Equal(t, 500.0, updated.TotalWon)
	})

	t.Run("should insert session if not exists when using Save", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)
		s.ID = uuid.New() // Not in database

		// GORM's Save() will insert if record doesn't exist
		err := repo.Update(ctx, s)

		// Save() acts as upsert, so this should succeed
		require.NoError(t, err)

		// Verify session was inserted
		retrieved, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.Equal(t, playerID, retrieved.PlayerID)
	})
}

// ============================================================================
// EndSession TESTS
// ============================================================================

func TestSessionGormRepository_EndSession(t *testing.T) {
	ctx := context.Background()

	t.Run("should end session successfully", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)
		s.StartingBalance = 10000.0
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		endingBalance := 9500.0
		err = repo.EndSession(ctx, s.ID, endingBalance)

		require.NoError(t, err)

		// Verify session was ended
		ended, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.NotNil(t, ended.EndedAt)
		assert.NotNil(t, ended.EndingBalance)
		assert.Equal(t, endingBalance, *ended.EndingBalance)
		assert.Equal(t, -500.0, ended.NetChange) // 9500 - 10000
	})

	t.Run("should calculate positive net change", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)
		s.StartingBalance = 10000.0
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		endingBalance := 12000.0
		err = repo.EndSession(ctx, s.ID, endingBalance)

		require.NoError(t, err)

		ended, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.Equal(t, 2000.0, ended.NetChange) // 12000 - 10000
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		nonExistentID := uuid.New()

		err := repo.EndSession(ctx, nonExistentID, 5000.0)

		assert.Error(t, err)
		assert.Equal(t, session.ErrSessionNotFound, err)
	})
}

// ============================================================================
// UpdateStatistics TESTS
// ============================================================================

func TestSessionGormRepository_UpdateStatistics(t *testing.T) {
	ctx := context.Background()

	t.Run("should update statistics successfully", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)
		s.TotalSpins = 0
		s.TotalWagered = 0.0
		s.TotalWon = 0.0
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		err = repo.UpdateStatistics(ctx, s.ID, 5, 500.0, 250.0)

		require.NoError(t, err)

		// Verify statistics were updated
		updated, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.Equal(t, 5, updated.TotalSpins)
		assert.Equal(t, 500.0, updated.TotalWagered)
		assert.Equal(t, 250.0, updated.TotalWon)
	})

	t.Run("should accumulate statistics on multiple updates", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()
		s := createTestSession(playerID)
		s.TotalSpins = 3
		s.TotalWagered = 300.0
		s.TotalWon = 150.0
		err := repo.Create(ctx, s)
		require.NoError(t, err)

		// First update
		err = repo.UpdateStatistics(ctx, s.ID, 2, 200.0, 100.0)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.Equal(t, 5, updated.TotalSpins)     // 3 + 2
		assert.Equal(t, 500.0, updated.TotalWagered) // 300 + 200
		assert.Equal(t, 250.0, updated.TotalWon)     // 150 + 100

		// Second update
		err = repo.UpdateStatistics(ctx, s.ID, 3, 300.0, 150.0)
		require.NoError(t, err)

		updated, err = repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.Equal(t, 8, updated.TotalSpins)     // 5 + 3
		assert.Equal(t, 800.0, updated.TotalWagered) // 500 + 300
		assert.Equal(t, 400.0, updated.TotalWon)     // 250 + 150
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		nonExistentID := uuid.New()

		err := repo.UpdateStatistics(ctx, nonExistentID, 5, 500.0, 250.0)

		assert.Error(t, err)
		assert.Equal(t, session.ErrSessionNotFound, err)
	})
}

// ============================================================================
// GetByPlayer TESTS
// ============================================================================

func TestSessionGormRepository_GetByPlayer(t *testing.T) {
	ctx := context.Background()

	t.Run("should get sessions by player successfully", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()

		// Create multiple sessions
		s1 := createTestSession(playerID)
		err := repo.Create(ctx, s1)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		s2 := createTestSession(playerID)
		err = repo.Create(ctx, s2)
		require.NoError(t, err)

		sessions, err := repo.GetByPlayer(ctx, playerID, 10, 0)

		require.NoError(t, err)
		assert.Len(t, sessions, 2)
		// Should be ordered by created_at DESC
		assert.Equal(t, s2.ID, sessions[0].ID)
		assert.Equal(t, s1.ID, sessions[1].ID)
	})

	t.Run("should paginate results", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()

		// Create 5 sessions
		for i := 0; i < 5; i++ {
			s := createTestSession(playerID)
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

		// Get third page
		page3, err := repo.GetByPlayer(ctx, playerID, 2, 4)
		require.NoError(t, err)
		assert.Len(t, page3, 1)

		// Verify no overlap
		assert.NotEqual(t, page1[0].ID, page2[0].ID)
		assert.NotEqual(t, page1[0].ID, page3[0].ID)
		assert.NotEqual(t, page2[0].ID, page3[0].ID)
	})

	t.Run("should return empty list when no sessions", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		playerID := uuid.New()

		sessions, err := repo.GetByPlayer(ctx, playerID, 10, 0)

		require.NoError(t, err)
		assert.NotNil(t, sessions)
		assert.Len(t, sessions, 0)
	})

	t.Run("should only return sessions for specified player", func(t *testing.T) {
		db := setupSessionTestDB(t)
		repo := NewSessionGormRepository(db)

		player1ID := uuid.New()
		player2ID := uuid.New()

		// Create sessions for player 1
		s1 := createTestSession(player1ID)
		err := repo.Create(ctx, s1)
		require.NoError(t, err)

		// Create sessions for player 2
		s2 := createTestSession(player2ID)
		err = repo.Create(ctx, s2)
		require.NoError(t, err)

		// Get sessions for player 1
		sessions, err := repo.GetByPlayer(ctx, player1ID, 10, 0)

		require.NoError(t, err)
		assert.Len(t, sessions, 1)
		assert.Equal(t, s1.ID, sessions[0].ID)
		assert.Equal(t, player1ID, sessions[0].PlayerID)
	})
}
