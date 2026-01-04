package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/freespins"
	"github.com/slotmachine/backend/domain/spin"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MOCKS
// ============================================================================

// MockFreeSpinsRepository is a mock implementation of freespins.Repository
type MockFreeSpinsRepository struct {
	mock.Mock
}

func (m *MockFreeSpinsRepository) Create(ctx context.Context, fs *freespins.FreeSpinsSession) error {
	args := m.Called(ctx, fs)
	return args.Error(0)
}

func (m *MockFreeSpinsRepository) GetByID(ctx context.Context, id uuid.UUID) (*freespins.FreeSpinsSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*freespins.FreeSpinsSession), args.Error(1)
}

func (m *MockFreeSpinsRepository) GetActiveByPlayer(ctx context.Context, playerID uuid.UUID) (*freespins.FreeSpinsSession, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*freespins.FreeSpinsSession), args.Error(1)
}

func (m *MockFreeSpinsRepository) Update(ctx context.Context, fs *freespins.FreeSpinsSession) error {
	args := m.Called(ctx, fs)
	return args.Error(0)
}

func (m *MockFreeSpinsRepository) UpdateSpins(ctx context.Context, id uuid.UUID, spinsCompleted, remainingSpins int) error {
	args := m.Called(ctx, id, spinsCompleted, remainingSpins)
	return args.Error(0)
}

func (m *MockFreeSpinsRepository) UpdateTotalWon(ctx context.Context, id uuid.UUID, totalWon float64) error {
	args := m.Called(ctx, id, totalWon)
	return args.Error(0)
}

func (m *MockFreeSpinsRepository) AddTotalWon(ctx context.Context, id uuid.UUID, totalWon float64) error {
	args := m.Called(ctx, id, totalWon)
	return args.Error(0)
}

func (m *MockFreeSpinsRepository) GetAvailableSessionByID(ctx context.Context, id uuid.UUID) (*freespins.FreeSpinsSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*freespins.FreeSpinsSession), args.Error(1)
}

func (m *MockFreeSpinsRepository) RollbackSpin(ctx context.Context, id uuid.UUID, additionalSpins int) error {
	args := m.Called(ctx, id, additionalSpins)
	return args.Error(0)
}

func (m *MockFreeSpinsRepository) ExecuteSpinWithLock(ctx context.Context, id uuid.UUID, additionalSpins int, lockVersion int) error {
	args := m.Called(ctx, id, additionalSpins, lockVersion)
	return args.Error(0)
}

func (m *MockFreeSpinsRepository) AddSpins(ctx context.Context, id uuid.UUID, additionalSpins int) error {
	args := m.Called(ctx, id, additionalSpins)
	return args.Error(0)
}

func (m *MockFreeSpinsRepository) CompleteSession(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFreeSpinsRepository) GetByPlayer(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*freespins.FreeSpinsSession, error) {
	args := m.Called(ctx, playerID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*freespins.FreeSpinsSession), args.Error(1)
}

// MockSpinRepository is a mock implementation of spin.Repository
type MockSpinRepository struct {
	mock.Mock
}

func (m *MockSpinRepository) Create(ctx context.Context, s *spin.Spin) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSpinRepository) GetByID(ctx context.Context, id uuid.UUID) (*spin.Spin, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*spin.Spin), args.Error(1)
}

func (m *MockSpinRepository) GetByPlayer(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*spin.Spin, error) {
	args := m.Called(ctx, playerID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*spin.Spin), args.Error(1)
}

func (m *MockSpinRepository) Count(ctx context.Context, playerID uuid.UUID) (int64, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSpinRepository) CountInTimeRange(ctx context.Context, playerID uuid.UUID, start, end time.Time) (int64, error) {
	args := m.Called(ctx, playerID, start, end)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSpinRepository) GetBySession(ctx context.Context, sessionID uuid.UUID) ([]*spin.Spin, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*spin.Spin), args.Error(1)
}

