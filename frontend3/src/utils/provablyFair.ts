/**
 * Provably Fair Utility
 * Client-side utilities for provably fair gaming verification
 *
 * This module implements:
 * 1. Dual Commitment Protocol - Both client and server commit before seeing each other's seeds
 * 2. RFC 5869 HKDF for per-reel key derivation
 * 3. Cryptographic domain separation between reel outcomes
 *
 * Security Flow (Dual Commitment):
 * 1. Client generates theta_seed, computes theta_commitment = SHA256(theta_seed)
 * 2. Client sends theta_commitment to server (before seeing server_seed)
 * 3. Server generates server_seed, computes server_seed_hash = SHA256(server_seed)
 * 4. Server responds with server_seed_hash (commitment)
 * 5. On first spin, client reveals theta_seed
 * 6. Server verifies: SHA256(theta_seed) === theta_commitment
 * 7. Game uses both seeds for RNG - neither party could bias the result
 */

// ============================================================================
// Dual Commitment Protocol - Theta Seed Management
// ============================================================================

/**
 * Generate a cryptographically secure theta seed (32 bytes / 256 bits)
 * This is the client's secret seed that will be committed before seeing server's seed
 */
export function generateThetaSeed(): string {
  const array = new Uint8Array(32) // 256 bits for maximum security
  crypto.getRandomValues(array)
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('')
}

/**
 * Compute theta commitment (SHA256 hash of theta_seed)
 * This commitment is sent to server BEFORE server generates its seed
 */
export async function computeThetaCommitment(thetaSeed: string): Promise<string> {
  return sha256(thetaSeed)
}

/**
 * Verify theta seed matches its commitment
 * Used by server (and can be used by client for self-verification)
 */
export async function verifyThetaCommitment(
  thetaSeed: string,
  thetaCommitment: string
): Promise<boolean> {
  const computedCommitment = await sha256(thetaSeed)
  return computedCommitment === thetaCommitment
}

/**
 * Generate a cryptographically secure random client seed (per-spin)
 * Uses Web Crypto API for secure randomness
 * Note: This is different from theta_seed which is per-session
 */
export function generateClientSeed(): string {
  const array = new Uint8Array(16) // 128 bits per-spin seed
  crypto.getRandomValues(array)
  return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('')
}

/**
 * Compute SHA-256 hash of input string
 * Used for verifying spin hashes
 */
export async function sha256(message: string): Promise<string> {
  const encoder = new TextEncoder()
  const data = encoder.encode(message)
  const hashBuffer = await crypto.subtle.digest('SHA-256', data)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map(byte => byte.toString(16).padStart(2, '0')).join('')
}

/**
 * Verify a single spin hash
 * Formula: SHA256(prev_spin_hash + server_seed + client_seed + nonce)
 */
export async function verifySpinHash(
  prevSpinHash: string,
  serverSeed: string,
  clientSeed: string,
  nonce: number,
  expectedHash: string
): Promise<boolean> {
  const input = `${prevSpinHash}${serverSeed}${clientSeed}${nonce}`
  const computedHash = await sha256(input)
  return computedHash === expectedHash
}

/**
 * Verify server seed commitment
 * Server seed hash should match SHA256(server_seed)
 */
export async function verifyServerSeedCommitment(
  serverSeed: string,
  serverSeedHash: string
): Promise<boolean> {
  const computedHash = await sha256(serverSeed)
  return computedHash === serverSeedHash
}

/**
 * Verify entire session hash chain
 * Returns verification result with details
 */
export interface SpinVerificationData {
  nonce: number
  clientSeed: string
  spinHash: string
  prevSpinHash: string
  reelPositions: number[]
}

export interface SessionVerificationResult {
  isValid: boolean
  serverSeedValid: boolean
  spinsVerified: number
  spinsFailed: number
  failedSpins: number[] // Nonces of failed spins
}

