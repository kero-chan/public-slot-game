package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/provablyfair"
	"github.com/slotmachine/backend/internal/api/dto"
	"github.com/slotmachine/backend/internal/pkg/logger"
	"github.com/slotmachine/backend/internal/service"
)

// ProvablyFairHandler handles provably fair endpoints
type ProvablyFairHandler struct {
	pfService provablyfair.Service
	logger    *logger.Logger
}

// NewProvablyFairHandler creates a new provably fair handler
func NewProvablyFairHandler(
	pfService *service.ProvablyFairService,
	log *logger.Logger,
) *ProvablyFairHandler {
	return &ProvablyFairHandler{
		pfService: pfService,
		logger:    log,
	}
}

// StartPFSession starts a new provably fair session
// POST /api/pf/sessions
func (h *ProvablyFairHandler) StartPFSession(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get player ID from context
	playerIDStr := c.Locals("user_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid player ID in token",
		})
	}

	// Get game session ID from context (must have active game session)
	gameSessionIDStr, ok := c.Locals("session_id").(string)
	if !ok || gameSessionIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "no_active_session",
			Message: "No active game session. Start a game session first.",
		})
	}
	gameSessionID, err := uuid.Parse(gameSessionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_session",
			Message: "Invalid game session ID",
		})
	}

	// Start PF session (no client_seed needed - it's per-spin now)
	// Note: This standalone endpoint doesn't use Dual Commitment Protocol
	// For Dual Commitment, use the session handler which passes theta_commitment from StartSessionRequest
	result, err := h.pfService.StartSession(c.Context(), playerID, gameSessionID, "")
	if err != nil {
		log.Error().Err(err).Str("player_id", playerID.String()).Msg("Failed to start PF session")

		if err == provablyfair.ErrSessionAlreadyActive {
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "pf_session_exists",
				Message: "Player already has an active provably fair session",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_start_pf_session",
			Message: "Failed to start provably fair session",
		})
	}

	// Build response (no client_seed - it's per-spin now)
	response := dto.StartPFSessionResponse{
		SessionID:      result.SessionID.String(),
		ServerSeedHash: result.ServerSeedHash,
		NonceStart:     result.NonceStart,
	}

	log.Info().
		Str("pf_session_id", result.SessionID.String()).
		Str("player_id", playerID.String()).
		Str("server_seed_hash", result.ServerSeedHash).
		Msg("PF session started")

	return c.Status(fiber.StatusCreated).JSON(response)
}

