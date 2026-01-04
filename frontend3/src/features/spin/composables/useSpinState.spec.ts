import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useSpinState } from './useSpinState'
import { createMockSpinResponse, createMockCascade } from '@/tests/factories/gameStateFactory'
import type { SpinResponse, CascadeData } from '../types'

describe('useSpinState', () => {
  describe('initialization', () => {
    it('should initialize with null/empty values', () => {
      const spinState = useSpinState()

      expect(spinState.backendTargetGrid.value).toBeNull()
      expect(spinState.backendCascades.value).toEqual([])
      expect(spinState.currentCascadeIndex.value).toBe(0)
      expect(spinState.currentCascade.value).toBeNull()
    })

    it('should expose all required methods', () => {
      const spinState = useSpinState()

      expect(typeof spinState.setBackendSpinResult).toBe('function')
      expect(typeof spinState.clearBackendSpinData).toBe('function')
      expect(typeof spinState.hasMoreBackendCascades).toBe('function')
      expect(typeof spinState.advanceToCascade).toBe('function')
    })
  })

  describe('setBackendSpinResult', () => {
    it('should set backend target grid', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        grid: [['fa', 'chu'], ['fa', 'chu'], ['fa', 'chu'], ['fa', 'chu'], ['fa', 'chu']]
      })

      spinState.setBackendSpinResult(response)

      expect(spinState.backendTargetGrid.value).toEqual(response.grid)
    })

    it('should set backend cascades', () => {
      const spinState = useSpinState()
      const cascades: CascadeData[] = [
        createMockCascade({ win_amount: 50 }),
        createMockCascade({ win_amount: 100 })
      ]
      const response = createMockSpinResponse({ cascades })

      spinState.setBackendSpinResult(response)

      expect(spinState.backendCascades.value).toHaveLength(2)
      expect(spinState.backendCascades.value[0].win_amount).toBe(50)
      expect(spinState.backendCascades.value[1].win_amount).toBe(100)
    })

    it('should reset currentCascadeIndex to 0', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse()

      // Set index to something other than 0
      spinState.currentCascadeIndex.value = 5

      spinState.setBackendSpinResult(response)

      expect(spinState.currentCascadeIndex.value).toBe(0)
    })

    it('should handle response with no cascades', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({ cascades: [] })

      spinState.setBackendSpinResult(response)

      expect(spinState.backendCascades.value).toEqual([])
    })

    it('should handle response with undefined cascades', () => {
      const spinState = useSpinState()
      const response = {
        ...createMockSpinResponse(),
        cascades: undefined as any
      }

      spinState.setBackendSpinResult(response)

      expect(spinState.backendCascades.value).toEqual([])
    })

    it('should log spin result information', () => {
      const spinState = useSpinState()
      const consoleLogSpy = vi.spyOn(console, 'log').mockImplementation(() => {})

      const response = createMockSpinResponse({
        cascades: [createMockCascade()],
        spin_total_win: 150
      })

      spinState.setBackendSpinResult(response)

      expect(consoleLogSpy).toHaveBeenCalled()
      const logMessage = consoleLogSpy.mock.calls[0][0]
      expect(logMessage).toContain('cascades=1')
      expect(logMessage).toContain('spin_total_win=150')

      consoleLogSpy.mockRestore()
    })
  })

  describe('clearBackendSpinData', () => {
    it('should reset all state to initial values', () => {
      const spinState = useSpinState()

      // Set some data first
      const response = createMockSpinResponse({
        cascades: [createMockCascade(), createMockCascade()]
      })
      spinState.setBackendSpinResult(response)
      spinState.advanceToCascade(0)

      // Clear data
      spinState.clearBackendSpinData()

      expect(spinState.backendTargetGrid.value).toBeNull()
      expect(spinState.backendCascades.value).toEqual([])
      expect(spinState.currentCascadeIndex.value).toBe(0)
      expect(spinState.currentCascade.value).toBeNull()
    })

    it('should be safe to call multiple times', () => {
      const spinState = useSpinState()

      spinState.clearBackendSpinData()
      spinState.clearBackendSpinData()
      spinState.clearBackendSpinData()

      expect(spinState.backendTargetGrid.value).toBeNull()
      expect(spinState.backendCascades.value).toEqual([])
    })

    it('should be safe to call on fresh instance', () => {
      const spinState = useSpinState()

      expect(() => spinState.clearBackendSpinData()).not.toThrow()
    })
  })

  describe('hasMoreBackendCascades', () => {
    it('should return false initially', () => {
      const spinState = useSpinState()

      expect(spinState.hasMoreBackendCascades()).toBe(false)
    })

    it('should return true when there are unprocessed cascades', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade(), createMockCascade()]
      })

      spinState.setBackendSpinResult(response)

      expect(spinState.hasMoreBackendCascades()).toBe(true)
    })

    it('should return false when all cascades are processed', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade()]
      })

      spinState.setBackendSpinResult(response)
      spinState.currentCascadeIndex.value = 1  // Processed all

      expect(spinState.hasMoreBackendCascades()).toBe(false)
    })

    it('should return false when currentCascadeIndex equals cascade count', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade(), createMockCascade(), createMockCascade()]
      })

      spinState.setBackendSpinResult(response)
      spinState.currentCascadeIndex.value = 3  // Same as length

      expect(spinState.hasMoreBackendCascades()).toBe(false)
    })

    it('should return false when currentCascadeIndex exceeds cascade count', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({
        cascades: [createMockCascade()]
      })

      spinState.setBackendSpinResult(response)
      spinState.currentCascadeIndex.value = 10  // Way past end

      expect(spinState.hasMoreBackendCascades()).toBe(false)
    })

    it('should return false when cascades array is empty', () => {
      const spinState = useSpinState()
      const response = createMockSpinResponse({ cascades: [] })

      spinState.setBackendSpinResult(response)

      expect(spinState.hasMoreBackendCascades()).toBe(false)
    })
  })

  describe('advanceToCascade', () => {
    let spinState: ReturnType<typeof useSpinState>
    let mockCascades: CascadeData[]

    beforeEach(() => {
      spinState = useSpinState()
      mockCascades = [
        createMockCascade({ win_amount: 50 }),
        createMockCascade({ win_amount: 100 }),
        createMockCascade({ win_amount: 150 })
      ]
      const response = createMockSpinResponse({ cascades: mockCascades })
      spinState.setBackendSpinResult(response)
    })

    it('should set currentCascade to the cascade at given index', () => {
      spinState.advanceToCascade(0)

      expect(spinState.currentCascade.value).toEqual(mockCascades[0])
      expect(spinState.currentCascade.value?.win_amount).toBe(50)
    })

    it('should update currentCascadeIndex', () => {
      spinState.advanceToCascade(1)

      expect(spinState.currentCascadeIndex.value).toBe(1)
      expect(spinState.currentCascade.value?.win_amount).toBe(100)
    })

    it('should handle advancing to last cascade', () => {
      spinState.advanceToCascade(2)

      expect(spinState.currentCascadeIndex.value).toBe(2)
      expect(spinState.currentCascade.value?.win_amount).toBe(150)
    })

    it('should handle advancing to first cascade', () => {
      spinState.advanceToCascade(1)  // Go to middle
      spinState.advanceToCascade(0)  // Go back to first

      expect(spinState.currentCascadeIndex.value).toBe(0)
      expect(spinState.currentCascade.value?.win_amount).toBe(50)
    })

    it('should clear currentCascade when index is null', () => {
      spinState.advanceToCascade(0)  // Set a cascade first
      expect(spinState.currentCascade.value).not.toBeNull()

      spinState.advanceToCascade(null)

      expect(spinState.currentCascade.value).toBeNull()
    })

    it('should handle invalid index gracefully', () => {
      spinState.advanceToCascade(999)  // Out of bounds

      expect(spinState.currentCascadeIndex.value).toBe(999)
      // currentCascade should not be set for invalid index
      expect(spinState.currentCascade.value).toBeNull()
    })

    it('should not throw when advancing on empty cascades', () => {
      const emptySpinState = useSpinState()
      const response = createMockSpinResponse({ cascades: [] })
      emptySpinState.setBackendSpinResult(response)

      expect(() => emptySpinState.advanceToCascade(0)).not.toThrow()
    })
  })

  describe('reactive state', () => {
    it('should have reactive backendTargetGrid', () => {
      const spinState = useSpinState()
      const grid1 = [['fa', 'chu'], ['fa', 'chu'], ['fa', 'chu'], ['fa', 'chu'], ['fa', 'chu']]
      const grid2 = [['chu', 'fa'], ['chu', 'fa'], ['chu', 'fa'], ['chu', 'fa'], ['chu', 'fa']]

      spinState.backendTargetGrid.value = grid1
      expect(spinState.backendTargetGrid.value).toEqual(grid1)

      spinState.backendTargetGrid.value = grid2
      expect(spinState.backendTargetGrid.value).toEqual(grid2)
    })

    it('should have reactive backendCascades', () => {
      const spinState = useSpinState()
      const cascades = [createMockCascade()]

      spinState.backendCascades.value = cascades
      expect(spinState.backendCascades.value).toEqual(cascades)

      spinState.backendCascades.value = []
      expect(spinState.backendCascades.value).toEqual([])
    })

    it('should have reactive currentCascadeIndex', () => {
      const spinState = useSpinState()

      spinState.currentCascadeIndex.value = 5
      expect(spinState.currentCascadeIndex.value).toBe(5)

      spinState.currentCascadeIndex.value = 0
      expect(spinState.currentCascadeIndex.value).toBe(0)
    })
  })

  describe('integration scenarios', () => {
    it('should handle complete spin cycle', () => {
      const spinState = useSpinState()

      // 1. Receive spin result
      const response = createMockSpinResponse({
        cascades: [
          createMockCascade({ win_amount: 50 }),
          createMockCascade({ win_amount: 100 })
        ]
      })
      spinState.setBackendSpinResult(response)

      // 2. Check if cascades available
      expect(spinState.hasMoreBackendCascades()).toBe(true)

      // 3. Process first cascade
      spinState.advanceToCascade(0)
      expect(spinState.currentCascade.value?.win_amount).toBe(50)

      // 4. Move to next cascade (still has more - cascade at index 1)
      spinState.currentCascadeIndex.value++
      expect(spinState.hasMoreBackendCascades()).toBe(true)

      // 5. Process second cascade
      spinState.advanceToCascade(1)
      expect(spinState.currentCascade.value?.win_amount).toBe(100)

      // 6. Move past all cascades
      spinState.currentCascadeIndex.value++
      expect(spinState.hasMoreBackendCascades()).toBe(false)

      // 7. Clear after spin complete
      spinState.clearBackendSpinData()
      expect(spinState.hasMoreBackendCascades()).toBe(false)
    })

    it('should handle spin with no wins', () => {
      const spinState = useSpinState()

      const response = createMockSpinResponse({ cascades: [] })
      spinState.setBackendSpinResult(response)

      expect(spinState.hasMoreBackendCascades()).toBe(false)
      expect(spinState.backendCascades.value).toEqual([])
    })

    it('should handle multiple spins sequentially', () => {
      const spinState = useSpinState()

      // First spin
      const response1 = createMockSpinResponse({
        cascades: [createMockCascade({ win_amount: 50 })]
      })
      spinState.setBackendSpinResult(response1)
      expect(spinState.backendCascades.value).toHaveLength(1)

      // Second spin (should replace first)
      const response2 = createMockSpinResponse({
        cascades: [
          createMockCascade({ win_amount: 100 }),
          createMockCascade({ win_amount: 150 })
        ]
      })
      spinState.setBackendSpinResult(response2)
      expect(spinState.backendCascades.value).toHaveLength(2)
      expect(spinState.currentCascadeIndex.value).toBe(0)  // Reset
    })
  })
})