export async function verifySession(
  serverSeed: string,
  serverSeedHash: string,
  spins: SpinVerificationData[]
): Promise<SessionVerificationResult> {
  const result: SessionVerificationResult = {
    isValid: true,
    serverSeedValid: false,
    spinsVerified: 0,
    spinsFailed: 0,
    failedSpins: [],
  }

  // Verify server seed commitment
  result.serverSeedValid = await verifyServerSeedCommitment(serverSeed, serverSeedHash)
  if (!result.serverSeedValid) {
    result.isValid = false
    return result
  }

  // Verify each spin in the chain
  for (const spin of spins) {
    const isValid = await verifySpinHash(
      spin.prevSpinHash,
      serverSeed,
      spin.clientSeed,
      spin.nonce,
      spin.spinHash
    )

    if (isValid) {
      result.spinsVerified++
    } else {
      result.spinsFailed++
      result.failedSpins.push(spin.nonce)
      result.isValid = false
    }
  }

  return result
}

/**
 * Store for managing provably fair session state
 * Extended for Dual Commitment Protocol
 */
export interface ProvablyFairState {
  // Server commitment (received after client sends theta_commitment)
  serverSeedHash: string | null
  nonceStart: number
  currentNonce: number
  isActive: boolean

  // Dual Commitment Protocol fields
  thetaSeed: string | null // Client's session seed (kept secret until first spin)
  thetaCommitment: string | null // SHA256(theta_seed) - sent to server before seeing server_seed
  thetaRevealed: boolean // True after theta_seed has been revealed to server
}

export function createInitialPFState(): ProvablyFairState {
  return {
    serverSeedHash: null,
    nonceStart: 0,
    currentNonce: 0,
    isActive: false,
    // Dual Commitment Protocol
    thetaSeed: null,
    thetaCommitment: null,
    thetaRevealed: false,
  }
}

// ============================================================================
// HKDF (RFC 5869) Implementation for Per-Reel Key Derivation
// ============================================================================

/**
 * HMAC-SHA256 implementation using Web Crypto API
 */
async function hmacSha256(key: Uint8Array, data: Uint8Array): Promise<Uint8Array> {
  // Create a new ArrayBuffer copy to avoid SharedArrayBuffer issues
  const keyBuffer = new ArrayBuffer(key.length)
  new Uint8Array(keyBuffer).set(key)

  const dataBuffer = new ArrayBuffer(data.length)
  new Uint8Array(dataBuffer).set(data)

  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    keyBuffer,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  )
  const signature = await crypto.subtle.sign('HMAC', cryptoKey, dataBuffer)
  return new Uint8Array(signature)
}

/**
 * HKDF-Extract (RFC 5869 Section 2.2)
 * PRK = HMAC-Hash(salt, IKM)
 */
async function hkdfExtract(salt: Uint8Array, ikm: Uint8Array): Promise<Uint8Array> {
  // If salt is empty, use a zero-filled buffer of hash length (32 for SHA-256)
  const actualSalt = salt.length > 0 ? salt : new Uint8Array(32)
  return hmacSha256(actualSalt, ikm)
}

/**
 * HKDF-Expand (RFC 5869 Section 2.3)
 * Expands PRK to desired length using info as context
 */
async function hkdfExpand(prk: Uint8Array, info: Uint8Array, length: number): Promise<Uint8Array> {
  const hashLen = 32 // SHA-256 output length
  const n = Math.ceil(length / hashLen)

  if (n > 255) {
    throw new Error('HKDF-Expand: requested length too large')
  }

  const okm = new Uint8Array(n * hashLen)
  let t: Uint8Array = new Uint8Array(0)

  for (let i = 1; i <= n; i++) {
    // T(i) = HMAC-Hash(PRK, T(i-1) | info | i)
    const input = new Uint8Array(t.length + info.length + 1)
    input.set(t, 0)
    input.set(info, t.length)
    input[t.length + info.length] = i

    const result = await hmacSha256(prk, input)
    t = new Uint8Array(result)
    okm.set(t, (i - 1) * hashLen)
  }

  return new Uint8Array(okm.buffer, 0, length)
}

