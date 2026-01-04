package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// MOCKS
// ============================================================================

// MockReelStripRepository is a mock implementation of reelstrip.Repository
type MockReelStripRepository struct {
	mock.Mock
}

// ReelStrip operations
func (m *MockReelStripRepository) Create(ctx context.Context, strip *reelstrip.ReelStrip) error {
	args := m.Called(ctx, strip)
	return args.Error(0)
}

func (m *MockReelStripRepository) CreateBatch(ctx context.Context, strips []*reelstrip.ReelStrip) error {
	args := m.Called(ctx, strips)
	return args.Error(0)
}

func (m *MockReelStripRepository) GetByID(ctx context.Context, id uuid.UUID) (*reelstrip.ReelStrip, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reelstrip.ReelStrip), args.Error(1)
}

func (m *MockReelStripRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*reelstrip.ReelStrip, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*reelstrip.ReelStrip), args.Error(1)
}

func (m *MockReelStripRepository) Update(ctx context.Context, strip *reelstrip.ReelStrip) error {
	args := m.Called(ctx, strip)
	return args.Error(0)
}

func (m *MockReelStripRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReelStripRepository) GetAllActive(ctx context.Context, gameMode string) ([]*reelstrip.ReelStrip, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*reelstrip.ReelStrip), args.Error(1)
}

func (m *MockReelStripRepository) GetByGameModeAndReel(ctx context.Context, gameMode string, reelNumber int) ([]*reelstrip.ReelStrip, error) {
	args := m.Called(ctx, gameMode, reelNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*reelstrip.ReelStrip), args.Error(1)
}

func (m *MockReelStripRepository) CountActive(ctx context.Context, gameMode string) (map[int]int, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int]int), args.Error(1)
}

func (m *MockReelStripRepository) DeactivateOldVersions(ctx context.Context, gameMode string, keepVersion int) error {
	args := m.Called(ctx, gameMode, keepVersion)
	return args.Error(0)
}

// ReelStripConfig operations
func (m *MockReelStripRepository) CreateConfig(ctx context.Context, config *reelstrip.ReelStripConfig) error {
	args := m.Called(ctx, config)
	// Simulate ID assignment
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockReelStripRepository) GetConfigByID(ctx context.Context, id uuid.UUID) (*reelstrip.ReelStripConfig, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reelstrip.ReelStripConfig), args.Error(1)
}

func (m *MockReelStripRepository) GetConfigByName(ctx context.Context, name string) (*reelstrip.ReelStripConfig, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reelstrip.ReelStripConfig), args.Error(1)
}

func (m *MockReelStripRepository) GetDefaultConfig(ctx context.Context, gameMode string) (*reelstrip.ReelStripConfig, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reelstrip.ReelStripConfig), args.Error(1)
}

func (m *MockReelStripRepository) UpdateConfig(ctx context.Context, config *reelstrip.ReelStripConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockReelStripRepository) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReelStripRepository) SetDefaultConfig(ctx context.Context, id uuid.UUID, gameMode string) error {
	args := m.Called(ctx, id, gameMode)
	return args.Error(0)
}

func (m *MockReelStripRepository) GetSetByConfigID(ctx context.Context, configID uuid.UUID) (*reelstrip.ReelStripConfigSet, error) {
	args := m.Called(ctx, configID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reelstrip.ReelStripConfigSet), args.Error(1)
}

// PlayerReelStripAssignment operations
func (m *MockReelStripRepository) CreateAssignment(ctx context.Context, assignment *reelstrip.PlayerReelStripAssignment) error {
	args := m.Called(ctx, assignment)
	// Simulate ID assignment
	if assignment.ID == uuid.Nil {
		assignment.ID = uuid.New()
	}
	return args.Error(0)
}

func (m *MockReelStripRepository) GetPlayerAssignment(ctx context.Context, playerID uuid.UUID) (*reelstrip.PlayerReelStripAssignment, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reelstrip.PlayerReelStripAssignment), args.Error(1)
}

func (m *MockReelStripRepository) UpdateAssignment(ctx context.Context, assignment *reelstrip.PlayerReelStripAssignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *MockReelStripRepository) DeleteAssignment(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReelStripRepository) GetPlayerAssignmentsByPlayerIDs(ctx context.Context, playerIDs []uuid.UUID) (map[uuid.UUID]*reelstrip.PlayerReelStripAssignment, error) {
	args := m.Called(ctx, playerIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]*reelstrip.PlayerReelStripAssignment), args.Error(1)
}

