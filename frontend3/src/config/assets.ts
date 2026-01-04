import { reactive } from 'vue'

/**
 * Image paths interface
 */
interface ImagePaths {
  [key: string]: string
}

/**
 * Loaded images interface (populated at runtime by imageLoader)
 */
interface LoadedImages {
  [key: string]: any
}

/**
 * Audio paths interface
 */
interface AudioPaths {
  background_music: string
  background_music_jackpot: string
  game_start: string
  consecutive_wins_2x: string
  consecutive_wins_3x: string
  consecutive_wins_4x: string
  consecutive_wins_5x: string
  consecutive_wins_6x: string
  consecutive_wins_10x: string
  win_bai: string
  win_zhong: string
  win_fa: string
  win_liangsuo: string
  win_liangtong: string
  win_wusuo: string
  win_wutong: string
  win_bawan: string
  win_jackpot: string
  winning_announcement: string
  winning_highlight: string
  background_noises: string[]
  lot: string
  reel_spin: string
  reel_spin_stop: string
  reach_bonus: string
  jackpot_finalize: string
  jackpot_start: string
  increase_bet: string
  decrease_bet: string
  generic_ui: string
  start_button: string
  card_transition: string
  tile_break: string
  line_win_sound: string
  [key: string]: string | string[]
}

/**
 * Video paths interface
 */
interface VideoPaths {
  [key: string]: string
}

/**
 * Assets configuration interface
 */
export interface AssetsConfig {
  gameName: string
  imagePaths: ImagePaths
  loadedImages: LoadedImages
  audioPaths: AudioPaths
  videoPaths: VideoPaths
  initialized: boolean
}

// Make ASSETS reactive so Vue components can detect changes
export const ASSETS: AssetsConfig = reactive({
  // Game name from API
  gameName: '',
  // Image paths will be populated dynamically from API
  imagePaths: {},
  loadedImages: {},
  initialized: false,
  // Audio paths - loaded from API (required)
  audioPaths: {} as AudioPaths,
  // Video paths - loaded from API
  videoPaths: {}
})

/**
 * Initialize assets from API response
 * Maps the API image/audio/video URLs to the expected keys used by the asset loader
 *
 * NOTE: Spritesheets are no longer used. Individual images are loaded directly
 * from spritesheet.ts using Vite's static asset imports.
 */
export function initializeGameAssets(
  gameName: string,
  images: {
    backgroundMain?: string
    backgroundStart?: string
    startBtn?: string
  },
  audios?: Record<string, string | string[]>,
  videos?: Record<string, string>
): void {
  // Set game name
  ASSETS.gameName = gameName
  ASSETS.imagePaths = {
    // Background images
    backgroundMain: images.backgroundMain || '',
    backgroundStart: images.backgroundStart || '',
    // Start screen assets
    startBtn: images.startBtn || '',
  }

  // Set audio paths from server (required)
  if (audios && Object.keys(audios).length > 0) {
    ASSETS.audioPaths = audios as AudioPaths
  } else {
    console.warn('⚠️ No audio assets received from server. Audio will not work.')
  }

  // Set video paths from server
  if (videos && Object.keys(videos).length > 0) {
    ASSETS.videoPaths = { ...videos }
  }

  ASSETS.initialized = true
}

/**
 * Check if assets have been initialized
 */
export function areAssetsInitialized(): boolean {
  return ASSETS.initialized
}
