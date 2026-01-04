// @ts-nocheck
/**
 * Asset Loading Utilities
 * Handles loading of images and audio using Pixi.js Assets API
 * Uses local assets from assets/images and assets/audios directories
 *
 * NOTE: Individual sprite images (tiles, icons, glyphs, etc.) are now loaded directly
 * from spritesheet.ts using single image imports instead of sprite atlases.
 */

import { Assets, Texture } from 'pixi.js'
import { ASSETS, initializeGameAssets as initializeAssets, areAssetsInitialized } from '@/config/assets'
import { getAllImagePaths, initializeTextureCache } from '@/config/spritesheet'

// Import background images from slotMachine folder
import backgroundMainImg from '@/assets/images/japaneseOmakase/background/background.webp'
import backgroundStartImg from '@/assets/images/japaneseOmakase/background/background_start_game.webp'

// Import start screen images
import startBtnImg from '@/assets/images/japaneseOmakase/glyphs/glyph_start_button.webp'

// Import local audio assets
import bgMusic from '@/assets/audios/background_music.mp3'
import bgMusicJackpot from '@/assets/audios/background_music_jackpot.mp3'
import gameStartAudio from '@/assets/audios/game_start.mp3'
import jackpotFinalizeAudio from '@/assets/audios/jackpot_finalize.m4a'

// Consecutive wins
import consecutive2x from '@/assets/audios/consecutive_wins/2x.mp3'
import consecutive3x from '@/assets/audios/consecutive_wins/3x.mp3'
import consecutive4x from '@/assets/audios/consecutive_wins/4x.mp3'
import consecutive5x from '@/assets/audios/consecutive_wins/5x.mp3'
import consecutive6x from '@/assets/audios/consecutive_wins/6x.mp3'
import consecutive10x from '@/assets/audios/consecutive_wins/10x.mp3'

// Win sounds
import winBai from '@/assets/audios/wins/bai.mp3'
import winZhong from '@/assets/audios/wins/zhong.mp3'
import winFa from '@/assets/audios/wins/fa.mp3'
import winLiangsuo from '@/assets/audios/wins/liangsuo.mp3'
import winLiangtong from '@/assets/audios/wins/liangtong.mp3'
import winWusuo from '@/assets/audios/wins/wusuo.mp3'
import winWutong from '@/assets/audios/wins/wutong.mp3'
import winBawan from '@/assets/audios/wins/bawan.mp3'
import winJackpot from '@/assets/audios/wins/jackpot.mp3'
import winningAnnouncement from '@/assets/audios/wins/winning_announcement.mp3'
import winningHighlight from '@/assets/audios/wins/winning_highlight.mp3'

// Effects
import lotAudio from '@/assets/audios/effect/lot.m4a'
import reelSpinAudio from '@/assets/audios/effect/reel_spin.m4a'
import reelSpinStopAudio from '@/assets/audios/effect/reel_spin_stop.m4a'
import reachBonusAudio from '@/assets/audios/effect/reach_bonus.m4a'
import jackpotStartAudio from '@/assets/audios/effect/jackpot_start.mp3'
import increaseBetAudio from '@/assets/audios/effect/increase_bet.mp3'
import decreaseBetAudio from '@/assets/audios/effect/decrease_bet.mp3'
import startButtonAudio from '@/assets/audios/effect/start_button.mp3'
import tileBreakAudio from '@/assets/audios/effect/tile_break.mp3'
import cardTransitionAudio from '@/assets/audios/effect/card_transition.mp3'
import genericUiAudio from '@/assets/audios/effect/generic_ui.mp3'
import lineWinAudio from '@/assets/audios/effect/line_win_sound.mp3'

