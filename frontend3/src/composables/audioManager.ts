/**
 * Global Audio Manager Singleton - Event-Driven Architecture
 * Provides centralized access to audio controls via events
 */

import { useBackgroundMusic, type UseBackgroundMusic } from './useBackgroundMusic'
import { audioEvents, AUDIO_EVENTS } from './audioEventBus'
import { audioPlayer } from './audioPlayer'

/**
 * Audio Manager Class
 * Centralized control for all audio in the application
 */
class AudioManager {
  private backgroundMusic: UseBackgroundMusic | null = null
  private gameSoundEnabled: boolean = true

  /**
   * Initialize the audio manager and audio player
   * @returns Background music instance for compatibility
   */
  initialize(): UseBackgroundMusic {
    // Initialize the audio player (sets up event listeners)
    audioPlayer.initialize()

    if (!this.backgroundMusic) {
      this.backgroundMusic = useBackgroundMusic()
      // Sync the initial gameSound state
      this.backgroundMusic.setGameSoundEnabled(this.gameSoundEnabled)
    }

    return this.backgroundMusic
  }

  /**
   * Set game sound enabled state
   * @param enabled - Whether game sound is enabled
   */
  setGameSoundEnabled(enabled: boolean): void {
    this.gameSoundEnabled = enabled

    // Update background music state for compatibility
    if (this.backgroundMusic) {
      this.backgroundMusic.setGameSoundEnabled(enabled)
    }
    
    // Also emit global audio control events
    if (enabled) {
      audioEvents.emit(AUDIO_EVENTS.AUDIO_ENABLE)
    } else {
      audioEvents.emit(AUDIO_EVENTS.AUDIO_DISABLE)
    }
  }

  /**
   * Get current game sound state
   * @returns Whether game sound is enabled
   */
  isGameSoundEnabled(): boolean {
    return this.gameSoundEnabled
  }

  /**
   * Check if audio system is ready
   * @returns Whether audio system is ready
   */
  isAudioReady(): boolean {
    return audioPlayer.isReady()
  }

  /**
   * Get background music instance (for compatibility)
   * @returns Background music instance or null
   */
  getInstance(): UseBackgroundMusic | null {
    return this.backgroundMusic
  }

  /**
   * Pause background music
   */
  pause(): void {
    audioEvents.emit(AUDIO_EVENTS.MUSIC_PAUSE)
  }

  /**
   * Resume background music
   */
  resume(): void {
    audioEvents.emit(AUDIO_EVENTS.MUSIC_RESUME)
  }

  /**
   * Start background music
   */
  start(): void {
    audioEvents.emit(AUDIO_EVENTS.MUSIC_START)
  }

  /**
   * Stop background music
   */
  stop(): void {
    audioEvents.emit(AUDIO_EVENTS.MUSIC_STOP)
  }

  /**
   * Stop all audio (background music and all sound effects)
   */
  stopAll(): void {
    // Stop background music
    this.stop()

    // Stop all Howler sounds (if needed for compatibility)
    import('./useHowlerAudio').then(({ howlerAudio }) => {
      howlerAudio.stopAll()
    }).catch(() => {})
  }

  /**
   * Reset audio system (called on logout)
   * Clears all audio state and stops all playback
   */
  reset(): void {
    // Reset background music composable state
    if (this.backgroundMusic) {
      this.backgroundMusic.reset()
    }
    
    // Reset audio player state
    audioPlayer.reset()
  }

  /**
   * Switch to jackpot music
   */
  switchToJackpotMusic(): void {
    audioEvents.emit(AUDIO_EVENTS.MUSIC_SWITCH_JACKPOT)
  }

  /**
   * Switch to normal music
   */
  switchToNormalMusic(): void {
    audioEvents.emit(AUDIO_EVENTS.MUSIC_SWITCH_NORMAL)
  }

  /**
   * Set global volume
   * @param volume - Volume level (0.0 to 1.0)
   */
  setGlobalVolume(volume: number): void {
    audioEvents.emit(AUDIO_EVENTS.AUDIO_SET_GLOBAL_VOLUME, { volume })
  }

  /**
   * Set music volume
   * @param volume - Volume level (0.0 to 1.0)
   */
  setMusicVolume(volume: number): void {
    audioEvents.emit(AUDIO_EVENTS.MUSIC_SET_VOLUME, { volume })
  }
}

// Export singleton instance
export const audioManager = new AudioManager()
