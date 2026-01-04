/**
 * Howler.js Audio Manager
 * Industry-standard solution for web audio with mobile compatibility
 */

import { Howl, Howler } from 'howler'
import { ASSETS } from '@/config/assets'

/**
 * HTML Audio Element-like interface for Howler wrapper
 */
export interface HowlerAudioElement {
  _howl: Howl
  _volume: number
  _loop: boolean
  volume: number
  loop: boolean
  paused: boolean
  currentTime: number
  play(): Promise<void>
  pause(): void
  addEventListener(event: string, handler: EventListener): void
  removeEventListener(event: string, handler: EventListener): void
  _isHowlerWrapper: boolean
}

/**
 * Howler.js Audio Manager Class
 *
 * Howler.js automatically handles mobile audio restrictions:
 * - Unlocks audio on first user interaction
 * - Works across all browsers
 * - Handles autoplay policies
 */
class HowlerAudioManager {
  private howls: Record<string, Howl> = {}
  private isInitialized: boolean = false
  private isUnlocked: boolean = false

  constructor() {
    // Handle page visibility for mobile power management and background tab optimization
    if (typeof document !== 'undefined') {
      document.addEventListener('visibilitychange', () => {
        if (document.hidden) {
          // Page is now hidden - suspend AudioContext to save resources
          this.suspendAudioContext()
        } else {
          // Page became visible - resume AudioContext if suspended
          this.resumeAudioContext()
        }
      })
    }
  }

  /**
   * Suspend AudioContext when tab becomes hidden to save CPU
   */
  async suspendAudioContext(): Promise<void> {
    const ctx = Howler.ctx
    if (ctx && ctx.state === 'running') {
      try {
        await ctx.suspend()
      } catch {
        // Silent fail
      }
    }
  }

  /**
   * Resume AudioContext if it got suspended (e.g., after tab switch or AFK)
   */
  async resumeAudioContext(): Promise<void> {
    const ctx = Howler.ctx
    if (ctx && ctx.state === 'suspended') {
      try {
        await ctx.resume()
      } catch {
        // Silent fail
      }
    }
  }

  /**
   * Initialize all audio as Howl instances
   * Call this after assets are loaded
   */
  initialize(): void {
    if (this.isInitialized) {
      return
    }

    if (!ASSETS.audioPaths) {
      return
    }

    // Get audio paths and flatten arrays (like background_noises)
    const audioEntries: [string, string][] = []
    Object.entries(ASSETS.audioPaths).forEach(([key, value]) => {
      if (Array.isArray(value)) {
        // Handle arrays (e.g., background_noises)
        value.forEach((path: string, index: number) => {
          audioEntries.push([`${key}_${index}`, path])
        })
      } else {
        audioEntries.push([key, value as string])
      }
    })

    // Create Howl instances using original file paths
    audioEntries.forEach(([key, src]) => {
      try {
        if (!src) {
          return
        }

        // Determine if this should use HTML5 Audio (for background music/long sounds)
        // or Web Audio API (for sound effects)
        const isLongAudio = key.includes('background_music') || key.includes('winning_announcement')
        const isLoop = key.includes('background_music') || key.includes('background_music_jackpot')

        // Create Howl instance
        this.howls[key] = new Howl({
          src: [src],
          preload: true,
          html5: false,
          pool: isLongAudio ? 2 : 5, // Smaller pool for long sounds, larger for effects
          loop: isLoop,
          onloaderror: () => {},
          onplayerror: () => {}
        })
      } catch {
        // Silent fail
      }
    })

    this.isInitialized = true

    // Enable Howler's built-in autoUnlock (handles Web Audio API automatically)
    Howler.autoUnlock = true
    Howler.html5PoolSize = 10 // Increase HTML5 audio pool to prevent "pool exhausted" errors
  }

  /**
   * Get a Howl instance
   * @param audioKey - Key of the audio to get
   * @returns Howl instance or null
   */
  getHowl(audioKey: string): Howl | null {
    const howl = this.howls[audioKey]

    if (!howl) {
      return null
    }

    return howl
  }

  /**
   * Play audio
   * @param audioKey - Key of the audio to play
   * @param volume - Volume level (0.0 to 1.0)
   * @param loop - Whether to loop the audio
   * @returns Sound ID or null
   */
  play(audioKey: string, volume: number = 1.0, loop: boolean = false): number | null {
    const howl = this.getHowl(audioKey)

    if (!howl) {
      return null
    }

    howl.volume(volume)
    howl.loop(loop)

    return howl.play()
  }

  /**
   * Stop audio
   * @param audioKey - Key of the audio to stop
   * @param soundId - Specific sound ID to stop (optional)
   */
  stop(audioKey: string, soundId: number | null = null): void {
    const howl = this.getHowl(audioKey)

    if (!howl) {
      return
    }

    if (soundId !== null) {
      howl.stop(soundId)
    } else {
      howl.stop()
    }
  }

  /**
   * Fade audio volume over time
   * @param audioKey - Key of the audio to fade
   * @param from - Starting volume (0.0 to 1.0)
   * @param to - Target volume (0.0 to 1.0)
   * @param duration - Duration in milliseconds
   * @param soundId - Specific sound ID to fade (optional)
   */
  fade(audioKey: string, from: number, to: number, duration: number, soundId: number | null = null): void {
    const howl = this.getHowl(audioKey)

    if (!howl) {
      return
    }

    // If soundId provided, fade that specific sound, otherwise fade all
    if (soundId !== null) {
      howl.fade(from, to, duration, soundId)
    } else {
      howl.fade(from, to, duration)
    }

    // If fading to 0, stop the sound after fade completes
    if (to === 0) {
      setTimeout(() => {
        if (soundId !== null) {
          howl.stop(soundId)
        } else {
          howl.stop()
        }
      }, duration)
    }
  }

