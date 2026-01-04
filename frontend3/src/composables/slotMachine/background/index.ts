/**
 * Background composable for PixiJS canvas
 * Renders the full-screen background image inside the PixiJS canvas
 * Uses slotMachine background images from spritesheet.ts
 */

import { Container, Sprite, Texture } from 'pixi.js'
import { getBackgroundTexture as getSpritesheetBackgroundTexture } from '@/config/spritesheet'

/**
 * Background texture keys
 * - 'main': Default game background
 * - 'start': Start screen background
 * - 'fa', 'zhong', 'bai', 'bawan': High-value symbol win backgrounds
 */
export type BackgroundTextureKey = 'main' | 'start' | 'fa' | 'zhong' | 'bai' | 'bawan'

/**
 * Background composable interface
 */
export interface UseBackground {
  container: Container
  build: (width: number, height: number) => void
  update: (width: number, height: number) => void
  setTexture: (textureKey: BackgroundTextureKey) => void
}

/**
 * Creates a background manager for the PixiJS canvas
 * Handles loading and rendering of full-screen background images
 */
export function useBackground(): UseBackground {
  const container = new Container()
  container.zIndex = -1000 // Behind everything

  let backgroundSprite: Sprite | null = null
  let currentTextureKey: BackgroundTextureKey = 'start'
  let lastWidth = 0
  let lastHeight = 0
  let lastTextureKey: BackgroundTextureKey | null = null

  /**
   * Map texture keys to spritesheet asset keys
   */
  const textureKeyMap: Record<BackgroundTextureKey, string> = {
    'main': 'backgroundMain',
    'start': 'backgroundStart',
    'fa': 'backgroundFa',
    'zhong': 'backgroundZhong',
    'bai': 'backgroundBai',
    'bawan': 'backgroundBawan'
  }

  /**
   * Load and create background texture from spritesheet
   */
  function getBackgroundTexture(key: BackgroundTextureKey): Texture | null {
    const assetKey = textureKeyMap[key]
    return getSpritesheetBackgroundTexture(assetKey)
  }

  /**
   * Build/rebuild the background for given dimensions
   */
  function build(width: number, height: number): void {
    // Clear existing sprite
    if (backgroundSprite) {
      container.removeChild(backgroundSprite)
      backgroundSprite.destroy()
      backgroundSprite = null
    }

    const texture = getBackgroundTexture(currentTextureKey)

    if (!texture) {
      // Fallback: create a solid color background
      const fallback = new Sprite(Texture.WHITE)
      fallback.tint = 0x1a1a2e // Dark blue fallback color
      fallback.width = width
      fallback.height = height
      container.addChild(fallback)
      backgroundSprite = fallback
    } else {
      backgroundSprite = new Sprite(texture)

      // Use texture.source dimensions if available, otherwise use texture dimensions
      const texW = texture.width || texture.source?.width || width
      const texH = texture.height || texture.source?.height || height
      const textureRatio = texW / texH
      const canvasRatio = width / height

      // Both start screen and main game use "cover" mode to fill the entire screen
      // Cover mode - fill entire canvas (may crop)
      if (canvasRatio > textureRatio) {
        // Canvas is wider - scale by width
        backgroundSprite.width = width
        backgroundSprite.height = width / textureRatio
        backgroundSprite.x = 0
        backgroundSprite.y = (height - backgroundSprite.height) / 2
      } else {
        // Canvas is taller - scale by height
        backgroundSprite.height = height
        backgroundSprite.width = height * textureRatio
        backgroundSprite.x = (width - backgroundSprite.width) / 2
        backgroundSprite.y = 0
      }

      container.addChild(backgroundSprite)
    }

    lastWidth = width
    lastHeight = height
    lastTextureKey = currentTextureKey
  }

  /**
   * Update background for new dimensions or texture change
   */
  function update(width: number, height: number): void {
    // Rebuild if dimensions changed OR if texture key changed OR if never built
    const needsRebuild = width !== lastWidth ||
                         height !== lastHeight ||
                         currentTextureKey !== lastTextureKey ||
                         !backgroundSprite

    if (needsRebuild && width > 0 && height > 0) {
      build(width, height)
    }
  }

  /**
   * Switch background texture
   * Supports: 'main', 'start', 'fa', 'zhong', 'bai', 'bawan'
   */
  function setTexture(textureKey: BackgroundTextureKey): void {
    currentTextureKey = textureKey
    // Don't rebuild here - let update() handle it to avoid double builds
  }

  return {
    container,
    build,
    update,
    setTexture
  }
}