// Background noises
import noise1 from '@/assets/audios/background_noise/noise_1.mp3'
import noise2 from '@/assets/audios/background_noise/noise_2.mp3'
import noise3 from '@/assets/audios/background_noise/noise_3.mp3'
import noise4 from '@/assets/audios/background_noise/noise_4.mp3'
import noise5 from '@/assets/audios/background_noise/noise_5.mp3'
import noise6 from '@/assets/audios/background_noise/noise_6.mp3'
import noise7 from '@/assets/audios/background_noise/noise_7.mp3'
import noise8 from '@/assets/audios/background_noise/noise_8.mp3'
import noise9 from '@/assets/audios/background_noise/noise_9.mp3'
import noise10 from '@/assets/audios/background_noise/noise_10.mp3'
import noise11 from '@/assets/audios/background_noise/noise_11.mp3'

/**
 * Progress callback function
 * @param loaded - Number of assets loaded so far
 * @param total - Total number of assets to load
 */
export type ProgressCallback = (loaded: number, total: number) => void

/**
 * Error that occurred during asset initialization
 */
export class AssetInitializationError extends Error {
  constructor(message: string, public readonly cause?: unknown) {
    super(message)
    this.name = 'AssetInitializationError'
  }
}

// Audio is now preloaded by Howler.js in useHowlerAudio.ts
// The loadAllAssets function waits for all audio to be fully loaded

// Flag to track if assets have been fully loaded (prevents double loading)
let assetsLoaded = false
let assetsLoading: Promise<void> | null = null

/**
 * Local image paths configuration
 * NOTE: Spritesheet images (backgrounds.png, glyphs.png, icons.png, tiles.png, winAnnouncements.png)
 * are no longer used. Individual sprites are now loaded directly from spritesheet.ts
 */
const LOCAL_IMAGES = {
  backgroundMain: backgroundMainImg,
  backgroundStart: backgroundStartImg,
  startBtn: startBtnImg,
}

/**
 * Local audio paths configuration
 */
const LOCAL_AUDIOS = {
  background_music: bgMusic,
  background_music_jackpot: bgMusicJackpot,
  game_start: gameStartAudio,
  consecutive_wins_2x: consecutive2x,
  consecutive_wins_3x: consecutive3x,
  consecutive_wins_4x: consecutive4x,
  consecutive_wins_5x: consecutive5x,
  consecutive_wins_6x: consecutive6x,
  consecutive_wins_10x: consecutive10x,
  win_bai: winBai,
  win_zhong: winZhong,
  win_fa: winFa,
  win_liangsuo: winLiangsuo,
  win_liangtong: winLiangtong,
  win_wusuo: winWusuo,
  win_wutong: winWutong,
  win_bawan: winBawan,
  win_jackpot: winJackpot,
  winning_announcement: winningAnnouncement,
  winning_highlight: winningHighlight,
  background_noises: [noise1, noise2, noise3, noise4, noise5, noise6, noise7, noise8, noise9, noise10, noise11],
  lot: lotAudio,
  reel_spin: reelSpinAudio,
  reel_spin_stop: reelSpinStopAudio,
  reach_bonus: reachBonusAudio,
  jackpot_finalize: jackpotFinalizeAudio,
  jackpot_start: jackpotStartAudio,
  increase_bet: increaseBetAudio,
  decrease_bet: decreaseBetAudio,
  start_button: startButtonAudio,
  tile_break: tileBreakAudio,
  card_transition: cardTransitionAudio,
  generic_ui: genericUiAudio,
  line_win_sound: lineWinAudio,
}

/**
 * Initialize game assets from local files
 * Uses assets from assets/images and assets/audios directories
 */
export async function initializeGameAssets(): Promise<void> {
  // Check if already initialized
  if (areAssetsInitialized()) {
    return
  }

  try {
    // Initialize game assets from local files
    initializeAssets(
      'Mahjong Ways', // Game name
      LOCAL_IMAGES,
      LOCAL_AUDIOS,
      {} // No videos
    )

    // NOTE: Spritesheet data is no longer needed - individual images are loaded directly
    // from spritesheet.ts using Vite's static asset imports

    console.log('✅ Assets initialized from local files')
    console.log(`🖼️ Image assets loaded: ${Object.keys(LOCAL_IMAGES).length} items`)
    console.log(`🔊 Audio assets loaded: ${Object.keys(LOCAL_AUDIOS).length} items`)
  } catch (error) {
    throw new AssetInitializationError('Failed to load local game assets', error)
  }
}

