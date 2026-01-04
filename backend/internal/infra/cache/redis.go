package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// RedisClient wraps Redis client
type RedisClient struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.Config, log *logger.Logger) (*RedisClient, error) {
	if !cfg.Redis.Enabled {
		log.Info().Msg("Redis is disabled, skipping connection")
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     10 * runtime.GOMAXPROCS(0), // Pool size = 10 * CPU cores
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().
		Str("addr", cfg.Redis.Addr).
		Msg("Redis connection established")

	return &RedisClient{
		client: client,
		logger: log,
	}, nil
}

// Get retrieves a value from Redis
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key does not exist
	}
	return val, err
}

// Set stores a value in Redis with expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Del deletes a key from Redis
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in Redis
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Incr increments a key's value
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Expire sets a timeout on a key
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	if r.client != nil {
		r.logger.Info().Msg("Closing Redis connection")
		return r.client.Close()
	}
	return nil
}

// GetClient returns the underlying Redis client
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Session cache key prefixes
const (
	SessionKeyPrefix       = "player_session:"
	PlayerSessionsSetKey   = "player_sessions_set:" // SET to track sessions by player ID
	ActivePlayersSetKey    = "active_players_set"   // SET to track all players with active sessions

	// Trial session cache key prefixes
	TrialSessionKeyPrefix     = "trial_session:"
	TrialGameSessionKeyPrefix = "trial_game_session:"
	TrialFreeSpinsKeyPrefix   = "trial_free_spins:"
)

// SessionData represents cached session data in Redis
type SessionData struct {
	SessionID string `json:"session_id"`
	PlayerID  string `json:"player_id"`
	GameID    string `json:"game_id,omitempty"` // empty string for cross-game
	ExpiresAt int64  `json:"expires_at"`        // Unix timestamp
}

// SetSession stores session data in Redis with secondary index for O(1) player lookup
func (r *RedisClient) SetSession(ctx context.Context, sessionToken string, data *SessionData, expiration time.Duration) error {
	key := SessionKeyPrefix + sessionToken
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Store session data
	pipe.Set(ctx, key, jsonData, expiration)

	// Add session token to player's session set (secondary index)
	playerSetKey := PlayerSessionsSetKey + data.PlayerID
	pipe.SAdd(ctx, playerSetKey, sessionToken)
	pipe.Expire(ctx, playerSetKey, expiration)

	// Add player to active players set
	pipe.SAdd(ctx, ActivePlayersSetKey, data.PlayerID)

	_, err = pipe.Exec(ctx)
	return err
}

// GetSession retrieves session data from Redis
func (r *RedisClient) GetSession(ctx context.Context, sessionToken string) (*SessionData, error) {
	key := SessionKeyPrefix + sessionToken
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Session not found in cache
	}
	if err != nil {
		return nil, err
	}

	var data SessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to parse session data: %w", err)
	}

	return &data, nil
}

