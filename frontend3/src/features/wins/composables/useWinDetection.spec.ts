import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  getBackendWins,
  findWinningCombinations,
  findWinningCombinations_DEPRECATED_FORBIDDEN,
} from './useWinDetection'
import { useSpinState } from '../../spin/composables/useSpinState'
import { createMockSpinResponse, createMockCascade, createMockWin } from '@/tests/factories/gameStateFactory'

describe('useWinDetection', () => {
  describe('getBackendWins', () => {
    it('should throw error when backendCascades is null', () => {
      const spinState = useSpinState()
      spinState.backendCascades.value = null as any

      expect(() => getBackendWins(spinState)).toThrow(
        'Backend cascade data missing - cannot proceed with spin'
      )
    })

    it('should throw error when backendCascades is undefined', () => {
      const spinState = useSpinState()
      spinState.backendCascades.value = undefined as any

      expect(() => getBackendWins(spinState)).toThrow(
        'Backend cascade data missing - cannot proceed with spin'
      )
    })

    it('should log security error when backend data is missing', () => {
      const spinState = useSpinState()
      spinState.backendCascades.value = null as any
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      try {
        getBackendWins(spinState)
      } catch (error) {
        // Expected
      }

      expect(consoleErrorSpy).toHaveBeenCalled()
      expect(consoleErrorSpy.mock.calls.some(call =>
        call[0].includes('CRITICAL SECURITY ERROR')
      )).toBe(true)

      consoleErrorSpy.mockRestore()
    })

    it('should return empty array for empty cascades', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({ cascades: [] })
      spinState.setBackendSpinResult(response)

      const wins = getBackendWins(spinState)

      expect(wins).toEqual([])
    })

    it('should call advanceToCascade(null) for empty cascades', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({ cascades: [] })
      spinState.setBackendSpinResult(response)
      const advanceSpy = vi.spyOn(spinState, 'advanceToCascade')

      getBackendWins(spinState)

      expect(advanceSpy).toHaveBeenCalledWith(null)
      advanceSpy.mockRestore()
    })

    it('should return wins from first cascade', () => {
      const spinState = useSpinState()
      const mockWins = [
        createMockWin({ symbol: 3, count: 5, win_intensity: 'mega' }),  // zhong
        createMockWin({ symbol: 2, count: 4, win_intensity: 'big' })    // fa
      ]
      const response = createMockSpinResponse({
        cascades: [createMockCascade({ wins: mockWins })]
      })
      spinState.setBackendSpinResult(response)

      const wins = getBackendWins(spinState)

      expect(wins).toHaveLength(2)
      expect(wins[0].symbol).toBe(3)
      expect(wins[0].count).toBe(5)
      expect(wins[1].symbol).toBe(2)
      expect(wins[1].count).toBe(4)
    })

    it('should advance cascade index after retrieving wins', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade({ wins: [createMockWin()] })]
      })
      spinState.setBackendSpinResult(response)

      expect(spinState.currentCascadeIndex.value).toBe(0)
      getBackendWins(spinState)
      expect(spinState.currentCascadeIndex.value).toBe(1)
    })

    it('should call advanceToCascade with current index', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade()]
      })
      spinState.setBackendSpinResult(response)
      const advanceSpy = vi.spyOn(spinState, 'advanceToCascade')

      getBackendWins(spinState)

      expect(advanceSpy).toHaveBeenCalledWith(0)
      advanceSpy.mockRestore()
    })

    it('should return empty array when no wins in cascade', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade({ wins: [] })]
      })
      spinState.setBackendSpinResult(response)

      const wins = getBackendWins(spinState)

      expect(wins).toEqual([])
    })

    it('should handle multiple cascades sequentially', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [
          createMockCascade({ wins: [createMockWin({ symbol: 3, count: 5, win_intensity: 'mega' })] }),  // zhong
          createMockCascade({ wins: [createMockWin({ symbol: 2, count: 4, win_intensity: 'big' })] })    // fa
        ]
      })
      spinState.setBackendSpinResult(response)

      // First call - gets first cascade
      const wins1 = getBackendWins(spinState)
      expect(wins1[0].symbol).toBe(3)
      expect(spinState.currentCascadeIndex.value).toBe(1)

      // Second call - gets second cascade
      const wins2 = getBackendWins(spinState)
      expect(wins2[0].symbol).toBe(2)
      expect(spinState.currentCascadeIndex.value).toBe(2)
    })

    it('should return empty array after all cascades processed', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade()]
      })
      spinState.setBackendSpinResult(response)

      // Process the only cascade
      getBackendWins(spinState)

      // Try to get more cascades
      const wins = getBackendWins(spinState)

      expect(wins).toEqual([])
    })

    it('should call advanceToCascade(null) after all cascades processed', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade()]
      })
      spinState.setBackendSpinResult(response)

      getBackendWins(spinState) // Process first cascade

      const advanceSpy = vi.spyOn(spinState, 'advanceToCascade')
      getBackendWins(spinState) // Try to get more

      expect(advanceSpy).toHaveBeenCalledWith(null)
      advanceSpy.mockRestore()
    })

    it('should log cascade processing information', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade()]
      })
      spinState.setBackendSpinResult(response)
      const consoleLogSpy = vi.spyOn(console, 'log').mockImplementation(() => {})

      getBackendWins(spinState)

      expect(consoleLogSpy).toHaveBeenCalled()
      const logMessage = consoleLogSpy.mock.calls.find(call =>
        call[0].includes('Using backend cascade')
      )
      expect(logMessage).toBeDefined()

      consoleLogSpy.mockRestore()
    })

    it('should log when all cascades are processed', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade()]
      })
      spinState.setBackendSpinResult(response)
      const consoleLogSpy = vi.spyOn(console, 'log').mockImplementation(() => {})

      getBackendWins(spinState) // Process first
      getBackendWins(spinState) // Try to get more

      const logMessage = consoleLogSpy.mock.calls.find(call =>
        call[0].includes('All') && call[0].includes('cascades processed')
      )
      expect(logMessage).toBeDefined()

      consoleLogSpy.mockRestore()
    })

    it('should handle cascade with null wins gracefully', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [{ ...createMockCascade(), wins: null as any }]
      })
      spinState.setBackendSpinResult(response)

      const wins = getBackendWins(spinState)

      expect(wins).toEqual([])
    })

    it('should handle cascade with undefined wins gracefully', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [{ ...createMockCascade(), wins: undefined as any }]
      })
      spinState.setBackendSpinResult(response)

      const wins = getBackendWins(spinState)

      expect(wins).toEqual([])
    })
  })

  describe('findWinningCombinations_DEPRECATED_FORBIDDEN', () => {
    it('should always throw error', () => {
      expect(() => findWinningCombinations_DEPRECATED_FORBIDDEN()).toThrow(
        'findWinningCombinations is forbidden - backend calculates all wins'
      )
    })

    it('should log security violation', () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      try {
        findWinningCombinations_DEPRECATED_FORBIDDEN()
      } catch (error) {
        // Expected
      }

      expect(consoleErrorSpy).toHaveBeenCalled()
      expect(consoleErrorSpy.mock.calls.some(call =>
        call[0].includes('SECURITY VIOLATION')
      )).toBe(true)

      consoleErrorSpy.mockRestore()
    })
  })

  describe('findWinningCombinations', () => {
    it('should call getBackendWins', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade({ wins: [createMockWin()] })]
      })
      spinState.setBackendSpinResult(response)

      const wins = findWinningCombinations(spinState)

      expect(wins).toBeDefined()
      expect(Array.isArray(wins)).toBe(true)
    })

    it('should return same result as getBackendWins', () => {
      const spinState1 = useSpinState()
      const spinState2 = useSpinState()
      const mockWins = [createMockWin({ symbol: 3, count: 5, win_intensity: 'mega' })]  // zhong
      const response = createMockSpinResponse({
        cascades: [createMockCascade({ wins: mockWins })]
      })

      spinState1.setBackendSpinResult(response)
      spinState2.setBackendSpinResult(response)

      const wins1 = getBackendWins(spinState1)
      const wins2 = findWinningCombinations(spinState2)

      expect(wins1).toHaveLength(wins2.length)
      expect(wins1[0].symbol).toBe(wins2[0].symbol)
      expect(wins1[0].count).toBe(wins2[0].count)
    })

    it('should handle empty cascades', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({ cascades: [] })
      spinState.setBackendSpinResult(response)

      const wins = findWinningCombinations(spinState)

      expect(wins).toEqual([])
    })

    it('should throw error for missing backend data', () => {
      const spinState = useSpinState()
      spinState.backendCascades.value = null as any

      expect(() => findWinningCombinations(spinState)).toThrow(
        'Backend cascade data missing - cannot proceed with spin'
      )
    })
  })
})
