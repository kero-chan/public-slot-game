/**
 * Audio Player Module - Event-Driven Architecture
 * Standalone audio handler that listens to events and manages playback
 * Uses Howler.js for audio playback
 */

import { audioEvents, AUDIO_EVENTS, type AudioEventData, type AudioEventType } from './audioEventBus'
import { howlerAudio, type HowlerAudioElement } from './useHowlerAudio'
import { ASSETS } from '@/config/assets'

/**
 * Win data interface
 */
interface WinData {
  symbol: string | number
  count: number
}

/**
 * Base volume configuration
 */
interface BaseVolumeConfig {
  music: number
  gameStart: number
  noise: number
  effects: number
  wins: number
}

/**
 * Audio element type (can be either HTMLAudioElement or HowlerAudioElement)
 */
type AudioElement = HTMLAudioElement | HowlerAudioElement

/**
 * Audio Player Class
 * Listens to all audio events and manages audio playback
 */
class AudioPlayer {
  private isInitialized: boolean = false
  private isAudioEnabled: boolean = true
  private globalVolume: number = 1.0
  private musicVolume: number = 0.7
  
  // Audio state management
  private currentMusic: AudioElement | null = null
  private currentMusicType: 'normal' | 'jackpot' = 'normal'
  private isPlaying: boolean = false
  private shouldBePlayingMusic: boolean = false  // Track if music should be playing when audio is enabled
  private winningAnnouncementAudio: AudioElement | null = null
  
  // Background music timers
  private gameStartTimeout: NodeJS.Timeout | null = null
  private noiseStartTimeout: NodeJS.Timeout | null = null
  private noiseInterval: NodeJS.Timeout | null = null

  // Track highest multiplier sound played in current spin session
  // x5 (consecutiveWins >= 3 in base) and x10 (consecutiveWins >= 3 in free spin) should only play once
  private highestMultiplierSoundPlayed: 'none' | '5x' | '10x' = 'none'
  
  // Base volumes for different audio types
  private baseVolume: BaseVolumeConfig = {
    music: 0.7,
    gameStart: 1.0,  // Raised from 0.8 to 1.0
    noise: 0.3,
    effects: 0.6,
    wins: 1.0
  }

  /**
   * Initialize audio player and setup event listeners
   */
  initialize(): void {
    if (this.isInitialized) return

    this.setupEventListeners()
    this.setupVisibilityHandling()
    this.isInitialized = true
  }

  /**
   * Setup visibility change handling to pause/resume audio when tab is hidden/visible
   * This prevents background tabs from consuming CPU resources with audio timers
   */
  private setupVisibilityHandling(): void {
    if (typeof document === 'undefined') return

    document.addEventListener('visibilitychange', () => {
      if (document.hidden) {
        this.handleTabHidden()
      } else {
        this.handleTabVisible()
      }
    })
  }

  /**
   * Handle tab becoming hidden - pause music and clear timers
   */
  private handleTabHidden(): void {
    // Pause current music (but don't reset shouldBePlayingMusic state)
    if (this.currentMusic && !this.currentMusic.paused) {
      this.currentMusic.pause()
    }

    // Clear all background timers to prevent callbacks in background
    if (this.gameStartTimeout) {
      clearTimeout(this.gameStartTimeout)
      this.gameStartTimeout = null
    }
    if (this.noiseStartTimeout) {
      clearTimeout(this.noiseStartTimeout)
      this.noiseStartTimeout = null
    }
    this.stopNoiseLoop()

    // Pause winning announcement if playing
    if (this.winningAnnouncementAudio && !this.winningAnnouncementAudio.paused) {
      this.winningAnnouncementAudio.pause()
    }
  }

  /**
   * Handle tab becoming visible - resume music if it should be playing
   */
  private handleTabVisible(): void {
    // Only resume if audio is enabled and music should be playing
    if (!this.isAudioEnabled) return

    // Resume music if it was playing before
    if (this.shouldBePlayingMusic && this.currentMusic && this.currentMusic.paused) {
      this.currentMusic.play().catch(() => {})
      // Restart noise loop if music is playing
      this.startNoiseLoop()
    }

    // Resume winning announcement if it was playing
    if (this.winningAnnouncementAudio && this.winningAnnouncementAudio.paused) {
      this.winningAnnouncementAudio.play().catch(() => {})
    }
  }

