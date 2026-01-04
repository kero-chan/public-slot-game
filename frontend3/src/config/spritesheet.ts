import { Assets, Sprite, Texture } from 'pixi.js'
import { ASSETS } from '@/config/assets'

// Import all single images from slotMachine folder

// Background images
import backgroundMain from '@/assets/images/japaneseOmakase/background/background.webp'
import backgroundStartGame from '@/assets/images/japaneseOmakase/background/background_start_game.webp'
import backgroundMultiplier from '@/assets/images/japaneseOmakase/background/background_multiplier.webp'
import backgroundGrid from '@/assets/images/japaneseOmakase/background/background_grid.webp'
import backgroundFooter from '@/assets/images/japaneseOmakase/background/background_footer.webp'
import backgroundFooterFreespin from '@/assets/images/japaneseOmakase/background/background_footer_freespin.webp'
import freeSpinBackground from '@/assets/images/japaneseOmakase/background/free_spin_background.webp'
import freeSpinRetriggerBackground from '@/assets/images/japaneseOmakase/background/free_spin_retrigger_background.webp'
import activeStatusBg from '@/assets/images/japaneseOmakase/background/active_status_bg.webp'
// High-value symbol win backgrounds
import backgroundFa from '@/assets/images/japaneseOmakase/background/fa_background.webp'
import backgroundZhong from '@/assets/images/japaneseOmakase/background/zhong_background.webp'
import backgroundBai from '@/assets/images/japaneseOmakase/background/bai_background.webp'
import backgroundBawan from '@/assets/images/japaneseOmakase/background/bawan_background.webp'
// High-value symbol frames (grid, footer, multiplier)
import faBackgroundGrid from '@/assets/images/japaneseOmakase/background/fa_background_grid.webp'
import faBackgroundFooter from '@/assets/images/japaneseOmakase/background/fa_background_footer.webp'
import faBackgroundMultiplier from '@/assets/images/japaneseOmakase/background/fa_background_multiplier.webp'
import zhongBackgroundGrid from '@/assets/images/japaneseOmakase/background/zhong_background_grid.webp'
import zhongBackgroundFooter from '@/assets/images/japaneseOmakase/background/zhong_background_footer.webp'
import zhongBackgroundMultiplier from '@/assets/images/japaneseOmakase/background/zhong_background_multiplier.webp'
import baiBackgroundGrid from '@/assets/images/japaneseOmakase/background/bai_background_grid.webp'
import baiBackgroundFooter from '@/assets/images/japaneseOmakase/background/bai_background_footer.webp'
import baiBackgroundMultiplier from '@/assets/images/japaneseOmakase/background/bai_background_multiplier.webp'
import bawanBackgroundGrid from '@/assets/images/japaneseOmakase/background/bawan_background_grid.webp'
import bawanBackgroundFooter from '@/assets/images/japaneseOmakase/background/bawan_background_footer.webp'
import bawanBackgroundMultiplier from '@/assets/images/japaneseOmakase/background/bawan_background_multiplier.webp'

// Glyph images - Multipliers (simplified naming)
import glyphX1 from '@/assets/images/japaneseOmakase/glyphs/glyph_x1.webp'
import glyphX2 from '@/assets/images/japaneseOmakase/glyphs/glyph_x2.webp'
import glyphX3 from '@/assets/images/japaneseOmakase/glyphs/glyph_x3.webp'
import glyphX4 from '@/assets/images/japaneseOmakase/glyphs/glyph_x4.webp'
import glyphX5 from '@/assets/images/japaneseOmakase/glyphs/glyph_x5.webp'
import glyphX6 from '@/assets/images/japaneseOmakase/glyphs/glyph_x6.webp'
import glyphX10 from '@/assets/images/japaneseOmakase/glyphs/glyph_x10.webp'
// Glyph images - Other
import glyphStartButton from '@/assets/images/japaneseOmakase/glyphs/glyph_start_button.webp'
import glyphSkipButton from '@/assets/images/japaneseOmakase/glyphs/glyph_skip_button.webp'
import glyphDot from '@/assets/images/japaneseOmakase/glyphs/glyph_dot.webp'
import glyph0 from '@/assets/images/japaneseOmakase/glyphs/glyph_0.webp'
import glyph1 from '@/assets/images/japaneseOmakase/glyphs/glyph_1.webp'
import glyph2 from '@/assets/images/japaneseOmakase/glyphs/glyph_2.webp'
import glyph3 from '@/assets/images/japaneseOmakase/glyphs/glyph_3.webp'
import glyph4 from '@/assets/images/japaneseOmakase/glyphs/glyph_4.webp'
import glyph5 from '@/assets/images/japaneseOmakase/glyphs/glyph_5.webp'
import glyph6 from '@/assets/images/japaneseOmakase/glyphs/glyph_6.webp'
import glyph7 from '@/assets/images/japaneseOmakase/glyphs/glyph_7.webp'
import glyph8 from '@/assets/images/japaneseOmakase/glyphs/glyph_8.webp'
import glyph9 from '@/assets/images/japaneseOmakase/glyphs/glyph_9.webp'
import glyphComma from '@/assets/images/japaneseOmakase/glyphs/glyph_comma.webp'

