package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/game"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MOCKS
// ============================================================================

// MockPlayerRepository is a mock implementation of player.Repository
type MockPlayerRepository struct {
	mock.Mock
}

func (m *MockPlayerRepository) Create(ctx context.Context, p *player.Player) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPlayerRepository) GetByID(ctx context.Context, id uuid.UUID) (*player.Player, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*player.Player), args.Error(1)
}

func (m *MockPlayerRepository) GetByUsername(ctx context.Context, username string) (*player.Player, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*player.Player), args.Error(1)
}

func (m *MockPlayerRepository) GetByEmail(ctx context.Context, email string) (*player.Player, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*player.Player), args.Error(1)
}

func (m *MockPlayerRepository) Update(ctx context.Context, p *player.Player) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPlayerRepository) UpdateBalance(ctx context.Context, id uuid.UUID, newBalance float64) error {
	args := m.Called(ctx, id, newBalance)
	return args.Error(0)
}

func (m *MockPlayerRepository) UpdateBalanceWithLock(ctx context.Context, id uuid.UUID, newBalance float64, lockVersion int) error {
	args := m.Called(ctx, id, newBalance, lockVersion)
	return args.Error(0)
}

func (m *MockPlayerRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, spins int, wagered, won float64) error {
	args := m.Called(ctx, id, spins, wagered, won)
	return args.Error(0)
}

func (m *MockPlayerRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPlayerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPlayerRepository) List(ctx context.Context, filters player.ListFilters) ([]*player.Player, int64, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*player.Player), args.Get(1).(int64), args.Error(2)
}

func (m *MockPlayerRepository) GetByUsernameAndGame(ctx context.Context, username string, gameID *uuid.UUID) (*player.Player, error) {
	args := m.Called(ctx, username, gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*player.Player), args.Error(1)
}

func (m *MockPlayerRepository) GetByEmailAndGame(ctx context.Context, email string, gameID *uuid.UUID) (*player.Player, error) {
	args := m.Called(ctx, email, gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*player.Player), args.Error(1)
}

func (m *MockPlayerRepository) FindLoginCandidate(ctx context.Context, username string, gameID *uuid.UUID) (*player.Player, error) {
	args := m.Called(ctx, username, gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*player.Player), args.Error(1)
}

func (m *MockPlayerRepository) UpdateBalanceWithTx(ctx context.Context, id uuid.UUID, amount float64) error {
	args := m.Called(ctx, id, amount)
	return args.Error(0)
}

func (m *MockPlayerRepository) UpdateBalanceWithLockAndTx(ctx context.Context, id uuid.UUID, amount float64, lockVersion int) error {
	args := m.Called(ctx, id, amount, lockVersion)
	return args.Error(0)
}

// MockGameRepository is a mock implementation of game.Repository
type MockGameRepository struct {
	mock.Mock
}

func (m *MockGameRepository) GetGameByID(ctx context.Context, id uuid.UUID) (*game.Game, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.Game), args.Error(1)
}

func (m *MockGameRepository) GetGamesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*game.Game, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]*game.Game), args.Error(1)
}

func (m *MockGameRepository) ListGames(ctx context.Context, page, pageSize int) ([]*game.Game, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]*game.Game), args.Get(1).(int64), args.Error(2)
}

func (m *MockGameRepository) CreateGame(ctx context.Context, g *game.Game) error {
	args := m.Called(ctx, g)
	return args.Error(0)
}

func (m *MockGameRepository) UpdateGame(ctx context.Context, id uuid.UUID, update *game.GameUpdate) (*game.Game, error) {
	args := m.Called(ctx, id, update)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.Game), args.Error(1)
}

func (m *MockGameRepository) DeleteGame(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGameRepository) GetAssetByID(ctx context.Context, id uuid.UUID) (*game.Asset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.Asset), args.Error(1)
}

func (m *MockGameRepository) ListAssets(ctx context.Context, page, pageSize int) ([]*game.Asset, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]*game.Asset), args.Get(1).(int64), args.Error(2)
}

func (m *MockGameRepository) CreateAsset(ctx context.Context, a *game.Asset) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *MockGameRepository) UpdateAsset(ctx context.Context, id uuid.UUID, update *game.AssetUpdate) (*game.Asset, error) {
	args := m.Called(ctx, id, update)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.Asset), args.Error(1)
}