  /**
   * Setup all event listeners for audio events
   */
  private setupEventListeners(): void {
    // Background Music Events
    audioEvents.on(AUDIO_EVENTS.MUSIC_START, () => this.handleMusicStart())
    audioEvents.on(AUDIO_EVENTS.MUSIC_STOP, () => this.handleMusicStop())
    audioEvents.on(AUDIO_EVENTS.MUSIC_PAUSE, () => this.handleMusicPause())
    audioEvents.on(AUDIO_EVENTS.MUSIC_RESUME, () => this.handleMusicResume())
    audioEvents.on(AUDIO_EVENTS.MUSIC_SWITCH_JACKPOT, () => this.handleMusicSwitchJackpot())
    audioEvents.on(AUDIO_EVENTS.MUSIC_SWITCH_NORMAL, () => this.handleMusicSwitchNormal())
    audioEvents.on(AUDIO_EVENTS.MUSIC_SET_VOLUME, (data) => this.handleMusicSetVolume(data))

    // Sound Effect Events
    audioEvents.on(AUDIO_EVENTS.EFFECT_PLAY, (data) => this.handleEffectPlay(data))
    audioEvents.on(AUDIO_EVENTS.EFFECT_STOP, (data) => this.handleEffectStop(data))
    audioEvents.on(AUDIO_EVENTS.EFFECT_FADE, (data) => this.handleEffectFade(data))
    audioEvents.on(AUDIO_EVENTS.EFFECT_WIN, (data) => this.handleEffectWin(data))
    audioEvents.on(AUDIO_EVENTS.EFFECT_CONSECUTIVE_WIN, (data) => this.handleEffectConsecutiveWin(data))
    audioEvents.on(AUDIO_EVENTS.EFFECT_WINNING_ANNOUNCEMENT, () => this.handleWinningAnnouncement())
    audioEvents.on(AUDIO_EVENTS.EFFECT_WINNING_ANNOUNCEMENT_STOP, () => this.handleWinningAnnouncementStop())
    audioEvents.on(AUDIO_EVENTS.EFFECT_WINNING_HIGHLIGHT, () => this.handleWinningHighlight())
    audioEvents.on(AUDIO_EVENTS.EFFECT_LINE_WIN, () => this.handleLineWin())

    // Global Control Events
    audioEvents.on(AUDIO_EVENTS.AUDIO_ENABLE, () => this.handleAudioEnable())
    audioEvents.on(AUDIO_EVENTS.AUDIO_DISABLE, () => this.handleAudioDisable())
    audioEvents.on(AUDIO_EVENTS.AUDIO_SET_GLOBAL_VOLUME, (data) => this.handleSetGlobalVolume(data))
  }

  // ============ Background Music Handlers ============

  /**
   * Handle music start event
   */
  private async handleMusicStart(): Promise<void> {
    if (this.isPlaying) return

    // Always set intended state when start is requested (even if audio disabled)
    this.shouldBePlayingMusic = true
    this.currentMusicType = 'normal'

    // Check if audio is enabled before starting
    if (!this.isAudioEnabled) {
      return
    }

    try {
      // Stop any existing audio first
      if (this.currentMusic) {
        this.currentMusic.pause()
        this.currentMusic = null
      }

      // Get background music audio
      const audio = this.getAudio('background_music')
      if (!audio) {
        return
      }

      // Set playing immediately after we have audio
      this.isPlaying = true

      audio.volume = this.calculateVolume(this.baseVolume.music)
      audio.loop = true

      this.currentMusic = audio

      // Handle errors
      audio.addEventListener('error', () => {
        this.isPlaying = false
      })

      await audio.play()

      // Play game start sound after 2 seconds (only if audio is enabled)
      this.gameStartTimeout = setTimeout(() => {
        if (this.isAudioEnabled) {
          this.playGameStartSound()
        }
      }, 2000)

      // Start background noise loop after 10 seconds (only if audio is enabled)
      this.noiseStartTimeout = setTimeout(() => {
        if (this.isAudioEnabled) {
          this.startNoiseLoop()
        }
      }, 10000)

    } catch {
      this.isPlaying = false
    }
  }