// Icon images
import iconClose from '@/assets/images/japaneseOmakase/icons/icon_close.webp'
import iconSpin from '@/assets/images/japaneseOmakase/icons/icon_spin.webp'
import iconDecreaseBetAmount from '@/assets/images/japaneseOmakase/icons/icon_decrease_bet_amount.webp'
import iconIncreaseBetAmount from '@/assets/images/japaneseOmakase/icons/icon_increase_bet_amount.webp'
import iconBtnGameModeSetting from '@/assets/images/japaneseOmakase/icons/icon_btn_game_mode_setting.webp'
import iconBtnToggleTurboMode from '@/assets/images/japaneseOmakase/icons/icon_btn_toggle_turbo_mode.webp'
import iconBtnSettingMenu from '@/assets/images/japaneseOmakase/icons/icon_btn_setting_menu.webp'
import iconTurboLabel from '@/assets/images/japaneseOmakase/icons/icon_turbo_label.webp'

// Game mode images
import gameModeMenuCloseBtn from '@/assets/images/japaneseOmakase/icons/icon_close.webp'
import gameModeMenuBackground from '@/assets/images/japaneseOmakase/gameMode/game_mode_menu_background.webp'
import gameModeMenuItemBackground from '@/assets/images/japaneseOmakase/gameMode/game_mode_menu_item_background.webp'
import gameModeMenuItemCoinBackground from '@/assets/images/japaneseOmakase/gameMode/game_mode_menu_item_coin_background.webp'

// Tile images
import tileBawan from '@/assets/images/japaneseOmakase/tiles/tile_bawan.webp'
import tileWutong from '@/assets/images/japaneseOmakase/tiles/tile_wutong.webp'
import tileWusuo from '@/assets/images/japaneseOmakase/tiles/tile_wusuo.webp'
import tileZhongGold from '@/assets/images/japaneseOmakase/tiles/tile_zhong_gold.webp'
import tileZhong from '@/assets/images/japaneseOmakase/tiles/tile_zhong.webp'
import tileWild from '@/assets/images/japaneseOmakase/tiles/tile_wild.webp'
import tileLiangtong from '@/assets/images/japaneseOmakase/tiles/tile_liangtong.webp'
import tileLiangsuoGold from '@/assets/images/japaneseOmakase/tiles/tile_liangsuo_gold.webp'
import tileBai from '@/assets/images/japaneseOmakase/tiles/tile_bai.webp'
import tileLiangtongGold from '@/assets/images/japaneseOmakase/tiles/tile_liangtong_gold.webp'
import tileFaGold from '@/assets/images/japaneseOmakase/tiles/tile_fa_gold.webp'
import tileWusuoGold from '@/assets/images/japaneseOmakase/tiles/tile_wusuo_gold.webp'
import tileFa from '@/assets/images/japaneseOmakase/tiles/tile_fa.webp'
import tileBonus from '@/assets/images/japaneseOmakase/tiles/tile_bonus.webp'
import tileLiangsuo from '@/assets/images/japaneseOmakase/tiles/tile_liangsuo.webp'
import tileBawanGold from '@/assets/images/japaneseOmakase/tiles/tile_bawan_gold.webp'
import tileBaiGold from '@/assets/images/japaneseOmakase/tiles/tile_bai_gold.webp'
import tileWutongGold from '@/assets/images/japaneseOmakase/tiles/tile_wutong_gold.webp'

