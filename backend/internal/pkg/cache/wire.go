package cache

import (
	"fmt"

	"github.com/google/wire"
	infraCache "github.com/slotmachine/backend/internal/infra/cache"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// ProviderSet is the Wire provider set for cache
var ProviderSet = wire.NewSet(
	ProvideCache,
	ProvideRedisClient,
	ProvidePFSessionCache,
)

// ProvideRedisClient provides the Redis client for session caching
func ProvideRedisClient(cfg *config.Config, log *logger.Logger) *infraCache.RedisClient {
	redisClient, err := infraCache.NewRedisClient(cfg, log)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to Redis for session cache")
		return nil
	}
	return redisClient
}

// ProvidePFSessionCache provides the PF session cache
func ProvidePFSessionCache(redisClient *infraCache.RedisClient, log *logger.Logger) *infraCache.PFSessionCache {
	return infraCache.NewPFSessionCache(redisClient, log)
}

func ProvideCache(cfg *config.Config, log *logger.Logger) *Cache {
	var bus EventBus
	var redisCloser RedisCloser

	// Try to initialize Redis if enabled
	redisClient, err := infraCache.NewRedisClient(cfg, log)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to Redis, cache will run without distributed sync")
	}

	// If Redis is available, create the event bus
	if redisClient != nil && redisClient.GetClient() != nil {
		bus = infraCache.NewRedisBus(redisClient.GetClient(), log)
		redisCloser = redisClient
		log.Info().Msg("Cache initialized with Redis event bus")
	} else {
		log.Info().Msg("Cache initialized without event bus (local only)")
	}

	params := NewCacheParams{
		Bus:         bus,
		Channel:     fmt.Sprintf("%s:%s:cache", cfg.App.Name, cfg.App.Env),
		Config:      cfg,
		RedisClient: redisCloser,
	}
	return NewCache(params)
}