  /**
   * Handle music stop event
   */
  private handleMusicStop(): void {
    this.isPlaying = false
    this.shouldBePlayingMusic = false

    // Clear timeouts
    if (this.gameStartTimeout) {
      clearTimeout(this.gameStartTimeout)
      this.gameStartTimeout = null
    }
    if (this.noiseStartTimeout) {
      clearTimeout(this.noiseStartTimeout)
      this.noiseStartTimeout = null
    }

    // Stop noise loop
    this.stopNoiseLoop()

    if (this.currentMusic) {
      this.currentMusic.pause()
      this.currentMusic = null
    }
  }

  /**
   * Handle music pause event
   */
  private handleMusicPause(): void {
    if (this.currentMusic && !this.currentMusic.paused) {
      this.currentMusic.pause()
    }
    this.stopNoiseLoop()
  }

  /**
   * Handle music resume event
   */
  private handleMusicResume(): void {
    if (this.currentMusic && this.isPlaying && this.currentMusic.paused) {
      this.currentMusic.play().catch(() => {})
    }
    if (this.isPlaying) {
      this.startNoiseLoop()
    }
  }

  /**
   * Handle switch to jackpot music
   */
  private handleMusicSwitchJackpot(): void {
    // Only skip if already jackpot AND music is actually playing
    if (this.currentMusicType === 'jackpot' && this.currentMusic && this.isPlaying) {
      return
    }

    // Set intended state regardless of audio enabled status
    this.shouldBePlayingMusic = true
    this.currentMusicType = 'jackpot'

    // Don't switch if audio is disabled, but keep the intended state
    if (!this.isAudioEnabled) {
      return
    }

    // Stop current music
    if (this.currentMusic) {
      this.currentMusic.pause()
      this.currentMusic = null
    }

    // Stop noise loop when switching to jackpot music
    this.stopNoiseLoop()

    try {
      const audio = this.getAudio('background_music_jackpot')
      if (!audio) {
        return
      }

      audio.volume = this.calculateVolume(this.baseVolume.music)
      audio.loop = true

      this.currentMusic = audio
      this.currentMusicType = 'jackpot'

      audio.addEventListener('error', () => {
        this.isPlaying = false
      })

      audio.play().then(() => {
        this.isPlaying = true
      }).catch(() => {
        this.isPlaying = false
      })

    } catch {
      // Silent fail
    }
  }

  /**
   * Handle switch to normal music
   */
  private handleMusicSwitchNormal(): void {
    // Only skip if already normal AND music is actually playing
    if (this.currentMusicType === 'normal' && this.currentMusic && this.isPlaying) {
      return
    }

    // Set intended state regardless of audio enabled status
    this.shouldBePlayingMusic = true
    this.currentMusicType = 'normal'

    // Don't switch if audio is disabled, but keep the intended state
    if (!this.isAudioEnabled) {
      return
    }

    // Stop any existing audio
    if (this.currentMusic) {
      this.currentMusic.pause()
      this.currentMusic = null
    }

    try {
      const audio = this.getAudio('background_music')
      if (!audio) {
        return
      }

      audio.volume = this.calculateVolume(this.baseVolume.music)
      audio.loop = true

      this.currentMusic = audio

      audio.addEventListener('error', () => {
        this.isPlaying = false
      })

      audio.play().then(() => {
        this.isPlaying = true
      }).catch(() => {
        this.isPlaying = false
      })

    } catch {
      // Silent fail
    }
  }

  /**
   * Handle set music volume
   */
  private handleMusicSetVolume(data: AudioEventData[typeof AUDIO_EVENTS.MUSIC_SET_VOLUME]): void {
    if (!data || typeof data.volume !== 'number') return
    
    this.musicVolume = Math.max(0, Math.min(1, data.volume))
    
    if (this.currentMusic) {
      this.currentMusic.volume = this.calculateVolume(this.musicVolume)
    }
  }

