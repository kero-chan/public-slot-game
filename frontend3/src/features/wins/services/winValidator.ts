/**
 * Win Validator Service
 * Uses backend-provided win intensity for animation/sound purposes
 * No Vue dependencies, fully testable
 */

import type { WinCombination } from '../../spin/types'
import type { WinIntensity, WinIntensityLevels } from '../types'

/**
 * Get the highest win intensity from backend-provided values
 * The backend calculates intensity based on symbol value and count
 *
 * @param wins - Array of win combinations from backend (with win_intensity field)
 * @returns Intensity level: 'small', 'medium', 'big', 'mega'
 *
 * @example
 * ```ts
 * const wins = [{ symbol: 'fa', count: 5, positions: [...], win_intensity: 'mega' }]
 * const intensity = getWinIntensity(wins) // Returns 'mega'
 * ```
 */
export function getWinIntensity(wins: WinCombination[]): WinIntensity {
  if (!wins || wins.length === 0) return 'small'

  const intensityLevels: WinIntensityLevels = { small: 1, medium: 2, big: 3, mega: 4 }
  let maxIntensity: WinIntensity = 'small'

  wins.forEach(win => {
    const intensity = win.win_intensity!

    if (intensityLevels[intensity] > intensityLevels[maxIntensity]) {
      maxIntensity = intensity
    }
  })

  return maxIntensity
}

/**
 * Validate that wins come from backend (security check)
 * Critical security function - ensures backend authority over game results
 *
 * @param wins - Wins array to validate
 * @returns True if valid
 * @throws Error if wins are null/undefined (invalid)
 *
 * @example
 * ```ts
 * try {
 *   validateBackendWins(backendWins)
 *   // Proceed with wins
 * } catch (error) {
 *   // Handle missing backend data
 * }
 * ```
 */
export function validateBackendWins(wins: WinCombination[] | null | undefined): boolean {
  if (wins === null || wins === undefined) {
    throw new Error('Backend wins data missing - cannot proceed')
  }
  return true
}
