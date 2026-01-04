package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slotmachine/backend/domain/provablyfair"
	"github.com/slotmachine/backend/domain/reelstrip"
	"github.com/slotmachine/backend/internal/config"
	"github.com/slotmachine/backend/internal/game/rng"
	"github.com/slotmachine/backend/internal/pkg/crypto"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// ProvablyFairService implements the provably fair service
type ProvablyFairService struct {
	repo          provablyfair.Repository
	cache         provablyfair.CacheRepository
	reelstripRepo reelstrip.Repository
	hashGenerator *rng.HashChainGenerator
	encryptor     *crypto.AESEncryptor
	logger        *logger.Logger
}

// NewProvablyFairService creates a new provably fair service
func NewProvablyFairService(
	repo provablyfair.Repository,
	cache provablyfair.CacheRepository,
	reelstripRepo reelstrip.Repository,
	cfg *config.Config,
	log *logger.Logger,
) (provablyfair.Service, error) {
	// Initialize AES encryptor for server seed encryption
	encryptor, err := crypto.NewAESEncryptor(cfg.ProvablyFair.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AES encryptor: %w", err)
	}

	return &ProvablyFairService{
		repo:          repo,
		cache:         cache,
		reelstripRepo: reelstripRepo,
		hashGenerator: rng.NewHashChainGenerator(),
		encryptor:     encryptor,
		logger:        log,
	}, nil
}

// Ensure ProvablyFairService implements Service
var _ provablyfair.Service = (*ProvablyFairService)(nil)

// StartSession creates a new provably fair session with Dual Commitment Protocol
// thetaCommitment is SHA256(theta_seed) - client's commitment sent BEFORE seeing server_seed
// Client seed is now provided per-spin, not per-session
//
// Dual Commitment Protocol flow:
// 1. Client generates theta_seed, computes theta_commitment = SHA256(theta_seed)
// 2. Client sends theta_commitment to server (before seeing server_seed)
// 3. Server generates server_seed ONLY AFTER receiving theta_commitment
// 4. Server responds with server_seed_hash (commitment)
// 5. On first spin, client reveals theta_seed
// 6. Server verifies: SHA256(theta_seed) === theta_commitment
// 7. Game uses both seeds for RNG - neither party could bias the result
func (s *ProvablyFairService) StartSession(
	ctx context.Context,
	playerID, gameSessionID uuid.UUID,
	thetaCommitment string,
) (*provablyfair.StartSessionResult, error) {
	log := s.logger.WithTraceContext(ctx)

	// Check if player already has an active PF session
	existingState, _ := s.cache.GetSessionStateByPlayer(ctx, playerID)
	if existingState != nil && existingState.Status == provablyfair.SessionStatusActive {
		log.Warn().
			Str("player_id", playerID.String()).
			Str("existing_session_id", existingState.SessionID.String()).
			Msg("Player already has active PF session")
		return nil, provablyfair.ErrSessionAlreadyActive
	}

	// Validate theta_commitment (should be 64 hex chars = 256 bits)
	if thetaCommitment != "" && len(thetaCommitment) != 64 {
		log.Warn().
			Str("theta_commitment", thetaCommitment).
			Msg("Invalid theta_commitment length, ignoring")
		thetaCommitment = ""
	}

	// Generate server seed (256-bit) ONLY AFTER receiving theta_commitment
	// This is critical for Dual Commitment Protocol security
	serverSeed, err := s.hashGenerator.GenerateServerSeed()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate server seed")
		return nil, fmt.Errorf("failed to generate server seed: %w", err)
	}

	// Calculate server seed hash (commitment shown to player)
	serverSeedHash := s.hashGenerator.HashServerSeed(serverSeed)

	// Encrypt server seed for DB storage (recovery if Redis is lost)
	encryptedServerSeed, err := s.encryptor.Encrypt(serverSeed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encrypt server seed")
		return nil, fmt.Errorf("failed to encrypt server seed: %w", err)
	}

	// Create session ID
	sessionID := uuid.New()

	// Create PF session model for DB (includes encrypted server seed for recovery)
	// No client_seed here - it's provided per-spin
	pfSession := &provablyfair.PFSession{
		ID:                  sessionID,
		PlayerID:            playerID,
		GameSessionID:       gameSessionID,
		ServerSeedHash:      serverSeedHash,
		EncryptedServerSeed: encryptedServerSeed, // AES-256-GCM encrypted
		NonceStart:          1,
		LastNonce:           0,
		LastSpinHash:        "",
		Status:              provablyfair.SessionStatusActive,
		CreatedAt:           time.Now().UTC(),
		// Dual Commitment Protocol
		ThetaCommitment: thetaCommitment,
		ThetaSeed:       "", // Will be revealed on first spin
		ThetaVerified:   false,
	}

	// Save to DB (commit - includes encrypted seed for recovery)
	if err := s.repo.CreateSession(ctx, pfSession); err != nil {
		log.Error().Err(err).Msg("Failed to create PF session in DB")
		return nil, fmt.Errorf("failed to create PF session: %w", err)
	}

	// Create session state for Redis (includes plaintext seed for fast access)
	// First spin's prevSpinHash combines server_seed_hash and theta_commitment (if Dual Commitment)
	// This ensures both server and client commitments are included in the RNG chain
	initialPrevSpinHash := s.hashGenerator.GenerateInitialPrevSpinHash(serverSeedHash, thetaCommitment)

	sessionState := &provablyfair.PFSessionState{
		SessionID:      sessionID,
		PlayerID:       playerID,
		GameSessionID:  gameSessionID,
		ServerSeed:     serverSeed, // Plaintext - in Redis for fast access
		ServerSeedHash: serverSeedHash,
		LastSpinHash:   initialPrevSpinHash, // SHA256(server_seed_hash + theta_commitment) or server_seed_hash
		Nonce:          0,
		Status:         provablyfair.SessionStatusActive,
		UpdatedAt:      time.Now().UTC(),
		// Dual Commitment Protocol
		ThetaCommitment: thetaCommitment,
		ThetaSeed:       "", // Will be revealed on first spin
		ThetaVerified:   false,
	}

	// Save to Redis
	if err := s.cache.SetSessionState(ctx, sessionState); err != nil {
		log.Error().Err(err).Msg("Failed to cache PF session state")
		// Don't fail - we can recover from DB using encrypted seed
	}

	log.Info().
		Str("session_id", sessionID.String()).
		Str("player_id", playerID.String()).
		Str("server_seed_hash", serverSeedHash).
		Bool("has_theta_commitment", thetaCommitment != "").
		Msg("PF session started with Dual Commitment Protocol")

	return &provablyfair.StartSessionResult{
		SessionID:      sessionID,
		ServerSeedHash: serverSeedHash,
		NonceStart:     1,
	}, nil
}