  // ============ Sound Effect Handlers ============

  /**
   * Handle generic effect play
   */
  private handleEffectPlay(data: AudioEventData[typeof AUDIO_EVENTS.EFFECT_PLAY]): void {
    if (!data || !data.audioKey) return

    // Don't play effect if audio is disabled
    if (!this.isAudioEnabled) {
      return
    }

    const { audioKey, volume = 0.6 } = data
    this.playEffect(audioKey, volume)
  }

  /**
   * Handle generic effect stop
   */
  private handleEffectStop(data: AudioEventData[typeof AUDIO_EVENTS.EFFECT_STOP]): void {
    if (!data || !data.audioKey) return

    // Stop using Howler if available
    if (howlerAudio.isReady()) {
      howlerAudio.stop(data.audioKey)
    }
  }

  /**
   * Handle effect fade
   */
  private handleEffectFade(data: AudioEventData[typeof AUDIO_EVENTS.EFFECT_FADE]): void {
    if (!data || !data.audioKey) return

    const { audioKey, from, to, duration } = data

    // Fade using Howler if available
    if (howlerAudio.isReady()) {
      howlerAudio.fade(audioKey, from, to, duration)
    }
  }

  /**
   * Handle win sound effect
   */
  private handleEffectWin(data: AudioEventData[typeof AUDIO_EVENTS.EFFECT_WIN]): void {
    if (!data || !data.wins || !Array.isArray(data.wins)) return

    // Don't play effect if audio is disabled
    if (!this.isAudioEnabled) {
      return
    }

    const wins = data.wins
    if (wins.length === 0) return

    let audioKey: string | null = null
    
    // Convert wins to symbol strings for lookup
    const symbolStrings = wins.map(win => {
      if (typeof win.symbol === 'number') {
        return this.numberToSymbol(win.symbol)
      }
      return win.symbol as string
    })

    // Check for jackpot first
    const hasJackpot = wins.some((win, index) => {
      const symbolStr = symbolStrings[index]
      return win.count >= 3 && symbolStr === 'bonus'
    })

    if (hasJackpot) {
      audioKey = 'win_jackpot'
    } else {
      // Play sound for the highest priority symbol
      const symbolAudioMap: Record<string, string> = {
        fa: 'win_fa',
        fa_gold: 'win_fa',
        zhong: 'win_zhong',
        zhong_gold: 'win_zhong',
        bai: 'win_bai',
        bai_gold: 'win_bai',
        bawan: 'win_bawan',
        bawan_gold: 'win_bawan',
        wusuo: 'win_wusuo',
        wusuo_gold: 'win_wusuo',
        wutong: 'win_wutong',
        wutong_gold: 'win_wutong',
        liangsuo: 'win_liangsuo',
        liangsuo_gold: 'win_liangsuo',
        liangtong: 'win_liangtong',
        liangtong_gold: 'win_liangtong'
      }

      const symbolPriority = [
        'fa', 'fa_gold', 'zhong', 'zhong_gold', 'bai', 'bai_gold',
        'bawan', 'bawan_gold', 'wusuo', 'wusuo_gold', 'wutong', 'wutong_gold',
        'liangsuo', 'liangsuo_gold', 'liangtong', 'liangtong_gold'
      ]

      for (const symbol of symbolPriority) {
        const hasSymbol = symbolStrings.includes(symbol)
        if (hasSymbol && symbolAudioMap[symbol]) {
          audioKey = symbolAudioMap[symbol]
          break
        }
      }
    }

    if (audioKey) {
      this.playEffect(audioKey, this.baseVolume.wins)
    }
  }