func (m *MockReelStripRepository) ListConfigs(ctx context.Context, filters *reelstrip.ConfigListFilters) ([]*reelstrip.ReelStripConfig, int64, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*reelstrip.ReelStripConfig), args.Get(1).(int64), args.Error(2)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func setupReelStripService() (*ReelStripService, *MockReelStripRepository) {
	mockRepo := new(MockReelStripRepository)
	log := logger.New("info", "json")
	service := NewReelStripService(mockRepo, log).(*ReelStripService)
	return service, mockRepo
}

func createMockReelStrip(reelNum int, gameMode string) *reelstrip.ReelStrip {
	return &reelstrip.ReelStrip{
		ID:          uuid.New(),
		GameMode:    gameMode,
		ReelNumber:  reelNum,
		StripData:   []string{"cai", "fu", "shu", "zhong", "liangtong"},
		Checksum:    "mock-checksum",
		StripLength: 5,
		IsActive:    true,
	}
}

// ============================================================================
// GetStripByID TESTS
// ============================================================================

func TestGetStripByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get strip successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		stripID := uuid.New()
		mockStrip := createMockReelStrip(0, "base_game")
		mockStrip.ID = stripID

		mockRepo.On("GetByID", ctx, stripID).Return(mockStrip, nil)

		strip, err := service.GetStripByID(ctx, stripID)

		require.NoError(t, err)
		assert.NotNil(t, strip)
		assert.Equal(t, stripID, strip.ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when strip not found", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		stripID := uuid.New()

		mockRepo.On("GetByID", ctx, stripID).Return(nil, reelstrip.ErrReelStripNotFound)

		strip, err := service.GetStripByID(ctx, stripID)

		assert.Error(t, err)
		assert.Nil(t, strip)

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// CreateConfig TESTS
// ============================================================================

func TestCreateConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("should create config successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		name := "Test Config"
		gameMode := "base_game"
		description := "Test description"
		reelStripIDs := [5]uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
		targetRTP := 96.5

		// Mock: Verify all strips exist
		mockStrips := []*reelstrip.ReelStrip{
			createMockReelStrip(0, gameMode),
			createMockReelStrip(1, gameMode),
			createMockReelStrip(2, gameMode),
			createMockReelStrip(3, gameMode),
			createMockReelStrip(4, gameMode),
		}
		mockRepo.On("GetByIDs", ctx, reelStripIDs[:]).Return(mockStrips, nil)

		// Mock: Create config
		mockRepo.On("CreateConfig", ctx, mock.AnythingOfType("*reelstrip.ReelStripConfig")).Return(nil)

		config, err := service.CreateConfig(ctx, name, gameMode, description, reelStripIDs, targetRTP, nil)

		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, name, config.Name)
		assert.Equal(t, gameMode, config.GameMode)
		assert.Equal(t, targetRTP, config.TargetRTP)
		assert.True(t, config.IsActive)
		assert.False(t, config.IsDefault)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid game mode", func(t *testing.T) {
		service, _ := setupReelStripService()

		reelStripIDs := [5]uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}

		config, err := service.CreateConfig(ctx, "Test", "invalid_mode", "desc", reelStripIDs, 96.0, nil)

		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Equal(t, reelstrip.ErrInvalidGameMode, err)
	})

	t.Run("should return error for incomplete strip set", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		reelStripIDs := [5]uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}

		// Return only 3 strips instead of 5
		incompleteStrips := []*reelstrip.ReelStrip{
			createMockReelStrip(0, "base_game"),
			createMockReelStrip(1, "base_game"),
			createMockReelStrip(2, "base_game"),
		}
		mockRepo.On("GetByIDs", ctx, reelStripIDs[:]).Return(incompleteStrips, nil)

		config, err := service.CreateConfig(ctx, "Test", "base_game", "desc", reelStripIDs, 96.0, nil)

		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Equal(t, reelstrip.ErrIncompleteSet, err)

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// GetConfigByID & GetConfigByName TESTS
// ============================================================================

func TestGetConfigByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get config successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		configID := uuid.New()
		mockConfig := &reelstrip.ReelStripConfig{
			ID:       configID,
			Name:     "Test Config",
			GameMode: "base_game",
		}

		mockRepo.On("GetConfigByID", ctx, configID).Return(mockConfig, nil)

		config, err := service.GetConfigByID(ctx, configID)

		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, configID, config.ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when config not found", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		configID := uuid.New()

		mockRepo.On("GetConfigByID", ctx, configID).Return(nil, reelstrip.ErrConfigNotFound)

		config, err := service.GetConfigByID(ctx, configID)

		assert.Error(t, err)
		assert.Nil(t, config)

		mockRepo.AssertExpectations(t)
	})
}

func TestGetConfigByName(t *testing.T) {
	ctx := context.Background()

	t.Run("should get config by name successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		configName := "Test Config"
		mockConfig := &reelstrip.ReelStripConfig{
			ID:       uuid.New(),
			Name:     configName,
			GameMode: "base_game",
		}

		mockRepo.On("GetConfigByName", ctx, configName).Return(mockConfig, nil)

		config, err := service.GetConfigByName(ctx, configName)

		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, configName, config.Name)

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// SetDefaultConfig TESTS
// ============================================================================

func TestSetDefaultConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("should set default config successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		configID := uuid.New()
		gameMode := "base_game"

		mockRepo.On("SetDefaultConfig", ctx, configID, gameMode).Return(nil)

		err := service.SetDefaultConfig(ctx, configID, gameMode)

		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid game mode", func(t *testing.T) {
		service, _ := setupReelStripService()

		configID := uuid.New()

		err := service.SetDefaultConfig(ctx, configID, "invalid_mode")

		assert.Error(t, err)
		assert.Equal(t, reelstrip.ErrInvalidGameMode, err)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		configID := uuid.New()
		gameMode := "base_game"
		repoErr := errors.New("database error")

		mockRepo.On("SetDefaultConfig", ctx, configID, gameMode).Return(repoErr)

		err := service.SetDefaultConfig(ctx, configID, gameMode)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set default config")

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// ActivateConfig & DeactivateConfig TESTS
// ============================================================================

func TestActivateConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("should activate config successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		configID := uuid.New()
		mockConfig := &reelstrip.ReelStripConfig{
			ID:       configID,
			Name:     "Test Config",
			IsActive: false,
		}

		mockRepo.On("GetConfigByID", ctx, configID).Return(mockConfig, nil)
		mockRepo.On("UpdateConfig", ctx, mockConfig).Return(nil)

		err := service.ActivateConfig(ctx, configID)

		require.NoError(t, err)
		assert.True(t, mockConfig.IsActive)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when config not found", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		configID := uuid.New()

		mockRepo.On("GetConfigByID", ctx, configID).Return(nil, reelstrip.ErrConfigNotFound)

		err := service.ActivateConfig(ctx, configID)

		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestDeactivateConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("should deactivate config successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		configID := uuid.New()
		mockConfig := &reelstrip.ReelStripConfig{
			ID:       configID,
			Name:     "Test Config",
			IsActive: true,
		}

		mockRepo.On("GetConfigByID", ctx, configID).Return(mockConfig, nil)
		mockRepo.On("UpdateConfig", ctx, mockConfig).Return(nil)

		err := service.DeactivateConfig(ctx, configID)

		require.NoError(t, err)
		assert.False(t, mockConfig.IsActive)

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// ListConfigs TESTS
// ============================================================================

func TestListConfigs(t *testing.T) {
	ctx := context.Background()

	t.Run("should list configs successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		gameMode := "base_game"
		isActive := true
		filters := &reelstrip.ConfigListFilters{
			GameMode: &gameMode,
			IsActive: &isActive,
			Page:     1,
			Limit:    20,
		}

		mockConfigs := []*reelstrip.ReelStripConfig{
			{ID: uuid.New(), Name: "Config 1", GameMode: gameMode, IsActive: true},
			{ID: uuid.New(), Name: "Config 2", GameMode: gameMode, IsActive: true},
		}

		mockRepo.On("ListConfigs", ctx, filters).Return(mockConfigs, int64(2), nil)

		configs, total, err := service.ListConfigs(ctx, filters)

		require.NoError(t, err)
		assert.Len(t, configs, 2)
		assert.Equal(t, int64(2), total)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid game mode", func(t *testing.T) {
		service, _ := setupReelStripService()

		invalidGameMode := "invalid_mode"
		filters := &reelstrip.ConfigListFilters{
			GameMode: &invalidGameMode,
			Page:     1,
			Limit:    20,
		}

		configs, total, err := service.ListConfigs(ctx, filters)

		assert.Error(t, err)
		assert.Nil(t, configs)
		assert.Equal(t, int64(0), total)
		assert.Equal(t, reelstrip.ErrInvalidGameMode, err)
	})

	t.Run("should handle pagination", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		gameMode := "base_game"
		filters := &reelstrip.ConfigListFilters{
			GameMode: &gameMode,
			Page:     2,
			Limit:    10,
		}

		mockConfigs := []*reelstrip.ReelStripConfig{
			{ID: uuid.New(), Name: "Config 11", GameMode: gameMode},
		}

		mockRepo.On("ListConfigs", ctx, filters).Return(mockConfigs, int64(15), nil)

		configs, total, err := service.ListConfigs(ctx, filters)

		require.NoError(t, err)
		assert.Len(t, configs, 1)
		assert.Equal(t, int64(15), total)

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// AssignConfigToPlayer TESTS
// ============================================================================

func TestAssignConfigToPlayer(t *testing.T) {
	ctx := context.Background()

	t.Run("should create new assignment successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		playerID := uuid.New()
		configID := uuid.New()
		gameMode := "base_game"
		reason := "A/B testing"

		mockConfig := &reelstrip.ReelStripConfig{
			ID:       configID,
			GameMode: gameMode,
		}

		mockRepo.On("GetConfigByID", ctx, configID).Return(mockConfig, nil)
		mockRepo.On("GetPlayerAssignment", ctx, playerID).Return(nil, reelstrip.ErrAssignmentNotFound)
		mockRepo.On("CreateAssignment", ctx, mock.AnythingOfType("*reelstrip.PlayerReelStripAssignment")).Return(nil)

		err := service.AssignConfigToPlayer(ctx, playerID, configID, gameMode, reason, "admin", nil)

		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should update existing assignment successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		playerID := uuid.New()
		configID := uuid.New()
		gameMode := "base_game"
		reason := "Updated config"

		mockConfig := &reelstrip.ReelStripConfig{
			ID:       configID,
			GameMode: gameMode,
		}

		oldConfigID := uuid.New()
		existingAssignment := &reelstrip.PlayerReelStripAssignment{
			ID:                uuid.New(),
			PlayerID:          playerID,
			BaseGameConfigID:  &oldConfigID,
			FreeSpinsConfigID: nil,
			Reason:            "Old reason",
			IsActive:          true,
		}

		mockRepo.On("GetConfigByID", ctx, configID).Return(mockConfig, nil)
		mockRepo.On("GetPlayerAssignment", ctx, playerID).Return(existingAssignment, nil)
		mockRepo.On("UpdateAssignment", ctx, existingAssignment).Return(nil)

		err := service.AssignConfigToPlayer(ctx, playerID, configID, gameMode, reason, "admin", nil)

		require.NoError(t, err)
		assert.Equal(t, configID, *existingAssignment.BaseGameConfigID)
		assert.Equal(t, reason, existingAssignment.Reason)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for game mode mismatch", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		playerID := uuid.New()
		configID := uuid.New()

		mockConfig := &reelstrip.ReelStripConfig{
			ID:       configID,
			GameMode: "base_game",
		}

		mockRepo.On("GetConfigByID", ctx, configID).Return(mockConfig, nil)

		// Request free_spins but config is base_game
		err := service.AssignConfigToPlayer(ctx, playerID, configID, "free_spins", "test", "admin", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match requested game mode")

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when config not found", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		playerID := uuid.New()
		configID := uuid.New()

		mockRepo.On("GetConfigByID", ctx, configID).Return(nil, reelstrip.ErrConfigNotFound)

		err := service.AssignConfigToPlayer(ctx, playerID, configID, "base_game", "test", "admin", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config not found")

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// GetPlayerAssignment & RemovePlayerAssignment TESTS
// ============================================================================

func TestGetPlayerAssignment(t *testing.T) {
	ctx := context.Background()

	t.Run("should get player assignment successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		playerID := uuid.New()
		baseConfigID := uuid.New()
		freeSpinsConfigID := uuid.New()
		mockAssignment := &reelstrip.PlayerReelStripAssignment{
			ID:                uuid.New(),
			PlayerID:          playerID,
			BaseGameConfigID:  &baseConfigID,
			FreeSpinsConfigID: &freeSpinsConfigID,
			IsActive:          true,
		}

		mockRepo.On("GetPlayerAssignment", ctx, playerID).Return(mockAssignment, nil)

		assignment, err := service.GetPlayerAssignment(ctx, playerID)

		require.NoError(t, err)
		assert.NotNil(t, assignment)
		assert.Equal(t, playerID, assignment.PlayerID)

		mockRepo.AssertExpectations(t)
	})
}

func TestRemovePlayerAssignment(t *testing.T) {
	ctx := context.Background()

	t.Run("should remove assignment successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		playerID := uuid.New()
		assignmentID := uuid.New()
		mockAssignment := &reelstrip.PlayerReelStripAssignment{
			ID:       assignmentID,
			PlayerID: playerID,
			IsActive: true,
		}

		mockRepo.On("GetPlayerAssignment", ctx, playerID).Return(mockAssignment, nil)
		mockRepo.On("DeleteAssignment", ctx, assignmentID).Return(nil)

		err := service.RemovePlayerAssignment(ctx, playerID)

		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when assignment not found", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		playerID := uuid.New()

		mockRepo.On("GetPlayerAssignment", ctx, playerID).Return(nil, reelstrip.ErrAssignmentNotFound)

		err := service.RemovePlayerAssignment(ctx, playerID)

		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// GetReelSetForPlayer TESTS (Priority System)
// ============================================================================

func TestGetReelSetForPlayer(t *testing.T) {
	ctx := context.Background()

	t.Run("should use player assignment (priority 1)", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		playerID := uuid.New()
		configID := uuid.New()
		gameMode := "base_game"

		mockAssignment := &reelstrip.PlayerReelStripAssignment{
			PlayerID:         playerID,
			BaseGameConfigID: &configID,
		}

		mockConfigSet := &reelstrip.ReelStripConfigSet{
			Config: &reelstrip.ReelStripConfig{ID: configID},
			Strips: [5]*reelstrip.ReelStrip{
				createMockReelStrip(0, gameMode),
				createMockReelStrip(1, gameMode),
				createMockReelStrip(2, gameMode),
				createMockReelStrip(3, gameMode),
				createMockReelStrip(4, gameMode),
			},
		}

		mockRepo.On("GetPlayerAssignment", ctx, playerID).Return(mockAssignment, nil)
		mockRepo.On("GetSetByConfigID", ctx, configID).Return(mockConfigSet, nil)

		configSet, err := service.GetReelSetForPlayer(ctx, playerID, gameMode)

		require.NoError(t, err)
		assert.NotNil(t, configSet)
		assert.Equal(t, configID, configSet.Config.ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should use default config (priority 2)", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		playerID := uuid.New()
		configID := uuid.New()
		gameMode := "base_game"

		mockDefaultConfig := &reelstrip.ReelStripConfig{
			ID:        configID,
			IsDefault: true,
		}

		mockConfigSet := &reelstrip.ReelStripConfigSet{
			Config: mockDefaultConfig,
			Strips: [5]*reelstrip.ReelStrip{
				createMockReelStrip(0, gameMode),
				createMockReelStrip(1, gameMode),
				createMockReelStrip(2, gameMode),
				createMockReelStrip(3, gameMode),
				createMockReelStrip(4, gameMode),
			},
		}

		mockRepo.On("GetPlayerAssignment", ctx, playerID).Return(nil, reelstrip.ErrAssignmentNotFound)
		mockRepo.On("GetDefaultConfig", ctx, gameMode).Return(mockDefaultConfig, nil)
		mockRepo.On("GetSetByConfigID", ctx, configID).Return(mockConfigSet, nil)

		configSet, err := service.GetReelSetForPlayer(ctx, playerID, gameMode)

		require.NoError(t, err)
		assert.NotNil(t, configSet)
		assert.Equal(t, configID, configSet.Config.ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should fallback to legacy random (priority 3)", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		playerID := uuid.New()
		gameMode := "base_game"

		// Mock all strips available for legacy mode
		mockStrips := []*reelstrip.ReelStrip{createMockReelStrip(0, gameMode)}

		mockRepo.On("GetPlayerAssignment", ctx, playerID).Return(nil, reelstrip.ErrAssignmentNotFound)
		mockRepo.On("GetDefaultConfig", ctx, gameMode).Return(nil, reelstrip.ErrNoDefaultConfig)

		// Legacy mode needs strips for each reel
		for i := 0; i < 5; i++ {
			mockRepo.On("GetByGameModeAndReel", ctx, gameMode, i).Return(mockStrips, nil)
		}

		configSet, err := service.GetReelSetForPlayer(ctx, playerID, gameMode)

		require.NoError(t, err)
		assert.NotNil(t, configSet)
		assert.Nil(t, configSet.Config) // No config in legacy mode

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid game mode", func(t *testing.T) {
		service, _ := setupReelStripService()

		playerID := uuid.New()

		configSet, err := service.GetReelSetForPlayer(ctx, playerID, "invalid_mode")

		assert.Error(t, err)
		assert.Nil(t, configSet)
		assert.Equal(t, reelstrip.ErrInvalidGameMode, err)
	})
}

// ============================================================================
// GetDefaultReelSet TESTS
// ============================================================================

func TestGetDefaultReelSet(t *testing.T) {
	ctx := context.Background()

	t.Run("should get default reel set successfully", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		gameMode := "base_game"
		configID := uuid.New()

		mockDefaultConfig := &reelstrip.ReelStripConfig{
			ID:        configID,
			IsDefault: true,
		}

		mockConfigSet := &reelstrip.ReelStripConfigSet{
			Config: mockDefaultConfig,
			Strips: [5]*reelstrip.ReelStrip{
				createMockReelStrip(0, gameMode),
				createMockReelStrip(1, gameMode),
				createMockReelStrip(2, gameMode),
				createMockReelStrip(3, gameMode),
				createMockReelStrip(4, gameMode),
			},
		}

		mockRepo.On("GetDefaultConfig", ctx, gameMode).Return(mockDefaultConfig, nil)
		mockRepo.On("GetSetByConfigID", ctx, configID).Return(mockConfigSet, nil)

		configSet, err := service.GetDefaultReelSet(ctx, gameMode)

		require.NoError(t, err)
		assert.NotNil(t, configSet)
		assert.True(t, configSet.Config.IsDefault)

		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when no default config", func(t *testing.T) {
		service, mockRepo := setupReelStripService()

		gameMode := "base_game"

		mockRepo.On("GetDefaultConfig", ctx, gameMode).Return(nil, reelstrip.ErrNoDefaultConfig)

		configSet, err := service.GetDefaultReelSet(ctx, gameMode)

		assert.Error(t, err)
		assert.Nil(t, configSet)

		mockRepo.AssertExpectations(t)
	})
}

// ============================================================================
// ValidateStripIntegrity TESTS
// ============================================================================

func TestValidateStripIntegrity(t *testing.T) {
	service, _ := setupReelStripService()

	t.Run("should validate successfully for valid strip", func(t *testing.T) {
		strip := &reelstrip.ReelStrip{
			StripData:   []string{"cai", "fu", "shu"},
			StripLength: 3,
		}
		// Calculate correct checksum
		strip.Checksum = service.calculateChecksum(strip.StripData)

		err := service.ValidateStripIntegrity(strip)

		require.NoError(t, err)
	})

	t.Run("should return error for nil strip", func(t *testing.T) {
		err := service.ValidateStripIntegrity(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "strip is nil")
	})

	t.Run("should return error for checksum mismatch", func(t *testing.T) {
		strip := &reelstrip.ReelStrip{
			StripData:   []string{"cai", "fu", "shu"},
			Checksum:    "invalid-checksum",
			StripLength: 3,
		}

		err := service.ValidateStripIntegrity(strip)

		assert.Error(t, err)
		assert.Equal(t, reelstrip.ErrChecksumMismatch, err)
	})

	t.Run("should return error for invalid strip length", func(t *testing.T) {
		strip := &reelstrip.ReelStrip{
			StripData:   []string{"cai", "fu", "shu"},
			StripLength: 5, // Wrong length
		}
		strip.Checksum = service.calculateChecksum(strip.StripData)

		err := service.ValidateStripIntegrity(strip)

		assert.Error(t, err)
		assert.Equal(t, reelstrip.ErrInvalidStripLength, err)
	})
}