// GetSessionState retrieves the current session state for a spin
func (s *ProvablyFairService) GetSessionState(ctx context.Context, gameSessionID uuid.UUID) (*provablyfair.PFSessionState, error) {
	log := s.logger.WithTraceContext(ctx)

	// Try Redis first
	state, err := s.cache.GetSessionStateByGameSession(ctx, gameSessionID)
	if err == nil && state != nil {
		return state, nil
	}

	// Fallback to DB if not in cache (session recovery scenario)
	log.Warn().
		Str("game_session_id", gameSessionID.String()).
		Msg("PF session state not found in cache, recovering from DB")

	session, dbErr := s.repo.GetActiveSessionByGameSession(ctx, gameSessionID)
	if dbErr != nil {
		return nil, provablyfair.ErrStateNotFound
	}

	// Decrypt server seed from DB
	serverSeed, err := s.encryptor.Decrypt(session.EncryptedServerSeed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decrypt server seed from DB")
		return nil, fmt.Errorf("failed to recover server seed: %w", err)
	}

	// Rebuild session state (no client_seed - it's per-spin now)
	recoveredState := &provablyfair.PFSessionState{
		SessionID:      session.ID,
		PlayerID:       session.PlayerID,
		GameSessionID:  session.GameSessionID,
		ServerSeed:     serverSeed,
		ServerSeedHash: session.ServerSeedHash,
		LastSpinHash:   session.LastSpinHash,
		Nonce:          session.LastNonce,
		Status:         session.Status,
		UpdatedAt:      time.Now().UTC(),
		// Dual Commitment Protocol fields
		ThetaCommitment: session.ThetaCommitment,
		ThetaSeed:       session.ThetaSeed,
		ThetaVerified:   session.ThetaVerified,
	}

	// If no spins yet, calculate initial prevSpinHash using Dual Commitment if present
	if recoveredState.LastSpinHash == "" {
		recoveredState.LastSpinHash = s.hashGenerator.GenerateInitialPrevSpinHash(
			session.ServerSeedHash, session.ThetaCommitment)
	}

	// Re-cache the recovered state
	if err := s.cache.SetSessionState(ctx, recoveredState); err != nil {
		log.Warn().Err(err).Msg("Failed to re-cache recovered session state")
		// Continue anyway, we have the state
	}

	log.Info().
		Str("session_id", session.ID.String()).
		Int64("nonce", recoveredState.Nonce).
		Msg("PF session state recovered from DB")

	return recoveredState, nil
}

