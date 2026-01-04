import { describe, it, expect, vi } from 'vitest'
import { checkBonusTiles } from './bonusDetector'
import { createDefaultGrid } from '@/tests/factories/gameStateFactory'
import type { Grid } from '../../spin/types'

// Mock the tileHelpers module
vi.mock('@/utils/tileHelpers', () => ({
  isBonusTile: (tile: string) => tile === 'bonus'
}))

// Mock CONFIG
vi.mock('@/config/constants', () => ({
  CONFIG: {
    reels: {
      count: 5
    }
  }
}))

describe('bonusDetector', () => {
  describe('checkBonusTiles', () => {
    it('should return 0 when no bonus tiles are present', () => {
      const grid = createDefaultGrid('fa')

      const count = checkBonusTiles(grid, 2, 7)

      expect(count).toBe(0)
    })

    it('should count bonus tiles in specified row range', () => {
      const grid: Grid = [
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa']
      ]

      const count = checkBonusTiles(grid, 2, 7)

      expect(count).toBe(3)
    })

    it('should not count bonus tiles outside specified row range', () => {
      const grid: Grid = [
        ['bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'bonus', 'fa'],  // Row 0 and 8
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],  // Row 2 (in range)
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa']
      ]

      const count = checkBonusTiles(grid, 2, 7)

      expect(count).toBe(1)  // Only the one at row 2
    })

    it('should count multiple bonus tiles in same column', () => {
      const grid: Grid = [
        ['fa', 'fa', 'bonus', 'bonus', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa']
      ]

      const count = checkBonusTiles(grid, 2, 7)

      expect(count).toBe(3)
    })

    it('should count bonus tiles across all columns', () => {
      const grid: Grid = [
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa']
      ]

      const count = checkBonusTiles(grid, 2, 7)

      expect(count).toBe(5)
    })

    it('should handle start row equal to end row', () => {
      const grid: Grid = [
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa']
      ]

      const count = checkBonusTiles(grid, 2, 2)  // Only check row 2

      expect(count).toBe(2)
    })

    it('should handle first row (row 0)', () => {
      const grid: Grid = [
        ['bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa']
      ]

      const count = checkBonusTiles(grid, 0, 5)

      expect(count).toBe(2)
    })

    it('should handle last row (row 9)', () => {
      const grid: Grid = [
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'bonus'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'bonus'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa']
      ]

      const count = checkBonusTiles(grid, 5, 9)

      expect(count).toBe(2)
    })

    it('should count large number of bonus tiles', () => {
      const grid: Grid = [
        Array(10).fill('bonus'),
        Array(10).fill('bonus'),
        Array(10).fill('bonus'),
        Array(10).fill('bonus'),
        Array(10).fill('bonus')
      ]

      const count = checkBonusTiles(grid, 2, 7)

      expect(count).toBe(30)  // 5 columns * 6 rows
    })

    it('should handle mixed bonus and regular tiles', () => {
      const grid: Grid = [
        ['fa', 'chu', 'bonus', 'zhong', 'fa', 'bonus', 'fa', 'chu', 'fa', 'fa'],  // rows 2,5 in range
        ['bonus', 'fa', 'chu', 'fa', 'bonus', 'fa', 'zhong', 'fa', 'fa', 'fa'],  // row 4 in range
        ['fa', 'bonus', 'fa', 'chu', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],  // row 1 NOT in range
        ['chu', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa']
      ]

      const count = checkBonusTiles(grid, 2, 7)

      expect(count).toBe(3)  // Col0: rows 2,5 + Col1: row 4 = 3 total
    })
  })

  describe('integration - checkBonusTiles for visual display', () => {
    it('should correctly count 3 bonus tiles for visual highlighting', () => {
      // NOTE: checkBonusTiles is for VISUAL display only
      // Backend determines actual free spins trigger via response.free_spins_triggered
      const grid: Grid = [
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa']
      ]

      const count = checkBonusTiles(grid, 2, 7)

      expect(count).toBe(3)
      // Frontend should NOT make trigger decision - backend provides response.free_spins_triggered
    })

    it('should correctly count 2 bonus tiles for visual display', () => {
      const grid: Grid = [
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'bonus', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa'],
        ['fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa', 'fa']
      ]

      const count = checkBonusTiles(grid, 2, 7)

      expect(count).toBe(2)
      // Frontend should NOT make trigger decision - backend provides response.free_spins_triggered
    })
  })
})