// EndPFSession ends the current provably fair session and reveals the server seed
// POST /api/pf/sessions/end
func (h *ProvablyFairHandler) EndPFSession(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get player ID from context
	playerIDStr := c.Locals("user_id").(string)
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid player ID in token",
		})
	}

	// Get game session ID from context
	gameSessionIDStr, ok := c.Locals("session_id").(string)
	if !ok || gameSessionIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "no_active_session",
			Message: "No active game session",
		})
	}
	gameSessionID, err := uuid.Parse(gameSessionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_session",
			Message: "Invalid game session ID",
		})
	}

	// End PF session
	result, err := h.pfService.EndSession(c.Context(), gameSessionID)
	if err != nil {
		log.Error().Err(err).Str("game_session_id", gameSessionID.String()).Msg("Failed to end PF session")

		if err == provablyfair.ErrStateNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "pf_session_not_found",
				Message: "No active provably fair session found",
			})
		}

		if err == provablyfair.ErrSessionAlreadyEnded {
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Error:   "pf_session_already_ended",
				Message: "Provably fair session already ended",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_end_pf_session",
			Message: "Failed to end provably fair session",
		})
	}

	// Convert spins to DTO format - each spin has its own client_seed
	spins := make([]dto.SpinVerificationData, len(result.Spins))
	for i, s := range result.Spins {
		var configIDStr *string
		if s.ReelStripConfigID != nil {
			str := s.ReelStripConfigID.String()
			configIDStr = &str
		}
		spins[i] = dto.SpinVerificationData{
			SpinIndex:         s.SpinIndex,
			Nonce:             s.Nonce,
			ClientSeed:        s.ClientSeed, // Per-spin client seed
			SpinHash:          s.SpinHash,
			PrevSpinHash:      s.PrevSpinHash,
			ReelPositions:     s.ReelPositions,
			ReelStripConfigID: configIDStr,
			GameMode:          s.GameMode,
			IsFreeSpin:        s.IsFreeSpin,
		}
	}

	// Build response (no session-level client_seed - it's per-spin now)
	response := dto.EndPFSessionResponse{
		SessionID:      result.SessionID.String(),
		ServerSeed:     result.ServerSeed,
		ServerSeedHash: result.ServerSeedHash,
		TotalSpins:     result.TotalSpins,
		Spins:          spins,
	}

	log.Info().
		Str("pf_session_id", result.SessionID.String()).
		Str("player_id", playerID.String()).
		Int64("total_spins", result.TotalSpins).
		Msg("PF session ended, server seed revealed")

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetPFSessionStatus gets the current PF session status
// GET /api/pf/sessions/status
func (h *ProvablyFairHandler) GetPFSessionStatus(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get game session ID from context
	gameSessionIDStr, ok := c.Locals("session_id").(string)
	if !ok || gameSessionIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "no_active_session",
			Message: "No active game session",
		})
	}
	gameSessionID, err := uuid.Parse(gameSessionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_session",
			Message: "Invalid game session ID",
		})
	}

	// Get session state
	state, err := h.pfService.GetSessionState(c.Context(), gameSessionID)
	if err != nil {
		log.Debug().Err(err).Str("game_session_id", gameSessionID.String()).Msg("PF session not found")
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "pf_session_not_found",
			Message: "No provably fair session found",
		})
	}

	// Build response (never expose server_seed while session is active)
	// No client_seed - it's per-spin now
	response := dto.PFSessionStatusResponse{
		SessionID:      state.SessionID.String(),
		ServerSeedHash: state.ServerSeedHash,
		CurrentNonce:   state.Nonce,
		LastSpinHash:   state.LastSpinHash,
		Status:         state.Status,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetVerificationData gets all data needed to verify a completed session
// GET /api/pf/sessions/:sessionId/verify
func (h *ProvablyFairHandler) GetVerificationData(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get PF session ID from URL params
	pfSessionIDStr := c.Params("sessionId")
	pfSessionID, err := uuid.Parse(pfSessionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_session_id",
			Message: "Invalid provably fair session ID",
		})
	}

	// Get verification data
	data, err := h.pfService.GetVerificationData(c.Context(), pfSessionID)
	if err != nil {
		log.Error().Err(err).Str("pf_session_id", pfSessionIDStr).Msg("Failed to get verification data")

		if err == provablyfair.ErrSessionNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "pf_session_not_found",
				Message: "Provably fair session not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "failed_to_get_verification_data",
			Message: "Failed to retrieve verification data",
		})
	}

	// Convert spins to DTO format - each spin has its own client_seed
	spins := make([]dto.SpinVerificationData, len(data.Spins))
	for i, s := range data.Spins {
		var configIDStr *string
		if s.ReelStripConfigID != nil {
			str := s.ReelStripConfigID.String()
			configIDStr = &str
		}
		spins[i] = dto.SpinVerificationData{
			SpinIndex:         s.SpinIndex,
			Nonce:             s.Nonce,
			ClientSeed:        s.ClientSeed, // Per-spin client seed
			SpinHash:          s.SpinHash,
			PrevSpinHash:      s.PrevSpinHash,
			ReelPositions:     s.ReelPositions,
			ReelStripConfigID: configIDStr,
			GameMode:          s.GameMode,
			IsFreeSpin:        s.IsFreeSpin,
		}
	}

	// Build response (no session-level client_seed - it's per-spin now)
	response := dto.VerificationDataResponse{
		SessionID:      data.SessionID.String(),
		ServerSeed:     data.ServerSeed,
		ServerSeedHash: data.ServerSeedHash,
		Spins:          spins,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// VerifySession verifies a session's hash chain
// POST /api/pf/sessions/:sessionId/verify
func (h *ProvablyFairHandler) VerifySession(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get PF session ID from URL params
	pfSessionIDStr := c.Params("sessionId")
	pfSessionID, err := uuid.Parse(pfSessionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_session_id",
			Message: "Invalid provably fair session ID",
		})
	}

	// Parse request
	var req dto.VerifySessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Verify session
	valid, err := h.pfService.VerifySession(c.Context(), pfSessionID, req.ServerSeed)
	if err != nil {
		log.Error().Err(err).Str("pf_session_id", pfSessionIDStr).Msg("Session verification failed")

		if err == provablyfair.ErrSessionNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error:   "pf_session_not_found",
				Message: "Provably fair session not found",
			})
		}

		if err == provablyfair.ErrInvalidServerSeed {
			return c.Status(fiber.StatusOK).JSON(dto.VerifySessionResponse{
				SessionID: pfSessionIDStr,
				Valid:     false,
				Message:   "Server seed hash does not match",
			})
		}

		// Hash chain verification error
		return c.Status(fiber.StatusOK).JSON(dto.VerifySessionResponse{
			SessionID: pfSessionIDStr,
			Valid:     false,
			Message:   err.Error(),
		})
	}

	// Build response
	response := dto.VerifySessionResponse{
		SessionID: pfSessionIDStr,
		Valid:     valid,
		Message:   "Hash chain verification successful",
	}

	log.Info().
		Str("pf_session_id", pfSessionIDStr).
		Bool("valid", valid).
		Msg("Session verification completed")

	return c.Status(fiber.StatusOK).JSON(response)
}