/**
 * Load all assets (images and audio) with progress tracking
 * Uses Pixi.js Assets API for images
 *
 * @param onProgress - Optional callback for progress updates
 * @returns Promise that resolves when all assets are loaded
 *
 * @example
 * await loadAllAssets((loaded, total) => {
 *   console.log(`Loading: ${loaded}/${total}`)
 * })
 */
export async function loadAllAssets(onProgress: ProgressCallback | null = null): Promise<void> {
  // If already loaded, report 100% and return immediately
  if (assetsLoaded) {
    if (onProgress) onProgress(1, 1)
    return
  }

  // If currently loading, wait for it to complete
  if (assetsLoading) {
    await assetsLoading
    if (onProgress) onProgress(1, 1)
    return
  }

  // Start loading and store the promise
  assetsLoading = (async () => {
    // Ensure assets are initialized from API first
    await initializeGameAssets()

  const paths = ASSETS.imagePaths || {}
  ASSETS.loadedImages = {}

  // Configure PixiJS Assets for high-quality loading
  // Disable preferCreateImageBitmap to ensure compatibility with high-DPI displays
  Assets.resolver.preferCreateImageBitmap = false

  // Get all slotMachine images from spritesheet.ts
  const slotMachineImages = getAllImagePaths()

  // Combine base paths with slotMachine images
  const allImagePaths = { ...paths, ...slotMachineImages }
  const entries = Object.entries(allImagePaths)

  console.log(`🖼️ Total images to load: ${entries.length} (${Object.keys(paths).length} base + ${Object.keys(slotMachineImages).length} slotMachine)`)

  // Weight-based progress: Images = 60%, Audio = 40% for better UX
  // Audio is handled by Howler with separate progress tracking
  const IMAGE_WEIGHT = 0.6
  const AUDIO_WEIGHT = 0.4
  const totalProgressUnits = 100 // Use 100 units for percentage-like tracking

  if (entries.length === 0) {
    if (onProgress) onProgress(1, 1) // Report 100% complete
    return
  }

  // Report initial progress
  if (onProgress) {
    onProgress(0, totalProgressUnits)
  }

  // Register assets with Pixi's asset loader (only if not already registered)
  for (const [alias, src] of entries) {
    // Skip empty or invalid entries
    if (!src || !alias) continue

    try {
      // Check if alias is already registered to avoid "already has key" warnings
      const existingAsset = Assets.resolver.hasKey(alias)
      if (!existingAsset) {
        Assets.add({ alias, src })
      }
      // Also register by resolved URL so Texture.from(url) can find it in cache
      const existingSrc = Assets.resolver.hasKey(src)
      if (!existingSrc && src !== alias) {
        Assets.add({ alias: src, src })
      }
    } catch {
      // Silent fail - asset registration failed
    }
  }

  // Load all images by alias with progress tracking
  // Pixi Assets.load() downloads full images and caches them as textures in memory
  let loaded: Record<string, any> = {}
  let lastReportedProgress = 0

  // Build list of all aliases to load (both key names and resolved URLs)
  const aliasesToLoad: string[] = []
  for (const [alias, src] of entries) {
    // Skip empty or invalid entries
    if (!src || !alias) continue
    aliasesToLoad.push(alias)
    if (src !== alias) {
      aliasesToLoad.push(src)
    }
  }

  try {
    loaded = await Assets.load(aliasesToLoad, (progress: number) => {
      // Pixi's progress is 0 to 1 for images only
      // Map to weighted progress (0-60% of total)
      const weightedProgress = Math.floor(progress * IMAGE_WEIGHT * totalProgressUnits)

      // Only report if progress actually increased (avoid duplicate reports)
      if (weightedProgress > lastReportedProgress) {
        lastReportedProgress = weightedProgress
        if (onProgress) {
          onProgress(weightedProgress, totalProgressUnits)
        }
      }
    })
  } catch {
    // Silent fail - image loading failed
  }

  // Normalize and store textures into ASSETS.loadedImages
  // Store by both key name AND resolved URL so lookups work either way
  for (const [alias, src] of entries) {
    // Skip empty or invalid entries
    if (!src || !alias) continue

    let tex = loaded?.[alias] || loaded?.[src] || null

    // If the loader returned something non-Texture, try to create a Texture
    if (!(tex instanceof Texture) && src) {
      try {
        tex = Texture.from(src)
      } catch {
        tex = null
      }
    }

    // Set texture quality settings for crisp rendering
    if (tex && tex.source) {
      tex.source.scaleMode = 'linear'
      tex.source.autoGenerateMipmaps = true
    }

    // Store by key name
    ASSETS.loadedImages[alias] = tex
    // Also store by resolved URL for direct URL lookups
    if (src !== alias && tex) {
      ASSETS.loadedImages[src] = tex
    }
  }

  // Report images fully loaded (60% complete)
  const imageProgressComplete = Math.floor(IMAGE_WEIGHT * totalProgressUnits)
  if (onProgress) {
    onProgress(imageProgressComplete, totalProgressUnits)
  }

  // Count unique images loaded (by alias only, not duplicated URLs)
  const successCount = entries.filter(([alias]) => ASSETS.loadedImages[alias]).length
  console.log(`🖼️ Images loaded: ${successCount}/${entries.length} (${imageProgressComplete}% progress)`)

  // Initialize Howler.js and wait for all audio to be fully loaded
  // Audio represents the remaining 40% (60% to 100%)
  try {
    const { howlerAudio } = await import('../composables/useHowlerAudio')
    howlerAudio.initialize()

    // Get total audio count for logging
    const audioCount = howlerAudio.getAudioCount()

    // Wait for all audio files to be fully loaded with progress tracking
    await howlerAudio.waitForAllLoaded((audioLoaded, audioTotal) => {
      if (onProgress && audioTotal > 0) {
        // Map audio progress to remaining 40% (60-100%)
        const audioProgress = (audioLoaded / audioTotal) * AUDIO_WEIGHT * totalProgressUnits
        const combinedProgress = imageProgressComplete + Math.floor(audioProgress)
        onProgress(combinedProgress, totalProgressUnits)
      }
    })

    console.log(`🔊 Audio preloaded: ${audioCount} files (40% progress)`)
  } catch {
    // Silent fail - Howler initialization failed
  }

  // Initialize the event-based audio system
  try {
    const { audioManager } = await import('../composables/audioManager')
    audioManager.initialize()
  } catch {
    // Silent fail - audio system initialization failed
  }

  // Initialize texture cache from loaded assets to prevent runtime network requests
  // This pre-populates the local cache in spritesheet.ts with all preloaded textures
  initializeTextureCache()

  // Final progress report - use actual totals including audio
  if (onProgress) {
    onProgress(totalProgressUnits, totalProgressUnits)
  }
  })()

  // Wait for loading to complete
  await assetsLoading
  assetsLoaded = true
}

/**
 * Simple HTML image loader for non-Pixi usage or diagnostics.
 * Returns an HTMLImageElement or null.
 *
 * @param src - Image source path
 * @returns Promise resolving to HTMLImageElement or null on error
 *
 * @example
 * const img = await loadImage('/path/to/image.png')
 * if (img) {
 *   console.log(`Image loaded: ${img.width}×${img.height}`)
 * }
 */
export function loadImage(src: string): Promise<HTMLImageElement | null> {
  return new Promise((resolve) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = () => {
      resolve(null)
    }
    img.src = src
  })
}
