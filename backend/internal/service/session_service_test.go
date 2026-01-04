package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MOCKS
// ============================================================================

// MockSessionRepository is a mock implementation of session.Repository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, s *session.GameSession) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*session.GameSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.GameSession), args.Error(1)
}

func (m *MockSessionRepository) GetActiveSessionByPlayer(ctx context.Context, playerID uuid.UUID) (*session.GameSession, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.GameSession), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, s *session.GameSession) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionRepository) EndSession(ctx context.Context, id uuid.UUID, endingBalance float64) error {
	args := m.Called(ctx, id, endingBalance)
	return args.Error(0)
}

func (m *MockSessionRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, spins int, wagered, won float64) error {
	args := m.Called(ctx, id, spins, wagered, won)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByPlayer(ctx context.Context, playerID uuid.UUID, limit, offset int) ([]*session.GameSession, error) {
	args := m.Called(ctx, playerID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*session.GameSession), args.Error(1)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func setupSessionService() (*SessionService, *MockSessionRepository, *MockPlayerRepository) {
	mockSessionRepo := new(MockSessionRepository)
	mockPlayerRepo := new(MockPlayerRepository)
	log := logger.New("info", "json")
	service := NewSessionService(mockSessionRepo, mockPlayerRepo, log).(*SessionService)
	return service, mockSessionRepo, mockPlayerRepo
}

// ============================================================================
// StartSession TESTS
// ============================================================================

func TestStartSession(t *testing.T) {
	ctx := context.Background()

	t.Run("should start session successfully", func(t *testing.T) {
		service, mockSessionRepo, mockPlayerRepo := setupSessionService()

		playerID := uuid.New()
		betAmount := 100.0
		mockPlayer := &player.Player{
			ID:       playerID,
			Username: "testuser",
			Balance:  10000.00,
			IsActive: true,
		}

		// Mock: get player
		mockPlayerRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Mock: no existing active session
		mockSessionRepo.On("GetActiveSessionByPlayer", ctx, playerID).Return(nil, session.ErrSessionNotFound)

		// Mock: create session
		mockSessionRepo.On("Create", ctx, mock.AnythingOfType("*session.GameSession")).Return(nil)

		// Execute
		sess, err := service.StartSession(ctx, playerID, betAmount)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, sess)
		assert.Equal(t, playerID, sess.PlayerID)
		assert.Equal(t, betAmount, sess.BetAmount)
		assert.Equal(t, 10000.00, sess.StartingBalance)
		assert.Nil(t, sess.EndingBalance)
		assert.Equal(t, 0, sess.TotalSpins)
		assert.Equal(t, 0.0, sess.TotalWagered)
		assert.Equal(t, 0.0, sess.TotalWon)
		assert.NotEqual(t, uuid.Nil, sess.ID)

		mockSessionRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid bet amount", func(t *testing.T) {
		service, _, _ := setupSessionService()

		playerID := uuid.New()

		tests := []struct {
			name      string
			betAmount float64
		}{
			{"zero bet", 0.0},
			{"negative bet", -10.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				sess, err := service.StartSession(ctx, playerID, tt.betAmount)

				assert.Error(t, err)
				assert.Nil(t, sess)
				assert.Contains(t, err.Error(), "must be positive")
			})
		}
	})

	t.Run("should return error for non-existent player", func(t *testing.T) {
		service, _, mockPlayerRepo := setupSessionService()

		playerID := uuid.New()

		// Mock: player not found
		mockPlayerRepo.On("GetByID", ctx, playerID).Return(nil, player.ErrPlayerNotFound)

		// Execute
		sess, err := service.StartSession(ctx, playerID, 100.0)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, sess)
		assert.Equal(t, player.ErrPlayerNotFound, err)

		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("should return error for inactive player", func(t *testing.T) {
		service, _, mockPlayerRepo := setupSessionService()

		playerID := uuid.New()
		mockPlayer := &player.Player{
			ID:       playerID,
			Username: "testuser",
			Balance:  10000.00,
			IsActive: false, // Inactive
		}

		// Mock: get inactive player
		mockPlayerRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Execute
		sess, err := service.StartSession(ctx, playerID, 100.0)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, sess)
		assert.Contains(t, err.Error(), "not active")

		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("should return error when active session exists", func(t *testing.T) {
		service, mockSessionRepo, mockPlayerRepo := setupSessionService()

		playerID := uuid.New()
		existingSessionID := uuid.New()
		mockPlayer := &player.Player{
			ID:       playerID,
			Balance:  10000.00,
			IsActive: true,
		}

		existingSession := &session.GameSession{
			ID:        existingSessionID,
			PlayerID:  playerID,
			BetAmount: 50.0,
			EndedAt:   nil, // Still active
		}

		// Mock: get player
		mockPlayerRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Mock: existing active session
		mockSessionRepo.On("GetActiveSessionByPlayer", ctx, playerID).Return(existingSession, nil)

		// Execute
		sess, err := service.StartSession(ctx, playerID, 100.0)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, sess)
		assert.Equal(t, session.ErrActiveSessionExists, err)

		mockSessionRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("should handle repository create error", func(t *testing.T) {
		service, mockSessionRepo, mockPlayerRepo := setupSessionService()

		playerID := uuid.New()
		mockPlayer := &player.Player{
			ID:       playerID,
			Balance:  10000.00,
			IsActive: true,
		}
		repoErr := errors.New("database error")

		// Mock: get player
		mockPlayerRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Mock: no existing session
		mockSessionRepo.On("GetActiveSessionByPlayer", ctx, playerID).Return(nil, session.ErrSessionNotFound)

		// Mock: create fails
		mockSessionRepo.On("Create", ctx, mock.AnythingOfType("*session.GameSession")).Return(repoErr)

		// Execute
		sess, err := service.StartSession(ctx, playerID, 100.0)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, sess)
		assert.Contains(t, err.Error(), "failed to create session")

		mockSessionRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})
}

// ============================================================================
// EndSession TESTS
// ============================================================================

func TestEndSession(t *testing.T) {
	ctx := context.Background()

	t.Run("should end session successfully", func(t *testing.T) {
		service, mockSessionRepo, mockPlayerRepo := setupSessionService()

		sessionID := uuid.New()
		playerID := uuid.New()
		mockSession := &session.GameSession{
			ID:              sessionID,
			PlayerID:        playerID,
			BetAmount:       100.0,
			StartingBalance: 10000.00,
			TotalSpins:      10,
			TotalWagered:    1000.00,
			TotalWon:        1200.00,
			EndedAt:         nil,
		}

		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 10200.00, // Current balance
		}

		updatedSession := &session.GameSession{
			ID:              sessionID,
			PlayerID:        playerID,
			BetAmount:       100.0,
			StartingBalance: 10000.00,
			EndingBalance:   &mockPlayer.Balance,
			TotalSpins:      10,
			TotalWagered:    1000.00,
			TotalWon:        1200.00,
			NetChange:       200.00,
			EndedAt:         &time.Time{},
		}

		// Mock: get session
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil).Once()

		// Mock: get player balance
		mockPlayerRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Mock: end session
		mockSessionRepo.On("EndSession", ctx, sessionID, 10200.00).Return(nil)

		// Mock: get updated session
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(updatedSession, nil).Once()

		// Execute
		sess, err := service.EndSession(ctx, sessionID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, sess)
		assert.NotNil(t, sess.EndingBalance)
		assert.Equal(t, 10200.00, *sess.EndingBalance)
		assert.Equal(t, 200.00, sess.NetChange)

		mockSessionRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		service, mockSessionRepo, _ := setupSessionService()

		sessionID := uuid.New()

		// Mock: session not found
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(nil, session.ErrSessionNotFound)

		// Execute
		sess, err := service.EndSession(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, sess)
		assert.Equal(t, session.ErrSessionNotFound, err)

		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("should return error for already ended session", func(t *testing.T) {
		service, mockSessionRepo, _ := setupSessionService()

		sessionID := uuid.New()
		endedTime := time.Now().UTC()
		mockSession := &session.GameSession{
			ID:        sessionID,
			PlayerID:  uuid.New(),
			BetAmount: 100.0,
			EndedAt:   &endedTime, // Already ended
		}

		// Mock: get ended session
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)

		// Execute
		sess, err := service.EndSession(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, sess)
		assert.Equal(t, session.ErrSessionAlreadyEnded, err)

		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("should handle player not found error", func(t *testing.T) {
		service, mockSessionRepo, mockPlayerRepo := setupSessionService()

		sessionID := uuid.New()
		playerID := uuid.New()
		mockSession := &session.GameSession{
			ID:        sessionID,
			PlayerID:  playerID,
			BetAmount: 100.0,
			EndedAt:   nil,
		}

		// Mock: get session
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)

		// Mock: player not found
		mockPlayerRepo.On("GetByID", ctx, playerID).Return(nil, player.ErrPlayerNotFound)

		// Execute
		sess, err := service.EndSession(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, sess)
		assert.Equal(t, player.ErrPlayerNotFound, err)

		mockSessionRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("should handle repository end session error", func(t *testing.T) {
		service, mockSessionRepo, mockPlayerRepo := setupSessionService()

		sessionID := uuid.New()
		playerID := uuid.New()
		mockSession := &session.GameSession{
			ID:        sessionID,
			PlayerID:  playerID,
			BetAmount: 100.0,
			EndedAt:   nil,
		}

		mockPlayer := &player.Player{
			ID:      playerID,
			Balance: 10200.00,
		}

		repoErr := errors.New("database error")

		// Mock: get session
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)

		// Mock: get player
		mockPlayerRepo.On("GetByID", ctx, playerID).Return(mockPlayer, nil)

		// Mock: end session fails
		mockSessionRepo.On("EndSession", ctx, sessionID, 10200.00).Return(repoErr)

		// Execute
		sess, err := service.EndSession(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, sess)
		assert.Contains(t, err.Error(), "failed to end session")

		mockSessionRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})
}

// ============================================================================
// GetSession TESTS
// ============================================================================

func TestGetSession(t *testing.T) {
	ctx := context.Background()

	t.Run("should get session successfully", func(t *testing.T) {
		service, mockSessionRepo, _ := setupSessionService()

		sessionID := uuid.New()
		mockSession := &session.GameSession{
			ID:              sessionID,
			PlayerID:        uuid.New(),
			BetAmount:       100.0,
			StartingBalance: 10000.00,
			TotalSpins:      5,
			TotalWagered:    500.00,
			TotalWon:        600.00,
		}

		// Mock: get session
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(mockSession, nil)

		// Execute
		sess, err := service.GetSession(ctx, sessionID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, sess)
		assert.Equal(t, sessionID, sess.ID)
		assert.Equal(t, 100.0, sess.BetAmount)
		assert.Equal(t, 5, sess.TotalSpins)

		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		service, mockSessionRepo, _ := setupSessionService()

		sessionID := uuid.New()

		// Mock: session not found
		mockSessionRepo.On("GetByID", ctx, sessionID).Return(nil, session.ErrSessionNotFound)

		// Execute
		sess, err := service.GetSession(ctx, sessionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, sess)
		assert.Equal(t, session.ErrSessionNotFound, err)

		mockSessionRepo.AssertExpectations(t)
	})
}

// ============================================================================
// GetPlayerSessions TESTS
// ============================================================================

func TestGetPlayerSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("should get player sessions successfully", func(t *testing.T) {
		service, mockSessionRepo, _ := setupSessionService()

		playerID := uuid.New()
		mockSessions := []*session.GameSession{
			{
				ID:              uuid.New(),
				PlayerID:        playerID,
				BetAmount:       100.0,
				StartingBalance: 10000.00,
			},
			{
				ID:              uuid.New(),
				PlayerID:        playerID,
				BetAmount:       50.0,
				StartingBalance: 9500.00,
			},
		}

		// Mock: get sessions
		mockSessionRepo.On("GetByPlayer", ctx, playerID, 20, 0).Return(mockSessions, nil)

		// Execute
		sessions, err := service.GetPlayerSessions(ctx, playerID, 1, 20)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, sessions)
		assert.Len(t, sessions, 2)
		assert.Equal(t, playerID, sessions[0].PlayerID)
		assert.Equal(t, playerID, sessions[1].PlayerID)

		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("should use default pagination values", func(t *testing.T) {
		service, mockSessionRepo, _ := setupSessionService()

		playerID := uuid.New()
		mockSessions := []*session.GameSession{}

		tests := []struct {
			name           string
			page           int
			limit          int
			expectedLimit  int
			expectedOffset int
		}{
			{"invalid page", 0, 10, 10, 0},       // page < 1 defaults to 1
			{"invalid limit", 1, 0, 20, 0},       // limit < 1 defaults to 20
			{"limit too high", 1, 200, 20, 0},    // limit > 100 defaults to 20
			{"valid params", 2, 10, 10, 10},      // page 2, limit 10 -> offset 10
			{"page 3", 3, 15, 15, 30},            // page 3, limit 15 -> offset 30
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Mock: get sessions with expected offset
				mockSessionRepo.On("GetByPlayer", ctx, playerID, tt.expectedLimit, tt.expectedOffset).Return(mockSessions, nil).Once()

				// Execute
				sessions, err := service.GetPlayerSessions(ctx, playerID, tt.page, tt.limit)

				// Assert
				require.NoError(t, err)
				assert.NotNil(t, sessions)

				mockSessionRepo.AssertExpectations(t)
			})
		}
	})

	t.Run("should handle repository error", func(t *testing.T) {
		service, mockSessionRepo, _ := setupSessionService()

		playerID := uuid.New()
		repoErr := errors.New("database error")

		// Mock: repository error
		mockSessionRepo.On("GetByPlayer", ctx, playerID, 20, 0).Return(nil, repoErr)

		// Execute
		sessions, err := service.GetPlayerSessions(ctx, playerID, 1, 20)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, sessions)
		assert.Contains(t, err.Error(), "failed to get player sessions")

		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("should return empty list when no sessions", func(t *testing.T) {
		service, mockSessionRepo, _ := setupSessionService()

		playerID := uuid.New()
		mockSessions := []*session.GameSession{}

		// Mock: empty result
		mockSessionRepo.On("GetByPlayer", ctx, playerID, 20, 0).Return(mockSessions, nil)

		// Execute
		sessions, err := service.GetPlayerSessions(ctx, playerID, 1, 20)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, sessions)
		assert.Len(t, sessions, 0)

		mockSessionRepo.AssertExpectations(t)
	})
}
