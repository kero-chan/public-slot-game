package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/domain/spin"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func setupSpinServiceForValidation() (*SpinService, *MockSpinRepository, *MockPlayerRepository, *MockSessionRepository, *MockFreeSpinsRepository) {
	mockSpinRepo := new(MockSpinRepository)
	mockPlayerRepo := new(MockPlayerRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockFreeSpinsRepo := new(MockFreeSpinsRepository)
	log := logger.New("info", "json")

	// Create service without game engine for validation tests (pass nil for txManager, reelstripRepo, and pfService in tests)
	service := NewSpinService(mockSpinRepo, mockPlayerRepo, mockSessionRepo, nil, mockFreeSpinsRepo, nil, nil, nil, log).(*SpinService)

	return service, mockSpinRepo, mockPlayerRepo, mockSessionRepo, mockFreeSpinsRepo
}

// ============================================================================
// ExecuteSpin VALIDATION TESTS
// ============================================================================

func TestExecuteSpinValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("should return error for invalid bet amount - zero", func(t *testing.T) {
		service, _, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		sessionID := uuid.New()

		result, err := service.ExecuteSpin(ctx, playerID, sessionID, 0, "", "", "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "bet amount must be positive")
	})

	t.Run("should return error for invalid bet amount - negative", func(t *testing.T) {
		service, _, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		sessionID := uuid.New()

		result, err := service.ExecuteSpin(ctx, playerID, sessionID, -100, "", "", "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "bet amount must be positive")
	})

	t.Run("should return error when player not found", func(t *testing.T) {
		service, _, mockPlayerRepo, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		sessionID := uuid.New()

		mockPlayerRepo.On("GetByID", ctx, playerID).Return(nil, player.ErrPlayerNotFound)

		result, err := service.ExecuteSpin(ctx, playerID, sessionID, 100.0, "", "", "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, player.ErrPlayerNotFound, err)

		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("should return error when session not found", func(t *testing.T) {
		service, _, mockPlayerRepo, mockSessionRepo, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		sessionID := uuid.New()

		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 10000.0,
		}

		mockPlayerRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(nil, session.ErrSessionNotFound)

		result, err := service.ExecuteSpin(ctx, playerID, sessionID, 100.0, "", "", "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, session.ErrSessionNotFound, err)

		mockPlayerRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("should return error when session belongs to different player", func(t *testing.T) {
		service, _, mockPlayerRepo, mockSessionRepo, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		sessionID := uuid.New()
		differentPlayerID := uuid.New()

		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 10000.0,
		}

		mockSession := &session.GameSession{
			ID:       sessionID,
			PlayerID: differentPlayerID, // Different player
			EndedAt:  nil,
		}

		mockPlayerRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)

		result, err := service.ExecuteSpin(ctx, playerID, sessionID, 100.0, "", "", "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "session does not belong to player")

		mockPlayerRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("should return error when session already ended", func(t *testing.T) {
		service, _, mockPlayerRepo, mockSessionRepo, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		sessionID := uuid.New()
		endedAt := time.Now().UTC()

		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 10000.0,
		}

		mockSession := &session.GameSession{
			ID:       sessionID,
			PlayerID: playerID,
			EndedAt:  &endedAt, // Session ended
		}

		mockPlayerRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)

		result, err := service.ExecuteSpin(ctx, playerID, sessionID, 100.0, "", "", "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, session.ErrSessionAlreadyEnded, err)

		mockPlayerRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("should return error for insufficient balance", func(t *testing.T) {
		service, _, mockPlayerRepo, mockSessionRepo, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		sessionID := uuid.New()
		betAmount := 1000.0

		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 100.0, // Less than bet amount
		}

		mockSession := &session.GameSession{
			ID:       sessionID,
			PlayerID: playerID,
			EndedAt:  nil,
		}

		mockPlayerRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)

		result, err := service.ExecuteSpin(ctx, playerID, sessionID, betAmount, "", "", "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, player.ErrInsufficientBalance, err)

		mockPlayerRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})
}

// ============================================================================
// GetSpinDetails TESTS
// ============================================================================

func TestGetSpinDetails(t *testing.T) {
	ctx := context.Background()

	t.Run("should get spin details successfully", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		spinID := uuid.New()
		mockSpin := &spin.Spin{
			ID:         spinID,
			PlayerID:   uuid.New(),
			BetAmount:  100.0,
			TotalWin:   500.0,
			IsFreeSpin: false,
		}

		mockSpinRepo.On("GetByID", ctx, spinID).Return(mockSpin, nil)

		result, err := service.GetSpinDetails(ctx, spinID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, spinID, result.ID)
		assert.Equal(t, 100.0, result.BetAmount)
		assert.Equal(t, 500.0, result.TotalWin)

		mockSpinRepo.AssertExpectations(t)
	})

	t.Run("should return error when spin not found", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		spinID := uuid.New()

		mockSpinRepo.On("GetByID", ctx, spinID).Return(nil, errors.New("not found"))

		result, err := service.GetSpinDetails(ctx, spinID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, spin.ErrSpinNotFound, err)

		mockSpinRepo.AssertExpectations(t)
	})
}

// ============================================================================
// GetSpinHistory TESTS
// ============================================================================

func TestGetSpinHistory(t *testing.T) {
	ctx := context.Background()

	t.Run("should get spin history successfully", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		page := 1
		limit := 20

		mockSpins := []*spin.Spin{
			{ID: uuid.New(), PlayerID: playerID, BetAmount: 100.0, TotalWin: 200.0},
			{ID: uuid.New(), PlayerID: playerID, BetAmount: 100.0, TotalWin: 0.0},
		}

		mockSpinRepo.On("Count", ctx, playerID).Return(int64(50), nil)
		mockSpinRepo.On("GetByPlayer", ctx, playerID, limit, 0).Return(mockSpins, nil)

		result, err := service.GetSpinHistory(ctx, playerID, page, limit)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, page, result.Page)
		assert.Equal(t, limit, result.Limit)
		assert.Equal(t, int64(50), result.Total)
		assert.Len(t, result.Spins, 2)

		mockSpinRepo.AssertExpectations(t)
	})

	t.Run("should use default pagination values for invalid page", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()

		mockSpinRepo.On("Count", ctx, playerID).Return(int64(10), nil)
		mockSpinRepo.On("GetByPlayer", ctx, playerID, 20, 0).Return([]*spin.Spin{}, nil)

		result, err := service.GetSpinHistory(ctx, playerID, 0, 0) // Invalid params

		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)   // Default to page 1
		assert.Equal(t, 20, result.Limit) // Default limit

		mockSpinRepo.AssertExpectations(t)
	})

	t.Run("should use default pagination values for invalid limit", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()

		mockSpinRepo.On("Count", ctx, playerID).Return(int64(10), nil)
		mockSpinRepo.On("GetByPlayer", ctx, playerID, 20, 0).Return([]*spin.Spin{}, nil)

		result, err := service.GetSpinHistory(ctx, playerID, 1, -5) // Invalid limit

		require.NoError(t, err)
		assert.Equal(t, 20, result.Limit) // Default limit

		mockSpinRepo.AssertExpectations(t)
	})

	t.Run("should enforce max limit", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()

		mockSpinRepo.On("Count", ctx, playerID).Return(int64(10), nil)
		mockSpinRepo.On("GetByPlayer", ctx, playerID, 20, 0).Return([]*spin.Spin{}, nil)

		result, err := service.GetSpinHistory(ctx, playerID, 1, 200) // Over max

		require.NoError(t, err)
		assert.Equal(t, 20, result.Limit) // Capped to default

		mockSpinRepo.AssertExpectations(t)
	})

	t.Run("should calculate correct offset for page 2", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		page := 2
		limit := 10

		mockSpinRepo.On("Count", ctx, playerID).Return(int64(50), nil)
		mockSpinRepo.On("GetByPlayer", ctx, playerID, limit, 10).Return([]*spin.Spin{}, nil) // offset = (2-1) * 10 = 10

		result, err := service.GetSpinHistory(ctx, playerID, page, limit)

		require.NoError(t, err)
		assert.Equal(t, 2, result.Page)

		mockSpinRepo.AssertExpectations(t)
	})

	t.Run("should calculate correct offset for page 3", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		page := 3
		limit := 15

		mockSpinRepo.On("Count", ctx, playerID).Return(int64(50), nil)
		mockSpinRepo.On("GetByPlayer", ctx, playerID, limit, 30).Return([]*spin.Spin{}, nil) // offset = (3-1) * 15 = 30

		result, err := service.GetSpinHistory(ctx, playerID, page, limit)

		require.NoError(t, err)
		assert.Equal(t, 3, result.Page)

		mockSpinRepo.AssertExpectations(t)
	})

	t.Run("should handle count error", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		countErr := errors.New("database error")

		mockSpinRepo.On("Count", ctx, playerID).Return(int64(0), countErr)

		result, err := service.GetSpinHistory(ctx, playerID, 1, 20)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to count spins")

		mockSpinRepo.AssertExpectations(t)
	})

	t.Run("should handle get spins error", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()
		getErr := errors.New("database error")

		mockSpinRepo.On("Count", ctx, playerID).Return(int64(50), nil)
		mockSpinRepo.On("GetByPlayer", ctx, playerID, 20, 0).Return(nil, getErr)

		result, err := service.GetSpinHistory(ctx, playerID, 1, 20)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get spin history")

		mockSpinRepo.AssertExpectations(t)
	})

	t.Run("should return empty list when no spins", func(t *testing.T) {
		service, mockSpinRepo, _, _, _ := setupSpinServiceForValidation()

		playerID := uuid.New()

		mockSpinRepo.On("Count", ctx, playerID).Return(int64(0), nil)
		mockSpinRepo.On("GetByPlayer", ctx, playerID, 20, 0).Return([]*spin.Spin{}, nil)

		result, err := service.GetSpinHistory(ctx, playerID, 1, 20)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), result.Total)
		assert.Len(t, result.Spins, 0)

		mockSpinRepo.AssertExpectations(t)
	})
}