// RecordSpin records a spin and returns the spin hash
// Client seed is now provided per-spin in input.ClientSeed
//
// Dual Commitment Protocol: On first spin (nonce=1), client must reveal theta_seed
// and server verifies SHA256(theta_seed) === theta_commitment
func (s *ProvablyFairService) RecordSpin(
	ctx context.Context,
	input *provablyfair.RecordSpinInput,
) (*provablyfair.SpinResult, error) {
	log := s.logger.WithTraceContext(ctx)

	// Get session state (from Redis or recovered from DB)
	state, err := s.GetSessionState(ctx, input.GameSessionID)
	if err != nil {
		log.Error().Err(err).Str("game_session_id", input.GameSessionID.String()).Msg("Failed to get PF session state")
		return nil, err
	}

	if state.Status != provablyfair.SessionStatusActive {
		return nil, provablyfair.ErrSessionInactive
	}

	// Calculate new nonce
	newNonce := state.Nonce + 1

	// Dual Commitment Protocol: Persist theta verification to DB on first spin
	// Note: The actual verification happens in GetHKDFStreamRNG BEFORE RNG generation
	// Here we just persist the verified theta_seed to DB for recovery purposes
	if newNonce == 1 && state.ThetaCommitment != "" && input.ThetaSeed != "" {
		// Update DB with verified theta_seed (verification already done in GetHKDFStreamRNG)
		if err := s.updateDBThetaVerified(ctx, state.SessionID, input.ThetaSeed); err != nil {
			log.Error().Err(err).Msg("Failed to persist theta verification to DB")
			// Don't fail - verification already passed in GetHKDFStreamRNG, continue with spin
		}
		// Update state in memory (already set in GetHKDFStreamRNG but may not be persisted)
		state.ThetaSeed = input.ThetaSeed
		state.ThetaVerified = true
	}

	// Validate client seed
	clientSeed := input.ClientSeed
	if clientSeed == "" {
		// Generate fallback client seed if not provided
		clientSeed, err = s.hashGenerator.GenerateClientSeed()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate fallback client seed")
			return nil, fmt.Errorf("failed to generate client seed: %w", err)
		}
		log.Warn().Msg("Client seed not provided, using generated fallback")
	}

	// Get previous spin hash (serverSeedHash for first spin)
	prevSpinHash := state.LastSpinHash

	// Generate spin hash using per-spin client seed
	spinHash := s.hashGenerator.GenerateSpinHash(
		prevSpinHash,
		state.ServerSeed,
		clientSeed, // Per-spin client seed
		newNonce,
	)

	// Create spin log for DB (append-only) - includes client_seed per spin
	spinLog := &provablyfair.SpinLog{
		ID:                uuid.New(),
		PFSessionID:       state.SessionID,
		SpinID:            input.SpinID,
		SpinIndex:         newNonce,
		Nonce:             newNonce,
		ClientSeed:        clientSeed, // Store per-spin client seed
		SpinHash:          spinHash,
		PrevSpinHash:      prevSpinHash,
		ReelPositions:     provablyfair.IntSlice(input.ReelPositions),
		ReelStripConfigID: input.ReelStripConfigID,
		GameMode:          input.GameMode,
		IsFreeSpin:        input.IsFreeSpin,
		CreatedAt:         time.Now().UTC(),
	}

	// Save spin log to DB
	if err := s.repo.CreateSpinLog(ctx, spinLog); err != nil {
		log.Error().Err(err).Msg("Failed to create spin log")
		return nil, fmt.Errorf("failed to create spin log: %w", err)
	}

	// Update session state in Redis
	state.Nonce = newNonce
	state.LastSpinHash = spinHash
	state.UpdatedAt = time.Now().UTC()

	if err := s.cache.UpdateSessionState(ctx, state); err != nil {
		log.Error().Err(err).Msg("Failed to update PF session state in cache")
		// Don't fail - spin was already recorded
	}

	// Update DB session state
	if err := s.updateDBSessionState(ctx, state.SessionID, newNonce, spinHash); err != nil {
		log.Error().Err(err).Msg("Failed to update PF session in DB")
		// Don't fail - spin was already recorded
	}

	log.Debug().
		Str("session_id", state.SessionID.String()).
		Int64("nonce", newNonce).
		Str("spin_hash", spinHash).
		Interface("reel_positions", input.ReelPositions).
		Bool("theta_verified", state.ThetaVerified).
		Msg("Spin recorded in PF system")

	return &provablyfair.SpinResult{
		SpinIndex:    newNonce,
		Nonce:        newNonce,
		SpinHash:     spinHash,
		PrevSpinHash: prevSpinHash,
	}, nil
}

