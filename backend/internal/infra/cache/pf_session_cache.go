package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/slotmachine/backend/domain/provablyfair"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// PF Session cache key prefixes
const (
	PFSessionKeyPrefix           = "pf_session:"              // Primary: pf_session:{session_id}
	PFSessionByPlayerKeyPrefix   = "pf_session_player:"       // Index: pf_session_player:{player_id}
	PFSessionByGameSessionPrefix = "pf_session_game_session:" // Index: pf_session_game_session:{game_session_id}
	PFSessionTTL                 = 24 * time.Hour             // TTL for PF session state
)

// PFSessionCache implements provablyfair.CacheRepository using Redis
type PFSessionCache struct {
	client *RedisClient
	logger *logger.Logger
}

// NewPFSessionCache creates a new PF session cache
func NewPFSessionCache(client *RedisClient, log *logger.Logger) *PFSessionCache {
	return &PFSessionCache{
		client: client,
		logger: log,
	}
}

// Ensure PFSessionCache implements CacheRepository
var _ provablyfair.CacheRepository = (*PFSessionCache)(nil)

// SetSessionState stores PF session state in Redis with secondary indexes
func (c *PFSessionCache) SetSessionState(ctx context.Context, state *provablyfair.PFSessionState) error {
	if c.client == nil {
		return fmt.Errorf("redis client is not available")
	}

	// Serialize state
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal PF session state: %w", err)
	}

	// Use pipeline for atomic operations
	pipe := c.client.GetClient().Pipeline()

	// Primary key: pf_session:{session_id}
	primaryKey := PFSessionKeyPrefix + state.SessionID.String()
	pipe.Set(ctx, primaryKey, data, PFSessionTTL)

	// Secondary index: pf_session_player:{player_id} -> session_id
	playerIndexKey := PFSessionByPlayerKeyPrefix + state.PlayerID.String()
	pipe.Set(ctx, playerIndexKey, state.SessionID.String(), PFSessionTTL)

	// Secondary index: pf_session_game_session:{game_session_id} -> session_id
	gameSessionIndexKey := PFSessionByGameSessionPrefix + state.GameSessionID.String()
	pipe.Set(ctx, gameSessionIndexKey, state.SessionID.String(), PFSessionTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set PF session state: %w", err)
	}

	c.logger.Debug().
		Str("session_id", state.SessionID.String()).
		Str("player_id", state.PlayerID.String()).
		Msg("PF session state stored in Redis")

	return nil
}

// GetSessionState retrieves PF session state by session ID
func (c *PFSessionCache) GetSessionState(ctx context.Context, sessionID uuid.UUID) (*provablyfair.PFSessionState, error) {
	if c.client == nil {
		return nil, provablyfair.ErrStateNotFound
	}

	key := PFSessionKeyPrefix + sessionID.String()
	val, err := c.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get PF session state: %w", err)
	}
	if val == "" {
		return nil, provablyfair.ErrStateNotFound
	}

	var state provablyfair.PFSessionState
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PF session state: %w", err)
	}

	return &state, nil
}

// GetSessionStateByPlayer retrieves PF session state by player ID
func (c *PFSessionCache) GetSessionStateByPlayer(ctx context.Context, playerID uuid.UUID) (*provablyfair.PFSessionState, error) {
	if c.client == nil {
		return nil, provablyfair.ErrStateNotFound
	}

	// Look up session ID from player index
	indexKey := PFSessionByPlayerKeyPrefix + playerID.String()
	sessionIDStr, err := c.client.Get(ctx, indexKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get PF session by player: %w", err)
	}
	if sessionIDStr == "" {
		return nil, provablyfair.ErrStateNotFound
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID in index: %w", err)
	}

	return c.GetSessionState(ctx, sessionID)
}

// GetSessionStateByGameSession retrieves PF session state by game session ID
func (c *PFSessionCache) GetSessionStateByGameSession(ctx context.Context, gameSessionID uuid.UUID) (*provablyfair.PFSessionState, error) {
	if c.client == nil {
		return nil, provablyfair.ErrStateNotFound
	}

	// Look up session ID from game session index
	indexKey := PFSessionByGameSessionPrefix + gameSessionID.String()
	sessionIDStr, err := c.client.Get(ctx, indexKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get PF session by game session: %w", err)
	}
	if sessionIDStr == "" {
		return nil, provablyfair.ErrStateNotFound
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID in index: %w", err)
	}

	return c.GetSessionState(ctx, sessionID)
}

// UpdateSessionState updates PF session state in Redis
func (c *PFSessionCache) UpdateSessionState(ctx context.Context, state *provablyfair.PFSessionState) error {
	state.UpdatedAt = time.Now().UTC()
	return c.SetSessionState(ctx, state)
}

// DeleteSessionState removes PF session state from Redis
func (c *PFSessionCache) DeleteSessionState(ctx context.Context, sessionID uuid.UUID) error {
	if c.client == nil {
		return nil // No-op if Redis is disabled
	}

	// First get the state to find the indexes to delete
	state, err := c.GetSessionState(ctx, sessionID)
	if err != nil {
		// If not found, nothing to delete
		if err == provablyfair.ErrStateNotFound {
			return nil
		}
		return err
	}

	// Use pipeline to delete all keys atomically
	pipe := c.client.GetClient().Pipeline()

	// Delete primary key
	primaryKey := PFSessionKeyPrefix + sessionID.String()
	pipe.Del(ctx, primaryKey)

	// Delete player index
	playerIndexKey := PFSessionByPlayerKeyPrefix + state.PlayerID.String()
	pipe.Del(ctx, playerIndexKey)

	// Delete game session index
	gameSessionIndexKey := PFSessionByGameSessionPrefix + state.GameSessionID.String()
	pipe.Del(ctx, gameSessionIndexKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete PF session state: %w", err)
	}

	c.logger.Debug().
		Str("session_id", sessionID.String()).
		Msg("PF session state deleted from Redis")

	return nil
}

// IncrementNonce atomically increments the nonce for a session
// Returns the new nonce value
func (c *PFSessionCache) IncrementNonce(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("redis client is not available")
	}

	// Get current state
	state, err := c.GetSessionState(ctx, sessionID)
	if err != nil {
		return 0, err
	}

	// Use Lua script for atomic increment
	script := redis.NewScript(`
		local key = KEYS[1]
		local data = redis.call('GET', key)
		if not data then
			return redis.error_reply('session not found')
		end

		local state = cjson.decode(data)
		state.nonce = state.nonce + 1
		state.updated_at = ARGV[1]

		local newData = cjson.encode(state)
		redis.call('SET', key, newData, 'EX', ARGV[2])

		return state.nonce
	`)

	primaryKey := PFSessionKeyPrefix + sessionID.String()
	updatedAt := time.Now().UTC().Format(time.RFC3339)
	ttlSeconds := int64(PFSessionTTL.Seconds())

	result, err := script.Run(ctx, c.client.GetClient(), []string{primaryKey}, updatedAt, ttlSeconds).Result()
	if err != nil {
		// Fallback to non-atomic update if Lua is not available
		state.Nonce++
		state.UpdatedAt = time.Now().UTC()
		if updateErr := c.SetSessionState(ctx, state); updateErr != nil {
			return 0, fmt.Errorf("failed to increment nonce: %w", updateErr)
		}
		return state.Nonce, nil
	}

	newNonce, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type from increment: %T", result)
	}

	return newNonce, nil
}