// DeleteSession removes session data from Redis and cleans up secondary indexes
func (r *RedisClient) DeleteSession(ctx context.Context, sessionToken string) error {
	key := SessionKeyPrefix + sessionToken

	// Get session data first to know which player's set to update
	data, err := r.GetSession(ctx, sessionToken)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()

	// Delete the session
	pipe.Del(ctx, key)

	// Remove from player's session set if we have player info
	if data != nil {
		playerSetKey := PlayerSessionsSetKey + data.PlayerID
		pipe.SRem(ctx, playerSetKey, sessionToken)

		// Check if player has any remaining sessions, if not remove from active players
		// We'll do this check after the pipeline executes
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	// Clean up active players set if player has no more sessions
	if data != nil {
		playerSetKey := PlayerSessionsSetKey + data.PlayerID
		count, _ := r.client.SCard(ctx, playerSetKey).Result()
		if count == 0 {
			r.client.SRem(ctx, ActivePlayersSetKey, data.PlayerID)
			r.client.Del(ctx, playerSetKey) // Clean up empty set
		}
	}

	return nil
}

// DeletePlayerSessions removes all sessions for a player from Redis
// Uses O(n) where n = number of player's sessions, not total sessions
func (r *RedisClient) DeletePlayerSessions(ctx context.Context, playerID string) error {
	playerSetKey := PlayerSessionsSetKey + playerID

	// Get all session tokens for this player from the SET - O(n) for this player only
	sessionTokens, err := r.client.SMembers(ctx, playerSetKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get player sessions: %w", err)
	}

	if len(sessionTokens) == 0 {
		return nil
	}

	// Build keys to delete
	keysToDelete := make([]string, len(sessionTokens)+1)
	for i, token := range sessionTokens {
		keysToDelete[i] = SessionKeyPrefix + token
	}
	// Also delete the player's session set
	keysToDelete[len(sessionTokens)] = playerSetKey

	// Delete all session keys and the player's set in one call
	pipe := r.client.Pipeline()
	pipe.Del(ctx, keysToDelete...)
	pipe.SRem(ctx, ActivePlayersSetKey, playerID)
	_, err = pipe.Exec(ctx)

	return err
}

// SessionExists checks if a session exists in Redis
func (r *RedisClient) SessionExists(ctx context.Context, sessionToken string) (bool, error) {
	key := SessionKeyPrefix + sessionToken
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

// HasPlayerSession checks if a player has any active session in Redis - O(1) operation
func (r *RedisClient) HasPlayerSession(ctx context.Context, playerID string) (bool, error) {
	playerSetKey := PlayerSessionsSetKey + playerID
	count, err := r.client.SCard(ctx, playerSetKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check player sessions: %w", err)
	}
	return count > 0, nil
}

// GetPlayersWithActiveSessions returns a map of player IDs that have active sessions
// Uses O(n) where n = number of playerIDs to check, not total sessions
func (r *RedisClient) GetPlayersWithActiveSessions(ctx context.Context, playerIDs []string) (map[string]bool, error) {
	result := make(map[string]bool, len(playerIDs))

	if len(playerIDs) == 0 {
		return result, nil
	}

	// Use pipeline to batch check all players at once
	pipe := r.client.Pipeline()
	cmds := make([]*redis.IntCmd, len(playerIDs))

	for i, playerID := range playerIDs {
		playerSetKey := PlayerSessionsSetKey + playerID
		cmds[i] = pipe.SCard(ctx, playerSetKey)
		result[playerID] = false // Initialize as false
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to check player sessions: %w", err)
	}

	// Process results
	for i, cmd := range cmds {
		count, err := cmd.Result()
		if err == nil && count > 0 {
			result[playerIDs[i]] = true
		}
	}

	return result, nil
}

// TrialSessionData represents cached trial session data in Redis
type TrialSessionData struct {
	ID             string  `json:"id"`
	Balance        float64 `json:"balance"`
	GameID         string  `json:"game_id,omitempty"`
	TotalSpins     int     `json:"total_spins"`
	TotalWagered   float64 `json:"total_wagered"`
	TotalWon       float64 `json:"total_won"`
	CreatedAt      int64   `json:"created_at"`
	LastActivityAt int64   `json:"last_activity_at"`
	ExpiresAt      int64   `json:"expires_at"`
}

// TrialGameSessionData represents cached trial game session data in Redis
type TrialGameSessionData struct {
	ID              string  `json:"id"`
	TrialSessionID  string  `json:"trial_session_id"`
	BetAmount       float64 `json:"bet_amount"`
	StartingBalance float64 `json:"starting_balance"`
	TotalSpins      int     `json:"total_spins"`
	TotalWagered    float64 `json:"total_wagered"`
	TotalWon        float64 `json:"total_won"`
	CreatedAt       int64   `json:"created_at"`
}

// TrialFreeSpinsData represents cached trial free spins session in Redis
type TrialFreeSpinsData struct {
	ID              string  `json:"id"`
	TrialSessionID  string  `json:"trial_session_id"`
	GameSessionID   string  `json:"game_session_id"`
	TotalSpins      int     `json:"total_spins"`
	RemainingSpins  int     `json:"remaining_spins"`
	CompletedSpins  int     `json:"completed_spins"`
	LockedBetAmount float64 `json:"locked_bet_amount"`
	TotalWon        float64 `json:"total_won"`
	IsActive        bool    `json:"is_active"`
	CreatedAt       int64   `json:"created_at"`
}

// SetTrialSession stores trial session data in Redis with TTL
func (r *RedisClient) SetTrialSession(ctx context.Context, sessionToken string, data *TrialSessionData, expiration time.Duration) error {
	key := TrialSessionKeyPrefix + sessionToken
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal trial session data: %w", err)
	}
	return r.client.Set(ctx, key, jsonData, expiration).Err()
}

// GetTrialSession retrieves trial session data from Redis
func (r *RedisClient) GetTrialSession(ctx context.Context, sessionToken string) (*TrialSessionData, error) {
	key := TrialSessionKeyPrefix + sessionToken
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Session not found
	}
	if err != nil {
		return nil, err
	}

	var data TrialSessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to parse trial session data: %w", err)
	}

	return &data, nil
}

