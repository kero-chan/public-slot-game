import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { validateGridIntegrity, verifyGridMatch, logVerification } from './spinService'
import { createDefaultGrid } from '@/tests/factories/gameStateFactory'
import type { Grid } from '../types'

describe('spinService', () => {
  describe('validateGridIntegrity', () => {
    it('should return the same grid if all cells are valid', () => {
      const grid = createDefaultGrid('fa')
      const result = validateGridIntegrity(grid, 10)

      expect(result).toHaveLength(5)
      expect(result[0]).toHaveLength(10)
      expect(result[0][0]).toBe('fa')
    })

    it('should use default fallback symbol "fa" when not specified', () => {
      const grid: Grid = [
        ['fa', ''],  // Invalid cell at [0][1]
        ['chu', 'zhong'],
        ['fa', 'chu'],
        ['zhong', 'fa'],
        ['chu', 'fa']
      ]

      const result = validateGridIntegrity(grid, 2)

      expect(result[0][1]).toBe('fa')
    })

    it('should replace invalid cells with specified fallback symbol', () => {
      const grid: Grid = [
        ['fa', ''],  // Invalid empty string
        ['chu', 'zhong'],
        ['fa', 'chu'],
        ['zhong', 'fa'],
        ['chu', 'fa']
      ]

      const result = validateGridIntegrity(grid, 2, 'chu')

      expect(result[0][1]).toBe('chu')
    })

    it('should replace null cells with fallback', () => {
      const grid: Grid = [
        ['fa', null as any],  // Invalid null
        ['chu', 'zhong'],
        ['fa', 'chu'],
        ['zhong', 'fa'],
        ['chu', 'fa']
      ]

      const result = validateGridIntegrity(grid, 2, 'fa')

      expect(result[0][1]).toBe('fa')
    })

    it('should replace undefined cells with fallback', () => {
      const grid: Grid = [
        ['fa', undefined as any],  // Invalid undefined
        ['chu', 'zhong'],
        ['fa', 'chu'],
        ['zhong', 'fa'],
        ['chu', 'fa']
      ]

      const result = validateGridIntegrity(grid, 2, 'fa')

      expect(result[0][1]).toBe('fa')
    })

    it('should replace non-string cells with fallback', () => {
      const grid: Grid = [
        ['fa', 123 as any],  // Invalid number
        ['chu', 'zhong'],
        ['fa', 'chu'],
        ['zhong', 'fa'],
        ['chu', 'fa']
      ]

      const result = validateGridIntegrity(grid, 2, 'fa')

      expect(result[0][1]).toBe('fa')
    })

    it('should fill empty columns with fallback symbols', () => {
      const grid: Grid = [
        [],  // Empty column
        ['chu', 'zhong'],
        ['fa', 'chu'],
        ['zhong', 'fa'],
        ['chu', 'fa']
      ]

      const result = validateGridIntegrity(grid, 3, 'fa')

      expect(result[0]).toHaveLength(3)
      expect(result[0][0]).toBe('fa')
      expect(result[0][1]).toBe('fa')
      expect(result[0][2]).toBe('fa')
    })

    it('should handle missing columns by creating them', () => {
      const grid: Grid = [
        null as any,  // Missing column
        ['chu', 'zhong'],
        ['fa', 'chu'],
        ['zhong', 'fa'],
        ['chu', 'fa']
      ]

      const result = validateGridIntegrity(grid, 2, 'fa')

      expect(result[0]).toHaveLength(2)
      expect(result[0][0]).toBe('fa')
      expect(result[0][1]).toBe('fa')
    })

    it('should create a copy of the grid, not modify the original', () => {
      const grid = createDefaultGrid('fa')
      const originalFirstCell = grid[0][0]

      const result = validateGridIntegrity(grid, 10)

      // Original should be unchanged
      expect(grid[0][0]).toBe(originalFirstCell)
      // Result should have the same value but be a different array
      expect(result[0]).not.toBe(grid[0])
    })

    it('should validate all 5 columns', () => {
      const grid = createDefaultGrid('fa')
      const result = validateGridIntegrity(grid, 10)

      expect(result).toHaveLength(5)
      for (let col = 0; col < 5; col++) {
        expect(result[col]).toHaveLength(10)
      }
    })

    it('should log warning for invalid cells', () => {
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})

      const grid: Grid = [
        ['fa', ''],
        ['chu', 'zhong'],
        ['fa', 'chu'],
        ['zhong', 'fa'],
        ['chu', 'fa']
      ]

      validateGridIntegrity(grid, 2)

      expect(consoleWarnSpy).toHaveBeenCalled()
      consoleWarnSpy.mockRestore()
    })
  })

  describe('verifyGridMatch', () => {
    it('should return allMatch=true for identical grids', () => {
      const grid1 = createDefaultGrid('fa')
      const grid2 = createDefaultGrid('fa')

      const result = verifyGridMatch(grid1, grid2)

      expect(result.allMatch).toBe(true)
      expect(result.mismatches).toHaveLength(0)
      expect(result.totalCells).toBe(50) // 5 columns x 10 rows
    })

    it('should detect single mismatch', () => {
      const grid1 = createDefaultGrid('fa')
      const grid2 = createDefaultGrid('fa')
      grid2[2][3] = 'chu'  // Create mismatch at [2][3]

      const result = verifyGridMatch(grid1, grid2)

      expect(result.allMatch).toBe(false)
      expect(result.mismatches).toHaveLength(1)
      expect(result.mismatches[0]).toEqual({
        col: 2,
        row: 3,
        backend: 'fa',
        displayed: 'chu'
      })
    })

    it('should detect multiple mismatches', () => {
      const grid1 = createDefaultGrid('fa')
      const grid2 = createDefaultGrid('fa')
      grid2[0][0] = 'chu'
      grid2[1][1] = 'zhong'
      grid2[4][9] = 'wild'

      const result = verifyGridMatch(grid1, grid2)

      expect(result.allMatch).toBe(false)
      expect(result.mismatches).toHaveLength(3)
    })

    it('should calculate correct total cells', () => {
      const grid1: Grid = Array(5).fill(null).map(() => ['fa', 'chu'])  // 5x2
      const grid2: Grid = Array(5).fill(null).map(() => ['fa', 'chu'])

      const result = verifyGridMatch(grid1, grid2)

      expect(result.totalCells).toBe(10)  // 5 columns x 2 rows
    })

    it('should include all mismatch details', () => {
      const grid1: Grid = [
        ['fa', 'chu'],
        ['fa', 'chu'],
        ['fa', 'chu'],
        ['fa', 'chu'],
        ['fa', 'chu']
      ]
      const grid2: Grid = [
        ['fa', 'chu'],
        ['fa', 'chu'],
        ['fa', 'zhong'],  // Mismatch at [2][1]
        ['fa', 'chu'],
        ['fa', 'chu']
      ]

      const result = verifyGridMatch(grid1, grid2)

      expect(result.mismatches[0]).toMatchObject({
        col: 2,
        row: 1,
        backend: 'chu',
        displayed: 'zhong'
      })
    })

    it('should handle empty grids', () => {
      const grid1: Grid = Array(5).fill([])
      const grid2: Grid = Array(5).fill([])

      const result = verifyGridMatch(grid1, grid2)

      expect(result.allMatch).toBe(true)
      expect(result.mismatches).toHaveLength(0)
      expect(result.totalCells).toBe(0)
    })

    it('should verify all columns and rows', () => {
      const grid1 = createDefaultGrid('fa')
      const grid2 = createDefaultGrid('chu')  // All different

      const result = verifyGridMatch(grid1, grid2)

      expect(result.allMatch).toBe(false)
      expect(result.mismatches).toHaveLength(50)  // Every cell mismatches
    })
  })

  describe('logVerification', () => {
    let consoleLogSpy: ReturnType<typeof vi.spyOn>
    let consoleErrorSpy: ReturnType<typeof vi.spyOn>

    beforeEach(() => {
      consoleLogSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    })

    afterEach(() => {
      consoleLogSpy.mockRestore()
      consoleErrorSpy.mockRestore()
    })

    it('should log success message when allMatch is true', () => {
      const verification = {
        allMatch: true,
        mismatches: [],
        totalCells: 50
      }

      logVerification(verification, 'SPIN COMPLETE')

      expect(consoleLogSpy).toHaveBeenCalled()
      const logCalls = consoleLogSpy.mock.calls.flat().join(' ')
      expect(logCalls).toContain('PASSED')
      expect(logCalls).toContain('50 tiles match')
    })

    it('should log error messages when allMatch is false', () => {
      const verification = {
        allMatch: false,
        mismatches: [
          { col: 1, row: 2, backend: 'fa', displayed: 'chu' }
        ],
        totalCells: 50
      }

      logVerification(verification, 'SPIN COMPLETE')

      expect(consoleErrorSpy).toHaveBeenCalled()
      const errorCalls = consoleErrorSpy.mock.calls.flat().join(' ')
      expect(errorCalls).toContain('FAILED')
      expect(errorCalls).toContain('MISMATCH')
    })

    it('should use default context when not provided', () => {
      const verification = {
        allMatch: true,
        mismatches: [],
        totalCells: 50
      }

      logVerification(verification)

      const logCalls = consoleLogSpy.mock.calls.flat().join(' ')
      expect(logCalls).toContain('VERIFICATION')
    })

    it('should use custom context', () => {
      const verification = {
        allMatch: true,
        mismatches: [],
        totalCells: 50
      }

      logVerification(verification, 'CASCADE 3')

      const logCalls = consoleLogSpy.mock.calls.flat().join(' ')
      expect(logCalls).toContain('CASCADE 3')
    })

    it('should log all mismatches', () => {
      const verification = {
        allMatch: false,
        mismatches: [
          { col: 0, row: 0, backend: 'fa', displayed: 'chu' },
          { col: 1, row: 1, backend: 'fa', displayed: 'zhong' },
          { col: 2, row: 2, backend: 'fa', displayed: 'wild' }
        ],
        totalCells: 50
      }

      logVerification(verification, 'TEST')

      expect(consoleErrorSpy).toHaveBeenCalledTimes(4)  // 1 header + 3 mismatches
    })

    it('should format mismatch details correctly', () => {
      const verification = {
        allMatch: false,
        mismatches: [
          { col: 3, row: 7, backend: 'fa', displayed: 'chu' }
        ],
        totalCells: 50
      }

      logVerification(verification)

      const errorCalls = consoleErrorSpy.mock.calls.flat().join(' ')
      expect(errorCalls).toContain('[3,7]')
      expect(errorCalls).toContain('Backend="fa"')
      expect(errorCalls).toContain('Display="chu"')
    })
  })
})
