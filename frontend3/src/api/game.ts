import apiClient from './client'
import { API_ENDPOINTS } from '@/config/constants'
import { convertGridToSymbols } from '@/utils/symbolConverter'
import { useAuthStore } from '@/stores/user/auth'
import type {
  StartSessionRequest,
  SessionResponse,
  ExecuteSpinRequest,
  SpinResponse,
  FreeSpinsStatusResponse,
  ExecuteFreeSpinRequest,
  SpinHistoryResponse,
} from '@/types/api'

// Helper to get endpoints based on trial mode
function getEndpoints() {
  const authStore = useAuthStore()
  const isTrial = authStore.isTrial

  return {
    SESSION_START: isTrial ? API_ENDPOINTS.TRIAL.SESSION_START : API_ENDPOINTS.SESSION.START,
    SPIN_EXECUTE: isTrial ? API_ENDPOINTS.TRIAL.SPIN : API_ENDPOINTS.SPIN.EXECUTE,
    FREE_SPINS_STATUS: isTrial ? API_ENDPOINTS.TRIAL.FREE_SPINS_STATUS : API_ENDPOINTS.FREE_SPINS.STATUS,
    FREE_SPINS_EXECUTE: isTrial ? API_ENDPOINTS.TRIAL.FREE_SPINS_EXECUTE : API_ENDPOINTS.FREE_SPINS.EXECUTE,
  }
}

// Converted response types with string grids (after conversion from backend numbers)
interface ConvertedCascadeInfo {
  cascade_number: number
  grid_after: string[][]
  multiplier: number
  wins: Array<{
    symbol: number  // Symbol ID (same as grid values)
    count: number
    ways: number
    payout: number
    win_amount: number
    positions: Array<{ reel: number; row: number }>
    win_intensity: 'small' | 'medium' | 'big' | 'mega'
  }>
  total_cascade_win: number
}

interface ConvertedSpinResponse {
  spin_id: string
  session_id: string
  bet_amount: number
  balance_before: number
  balance_after_bet: number
  new_balance: number
  grid: string[][]
  cascades: ConvertedCascadeInfo[]
  spin_total_win: number
  scatter_count: number
  is_free_spin: boolean
  free_spins_triggered: boolean
  free_spins_session_id?: string
  free_spins_remaining_spins: number
  free_session_total_win: number
  timestamp: string

}

export const gameApi = {
  // Initial Grid (for display, no auth required)
  async getInitialGrid(): Promise<{ grid: string[][] }> {
    const response = await apiClient.get<{ grid: number[][] }>(API_ENDPOINTS.GAME.INITIAL_GRID)
    // Convert number grid to symbol grid
    return {
      grid: convertGridToSymbols(response.data.grid)
    }
  },

  // Session Management
  async startSession(data: StartSessionRequest): Promise<SessionResponse> {
    const endpoints = getEndpoints()
    const response = await apiClient.post<SessionResponse>(endpoints.SESSION_START, data)
    return response.data
  },

  async endSession(sessionId: string): Promise<SessionResponse> {
    const response = await apiClient.post<SessionResponse>(API_ENDPOINTS.SESSION.END(sessionId))
    return response.data
  },

  async getSessionHistory(page = 1, limit = 20): Promise<{ page: number; limit: number; sessions: SessionResponse[] }> {
    const response = await apiClient.get(API_ENDPOINTS.SESSION.HISTORY, {
      params: { page, limit },
    })
    return response.data
  },

  // Spin Execution
  async executeSpin(data: ExecuteSpinRequest, clientSeed?: string): Promise<ConvertedSpinResponse> {
    const endpoints = getEndpoints()
    const requestData = {
      ...data,
      client_seed: clientSeed,
      // Dual Commitment Protocol: theta_seed is included if present in data
    }
    const response = await apiClient.post<SpinResponse>(endpoints.SPIN_EXECUTE, requestData)
    const backendData = response.data

    // Convert all grids from numbers to symbols
    return {
      ...backendData,
      grid: convertGridToSymbols(backendData.grid),
      cascades: backendData.cascades.map(cascade => ({
        ...cascade,
        grid_after: convertGridToSymbols(cascade.grid_after)
      }))
    }
  },

  async getSpinHistory(page = 1, limit = 20): Promise<SpinHistoryResponse> {
    const response = await apiClient.get<SpinHistoryResponse>(API_ENDPOINTS.SPIN.HISTORY, {
      params: { page, limit },
    })
    return response.data
  },

  // Free Spins
  async getFreeSpinsStatus(): Promise<FreeSpinsStatusResponse> {
    const endpoints = getEndpoints()
    const response = await apiClient.get<FreeSpinsStatusResponse>(endpoints.FREE_SPINS_STATUS)
    return response.data
  },

  async executeFreeSpin(data: ExecuteFreeSpinRequest, clientSeed?: string): Promise<ConvertedSpinResponse> {
    const endpoints = getEndpoints()
    const requestData = {
      ...data,
      client_seed: clientSeed,
    }
    const response = await apiClient.post<SpinResponse>(endpoints.FREE_SPINS_EXECUTE, requestData)
    const backendData = response.data

    // Convert all grids from numbers to symbols
    return {
      ...backendData,
      grid: convertGridToSymbols(backendData.grid),
      cascades: backendData.cascades.map(cascade => ({
        ...cascade,
        grid_after: convertGridToSymbols(cascade.grid_after)
      }))
    }
  },
}

// Export converted types for use in other modules
export type { ConvertedSpinResponse, ConvertedCascadeInfo }
