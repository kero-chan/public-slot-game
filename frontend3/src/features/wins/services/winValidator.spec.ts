import { describe, it, expect } from 'vitest'
import { getWinIntensity, validateBackendWins } from './winValidator'
import { createMockWin } from '@/tests/factories/gameStateFactory'
import type { WinCombination } from '../../spin/types'

describe('winValidator', () => {
  describe('getWinIntensity', () => {
    it('should return "small" for empty wins array', () => {
      const result = getWinIntensity([])
      expect(result).toBe('small')
    })

    it('should return "small" for null wins', () => {
      const result = getWinIntensity(null as any)
      expect(result).toBe('small')
    })

    it('should return "small" for undefined wins', () => {
      const result = getWinIntensity(undefined as any)
      expect(result).toBe('small')
    })

    it('should return "small" when win_intensity is "small"', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'small' })
      ]
      const result = getWinIntensity(wins)
      expect(result).toBe('small')
    })

    it('should return "medium" when win_intensity is "medium"', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'medium' })
      ]
      const result = getWinIntensity(wins)
      expect(result).toBe('medium')
    })

    it('should return "big" when win_intensity is "big"', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'big' })
      ]
      const result = getWinIntensity(wins)
      expect(result).toBe('big')
    })

    it('should return "mega" when win_intensity is "mega"', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'mega' })
      ]
      const result = getWinIntensity(wins)
      expect(result).toBe('mega')
    })

    it('should return highest intensity from multiple wins', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'small' }),
        createMockWin({ win_intensity: 'medium' }),
        createMockWin({ win_intensity: 'mega' })
      ]
      const result = getWinIntensity(wins)
      expect(result).toBe('mega')
    })

    it('should return "big" when highest win is big', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'small' }),
        createMockWin({ win_intensity: 'medium' }),
        createMockWin({ win_intensity: 'big' })
      ]
      const result = getWinIntensity(wins)
      expect(result).toBe('big')
    })

    it('should return "medium" when highest win is medium', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'small' }),
        createMockWin({ win_intensity: 'medium' })
      ]
      const result = getWinIntensity(wins)
      expect(result).toBe('medium')
    })

    it('should handle single win correctly', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'mega' })
      ]
      const result = getWinIntensity(wins)
      expect(result).toBe('mega')
    })

    it('should handle multiple mega wins', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'mega' }),
        createMockWin({ win_intensity: 'mega' }),
        createMockWin({ win_intensity: 'mega' })
      ]
      const result = getWinIntensity(wins)
      expect(result).toBe('mega')
    })

    it('should use correct intensity level ordering', () => {
      const winsSmall: WinCombination[] = [createMockWin({ win_intensity: 'small' })]
      const winsMedium: WinCombination[] = [createMockWin({ win_intensity: 'medium' })]
      const winsBig: WinCombination[] = [createMockWin({ win_intensity: 'big' })]

      expect(getWinIntensity(winsSmall)).toBe('small')
      expect(getWinIntensity(winsMedium)).toBe('medium')
      expect(getWinIntensity(winsBig)).toBe('big')
    })
  })

  describe('validateBackendWins', () => {
    it('should return true for valid wins array', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'mega' })
      ]
      const result = validateBackendWins(wins)
      expect(result).toBe(true)
    })

    it('should return true for empty wins array', () => {
      const wins: WinCombination[] = []
      const result = validateBackendWins(wins)
      expect(result).toBe(true)
    })

    it('should throw error for null wins', () => {
      expect(() => validateBackendWins(null)).toThrow('Backend wins data missing - cannot proceed')
    })

    it('should throw error for undefined wins', () => {
      expect(() => validateBackendWins(undefined)).toThrow('Backend wins data missing - cannot proceed')
    })

    it('should not throw for valid data structures', () => {
      const wins: WinCombination[] = [
        createMockWin({ win_intensity: 'mega' }),
        createMockWin({ win_intensity: 'big' })
      ]
      expect(() => validateBackendWins(wins)).not.toThrow()
    })

    it('should throw error with exact message for null', () => {
      try {
        validateBackendWins(null)
        // Should not reach here
        expect(true).toBe(false)
      } catch (error) {
        expect((error as Error).message).toBe('Backend wins data missing - cannot proceed')
      }
    })

    it('should throw error with exact message for undefined', () => {
      try {
        validateBackendWins(undefined)
        // Should not reach here
        expect(true).toBe(false)
      } catch (error) {
        expect((error as Error).message).toBe('Backend wins data missing - cannot proceed')
      }
    })
  })
})
