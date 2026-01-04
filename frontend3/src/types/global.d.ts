import { Ref } from 'vue'
import type { Container, Sprite, Graphics, Text, Texture, Application } from 'pixi.js'

/**
 * Global Type Definitions
 * Shared types used across the application
 */

// ============================================
// Game State Types
// ============================================

export type GameStatus =
  | 'idle'
  | 'spinning'
  | 'checking_wins'
  | 'showing_wins'
  | 'cascading'
  | 'gold_transforming'
  | 'jackpot'
  | 'bonus_triggered'

export interface GameState {
  bet: Ref<number>
  balance: Ref<number>
  state: Ref<GameStatus>
  inFreeSpinMode: Ref<boolean>
  freeSpinsRemaining: Ref<number>
  consecutiveWins: Ref<number>
  anticipationMode: Ref<boolean>
  totalWinAmount: Ref<number>
  // Methods
  startSpinCycle: () => boolean
  completeSpinAnimation: () => void
  transitionTo: (newState: GameStatus) => void
  activateAnticipationMode: () => void
  deactivateAnticipationMode: () => void
  increaseBet: () => void
  decreaseBet: () => void
}

// ============================================
// Grid Types
// ============================================

export interface GridState {
  grid: string[][]  // 5 columns x 10 rows
  reelStrips: string[][]  // 5 columns x 100 rows (strip length)
  reelTopIndex: number[]  // Current top position for each reel
  spinOffsets: number[]  // Sub-pixel offsets for smooth animation
  spinVelocities: number[]  // Current velocity for each reel
  highlightWins: WinCombination[] | null
  highlightAnim: { start: number; duration: number }
  disappearPositions: Set<string>  // Positions marked for disappear animation
  disappearAnim: { start: number; duration: number }
  lastRemovedPositions: Set<string>
  goldTransformedPositions: Set<string>  // Positions where gold tiles transformed to wild (should not cascade)
  previousGridSnapshot: string[][]
  bufferRows: string[][]
  lastCascadeTime: number
  isDropAnimating: boolean
  activeSlowdownColumn: number
  hasGlowableTiles?: boolean  // Whether there are bonus/wild tiles visible
  glowableTileInfos?: Array<{ key: string; x: number; y: number; width: number; height: number }>  // Position info for sparkle effects
}

// ============================================
// Win Types
// ============================================

export interface WinCombination {
  symbol: number  // Symbol ID (same as grid values)
  count: number   // Number of matching reels (3, 4, or 5)
  positions: WinPosition[]
  payout?: number
  win_intensity: WinIntensity  // Visual intensity from backend
}

export interface WinPosition {
  reel: number  // 0-4 (column)
  row: number   // 0-9 (row in grid)
  is_gold_to_wild?: boolean  // True if this gold tile transforms to wild (from backend)
}

export type WinIntensity = 'small' | 'medium' | 'big' | 'mega'

// ============================================
// Backend Response Types
// ============================================

export interface SpinResponse {
  spin_id: string  // Unique identifier for this spin
  session_id: string  // Current game session ID
  bet_amount: number  // Bet amount for this spin
  balance_before: number  // Balance before placing bet
  balance_after_bet: number  // Balance after bet deduction
  new_balance: number  // Balance after winnings credited
  grid: string[][]  // Landing grid (5x10)
  cascades: CascadeData[]  // All cascades for this spin
  spin_total_win: number  // Total win amount for entire spin (including all cascades)
  scatter_count: number  // Number of scatter/bonus symbols
  is_free_spin: boolean  // Whether this was a free spin
  free_spins_triggered: boolean  // Whether free spins were triggered
  free_spins_retriggered: boolean  // Whether free spins were retriggered during free spin
  free_spins_additional?: number  // Additional free spins awarded on retrigger
  free_spins_session_id?: string  // Free spins session ID if triggered
  free_spins_remaining_spins: number  // Remaining free spins
  free_session_total_win: number  // Total accumulated win in free spins session
  timestamp: string  // ISO 8601 timestamp
}

export interface CascadeData {
  wins: WinCombination[]  // Wins detected in this cascade
  total_cascade_win: number  // Total win amount for this cascade
  grid_after: string[][]  // Grid state after cascade completes
  cascade_number?: number
  multiplier?: number  // Current multiplier for this cascade (from backend)
}
// ============================================
// Config Types
// ============================================

