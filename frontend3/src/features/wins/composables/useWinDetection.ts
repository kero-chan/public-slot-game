/**
 * Win Detection Composable
 * Manages backend win retrieval and validation
 *
 * SECURITY: All wins come from backend. No client-side calculation allowed.
 */

import type { UseSpinState, WinCombination } from '../../spin/types'
import { useGameStore } from '@/stores'

/**
 * Get wins from backend cascade data
 * Critical security function - only backend provides wins
 *
 * @param spinState - Spin state composable instance
 * @returns Array of win objects from backend
 * @throws Error if backend cascade data is missing
 *
 * @example
 * ```ts
 * const spinState = useSpinState()
 * const wins = getBackendWins(spinState)
 * // Process wins from backend
 * ```
 */
export function getBackendWins(spinState: UseSpinState): WinCombination[] {
  const { backendCascades, currentCascadeIndex, advanceToCascade } = spinState

  // ❌ CRITICAL ERROR: No backend data at all (null/undefined) - this should NEVER happen
  if (!backendCascades.value) {
    console.error('❌ CRITICAL SECURITY ERROR: No backend cascade data!')
    console.error('Backend MUST provide all game results. Client-side calculation is FORBIDDEN.')
    console.error('See /specs/09-security-architecture.md for details.')
    throw new Error('Backend cascade data missing - cannot proceed with spin')
  }

  // Empty array is valid - means no wins on this spin
  if (backendCascades.value.length === 0) {
    advanceToCascade(null)
    return [] // No wins
  }

  // Return next cascade if available
  if (currentCascadeIndex.value < backendCascades.value.length) {
    const cascade = backendCascades.value[currentCascadeIndex.value]
    console.log(
      `📊 Using backend cascade ${currentCascadeIndex.value + 1}/${backendCascades.value.length}:`,
      cascade
    )

    // Store current cascade and advance index
    advanceToCascade(currentCascadeIndex.value)
    spinState.currentCascadeIndex.value++

    // Set multiplier from backend cascade data (authoritative source)
    if (cascade.multiplier !== undefined) {
      const gameStore = useGameStore()
      gameStore.setCurrentMultiplier(cascade.multiplier)
    }

    return cascade.wins || []
  }

  // All cascades have been processed
  console.log(`✅ All ${backendCascades.value.length} backend cascades processed`)
  advanceToCascade(null)
  return [] // No more cascades
}

/**
 * ❌ DEPRECATED - DO NOT USE
 * @deprecated This function is FORBIDDEN for security reasons
 * @security Client-side win calculation is a CRITICAL SECURITY VIOLATION
 * @throws Error always throws - backend provides all wins
 */
export function findWinningCombinations_DEPRECATED_FORBIDDEN(): never {
  console.error('❌ SECURITY VIOLATION: findWinningCombinations called!')
  console.error('This function is DEPRECATED and FORBIDDEN.')
  console.error('Backend provides all win data. Client-side calculation is NOT ALLOWED.')
  console.error('See /specs/09-security-architecture.md')
  throw new Error('findWinningCombinations is forbidden - backend calculates all wins')
}

/**
 * Bridge function that redirects to backend wins
 * Provides compatibility with older code while enforcing security
 *
 * @param spinState - Spin state composable instance
 * @returns Win array from backend
 *
 * @example
 * ```ts
 * const spinState = useSpinState()
 * const wins = findWinningCombinations(spinState)
 * ```
 */
export function findWinningCombinations(spinState: UseSpinState): WinCombination[] {
  return getBackendWins(spinState)
}