// Win announcement images
import freeSpinsOverlayBg from '@/assets/images/japaneseOmakase/winAnnouncements/free_spins_overlay_bg.webp'
import winMega from '@/assets/images/japaneseOmakase/winAnnouncements/win_mega.webp'
import winSmall from '@/assets/images/japaneseOmakase/winAnnouncements/win_small.webp'
import winGrand from '@/assets/images/japaneseOmakase/winAnnouncements/win_grand.webp'
import winGold from '@/assets/images/japaneseOmakase/winAnnouncements/win_gold.webp'
import winJackpot from '@/assets/images/japaneseOmakase/winAnnouncements/win_jackpot.webp'
import winBig from '@/assets/images/japaneseOmakase/winAnnouncements/win_big.webp'
import winMedium from '@/assets/images/japaneseOmakase/winAnnouncements/win_medium.webp'
import winTotal from '@/assets/images/japaneseOmakase/winAnnouncements/win_total.webp'

// Winning frame images
import winningDefaultFrame from '@/assets/images/japaneseOmakase/winningFrames/winning_default_frame.webp'
import winningFaFrame from '@/assets/images/japaneseOmakase/winningFrames/winning_fa_frame.webp'
import winningZhongFrame from '@/assets/images/japaneseOmakase/winningFrames/winning_zong_frame.webp'
import winningBaiFrame from '@/assets/images/japaneseOmakase/winningFrames/winning_bai_frame.webp'
import winningBawanFrame from '@/assets/images/japaneseOmakase/winningFrames/winning_bawan_frame.webp'

// Winning high value images (front and back for flip animation)
import wonFaFront from '@/assets/images/japaneseOmakase/winningHighValue/won_fa_front.webp'
import wonFaBack from '@/assets/images/japaneseOmakase/winningHighValue/won_fa_back.webp'
import wonZhongFront from '@/assets/images/japaneseOmakase/winningHighValue/won_zhong_front.webp'
import wonZhongBack from '@/assets/images/japaneseOmakase/winningHighValue/won_zhong_back.webp'
import wonBaiFront from '@/assets/images/japaneseOmakase/winningHighValue/won_bai_front.webp'
import wonBaiBack from '@/assets/images/japaneseOmakase/winningHighValue/won_bai_back.webp'
import wonBawanFront from '@/assets/images/japaneseOmakase/winningHighValue/won_bawan_front.webp'
import wonBawanBack from '@/assets/images/japaneseOmakase/winningHighValue/won_bawan_back.webp'
import wonLiangsuoFront from '@/assets/images/japaneseOmakase/winningHighValue/won_liangsuo_front.webp'
import wonLiangsuoBack from '@/assets/images/japaneseOmakase/winningHighValue/won_liangsuo_back.webp'
import wonLiangtongFront from '@/assets/images/japaneseOmakase/winningHighValue/won_liangtong_front.webp'
import wonLiangtongBack from '@/assets/images/japaneseOmakase/winningHighValue/won_liangtong_back.webp'
import wonWusuoFront from '@/assets/images/japaneseOmakase/winningHighValue/won_wusuo_front.webp'
import wonWusuoBack from '@/assets/images/japaneseOmakase/winningHighValue/won_wusuo_back.webp'
import wonWutongFront from '@/assets/images/japaneseOmakase/winningHighValue/won_wutong_front.webp'
import wonWutongBack from '@/assets/images/japaneseOmakase/winningHighValue/won_wutong_back.webp'

