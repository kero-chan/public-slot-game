package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/infra/cache"
	"github.com/slotmachine/backend/internal/pkg/errors"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

const (
	// Redis key prefixes for trial security
	trialIPSessionsKey          = "trial:ip_sessions:"          // ZSET of session IDs per IP (score = timestamp)
	trialIPCooldownKey          = "trial:ip_cooldown:"          // Cooldown timestamp per IP
	trialGlobalCounterKey       = "trial:global_counter"        // Global trial session counter
	trialMaxGlobalLimit         = 100000                        // Max global trial sessions (safety cap)
	trialFingerprintSessionsKey = "trial:fp_sessions:"          // ZSET of session IDs per fingerprint (IP+UA)
	trialMaxSessionsPerFP       = 3                             // Max sessions per fingerprint (IP+User-Agent)
	trialSessionMetaKey         = "trial:session_meta:"         // Hash storing session metadata (ip, fingerprint)
)

// Lua script for atomic session removal with idempotent counter decrement
// Only decrements global counter if session was actually removed from ZSET
// Returns: 1 if session was removed, 0 if session didn't exist
var luaRemoveSession = redis.NewScript(`
local ip_key = KEYS[1]
local fp_key = KEYS[2]
local meta_key = KEYS[3]
local session_key = KEYS[4]
local counter_key = KEYS[5]
local session_token = ARGV[1]

local removed = 0

-- Remove from IP ZSET
local ip_removed = redis.call('ZREM', ip_key, session_token)
if ip_removed > 0 then
    removed = 1
end

-- Remove from fingerprint ZSET (if fp_key provided)
if fp_key ~= "" then
    redis.call('ZREM', fp_key, session_token)
end

-- Delete session metadata
redis.call('DEL', meta_key)

-- Delete actual session data
redis.call('DEL', session_key)

-- Only decrement counter if we actually removed a session
if removed > 0 then
    redis.call('DECR', counter_key)
end

return removed
`)

// TrialRateLimiter implements security controls for trial session creation
type TrialRateLimiter struct {
	redis            *cache.RedisClient
	config           *config.TrialConfig
	logger           *logger.Logger
	whitelistedNets  []*net.IPNet    // Parsed CIDR networks for whitelist
	whitelistedIPs   map[string]bool // Exact IP matches for whitelist
	trustedProxyNets []*net.IPNet    // Parsed CIDR networks for trusted proxies
	trustedProxyIPs  map[string]bool // Exact IP matches for trusted proxies
}

// NewTrialRateLimiter creates a new trial rate limiter
func NewTrialRateLimiter(redis *cache.RedisClient, cfg *config.TrialConfig, log *logger.Logger) *TrialRateLimiter {
	trl := &TrialRateLimiter{
		redis:           redis,
		config:          cfg,
		logger:          log,
		whitelistedIPs:  make(map[string]bool),
		trustedProxyIPs: make(map[string]bool),
	}

	// Parse whitelisted IPs/CIDRs
	if cfg.WhitelistedIPs != "" {
		for _, entry := range strings.Split(cfg.WhitelistedIPs, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}

			// Try parsing as CIDR
			if strings.Contains(entry, "/") {
				_, network, err := net.ParseCIDR(entry)
				if err == nil {
					trl.whitelistedNets = append(trl.whitelistedNets, network)
					log.Debug().Str("cidr", entry).Msg("Added whitelisted CIDR for trial")
				} else {
					log.Warn().Str("entry", entry).Err(err).Msg("Invalid CIDR in trial whitelist")
				}
			} else {
				// Exact IP match
				if ip := net.ParseIP(entry); ip != nil {
					trl.whitelistedIPs[entry] = true
					log.Debug().Str("ip", entry).Msg("Added whitelisted IP for trial")
				} else {
					log.Warn().Str("entry", entry).Msg("Invalid IP in trial whitelist")
				}
			}
		}
	}

	// Parse trusted proxy IPs/CIDRs (HIGH-1 fix: IP spoofing prevention)
	if cfg.TrustedProxies != "" {
		for _, entry := range strings.Split(cfg.TrustedProxies, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}

			// Try parsing as CIDR
			if strings.Contains(entry, "/") {
				_, network, err := net.ParseCIDR(entry)
				if err == nil {
					trl.trustedProxyNets = append(trl.trustedProxyNets, network)
					log.Debug().Str("cidr", entry).Msg("Added trusted proxy CIDR")
				} else {
					log.Warn().Str("entry", entry).Err(err).Msg("Invalid CIDR in trusted proxies")
				}
			} else {
				// Exact IP match
				if ip := net.ParseIP(entry); ip != nil {
					trl.trustedProxyIPs[entry] = true
					log.Debug().Str("ip", entry).Msg("Added trusted proxy IP")
				} else {
					log.Warn().Str("entry", entry).Msg("Invalid IP in trusted proxies")
				}
			}
		}
	}

	log.Info().
		Int("whitelisted_cidrs", len(trl.whitelistedNets)).
		Int("whitelisted_ips", len(trl.whitelistedIPs)).
		Int("trusted_proxy_cidrs", len(trl.trustedProxyNets)).
		Int("trusted_proxy_ips", len(trl.trustedProxyIPs)).
		Msg("Trial rate limiter initialized")

	return trl
}

