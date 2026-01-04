import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from '../user/auth'
import { gameApi } from '@/api/game'
import {
  generateClientSeed,
  generateThetaSeed,
  computeThetaCommitment,
  type ProvablyFairState,
  createInitialPFState
} from '@/utils/provablyFair'
import type { SessionProvablyFairData } from '@/types/api'

/**
 * Backend Game Integration Store
 * Singleton store to manage backend game state and API calls
 * Ensures only one session and prevents duplicate API calls
 */
export const useBackendGameStore = defineStore('backendGame', () => {
  const authStore = useAuthStore()
  const currentSessionId = ref<string | null>(null)
  const currentFreeSpinsSessionId = ref<string | null>(null)
  const isProcessing = ref(false)

  // Provably fair state
  const provablyFairState = ref<ProvablyFairState>(createInitialPFState())
  const sessionVerificationData = ref<SessionProvablyFairData | null>(null)

  // Get balance from auth store
  const balance = computed(() => authStore.balance)

  /**
   * Start a new game session with Dual Commitment Protocol
   *
   * Dual Commitment Protocol flow:
   * 1. Client generates theta_seed, computes theta_commitment = SHA256(theta_seed)
   * 2. Client sends theta_commitment to server (BEFORE seeing server_seed)
   * 3. Server generates server_seed ONLY AFTER receiving theta_commitment
   * 4. Server responds with server_seed_hash (commitment)
   * 5. On first spin, client reveals theta_seed
   * 6. Server verifies: SHA256(theta_seed) === theta_commitment
   * 7. Game uses both seeds for RNG - neither party could bias the result
   */
  async function startSession(betAmount: number) {
    try {
      // Dual Commitment Protocol: Generate theta_seed and compute commitment
      const thetaSeed = generateThetaSeed()
      const thetaCommitment = await computeThetaCommitment(thetaSeed)

      console.log('🔐 Dual Commitment: Generated theta_seed, sending theta_commitment to server')

      const response = await gameApi.startSession({
        bet_amount: betAmount,
        theta_commitment: thetaCommitment, // Send commitment BEFORE seeing server_seed
      })

      currentSessionId.value = response.id

      // Initialize provably fair state from session response
      if (response.provably_fair) {
        provablyFairState.value = {
          serverSeedHash: response.provably_fair.server_seed_hash,
          nonceStart: response.provably_fair.nonce_start,
          currentNonce: response.provably_fair.nonce_start,
          isActive: true,
          // Dual Commitment Protocol fields
          thetaSeed: thetaSeed, // Keep secret until first spin
          thetaCommitment: thetaCommitment,
          thetaRevealed: false,
        }
        console.log('🎲 Provably Fair session started with Dual Commitment Protocol')
        console.log('   Server seed hash:', response.provably_fair.server_seed_hash)
        console.log('   Theta commitment:', thetaCommitment.slice(0, 16) + '...')
      }

      return response
    } catch (error: any) {
      // If active session exists, end it first and create a new one
      if (error.error === 'active_session_exists') {
        try {
          // Get session history to find the active session
          const history = await gameApi.getSessionHistory(1, 1)
          if (history.sessions && history.sessions.length > 0) {
            const activeSession = history.sessions[0]
            // End the active session if it's not already ended
            if (!activeSession.ended_at) {
              await gameApi.endSession(activeSession.id)
            }
          }

          // Generate new theta_seed for retry
          const thetaSeed = generateThetaSeed()
          const thetaCommitment = await computeThetaCommitment(thetaSeed)

          // Now try to start a new session with Dual Commitment
          const response = await gameApi.startSession({
            bet_amount: betAmount,
            theta_commitment: thetaCommitment,
          })
          currentSessionId.value = response.id

          // Initialize provably fair state from session response
          if (response.provably_fair) {
            provablyFairState.value = {
              serverSeedHash: response.provably_fair.server_seed_hash,
              nonceStart: response.provably_fair.nonce_start,
              currentNonce: response.provably_fair.nonce_start,
              isActive: true,
              // Dual Commitment Protocol fields
              thetaSeed: thetaSeed,
              thetaCommitment: thetaCommitment,
              thetaRevealed: false,
            }
            console.log('🎲 Provably Fair session started with Dual Commitment Protocol')
          }

          return response
        } catch (retryError) {
          throw retryError
        }
      }
      throw error
    }
  }

  /**
   * Execute a spin
   * Generates a new client seed for provably fair RNG
   *
   * Dual Commitment Protocol: On first spin (nonce=1), reveals theta_seed to server
   * @param betAmount - The bet amount for this spin
   * @param gameMode - Optional game mode (e.g., 'bonus_spin_trigger' for guaranteed free spins)
   */
  async function executeSpin(betAmount: number, gameMode?: string) {
    // ATOMIC: Check and set in one operation to prevent race condition
    if (isProcessing.value) {
      return null
    }

    // Immediately set to true BEFORE any await to prevent race condition
    isProcessing.value = true

    try {
      // Start session if not exist (may return null if active session exists)
      if (!currentSessionId.value) {
        await startSession(betAmount)
        // If null returned (active session exists), we continue without session tracking
      }

      // Generate client seed for this spin (provably fair)
      const clientSeed = generateClientSeed()

      // Dual Commitment Protocol: Reveal theta_seed on first spin
      let thetaSeed: string | undefined = undefined
      if (provablyFairState.value.thetaSeed && !provablyFairState.value.thetaRevealed) {
        thetaSeed = provablyFairState.value.thetaSeed
        console.log('🔐 Dual Commitment: Revealing theta_seed on first spin')
      }

      // Execute spin with client seed, theta_seed (if first spin), and game mode
      const response = await gameApi.executeSpin({
        session_id: currentSessionId.value,
        bet_amount: betAmount,
        theta_seed: thetaSeed,
        game_mode: gameMode,
      }, clientSeed)

      // Mark theta as revealed after successful first spin
      if (thetaSeed) {
        provablyFairState.value.thetaRevealed = true
        console.log('🔐 Dual Commitment: theta_seed verified by server')
      }

      // Update balance from response
      authStore.updateBalance(response.balance_after_bet)

      // Check if free spins triggered
      if (response.free_spins_triggered && response.free_spins_session_id) {
        currentFreeSpinsSessionId.value = response.free_spins_session_id
      }

      return response
    } catch (error) {
      throw error
    } finally {
      isProcessing.value = false
    }
  }

  /**
   * Execute a free spin
   * Generates a new client seed for provably fair RNG
   */
  async function executeFreeSpin() {
    // ATOMIC: Check and set in one operation to prevent race condition
    if (isProcessing.value) {
      return null
    }
    if (!currentFreeSpinsSessionId.value) {
      return null
    }

    // Immediately set to true BEFORE any await to prevent race condition
    isProcessing.value = true

    try {
      // Generate client seed for this spin (provably fair)
      const clientSeed = generateClientSeed()

      const response = await gameApi.executeFreeSpin({
        free_spins_session_id: currentFreeSpinsSessionId.value,
      }, clientSeed)

      // Update balance from response
      authStore.updateBalance(response.balance_after_bet)

      // Note: Retrigger detection is handled by gameStore.setSpinResponse()
      // which checks response.free_spins_retriggered and triggers the overlay

      // Check if free spins ended
      if (response.free_spins_remaining_spins === 0) {
        currentFreeSpinsSessionId.value = null
      }

      return response
    } catch (error) {
      throw error
    } finally {
      isProcessing.value = false
    }
  }

  /**
   * Get free spins status
   */
  async function getFreeSpinsStatus() {
    try {
      const response = await gameApi.getFreeSpinsStatus()

      if (response.active && response.free_spins_session_id) {
        currentFreeSpinsSessionId.value = response.free_spins_session_id
      } else {
        currentFreeSpinsSessionId.value = null
      }

      return response
    } catch {
      return null
    }
  }

  /**
   * Refresh balance from server
   */
  async function refreshBalance() {
    try {
      await authStore.refreshBalance()
    } catch {
      // Silent fail
    }
  }

  function setCurrentSessionId(sessionId: string | null) {
    currentSessionId.value = sessionId
  }

  /**
   * End current session and get verification data
   * Returns provably fair data for client-side verification
   */
  async function endSession() {
    if (!currentSessionId.value) {
      return null
    }

    try {
      const response = await gameApi.endSession(currentSessionId.value)

      // Store verification data if present (contains revealed server seed)
      if (response.provably_fair) {
        sessionVerificationData.value = response.provably_fair
        console.log('🎲 Session ended, server seed revealed:', response.provably_fair.server_seed)
        console.log('🎲 Total spins for verification:', response.provably_fair.total_spins)
      }

      // Reset provably fair state
      provablyFairState.value = createInitialPFState()
      currentSessionId.value = null

      return response
    } catch (error) {
      throw error
    }
  }

  /**
   * Get current provably fair state
   */
  function getProvablyFairState() {
    return provablyFairState.value
  }

  /**
   * Get session verification data (after session ends)
   */
  function getSessionVerificationData() {
    return sessionVerificationData.value
  }

  /**
   * Reset session state
   */
  function reset() {
    currentSessionId.value = null
    currentFreeSpinsSessionId.value = null
    isProcessing.value = false
    provablyFairState.value = createInitialPFState()
    sessionVerificationData.value = null
  }


  return {
    // State
    balance,
    currentSessionId,
    currentFreeSpinsSessionId,
    isProcessing,
    provablyFairState,
    sessionVerificationData,

    // Actions
    startSession,
    endSession,
    executeSpin,
    executeFreeSpin,
    getFreeSpinsStatus,
    refreshBalance,
    setCurrentSessionId,
    getProvablyFairState,
    getSessionVerificationData,
    reset,
  }
})