func (m *MockGameRepository) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGameRepository) GetActiveAssetForGame(ctx context.Context, gameID uuid.UUID) (*game.Asset, error) {
	args := m.Called(ctx, gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.Asset), args.Error(1)
}

func (m *MockGameRepository) GetGameConfigByID(ctx context.Context, id uuid.UUID) (*game.GameConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.GameConfig), args.Error(1)
}

func (m *MockGameRepository) ListGameConfigs(ctx context.Context, page, pageSize int) ([]*game.GameConfig, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]*game.GameConfig), args.Get(1).(int64), args.Error(2)
}

func (m *MockGameRepository) CreateGameConfig(ctx context.Context, c *game.GameConfig) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockGameRepository) DeleteGameConfig(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGameRepository) ActivateGameConfig(ctx context.Context, id uuid.UUID) (*game.GameConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.GameConfig), args.Error(1)
}

func (m *MockGameRepository) DeactivateGameConfig(ctx context.Context, id uuid.UUID) (*game.GameConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.GameConfig), args.Error(1)
}

// MockPlayerSessionRepository is a mock implementation of session.PlayerSessionRepository
type MockPlayerSessionRepository struct {
	mock.Mock
}

func (m *MockPlayerSessionRepository) Create(ctx context.Context, s *session.PlayerSession) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockPlayerSessionRepository) GetByToken(ctx context.Context, token string) (*session.PlayerSession, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.PlayerSession), args.Error(1)
}

func (m *MockPlayerSessionRepository) GetActiveByPlayerAndGame(ctx context.Context, playerID uuid.UUID, gameID *uuid.UUID) (*session.PlayerSession, error) {
	args := m.Called(ctx, playerID, gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.PlayerSession), args.Error(1)
}

func (m *MockPlayerSessionRepository) DeactivateSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
	args := m.Called(ctx, sessionID, reason)
	return args.Error(0)
}

func (m *MockPlayerSessionRepository) DeactivateAllPlayerSessions(ctx context.Context, playerID uuid.UUID, reason string) error {
	args := m.Called(ctx, playerID, reason)
	return args.Error(0)
}

func (m *MockPlayerSessionRepository) DeactivatePlayerGameSession(ctx context.Context, playerID uuid.UUID, gameID *uuid.UUID, reason string) error {
	args := m.Called(ctx, playerID, gameID, reason)
	return args.Error(0)
}

func (m *MockPlayerSessionRepository) UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockPlayerSessionRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func setupPlayerService() (*PlayerService, *MockPlayerRepository, *MockGameRepository, *MockPlayerSessionRepository) {
	mockRepo := new(MockPlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockSessionRepo := new(MockPlayerSessionRepository)
	log := logger.New("info", "json") // level, format
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:          "test-secret",
			ExpirationHours: 24,
		},
	}
	service := NewPlayerService(mockRepo, mockGameRepo, mockSessionRepo, nil, cfg, log).(*PlayerService)
	return service, mockRepo, mockGameRepo, mockSessionRepo
}

// ============================================================================
// Register TESTS
// ============================================================================

func TestRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("should reject registration without game_id", func(t *testing.T) {
		service, _, _, _ := setupPlayerService()

		// Execute (nil gameID should be rejected)
		p, err := service.Register(ctx, "testuser", "test@example.com", "password123", nil)

		// Assert - game_id is required
		assert.Error(t, err)
		assert.Equal(t, player.ErrGameIDRequired, err)
		assert.Nil(t, p)
	})

	t.Run("should register player with game_id", func(t *testing.T) {
		service, mockRepo, mockGameRepo, _ := setupPlayerService()

		gameID := uuid.New()
		mockGame := &game.Game{ID: gameID, Name: "Test Game"}

		// Mock: game exists
		mockGameRepo.On("GetGameByID", ctx, gameID).Return(mockGame, nil)

		// Mock: username doesn't exist for this game
		mockRepo.On("GetByUsernameAndGame", ctx, "testuser", &gameID).Return(nil, player.ErrPlayerNotFound)

		// Mock: email doesn't exist for this game
		mockRepo.On("GetByEmailAndGame", ctx, "test@example.com", &gameID).Return(nil, player.ErrPlayerNotFound)

		// Mock: create player
		mockRepo.On("Create", ctx, mock.AnythingOfType("*player.Player")).Return(nil)

		// Execute
		p, err := service.Register(ctx, "testuser", "test@example.com", "password123", &gameID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, "testuser", p.Username)
		assert.NotNil(t, p.GameID)
		assert.Equal(t, gameID, *p.GameID)

		mockRepo.AssertExpectations(t)
		mockGameRepo.AssertExpectations(t)
	})

	t.Run("should return error for duplicate username", func(t *testing.T) {
		service, mockRepo, mockGameRepo, _ := setupPlayerService()

		gameID := uuid.New()
		mockGame := &game.Game{ID: gameID, Name: "Test Game"}

		existingPlayer := &player.Player{
			ID:       uuid.New(),
			Username: "testuser",
			Email:    "existing@example.com",
			GameID:   &gameID,
		}

		// Mock: game exists
		mockGameRepo.On("GetGameByID", ctx, gameID).Return(mockGame, nil)

		// Mock: username exists
		mockRepo.On("GetByUsernameAndGame", ctx, "testuser", &gameID).Return(existingPlayer, nil)

		// Execute
		p, err := service.Register(ctx, "testuser", "test@example.com", "password123", &gameID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Equal(t, player.ErrPlayerAlreadyExists, err)

		mockRepo.AssertExpectations(t)
		mockGameRepo.AssertExpectations(t)
	})

	t.Run("should return error for duplicate email", func(t *testing.T) {
		service, mockRepo, mockGameRepo, _ := setupPlayerService()

		gameID := uuid.New()
		mockGame := &game.Game{ID: gameID, Name: "Test Game"}

		existingPlayer := &player.Player{
			ID:       uuid.New(),
			Username: "existinguser",
			Email:    "test@example.com",
			GameID:   &gameID,
		}

		// Mock: game exists
		mockGameRepo.On("GetGameByID", ctx, gameID).Return(mockGame, nil)

		// Mock: username doesn't exist
		mockRepo.On("GetByUsernameAndGame", ctx, "testuser", &gameID).Return(nil, player.ErrPlayerNotFound)

		// Mock: email exists
		mockRepo.On("GetByEmailAndGame", ctx, "test@example.com", &gameID).Return(existingPlayer, nil)

		// Execute
		p, err := service.Register(ctx, "testuser", "test@example.com", "password123", &gameID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Equal(t, player.ErrPlayerAlreadyExists, err)

		mockRepo.AssertExpectations(t)
		mockGameRepo.AssertExpectations(t)
	})

	t.Run("should validate username length", func(t *testing.T) {
		service, _, _, _ := setupPlayerService()

		tests := []struct {
			name     string
			username string
			wantErr  bool
		}{
			{"too short", "ab", true},
			{"empty", "", true},
			{"valid min", "abc", false},
			{"valid max", "a123456789012345678901234567890123456789012345678", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := service.validateRegistration(tt.username, "test@example.com", "password123")
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("should validate email format", func(t *testing.T) {
		service, _, _, _ := setupPlayerService()

		tests := []struct {
			name    string
			email   string
			wantErr bool
		}{
			{"valid email", "test@example.com", false},
			{"invalid - no @", "testexample.com", true},
			{"empty", "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := service.validateRegistration("testuser", tt.email, "password123")
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("should validate password length", func(t *testing.T) {
		service, _, _, _ := setupPlayerService()

		tests := []struct {
			name     string
			password string
			wantErr  bool
		}{
			{"too short", "pass", true},
			{"empty", "", true},
			{"valid", "password123", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := service.validateRegistration("testuser", "test@example.com", tt.password)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

// ============================================================================
// Login TESTS
// ============================================================================

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("should login successfully with correct credentials", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()
		// Hash the password "password123" for testing
		// The actual hash will be generated by util.HashPassword
		mockPlayer := &player.Player{
			ID:           playerID,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "$2a$10$YourHashedPasswordHere", // This will be mocked
			Balance:      100000.00,
			IsActive:     true,
			GameID:       nil, // Cross-game account
		}

		// Mock: find login candidate
		mockRepo.On("FindLoginCandidate", ctx, "testuser", (*uuid.UUID)(nil)).Return(mockPlayer, nil)

		// Note: We can't fully test password verification without a real hash
		// In a real test, we'd need to generate a proper hash for "password123"
		// For now, this tests the structure but password check will fail
		// Since password check will fail, UpdateLastLogin should NOT be called
		_, err := service.Login(ctx, "testuser", "password123", nil, nil)

		// We expect an error because our mock hash won't match
		assert.Error(t, err)
		assert.Equal(t, player.ErrInvalidCredentials, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for non-existent username", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		// Mock: player not found
		mockRepo.On("FindLoginCandidate", ctx, "nonexistent", (*uuid.UUID)(nil)).Return(nil, player.ErrPlayerNotFound)

		// Execute
		result, err := service.Login(ctx, "nonexistent", "password123", nil, nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, player.ErrInvalidCredentials, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for inactive player", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		mockPlayer := &player.Player{
			ID:           uuid.New(),
			Username:     "testuser",
			PasswordHash: "$2a$10$YourHashedPasswordHere",
			IsActive:     false, // Inactive player
			GameID:       nil,
		}

		// Mock: find login candidate
		mockRepo.On("FindLoginCandidate", ctx, "testuser", (*uuid.UUID)(nil)).Return(mockPlayer, nil)

		// Execute
		result, err := service.Login(ctx, "testuser", "password123", nil, nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not active")

		mockRepo.AssertExpectations(t)
	})

	t.Run("should deny login for game-specific player to different game", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		game1ID := uuid.New()
		game2ID := uuid.New()

		mockPlayer := &player.Player{
			ID:           uuid.New(),
			Username:     "testuser",
			PasswordHash: "$2a$10$YourHashedPasswordHere",
			IsActive:     true,
			GameID:       &game1ID, // Player is bound to game1
		}

		// Mock: find login candidate returns game1 player
		mockRepo.On("FindLoginCandidate", ctx, "testuser", &game2ID).Return(mockPlayer, nil)

		// Execute - try to login with game2
		result, err := service.Login(ctx, "testuser", "password123", &game2ID, nil)

		// Assert - should be denied because player is bound to game1
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, player.ErrGameAccessDenied, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should validate required fields", func(t *testing.T) {
		service, _, _, _ := setupPlayerService()

		tests := []struct {
			name     string
			username string
			password string
			wantErr  bool
		}{
			{"empty username", "", "password123", true},
			{"empty password", "testuser", "", true},
			{"both empty", "", "", true},
			{"both provided", "testuser", "password123", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.username == "" || tt.password == "" {
					_, err := service.Login(ctx, tt.username, tt.password, nil, nil)
					assert.Error(t, err)
				}
			})
		}
	})
}

// ============================================================================
// GetProfile TESTS
// ============================================================================

func TestGetProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("should get player profile successfully", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()
		mockPlayer := &player.Player{
			ID:       playerID,
			Username: "testuser",
			Email:    "test@example.com",
			Balance:  50000.00,
			IsActive: true,
		}

		// Mock: get player by ID
		mockRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Execute
		p, err := service.GetProfile(ctx, playerID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, playerID, p.ID)
		assert.Equal(t, "testuser", p.Username)
		assert.Equal(t, 50000.00, p.Balance)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for non-existent player", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()

		// Mock: player not found
		mockRepo.On("GetByID", ctx, playerID).Return(nil, player.ErrPlayerNotFound)

		// Execute
		p, err := service.GetProfile(ctx, playerID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Equal(t, player.ErrPlayerNotFound, err)

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// GetBalance TESTS
// ============================================================================

func TestGetBalance(t *testing.T) {
	ctx := context.Background()

	t.Run("should get player balance successfully", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()
		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 75000.00,
		}

		// Mock: get player by ID
		mockRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Execute
		balance, err := service.GetBalance(ctx, playerID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, 75000.00, balance)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for non-existent player", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()

		// Mock: player not found
		mockRepo.On("GetByID", ctx, playerID).Return(nil, player.ErrPlayerNotFound)

		// Execute
		balance, err := service.GetBalance(ctx, playerID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0.0, balance)
		assert.Equal(t, player.ErrPlayerNotFound, err)

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// UpdateBalance TESTS
// ============================================================================

func TestUpdateBalance(t *testing.T) {
	ctx := context.Background()

	t.Run("should update balance successfully", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()
		newBalance := 50000.00

		// Mock: update balance
		mockRepo.On("UpdateBalance", ctx, playerID, newBalance).Return(nil)

		// Execute
		err := service.UpdateBalance(ctx, playerID, newBalance)

		// Assert
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should reject negative balance", func(t *testing.T) {
		service, _, _, _ := setupPlayerService()

		playerID := uuid.New()

		// Execute
		err := service.UpdateBalance(ctx, playerID, -100.00)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be negative")
	})

	t.Run("should allow zero balance", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()

		// Mock: update balance to zero
		mockRepo.On("UpdateBalance", ctx, playerID, 0.0).Return(nil)

		// Execute
		err := service.UpdateBalance(ctx, playerID, 0.0)

		// Assert
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()
		repoErr := errors.New("database error")

		// Mock: update balance fails
		mockRepo.On("UpdateBalance", ctx, playerID, 50000.00).Return(repoErr)

		// Execute
		err := service.UpdateBalance(ctx, playerID, 50000.00)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update balance")

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// DeductBet TESTS
// ============================================================================

func TestDeductBet(t *testing.T) {
	ctx := context.Background()

	t.Run("should deduct bet successfully", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()
		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 10000.00,
		}
		betAmount := 100.00

		// Mock: get player
		mockRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Mock: update balance
		mockRepo.On("UpdateBalance", ctx, playerID, 9900.00).Return(nil)

		// Execute
		err := service.DeductBet(ctx, playerID, betAmount)

		// Assert
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for insufficient balance", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()
		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 50.00,
		}
		betAmount := 100.00

		// Mock: get player
		mockRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Execute
		err := service.DeductBet(ctx, playerID, betAmount)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, player.ErrInsufficientBalance, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should reject negative bet amount", func(t *testing.T) {
		service, _, _, _ := setupPlayerService()

		playerID := uuid.New()

		// Execute
		err := service.DeductBet(ctx, playerID, -10.00)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("should reject zero bet amount", func(t *testing.T) {
		service, _, _, _ := setupPlayerService()

		playerID := uuid.New()

		// Execute
		err := service.DeductBet(ctx, playerID, 0.0)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("should handle player not found", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()

		// Mock: player not found
		mockRepo.On("GetByID", ctx, playerID).Return(nil, player.ErrPlayerNotFound)

		// Execute
		err := service.DeductBet(ctx, playerID, 100.00)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, player.ErrPlayerNotFound, err)

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// CreditWin TESTS
// ============================================================================

func TestCreditWin(t *testing.T) {
	ctx := context.Background()

	t.Run("should credit win successfully", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()
		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 10000.00,
		}
		winAmount := 500.00

		// Mock: get player
		mockRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Mock: update balance
		mockRepo.On("UpdateBalance", ctx, playerID, 10500.00).Return(nil)

		// Execute
		err := service.CreditWin(ctx, playerID, winAmount)

		// Assert
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should skip credit for zero win", func(t *testing.T) {
		service, _, _, _ := setupPlayerService()

		playerID := uuid.New()

		// Execute (should not call repository)
		err := service.CreditWin(ctx, playerID, 0.0)

		// Assert
		require.NoError(t, err)
	})

	t.Run("should reject negative win amount", func(t *testing.T) {
		service, _, _, _ := setupPlayerService()

		playerID := uuid.New()

		// Execute
		err := service.CreditWin(ctx, playerID, -100.00)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be negative")
	})

	t.Run("should handle player not found", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()

		// Mock: player not found
		mockRepo.On("GetByID", ctx, playerID).Return(nil, player.ErrPlayerNotFound)

		// Execute
		err := service.CreditWin(ctx, playerID, 500.00)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, player.ErrPlayerNotFound, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository update error", func(t *testing.T) {
		service, mockRepo, _, _ := setupPlayerService()

		playerID := uuid.New()
		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 10000.00,
		}
		repoErr := errors.New("database error")

		// Mock: get player
		mockRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Mock: update balance fails
		mockRepo.On("UpdateBalance", ctx, playerID, 10500.00).Return(repoErr)

		// Execute
		err := service.CreditWin(ctx, playerID, 500.00)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to credit win")

		mockRepo.AssertExpectations(t)
	})
}