// isWhitelisted checks if an IP is in the whitelist
func (trl *TrialRateLimiter) isWhitelisted(ipStr string) bool {
	// Check exact match
	if trl.whitelistedIPs[ipStr] {
		return true
	}

	// Check CIDR ranges
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, network := range trl.whitelistedNets {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// isTrustedProxy checks if an IP is a trusted reverse proxy (HIGH-1 fix)
// Only trusted proxies can set X-Real-IP/X-Forwarded-For headers
func (trl *TrialRateLimiter) isTrustedProxy(ipStr string) bool {
	// Check exact match
	if trl.trustedProxyIPs[ipStr] {
		return true
	}

	// Check CIDR ranges
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, network := range trl.trustedProxyNets {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// getClientIP extracts the real client IP with trusted proxy validation (HIGH-1 fix)
// Only trusts X-Real-IP/X-Forwarded-For if request comes from a trusted proxy
func (trl *TrialRateLimiter) getClientIP(c *fiber.Ctx) string {
	// Get the direct connection IP
	directIP := c.IP()

	// If no trusted proxies configured, use direct IP only (most secure)
	if len(trl.trustedProxyNets) == 0 && len(trl.trustedProxyIPs) == 0 {
		return directIP
	}

	// Only trust forwarded headers if request comes from a trusted proxy
	if !trl.isTrustedProxy(directIP) {
		// Direct connection from untrusted source - ignore forwarded headers
		return directIP
	}

	// Request is from trusted proxy - check forwarded headers
	// Priority: X-Real-IP > X-Forwarded-For (first IP)
	if xRealIP := c.Get("X-Real-IP"); xRealIP != "" {
		// Validate IP format to prevent injection
		if ip := net.ParseIP(xRealIP); ip != nil {
			return xRealIP
		}
	}

	// Fallback to X-Forwarded-For (take first/leftmost IP = original client)
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if ip := net.ParseIP(clientIP); ip != nil {
				return clientIP
			}
		}
	}

	// No valid forwarded header - use direct IP
	return directIP
}

// GenerateDeviceFingerprint creates a fingerprint from IP + User-Agent
// Used for per-browser session limiting
func (trl *TrialRateLimiter) GenerateDeviceFingerprint(clientIP, userAgent string) string {
	data := fmt.Sprintf("%s|%s", clientIP, userAgent)
	hash := sha256.Sum256([]byte(data))
	return "fp_" + hex.EncodeToString(hash[:16]) // 32 hex chars
}

// TrialCreationMiddleware validates trial session creation requests
// Enforces:
// - Cooldown period between session creations from same IP (skipped for whitelisted IPs)
// - Maximum concurrent sessions per IP (higher limit for whitelisted IPs)
// - Global session limit (safety cap)
func (trl *TrialRateLimiter) TrialCreationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := trl.logger.WithTrace(c)

		// Check if trial mode is enabled
		if !trl.config.Enabled {
			return respondError(c, errors.ServiceUnavailable("Trial mode is currently disabled"))
		}

		// If Redis is not available, deny trial (security-first approach)
		if trl.redis == nil {
			log.Warn().Msg("Trial creation denied: Redis not available")
			return respondError(c, errors.ServiceUnavailable("Trial mode temporarily unavailable"))
		}

		// Extract client IP with trusted proxy validation (HIGH-1 fix: prevents IP spoofing)
		clientIP := trl.getClientIP(c)

		ctx := c.Context()

		// Check if IP is whitelisted (internal testing, office networks)
		isWhitelisted := trl.isWhitelisted(clientIP)

		// Check 1: Global session limit (prevent Redis memory exhaustion) - applies to all
		globalCount, err := trl.getGlobalSessionCount(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check global trial session count")
			// Fail open for read errors, but log for monitoring
		} else if globalCount >= trialMaxGlobalLimit {
			log.Warn().
				Int64("global_count", globalCount).
				Str("ip", clientIP).
				Bool("whitelisted", isWhitelisted).
				Msg("Trial creation denied: global session limit reached")
			return respondError(c, errors.TooManyRequests("Trial service is at capacity. Please try again later."))
		}

		// Check 2: IP cooldown (prevent rapid session creation) - SKIP for whitelisted IPs
		if !isWhitelisted {
			inCooldown, remainingSeconds, err := trl.checkIPCooldown(ctx, clientIP)
			if err != nil {
				log.Error().Err(err).Str("ip", clientIP).Msg("Failed to check IP cooldown")
			}
			if inCooldown {
				log.Warn().
					Str("ip", clientIP).
					Int64("remaining_seconds", remainingSeconds).
					Msg("Trial creation denied: IP in cooldown period")
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"success": false,
					"error": fiber.Map{
						"code":              "TRIAL_COOLDOWN",
						"message":           fmt.Sprintf("Please wait %d seconds before creating another trial session", remainingSeconds),
						"retry_after_secs":  remainingSeconds,
					},
				})
			}
		}

		// Check 3: Max sessions per IP - use higher limit for whitelisted IPs
		maxSessions := trl.config.MaxSessionsPerIP
		if isWhitelisted && trl.config.WhitelistedMaxSessions > 0 {
			maxSessions = trl.config.WhitelistedMaxSessions
		}

		sessionCount, err := trl.getIPSessionCount(ctx, clientIP)
		if err != nil {
			log.Error().Err(err).Str("ip", clientIP).Msg("Failed to check IP session count")
		}
		if sessionCount >= int64(maxSessions) {
			log.Warn().
				Str("ip", clientIP).
				Int64("session_count", sessionCount).
				Int("max_allowed", maxSessions).
				Bool("whitelisted", isWhitelisted).
				Msg("Trial creation denied: max sessions per IP reached")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":         "TRIAL_LIMIT_REACHED",
					"message":      fmt.Sprintf("Maximum %d concurrent trial sessions allowed per IP", maxSessions),
					"current":      sessionCount,
					"max_allowed":  maxSessions,
				},
			})
		}

		// Check 4: Max sessions per fingerprint (IP + User-Agent)
		// This limits sessions per browser type, allowing different browsers on same machine
		userAgent := c.Get("User-Agent")
		fingerprint := trl.GenerateDeviceFingerprint(clientIP, userAgent)

		fpSessionCount, err := trl.getFingerprintSessionCount(ctx, fingerprint)
		if err != nil {
			log.Error().Err(err).Str("fingerprint", fingerprint).Msg("Failed to check fingerprint session count")
		}

		// If at limit, remove oldest session to make room for new one (FIFO)
		if fpSessionCount >= trialMaxSessionsPerFP {
			oldestToken, err := trl.removeOldestFingerprintSession(ctx, fingerprint, clientIP)
			if err != nil {
				log.Warn().Err(err).Str("fingerprint", fingerprint).Msg("Failed to remove oldest session")
			} else if oldestToken != "" {
				log.Info().
					Str("ip", clientIP).
					Str("fingerprint", fingerprint).
					Str("removed_session", oldestToken[:minInt(20, len(oldestToken))]+"...").
					Msg("Removed oldest session to make room for new one")
			}
		}

		// Store IP and whitelist status in context
		c.Locals("client_ip", clientIP)
		c.Locals("is_whitelisted", isWhitelisted)

		if isWhitelisted {
			log.Debug().
				Str("ip", clientIP).
				Int64("ip_sessions", sessionCount).
				Int64("fp_sessions", fpSessionCount).
				Msg("Whitelisted IP creating trial session")
		}

		return c.Next()
	}
}

