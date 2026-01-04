package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	adminDomain "github.com/slotmachine/backend/domain/admin"
	gameDomain "github.com/slotmachine/backend/domain/game"
	"github.com/slotmachine/backend/domain/player"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/domain/session"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/infra/cache"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/pkg/util"
	"golang.org/x/crypto/bcrypt"
)

// AdminService implements adminDomain.Service
type AdminService struct {
	repo              adminDomain.Repository
	playerRepo        player.Repository
	reelStripRepo     reelstrip.Repository
	gameRepo          gameDomain.Repository
	playerSessionRepo session.PlayerSessionRepository
	cache             *cache.RedisClient
	cfg               *config.Config
	logger            *logger.Logger
}

// NewAdminService creates a new admin service
func NewAdminService(
	repo adminDomain.Repository,
	playerRepo player.Repository,
	reelStripRepo reelstrip.Repository,
	gameRepo gameDomain.Repository,
	playerSessionRepo session.PlayerSessionRepository,
	redisCache *cache.RedisClient,
	cfg *config.Config,
	log *logger.Logger,
) adminDomain.Service {
	return &AdminService{
		repo:              repo,
		playerRepo:        playerRepo,
		reelStripRepo:     reelStripRepo,
		gameRepo:          gameRepo,
		playerSessionRepo: playerSessionRepo,
		cache:             redisCache,
		cfg:               cfg,
		logger:            log,
	}
}