/**
 * Full HKDF (Extract-then-Expand)
 */
async function hkdf(
  ikm: Uint8Array,
  salt: Uint8Array,
  info: Uint8Array,
  length: number
): Promise<Uint8Array> {
  const prk = await hkdfExtract(salt, ikm)
  return hkdfExpand(prk, info, length)
}

/**
 * Convert string to Uint8Array (UTF-8 encoding)
 */
function stringToBytes(str: string): Uint8Array {
  return new TextEncoder().encode(str)
}

/**
 * Convert Uint8Array to hex string
 */
function bytesToHex(bytes: Uint8Array): string {
  return Array.from(bytes, byte => byte.toString(16).padStart(2, '0')).join('')
}

/**
 * HKDF Stream RNG for provably fair verification
 * Matches the backend HKDFStreamRNG implementation exactly
 *
 * IMPORTANT: Uses "stream:N" domain pattern (NOT "reel:N")
 * This matches the game engine which calls rng.Int() for each reel
 */
export class HKDFStreamRNG {
  private masterKey: Uint8Array
  private counter: number = 0

  private constructor(masterKey: Uint8Array) {
    this.masterKey = masterKey
  }

  /**
   * Create a new HKDF Stream RNG from server seed, client seed, nonce, and prevSpinHash
   * This matches the backend NewHKDFStreamRNG function exactly
   *
   * @param serverSeed - Server's secret seed (Salt)
   * @param clientSeed - Client's per-spin seed
   * @param nonce - Spin number
   * @param prevSpinHash - Previous spin's hash (or initialPrevSpinHash for first spin)
   */
  static async create(
    serverSeed: string,
    clientSeed: string,
    nonce: number,
    prevSpinHash: string
  ): Promise<HKDFStreamRNG> {
    // IKM = prevSpinHash || clientSeed || nonce (as string)
    // MUST include prevSpinHash for hash chain integrity
    const ikm = stringToBytes(`${prevSpinHash}${clientSeed}${nonce}`)

    // Salt = serverSeed
    const salt = stringToBytes(serverSeed)

    // Info = "spin-master-v1" (must match backend)
    const info = stringToBytes('spin-master-v1')

    // Derive 32-byte master key
    const masterKey = await hkdf(ikm, salt, info, 32)

    return new HKDFStreamRNG(masterKey)
  }

  /**
   * Generate a random integer in range [0, max)
   * Uses "stream:N:M" domain pattern (N = counter, M = attempt for rejection sampling)
   */
  async int(max: number): Promise<number> {
    if (max <= 0) {
      throw new Error('max must be positive')
    }

    const domain = `stream:${this.counter}`
    this.counter++

    // Rejection sampling to eliminate modulo bias
    const maxBigInt = BigInt(max)
    const threshold = (BigInt(2) ** BigInt(64) - maxBigInt) % maxBigInt

    for (let attempt = 0; attempt < 100; attempt++) {
      const info = stringToBytes(`${domain}:${attempt}`)
      const key = await hkdfExpand(this.masterKey, info, 8)

      // Read as uint64 big-endian
      const view = new DataView(key.buffer, key.byteOffset, 8)
      const value = view.getBigUint64(0, false)

      // Reject values that would cause modulo bias
      if (value >= threshold) {
        return Number(value % maxBigInt)
      }
    }

    throw new Error('Rejection sampling failed after 100 attempts')
  }

  /**
   * Get positions for all reels using sequential Int() calls
   * This matches the game engine behavior exactly
   */
  async getAllReelPositions(reelLengths: number[]): Promise<number[]> {
    const positions: number[] = []

    for (const length of reelLengths) {
      const pos = await this.int(length)
      positions.push(pos)
    }

    return positions
  }