  /**
   * Handle consecutive win sound effect
   * x5 and x10 sounds only play once per spin session (consecutiveWins >= 3)
   */
  private handleEffectConsecutiveWin(data: AudioEventData[typeof AUDIO_EVENTS.EFFECT_CONSECUTIVE_WIN]): void {
    if (!data || typeof data.consecutiveWins !== 'number') return

    // Don't play effect if audio is disabled
    if (!this.isAudioEnabled) {
      return
    }

    const { consecutiveWins, isFreeSpin = false } = data
    let audioKey: string | null = null

    // Reset the multiplier tracking when consecutive wins reset to 0 (new spin)
    if (consecutiveWins === 0) {
      this.highestMultiplierSoundPlayed = 'none'
      return // No sound for first cascade (consecutiveWins = 0)
    }

    if (isFreeSpin) {
      if (consecutiveWins === 1) {
        audioKey = 'consecutive_wins_4x'
      } else if (consecutiveWins === 2) {
        audioKey = 'consecutive_wins_6x'
      } else if (consecutiveWins >= 3) {
        // x10 sound - only play once per spin session
        if (this.highestMultiplierSoundPlayed !== '10x') {
          audioKey = 'consecutive_wins_10x'
          this.highestMultiplierSoundPlayed = '10x'
        }
        // If already played, don't play again (audioKey stays null)
      }
    } else {
      if (consecutiveWins === 1) {
        audioKey = 'consecutive_wins_2x'
      } else if (consecutiveWins === 2) {
        audioKey = 'consecutive_wins_3x'
      } else if (consecutiveWins >= 3) {
        // x5 sound - only play once per spin session
        if (this.highestMultiplierSoundPlayed !== '5x' && this.highestMultiplierSoundPlayed !== '10x') {
          audioKey = 'consecutive_wins_5x'
          this.highestMultiplierSoundPlayed = '5x'
        }
        // If already played, don't play again (audioKey stays null)
      }
    }

    if (audioKey) {
      this.playEffect(audioKey, this.baseVolume.effects)
    }
  }

  /**
   * Handle winning announcement
   */
  private handleWinningAnnouncement(): void {
    // Don't play announcement if audio is disabled
    if (!this.isAudioEnabled) {
      return
    }

    // Stop any existing announcement first
    this.handleWinningAnnouncementStop()

    try {
      this.winningAnnouncementAudio = this.getAudio('winning_announcement')
      if (!this.winningAnnouncementAudio) {
        return
      }

      this.winningAnnouncementAudio.volume = this.calculateVolume(0.7)
      this.winningAnnouncementAudio.loop = true

      this.winningAnnouncementAudio.addEventListener('error', () => {})

      this.winningAnnouncementAudio.play().catch(() => {})

    } catch {
      // Silent fail
    }
  }

  /**
   * Handle stop winning announcement
   */
  private handleWinningAnnouncementStop(): void {
    if (this.winningAnnouncementAudio) {
      try {
        this.winningAnnouncementAudio.pause()
        // Only set currentTime if it's a regular HTMLAudioElement
        if ('currentTime' in this.winningAnnouncementAudio && typeof this.winningAnnouncementAudio.currentTime === 'number') {
          this.winningAnnouncementAudio.currentTime = 0
        }
        this.winningAnnouncementAudio = null
      } catch {
        // Silent fail
      }
    }
  }

  /**
   * Handle winning highlight
   */
  private handleWinningHighlight(): void {
    // Don't play highlight if audio is disabled
    if (!this.isAudioEnabled) {
      return
    }

    this.playEffect('winning_highlight', 0.5)
  }

  /**
   * Handle line win sound
   * Played after a win when the winning line is shown
   * Volume is relatively quiet (0.3) as per requirements
   */
  private handleLineWin(): void {
    // Don't play if audio is disabled
    if (!this.isAudioEnabled) {
      return
    }

    this.playEffect('line_win_sound', 0.5)
  }

  // ============ Global Control Handlers ============

  /**
   * Handle audio enable
   */
  private handleAudioEnable(): void {
    this.isAudioEnabled = true
    this.updateAllVolumes()

    // If music should be playing, restart it with correct type based on CURRENT game mode
    if (this.shouldBePlayingMusic && !this.isPlaying) {
      // Check actual game mode from store instead of relying on cached currentMusicType
      import('@/stores/game/freeSpinsStore').then(({ useFreeSpinsStore }) => {
        const freeSpinsStore = useFreeSpinsStore()
        if (freeSpinsStore.inFreeSpinMode) {
          this.handleMusicSwitchJackpot()
        } else {
          this.handleMusicStart()
        }
      }).catch(() => {
        // Fallback to normal music if store import fails
        this.handleMusicStart()
      })
    }
  }

