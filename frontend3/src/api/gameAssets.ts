import apiClient from './client'

/**
 * Spritesheet frame data
 */
export interface SpritesheetFrame {
  frame: {
    x: number
    y: number
    w: number
    h: number
  }
}

/**
 * Spritesheet JSON type
 */
export type SpritesheetJSON = Record<string, SpritesheetFrame>

/**
 * Image URLs from the API
 */
export interface ImageURLs {
  backgrounds: string
  glyphs: string
  icons: string
  tiles: string
  winAnnouncements: string
  backgroundMain: string
  backgroundStart: string
  startBtn: string
  iconAvatar: string
  iconExit: string
  iconHistory: string
}

/**
 * Audio URLs from the API
 * Keys are audio names, values can be string (single audio) or string[] (array of audios)
 */
export interface AudioURLs {
  background_music?: string
  background_music_jackpot?: string
  game_start?: string
  consecutive_wins_2x?: string
  consecutive_wins_3x?: string
  consecutive_wins_4x?: string
  consecutive_wins_5x?: string
  consecutive_wins_6x?: string
  consecutive_wins_10x?: string
  win_bai?: string
  win_zhong?: string
  win_fa?: string
  win_liangsuo?: string
  win_liangtong?: string
  win_wusuo?: string
  win_wutong?: string
  win_bawan?: string
  win_jackpot?: string
  winning_announcement?: string
  winning_highlight?: string
  background_noises?: string[]
  lot?: string
  reel_spin?: string
  reel_spin_stop?: string
  reach_bonus?: string
  jackpot_finalize?: string
  jackpot_start?: string
  [key: string]: string | string[] | undefined
}

/**
 * Video URLs from the API
 */
export interface VideoURLs {
  [key: string]: string
}

/**
 * Game assets response from API
 */
export interface GameAssetsResponse {
  id: string
  name: string
  spritesheetJson: SpritesheetJSON
  images: ImageURLs
  audios: AudioURLs
  videos: VideoURLs
}

/**
 * Fetch game assets from the API
 * @param gameId - The game ID to fetch assets for
 * @returns Promise resolving to the game assets
 * @throws Error if game ID is not provided or API call fails
 */
export async function fetchGameAssets(): Promise<GameAssetsResponse> {
  const response = await apiClient.get<GameAssetsResponse>('/game-assets')
  return response.data
}

/**
 * Get the game ID from environment variable
 * @returns The game ID or null if not configured
 */
export function getGameId(): string | null {
  const gameId = import.meta.env.VITE_GAME_ID
  return gameId || null
}
