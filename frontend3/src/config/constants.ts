/**
 * Application-wide constants and configuration
 */

// API Configuration
export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_URL || '/v1',
  TIMEOUT: 10000,
  DEFAULT_HEADERS: {
    'Content-Type': 'application/json',
  },
}

// API Endpoints
export const API_ENDPOINTS = {
  AUTH: {
    REGISTER: '/auth/register',
    LOGIN: '/auth/login',
    LOGOUT: '/auth/logout',
    TRIAL: '/auth/trial',
  },
  TRIAL: {
    PROFILE: '/trial/profile',
    BALANCE: '/trial/balance',
    PLAYER_BALANCE: '/trial/player/balance',
    SESSION_START: '/trial/session/start',
    SPIN: '/trial/spin',
    FREE_SPINS_STATUS: '/trial/free-spins/status',
    FREE_SPINS_EXECUTE: '/trial/free-spins/spin',
  },
  PLAYER: {
    PROFILE: '/player/profile',
    BALANCE: '/player/balance',
  },
  GAME: {
    INITIAL_GRID: '/initial-grid',
  },
  SESSION: {
    START: '/session/start',
    END: (sessionId: string) => `/session/${sessionId}/end`,
    HISTORY: '/session/history',
  },
  SPIN: {
    EXECUTE: '/base-spins/spin',
    HISTORY: '/base-spins/histories',
  },
  FREE_SPINS: {
    STATUS: '/free-spins/status',
    EXECUTE: '/free-spins/spin',
  },
}

// Game Configuration
export const GAME_CONFIG = {
  BET_OPTIONS: [0.25, 0.50, 1.0, 2.0, 5.0, 10.0, 25.0, 50.0, 100.0],
  DEFAULT_BET: 1.0,
}

// Animation Configuration
export const ANIMATION_CONFIG = {
  SPIN_NORMAL: {
    DURATION: 0.5,       // Normal speed
    STAGGER: 0.15,       // Normal stagger
    BOUNCE_DURATION: 0.2, // Normal bounce
    BOUNCE_OFFSET: 20,   // Normal offset
  },
  SPIN_FAST: {
    DURATION: 0.1,       // Fast speed
    STAGGER: 0.03,       // Fast stagger
    BOUNCE_DURATION: 0.05, // Fast bounce
    BOUNCE_OFFSET: 10,   // Minimal offset
  },
  CASCADE_NORMAL: {
    HIGHLIGHT_DURATION: 0.3,     // Normal speed
    REMOVE_DURATION: 0.3,        // Normal speed
    DROP_DURATION: 0.4,          // Normal speed
    FILL_DURATION: 0.4,          // Normal speed
  },
  CASCADE_FAST: {
    HIGHLIGHT_DURATION: 0.1,     // Fast speed
    REMOVE_DURATION: 0.1,        // Fast speed
    DROP_DURATION: 0.15,         // Fast speed
    FILL_DURATION: 0.15,         // Fast speed
  },
  FREE_SPINS_TRIGGER: {
    SCALE_IN: 0.5,
    PULSE: 0.3,
    PULSE_REPEAT: 3,
    FADE_OUT: 0.5,
  },
}

// Helper functions to get animation config based on fast spin setting
export function getSpinAnimationConfig(isFastSpin: boolean = false, isInFreeSpinMode: boolean = false) {
  // Only use fast spin for normal spins (not during free spin mode)
  const shouldUseFastSpin = isFastSpin && !isInFreeSpinMode
  return shouldUseFastSpin ? ANIMATION_CONFIG.SPIN_FAST : ANIMATION_CONFIG.SPIN_NORMAL
}

export function getCascadeAnimationConfig(isFastSpin: boolean = false, isInFreeSpinMode: boolean = false) {
  // Only use fast spin for normal spins (not during free spin mode)
  const shouldUseFastSpin = isFastSpin && !isInFreeSpinMode
  return shouldUseFastSpin ? ANIMATION_CONFIG.CASCADE_FAST : ANIMATION_CONFIG.CASCADE_NORMAL
}

// Grid Configuration
export const GRID_CONFIG = {
  COLS: 5,
  ROWS: 6,
  SYMBOL_WIDTH: 120,
  SYMBOL_HEIGHT: 120,
  PADDING: 10,
  START_X: 100,
  START_Y: 50,
}

// Symbol Colors
export const SYMBOL_COLORS: Record<string, number> = {
  fa: 0xff6b6b,
  zhong: 0x4ecdc4,
  guo: 0xffd93d,
  liangtong: 0x95e1d3,
  fu: 0xf38181,
  lu: 0xaa96da,
  shou: 0xfcbad3,
  xi: 0x6c5ce7,
  wild: 0xffd700,
  scatter: 0xff69b4,
}

