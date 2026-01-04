// @ts-nocheck
/**
 * Background Music Composable - Event-Driven Architecture
 * Helper composable that emits audio events for background music
 */

import { ref, type Ref } from 'vue'
import { audioEvents, AUDIO_EVENTS } from './audioEventBus'

/**
 * Music type (normal or jackpot mode)
 */
type MusicType = 'normal' | 'jackpot'

/**
 * Background music composable interface
 */
export interface UseBackgroundMusic {
  isPlaying: Ref<boolean>
  currentMusicType: Ref<MusicType>
  gameSoundEnabled: Ref<boolean>
  
  start: () => Promise<boolean>
  stop: () => void
  pause: () => void
  resume: () => void
  setVolume: (volume: number) => void
  switchToJackpotMusic: () => void
  switchToNormalMusic: () => void
  setGameSoundEnabled: (enabled: boolean) => void
  reset: () => void
}

/**
 * Background music composable - Event-driven implementation
 * All methods emit events to the audio event bus
 */
export function useBackgroundMusic(): UseBackgroundMusic {
  
  // State management (kept for compatibility with existing code)
  const isPlaying: Ref<boolean> = ref(false)
  const currentMusicType: Ref<MusicType> = ref('normal')
  const gameSoundEnabled: Ref<boolean> = ref(true)

  /**
   * Start playing background music
   */
  const start = async (): Promise<boolean> => {
    if (isPlaying.value) {
      return true
    }

    // Emit event to start music
    audioEvents.emit(AUDIO_EVENTS.MUSIC_START)
    
    // Update local state for compatibility
    isPlaying.value = true
    currentMusicType.value = 'normal'
    
    return true
  }

  /**
   * Stop playing background music
   */
  const stop = (): void => {
    if (!isPlaying.value) return

    // Emit event to stop music
    audioEvents.emit(AUDIO_EVENTS.MUSIC_STOP)
    
    // Update local state
    isPlaying.value = false
  }

  /**
   * Set volume (0.0 to 1.0)
   */
  const setVolume = (volume: number): void => {
    // Emit event to set music volume
    audioEvents.emit(AUDIO_EVENTS.MUSIC_SET_VOLUME, { volume })
  }

  /**
   * Pause background music (for video playback)
   */
  const pause = (): void => {
    // Emit event to pause music
    audioEvents.emit(AUDIO_EVENTS.MUSIC_PAUSE)
  }

  /**
   * Resume background music (after video playback)
   */
  const resume = (): void => {
    // Emit event to resume music
    audioEvents.emit(AUDIO_EVENTS.MUSIC_RESUME)
  }

  /**
   * Switch to jackpot background music
   */
  const switchToJackpotMusic = (): void => {
    if (currentMusicType.value === 'jackpot') return

    // Emit event to switch to jackpot music
    audioEvents.emit(AUDIO_EVENTS.MUSIC_SWITCH_JACKPOT)
    
    // Update local state
    currentMusicType.value = 'jackpot'
  }

  /**
   * Switch back to normal background music
   */
  const switchToNormalMusic = (): void => {
    if (currentMusicType.value === 'normal') return

    // Emit event to switch to normal music
    audioEvents.emit(AUDIO_EVENTS.MUSIC_SWITCH_NORMAL)
    
    // Update local state
    currentMusicType.value = 'normal'
  }

  /**
   * Enable/disable game sound
   */
  const setGameSoundEnabled = (enabled: boolean): void => {
    gameSoundEnabled.value = enabled
    
    // Emit global audio control event
    if (enabled) {
      audioEvents.emit(AUDIO_EVENTS.AUDIO_ENABLE)
    } else {
      audioEvents.emit(AUDIO_EVENTS.AUDIO_DISABLE)
    }
  }

  /**
   * Reset background music state (called on logout)
   */
  const reset = (): void => {
    isPlaying.value = false
    currentMusicType.value = 'normal'
    gameSoundEnabled.value = true
  }

  return {
    // State (for compatibility)
    isPlaying,
    currentMusicType,
    gameSoundEnabled,
    
    // Actions (emit events)
    start,
    stop,
    pause,
    resume,
    setVolume,
    switchToJackpotMusic,
    switchToNormalMusic,
    setGameSoundEnabled,
    reset
  }
}