// Image path mappings
const backgroundImages: Record<string, string> = {
  'background.webp': backgroundMain,
  'background_start_game.webp': backgroundStartGame,
  'background_multiplier.webp': backgroundMultiplier,
  'background_grid.webp': backgroundGrid,
  'background_footer.webp': backgroundFooter,
  'background_footer_freespin.webp': backgroundFooterFreespin,
  'free_spin_background.webp': freeSpinBackground,
  'free_spin_retrigger_background.webp': freeSpinRetriggerBackground,
  'active_status_bg.webp': activeStatusBg,
  'backgroundMain': backgroundMain,
  'backgroundStart': backgroundStartGame,
  // High-value symbol win backgrounds
  'backgroundFa': backgroundFa,
  'backgroundZhong': backgroundZhong,
  'backgroundBai': backgroundBai,
  'backgroundBawan': backgroundBawan,
  // High-value symbol frames (grid)
  'fa_background_grid.webp': faBackgroundGrid,
  'zhong_background_grid.webp': zhongBackgroundGrid,
  'bai_background_grid.webp': baiBackgroundGrid,
  'bawan_background_grid.webp': bawanBackgroundGrid,
  // High-value symbol frames (footer)
  'fa_background_footer.webp': faBackgroundFooter,
  'zhong_background_footer.webp': zhongBackgroundFooter,
  'bai_background_footer.webp': baiBackgroundFooter,
  'bawan_background_footer.webp': bawanBackgroundFooter,
  // High-value symbol frames (multiplier)
  'fa_background_multiplier.webp': faBackgroundMultiplier,
  'zhong_background_multiplier.webp': zhongBackgroundMultiplier,
  'bai_background_multiplier.webp': baiBackgroundMultiplier,
  'bawan_background_multiplier.webp': bawanBackgroundMultiplier,
}

const glyphImages: Record<string, string> = {
  // Multiplier glyphs (simplified naming)
  'glyph_x1.webp': glyphX1,
  'glyph_x2.webp': glyphX2,
  'glyph_x3.webp': glyphX3,
  'glyph_x4.webp': glyphX4,
  'glyph_x5.webp': glyphX5,
  'glyph_x6.webp': glyphX6,
  'glyph_x10.webp': glyphX10,
  // Text glyphs
  'glyph_start_button.webp': glyphStartButton,
  'glyph_skip_button.webp': glyphSkipButton,
  // Number glyphs
  'glyph_dot.webp': glyphDot,
  'glyph_0.webp': glyph0,
  'glyph_1.webp': glyph1,
  'glyph_2.webp': glyph2,
  'glyph_3.webp': glyph3,
  'glyph_4.webp': glyph4,
  'glyph_5.webp': glyph5,
  'glyph_6.webp': glyph6,
  'glyph_7.webp': glyph7,
  'glyph_8.webp': glyph8,
  'glyph_9.webp': glyph9,
  'glyph_comma.webp': glyphComma,
}

const iconImages: Record<string, string> = {
  'icon_close.webp': iconClose,
  'icon_spin.webp': iconSpin,
  'icon_decrease_bet_amount.webp': iconDecreaseBetAmount,
  'icon_increase_bet_amount.webp': iconIncreaseBetAmount,
  'icon_turbo_label.webp': iconTurboLabel,
  'icon_btn_game_mode_setting.webp': iconBtnGameModeSetting,
  'icon_btn_toggle_turbo_mode.webp': iconBtnToggleTurboMode,
  'icon_btn_setting_menu.webp': iconBtnSettingMenu,
}

const gameModeImages: Record<string, string> = {
  'game_mode_menu_close_btn.webp': gameModeMenuCloseBtn,
  'game_mode_menu_background.webp': gameModeMenuBackground,
  'game_mode_menu_item_background.webp': gameModeMenuItemBackground,
  'game_mode_menu_item_coin_background.webp': gameModeMenuItemCoinBackground,
}

const tileImages: Record<string, string> = {
  'tile_bawan.webp': tileBawan,
  'tile_wutong.webp': tileWutong,
  'tile_wusuo.webp': tileWusuo,
  'tile_zhong_gold.webp': tileZhongGold,
  'tile_zhong.webp': tileZhong,
  'tile_wild.webp': tileWild,
  'tile_liangtong.webp': tileLiangtong,
  'tile_liangsuo_gold.webp': tileLiangsuoGold,
  'tile_bai.webp': tileBai,
  'tile_liangtong_gold.webp': tileLiangtongGold,
  'tile_fa_gold.webp': tileFaGold,
  'tile_wusuo_gold.webp': tileWusuoGold,
  'tile_fa.webp': tileFa,
  'tile_bonus.webp': tileBonus,
  'tile_liangsuo.webp': tileLiangsuo,
  'tile_bawan_gold.webp': tileBawanGold,
  'tile_bai_gold.webp': tileBaiGold,
  'tile_wutong_gold.webp': tileWutongGold,
}

