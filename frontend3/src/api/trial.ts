import apiClient from './client'
import { API_ENDPOINTS } from '@/config/constants'

// Trial Types
export interface StartTrialRequest {
  game_id?: string
}

export interface TrialProfile {
  id: string
  username: string
  balance: number
  total_spins: number
  total_wagered: number
  total_won: number
  is_trial: boolean
}

export interface TrialResponse {
  session_token: string
  expires_at: number
  player: TrialProfile
}

export interface TrialBalanceResponse {
  balance: number
}

export const trialApi = {
  // Start a new trial session (no auth required)
  async startTrial(gameId?: string): Promise<TrialResponse> {
    const response = await apiClient.post<TrialResponse>(
      API_ENDPOINTS.AUTH.TRIAL,
      gameId ? { game_id: gameId } : {}
    )
    return response.data
  },

  // Get trial profile (requires trial token)
  async getProfile(): Promise<TrialProfile> {
    const response = await apiClient.get<TrialProfile>(API_ENDPOINTS.TRIAL.PROFILE)
    return response.data
  },

  // Get trial balance (requires trial token)
  async getBalance(): Promise<TrialBalanceResponse> {
    const response = await apiClient.get<TrialBalanceResponse>(API_ENDPOINTS.TRIAL.BALANCE)
    return response.data
  },
}
