package provablyfair

import "errors"

var (
	// Session errors
	ErrSessionNotFound      = errors.New("provably fair session not found")
	ErrSessionAlreadyActive = errors.New("player already has an active provably fair session")
	ErrSessionAlreadyEnded  = errors.New("provably fair session has already ended")
	ErrSessionInactive      = errors.New("provably fair session is not active")

	// Nonce errors
	ErrInvalidNonce     = errors.New("invalid nonce: must be sequential")
	ErrNonceMismatch    = errors.New("nonce mismatch: expected next sequential value")

	// Hash chain errors
	ErrHashChainBroken   = errors.New("hash chain verification failed")
	ErrInvalidGenesisHash = errors.New("genesis hash verification failed")
	ErrInvalidServerSeed = errors.New("server seed hash verification failed")

	// Verification errors
	ErrVerificationFailed = errors.New("provably fair verification failed")
	ErrSpinNotFound       = errors.New("spin log not found")

	// State errors
	ErrStateNotFound    = errors.New("session state not found in cache")
	ErrStateCorrupted   = errors.New("session state is corrupted")

	// Dual Commitment Protocol errors
	ErrThetaSeedRequired        = errors.New("theta_seed is required on first spin when theta_commitment was provided")
	ErrThetaVerificationFailed  = errors.New("theta_seed verification failed: SHA256(theta_seed) does not match theta_commitment")
	ErrThetaAlreadyVerified     = errors.New("theta_seed has already been verified")
)