  /**
   * Stop all audio
   */
  stopAll(): void {
    Object.values(this.howls).forEach(howl => {
      howl.stop()
    })
  }

  /**
   * Set global volume
   * @param volume - Volume level (0.0 to 1.0)
   */
  setVolume(volume: number): void {
    Howler.volume(volume)
  }

  /**
   * Check if Howler is ready
   */
  isReady(): boolean {
    return this.isInitialized
  }

  /**
   * Wait for all audio files to be fully loaded
   * @param onProgress - Optional callback for progress updates (loaded, total)
   * @returns Promise that resolves when all audio is loaded
   */
  async waitForAllLoaded(onProgress?: (loaded: number, total: number) => void): Promise<void> {
    const howlEntries = Object.entries(this.howls)
    const total = howlEntries.length

    if (total === 0) {
      onProgress?.(0, 0)
      return
    }

    let loadedCount = 0
    onProgress?.(0, total)

    const loadPromises = howlEntries.map(([key, howl]) => {
      return new Promise<void>((resolve) => {
        // Check if already loaded
        if (howl.state() === 'loaded') {
          loadedCount++
          onProgress?.(loadedCount, total)
          resolve()
          return
        }

        // Wait for load event
        const onLoad = () => {
          loadedCount++
          onProgress?.(loadedCount, total)
          resolve()
        }

        const onError = () => {
          // Still count as "loaded" to not block progress
          loadedCount++
          onProgress?.(loadedCount, total)
          resolve()
        }

        howl.once('load', onLoad)
        howl.once('loaderror', onError)

        // Timeout fallback - don't block forever if audio fails to load
        setTimeout(() => {
          if (howl.state() !== 'loaded') {
            loadedCount++
            onProgress?.(loadedCount, total)
          }
          resolve()
        }, 10000) // 10 second timeout per audio file
      })
    })

    await Promise.all(loadPromises)
  }

  /**
   * Get the total number of audio files
   */
  getAudioCount(): number {
    return Object.keys(this.howls).length
  }

  /**
   * Unlock audio context after user gesture (e.g., Start button click)
   * Required for mobile browsers
   */
  async unlockAudioContext(): Promise<void> {
    try {
      // Check if Howler is initialized
      if (!this.isInitialized) {
        return
      }

      // Step 1: Resume Web Audio API context
      const ctx = Howler.ctx
      if (ctx) {
        if (ctx.state === 'suspended') {
          await ctx.resume()
        }
      }

      // Step 2: Unlock only critical sounds - Howler will unlock others automatically
      // This makes unlock instant instead of taking 2 seconds
      const criticalSounds = ['background_music', 'background_music_jackpot', 'game_start']
      let unlockedCount = 0

      for (const key of criticalSounds) {
        const howl = this.howls[key]
        if (howl) {
          try {
            const vol = howl.volume()
            howl.volume(0)
            const id = howl.play()

            // Wait a tiny bit to ensure playback starts
            await new Promise(resolve => setTimeout(resolve, 10))

            howl.stop(id)
            howl.volume(vol)
            unlockedCount++
          } catch (err) {
            // Silently continue
          }
        }
      }

      this.isUnlocked = true
    } catch {
      this.isUnlocked = true
    }
  }

  /**
   * Create an HTMLAudioElement-like wrapper for Howler
   * This allows existing code to work without changes
   * @param audioKey - Key of the audio
   * @returns Audio element wrapper or null
   */
  createAudioElement(audioKey: string): HowlerAudioElement | null {
    const howl = this.getHowl(audioKey)

    if (!howl) {
      return null
    }

    let soundId: number | null = null
    let isPaused = false

    // Return object that mimics HTMLAudioElement API
    return {
      _howl: howl,
      _volume: 1.0,
      _loop: false,

      set volume(val: number) {
        this._volume = val
        if (soundId !== null) {
          howl.volume(val, soundId)
        }
      },

      get volume(): number {
        return this._volume
      },

      set loop(val: boolean) {
        this._loop = val
        if (soundId !== null) {
          howl.loop(val, soundId)
        }
      },

      get loop(): boolean {
        return this._loop
      },

      get paused(): boolean {
        return isPaused
      },

      set currentTime(val: number) {
        if (soundId !== null) {
          howl.seek(val, soundId)
        }
      },

      get currentTime(): number {
        if (soundId !== null) {
          return howl.seek(soundId) as number || 0
        }
        return 0
      },

      play(): Promise<void> {
        // If paused, resume the existing sound
        if (isPaused && soundId !== null) {
          howl.play(soundId)
          isPaused = false
          return Promise.resolve()
        }

        // Stop previous instance if exists and not paused
        if (soundId !== null) {
          howl.stop(soundId)
        }

        // Play new instance
        soundId = howl.play()

        // Apply settings
        howl.volume(this._volume, soundId)
        howl.loop(this._loop, soundId)

        isPaused = false

        // Return a promise for compatibility
        return Promise.resolve()
      },

      pause(): void {
        if (soundId !== null) {
          howl.pause(soundId)
          isPaused = true
        }
      },

      addEventListener(event: string, handler: EventListener): void {
        if (event === 'ended' && soundId !== null) {
          howl.on('end', handler as any, soundId)
        }
        if (event === 'error' && soundId !== null) {
          howl.on('loaderror', handler as any, soundId)
          howl.on('playerror', handler as any, soundId)
        }
      },

      removeEventListener(event: string, handler: EventListener): void {
        if (soundId !== null) {
          howl.off(event, handler as any, soundId)
        }
      },

      // Flag to identify this as a Howler wrapper
      _isHowlerWrapper: true
    }
  }
}

// Export singleton
export const howlerAudio = new HowlerAudioManager()