  /**
   * Get the master key as hex string (for debugging/logging)
   */
  getSpinHash(): string {
    return bytesToHex(this.masterKey)
  }
}

// Backward compatibility alias
export const HKDFRNG = HKDFStreamRNG

/**
 * Verify all reel positions for a spin using HKDF Stream RNG
 * IMPORTANT: Must include prevSpinHash to match backend
 */
export async function verifyAllReelPositionsHKDF(
  serverSeed: string,
  clientSeed: string,
  nonce: number,
  prevSpinHash: string,
  reelLengths: number[],
  expectedPositions: number[]
): Promise<boolean> {
  if (reelLengths.length !== expectedPositions.length) {
    return false
  }

  const rng = await HKDFStreamRNG.create(serverSeed, clientSeed, nonce, prevSpinHash)
  const actualPositions = await rng.getAllReelPositions(reelLengths)

  for (let i = 0; i < expectedPositions.length; i++) {
    if (actualPositions[i] !== expectedPositions[i]) {
      return false
    }
  }

  return true
}

/**
 * Calculate initial prevSpinHash for first spin (Dual Commitment Protocol)
 * Formula: SHA256(serverSeedHash + thetaCommitment)
 */
export async function calculateInitialPrevSpinHash(
  serverSeedHash: string,
  thetaCommitment: string
): Promise<string> {
  if (!thetaCommitment) {
    // Legacy mode without Dual Commitment - just use serverSeedHash
    return serverSeedHash
  }
  return sha256(serverSeedHash + thetaCommitment)
}

/**
 * Enhanced session verification with HKDF reel position verification
 */
export interface EnhancedSpinVerificationData extends SpinVerificationData {
  reelLengths?: number[] // Required for HKDF verification
}

export interface EnhancedSessionVerificationResult extends SessionVerificationResult {
  reelPositionsVerified: number
  reelPositionsFailed: number
  failedReelSpins: number[] // Nonces of spins with failed reel positions
}

export async function verifySessionWithHKDF(
  serverSeed: string,
  serverSeedHash: string,
  spins: EnhancedSpinVerificationData[],
  defaultReelLengths?: number[]
): Promise<EnhancedSessionVerificationResult> {
  const result: EnhancedSessionVerificationResult = {
    isValid: true,
    serverSeedValid: false,
    spinsVerified: 0,
    spinsFailed: 0,
    failedSpins: [],
    reelPositionsVerified: 0,
    reelPositionsFailed: 0,
    failedReelSpins: [],
  }

  // Verify server seed commitment
  result.serverSeedValid = await verifyServerSeedCommitment(serverSeed, serverSeedHash)
  if (!result.serverSeedValid) {
    result.isValid = false
    return result
  }

  // Verify each spin
  for (const spin of spins) {
    // Verify spin hash
    const hashValid = await verifySpinHash(
      spin.prevSpinHash,
      serverSeed,
      spin.clientSeed,
      spin.nonce,
      spin.spinHash
    )

    if (hashValid) {
      result.spinsVerified++
    } else {
      result.spinsFailed++
      result.failedSpins.push(spin.nonce)
      result.isValid = false
    }

    // Verify reel positions using HKDF (if reel lengths provided)
    const reelLengths = spin.reelLengths || defaultReelLengths
    if (reelLengths && spin.reelPositions.length === reelLengths.length) {
      const reelValid = await verifyAllReelPositionsHKDF(
        serverSeed,
        spin.clientSeed,
        spin.nonce,
        spin.prevSpinHash, // Required for hash chain
        reelLengths,
        spin.reelPositions
      )

      if (reelValid) {
        result.reelPositionsVerified++
      } else {
        result.reelPositionsFailed++
        result.failedReelSpins.push(spin.nonce)
        result.isValid = false
      }
    }
  }

  return result
}