func (m *MockSpinRepository) GetByPlayerInTimeRange(ctx context.Context, playerID uuid.UUID, start, end time.Time, limit, offset int) ([]*spin.Spin, error) {
	args := m.Called(ctx, playerID, start, end, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*spin.Spin), args.Error(1)
}

func (m *MockSpinRepository) GetByFreeSpinsSession(ctx context.Context, freeSpinsSessionID uuid.UUID) ([]*spin.Spin, error) {
	args := m.Called(ctx, freeSpinsSessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*spin.Spin), args.Error(1)
}

func (m *MockSpinRepository) UpdateFreeSpinsSessionId(ctx context.Context, spinID, freeSpinsSessionID uuid.UUID) error {
	args := m.Called(ctx, spinID, freeSpinsSessionID)
	return args.Error(0)
}

// Note: For freespins tests, we don't need the actual GameEngine
// since we're testing the service layer logic, not the engine execution

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func setupFreeSpinsService() (*FreeSpinsService, *MockFreeSpinsRepository, *MockSpinRepository, *MockPlayerRepository, *MockSessionRepository) {
	mockSessionRepo := new(MockSessionRepository)
	mockFSRepo := new(MockFreeSpinsRepository)
	mockSpinRepo := new(MockSpinRepository)
	mockPlayerRepo := new(MockPlayerRepository)
	log := logger.New("info", "json")
	// Pass nil for game engine and pfService since we're only testing non-execution methods
	service := NewFreeSpinsService(mockSessionRepo, mockFSRepo, mockSpinRepo, mockPlayerRepo, nil, nil, log).(*FreeSpinsService)
	return service, mockFSRepo, mockSpinRepo, mockPlayerRepo, mockSessionRepo
}

// ============================================================================
// TriggerFreeSpins TESTS
// ============================================================================