const winAnnouncementImages: Record<string, string> = {
  'free_spins_overlay_bg.webp': freeSpinsOverlayBg,
  'win_mega.webp': winMega,
  'win_small.webp': winSmall,
  'win_grand.webp': winGrand,
  'win_gold.webp': winGold,
  'win_jackpot.webp': winJackpot,
  'win_big.webp': winBig,
  'win_medium.webp': winMedium,
  'win_total.webp': winTotal,
}

const winningFrameImages: Record<string, string> = {
  'winning_default_frame.webp': winningDefaultFrame,
  'winning_fa_frame.webp': winningFaFrame,
  'winning_zhong_frame.webp': winningZhongFrame,
  'winning_bai_frame.webp': winningBaiFrame,
  'winning_bawan_frame.webp': winningBawanFrame,
}

const winningHighValueImages: Record<string, string> = {
  // Front images
  'won_fa_front.webp': wonFaFront,
  'won_zhong_front.webp': wonZhongFront,
  'won_bai_front.webp': wonBaiFront,
  'won_bawan_front.webp': wonBawanFront,
  'won_liangsuo_front.webp': wonLiangsuoFront,
  'won_liangtong_front.webp': wonLiangtongFront,
  'won_wusuo_front.webp': wonWusuoFront,
  'won_wutong_front.webp': wonWutongFront,
  // Back images
  'won_fa_back.webp': wonFaBack,
  'won_zhong_back.webp': wonZhongBack,
  'won_bai_back.webp': wonBaiBack,
  'won_bawan_back.webp': wonBawanBack,
  'won_liangsuo_back.webp': wonLiangsuoBack,
  'won_liangtong_back.webp': wonLiangtongBack,
  'won_wusuo_back.webp': wonWusuoBack,
  'won_wutong_back.webp': wonWutongBack,
}

// Texture cache - populated during preload phase to avoid runtime network requests
const textureCache: Record<string, Texture> = {}

// Flag to track if texture cache has been initialized from preloaded assets
let textureCacheInitialized = false

/**
 * Reset texture cache when PixiJS app is destroyed (e.g., on logout)
 * This ensures fresh textures are loaded when the app is re-initialized
 */
export function resetTextureCache(): void {
  // Clear all cached textures
  for (const key in textureCache) {
    delete textureCache[key]
  }
  textureCacheInitialized = false
}

/**
 * Set texture quality settings for crisp rendering
 */
function setTextureQuality(texture: Texture): void {
  if (texture.source) {
    // Use linear filtering for smooth scaling
    texture.source.scaleMode = 'linear'
    // Ensure mipmaps are generated for better quality at different scales
    texture.source.autoGenerateMipmaps = true
  }
}

/**
 * Initialize texture cache from preloaded ASSETS.loadedImages
 * Call this after loadAllAssets() completes to ensure all textures are cached
 * This prevents any runtime network requests for already-preloaded assets
 */
