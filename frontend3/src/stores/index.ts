/**
 * Store Exports
 * Central export point for all Pinia stores
 */

// ==================== Game Stores ====================
// Core game logic stores
export {
  // Main game store (facade)
  useGameStore,
  GAME_STATES,
  type GameState,
  type GameStoreState,

  // Game flow
  useGameFlowStore,
  type GameFlowState,

  // Betting
  useBettingStore,
  type BettingState,

  // Free spins
  useFreeSpinsStore,
  type FreeSpinsState,

  // Spin wins
  useSpinWinsStore,
  type SpinWinsState,

  // Grid
  useGridStore,
  type GridStoreState,
  type AnimationState,

  // Winning animations
  useWinningStore,
  WINNING_STATES,
  type WinningState,
  type WinningStoreState,
  type WinningTimings,

  // Timing
  useTimingStore,
  type TimingState,

  // Backend integration
  useBackendGameStore,
} from './game'

// ==================== User Stores ====================
// User-related stores
export {
  useAuthStore,
  useSettingsStore,
  type SettingsState,
} from './user'

// ==================== UI Stores ====================
// UI state stores
export {
  useUIStore,
  type UIState,
  type LoadingProgress,
} from './ui'
