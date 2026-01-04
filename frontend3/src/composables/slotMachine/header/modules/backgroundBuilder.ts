import { Container, Sprite } from 'pixi.js'
import type { HeaderRect } from './types'
import { getBackgroundTexture } from '@/config/spritesheet'

/**
 * High-value symbol types for frame theming
 */
export type FrameTheme = 'default' | 'fa' | 'zhong' | 'bai' | 'bawan'

export interface BackgroundBuildResult {
  bg02Y: number
  bg02Height: number
  multiplierBgSprite: Sprite | null
}

/**
 * Get the multiplier frame texture key based on theme
 */
function getMultiplierTextureKey(theme: FrameTheme): string {
  if (theme === 'default') return 'background_multiplier.webp'
  return `${theme}_background_multiplier.webp`
}

export function buildBackgrounds(
  container: Container,
  rect: HeaderRect,
  theme: FrameTheme = 'default'
): BackgroundBuildResult {
  const { x, y, w } = rect

  let bg02Y = y
  let bg02Height = 0
  let multiplierBgSprite: Sprite | null = null

  // Multiplier background strip (scaled down by 10%)
  const textureKey = getMultiplierTextureKey(theme)
  const bg02Texture = getBackgroundTexture(textureKey)
  if (bg02Texture && bg02Texture.width > 0 && bg02Texture.height > 0) {
    const bg02 = new Sprite(bg02Texture)
    const scaledWidth = w * 0.9  // 10% smaller
    bg02.x = x + (w - scaledWidth) / 2  // Center horizontally
    bg02.width = scaledWidth
    // Maintain aspect ratio
    bg02.height = (bg02Texture.height / bg02Texture.width) * scaledWidth
    // Position at top
    bg02.y = y - bg02.height * 0.02
    container.addChild(bg02)

    bg02Y = bg02.y
    bg02Height = bg02.height
    multiplierBgSprite = bg02
  }

  return { bg02Y, bg02Height, multiplierBgSprite }
}

/**
 * Update the multiplier background texture based on theme
 */
export function updateMultiplierBackground(
  sprite: Sprite | null,
  theme: FrameTheme
): void {
  if (!sprite) return

  const textureKey = getMultiplierTextureKey(theme)
  const texture = getBackgroundTexture(textureKey)
  if (texture) {
    sprite.texture = texture
  }
}