// VerifySpin verifies a single spin's hash
// POST /api/pf/verify/spin
func (h *ProvablyFairHandler) VerifySpin(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse request
	var req dto.VerifySpinRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.ServerSeed == "" || len(req.ServerSeed) != 64 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_server_seed",
			Message: "Server seed must be a 64-character hex string",
		})
	}

	if req.ClientSeed == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_client_seed",
			Message: "Client seed is required",
		})
	}

	if req.Nonce < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_nonce",
			Message: "Nonce must be at least 1",
		})
	}

	if req.PrevSpinHash == "" || len(req.PrevSpinHash) != 64 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_prev_spin_hash",
			Message: "Previous spin hash must be a 64-character hex string",
		})
	}

	if req.SpinHash == "" || len(req.SpinHash) != 64 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_spin_hash",
			Message: "Spin hash must be a 64-character hex string",
		})
	}

	// Verify spin
	result, err := h.pfService.VerifySpin(c.Context(), &provablyfair.VerifySpinInput{
		ServerSeed:   req.ServerSeed,
		ClientSeed:   req.ClientSeed,
		Nonce:        req.Nonce,
		PrevSpinHash: req.PrevSpinHash,
		SpinHash:     req.SpinHash,
	})
	if err != nil {
		log.Error().Err(err).Msg("Spin verification failed")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "verification_failed",
			Message: "Failed to verify spin",
		})
	}

	// Build response
	message := "Spin hash verification successful"
	if !result.Valid {
		message = "Spin hash does not match expected value"
	}

	response := dto.VerifySpinResponse{
		Valid:            result.Valid,
		ExpectedSpinHash: result.ExpectedSpinHash,
		ProvidedSpinHash: req.SpinHash,
		ServerSeedHashOK: true, // Always true since we calculate from provided server_seed
		ServerSeedHash:   result.ServerSeedHash,
		Message:          message,
	}

	log.Info().
		Bool("valid", result.Valid).
		Int64("nonce", req.Nonce).
		Msg("Spin verification completed")

	return c.Status(fiber.StatusOK).JSON(response)
}

// VerifySpinWithReel verifies a spin hash and its reel positions
// POST /api/pf/verify/spin-with-reel
func (h *ProvablyFairHandler) VerifySpinWithReel(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Parse request
	var req dto.VerifySpinWithReelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.ServerSeed == "" || len(req.ServerSeed) != 64 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_server_seed",
			Message: "Server seed must be a 64-character hex string",
		})
	}

	if req.ClientSeed == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_client_seed",
			Message: "Client seed is required",
		})
	}

	if req.Nonce < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_nonce",
			Message: "Nonce must be at least 1",
		})
	}

	if req.PrevSpinHash == "" || len(req.PrevSpinHash) != 64 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_prev_spin_hash",
			Message: "Previous spin hash must be a 64-character hex string",
		})
	}

	if req.SpinHash == "" || len(req.SpinHash) != 64 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_spin_hash",
			Message: "Spin hash must be a 64-character hex string",
		})
	}

	if len(req.ReelPositions) != 5 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_reel_positions",
			Message: "Reel positions must be an array of exactly 5 integers",
		})
	}

	// Parse reel strip config ID
	reelStripConfigID, err := uuid.Parse(req.ReelStripConfigID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_reel_strip_config_id",
			Message: "Invalid reel strip config ID format",
		})
	}

	// Verify spin with reel positions
	result, err := h.pfService.VerifySpinWithReel(c.Context(), &provablyfair.VerifySpinWithReelInput{
		ServerSeed:        req.ServerSeed,
		ClientSeed:        req.ClientSeed,
		Nonce:             req.Nonce,
		PrevSpinHash:      req.PrevSpinHash,
		SpinHash:          req.SpinHash,
		ReelPositions:     req.ReelPositions,
		ReelStripConfigID: reelStripConfigID,
	})
	if err != nil {
		log.Error().Err(err).Msg("Spin verification with reel failed")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "verification_failed",
			Message: "Failed to verify spin",
		})
	}

	// Build message
	var message string
	switch {
	case result.Valid:
		message = "Spin hash and reel positions verification successful"
	case !result.SpinHashValid && !result.ReelPositionsValid:
		message = "Both spin hash and reel positions do not match"
	case !result.SpinHashValid:
		message = "Spin hash does not match expected value"
	case !result.ReelPositionsValid:
		message = "Reel positions do not match expected values"
	}

	response := dto.VerifySpinWithReelResponse{
		Valid:                 result.Valid,
		SpinHashValid:         result.SpinHashValid,
		ReelPositionsValid:    result.ReelPositionsValid,
		ExpectedSpinHash:      result.ExpectedSpinHash,
		ExpectedReelPositions: result.ExpectedReelPositions,
		ProvidedReelPositions: req.ReelPositions,
		ServerSeedHash:        result.ServerSeedHash,
		Message:               message,
	}

	log.Info().
		Bool("valid", result.Valid).
		Bool("spin_hash_valid", result.SpinHashValid).
		Bool("reel_positions_valid", result.ReelPositionsValid).
		Int64("nonce", req.Nonce).
		Msg("Spin verification with reel completed")

	return c.Status(fiber.StatusOK).JSON(response)
}