// updateDBThetaVerified updates the theta verification status in the database
func (s *ProvablyFairService) updateDBThetaVerified(ctx context.Context, sessionID uuid.UUID, thetaSeed string) error {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ThetaSeed = thetaSeed
	session.ThetaVerified = true

	return s.repo.UpdateSession(ctx, session)
}

// EndSession ends the session and reveals the server seed
func (s *ProvablyFairService) EndSession(ctx context.Context, gameSessionID uuid.UUID) (*provablyfair.EndSessionResult, error) {
	log := s.logger.WithTraceContext(ctx)

	// Get session state (from Redis or recovered from DB)
	state, err := s.GetSessionState(ctx, gameSessionID)
	if err != nil {
		log.Error().Err(err).Str("game_session_id", gameSessionID.String()).Msg("Failed to get PF session state")
		return nil, err
	}

	if state.Status == provablyfair.SessionStatusEnded {
		return nil, provablyfair.ErrSessionAlreadyEnded
	}

	// Get all spin logs for verification data
	spinLogs, err := s.repo.GetSpinLogsBySession(ctx, state.SessionID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get spin logs")
		return nil, fmt.Errorf("failed to get spin logs: %w", err)
	}

	// Convert to verification format - each spin has its own client_seed
	spins := make([]provablyfair.SpinVerification, len(spinLogs))
	for i, spinLog := range spinLogs {
		spins[i] = provablyfair.SpinVerification{
			SpinIndex:         spinLog.SpinIndex,
			Nonce:             spinLog.Nonce,
			ClientSeed:        spinLog.ClientSeed, // Per-spin client seed
			SpinHash:          spinLog.SpinHash,
			PrevSpinHash:      spinLog.PrevSpinHash,
			ReelPositions:     []int(spinLog.ReelPositions),
			ReelStripConfigID: spinLog.ReelStripConfigID,
			GameMode:          spinLog.GameMode,
			IsFreeSpin:        spinLog.IsFreeSpin,
		}
	}

	// Create session audit (reveals server seed)
	audit := &provablyfair.SessionAudit{
		ID:                  uuid.New(),
		PFSessionID:         state.SessionID,
		ServerSeedPlaintext: state.ServerSeed,
		ServerSeedHash:      state.ServerSeedHash,
		TotalSpins:          state.Nonce,
		RevealedAt:          time.Now().UTC(),
	}

	if err := s.repo.CreateSessionAudit(ctx, audit); err != nil {
		log.Error().Err(err).Msg("Failed to create session audit")
		return nil, fmt.Errorf("failed to create session audit: %w", err)
	}

	// End session in DB
	if err := s.repo.EndSession(ctx, state.SessionID); err != nil {
		log.Error().Err(err).Msg("Failed to end PF session in DB")
		return nil, fmt.Errorf("failed to end session: %w", err)
	}

	// Delete from Redis
	if err := s.cache.DeleteSessionState(ctx, state.SessionID); err != nil {
		log.Warn().Err(err).Msg("Failed to delete PF session state from cache")
		// Don't fail - session was already ended
	}

	log.Info().
		Str("session_id", state.SessionID.String()).
		Int64("total_spins", state.Nonce).
		Msg("PF session ended and server seed revealed")

	return &provablyfair.EndSessionResult{
		SessionID:      state.SessionID,
		ServerSeed:     state.ServerSeed,
		ServerSeedHash: state.ServerSeedHash,
		TotalSpins:     state.Nonce,
		Spins:          spins, // Each spin has its own client_seed
	}, nil
}