export interface GameConfig {
  reels: ReelsConfig
  game: GameRulesConfig
  paytable: Record<string, number[]>
  symbols: SymbolsConfig
  animations: AnimationConfig
}

export interface ReelsConfig {
  count: number  // 5
  rows: number  // 10 total rows
  bufferRows: number  // 4
  fullyVisibleRows: number  // 4
  winCheckStartRow: number  // 5
  winCheckEndRow: number  // 8
  stripLength: number  // 100
}

export interface GameRulesConfig {
  maxBonusPerColumn: number  // 2
  defaultBet: number
  minBet: number
  maxBet: number
  betIncrements: number[]
  freeSpinsCount: number
  freeSpinsMultiplier: number
}

export interface SymbolsConfig {
  symbols: string[]
  bonusSymbol: string
  wildSymbol: string
  goldenSuffix: string
}

export interface AnimationConfig {
  spinDuration: number
  cascadeDuration: number
  winHighlightDuration: number
  anticipationSlowdown: number
}

// ============================================
// PixiJS Renderer Types
// ============================================

export interface PixiRenderer {
  app: Application | null
  stage: Container
  resize: (width: number, height: number) => void
  render: () => void
  destroy: () => void
}

export interface HeaderElements {
  container: Container
  multiplierSprites: Sprite[]
  multiplierBackgrounds: Sprite[]
}

export interface FooterElements {
  container: Container
  spinButton: Sprite
  betDisplay: Text
  balanceDisplay: Text
  winAmountDisplay: Text
}

export interface ReelsElements {
  container: Container
  reelContainers: Container[]
  tileSprites: Sprite[][][]  // [col][row][sprite]
  backdrop: Graphics
}

export interface Rectangle {
  x: number
  y: number
  w: number
  h: number
}

// ============================================
// Audio Types
// ============================================

export interface AudioManager {
  playEffect: (effectName: string) => void
  playWinSound: (intensity: WinIntensity) => void
  playConsecutiveWinSound: (consecutiveWins: number) => void
  playBackgroundMusic: () => void
  stopBackgroundMusic: () => void
  setVolume: (volume: number) => void
  mute: () => void
  unmute: () => void
}

export interface AudioEffectName {
  SPIN: 'lot' | 'reel_spin'
  WIN: 'small_win' | 'medium_win' | 'big_win' | 'mega_win'
  BONUS: 'reach_bonus' | 'bonus_trigger'
  CASCADE: 'cascade_drop'
  GOLD: 'gold_transform'
  UI: 'button_click' | 'button_hover'
}

// ============================================
// Spin State Types (Refactored Feature)
// ============================================

export interface SpinState {
  backendTargetGrid: Ref<string[][] | null>
  backendCascades: Ref<CascadeData[]>
  currentCascadeIndex: Ref<number>
  currentCascade: Ref<CascadeData | null>
  setBackendSpinResult: (response: SpinResponse) => void
  hasMoreBackendCascades: () => boolean
  advanceToCascade: (index: number | null) => void
}

// ============================================
// Grid Verification Types
// ============================================

export interface GridVerification {
  allMatch: boolean
  mismatches: GridMismatch[]
  totalCells: number
}

export interface GridMismatch {
  position: [number, number]  // [col, row]
  backend: string
  displayed: string
}

// ============================================
// Utility Types
// ============================================

export interface TileData {
  symbol: string
  isGolden: boolean
  isWild: boolean
  isBonus: boolean
}

export interface RandomSymbolOptions {
  col: number
  visualRow?: number
  allowGold?: boolean
  allowBonus?: boolean
}

// ============================================
// API Client Types
// ============================================

export interface ApiClient {
  post: <T>(url: string, data: any) => Promise<T>
  get: <T>(url: string) => Promise<T>
  setAuthToken: (token: string) => void
}

// ============================================
// Store Types
// ============================================

export interface TimingConfig {
  SPIN_BASE_DURATION: number
  SPIN_REEL_STAGGER: number
  DROP_DURATION: number
  CASCADE_MAX_WAIT: number
  HIGHLIGHT_ANIMATION_DURATION: number
  DISAPPEAR_WAIT: number
  ANTICIPATION_SLOWDOWN_PER_COLUMN: number
}

export interface SettingsState {
  soundEnabled: Ref<boolean>
  musicEnabled: Ref<boolean>
  volume: Ref<number>
  toggleSound: () => void
  toggleMusic: () => void
  setVolume: (volume: number) => void
}
