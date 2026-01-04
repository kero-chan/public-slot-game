import axios, { AxiosInstance, AxiosError } from 'axios'
import type { ErrorResponse } from '@/types/api'
import { API_CONFIG } from '@/config/constants'

// Create axios instance with centralized configuration
const apiClient: AxiosInstance = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
  headers: API_CONFIG.DEFAULT_HEADERS,
})

// Request interceptor - add session token and game ID
apiClient.interceptors.request.use(
  async (config) => {
    // Import auth store dynamically to avoid circular dependency
    // and to ensure we get the latest session token value
    const { useAuthStore } = await import('@/stores')
    const authStore = useAuthStore()

    // Add session token to Authorization header
    if (authStore.sessionToken) {
      config.headers.Authorization = `Bearer ${authStore.sessionToken}`
    }

    // Add game ID header for session validation
    if (authStore.gameId) {
      config.headers['X-Game-ID'] = authStore.gameId
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor - handle errors
apiClient.interceptors.response.use(
  (response) => {
    return response
  },
  async (error: AxiosError<ErrorResponse>) => {
    if (error.response) {
      // Server responded with error
      const errorData = error.response.data

      // Handle 401 Unauthorized
      // Note: Don't redirect here, let the auth store handle it
      // to avoid hard reloads and allow proper state cleanup
      if (error.response.status === 401) {
        // Clear auth via store (pinia will persist automatically)
        const { useAuthStore } = await import('@/stores')
        const authStore = useAuthStore()
        authStore.logout()
      }

      // Handle 409 Conflict (e.g., active_session_exists)
      // Log but don't treat as a hard error for certain cases
      if (error.response.status === 409) {
        console.warn('⚠️ Conflict error (409):', errorData?.error || errorData?.message)
      }

      // Return structured error
      return Promise.reject({
        status: error.response.status,
        error: errorData?.error || 'unknown_error',
        message: errorData?.message || 'An error occurred',
        details: errorData?.details,
      })
    } else if (error.request) {
      // Request made but no response
      return Promise.reject({
        error: 'network_error',
        message: 'Network error. Please check your connection.',
      })
    } else {
      // Something else happened
      return Promise.reject({
        error: 'request_error',
        message: error.message || 'Failed to make request',
      })
    }
  }
)

export default apiClient