export function initializeTextureCache(): void {
  if (textureCacheInitialized) return

  let cachedCount = 0

  // First, copy ALL textures from ASSETS.loadedImages directly into textureCache
  // This ensures any texture that was preloaded is available regardless of key format
  if (ASSETS.loadedImages) {
    for (const [key, texture] of Object.entries(ASSETS.loadedImages)) {
      if (texture instanceof Texture) {
        setTextureQuality(texture)
        textureCache[key] = texture
        cachedCount++
      }
    }
  }

  // Also map our local image path keys to their resolved URLs in the cache
  // This ensures lookups work with either the key name or the resolved URL
  const allPaths = getAllImagePaths()
  for (const [key, path] of Object.entries(allPaths)) {
    // If we already have the texture by the resolved URL, also store by key
    if (textureCache[path] instanceof Texture && !textureCache[key]) {
      textureCache[key] = textureCache[path]
    }
    // If we have it by key but not by path, store by path too
    if (textureCache[key] instanceof Texture && !textureCache[path]) {
      textureCache[path] = textureCache[key]
    }

    // If still not found, try PixiJS Assets cache
    if (!textureCache[path] || !textureCache[key]) {
      try {
        let texture: Texture | null = null
        if (Assets.resolver.hasKey(path)) {
          texture = Assets.get(path)
        } else if (Assets.resolver.hasKey(key)) {
          texture = Assets.get(key)
        }
        if (texture instanceof Texture) {
          setTextureQuality(texture)
          if (!textureCache[path]) textureCache[path] = texture
          if (!textureCache[key]) textureCache[key] = texture
        }
      } catch {
        // Ignore errors
      }
    }
  }

  textureCacheInitialized = true
  console.log(`✅ Texture cache initialized: ${cachedCount} textures cached from preloaded assets`)
}

/**
 * Get texture from image path (Vite resolved URL)
 * Uses cached textures from preload phase - no network requests during gameplay
 *
 * Cache lookup order:
 * 1. Local textureCache (fastest, populated during initializeTextureCache)
 * 2. ASSETS.loadedImages (populated by imageLoader.ts during preload)
 * 3. PixiJS Assets cache (populated by Assets.load)
 * 4. Fallback to Texture.from() with warning (should not happen in normal flow)
 */
function getTexture(imagePath: string): Texture | null {
  if (!imagePath) return null

  // Check local cache first (fastest path)
  const cachedTex = textureCache[imagePath]
  if (cachedTex) {
    // Validate texture is still usable (not destroyed)
    if (!cachedTex.destroyed && cachedTex.source && !cachedTex.source.destroyed) {
      return cachedTex
    }
    // Texture was destroyed, remove from cache and re-fetch
    delete textureCache[imagePath]
  }

  // Try to get from ASSETS.loadedImages (populated by imageLoader.ts)
  const loadedTexture = ASSETS.loadedImages?.[imagePath]
  if (loadedTexture instanceof Texture && !loadedTexture.destroyed && loadedTexture.source && !loadedTexture.source.destroyed) {
    setTextureQuality(loadedTexture)
    textureCache[imagePath] = loadedTexture
    return loadedTexture
  }

  // Try to get from PixiJS Assets cache
  try {
    if (Assets.resolver.hasKey(imagePath)) {
      const cachedTexture = Assets.get(imagePath)
      if (cachedTexture instanceof Texture && !cachedTexture.destroyed && cachedTexture.source && !cachedTexture.source.destroyed) {
        setTextureQuality(cachedTexture)
        textureCache[imagePath] = cachedTexture
        return cachedTexture
      }
    }
  } catch {
    // Not in cache, continue to fallback
  }

  // Fallback to Texture.from() - this may trigger a network request
  // This should not happen in normal gameplay if preloading worked correctly
  console.warn(`[getTexture] Cache miss for: ${imagePath} - may trigger network request`)
  try {
    const texture = Texture.from(imagePath)
    setTextureQuality(texture)
    textureCache[imagePath] = texture
    return texture
  } catch (e) {
    console.error('[getTexture] Failed to create texture:', imagePath, e)
    return null
  }
}

/**
 * Get sprite from image path
 */
function getSprite(imagePath: string): Sprite | null {
  const texture = getTexture(imagePath)
  if (!texture) return null

  const sprite = new Sprite(texture)
  sprite.anchor.set(0.5)
  return sprite
}

// Glyph functions
export function getGlyphTexture(key: string): Texture | null {
  const path = glyphImages[key]
  if (!path) {
    console.warn('[Glyph] Image key not found:', key)
    return null
  }
  return getTexture(path)
}

export function getGlyphSprite(key: string): Sprite | null {
  const path = glyphImages[key]
  if (!path) {
    console.warn('Glyph image not found:', key)
    return null
  }
  return getSprite(path)
}