// RegisterTrialSession records a new trial session for IP and fingerprint tracking
// Called by TrialHandler after successful session creation
// fingerprint: IP+User-Agent hash for per-browser limiting
// Uses ZSET with timestamp as score to track session age (for FIFO removal)
// Stores session metadata (IP, fingerprint) for proper cleanup
func (trl *TrialRateLimiter) RegisterTrialSession(ctx context.Context, clientIP, sessionToken, fingerprint string, sessionTTL time.Duration) error {
	if trl.redis == nil {
		return nil
	}

	isWhitelisted := trl.isWhitelisted(clientIP)
	now := float64(time.Now().Unix())

	pipe := trl.redis.GetClient().Pipeline()

	// Add session to IP's session ZSET (sorted by timestamp)
	ipSessionsKey := trialIPSessionsKey + clientIP
	pipe.ZAdd(ctx, ipSessionsKey, redis.Z{Score: now, Member: sessionToken})
	pipe.Expire(ctx, ipSessionsKey, sessionTTL+time.Hour)

	// Add session to fingerprint's session ZSET (sorted by timestamp)
	if fingerprint != "" {
		fpSessionsKey := trialFingerprintSessionsKey + fingerprint
		pipe.ZAdd(ctx, fpSessionsKey, redis.Z{Score: now, Member: sessionToken})
		pipe.Expire(ctx, fpSessionsKey, sessionTTL+time.Hour)
	}

	// Store session metadata for proper cleanup (CRITICAL-2 fix: enables fingerprint cleanup)
	metaKey := trialSessionMetaKey + sessionToken
	pipe.HSet(ctx, metaKey, map[string]interface{}{
		"ip":          clientIP,
		"fingerprint": fingerprint,
		"created_at":  time.Now().Unix(),
	})
	pipe.Expire(ctx, metaKey, sessionTTL+time.Hour)

	// Set IP cooldown - SKIP for whitelisted IPs
	if !isWhitelisted {
		cooldownKey := trialIPCooldownKey + clientIP
		cooldownDuration := time.Duration(trl.config.SessionCooldownSeconds) * time.Second
		pipe.Set(ctx, cooldownKey, time.Now().Unix(), cooldownDuration)
	}

	// Increment global counter
	pipe.Incr(ctx, trialGlobalCounterKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to register trial session: %w", err)
	}

	trl.logger.Info().
		Str("ip", clientIP).
		Str("session_prefix", sessionToken[:minInt(20, len(sessionToken))]+"...").
		Str("fingerprint", fingerprint).
		Bool("whitelisted", isWhitelisted).
		Msg("Trial session registered")

	return nil
}