// VerifyActiveSpin verifies a spin in an active session using server's stored server_seed
// POST /api/pf/sessions/verify-spin
func (h *ProvablyFairHandler) VerifyActiveSpin(c *fiber.Ctx) error {
	log := h.logger.WithTrace(c)

	// Get player ID from context (set by auth middleware)
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
	}

	playerID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
	}

	// Get active PF session for this player
	pfState, err := h.pfService.GetActiveSessionByPlayer(c.Context(), playerID)
	if err != nil || pfState == nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "no_active_session",
			Message: "No active provably fair session found. Start a session first.",
		})
	}
	gameSessionID := pfState.GameSessionID

	// Parse request
	var req dto.VerifyActiveSpinRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.ClientSeed == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_client_seed",
			Message: "Client seed is required",
		})
	}

	if req.Nonce < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_nonce",
			Message: "Nonce must be at least 1",
		})
	}

	if req.SpinHash == "" || len(req.SpinHash) != 64 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_spin_hash",
			Message: "Spin hash must be a 64-character hex string",
		})
	}

	// Parse optional reel strip config ID
	var reelStripConfigID uuid.UUID
	if req.ReelStripConfigID != "" {
		var err error
		reelStripConfigID, err = uuid.Parse(req.ReelStripConfigID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Error:   "invalid_reel_strip_config_id",
				Message: "Invalid reel strip config ID format",
			})
		}
	}

	// Validate reel positions if provided
	if len(req.ReelPositions) > 0 && len(req.ReelPositions) != 5 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "invalid_reel_positions",
			Message: "Reel positions must be an array of exactly 5 integers",
		})
	}

	if len(req.ReelPositions) == 5 && reelStripConfigID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   "missing_reel_strip_config_id",
			Message: "reel_strip_config_id is required when verifying reel positions",
		})
	}

	// Verify spin
	result, err := h.pfService.VerifyActiveSpin(c.Context(), gameSessionID, &provablyfair.VerifyActiveSpinInput{
		ClientSeed:        req.ClientSeed,
		Nonce:             req.Nonce,
		SpinHash:          req.SpinHash,
		PrevSpinHash:      req.PrevSpinHash, // Optional: if provided, use this instead of server-calculated value
		ReelPositions:     req.ReelPositions,
		ReelStripConfigID: reelStripConfigID,
	})
	if err != nil {
		log.Error().Err(err).Msg("Active spin verification failed")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error:   "verification_failed",
			Message: err.Error(),
		})
	}

	// Build message
	var message string
	if result.Valid {
		if result.ReelPositionsValid != nil {
			message = "Spin hash and reel positions verification successful"
		} else {
			message = "Spin hash verification successful"
		}
	} else if !result.SpinHashValid {
		message = "Spin hash does not match expected value"
	} else if result.ReelPositionsValid != nil && !*result.ReelPositionsValid {
		message = "Reel positions do not match expected values"
	}

	response := dto.VerifyActiveSpinResponse{
		Valid:                 result.Valid,
		SpinHashValid:         result.SpinHashValid,
		ReelPositionsValid:    result.ReelPositionsValid,
		ExpectedSpinHash:      result.ExpectedSpinHash,
		ExpectedReelPositions: result.ExpectedReelPositions,
		ServerSeedHash:        result.ServerSeedHash,
		PrevSpinHash:          result.PrevSpinHash,
		Message:               message,
	}

	log.Info().
		Bool("valid", result.Valid).
		Bool("spin_hash_valid", result.SpinHashValid).
		Int64("nonce", req.Nonce).
		Msg("Active spin verification completed")

	return c.Status(fiber.StatusOK).JSON(response)
}