// Symbol Display Names
export const SYMBOL_NAMES: Record<string, string> = {
  fa: '發',
  zhong: '中',
  guo: '國',
  liangtong: '兩通',
  fu: '福',
  lu: '祿',
  shou: '壽',
  xi: '喜',
  wild: '💎',
  scatter: '⭐',
}

// Storage Keys
export const STORAGE_KEYS = {
  AUTH: 'auth',
}

// Main CONFIG object (combines all configuration)
export const CONFIG = {
  canvas: {
    baseWidth: 600,
    baseHeight: 800,
    aspectRatio: 8 / 16, // 0.5 - Width to height ratio
    tileAspectRatio: 600 / 480, // 1.25 - Tile height/width ratio (actual tile image dimensions)
  },
  reels: {
    count: GRID_CONFIG.COLS,
    rows: 10, // Total rows from backend (4 buffer + 6 visible)
    bufferRows: 4, // Buffer rows (rows 0-3) above visible area
    fullyVisibleRows: 4, // Number of fully visible rows to check for wins
    stripLength: 100, // Longer strip = smoother spin animation, prevents gaps
    symbolSize: 70,
    spacing: 8,

    // WINNING CHECK ROWS - Single source of truth for all winning/bonus/effect calculations
    // Backend now sends 10 rows: rows 0-3 are buffer, rows 4-9 are visible area
    winCheckStartRow: 5,  // Start checking wins from grid row 5 (visual row 1 = first fully visible row)
    winCheckEndRow: 8,    // End checking wins at grid row 8 (visual row 4 = last fully visible row)

    // For renderer: visual row equivalents (calculated from grid rows)
    get visualWinStartRow(): number { return this.winCheckStartRow - this.bufferRows }, // 5 - 4 = 1
    get visualWinEndRow(): number { return this.winCheckEndRow - this.bufferRows }       // 8 - 4 = 4
  },
  animation: {
    spinDuration: 1600,
    reelStagger: 150,
  },
  game: {
    initialCredits: 100000,
    minBet: 10,
    maxBet: 100,
    betStep: 2,
    bettingMultiplierRate: 0.1,
    bonusScattersPerSpin: 2,       // Max bonus tiles per spin (for spawn control)
    maxBonusPerColumn: 1           // Maximum bonus tiles allowed per column in visible rows
  },
  spawnRates: {
    bonusChance: 0.03,             // 3% chance for bonus tiles
    wildChance: 0.02,              // 2% chance for wild tiles
    goldChance: 0.15               // 15% chance for gold variants
  }
}

// Paytable Configuration
export const PAYTABLE_CONFIG = {
  // High-value symbols
  fa: { 3: 10, 4: 25, 5: 50 },
  zhong: { 3: 8, 4: 20, 5: 40 },
  bai: { 3: 6, 4: 15, 5: 30 },
  bawan: { 3: 5, 4: 10, 5: 15 },

  // Low-value symbols
  wusuo: { 3: 3, 4: 5, 5: 12 },
  wutong: { 3: 3, 4: 5, 5: 12 },
  liangsuo: { 3: 2, 4: 4, 5: 10 },
  liangtong: { 3: 1, 4: 3, 5: 6 },
} as const

// Free Spins Configuration
export const FREE_SPINS_CONFIG = {
  awards: {
    3: 12,  // 3 scatters = 12 free spins
    4: 14,  // 4 scatters = 14 free spins (12 + 2)
    5: 16,  // 5 scatters = 16 free spins (12 + 4)
  },
  minScatters: 3,
  maxScatters: 5,
} as const

// Helper functions
export const GAME_RULES = {
  minSymbolsForPayout: 3,
  maxSymbolsForPayout: 5,
  minScattersForFreeSpin: 3,
} as const

// Type definitions for better type safety
export type PaytableSymbol = keyof typeof PAYTABLE_CONFIG
export type SymbolCount = 3 | 4 | 5
export type PayoutMultiplier = number

export interface SymbolPayout {
  3?: PayoutMultiplier
  4?: PayoutMultiplier
  5?: PayoutMultiplier
}

// Utility functions
export function getPayout(symbol: PaytableSymbol, count: SymbolCount): number {
  return PAYTABLE_CONFIG[symbol]?.[count] || 0
}

export function getFreeSpinsAward(scatterCount: number): number {
  if (scatterCount < 3) return 0
  if (scatterCount in FREE_SPINS_CONFIG.awards) {
    return FREE_SPINS_CONFIG.awards[scatterCount as keyof typeof FREE_SPINS_CONFIG.awards]
  }
  // For counts > 5, use formula: 12 + (2 × (scatter_count - 3))
  return 12 + (2 * (scatterCount - 3))
}