  /**
   * Handle audio disable
   */
  private handleAudioDisable(): void {
    this.isAudioEnabled = false
    
    // Stop all audio immediately when disabled but preserve intended state
    if (this.currentMusic) {
      this.currentMusic.pause()
      this.currentMusic = null
    }
    
    // Stop background noise loop
    this.stopNoiseLoop()
    
    // Clear any pending timeouts
    if (this.gameStartTimeout) {
      clearTimeout(this.gameStartTimeout)
      this.gameStartTimeout = null
    }
    if (this.noiseStartTimeout) {
      clearTimeout(this.noiseStartTimeout)
      this.noiseStartTimeout = null
    }
    
    // Stop winning announcement if playing
    if (this.winningAnnouncementAudio) {
      this.winningAnnouncementAudio.pause()
      this.winningAnnouncementAudio = null
    }
    
    // Reset playing state
    this.isPlaying = false
    
    this.updateAllVolumes()
  }

  /**
   * Handle set global volume
   */
  private handleSetGlobalVolume(data: AudioEventData[typeof AUDIO_EVENTS.AUDIO_SET_GLOBAL_VOLUME]): void {
    if (!data || typeof data.volume !== 'number') return
    
    this.globalVolume = Math.max(0, Math.min(1, data.volume))
    this.updateAllVolumes()
  }

  // ============ Helper Methods ============

  /**
   * Get audio element (Howler or HTML5)
   */
  private getAudio(audioKey: string): AudioElement | null {
    // Try Howler first (best for mobile)
    if (howlerAudio.isReady()) {
      const audio = howlerAudio.createAudioElement(audioKey)
      if (audio) return audio
    }

    // Fallback to preloaded audio
    const preloadedAudio = (ASSETS as any).loadedAudios?.[audioKey]
    if (preloadedAudio) {
      return preloadedAudio.cloneNode()
    }

    // Last resort: create from path
    const audioPath = (ASSETS as any).audioPaths?.[audioKey]
    if (audioPath) {
      return new Audio(audioPath)
    }

    return null
  }

  /**
   * Play a sound effect
   */
  private playEffect(audioKey: string, baseVolume: number = 0.6): void {
    try {
      const audio = this.getAudio(audioKey)
      if (!audio) return

      // Adjust volume for specific effects
      let adjustedVolume = baseVolume
      if (audioKey === 'reel_spin' || audioKey === 'reel_spin_stop') {
        adjustedVolume = 0.9 // 90% volume for reel sounds
      } else if (audioKey === 'reach_bonus') {
        adjustedVolume = 1.0 // Raised volume for reach_bonus
      } else if (audioKey === 'game_start') {
        adjustedVolume = 1.0 // Raised volume for game_start
      }

      audio.volume = this.calculateVolume(adjustedVolume)

      audio.addEventListener('error', () => {})

      audio.play().catch(() => {})

    } catch {
      // Silent fail
    }
  }

  /**
   * Play game start sound
   */
  private playGameStartSound(): void {
    // Don't play if audio is disabled
    if (!this.isAudioEnabled) {
      return
    }

    try {
      const gameStartAudio = this.getAudio('game_start')
      if (!gameStartAudio) {
        return
      }

      gameStartAudio.volume = this.calculateVolume(this.baseVolume.gameStart)

      gameStartAudio.addEventListener('error', () => {})

      gameStartAudio.play().catch(() => {})

    } catch {
      // Silent fail
    }
  }