// Background functions
export function getBackgroundTexture(key: string): Texture | null {
  const path = backgroundImages[key]
  if (!path) {
    console.warn('Background image not found:', key)
    return null
  }
  return getTexture(path)
}

export function getBackgroundSprite(key: string): Sprite | null {
  const path = backgroundImages[key]
  if (!path) {
    console.warn('Background image not found:', key)
    return null
  }
  return getSprite(path)
}

// Icon functions
export function getIconTexture(key: string): Texture | null {
  const path = iconImages[key]
  if (!path) {
    console.warn('Icon image not found:', key)
    return null
  }
  return getTexture(path)
}

export function getIconSprite(key: string): Sprite | null {
  const path = iconImages[key]
  if (!path) {
    console.warn('Icon image not found:', key)
    return null
  }
  return getSprite(path)
}

// Game mode functions
export function getGameModeTexture(key: string): Texture | null {
  const path = gameModeImages[key]
  if (!path) {
    console.warn('Game mode image not found:', key)
    return null
  }
  return getTexture(path)
}

export function getGameModeSprite(key: string): Sprite | null {
  const path = gameModeImages[key]
  if (!path) {
    console.warn('Game mode image not found:', key)
    return null
  }
  return getSprite(path)
}

// Win announcement functions
export function getWinAnnouncementTexture(key: string): Texture | null {
  const path = winAnnouncementImages[key]
  if (!path) {
    console.warn('Win announcement image not found:', key)
    return null
  }
  return getTexture(path)
}

export function getWinAnnouncementSprite(key: string): Sprite | null {
  const path = winAnnouncementImages[key]
  if (!path) {
    console.warn('Win announcement image not found:', key)
    return null
  }
  return getSprite(path)
}

// Tile functions
export function getTileTexture(key: string): Texture | null {
  const path = tileImages[key]
  if (!path) {
    console.warn('Tile image not found:', key)
    return null
  }
  return getTexture(path)
}

export function getTileSprite(key: string): Sprite | null {
  const path = tileImages[key]
  if (!path) {
    console.warn('Tile image not found:', key)
    return null
  }
  return getSprite(path)
}

// Winning frame functions
export function getWinningFrameTexture(key: string): Texture | null {
  const path = winningFrameImages[key]
  if (!path) {
    console.warn('Winning frame image not found:', key)
    return null
  }
  return getTexture(path)
}

export function getWinningFrameSprite(key: string): Sprite | null {
  const path = winningFrameImages[key]
  if (!path) {
    console.warn('Winning frame image not found:', key)
    return null
  }
  return getSprite(path)
}

/**
 * Get winning frame texture for a symbol
 * High-value symbols (fa, zhong, bai, bawan) get special frames
 * Other symbols get the default frame
 */
export function getWinningFrameForSymbol(symbol: string): Texture | null {
  // Remove _gold suffix if present
  const baseSymbol = symbol.replace('_gold', '')

  switch (baseSymbol) {
    case 'fa':
      return getWinningFrameTexture('winning_fa_frame.webp')
    case 'zhong':
      return getWinningFrameTexture('winning_zhong_frame.webp')
    case 'bai':
      return getWinningFrameTexture('winning_bai_frame.webp')
    case 'bawan':
      return getWinningFrameTexture('winning_bawan_frame.webp')
    default:
      return getWinningFrameTexture('winning_default_frame.webp')
  }
}

// Winning high value functions
export function getWinningHighValueTexture(key: string): Texture | null {
  const path = winningHighValueImages[key]
  if (!path) {
    console.warn('Winning high value image not found:', key)
    return null
  }
  return getTexture(path)
}

export function getWinningHighValueSprite(key: string): Sprite | null {
  const path = winningHighValueImages[key]
  if (!path) {
    console.warn('Winning high value image not found:', key)
    return null
  }
  return getSprite(path)
}

/**
 * Get all image paths for preloading
 * Returns a flat object with all images from all categories
 */
export function getAllImagePaths(): Record<string, string> {
  return {
    ...backgroundImages,
    ...glyphImages,
    ...iconImages,
    ...gameModeImages,
    ...tileImages,
    ...winAnnouncementImages,
    ...winningFrameImages,
    ...winningHighValueImages,
  }
}