// GetVerificationData returns all data needed to verify a completed session
// Each spin has its own client_seed stored in spin_logs
func (s *ProvablyFairService) GetVerificationData(ctx context.Context, pfSessionID uuid.UUID) (*provablyfair.VerificationData, error) {
	// Get session from DB
	session, err := s.repo.GetSessionByID(ctx, pfSessionID)
	if err != nil {
		return nil, err
	}

	// Session must be ended to get verification data
	if session.Status != provablyfair.SessionStatusEnded {
		return nil, fmt.Errorf("session must be ended to retrieve verification data")
	}

	// Get audit (contains revealed server seed)
	audit, err := s.repo.GetSessionAudit(ctx, pfSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session audit: %w", err)
	}

	// Get all spin logs
	spinLogs, err := s.repo.GetSpinLogsBySession(ctx, pfSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get spin logs: %w", err)
	}

	// Convert to verification format - each spin has its own client_seed
	spins := make([]provablyfair.SpinVerification, len(spinLogs))
	for i, spinLog := range spinLogs {
		spins[i] = provablyfair.SpinVerification{
			SpinIndex:         spinLog.SpinIndex,
			Nonce:             spinLog.Nonce,
			ClientSeed:        spinLog.ClientSeed, // Per-spin client seed
			SpinHash:          spinLog.SpinHash,
			PrevSpinHash:      spinLog.PrevSpinHash,
			ReelPositions:     []int(spinLog.ReelPositions),
			ReelStripConfigID: spinLog.ReelStripConfigID,
			GameMode:          spinLog.GameMode,
			IsFreeSpin:        spinLog.IsFreeSpin,
		}
	}

	return &provablyfair.VerificationData{
		SessionID:      pfSessionID,
		ServerSeed:     audit.ServerSeedPlaintext,
		ServerSeedHash: session.ServerSeedHash,
		Spins:          spins, // Each spin has its own client_seed
	}, nil
}

// VerifySession verifies a completed session's hash chain
// Each spin has its own client_seed stored in spin_logs
func (s *ProvablyFairService) VerifySession(ctx context.Context, pfSessionID uuid.UUID, serverSeed string) (bool, error) {
	// Get session from DB
	session, err := s.repo.GetSessionByID(ctx, pfSessionID)
	if err != nil {
		return false, err
	}

	// Verify server seed hash
	calculatedHash := s.hashGenerator.HashServerSeed(serverSeed)
	if calculatedHash != session.ServerSeedHash {
		return false, provablyfair.ErrInvalidServerSeed
	}

	// Get all spin logs
	spinLogs, err := s.repo.GetSpinLogsBySession(ctx, pfSessionID)
	if err != nil {
		return false, fmt.Errorf("failed to get spin logs: %w", err)
	}

	// Convert to verification format - each spin has its own client_seed
	spins := make([]provablyfair.SpinVerification, len(spinLogs))
	for i, spinLog := range spinLogs {
		spins[i] = provablyfair.SpinVerification{
			SpinIndex:         spinLog.SpinIndex,
			Nonce:             spinLog.Nonce,
			ClientSeed:        spinLog.ClientSeed, // Per-spin client seed
			SpinHash:          spinLog.SpinHash,
			PrevSpinHash:      spinLog.PrevSpinHash,
			ReelPositions:     []int(spinLog.ReelPositions),
			ReelStripConfigID: spinLog.ReelStripConfigID,
			GameMode:          spinLog.GameMode,
			IsFreeSpin:        spinLog.IsFreeSpin,
		}
	}

	// Verify hash chain - uses serverSeedHash as first prevSpinHash
	return s.hashGenerator.VerifyHashChain(
		serverSeed,
		session.ServerSeedHash,
		spins,
	)
}

// updateDBSessionState updates the session state in the database
func (s *ProvablyFairService) updateDBSessionState(ctx context.Context, sessionID uuid.UUID, nonce int64, lastSpinHash string) error {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastNonce = nonce
	session.LastSpinHash = lastSpinHash

	return s.repo.UpdateSession(ctx, session)
}

