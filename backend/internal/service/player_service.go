package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/game"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/infra/cache"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/pkg/util"
)

// PlayerService implements the player.Service interface
type PlayerService struct {
	repo        player.Repository
	gameRepo    game.Repository
	sessionRepo session.PlayerSessionRepository
	cache       *cache.RedisClient
	config      *config.Config
	logger      *logger.Logger
}

// NewPlayerService creates a new player service
func NewPlayerService(
	repo player.Repository,
	gameRepo game.Repository,
	sessionRepo session.PlayerSessionRepository,
	cache *cache.RedisClient,
	cfg *config.Config,
	log *logger.Logger,
) player.Service {
	return &PlayerService{
		repo:        repo,
		gameRepo:    gameRepo,
		sessionRepo: sessionRepo,
		cache:       cache,
		config:      cfg,
		logger:      log,
	}
}

// generateSessionToken generates a cryptographically secure session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// Register creates a new player account
// gameID is REQUIRED for player registration from frontend
// Only admin can create cross-game accounts (gameID = nil)
func (s *PlayerService) Register(ctx context.Context, username, email, password string, gameID *uuid.UUID) (*player.Player, error) {
	log := s.logger.WithTraceContext(ctx)

	// Validate inputs
	if err := s.validateRegistration(username, email, password); err != nil {
		return nil, err
	}

	// Trim whitespace but preserve original case (DB index handles case-insensitivity)
	trimmedUsername := strings.TrimSpace(username)
	trimmedEmail := strings.TrimSpace(email)

	// game_id is required for player registration
	if gameID == nil {
		return nil, player.ErrGameIDRequired
	}

	// Validate the game exists
	if err := s.validateGameExists(ctx, *gameID); err != nil {
		return nil, err
	}

	// Check if username already exists for this game (or cross-game)
	// Repository uses LOWER() for case-insensitive comparison
	existingUser, _ := s.repo.GetByUsernameAndGame(ctx, trimmedUsername, gameID)
	if existingUser != nil {
		return nil, player.ErrPlayerAlreadyExists
	}

	// Check if email already exists for this game (or cross-game)
	// Repository uses LOWER() for case-insensitive comparison
	existingEmail, _ := s.repo.GetByEmailAndGame(ctx, trimmedEmail, gameID)
	if existingEmail != nil {
		return nil, player.ErrPlayerAlreadyExists
	}

	// Hash password
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create player with original case preserved
	newPlayer := &player.Player{
		ID:           uuid.New(),
		Username:     trimmedUsername,
		Email:        trimmedEmail,
		PasswordHash: hashedPassword,
		GameID:       gameID,    // Set game_id (can be nil for cross-game account)
		Balance:      100000.00, // Default starting balance
		TotalSpins:   0,
		TotalWagered: 0.0,
		TotalWon:     0.0,
		IsActive:     true,
		IsVerified:   false,
	}

	// Save to database
	if err := s.repo.Create(ctx, newPlayer); err != nil {
		log.Error().Err(err).Str("username", trimmedUsername).Msg("Failed to create player")
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	log.Info().
		Str("player_id", newPlayer.ID.String()).
		Str("username", trimmedUsername).
		Interface("game_id", gameID).
		Msg("Player registered successfully")

	return newPlayer, nil
}

// Login authenticates a player and creates a session
func (s *PlayerService) Login(ctx context.Context, username, password string, gameID *uuid.UUID, opts *player.LoginOptions) (*player.LoginResult, error) {
	log := s.logger.WithTraceContext(ctx)

	// Validate inputs
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	// Default options
	if opts == nil {
		opts = &player.LoginOptions{}
	}

	// Find a player who can login (game-specific or cross-game account)
	p, err := s.repo.FindLoginCandidate(ctx, username, gameID)
	if err != nil {
		log.Warn().Str("username", username).Msg("Login attempt with non-existent username")
		return nil, player.ErrInvalidCredentials
	}

	// Validate game access
	if err := s.validateGameAccess(p, gameID); err != nil {
		log.Warn().
			Str("username", username).
			Interface("player_game_id", p.GameID).
			Interface("requested_game_id", gameID).
			Msg("Game access denied")
		return nil, err
	}

	// Check if player is active
	if !p.IsActive {
		return nil, fmt.Errorf("player account is not active")
	}

	// Verify password
	if !util.CheckPassword(password, p.PasswordHash) {
		log.Warn().Str("username", username).Msg("Login attempt with invalid password")
		return nil, player.ErrInvalidCredentials
	}

	// Check for existing active session
	existingSession, err := s.sessionRepo.GetActiveByPlayerAndGame(ctx, p.ID, gameID)
	if err == nil && existingSession != nil {
		// Player has an active session
		if !opts.ForceLogout {
			// Return error - player must explicitly force logout
			log.Warn().
				Str("player_id", p.ID.String()).
				Str("username", username).
				Msg("Player already logged in on another device")
			return nil, session.ErrPlayerAlreadyLoggedIn
		}

		// Force logout the existing session
		if err := s.deactivateSession(ctx, existingSession, session.LogoutReasonForced); err != nil {
			log.Error().Err(err).Str("session_id", existingSession.ID.String()).Msg("Failed to force logout existing session")
			// Continue with login anyway
		} else {
			log.Info().
				Str("player_id", p.ID.String()).
				Str("old_session_id", existingSession.ID.String()).
				Msg("Force logged out existing session")
		}
	}

	// Generate new session token
	sessionToken, err := generateSessionToken()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate session token")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Calculate expiration time
	expirationHours := s.config.JWT.ExpirationHours
	if expirationHours <= 0 {
		expirationHours = 24
	}
	expiresAt := time.Now().UTC().Add(time.Duration(expirationHours) * time.Hour)

	// Create player session
	var ipAddr, userAgent, deviceInfo *string
	if opts.IPAddress != "" {
		ipAddr = &opts.IPAddress
	}
	if opts.UserAgent != "" {
		userAgent = &opts.UserAgent
	}
	if opts.DeviceInfo != "" {
		deviceInfo = &opts.DeviceInfo
	}

	newSession := &session.PlayerSession{
		ID:             uuid.New(),
		PlayerID:       p.ID,
		GameID:         gameID,
		SessionToken:   sessionToken,
		DeviceInfo:     deviceInfo,
		IPAddress:      ipAddr,
		UserAgent:      userAgent,
		IsActive:       true,
		CreatedAt:      time.Now().UTC(),
		ExpiresAt:      expiresAt,
		LastActivityAt: time.Now().UTC(),
	}

	// Save session to database
	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		log.Error().Err(err).Str("player_id", p.ID.String()).Msg("Failed to create player session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Cache session in Redis for fast validation
	if s.cache != nil {
		gameIDStr := ""
		if gameID != nil {
			gameIDStr = gameID.String()
		}
		sessionData := &cache.SessionData{
			SessionID: newSession.ID.String(),
			PlayerID:  p.ID.String(),
			GameID:    gameIDStr,
			ExpiresAt: expiresAt.Unix(),
		}
		if err := s.cache.SetSession(ctx, sessionToken, sessionData, time.Duration(expirationHours)*time.Hour); err != nil {
			log.Warn().Err(err).Msg("Failed to cache session in Redis, falling back to DB validation")
			// Don't fail login - DB validation will still work
		}
	}

	// Update last login timestamp
	if err := s.repo.UpdateLastLogin(ctx, p.ID); err != nil {
		log.Error().Err(err).Str("player_id", p.ID.String()).Msg("Failed to update last login")
		// Don't fail the login for this
	}

	log.Info().
		Str("player_id", p.ID.String()).
		Str("session_id", newSession.ID.String()).
		Str("username", username).
		Interface("game_id", gameID).
		Bool("force_logout", opts.ForceLogout).
		Msg("Player logged in successfully")

	return &player.LoginResult{
		Player:       p,
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

// Logout invalidates the player's session
func (s *PlayerService) Logout(ctx context.Context, sessionToken string) error {
	log := s.logger.WithTraceContext(ctx)

	if sessionToken == "" {
		return fmt.Errorf("session token is required")
	}

	// Get session from database
	sess, err := s.sessionRepo.GetByToken(ctx, sessionToken)
	if err != nil {
		if errors.Is(err, session.ErrPlayerSessionNotFound) {
			// Session already logged out or doesn't exist
			return nil
		}
		log.Error().Err(err).Msg("Failed to get session for logout")
		return fmt.Errorf("failed to logout: %w", err)
	}

	// Deactivate session
	if err := s.deactivateSession(ctx, sess, session.LogoutReasonManual); err != nil {
		log.Error().Err(err).Str("session_id", sess.ID.String()).Msg("Failed to deactivate session")
		return fmt.Errorf("failed to logout: %w", err)
	}

	log.Info().
		Str("player_id", sess.PlayerID.String()).
		Str("session_id", sess.ID.String()).
		Msg("Player logged out successfully")

	return nil
}

// ValidateSession validates a session token and returns session info
func (s *PlayerService) ValidateSession(ctx context.Context, sessionToken string, requestedGameID *uuid.UUID) (*player.LoginResult, error) {
	log := s.logger.WithTraceContext(ctx)

	if sessionToken == "" {
		return nil, session.ErrPlayerSessionNotFound
	}

	// Try to get session from Redis cache first
	if s.cache != nil {
		cachedSession, err := s.cache.GetSession(ctx, sessionToken)
		if err != nil {
			log.Warn().Err(err).Msg("Redis cache error, falling back to DB")
		} else if cachedSession != nil {
			// Validate expiration
			if time.Now().Unix() > cachedSession.ExpiresAt {
				// Session expired - remove from cache
				_ = s.cache.DeleteSession(ctx, sessionToken)
				return nil, session.ErrPlayerSessionExpired
			}

			// Validate game access
			if err := s.validateSessionGameAccess(cachedSession.GameID, requestedGameID); err != nil {
				return nil, err
			}

			// Parse player ID
			playerID, err := uuid.Parse(cachedSession.PlayerID)
			if err != nil {
				log.Error().Err(err).Msg("Invalid player ID in cached session")
				// Fall through to DB validation
			} else {
				// Get player from database
				p, err := s.repo.GetByID(ctx, playerID)
				if err != nil {
					log.Warn().Err(err).Str("player_id", cachedSession.PlayerID).Msg("Failed to get player for cached session")
					// Fall through to DB validation
				} else {
					return &player.LoginResult{
						Player:       p,
						SessionToken: sessionToken,
						ExpiresAt:    cachedSession.ExpiresAt,
					}, nil
				}
			}
		}
	}

	// Fallback to database validation
	sess, err := s.sessionRepo.GetByToken(ctx, sessionToken)
	if err != nil {
		if errors.Is(err, session.ErrPlayerSessionNotFound) {
			return nil, session.ErrPlayerSessionNotFound
		}
		log.Error().Err(err).Msg("Failed to get session from database")
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	// Check if session is active
	if !sess.IsActive {
		if sess.LogoutReason != nil && *sess.LogoutReason == session.LogoutReasonForced {
			return nil, session.ErrPlayerSessionForcedLogout
		} else {
			return nil, session.ErrPlayerSessionInactive
		}
	}

	// Check if session is expired
	if time.Now().UTC().After(sess.ExpiresAt) {
		// Mark session as expired in database
		_ = s.sessionRepo.DeactivateSession(ctx, sess.ID, session.LogoutReasonExpired)
		return nil, session.ErrPlayerSessionExpired
	}

	// Validate game access
	sessionGameID := ""
	if sess.GameID != nil {
		sessionGameID = sess.GameID.String()
	}
	if err := s.validateSessionGameAccess(sessionGameID, requestedGameID); err != nil {
		return nil, err
	}

	// Get player
	p, err := s.repo.GetByID(ctx, sess.PlayerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", sess.PlayerID.String()).Msg("Failed to get player for session")
		return nil, player.ErrPlayerNotFound
	}

	// Cache session in Redis for faster subsequent validations
	if s.cache != nil {
		gameIDStr := ""
		if sess.GameID != nil {
			gameIDStr = sess.GameID.String()
		}
		sessionData := &cache.SessionData{
			SessionID: sess.ID.String(),
			PlayerID:  sess.PlayerID.String(),
			GameID:    gameIDStr,
			ExpiresAt: sess.ExpiresAt.Unix(),
		}
		// Calculate remaining TTL
		ttl := time.Until(sess.ExpiresAt)
		if ttl > 0 {
			if err := s.cache.SetSession(ctx, sessionToken, sessionData, ttl); err != nil {
				log.Warn().Err(err).Msg("Failed to cache session in Redis after DB validation")
				// Don't fail - DB validation succeeded
			}
		}
	}

	// Update last activity (async, don't block)
	go func() {
		bgCtx := context.Background()
		if err := s.sessionRepo.UpdateLastActivity(bgCtx, sess.ID); err != nil {
			log.Warn().Err(err).Str("session_id", sess.ID.String()).Msg("Failed to update last activity")
		}
	}()

	return &player.LoginResult{
		Player:       p,
		SessionToken: sessionToken,
		ExpiresAt:    sess.ExpiresAt.Unix(),
	}, nil
}

// deactivateSession deactivates a session in both database and cache
func (s *PlayerService) deactivateSession(ctx context.Context, sess *session.PlayerSession, reason string) error {
	// Deactivate in database
	if err := s.sessionRepo.DeactivateSession(ctx, sess.ID, reason); err != nil {
		return err
	}

	// Remove from Redis cache
	if s.cache != nil {
		if err := s.cache.DeleteSession(ctx, sess.SessionToken); err != nil {
			s.logger.Warn().Err(err).Str("session_token", sess.SessionToken).Msg("Failed to remove session from cache")
			// Don't fail - session is already deactivated in DB
		}
	}

	return nil
}

// validateSessionGameAccess validates that the session's game matches the requested game
func (s *PlayerService) validateSessionGameAccess(sessionGameID string, requestedGameID *uuid.UUID) error {
	// Cross-game sessions (sessionGameID == "") can access any game
	if sessionGameID == "" {
		return nil
	}

	// If no game requested, session is valid
	if requestedGameID == nil {
		return nil
	}

	// Session is game-specific - must match requested game
	if sessionGameID != requestedGameID.String() {
		return session.ErrPlayerSessionGameMismatch
	}

	return nil
}

// GetProfile retrieves a player's profile
func (s *PlayerService) GetProfile(ctx context.Context, playerID uuid.UUID) (*player.Player, error) {
	log := s.logger.WithTraceContext(ctx)

	p, err := s.repo.GetByID(ctx, playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player profile")
		return nil, player.ErrPlayerNotFound
	}

	return p, nil
}

// GetBalance retrieves a player's current balance
func (s *PlayerService) GetBalance(ctx context.Context, playerID uuid.UUID) (float64, error) {
	log := s.logger.WithTraceContext(ctx)

	p, err := s.repo.GetByID(ctx, playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player balance")
		return 0, player.ErrPlayerNotFound
	}

	return p.Balance, nil
}

// UpdateBalance updates a player's balance (for internal use)
func (s *PlayerService) UpdateBalance(ctx context.Context, playerID uuid.UUID, newBalance float64) error {
	log := s.logger.WithTraceContext(ctx)

	// Validate balance
	if newBalance < 0 {
		return fmt.Errorf("balance cannot be negative")
	}

	if err := s.repo.UpdateBalance(ctx, playerID, newBalance); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Float64("new_balance", newBalance).Msg("Failed to update balance")
		return fmt.Errorf("failed to update balance: %w", err)
	}

	log.Info().Str("player_id", playerID.String()).Float64("new_balance", newBalance).Msg("Balance updated")

	return nil
}

// DeductBet deducts bet amount from player balance
func (s *PlayerService) DeductBet(ctx context.Context, playerID uuid.UUID, betAmount float64) error {
	log := s.logger.WithTraceContext(ctx)

	// Validate bet amount
	if betAmount <= 0 {
		return fmt.Errorf("bet amount must be positive")
	}

	// Get current balance
	p, err := s.repo.GetByID(ctx, playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player for bet deduction")
		return player.ErrPlayerNotFound
	}

	// Check if player has sufficient balance
	if p.Balance < betAmount {
		log.Warn().
			Str("player_id", playerID.String()).
			Float64("balance", p.Balance).
			Float64("bet_amount", betAmount).
			Msg("Insufficient balance for bet")
		return player.ErrInsufficientBalance
	}

	// Calculate new balance
	newBalance := p.Balance - betAmount

	// Update balance
	if err := s.repo.UpdateBalance(ctx, playerID, newBalance); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to deduct bet")
		return fmt.Errorf("failed to deduct bet: %w", err)
	}

	log.Info().
		Str("player_id", playerID.String()).
		Float64("bet_amount", betAmount).
		Float64("new_balance", newBalance).
		Msg("Bet deducted from balance")

	return nil
}

// CreditWin credits win amount to player balance
func (s *PlayerService) CreditWin(ctx context.Context, playerID uuid.UUID, winAmount float64) error {
	log := s.logger.WithTraceContext(ctx)

	// Validate win amount
	if winAmount < 0 {
		return fmt.Errorf("win amount cannot be negative")
	}

	// If win is 0, skip (no need to update)
	if winAmount == 0 {
		return nil
	}

	// Get current balance
	p, err := s.repo.GetByID(ctx, playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player for win credit")
		return player.ErrPlayerNotFound
	}

	// Calculate new balance
	newBalance := p.Balance + winAmount

	// Update balance
	if err := s.repo.UpdateBalance(ctx, playerID, newBalance); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to credit win")
		return fmt.Errorf("failed to credit win: %w", err)
	}

	log.Info().
		Str("player_id", playerID.String()).
		Float64("win_amount", winAmount).
		Float64("new_balance", newBalance).
		Msg("Win credited to balance")

	return nil
}

// validateRegistration validates registration inputs
func (s *PlayerService) validateRegistration(username, email, password string) error {
	// Validate username
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}

	// Validate email
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email format")
	}

	// Validate password
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	return nil
}

// validateGameExists checks if the game exists
func (s *PlayerService) validateGameExists(ctx context.Context, gameID uuid.UUID) error {
	_, err := s.gameRepo.GetGameByID(ctx, gameID)
	if err != nil {
		return player.ErrGameNotFound
	}
	return nil
}

// validateGameAccess checks if player can access the requested game
func (s *PlayerService) validateGameAccess(p *player.Player, requestedGameID *uuid.UUID) error {
	// Cross-game accounts (player.GameID == nil) can login to any game
	if p.GameID == nil {
		return nil
	}

	// Game-specific accounts must match the requested game
	if requestedGameID == nil {
		// Player is game-specific but no game was specified in login
		return player.ErrGameAccessDenied
	}

	if *p.GameID != *requestedGameID {
		// Player is bound to a different game
		return player.ErrGameAccessDenied
	}

	return nil
}