// minInt returns the smaller of two integers (helper for Go < 1.21)
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getFingerprintSessionCount returns the number of active sessions for a fingerprint
// Uses ZSET for ordered session tracking (by timestamp)
func (trl *TrialRateLimiter) getFingerprintSessionCount(ctx context.Context, fingerprint string) (int64, error) {
	fpSessionsKey := trialFingerprintSessionsKey + fingerprint
	count, err := trl.redis.GetClient().ZCard(ctx, fpSessionsKey).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// removeOldestFingerprintSession removes the oldest session for a fingerprint (FIFO)
// Uses Lua script for atomic removal with idempotent counter decrement (CRITICAL-1 fix)
// Returns the removed session token, or empty string if no session was removed
func (trl *TrialRateLimiter) removeOldestFingerprintSession(ctx context.Context, fingerprint, clientIP string) (string, error) {
	fpSessionsKey := trialFingerprintSessionsKey + fingerprint

	// Get the oldest session (lowest score = earliest timestamp)
	sessions, err := trl.redis.GetClient().ZRange(ctx, fpSessionsKey, 0, 0).Result()
	if err != nil {
		return "", err
	}
	if len(sessions) == 0 {
		return "", nil
	}

	oldestToken := sessions[0]

	// Look up session metadata to get correct IP (may differ from current request IP)
	metaKey := trialSessionMetaKey + oldestToken
	meta, err := trl.redis.GetClient().HGetAll(ctx, metaKey).Result()
	if err != nil {
		return "", err
	}

	// Use stored IP if available, otherwise use provided clientIP
	sessionIP := clientIP
	if storedIP, ok := meta["ip"]; ok && storedIP != "" {
		sessionIP = storedIP
	}

	ipSessionsKey := trialIPSessionsKey + sessionIP
	sessionKey := "trial_session:" + oldestToken

	// Use Lua script for atomic removal (CRITICAL-1 fix: prevents double decrement)
	_, err = luaRemoveSession.Run(ctx, trl.redis.GetClient(),
		[]string{ipSessionsKey, fpSessionsKey, metaKey, sessionKey, trialGlobalCounterKey},
		oldestToken,
	).Result()

	if err != nil {
		return "", err
	}

	return oldestToken, nil
}

// UnregisterTrialSession removes a trial session from IP and fingerprint tracking
// Called when a trial session expires or is manually ended
// Uses Lua script for atomic removal with idempotent counter decrement (CRITICAL-1 fix)
// Automatically looks up fingerprint from session metadata (CRITICAL-2 fix)
func (trl *TrialRateLimiter) UnregisterTrialSession(ctx context.Context, clientIP, sessionToken string) error {
	if trl.redis == nil {
		return nil
	}

	// Look up session metadata to get fingerprint
	metaKey := trialSessionMetaKey + sessionToken
	meta, err := trl.redis.GetClient().HGetAll(ctx, metaKey).Result()
	if err != nil {
		// Log but continue - we can still remove from IP tracking
		trl.logger.Warn().Err(err).Str("session", sessionToken).Msg("Failed to get session metadata")
	}

	// Get fingerprint from metadata (CRITICAL-2 fix: enables fingerprint cleanup)
	fingerprint := ""
	if fp, ok := meta["fingerprint"]; ok {
		fingerprint = fp
	}

	// Use stored IP if available (in case provided IP is different)
	sessionIP := clientIP
	if storedIP, ok := meta["ip"]; ok && storedIP != "" {
		sessionIP = storedIP
	}

	ipSessionsKey := trialIPSessionsKey + sessionIP
	fpSessionsKey := ""
	if fingerprint != "" {
		fpSessionsKey = trialFingerprintSessionsKey + fingerprint
	}
	sessionKey := "trial_session:" + sessionToken

	// Use Lua script for atomic removal (CRITICAL-1 fix: prevents double decrement)
	_, err = luaRemoveSession.Run(ctx, trl.redis.GetClient(),
		[]string{ipSessionsKey, fpSessionsKey, metaKey, sessionKey, trialGlobalCounterKey},
		sessionToken,
	).Result()

	return err
}

// CleanupExpiredSessions removes expired sessions from IP tracking
// Should be called periodically or on session validation failure
// Uses Lua script for atomic removal with idempotent counter decrement (CRITICAL-1 fix)
// Also cleans up from fingerprint tracking (CRITICAL-2 fix)
func (trl *TrialRateLimiter) CleanupExpiredSessions(ctx context.Context, clientIP string, validSessionTokens []string) error {
	if trl.redis == nil {
		return nil
	}

	ipSessionsKey := trialIPSessionsKey + clientIP

	// Get all sessions currently tracked for this IP (using ZSET)
	trackedSessions, err := trl.redis.GetClient().ZRange(ctx, ipSessionsKey, 0, -1).Result()
	if err != nil {
		return err
	}

	// Create a set of valid sessions for quick lookup
	validSet := make(map[string]bool)
	for _, token := range validSessionTokens {
		validSet[token] = true
	}

	// Remove expired sessions from tracking
	var toRemove []string
	for _, session := range trackedSessions {
		if !validSet[session] {
			toRemove = append(toRemove, session)
		}
	}

	if len(toRemove) > 0 {
		removedCount := 0
		for _, sessionToken := range toRemove {
			// Look up session metadata to get fingerprint
			metaKey := trialSessionMetaKey + sessionToken
			meta, _ := trl.redis.GetClient().HGetAll(ctx, metaKey).Result()

			fingerprint := ""
			if fp, ok := meta["fingerprint"]; ok {
				fingerprint = fp
			}

			fpSessionsKey := ""
			if fingerprint != "" {
				fpSessionsKey = trialFingerprintSessionsKey + fingerprint
			}
			sessionKey := "trial_session:" + sessionToken

			// Use Lua script for atomic removal (CRITICAL-1 fix: prevents double decrement)
			result, err := luaRemoveSession.Run(ctx, trl.redis.GetClient(),
				[]string{ipSessionsKey, fpSessionsKey, metaKey, sessionKey, trialGlobalCounterKey},
				sessionToken,
			).Int()

			if err == nil && result > 0 {
				removedCount++
			}
		}

		if removedCount > 0 {
			trl.logger.Info().
				Str("ip", clientIP).
				Int("removed_count", removedCount).
				Msg("Cleaned up expired trial sessions from IP tracking")
		}
	}

	return nil
}

// Helper methods

func (trl *TrialRateLimiter) checkIPCooldown(ctx context.Context, clientIP string) (bool, int64, error) {
	cooldownKey := trialIPCooldownKey + clientIP
	ttl, err := trl.redis.GetClient().TTL(ctx, cooldownKey).Result()
	if err != nil {
		return false, 0, err
	}
	if ttl > 0 {
		return true, int64(ttl.Seconds()), nil
	}
	return false, 0, nil
}

func (trl *TrialRateLimiter) getIPSessionCount(ctx context.Context, clientIP string) (int64, error) {
	ipSessionsKey := trialIPSessionsKey + clientIP
	count, err := trl.redis.GetClient().ZCard(ctx, ipSessionsKey).Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (trl *TrialRateLimiter) getGlobalSessionCount(ctx context.Context) (int64, error) {
	count, err := trl.redis.GetClient().Get(ctx, trialGlobalCounterKey).Int64()
	if err != nil {
		// Key doesn't exist yet, return 0
		return 0, nil
	}
	return count, nil
}

// GetTrialStats returns current trial session statistics (for admin/monitoring)
func (trl *TrialRateLimiter) GetTrialStats(ctx context.Context) (map[string]interface{}, error) {
	if trl.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}

	globalCount, _ := trl.getGlobalSessionCount(ctx)

	return map[string]interface{}{
		"global_session_count": globalCount,
		"max_global_limit":     trialMaxGlobalLimit,
		"max_per_ip":           trl.config.MaxSessionsPerIP,
		"cooldown_seconds":     trl.config.SessionCooldownSeconds,
		"enabled":              trl.config.Enabled,
	}, nil
}
