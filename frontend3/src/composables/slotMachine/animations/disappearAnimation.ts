import type { GridState, WinCombination } from '@/types/global'
import { audioEvents, AUDIO_EVENTS } from '@/composables/audioEventBus'
import { howlerAudio } from '@/composables/useHowlerAudio'

/**
 * Play tile break sound with pitch randomization (0.75-1.25)
 */
function playTileBreakSound(): void {
  const howl = howlerAudio.getHowl('tile_break')
  if (!howl) return

  // Apply pitch randomization 0.75-1.25 to prevent audio fatigue
  const randomPitch = 0.75 + Math.random() * 0.5
  howl.rate(randomPitch)

  audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { audioKey: 'tile_break', volume: 0.6 })
}

/**
 * Run disappear animation for winning tiles
 */
export function animateDisappear(
  gridState: GridState,
  wins: WinCombination[],
  duration: number
): Promise<void> {
  // Extract positions from wins
  const positions: Array<[number, number]> = []

  wins.forEach(win => {
    win.positions.forEach(pos => {
      positions.push([pos.reel, pos.row])
    })
  })

  // Play tile break sound when tiles are removed
  if (positions.length > 0) {
    playTileBreakSound()
  }

  // Set disappear state
  gridState.disappearPositions = new Set(
    positions.map(([c, r]) => `${c},${r}`)
  )
  gridState.disappearAnim = { start: Date.now(), duration }

  return new Promise(resolve => {
    setTimeout(resolve, duration)
  })
}

/**
 * Clear disappear animation state
 */
export function clearDisappearState(gridState: GridState): void {
  gridState.disappearPositions = new Set()
  gridState.disappearAnim = { start: 0, duration: 0 }
}