// Login authenticates an admin and returns a JWT token
func (s *AdminService) Login(ctx context.Context, username, password, ip string) (*adminDomain.Admin, string, error) {
	log := s.logger.WithTraceContext(ctx)
	// Get admin by username
	adm, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if err == adminDomain.ErrAdminNotFound {
			return nil, "", adminDomain.ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("failed to get admin: %w", err)
	}

	// Check if account is locked
	if adm.IsLocked() {
		log.Warn().
			Str("username", username).
			Msg("Login attempt on locked account")
		return nil, "", adminDomain.ErrAccountLocked
	}

	// Check account status
	if adm.Status == adminDomain.StatusInactive {
		return nil, "", adminDomain.ErrAccountInactive
	}
	if adm.Status == adminDomain.StatusSuspended {
		return nil, "", adminDomain.ErrAccountSuspended
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(adm.PasswordHash), []byte(password)); err != nil {
		// Increment failed attempts
		if err := s.repo.IncrementFailedAttempts(ctx, adm.ID); err != nil {
			log.Error().Err(err).Msg("Failed to increment failed attempts")
		}

		log.Warn().
			Str("username", username).
			Msg("Invalid password attempt")

		return nil, "", adminDomain.ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	if err := s.repo.ResetFailedAttempts(ctx, adm.ID); err != nil {
		log.Error().Err(err).Msg("Failed to reset failed attempts")
	}

	// Update last login
	if err := s.repo.UpdateLastLogin(ctx, adm.ID, ip); err != nil {
		log.Error().Err(err).Msg("Failed to update last login")
	}

	// Generate JWT token (admin has no game_id, pass nil)
	token, err := util.GenerateJWT(adm.ID.String(), adm.Username, nil, s.cfg.JWT.Secret, s.cfg.JWT.ExpirationHours)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Info().
		Str("admin_id", adm.ID.String()).
		Str("username", username).
		Msg("Admin logged in successfully")

	// Sanitize before returning
	adm.Sanitize()

	return adm, token, nil
}

// ValidateToken validates a JWT token and returns the admin
func (s *AdminService) ValidateToken(ctx context.Context, token string) (*adminDomain.Admin, error) {
	// log := s.logger.WithTraceContext(ctx)
	// Parse and validate token
	claims, err := util.ValidateJWT(token, s.cfg.JWT.Secret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get admin by ID
	adminID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid admin ID in token: %w", err)
	}

	adm, err := s.repo.GetByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	// Check if account is active
	if !adm.IsActive() {
		return nil, adminDomain.ErrAccountInactive
	}

	return adm, nil
}

// CreateAdmin creates a new admin
func (s *AdminService) CreateAdmin(ctx context.Context, req adminDomain.CreateAdminRequest, createdBy uuid.UUID) (*adminDomain.Admin, error) {
	log := s.logger.WithTraceContext(ctx)

	// Validate password strength
	if len(req.Password) < 8 {
		return nil, adminDomain.ErrWeakPassword
	}

	// Check if username exists
	if _, err := s.repo.GetByUsername(ctx, req.Username); err == nil {
		return nil, adminDomain.ErrDuplicateUsername
	}

	// Check if email exists
	if _, err := s.repo.GetByEmail(ctx, req.Email); err == nil {
		return nil, adminDomain.ErrDuplicateEmail
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin
	adm := &adminDomain.Admin{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		FullName:     req.FullName,
		Role:         req.Role,
		Status:       adminDomain.StatusActive,
		Permissions:  adminDomain.StringArray(req.Permissions),
		CreatedBy:    &createdBy,
	}

	if err := s.repo.Create(ctx, adm); err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	log.Info().
		Str("admin_id", adm.ID.String()).
		Str("username", adm.Username).
		Str("created_by", createdBy.String()).
		Msg("Admin created successfully")

	adm.Sanitize()
	return adm, nil
}

// GetAdmin retrieves an admin by ID
func (s *AdminService) GetAdmin(ctx context.Context, id uuid.UUID) (*adminDomain.Admin, error) {
	adm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	adm.Sanitize()
	return adm, nil
}

// UpdateAdmin updates an admin
func (s *AdminService) UpdateAdmin(ctx context.Context, id uuid.UUID, req adminDomain.UpdateAdminRequest, updatedBy uuid.UUID) (*adminDomain.Admin, error) {
	log := s.logger.WithTraceContext(ctx)

	adm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Email != nil {
		// Check if email is already used by another admin
		if existingAdmin, err := s.repo.GetByEmail(ctx, *req.Email); err == nil && existingAdmin.ID != id {
			return nil, adminDomain.ErrDuplicateEmail
		}
		adm.Email = *req.Email
	}
	if req.FullName != nil {
		adm.FullName = *req.FullName
	}
	if req.Role != nil {
		adm.Role = *req.Role
	}
	if req.Permissions != nil {
		adm.Permissions = adminDomain.StringArray(*req.Permissions)
	}

	adm.UpdatedBy = &updatedBy
	adm.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, adm); err != nil {
		return nil, fmt.Errorf("failed to update admin: %w", err)
	}

	log.Info().
		Str("admin_id", adm.ID.String()).
		Str("updated_by", updatedBy.String()).
		Msg("Admin updated successfully")

	adm.Sanitize()
	return adm, nil
}

// DeleteAdmin soft deletes an admin
func (s *AdminService) DeleteAdmin(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	log := s.logger.WithTraceContext(ctx)

	// Prevent self-deletion
	if id == deletedBy {
		return adminDomain.ErrCannotDeleteSelf
	}

	// Get admin to check role
	adm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Prevent deleting super admin
	if adm.Role == adminDomain.RoleSuperAdmin {
		deleter, _ := s.repo.GetByID(ctx, deletedBy)
		if deleter == nil || deleter.Role != adminDomain.RoleSuperAdmin {
			return adminDomain.ErrCannotModifySuperAdmin
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	log.Info().
		Str("admin_id", id.String()).
		Str("deleted_by", deletedBy.String()).
		Msg("Admin deleted successfully")

	return nil
}

// ListAdmins retrieves all admins with filters
func (s *AdminService) ListAdmins(ctx context.Context, filters adminDomain.ListFilters) ([]*adminDomain.Admin, int64, error) {
	admins, total, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	// Sanitize all admins
	for _, adm := range admins {
		adm.Sanitize()
	}

	return admins, total, nil
}

// ChangePassword changes an admin's password
func (s *AdminService) ChangePassword(ctx context.Context, adminID uuid.UUID, oldPassword, newPassword string) error {
	log := s.logger.WithTraceContext(ctx)

	// Validate new password strength
	if len(newPassword) < 8 {
		return adminDomain.ErrWeakPassword
	}

	adm, err := s.repo.GetByID(ctx, adminID)
	if err != nil {
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(adm.PasswordHash), []byte(oldPassword)); err != nil {
		return adminDomain.ErrInvalidPassword
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adm.PasswordHash = string(newPasswordHash)
	adm.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, adm); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	log.Info().
		Str("admin_id", adminID.String()).
		Msg("Password changed successfully")

	return nil
}

// ResetPassword resets an admin's password (admin action)
func (s *AdminService) ResetPassword(ctx context.Context, adminID uuid.UUID, newPassword string, resetBy uuid.UUID) error {
	log := s.logger.WithTraceContext(ctx)

	// Validate new password strength
	if len(newPassword) < 8 {
		return adminDomain.ErrWeakPassword
	}

	adm, err := s.repo.GetByID(ctx, adminID)
	if err != nil {
		return err
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adm.PasswordHash = string(newPasswordHash)
	adm.UpdatedBy = &resetBy
	adm.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, adm); err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	log.Info().
		Str("admin_id", adminID.String()).
		Str("reset_by", resetBy.String()).
		Msg("Password reset successfully")

	return nil
}

// ActivateAdmin activates an admin account
func (s *AdminService) ActivateAdmin(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	adm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	adm.Status = adminDomain.StatusActive
	adm.UpdatedBy = &updatedBy
	adm.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, adm); err != nil {
		return fmt.Errorf("failed to activate admin: %w", err)
	}

	s.logger.Info().
		Str("admin_id", id.String()).
		Str("updated_by", updatedBy.String()).
		Msg("Admin activated")

	return nil
}

// DeactivateAdmin deactivates an admin account
func (s *AdminService) DeactivateAdmin(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	adm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	adm.Status = adminDomain.StatusInactive
	adm.UpdatedBy = &updatedBy
	adm.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, adm); err != nil {
		return fmt.Errorf("failed to deactivate admin: %w", err)
	}

	s.logger.Info().
		Str("admin_id", id.String()).
		Str("updated_by", updatedBy.String()).
		Msg("Admin deactivated")

	return nil
}

// SuspendAdmin suspends an admin account
func (s *AdminService) SuspendAdmin(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	adm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	adm.Status = adminDomain.StatusSuspended
	adm.UpdatedBy = &updatedBy
	adm.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, adm); err != nil {
		return fmt.Errorf("failed to suspend admin: %w", err)
	}

	s.logger.Info().
		Str("admin_id", id.String()).
		Str("updated_by", updatedBy.String()).
		Msg("Admin suspended")

	return nil
}

