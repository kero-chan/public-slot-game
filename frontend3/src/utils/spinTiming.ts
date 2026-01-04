/**
 * Spin Timing Utilities
 * Provides functions to get correct timing values based on game state
 */

/**
 * Get current spin timing based on fastSpin setting and game mode
 * @param fastSpinEnabled - Whether fast spin is enabled
 * @param isInFreeSpinMode - Whether currently in free spin mode
 * @param normalDuration - Normal timing value
 * @param fastDuration - Fast timing value
 * @returns Appropriate timing value
 */
export function getSpinTiming(
  fastSpinEnabled: boolean,
  isInFreeSpinMode: boolean,
  normalDuration: number,
  fastDuration: number
): number {
  // Only use fast spin for normal spins (not during free spin mode)
  const shouldUseFastSpin = fastSpinEnabled && !isInFreeSpinMode
  return shouldUseFastSpin ? fastDuration : normalDuration
}

/**
 * Get current spin base duration
 */
export function getCurrentSpinBaseDuration(fastSpinEnabled: boolean, isInFreeSpinMode: boolean): number {
  return getSpinTiming(fastSpinEnabled, isInFreeSpinMode, 1500, 250)
}

/**
 * Get current spin reel stagger
 */
export function getCurrentSpinReelStagger(fastSpinEnabled: boolean, isInFreeSpinMode: boolean): number {
  return getSpinTiming(fastSpinEnabled, isInFreeSpinMode, 200, 30)
}

/**
 * Get current anticipation slowdown
 */
export function getCurrentAnticipationSlowdown(fastSpinEnabled: boolean, isInFreeSpinMode: boolean): number {
  return getSpinTiming(fastSpinEnabled, isInFreeSpinMode, 4000, 800)
}