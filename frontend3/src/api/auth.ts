import apiClient from './client'
import { API_ENDPOINTS } from '@/config/constants'
import type {
  RegisterRequest,
  LoginRequest,
  AuthResponse,
  RegisterResponse,
  PlayerProfile,
} from '@/types/api'

export const authApi = {
  // Register a new player (returns player info, must login to get session)
  async register(data: RegisterRequest): Promise<RegisterResponse> {
    const response = await apiClient.post<RegisterResponse>(API_ENDPOINTS.AUTH.REGISTER, data)
    return response.data
  },

  // Login (returns session token)
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await apiClient.post<AuthResponse>(API_ENDPOINTS.AUTH.LOGIN, data)
    return response.data
  },

  // Logout (invalidates current session)
  async logout(): Promise<void> {
    await apiClient.post(API_ENDPOINTS.AUTH.LOGOUT)
  },

  // Get current player profile
  async getProfile(): Promise<PlayerProfile> {
    const response = await apiClient.get<PlayerProfile>(API_ENDPOINTS.PLAYER.PROFILE)
    return response.data
  },

  // Get current balance
  async getBalance(): Promise<{ balance: number }> {
    const response = await apiClient.get<{ balance: number }>(API_ENDPOINTS.PLAYER.BALANCE)
    return response.data
  },
}