// CreatePlayer creates a new player from admin panel
func (s *AdminService) CreatePlayer(ctx context.Context, req adminDomain.CreatePlayerRequest, createdBy uuid.UUID) (interface{}, error) {
	log := s.logger.WithTraceContext(ctx)

	// Validate password strength
	if len(req.Password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	// Check if username exists for the same game
	if _, err := s.playerRepo.GetByUsernameAndGame(ctx, req.Username, req.GameID); err == nil {
		return nil, fmt.Errorf("username already exists for this game")
	}

	// Check if email exists for the same game
	if _, err := s.playerRepo.GetByEmailAndGame(ctx, req.Email, req.GameID); err == nil {
		return nil, fmt.Errorf("email already exists for this game")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default balance if not provided
	balance := req.Balance
	if balance <= 0 {
		balance = 100000.00 // Default balance
	}

	// Create player
	newPlayer := &player.Player{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		Balance:      balance,
		GameID:       req.GameID,
		IsActive:     true,
		IsVerified:   false,
	}

	if err := s.playerRepo.Create(ctx, newPlayer); err != nil {
		log.Error().Err(err).Str("username", req.Username).Msg("Failed to create player")
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	log.Info().
		Str("player_id", newPlayer.ID.String()).
		Str("username", newPlayer.Username).
		Str("created_by", createdBy.String()).
		Msg("Player created successfully by admin")

	// Build response with game info
	result := map[string]interface{}{
		"id":            newPlayer.ID,
		"username":      newPlayer.Username,
		"email":         newPlayer.Email,
		"balance":       newPlayer.Balance,
		"game_id":       newPlayer.GameID,
		"is_active":     newPlayer.IsActive,
		"is_verified":   newPlayer.IsVerified,
		"created_at":    newPlayer.CreatedAt,
		"updated_at":    newPlayer.UpdatedAt,
	}

	// Add game info if player is bound to a game
	if newPlayer.GameID != nil {
		if game, err := s.gameRepo.GetGameByID(ctx, *newPlayer.GameID); err == nil {
			result["game"] = map[string]interface{}{
				"id":   game.ID,
				"name": game.Name,
			}
		}
	}

	return result, nil
}

// GetPlayer retrieves a player by ID
func (s *AdminService) GetPlayer(ctx context.Context, id uuid.UUID) (interface{}, error) {
	log := s.logger.WithTraceContext(ctx)

	p, err := s.playerRepo.GetByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("player_id", id.String()).Msg("Failed to get player")
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	// Build response with game info
	result := map[string]interface{}{
		"id":            p.ID,
		"username":      p.Username,
		"email":         p.Email,
		"balance":       p.Balance,
		"game_id":       p.GameID,
		"total_spins":   p.TotalSpins,
		"total_wagered": p.TotalWagered,
		"total_won":     p.TotalWon,
		"is_active":     p.IsActive,
		"is_verified":   p.IsVerified,
		"created_at":    p.CreatedAt,
		"updated_at":    p.UpdatedAt,
		"last_login_at": p.LastLoginAt,
	}

	// Add game info if player is bound to a game
	if p.GameID != nil {
		if game, err := s.gameRepo.GetGameByID(ctx, *p.GameID); err == nil {
			result["game"] = map[string]interface{}{
				"id":   game.ID,
				"name": game.Name,
			}
		}
	}

	return result, nil
}

// ListPlayers retrieves a list of players with filters
func (s *AdminService) ListPlayers(ctx context.Context, filters adminDomain.PlayerListFilters) (interface{}, int64, error) {
	log := s.logger.WithTraceContext(ctx)

	// Convert admin filters to player filters
	playerFilters := player.ListFilters{
		Username: filters.Username,
		Email:    filters.Email,
		GameID:   filters.GameID,
		IsActive: filters.IsActive,
		Page:     filters.Page,
		Limit:    filters.Limit,
		SortBy:   filters.SortBy,
		SortDesc: filters.SortDesc,
	}

	players, total, err := s.playerRepo.List(ctx, playerFilters)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list players")
		return nil, 0, fmt.Errorf("failed to list players: %w", err)
	}

	// Get player IDs for bulk assignment lookup
	playerIDs := make([]uuid.UUID, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID
	}

	// Bulk fetch assignments for all players
	assignments, err := s.reelStripRepo.GetPlayerAssignmentsByPlayerIDs(ctx, playerIDs)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch player assignments, continuing without assignments")
		assignments = make(map[uuid.UUID]*reelstrip.PlayerReelStripAssignment)
	}

	// Collect unique game IDs for bulk lookup
	var uniqueGameIDs []uuid.UUID
	gameIDSet := make(map[uuid.UUID]bool)
	for _, p := range players {
		if p.GameID != nil && !gameIDSet[*p.GameID] {
			gameIDSet[*p.GameID] = true
			uniqueGameIDs = append(uniqueGameIDs, *p.GameID)
		}
	}

	// Fetch game info for all unique game IDs in a single query (batch fetch)
	gameMap, err := s.gameRepo.GetGamesByIDs(ctx, uniqueGameIDs)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch games, continuing without game info")
		gameMap = make(map[uuid.UUID]*gameDomain.Game)
	}

	// Get active sessions for all players
	playerIDStrings := make([]string, len(players))
	for i, p := range players {
		playerIDStrings[i] = p.ID.String()
	}
	activeSessions := make(map[string]bool)
	if s.cache != nil {
		activeSessions, _ = s.cache.GetPlayersWithActiveSessions(ctx, playerIDStrings)
	}

	// Convert to safe response format (without password hash)
	result := make([]map[string]interface{}, len(players))
	for i, p := range players {
		playerData := map[string]interface{}{
			"id":            p.ID,
			"username":      p.Username,
			"email":         p.Email,
			"balance":       p.Balance,
			"game_id":       p.GameID,
			"total_spins":   p.TotalSpins,
			"total_wagered": p.TotalWagered,
			"total_won":     p.TotalWon,
			"is_active":     p.IsActive,
			"is_verified":   p.IsVerified,
			"created_at":    p.CreatedAt,
			"updated_at":    p.UpdatedAt,
			"last_login_at": p.LastLoginAt,
		}

		// Add game info if player is bound to a game
		if p.GameID != nil {
			if game, ok := gameMap[*p.GameID]; ok {
				playerData["game"] = map[string]interface{}{
					"id":   game.ID,
					"name": game.Name,
				}
			}
		}

		// Add assignment if exists
		if assignment, ok := assignments[p.ID]; ok && assignment.ID != uuid.Nil {
			playerData["assignment"] = map[string]interface{}{
				"id":                    assignment.ID,
				"base_game_config_id":   assignment.BaseGameConfigID,
				"free_spins_config_id":  assignment.FreeSpinsConfigID,
				"reason":                assignment.Reason,
				"assigned_by":           assignment.AssignedBy,
				"assigned_at":           assignment.AssignedAt,
				"expires_at":            assignment.ExpiresAt,
				"is_active":             assignment.IsActive,
			}
		}

		// Add active session status
		playerData["has_active_session"] = activeSessions[p.ID.String()]

		result[i] = playerData
	}

	log.Info().
		Int("count", len(players)).
		Int64("total", total).
		Msg("Listed players")

	return result, total, nil
}

