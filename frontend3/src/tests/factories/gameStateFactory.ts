import { ref, type Ref } from 'vue'

/**
 * Factory for creating mock game state objects
 * Used in tests to create consistent test data
 */

export interface MockGameState {
  bet: Ref<number>
  balance: Ref<number>
  state: Ref<string>
  inFreeSpinMode: Ref<boolean>
  freeSpinsRemaining: Ref<number>
  consecutiveWins: Ref<number>
  anticipationMode: Ref<boolean>
  totalWinAmount: Ref<number>
}

export function createMockGameState(overrides?: Partial<MockGameState>): MockGameState {
  return {
    bet: ref(10),
    balance: ref(1000),
    state: ref('idle'),
    inFreeSpinMode: ref(false),
    freeSpinsRemaining: ref(0),
    consecutiveWins: ref(0),
    anticipationMode: ref(false),
    totalWinAmount: ref(0),
    ...overrides
  }
}

export interface MockGridState {
  grid: string[][]
  reelStrips: string[][]
  reelTopIndex: number[]
  spinOffsets: number[]
  spinVelocities: number[]
  highlightWins: any[] | null
  disappearPositions: Set<string>
  lastRemovedPositions: Set<string>
  isDropAnimating: boolean
}

export function createMockGridState(overrides?: Partial<MockGridState>): MockGridState {
  return {
    grid: createDefaultGrid(),
    reelStrips: Array(5).fill(null).map(() => Array(100).fill('fa')),
    reelTopIndex: Array(5).fill(0),
    spinOffsets: Array(5).fill(0),
    spinVelocities: Array(5).fill(0),
    highlightWins: null,
    disappearPositions: new Set(),
    lastRemovedPositions: new Set(),
    isDropAnimating: false,
    ...overrides
  }
}

export function createDefaultGrid(symbol: string = 'fa'): string[][] {
  return Array(5).fill(null).map(() => Array(10).fill(symbol))
}

export function createGridWithSymbols(symbols: string[][]): string[][] {
  if (symbols.length !== 5) {
    throw new Error('Grid must have exactly 5 columns')
  }
  return symbols.map(col => {
    if (col.length !== 10) {
      throw new Error('Each column must have exactly 10 rows')
    }
    return [...col]
  })
}

export interface MockWinCombination {
  symbol: number  // Symbol ID (same as grid values)
  count: number
  positions: Array<{ reel: number; row: number }>
  payout?: number
  win_intensity: 'small' | 'medium' | 'big' | 'mega'
}

export function createMockWin(overrides?: Partial<MockWinCombination>): MockWinCombination {
  return {
    symbol: 2,  // fa = 2
    count: 3,
    positions: [
      { reel: 0, row: 5 },
      { reel: 1, row: 5 },
      { reel: 2, row: 5 },
    ],
    payout: 30,
    win_intensity: 'small',
    ...overrides
  }
}

export interface MockSpinResponse {
  spin_id: string
  session_id: string
  bet_amount: number
  balance_before: number
  balance_after_bet: number
  new_balance: number
  grid: string[][]
  cascades: MockCascadeData[]
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

export interface MockCascadeData {
  wins: MockWinCombination[]
  win_amount: number
  grid_after: string[][]
}

export function createMockSpinResponse(overrides?: Partial<MockSpinResponse>): MockSpinResponse {
  return {
    spin_id: 'test-spin-123',
    session_id: 'test-session-456',
    bet_amount: 10,
    balance_before: 1000,
    balance_after_bet: 990,
    new_balance: 990,
    grid: createDefaultGrid(),
    cascades: [],
    spin_total_win: 0,
    scatter_count: 0,
    is_free_spin: false,
    free_spins_triggered: false,
    free_spins_retriggered: false,
    free_spins_remaining_spins: 0,
    free_session_total_win: 0,
    timestamp: new Date().toISOString(),
    // Legacy fields
    ...overrides
  }
}

export function createMockCascade(overrides?: Partial<MockCascadeData>): MockCascadeData {
  return {
    wins: [createMockWin()],
    win_amount: 30,
    grid_after: createDefaultGrid(),
    ...overrides
  }
}