  /**
   * Start background noise loop
   */
  private startNoiseLoop(): void {
    if (this.noiseInterval) return // Already running

    // Don't start noise loop if audio is disabled
    if (!this.isAudioEnabled) {
      return
    }

    const playRandomNoise = () => {
      // Don't play noise if audio is disabled
      if (!this.isAudioEnabled) {
        return
      }

      try {
        const noiseIndex = Math.floor(Math.random() * 11) // 0-10 for background_noises_0 to background_noises_10
        const noiseKey = `background_noises_${noiseIndex}`

        const noiseAudio = this.getAudio(noiseKey)
        if (!noiseAudio) return

        noiseAudio.volume = this.calculateVolume(this.baseVolume.noise)

        noiseAudio.play().catch(() => {})

      } catch {
        // Silent fail
      }
    }

    // Start immediately
    playRandomNoise()

    // Then repeat every 5-15 seconds
    this.noiseInterval = setInterval(() => {
      const delay = Math.random() * 10000 + 5000 // 5-15 seconds
      setTimeout(playRandomNoise, delay)
    }, 15000)
  }

  /**
   * Stop background noise loop
   */
  private stopNoiseLoop(): void {
    if (this.noiseInterval) {
      clearInterval(this.noiseInterval)
      this.noiseInterval = null
    }
  }

  /**
   * Calculate final volume based on global settings
   */
  private calculateVolume(baseVolume: number): number {
    if (!this.isAudioEnabled) return 0
    return baseVolume * this.globalVolume
  }

  /**
   * Update volumes for all currently playing audio
   */
  private updateAllVolumes(): void {
    if (this.currentMusic) {
      this.currentMusic.volume = this.calculateVolume(this.baseVolume.music)
    }
  }

  /**
   * Convert number to symbol string
   */
  private numberToSymbol(symbolNumber: number): string {
    const symbolMap: Record<number, string> = {
      1: 'fa', 2: 'zhong', 3: 'bai', 4: 'bawan',
      5: 'wusuo', 6: 'wutong', 7: 'liangsuo', 8: 'liangtong',
      9: 'wild', 10: 'bonus',
      11: 'fa_gold', 12: 'zhong_gold', 13: 'bai_gold', 14: 'bawan_gold',
      15: 'wusuo_gold', 16: 'wutong_gold', 17: 'liangsuo_gold', 18: 'liangtong_gold'
    }
    return symbolMap[symbolNumber] || 'unknown'
  }

  /**
   * Reset audio player state (called on logout)
   * Clears all music state and stops all playback
   */
  reset(): void {
    // Stop and clear current music (properly stop Howler instances)
    if (this.currentMusic) {
      // Check if it's a Howler wrapper and stop the underlying Howl
      if ('_isHowlerWrapper' in this.currentMusic && (this.currentMusic as any)._howl) {
        (this.currentMusic as any)._howl.stop()
      }
      this.currentMusic.pause()
      this.currentMusic = null
    }

    // Stop winning announcement (properly stop Howler instances)
    if (this.winningAnnouncementAudio) {
      if ('_isHowlerWrapper' in this.winningAnnouncementAudio && (this.winningAnnouncementAudio as any)._howl) {
        (this.winningAnnouncementAudio as any)._howl.stop()
      }
      this.winningAnnouncementAudio.pause()
      this.winningAnnouncementAudio = null
    }

    // Stop all Howler sounds to ensure clean state
    import('./useHowlerAudio').then(({ howlerAudio }) => {
      howlerAudio.stopAll()
    }).catch(() => {})

    // Clear all timers
    if (this.gameStartTimeout) {
      clearTimeout(this.gameStartTimeout)
      this.gameStartTimeout = null
    }
    if (this.noiseStartTimeout) {
      clearTimeout(this.noiseStartTimeout)
      this.noiseStartTimeout = null
    }
    this.stopNoiseLoop()

    // Reset state flags
    this.currentMusicType = 'normal'
    this.isPlaying = false
    this.shouldBePlayingMusic = false
    this.highestMultiplierSoundPlayed = 'none'
  }

  /**
   * Check if audio system is ready
   */
  isReady(): boolean {
    return this.isInitialized
  }

  /**
   * Check if audio is enabled
   */
  isEnabled(): boolean {
    return this.isAudioEnabled
  }
}

// Export singleton instance
export const audioPlayer = new AudioPlayer()