// ActivatePlayer activates a player account
func (s *AdminService) ActivatePlayer(ctx context.Context, playerID uuid.UUID, updatedBy uuid.UUID) error {
	log := s.logger.WithTraceContext(ctx)

	p, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player")
		return fmt.Errorf("failed to get player: %w", err)
	}

	if p.IsActive {
		return fmt.Errorf("player is already active")
	}

	p.IsActive = true
	p.UpdatedAt = time.Now()

	if err := s.playerRepo.Update(ctx, p); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to activate player")
		return fmt.Errorf("failed to activate player: %w", err)
	}

	log.Info().
		Str("player_id", playerID.String()).
		Str("updated_by", updatedBy.String()).
		Msg("Player activated")

	return nil
}

// DeactivatePlayer deactivates a player account
func (s *AdminService) DeactivatePlayer(ctx context.Context, playerID uuid.UUID, updatedBy uuid.UUID) error {
	log := s.logger.WithTraceContext(ctx)

	p, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player")
		return fmt.Errorf("failed to get player: %w", err)
	}

	if !p.IsActive {
		return fmt.Errorf("player is already inactive")
	}

	p.IsActive = false
	p.UpdatedAt = time.Now()

	if err := s.playerRepo.Update(ctx, p); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to deactivate player")
		return fmt.Errorf("failed to deactivate player: %w", err)
	}

	log.Info().
		Str("player_id", playerID.String()).
		Str("updated_by", updatedBy.String()).
		Msg("Player deactivated")

	return nil
}

// ForceLogoutPlayer force logout a player from all sessions
func (s *AdminService) ForceLogoutPlayer(ctx context.Context, playerID uuid.UUID, adminID uuid.UUID) error {
	log := s.logger.WithTraceContext(ctx)

	// Check if player exists
	p, err := s.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to get player")
		return fmt.Errorf("player not found: %w", err)
	}

	// Deactivate all player sessions in database
	if err := s.playerSessionRepo.DeactivateAllPlayerSessions(ctx, playerID, session.LogoutReasonForced); err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to deactivate player sessions")
		return fmt.Errorf("failed to force logout player: %w", err)
	}

	// Clear player sessions from Redis cache
	if s.cache != nil {
		if err := s.cache.DeletePlayerSessions(ctx, playerID.String()); err != nil {
			log.Warn().Err(err).Str("player_id", playerID.String()).Msg("Failed to clear player sessions from cache")
			// Don't fail - DB sessions are already invalidated
		}
	}

	log.Info().
		Str("player_id", playerID.String()).
		Str("username", p.Username).
		Str("admin_id", adminID.String()).
		Msg("Player force logged out by admin")

	return nil
}
