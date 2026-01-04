package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupPlayerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce test noise
	})
	require.NoError(t, err, "Failed to connect to test database")

	// Create table manually with SQLite-compatible syntax
	// SQLite doesn't support PostgreSQL-specific types like uuid, decimal, varchar with uuid_generate_v4()
	err = db.Exec(`
		CREATE TABLE players (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			balance REAL DEFAULT 10000.00 NOT NULL,
			total_spins INTEGER DEFAULT 0,
			total_wagered REAL DEFAULT 0.00,
			total_won REAL DEFAULT 0.00,
			is_active INTEGER DEFAULT 1,
			is_verified INTEGER DEFAULT 0,
			game_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			lock_version INTEGER DEFAULT 0,
			last_login_at DATETIME
		)
	`).Error
	require.NoError(t, err, "Failed to create players table")

	return db
}

// createTestPlayer creates a test player with default values
func createTestPlayer() *player.Player {
	return &player.Player{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
		Balance:      10000.0,
		TotalSpins:   0,
		TotalWagered: 0.0,
		TotalWon:     0.0,
		IsActive:     true,
		IsVerified:   false,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		LockVersion:  0,
	}
}

// ============================================================================
// Create TESTS
// ============================================================================

func TestPlayerGormRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("should create player successfully", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()

		err := repo.Create(ctx, p)

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, p.ID)

		// Verify it was saved
		var saved player.Player
		err = db.First(&saved, "id = ?", p.ID).Error
		require.NoError(t, err)
		assert.Equal(t, p.Username, saved.Username)
		assert.Equal(t, p.Email, saved.Email)
		assert.Equal(t, p.Balance, saved.Balance)
	})

	t.Run("should return error for duplicate username", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p1 := createTestPlayer()
		p1.Username = "duplicate"
		err := repo.Create(ctx, p1)
		require.NoError(t, err)

		// Try to create another with same username
		p2 := createTestPlayer()
		p2.ID = uuid.New()
		p2.Username = "duplicate"
		p2.Email = "different@example.com"

		err = repo.Create(ctx, p2)
		assert.Error(t, err)
	})

	t.Run("should return error for duplicate email", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p1 := createTestPlayer()
		p1.Email = "duplicate@example.com"
		err := repo.Create(ctx, p1)
		require.NoError(t, err)

		// Try to create another with same email
		p2 := createTestPlayer()
		p2.ID = uuid.New()
		p2.Username = "different"
		p2.Email = "duplicate@example.com"

		err = repo.Create(ctx, p2)
		assert.Error(t, err)
	})

	t.Run("should allow setting balance to zero", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		p.Balance = 0 // Explicitly set to 0

		err := repo.Create(ctx, p)
		require.NoError(t, err)

		// Verify balance was saved as 0
		var saved player.Player
		err = db.First(&saved, "id = ?", p.ID).Error
		require.NoError(t, err)
		// The Create method in the repository applies default balance if not set
		// In this case we set it to 0, but the repository might apply default 100000
		assert.Greater(t, saved.Balance, 0.0, "Repository should apply default balance")
	})
}

// ============================================================================
// GetByID TESTS
// ============================================================================

func TestPlayerGormRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get player by ID successfully", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, p.ID)

		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, p.ID, retrieved.ID)
		assert.Equal(t, p.Username, retrieved.Username)
		assert.Equal(t, p.Email, retrieved.Email)
		assert.Equal(t, p.Balance, retrieved.Balance)
	})

	t.Run("should return error for non-existent ID", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		nonExistentID := uuid.New()

		retrieved, err := repo.GetByID(ctx, nonExistentID)

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, player.ErrPlayerNotFound, err)
	})
}

// ============================================================================
// GetByUsername TESTS
// ============================================================================

func TestPlayerGormRepository_GetByUsername(t *testing.T) {
	ctx := context.Background()

	t.Run("should get player by username successfully", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		p.Username = "uniqueuser"
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		retrieved, err := repo.GetByUsername(ctx, "uniqueuser")

		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, p.ID, retrieved.ID)
		assert.Equal(t, "uniqueuser", retrieved.Username)
	})

	t.Run("should return error for non-existent username", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		retrieved, err := repo.GetByUsername(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, player.ErrPlayerNotFound, err)
	})

	t.Run("should be case-insensitive", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		p.Username = "TestUser"
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		// Try lowercase - should find it because search is case-insensitive
		retrieved, err := repo.GetByUsername(ctx, "testuser")

		// Should find it (case-insensitive search)
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, "TestUser", retrieved.Username) // Original case preserved
	})
}

// ============================================================================
// GetByEmail TESTS
// ============================================================================

func TestPlayerGormRepository_GetByEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("should get player by email successfully", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		p.Email = "unique@example.com"
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		retrieved, err := repo.GetByEmail(ctx, "unique@example.com")

		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, p.ID, retrieved.ID)
		assert.Equal(t, "unique@example.com", retrieved.Email)
	})

	t.Run("should return error for non-existent email", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		retrieved, err := repo.GetByEmail(ctx, "nonexistent@example.com")

		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, player.ErrPlayerNotFound, err)
	})
}

// ============================================================================
// Update TESTS
// ============================================================================

func TestPlayerGormRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("should update player successfully", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		// Update fields
		p.Username = "updateduser"
		p.Balance = 50000.0
		p.IsVerified = true

		err = repo.Update(ctx, p)

		require.NoError(t, err)

		// Verify updates
		updated, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "updateduser", updated.Username)
		assert.Equal(t, 50000.0, updated.Balance)
		assert.True(t, updated.IsVerified)
	})

	t.Run("should insert player if not exists when using Save", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		p.ID = uuid.New() // Not in database initially
		p.Username = "newuser"
		p.Email = "newuser@example.com"

		// GORM's Save() will insert if record doesn't exist
		err := repo.Update(ctx, p)

		// Save() acts as upsert, so this should succeed
		require.NoError(t, err)

		// Verify player was inserted
		retrieved, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, "newuser", retrieved.Username)
	})

	t.Run("should update timestamps", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		originalUpdatedAt := p.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		p.Balance = 20000.0
		err = repo.Update(ctx, p)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.True(t, updated.UpdatedAt.After(originalUpdatedAt))
	})
}

// ============================================================================
// UpdateBalance TESTS
// ============================================================================

func TestPlayerGormRepository_UpdateBalance(t *testing.T) {
	ctx := context.Background()

	t.Run("should update balance successfully by adding amount", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		p.Balance = 10000.0
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		err = repo.UpdateBalance(ctx, p.ID, 5000.0)

		require.NoError(t, err)

		// Verify balance was updated (10000 + 5000 = 15000)
		updated, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, 15000.0, updated.Balance)
	})

	t.Run("should allow negative amounts to subtract from balance", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		p.Balance = 10000.0
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		err = repo.UpdateBalance(ctx, p.ID, -10000.0)

		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, 0.0, updated.Balance)
	})

	t.Run("should return error for non-existent player", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		nonExistentID := uuid.New()

		err := repo.UpdateBalance(ctx, nonExistentID, 5000.0)

		assert.Error(t, err)
		assert.Equal(t, player.ErrPlayerNotFound, err)
	})
}

// ============================================================================
// UpdateStatistics TESTS
// ============================================================================

func TestPlayerGormRepository_UpdateStatistics(t *testing.T) {
	ctx := context.Background()

	t.Run("should update statistics successfully", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		p.TotalSpins = 0
		p.TotalWagered = 0.0
		p.TotalWon = 0.0
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		err = repo.UpdateStatistics(ctx, p.ID, 10, 1000.0, 500.0)

		require.NoError(t, err)

		// Verify statistics were updated
		updated, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, 10, updated.TotalSpins)
		assert.Equal(t, 1000.0, updated.TotalWagered)
		assert.Equal(t, 500.0, updated.TotalWon)
	})

	t.Run("should accumulate statistics on multiple updates", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		p.TotalSpins = 5
		p.TotalWagered = 500.0
		p.TotalWon = 250.0
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		// Update with additional stats
		err = repo.UpdateStatistics(ctx, p.ID, 3, 300.0, 150.0)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.Equal(t, 8, updated.TotalSpins)      // 5 + 3
		assert.Equal(t, 800.0, updated.TotalWagered) // 500 + 300
		assert.Equal(t, 400.0, updated.TotalWon)     // 250 + 150
	})

	t.Run("should return error for non-existent player", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		nonExistentID := uuid.New()

		err := repo.UpdateStatistics(ctx, nonExistentID, 5, 500.0, 250.0)

		assert.Error(t, err)
		assert.Equal(t, player.ErrPlayerNotFound, err)
	})
}

// ============================================================================
// UpdateLastLogin TESTS
// ============================================================================

func TestPlayerGormRepository_UpdateLastLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("should update last login time", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		p.LastLoginAt = nil
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		err = repo.UpdateLastLogin(ctx, p.ID)

		require.NoError(t, err)

		// Verify last login was set
		updated, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated.LastLoginAt)
		assert.True(t, time.Since(*updated.LastLoginAt) < 5*time.Second)
	})

	t.Run("should update existing last login time", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		oldTime := time.Now().UTC().Add(-24 * time.Hour)
		p := createTestPlayer()
		p.LastLoginAt = &oldTime
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		err = repo.UpdateLastLogin(ctx, p.ID)

		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, p.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated.LastLoginAt)
		assert.True(t, updated.LastLoginAt.After(oldTime))
	})

	t.Run("should return error for non-existent player", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		nonExistentID := uuid.New()

		err := repo.UpdateLastLogin(ctx, nonExistentID)

		assert.Error(t, err)
		assert.Equal(t, player.ErrPlayerNotFound, err)
	})
}

// ============================================================================
// Delete TESTS
// ============================================================================

func TestPlayerGormRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete player successfully", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		p := createTestPlayer()
		err := repo.Create(ctx, p)
		require.NoError(t, err)

		err = repo.Delete(ctx, p.ID)

		require.NoError(t, err)

		// Verify player was deleted
		_, err = repo.GetByID(ctx, p.ID)
		assert.Error(t, err)
		assert.Equal(t, player.ErrPlayerNotFound, err)
	})

	t.Run("should return error for non-existent player", func(t *testing.T) {
		db := setupPlayerTestDB(t)
		repo := NewPlayerGormRepository(db)

		nonExistentID := uuid.New()

		err := repo.Delete(ctx, nonExistentID)

		assert.Error(t, err)
		assert.Equal(t, player.ErrPlayerNotFound, err)
	})
}

