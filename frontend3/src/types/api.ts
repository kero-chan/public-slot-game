// API Request Types
export interface RegisterRequest {
  username: string
  email: string
  password: string
  game_id?: string // Optional: null/undefined = cross-game account
}

export interface LoginRequest {
  username: string
  password: string
  game_id?: string // Optional: game being logged into
  force_logout?: boolean // Force logout existing session on another device
  device_info?: string // Device information (browser, OS, screen size, etc.)
}

export interface StartSessionRequest {
  bet_amount: number
  // Dual Commitment Protocol: SHA256(theta_seed) - client's commitment sent BEFORE seeing server_seed
  theta_commitment?: string
}

export interface ExecuteSpinRequest {
  session_id: string | null
  bet_amount: number
  client_seed?: string // Provably fair: client-generated seed for RNG
  // Dual Commitment Protocol: theta_seed is revealed on first spin
  theta_seed?: string
  // Game mode: 'bonus_spin_trigger' for guaranteed free spins
  game_mode?: string
}

export interface ExecuteFreeSpinRequest {
  free_spins_session_id: string
  client_seed?: string // Provably fair: client-generated seed for RNG
}

// API Response Types
export interface PlayerProfile {
  id: string
  username: string
  email: string
  balance: number
  game_id?: string // null = cross-game account
  total_spins: number
  total_wagered: number
  total_won: number
  is_active: boolean
  is_verified: boolean
  created_at: string
  last_login_at?: string
}

// Login response - contains session token
export interface AuthResponse {
  session_token: string
  expires_at: number // Unix timestamp when session expires
  player: PlayerProfile
}

// Register response - no session, must login after
export interface RegisterResponse {
  message: string
  player: PlayerProfile
}

// Provably Fair Types
export interface SpinVerificationData {
  nonce: number
  client_seed: string
  spin_hash: string
  prev_spin_hash: string
  reel_positions: number[]
}

export interface SessionProvablyFairData {
  session_id: string
  server_seed_hash: string
  nonce_start: number
  // Only present when session is ended (revealed)
  server_seed?: string
  total_spins?: number
  spins?: SpinVerificationData[]
}

export interface SessionResponse {
  id: string
  player_id: string
  bet_amount: number
  starting_balance: number
  ending_balance?: number
  total_spins: number
  total_wagered: number
  total_won: number
  net_change: number
  created_at: string
  ended_at?: string
  // Provably fair data (present when session started or ended)
  provably_fair?: SessionProvablyFairData
}

export interface Position {
  reel: number
  row: number
}

export interface WinInfo {
  symbol: number  // Symbol ID (same as grid values)
  count: number
  ways: number
  payout: number
  win_amount: number
  positions: Position[]  // Grid positions that form this win
  win_intensity: 'small' | 'medium' | 'big' | 'mega'
}

export interface CascadeInfo {
  cascade_number: number
  grid_after: number[][]  // Grid state after this cascade (after removal and drop) - numbers from backend
  multiplier: number
  wins: WinInfo[]
  total_cascade_win: number
}

export interface SpinResponse {
  spin_id: string
  session_id: string
  bet_amount: number
  balance_before: number
  balance_after_bet: number
  new_balance: number
  grid: number[][]  // Grid as numbers from backend (use convertGridToSymbols to convert)
  cascades: CascadeInfo[]
  spin_total_win: number
  scatter_count: number
  is_free_spin: boolean
  free_spins_triggered: boolean
  free_spins_retriggered: boolean
  free_spins_additional?: number
  free_spins_session_id?: string
  free_spins_remaining_spins: number
  free_session_total_win: number
  timestamp: string
}

export interface FreeSpinsStatusResponse {
  active: boolean
  free_spins_session_id?: string
  session_id?: string // Game session ID for provably fair recovery
  total_spins_awarded: number
  spins_completed: number
  remaining_spins: number
  locked_bet_amount: number
  total_won: number
}

export interface SpinSummary {
  spin_id: string
  session_id: string
  bet_amount: number
  total_win: number
  scatter_count: number
  is_free_spin: boolean
  free_spins_triggered: boolean
  created_at: string
}

export interface SpinHistoryResponse {
  page: number
  limit: number
  total: number
  spins: SpinSummary[]
}

export interface ErrorResponse {
  error: string
  message: string
  details?: any
}
