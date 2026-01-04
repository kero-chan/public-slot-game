import { describe, it, expect } from 'vitest'
import { createMockGameState, createMockGridState, createMockWin } from './factories/gameStateFactory'

/**
 * Example test file to verify testing infrastructure is working
 * This can be deleted once real tests are written
 */

describe('Testing Infrastructure', () => {
  describe('Basic Vitest Functionality', () => {
    it('should run basic assertions', () => {
      expect(true).toBe(true)
      expect(1 + 1).toBe(2)
      expect('hello').toBe('hello')
    })

    it('should support async tests', async () => {
      const promise = Promise.resolve(42)
      const result = await promise
      expect(result).toBe(42)
    })
  })

  describe('Test Factories', () => {
    it('should create mock game state', () => {
      const gameState = createMockGameState()

      expect(gameState.bet.value).toBe(10)
      expect(gameState.balance.value).toBe(1000)
      expect(gameState.state.value).toBe('idle')
      expect(gameState.inFreeSpinMode.value).toBe(false)
    })

    it('should create mock game state with overrides', () => {
      const gameState = createMockGameState({
        bet: { value: 20 } as any,
        balance: { value: 500 } as any,
      })

      expect(gameState.bet.value).toBe(20)
      expect(gameState.balance.value).toBe(500)
    })

    it('should create mock grid state', () => {
      const gridState = createMockGridState()

      expect(gridState.grid).toHaveLength(5) // 5 columns
      expect(gridState.grid[0]).toHaveLength(10) // 10 rows
      expect(gridState.reelTopIndex).toHaveLength(5)
      expect(gridState.spinOffsets).toHaveLength(5)
    })

    it('should create mock win combination', () => {
      const win = createMockWin({
        symbol: 3,  // zhong
        count: 5,
        payout: 100,
        win_intensity: 'mega'
      })

      expect(win.symbol).toBe(3)
      expect(win.count).toBe(5)
      expect(win.payout).toBe(100)
      expect(win.positions).toBeDefined()
    })
  })

  describe('TypeScript Support', () => {
    it('should have proper type inference', () => {
      const gameState = createMockGameState()

      // TypeScript should infer these types correctly
      const bet: number = gameState.bet.value
      const balance: number = gameState.balance.value
      const state: string = gameState.state.value

      expect(typeof bet).toBe('number')
      expect(typeof balance).toBe('number')
      expect(typeof state).toBe('string')
    })
  })
})