func TestTriggerFreeSpins(t *testing.T) {
	ctx := context.Background()

	t.Run("should trigger free spins successfully with 3 scatters", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		playerID := uuid.New()
		spinID := uuid.New()
		scatterCount := 3
		betAmount := 100.0

		// Mock: no existing active session
		mockFSRepo.On("GetActiveByPlayer", ctx, playerID).Return(nil, freespins.ErrFreeSpinsNotFound)

		// Mock: create session
		mockFSRepo.On("Create", ctx, mock.AnythingOfType("*freespins.FreeSpinsSession")).Return(nil)

		// Execute
		session, err := service.TriggerFreeSpins(ctx, playerID, spinID, scatterCount, betAmount)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, playerID, session.PlayerID)
		assert.Equal(t, scatterCount, session.ScatterCount)
		assert.Equal(t, 12, session.TotalSpinsAwarded) // 3 scatters = 12 spins
		assert.Equal(t, 12, session.RemainingSpins)
		assert.Equal(t, 0, session.SpinsCompleted)
		assert.Equal(t, betAmount, session.LockedBetAmount)
		assert.True(t, session.IsActive)
		assert.False(t, session.IsCompleted)

		mockFSRepo.AssertExpectations(t)
	})

	t.Run("should trigger free spins with 4 scatters", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		playerID := uuid.New()
		spinID := uuid.New()
		scatterCount := 4
		betAmount := 100.0

		mockFSRepo.On("GetActiveByPlayer", ctx, playerID).Return(nil, freespins.ErrFreeSpinsNotFound)
		mockFSRepo.On("Create", ctx, mock.AnythingOfType("*freespins.FreeSpinsSession")).Return(nil)

		session, err := service.TriggerFreeSpins(ctx, playerID, spinID, scatterCount, betAmount)

		require.NoError(t, err)
		assert.Equal(t, 14, session.TotalSpinsAwarded) // 4 scatters = 14 spins

		mockFSRepo.AssertExpectations(t)
	})

	t.Run("should trigger free spins with 5 scatters", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		playerID := uuid.New()
		spinID := uuid.New()
		scatterCount := 5
		betAmount := 100.0

		mockFSRepo.On("GetActiveByPlayer", ctx, playerID).Return(nil, freespins.ErrFreeSpinsNotFound)
		mockFSRepo.On("Create", ctx, mock.AnythingOfType("*freespins.FreeSpinsSession")).Return(nil)

		session, err := service.TriggerFreeSpins(ctx, playerID, spinID, scatterCount, betAmount)

		require.NoError(t, err)
		assert.Equal(t, 16, session.TotalSpinsAwarded) // 5 scatters = 16 spins

		mockFSRepo.AssertExpectations(t)
	})

	t.Run("should return error for insufficient scatters", func(t *testing.T) {
		service, _, _, _, _ := setupFreeSpinsService()

		playerID := uuid.New()
		spinID := uuid.New()

		tests := []struct {
			name         string
			scatterCount int
		}{
			{"0 scatters", 0},
			{"1 scatter", 1},
			{"2 scatters", 2},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				session, err := service.TriggerFreeSpins(ctx, playerID, spinID, tt.scatterCount, 100.0)

				assert.Error(t, err)
				assert.Nil(t, session)
				assert.Contains(t, err.Error(), "insufficient scatters")
			})
		}
	})

	t.Run("should return error when active session exists", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		playerID := uuid.New()
		spinID := uuid.New()
		existingSession := &freespins.FreeSpinsSession{
			ID:            uuid.New(),
			PlayerID:      playerID,
			RemainingSpins: 5,
			IsActive:      true,
		}

		// Mock: existing active session
		mockFSRepo.On("GetActiveByPlayer", ctx, playerID).Return(existingSession, nil)

		// Execute
		session, err := service.TriggerFreeSpins(ctx, playerID, spinID, 3, 100.0)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Equal(t, freespins.ErrActiveFreeSpinsExists, err)

		mockFSRepo.AssertExpectations(t)
	})

	t.Run("should handle repository create error", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		playerID := uuid.New()
		spinID := uuid.New()
		repoErr := errors.New("database error")

		mockFSRepo.On("GetActiveByPlayer", ctx, playerID).Return(nil, freespins.ErrFreeSpinsNotFound)
		mockFSRepo.On("Create", ctx, mock.AnythingOfType("*freespins.FreeSpinsSession")).Return(repoErr)

		session, err := service.TriggerFreeSpins(ctx, playerID, spinID, 3, 100.0)

		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "failed to create free spins session")

		mockFSRepo.AssertExpectations(t)
	})
}

// ============================================================================
// GetStatus TESTS
// ============================================================================

func TestGetStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("should get status successfully", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		sessionID := uuid.New()
		mockSession := &freespins.FreeSpinsSession{
			ID:                sessionID,
			PlayerID:          uuid.New(),
			TotalSpinsAwarded: 12,
			SpinsCompleted:    5,
			RemainingSpins:    7,
			LockedBetAmount:   100.0,
			TotalWon:          500.0,
			IsActive:          true,
		}

		mockFSRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)

		status, err := service.GetStatus(ctx, sessionID)

		require.NoError(t, err)
		assert.NotNil(t, status)
		assert.True(t, status.Active)
		assert.Equal(t, sessionID, status.FreeSpinsSessionID)
		assert.Equal(t, 12, status.TotalSpinsAwarded)
		assert.Equal(t, 5, status.SpinsCompleted)
		assert.Equal(t, 7, status.RemainingSpins)
		assert.Equal(t, 100.0, status.LockedBetAmount)
		assert.Equal(t, 500.0, status.TotalWon)

		mockFSRepo.AssertExpectations(t)
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		sessionID := uuid.New()

		mockFSRepo.On("GetByID", ctx, sessionID).Return(nil, freespins.ErrFreeSpinsNotFound)

		status, err := service.GetStatus(ctx, sessionID)

		assert.Error(t, err)
		assert.Nil(t, status)
		assert.Equal(t, freespins.ErrFreeSpinsNotFound, err)

		mockFSRepo.AssertExpectations(t)
	})
}

// ============================================================================
// GetActiveSession TESTS
// ============================================================================

