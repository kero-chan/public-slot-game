import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import { trialApi } from '@/api/trial'
import { STORAGE_KEYS } from '@/config/constants'
import { getGameId } from '@/api/gameAssets'
import { getDeviceInfoString } from '@/utils/deviceInfo'
import type { PlayerProfile, RegisterRequest, LoginRequest } from '@/types/api'

export const useAuthStore = defineStore('auth', () => {
  // State
  const sessionToken = ref<string | null>(null)
  const sessionExpiresAt = ref<number | null>(null)
  const player = ref<PlayerProfile | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const isTrial = ref(false) // Trial mode flag

  // Getters
  const isAuthenticated = computed(() => !!sessionToken.value && !!player.value && !isSessionExpired.value)
  const isSessionExpired = computed(() => {
    if (!sessionExpiresAt.value) return true
    return Date.now() > sessionExpiresAt.value * 1000 // expires_at is Unix timestamp in seconds
  })
  const username = computed(() => player.value?.username || '')
  const balance = computed(() => player.value?.balance || 0)
  const gameId = computed(() => getGameId())

  // Actions
  async function register(data: Omit<RegisterRequest, 'game_id'>) {
    isLoading.value = true
    error.value = null

    // game_id is REQUIRED for registration
    const currentGameId = getGameId()
    if (!currentGameId) {
      error.value = 'Game configuration error: VITE_GAME_ID is not set'
      isLoading.value = false
      throw new Error('VITE_GAME_ID is required for registration')
    }

    try {
      // Add game_id from environment (required)
      const requestData: RegisterRequest = {
        ...data,
        game_id: currentGameId,
      }
      const response = await authApi.register(requestData)

      // Registration successful - user must login to get session
      // Return response so UI can show success message
      return response
    } catch (err: any) {
      error.value = err.response?.data?.message || err.message || 'Registration failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function login(data: Omit<LoginRequest, 'game_id'>) {
    isLoading.value = true
    error.value = null

    // game_id is REQUIRED for login (each client is bound to a game)
    const currentGameId = getGameId()
    if (!currentGameId) {
      error.value = 'Game configuration error: VITE_GAME_ID is not set'
      isLoading.value = false
      throw new Error('VITE_GAME_ID is required for login')
    }

    try {
      // Add game_id from environment (required) and device info
      const requestData: LoginRequest = {
        ...data,
        game_id: currentGameId,
        device_info: getDeviceInfoString(),
      }
      const response = await authApi.login(requestData)

      // Store session token and player data (persisted automatically)
      sessionToken.value = response.session_token
      sessionExpiresAt.value = response.expires_at
      player.value = response.player
      isTrial.value = false // Ensure trial mode is off for regular login

      return response
    } catch (err: any) {
      // Check for already_logged_in error
      // Note: API client interceptor transforms error to { status, error, message, details }
      if (err.error === 'already_logged_in') {
        error.value = 'You are already logged in on another device. Enable force_logout to logout other device.'
      } else {
        error.value = err.message || 'Login failed'
      }
      throw err
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Start a trial session (no registration required)
   * Trial sessions have HUGE RTP for better winning experience
   */
  async function startTrialMode() {
    isLoading.value = true
    error.value = null

    const currentGameId = getGameId()

    try {
      const response = await trialApi.startTrial(currentGameId || undefined)

      // Store trial session token and player data
      sessionToken.value = response.session_token
      sessionExpiresAt.value = response.expires_at
      isTrial.value = true

      // Convert trial player to PlayerProfile format
      player.value = {
        id: response.player.id,
        username: response.player.username,
        email: '',
        balance: response.player.balance,
        total_spins: response.player.total_spins,
        total_wagered: response.player.total_wagered,
        total_won: response.player.total_won,
        is_active: true,
        is_verified: false,
        created_at: new Date().toISOString(),
      }

      return response
    } catch (err: any) {
      error.value = err.message || 'Failed to start trial mode'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchProfile() {
    if (!sessionToken.value) return

    isLoading.value = true
    error.value = null

    try {
      if (isTrial.value) {
        // For trial sessions, use trial API
        const trialProfile = await trialApi.getProfile()
        player.value = {
          id: trialProfile.id,
          username: trialProfile.username,
          email: '',
          balance: trialProfile.balance,
          total_spins: trialProfile.total_spins,
          total_wagered: trialProfile.total_wagered,
          total_won: trialProfile.total_won,
          is_active: true,
          is_verified: false,
          created_at: new Date().toISOString(),
        }
      } else {
        player.value = await authApi.getProfile()
      }
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch profile'

      // If unauthorized, logout
      if (err.status === 401) {
        logout()
      }

      throw err
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Initialize auth state on app mount
   * Validates session and fetches fresh profile
   * Trial sessions are NOT persisted - user must restart trial on each page load
   */
  async function init() {
    // Trial sessions should not persist across page reloads
    // Clear trial session and redirect to login
    if (isTrial.value) {
      console.log('Trial session detected on init - clearing (trials do not persist)')
      sessionToken.value = null
      sessionExpiresAt.value = null
      player.value = null
      isTrial.value = false
      return false
    }

    if (!sessionToken.value) {
      return false
    }

    // Check if session expired
    if (isSessionExpired.value) {
      logout()
      return false
    }

    try {
      // Session exists, fetch fresh profile to validate
      await fetchProfile()
      return true
    } catch (err) {
      // Session invalid, clear state
      logout()
      return false
    }
  }

  async function refreshBalance() {
    if (!sessionToken.value) return

    try {
      const response = await authApi.getBalance()
      if (player.value) {
        player.value.balance = response.balance
      }
    } catch (err: any) {
      console.error('Failed to refresh balance:', err)
    }
  }

  async function logout() {
    // Call server logout if we have a session (skip for trial - no server logout needed)
    if (sessionToken.value && !isTrial.value) {
      try {
        await authApi.logout()
      } catch (err) {
        // Ignore errors on logout - we're clearing local state anyway
        console.warn('Logout request failed:', err)
      }
    }

    // Stop and reset all audio (must happen before clearing state)
    const { audioManager } = await import('@/composables/audioManager')
    audioManager.reset()

    // Clear local state
    sessionToken.value = null
    sessionExpiresAt.value = null
    player.value = null
    isTrial.value = false

    const { useGameStore, useBackendGameStore } = await import('@/stores')
    useGameStore().resetGame()
    useBackendGameStore().reset()
  }

  function updateBalance(newBalance: number) {
    if (player.value) {
      player.value.balance = newBalance
    }
  }

  return {
    // State
    sessionToken,
    sessionExpiresAt,
    player,
    isLoading,
    error,
    isTrial,

    // Getters
    isAuthenticated,
    isSessionExpired,
    username,
    balance,
    gameId,

    // Actions
    init,
    register,
    login,
    startTrialMode,
    logout,
    fetchProfile,
    refreshBalance,
    updateBalance,
  }
}, {
  persist: {
    key: STORAGE_KEYS.AUTH,
    storage: localStorage,
  }
})
