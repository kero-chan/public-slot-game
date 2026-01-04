import { describe, it, expect, vi } from 'vitest'
import { transformGoldTilesToWild } from './goldTransformService'
import { createDefaultGrid, createMockWin } from '@/tests/factories/gameStateFactory'
import type { Grid, WinCombination } from '../../spin/types'

// Mock the tileHelpers module
vi.mock('@/utils/tileHelpers', () => ({
  istileWilden: (tile: string) => tile.includes('_gold'),
  isTileWildcard: (tile: string) => tile === 'wild'
}))

describe('goldTransformService', () => {
  describe('transformGoldTilesToWild', () => {
    it('should return empty result for no wins', () => {
      const grid = createDefaultGrid('fa')
      const wins: WinCombination[] = []

      const result = transformGoldTilesToWild(grid, wins)

      expect(result.transformCount).toBe(0)
      expect(result.transformedPositions.size).toBe(0)
    })

    it('should not transform non-gold tiles', () => {
      const grid: Grid = [
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const wins = [
        createMockWin({
          positions: [
            { reel: 0, row: 0 },
            { reel: 1, row: 0 }
          ]
        })
      ]

      const result = transformGoldTilesToWild(grid, wins)

      expect(result.transformCount).toBe(0)
      expect(grid[0][0]).toBe('fa')  // Not transformed
      expect(grid[1][0]).toBe('fa')  // Not transformed
    })

    it('should transform gold tiles to wild', () => {
      const grid: Grid = [
        ['fa_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['chu_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const wins = [
        createMockWin({
          positions: [
            { reel: 0, row: 0 },
            { reel: 1, row: 0 }
          ]
        })
      ]

      const result = transformGoldTilesToWild(grid, wins)

      expect(result.transformCount).toBe(2)
      expect(grid[0][0]).toBe('wild')
      expect(grid[1][0]).toBe('wild')
    })

    it('should not transform tiles that are already wild', () => {
      const grid: Grid = [
        ['wild', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const wins = [
        createMockWin({
          positions: [
            { reel: 0, row: 0 }
          ]
        })
      ]

      const result = transformGoldTilesToWild(grid, wins)

      expect(result.transformCount).toBe(0)
      expect(grid[0][0]).toBe('wild')  // Stays wild, not counted as transform
    })

    it('should add transformed positions to set', () => {
      const grid: Grid = [
        ['fa_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['chu_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const wins = [
        createMockWin({
          positions: [
            { reel: 0, row: 0 },
            { reel: 1, row: 0 }
          ]
        })
      ]

      const result = transformGoldTilesToWild(grid, wins)

      expect(result.transformedPositions.has('0,0')).toBe(true)
      expect(result.transformedPositions.has('1,0')).toBe(true)
      expect(result.transformedPositions.size).toBe(2)
    })

    it('should remove transformed positions from win.positions', () => {
      const grid: Grid = [
        ['fa_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const wins = [
        createMockWin({
          positions: [
            { reel: 0, row: 0 },  // Gold - will be removed
            { reel: 1, row: 0 }   // Not gold - will stay
          ]
        })
      ]

      transformGoldTilesToWild(grid, wins)

      expect(wins[0].positions).toHaveLength(1)
      expect(wins[0].positions[0]).toEqual({ reel: 1, row: 0 })
    })

    it('should handle multiple wins', () => {
      const grid: Grid = [
        ['fa_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['chu_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['zhong_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const wins = [
        createMockWin({
          symbol: 2,  // fa
          positions: [{ reel: 0, row: 0 }]
        }),
        createMockWin({
          symbol: 3,  // zhong
          positions: [{ reel: 1, row: 0 }]
        }),
        createMockWin({
          symbol: 3,  // zhong
          positions: [{ reel: 2, row: 0 }]
        })
      ]

      const result = transformGoldTilesToWild(grid, wins)

      expect(result.transformCount).toBe(3)
      expect(grid[0][0]).toBe('wild')
      expect(grid[1][0]).toBe('wild')
      expect(grid[2][0]).toBe('wild')
    })

    it('should handle mixed gold and non-gold positions in single win', () => {
      const grid: Grid = [
        ['fa_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const wins = [
        createMockWin({
          positions: [
            { reel: 0, row: 0 },  // Gold
            { reel: 1, row: 0 },  // Not gold
            { reel: 2, row: 0 }   // Gold
          ]
        })
      ]

      const result = transformGoldTilesToWild(grid, wins)

      expect(result.transformCount).toBe(2)
      expect(wins[0].positions).toHaveLength(1)
      expect(wins[0].positions[0]).toEqual({ reel: 1, row: 0 })
    })

    it('should handle multiple positions in different rows', () => {
      const grid: Grid = [
        ['fa_gold', 'chu', 'fa_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const wins = [
        createMockWin({
          positions: [
            { reel: 0, row: 0 },
            { reel: 0, row: 2 }
          ]
        })
      ]

      const result = transformGoldTilesToWild(grid, wins)

      expect(result.transformCount).toBe(2)
      expect(result.transformedPositions.has('0,0')).toBe(true)
      expect(result.transformedPositions.has('0,2')).toBe(true)
    })

    it('should not duplicate positions in set', () => {
      const grid: Grid = [
        ['fa_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const wins = [
        createMockWin({
          positions: [{ reel: 0, row: 0 }]
        }),
        createMockWin({
          positions: [{ reel: 0, row: 0 }]  // Duplicate position
        })
      ]

      const result = transformGoldTilesToWild(grid, wins)

      // Should only transform once, but count might be 2 if it processes both
      expect(result.transformedPositions.size).toBe(1)
      expect(grid[0][0]).toBe('wild')
    })

    it('should handle wins with no positions', () => {
      const grid = createDefaultGrid('fa')
      const wins = [
        createMockWin({ positions: [] })
      ]

      const result = transformGoldTilesToWild(grid, wins)

      expect(result.transformCount).toBe(0)
      expect(result.transformedPositions.size).toBe(0)
    })

    it('should modify grid in place', () => {
      const grid: Grid = [
        ['fa_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const originalGrid = grid
      const wins = [
        createMockWin({
          positions: [{ reel: 0, row: 0 }]
        })
      ]

      transformGoldTilesToWild(grid, wins)

      expect(grid).toBe(originalGrid)  // Same reference
      expect(grid[0][0]).toBe('wild')  // Modified in place
    })

    it('should use correct position format', () => {
      const grid: Grid = [
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa_gold', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu'],
        ['fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu', 'fa', 'chu']
      ]
      const wins = [
        createMockWin({
          positions: [{ reel: 3, row: 2 }]  // reel=col, row=row
        })
      ]

      const result = transformGoldTilesToWild(grid, wins)

      expect(result.transformedPositions.has('3,2')).toBe(true)
      expect(grid[3][2]).toBe('wild')
    })
  })
})
