import { ref } from 'vue'
import type { UseSpinState, SpinResponse, CascadeData, Grid } from '../types'

/**
 * Spin State Management Composable
 * Manages backend target grid and cascade data
 *
 * This composable maintains the authoritative state received from the backend
 * and provides methods to navigate through cascades.
 *
 * @returns Spin state management object
 *
 * @example
 * ```ts
 * const spinState = useSpinState()
 *
 * // After receiving backend response
 * spinState.setBackendSpinResult(response)
 *
 * // Check if more cascades to process
 * if (spinState.hasMoreBackendCascades()) {
 *   spinState.advanceToCascade(0)
 * }
 * ```
 */
export function useSpinState(): UseSpinState {
  // Backend target grid - set by spin() before animateSpin()
  const backendTargetGrid = ref<Grid | null>(null)

  // Backend cascades - contains wins for each cascade
  const backendCascades = ref<CascadeData[]>([])

  // Current cascade index being processed
  const currentCascadeIndex = ref<number>(0)

  // Current cascade object (for accessing grid_after)
  const currentCascade = ref<CascadeData | null>(null)

  /**
   * Set backend spin result
   * Initializes state with data from backend spin response
   *
   * @param response - Backend spin response with grid and cascades
   */
  function setBackendSpinResult(response: SpinResponse): void {
    backendTargetGrid.value = response.grid
    backendCascades.value = response.cascades || []
    currentCascadeIndex.value = 0
    console.log(
      `🎲 Server spin result: cascades=${backendCascades.value.length}, spin_total_win=${response.spin_total_win ?? 0}, free_session_total_win=${response.free_session_total_win ?? 0}`
    )
  }

  /**
   * Clear backend spin data
   * Resets all state to initial values
   */
  function clearBackendSpinData(): void {
    backendTargetGrid.value = null
    backendCascades.value = []
    currentCascadeIndex.value = 0
    currentCascade.value = null
  }

  /**
   * Check if we have more backend cascades to process
   *
   * @returns True if there are more cascades, false otherwise
   */
  function hasMoreBackendCascades(): boolean {
    return Boolean(
      backendCascades.value &&
      currentCascadeIndex.value < backendCascades.value.length
    )
  }

  /**
   * Move to specified cascade
   * Updates currentCascade with the cascade at the given index
   *
   * @param index - Index of cascade to advance to (or null to clear)
   */
  function advanceToCascade(index: number | null): void {
    if (index === null) {
      currentCascade.value = null
      currentCascadeIndex.value = 0
      return
    }

    currentCascadeIndex.value = index
    if (backendCascades.value && backendCascades.value[index]) {
      currentCascade.value = backendCascades.value[index]
    } else {
      currentCascade.value = null
    }
  }

  return {
    // State
    backendTargetGrid,
    backendCascades,
    currentCascadeIndex,
    currentCascade,

    // Actions
    setBackendSpinResult,
    clearBackendSpinData,
    hasMoreBackendCascades,
    advanceToCascade,
  }
}