// GetHKDFStreamRNG returns an HKDF-based RNG implementing RFC 5869
// This provides cryptographic domain separation - each reel has its own
// independent key derived from the master seed, ensuring no correlation
// between reel outcomes.
//
// Benefits:
// - Each reel key is cryptographically independent
// - Standard RFC 5869 compliance for audit/compliance
// - Domain separation via 'info' parameter
// - Extensible to any number of keys needed
//
// Dual Commitment Protocol:
// - thetaSeed is required on first spin if theta_commitment was provided during StartSession
// - Server verifies SHA256(thetaSeed) === theta_commitment BEFORE generating RNG
// - This ensures client cannot change their commitment after seeing server_seed
func (s *ProvablyFairService) GetHKDFStreamRNG(ctx context.Context, gameSessionID uuid.UUID, clientSeed, thetaSeed string) (*rng.HKDFStreamRNG, string, error) {
	log := s.logger.WithTraceContext(ctx)

	// Get session state (from Redis or recovered from DB)
	state, err := s.GetSessionState(ctx, gameSessionID)
	if err != nil {
		return nil, "", err
	}

	// Calculate the nonce for the NEXT spin (current nonce + 1)
	nextNonce := state.Nonce + 1

	// Dual Commitment Protocol: Verify theta_seed on first spin BEFORE generating RNG
	// This is critical - verification must happen BEFORE RNG generation to prevent
	// the client from seeing RNG output and then providing a different theta_seed
	if nextNonce == 1 && state.ThetaCommitment != "" {
		// First spin with theta_commitment - verify theta_seed
		if thetaSeed == "" {
			log.Error().
				Str("session_id", state.SessionID.String()).
				Msg("Dual Commitment: theta_seed required on first spin but not provided")
			return nil, "", provablyfair.ErrThetaSeedRequired
		}

		// Verify theta_seed matches theta_commitment
		expectedCommitment := s.hashGenerator.HashServerSeed(thetaSeed) // Same SHA256 function
		if expectedCommitment != state.ThetaCommitment {
			log.Error().
				Str("session_id", state.SessionID.String()).
				Str("provided_theta_seed", thetaSeed[:min(8, len(thetaSeed))]+"...").
				Str("expected_commitment", state.ThetaCommitment[:8]+"...").
				Str("calculated_commitment", expectedCommitment[:8]+"...").
				Msg("Dual Commitment: theta_seed verification failed")
			return nil, "", provablyfair.ErrThetaVerificationFailed
		}

		// Mark as verified in state (will be persisted in RecordSpin)
		state.ThetaSeed = thetaSeed
		state.ThetaVerified = true

		log.Info().
			Str("session_id", state.SessionID.String()).
			Str("theta_seed", thetaSeed[:8]+"...").
			Msg("Dual Commitment: theta_seed verified successfully BEFORE RNG generation")
	}

	// Validate/generate client seed
	if clientSeed == "" {
		// Generate fallback client seed if not provided
		clientSeed, err = s.hashGenerator.GenerateClientSeed()
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate fallback client seed")
			return nil, "", fmt.Errorf("failed to generate client seed: %w", err)
		}
		log.Warn().Msg("Client seed not provided for GetHKDFStreamRNG, using generated fallback")
	}

	// Create HKDF-based RNG using server seed, client seed, nonce, and prevSpinHash
	// This implements RFC 5869 HKDF for per-reel key derivation
	// prevSpinHash is REQUIRED to maintain hash chain integrity:
	// - Each spin's RNG depends on ALL previous spins (hash chain)
	// - Ensures entropy accumulation across the session
	// - Server cannot pre-compute outcomes for multiple spins
	hkdfRNG, err := rng.NewHKDFStreamRNG(state.ServerSeed, clientSeed, nextNonce, state.LastSpinHash)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create HKDF stream RNG: %w", err)
	}

	// Get the spin hash for logging/verification
	spinHash := hkdfRNG.GetSpinHash()

	log.Debug().
		Str("session_id", state.SessionID.String()).
		Int64("nonce", nextNonce).
		Str("spin_hash", spinHash[:16]+"..."). // Truncate for logging
		Msg("Created HKDF stream RNG for provably fair spin")

	return hkdfRNG, spinHash, nil
}

// VerifySpin verifies a single spin's hash
// This is a stateless verification - no database access required
func (s *ProvablyFairService) VerifySpin(
	ctx context.Context,
	input *provablyfair.VerifySpinInput,
) (*provablyfair.VerifySpinResult, error) {
	// Calculate expected spin hash
	expectedSpinHash := s.hashGenerator.GenerateSpinHash(
		input.PrevSpinHash,
		input.ServerSeed,
		input.ClientSeed,
		input.Nonce,
	)

	// Calculate server seed hash for reference
	serverSeedHash := s.hashGenerator.HashServerSeed(input.ServerSeed)

	// Check if the provided spin hash matches
	valid := expectedSpinHash == input.SpinHash

	return &provablyfair.VerifySpinResult{
		Valid:            valid,
		ExpectedSpinHash: expectedSpinHash,
		ServerSeedHash:   serverSeedHash,
	}, nil
}