// UpdateTrialSession updates trial session data in Redis (preserves TTL)
func (r *RedisClient) UpdateTrialSession(ctx context.Context, sessionToken string, data *TrialSessionData) error {
	key := TrialSessionKeyPrefix + sessionToken

	// Get remaining TTL
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}
	if ttl <= 0 {
		return fmt.Errorf("trial session expired or not found")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal trial session data: %w", err)
	}

	return r.client.Set(ctx, key, jsonData, ttl).Err()
}

// DeleteTrialSession removes trial session from Redis
func (r *RedisClient) DeleteTrialSession(ctx context.Context, sessionToken string) error {
	key := TrialSessionKeyPrefix + sessionToken
	return r.client.Del(ctx, key).Err()
}

// SetTrialGameSession stores trial game session data in Redis
func (r *RedisClient) SetTrialGameSession(ctx context.Context, sessionID string, data *TrialGameSessionData, expiration time.Duration) error {
	key := TrialGameSessionKeyPrefix + sessionID
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal trial game session data: %w", err)
	}
	return r.client.Set(ctx, key, jsonData, expiration).Err()
}

// GetTrialGameSession retrieves trial game session data from Redis
func (r *RedisClient) GetTrialGameSession(ctx context.Context, sessionID string) (*TrialGameSessionData, error) {
	key := TrialGameSessionKeyPrefix + sessionID
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var data TrialGameSessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to parse trial game session data: %w", err)
	}

	return &data, nil
}

// DeleteTrialGameSession removes trial game session from Redis
func (r *RedisClient) DeleteTrialGameSession(ctx context.Context, sessionID string) error {
	key := TrialGameSessionKeyPrefix + sessionID
	return r.client.Del(ctx, key).Err()
}

// SetTrialFreeSpins stores trial free spins session in Redis
func (r *RedisClient) SetTrialFreeSpins(ctx context.Context, sessionID string, data *TrialFreeSpinsData, expiration time.Duration) error {
	key := TrialFreeSpinsKeyPrefix + sessionID
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal trial free spins data: %w", err)
	}
	return r.client.Set(ctx, key, jsonData, expiration).Err()
}

// GetTrialFreeSpins retrieves trial free spins session from Redis
func (r *RedisClient) GetTrialFreeSpins(ctx context.Context, sessionID string) (*TrialFreeSpinsData, error) {
	key := TrialFreeSpinsKeyPrefix + sessionID
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var data TrialFreeSpinsData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to parse trial free spins data: %w", err)
	}

	return &data, nil
}

// UpdateTrialFreeSpins updates trial free spins session in Redis
func (r *RedisClient) UpdateTrialFreeSpins(ctx context.Context, sessionID string, data *TrialFreeSpinsData) error {
	key := TrialFreeSpinsKeyPrefix + sessionID

	// Get remaining TTL
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}
	if ttl <= 0 {
		// Set a default TTL if not found (2 hours for trial sessions)
		ttl = 2 * time.Hour
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal trial free spins data: %w", err)
	}

	return r.client.Set(ctx, key, jsonData, ttl).Err()
}

// DeleteTrialFreeSpins removes trial free spins session from Redis
func (r *RedisClient) DeleteTrialFreeSpins(ctx context.Context, sessionID string) error {
	key := TrialFreeSpinsKeyPrefix + sessionID
	return r.client.Del(ctx, key).Err()
}

// GetActiveTrialFreeSpinsByTrialSession finds active free spins for a trial session
func (r *RedisClient) GetActiveTrialFreeSpinsByTrialSession(ctx context.Context, trialSessionID string) (*TrialFreeSpinsData, error) {
	// Scan for free spins keys matching this trial session
	// This is a simple approach - for production you might want a secondary index
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, TrialFreeSpinsKeyPrefix+"*", 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			val, err := r.client.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var data TrialFreeSpinsData
			if err := json.Unmarshal([]byte(val), &data); err != nil {
				continue
			}

			if data.TrialSessionID == trialSessionID && data.IsActive {
				return &data, nil
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil, nil
}
