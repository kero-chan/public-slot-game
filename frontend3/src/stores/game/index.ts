/**
 * Game Stores
 * Core game logic stores for the slot machine
 */

// Main game store (facade for backward compatibility)
export { useGameStore, GAME_STATES, type GameState, type LoadingProgress, type GameStoreState } from './gameStore'

// Game flow and state machine
export { useGameFlowStore, type GameFlowState } from './gameFlowStore'

// Betting management
export { useBettingStore, type BettingState } from './bettingStore'

// Free spins mode
export { useFreeSpinsStore, type FreeSpinsState } from './freeSpinsStore'

// Spin wins tracking
export { useSpinWinsStore, type SpinWinsState } from './spinWinsStore'

// Grid and reel management
export { useGridStore, type GridStoreState, type AnimationState } from './gridStore'

// Winning tile animations
export { useWinningStore, WINNING_STATES, type WinningState, type WinningStoreState, type WinningTimings } from './winningStore'

// Timing constants
export { useTimingStore, type TimingState } from './timingStore'

// Backend game API integration
export { useBackendGameStore } from './backendGame'
