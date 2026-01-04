package service

import (
	"github.com/google/wire"
	"github.com/slotmachine/backend/domain/freespins"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/domain/spin"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/game/engine"
	infraCache "github.com/slotmachine/backend/internal/infra/cache"
	redisCache "github.com/slotmachine/backend/internal/infra/cache"
	"github.com/slotmachine/backend/internal/infra/repository"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// ProviderSet is the Wire provider set for services
var ProviderSet = wire.NewSet(
	NewPlayerService,
	NewSessionService,
	ProvideSpinService,
	wire.Bind(new(spin.Service), new(*SpinService)),
	ProvideFreeSpinsService,
	wire.Bind(new(freespins.Service), new(*FreeSpinsService)),
	NewReelStripService,
	NewAdminService,
	ProvideProvablyFairService,
	ProvideTrialService,
)

// ProvideTrialService provides the TrialService
func ProvideTrialService(
	cache *redisCache.RedisClient,
	log *logger.Logger,
) *TrialService {
	return NewTrialService(cache, log)
}

// ProvideSpinService provides a concrete SpinService with required pfService
func ProvideSpinService(
	spinRepo spin.Repository,
	playerRepo player.Repository,
	sessionRepo session.Repository,
	gameEngine *engine.GameEngine,
	freespinsRepo freespins.Repository,
	reelstripRepo reelstrip.Repository,
	txManager *repository.TxManager,
	pfService *ProvablyFairService,
	log *logger.Logger,
) *SpinService {
	return &SpinService{
		spinRepo:      spinRepo,
		playerRepo:    playerRepo,
		sessionRepo:   sessionRepo,
		gameEngine:    gameEngine,
		freespinsRepo: freespinsRepo,
		reelstripRepo: reelstripRepo,
		txManager:     txManager,
		pfService:     pfService,
		trialService:  nil, // Trial service set separately via SetTrialService
		logger:        log,
	}
}

// ProvideFreeSpinsService provides a concrete FreeSpinsService with required pfService
func ProvideFreeSpinsService(
	sessionRepo session.Repository,
	freespinsRepo freespins.Repository,
	spinRepo spin.Repository,
	playerRepo player.Repository,
	gameEngine *engine.GameEngine,
	pfService *ProvablyFairService,
	log *logger.Logger,
) *FreeSpinsService {
	return &FreeSpinsService{
		sessionRepo:   sessionRepo,
		freespinsRepo: freespinsRepo,
		spinRepo:      spinRepo,
		playerRepo:    playerRepo,
		gameEngine:    gameEngine,
		pfService:     pfService,
		logger:        log,
	}
}

// ProvideProvablyFairService provides the ProvablyFairService
// Takes concrete types directly (from repository and cache ProviderSets)
func ProvideProvablyFairService(
	repo *repository.ProvablyFairGormRepository,
	cache *infraCache.PFSessionCache,
	reelstripRepo reelstrip.Repository,
	cfg *config.Config,
	log *logger.Logger,
) (*ProvablyFairService, error) {
	svc, err := NewProvablyFairService(repo, cache, reelstripRepo, cfg, log)
	if err != nil {
		return nil, err
	}
	return svc.(*ProvablyFairService), nil
}