func TestGetActiveSession(t *testing.T) {
	ctx := context.Background()

	t.Run("should get active session successfully", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		playerID := uuid.New()
		mockSession := &freespins.FreeSpinsSession{
			ID:             uuid.New(),
			PlayerID:       playerID,
			RemainingSpins: 10,
			IsActive:       true,
		}

		mockFSRepo.On("GetActiveByPlayer", ctx, playerID).Return(mockSession, nil)

		session, err := service.GetActiveSession(ctx, playerID)

		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, playerID, session.PlayerID)
		assert.True(t, session.IsActive)

		mockFSRepo.AssertExpectations(t)
	})

	t.Run("should return error when no active session", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		playerID := uuid.New()

		mockFSRepo.On("GetActiveByPlayer", ctx, playerID).Return(nil, freespins.ErrFreeSpinsNotFound)

		session, err := service.GetActiveSession(ctx, playerID)

		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Equal(t, freespins.ErrFreeSpinsNotFound, err)

		mockFSRepo.AssertExpectations(t)
	})
}

// ============================================================================
// RetriggerFreeSpins TESTS
// ============================================================================

func TestRetriggerFreeSpins(t *testing.T) {
	ctx := context.Background()

	t.Run("should retrigger with 3 scatters", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		sessionID := uuid.New()
		mockSession := &freespins.FreeSpinsSession{
			ID:             sessionID,
			RemainingSpins: 5,
			IsActive:       true,
		}

		mockFSRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)
		mockFSRepo.On("AddSpins", ctx, sessionID, 12).Return(nil) // 3 scatters = 12 additional spins

		err := service.RetriggerFreeSpins(ctx, sessionID, 3)

		require.NoError(t, err)

		mockFSRepo.AssertExpectations(t)
	})

	t.Run("should retrigger with 4 scatters", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		sessionID := uuid.New()
		mockSession := &freespins.FreeSpinsSession{
			ID:       sessionID,
			IsActive: true,
		}

		mockFSRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)
		mockFSRepo.On("AddSpins", ctx, sessionID, 14).Return(nil) // 4 scatters = 14 spins

		err := service.RetriggerFreeSpins(ctx, sessionID, 4)

		require.NoError(t, err)

		mockFSRepo.AssertExpectations(t)
	})

	t.Run("should return error for insufficient scatters", func(t *testing.T) {
		service, _, _, _, _ := setupFreeSpinsService()

		sessionID := uuid.New()

		err := service.RetriggerFreeSpins(ctx, sessionID, 2)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient scatters")
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		sessionID := uuid.New()

		mockFSRepo.On("GetByID", ctx, sessionID).Return(nil, freespins.ErrFreeSpinsNotFound)

		err := service.RetriggerFreeSpins(ctx, sessionID, 3)

		assert.Error(t, err)
		assert.Equal(t, freespins.ErrFreeSpinsNotFound, err)

		mockFSRepo.AssertExpectations(t)
	})

	t.Run("should return error for inactive session", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		sessionID := uuid.New()
		completedTime := time.Now().UTC()
		mockSession := &freespins.FreeSpinsSession{
			ID:          sessionID,
			IsActive:    false,
			IsCompleted: true,
			CompletedAt: &completedTime,
		}

		mockFSRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)

		err := service.RetriggerFreeSpins(ctx, sessionID, 3)

		assert.Error(t, err)
		assert.Equal(t, freespins.ErrFreeSpinsNotActive, err)

		mockFSRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		service, mockFSRepo, _, _, _ := setupFreeSpinsService()

		sessionID := uuid.New()
		mockSession := &freespins.FreeSpinsSession{
			ID:       sessionID,
			IsActive: true,
		}
		repoErr := errors.New("database error")

		mockFSRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)
		mockFSRepo.On("AddSpins", ctx, sessionID, 12).Return(repoErr)

		err := service.RetriggerFreeSpins(ctx, sessionID, 3)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrigger free spins")

		mockFSRepo.AssertExpectations(t)
	})
}