// VerifySpinWithReel verifies a spin hash and its reel positions using HKDF
// This implements RFC 5869 HKDF with the same "stream:N" domain pattern as the game engine
func (s *ProvablyFairService) VerifySpinWithReel(
	ctx context.Context,
	input *provablyfair.VerifySpinWithReelInput,
) (*provablyfair.VerifySpinWithReelResult, error) {
	// Lookup reel strip config to get strip lengths
	configSet, err := s.reelstripRepo.GetSetByConfigID(ctx, input.ReelStripConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reel strip config: %w", err)
	}

	// Calculate expected spin hash (for backward compatibility logging)
	expectedSpinHash := s.hashGenerator.GenerateSpinHash(
		input.PrevSpinHash,
		input.ServerSeed,
		input.ClientSeed,
		input.Nonce,
	)

	// Calculate server seed hash for reference
	serverSeedHash := s.hashGenerator.HashServerSeed(input.ServerSeed)

	// Check if the provided spin hash matches
	spinHashValid := expectedSpinHash == input.SpinHash

	// Create HKDFStreamRNG for reel position verification
	// IMPORTANT: Must use HKDFStreamRNG.Int() to match game engine behavior
	// Game engine uses reels.GenerateGrid() which calls rng.Int() for each reel
	// This uses domain "stream:N:M" pattern, NOT "reel:N:M"
	streamRNG, err := rng.NewHKDFStreamRNG(input.ServerSeed, input.ClientSeed, input.Nonce, input.PrevSpinHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create HKDF Stream RNG for verification: %w", err)
	}

	// Generate expected reel positions using the same method as game engine
	// Game engine: reels.GenerateGrid() calls rng.Int(stripLength) for each reel
	// This uses sequential "stream:0", "stream:1", etc. domains
	expectedReelPositions := make([]int, 5)
	for i := 0; i < 5; i++ {
		stripLength := len(configSet.Strips[i].StripData)
		pos, err := streamRNG.Int(stripLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate expected reel position %d: %w", i, err)
		}
		expectedReelPositions[i] = pos
	}

	// Check if reel positions match
	reelPositionsValid := len(input.ReelPositions) == 5
	if reelPositionsValid {
		for i := 0; i < 5; i++ {
			if input.ReelPositions[i] != expectedReelPositions[i] {
				reelPositionsValid = false
				break
			}
		}
	}

	return &provablyfair.VerifySpinWithReelResult{
		Valid:                 spinHashValid && reelPositionsValid,
		SpinHashValid:         spinHashValid,
		ReelPositionsValid:    reelPositionsValid,
		ExpectedSpinHash:      expectedSpinHash,
		ExpectedReelPositions: expectedReelPositions,
		ServerSeedHash:        serverSeedHash,
	}, nil
}

// GetActiveSessionByPlayer retrieves the active PF session state by player ID
func (s *ProvablyFairService) GetActiveSessionByPlayer(ctx context.Context, playerID uuid.UUID) (*provablyfair.PFSessionState, error) {
	log := s.logger.WithTraceContext(ctx)

	// Try Redis first
	state, err := s.cache.GetSessionStateByPlayer(ctx, playerID)
	if err == nil && state != nil && state.Status == provablyfair.SessionStatusActive {
		return state, nil
	}

	// Fallback to DB if not in cache
	log.Debug().
		Str("player_id", playerID.String()).
		Msg("PF session state not found in cache, trying DB")

	session, dbErr := s.repo.GetActiveSessionByPlayer(ctx, playerID)
	if dbErr != nil {
		return nil, provablyfair.ErrStateNotFound
	}

	// Decrypt server seed from DB
	serverSeed, err := s.encryptor.Decrypt(session.EncryptedServerSeed)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decrypt server seed from DB")
		return nil, fmt.Errorf("failed to recover server seed: %w", err)
	}

	// Rebuild session state
	recoveredState := &provablyfair.PFSessionState{
		SessionID:      session.ID,
		PlayerID:       session.PlayerID,
		GameSessionID:  session.GameSessionID,
		ServerSeed:     serverSeed,
		ServerSeedHash: session.ServerSeedHash,
		LastSpinHash:   session.LastSpinHash,
		Nonce:          session.LastNonce,
		Status:         session.Status,
		UpdatedAt:      time.Now().UTC(),
	}

	// If no spins yet, last spin hash is server_seed_hash
	if recoveredState.LastSpinHash == "" {
		recoveredState.LastSpinHash = session.ServerSeedHash
	}

	// Re-cache the recovered state
	if err := s.cache.SetSessionState(ctx, recoveredState); err != nil {
		log.Warn().Err(err).Msg("Failed to re-cache recovered session state")
	}

	return recoveredState, nil
}

// VerifyActiveSpin verifies a spin in an active session using server's stored server_seed
// Uses HKDF (RFC 5869) for per-reel key derivation verification
func (s *ProvablyFairService) VerifyActiveSpin(
	ctx context.Context,
	gameSessionID uuid.UUID,
	input *provablyfair.VerifyActiveSpinInput,
) (*provablyfair.VerifyActiveSpinResult, error) {
	// Get session state (contains server_seed)
	state, err := s.GetSessionState(ctx, gameSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session state: %w", err)
	}

	// Get prev_spin_hash: use user-provided value if available, otherwise calculate from DB
	var prevSpinHash string
	if input.PrevSpinHash != "" {
		// User provided prev_spin_hash - use it directly for verification
		prevSpinHash = input.PrevSpinHash
	} else if input.Nonce == 1 {
		// First spin: prev_spin_hash = SHA256(server_seed_hash + theta_commitment)
		// For Dual Commitment Protocol, both commitments are combined
		// If no theta_commitment (legacy), falls back to just server_seed_hash
		prevSpinHash = s.hashGenerator.GenerateInitialPrevSpinHash(state.ServerSeedHash, state.ThetaCommitment)
	} else {
		// Get previous spin's hash from spin_logs (nonce-1 = spinIndex-1)
		prevSpin, err := s.repo.GetSpinLogByIndex(ctx, state.SessionID, input.Nonce-1)
		if err != nil {
			return nil, fmt.Errorf("failed to get previous spin: %w", err)
		}
		prevSpinHash = prevSpin.SpinHash
	}

	// Calculate expected spin hash
	expectedSpinHash := s.hashGenerator.GenerateSpinHash(
		prevSpinHash,
		state.ServerSeed,
		input.ClientSeed,
		input.Nonce,
	)

	// Check if the provided spin hash matches
	spinHashValid := expectedSpinHash == input.SpinHash

	result := &provablyfair.VerifyActiveSpinResult{
		Valid:            spinHashValid,
		SpinHashValid:    spinHashValid,
		ExpectedSpinHash: expectedSpinHash,
		ServerSeedHash:   state.ServerSeedHash,
		PrevSpinHash:     prevSpinHash, // For debugging - shows what prev_spin_hash was used
	}

	// Optionally verify reel positions using HKDFStreamRNG
	if len(input.ReelPositions) == 5 && input.ReelStripConfigID != uuid.Nil {
		// Lookup reel strip config
		configSet, err := s.reelstripRepo.GetSetByConfigID(ctx, input.ReelStripConfigID)
		if err != nil {
			return nil, fmt.Errorf("failed to get reel strip config: %w", err)
		}

		// Create HKDFStreamRNG for reel position verification
		// IMPORTANT: Must use HKDFStreamRNG.Int() to match game engine behavior
		// Game engine uses reels.GenerateGrid() which calls rng.Int() for each reel
		// This uses domain "stream:N:M" pattern, NOT "reel:N:M"
		streamRNG, err := rng.NewHKDFStreamRNG(state.ServerSeed, input.ClientSeed, input.Nonce, prevSpinHash)
		if err != nil {
			return nil, fmt.Errorf("failed to create HKDF Stream RNG: %w", err)
		}

		// Generate expected reel positions using the same method as game engine
		// Game engine: reels.GenerateGrid() calls rng.Int(stripLength) for each reel
		expectedReelPositions := make([]int, 5)
		for i := 0; i < 5; i++ {
			stripLength := len(configSet.Strips[i].StripData)
			pos, err := streamRNG.Int(stripLength)
			if err != nil {
				return nil, fmt.Errorf("failed to generate expected reel position %d: %w", i, err)
			}
			expectedReelPositions[i] = pos
		}

		// Check if reel positions match
		reelPositionsValid := true
		for i := 0; i < 5; i++ {
			if input.ReelPositions[i] != expectedReelPositions[i] {
				reelPositionsValid = false
				break
			}
		}

		result.ReelPositionsValid = &reelPositionsValid
		result.ExpectedReelPositions = expectedReelPositions
		result.Valid = spinHashValid && reelPositionsValid
	}

	return result, nil
}
