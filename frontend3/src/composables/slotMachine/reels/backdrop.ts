import { Graphics, Sprite, Container } from 'pixi.js'
import { getBackgroundTexture } from '@/config/spritesheet'

/**
 * Rectangle configuration for backdrop
 */
export interface BackdropRect {
  x: number
  y: number
  w: number
  h: number
}

/**
 * High-value symbol types for frame theming
 */
export type FrameTheme = 'default' | 'fa' | 'zhong' | 'bai' | 'bawan'

/**
 * Backdrop composable interface
 */
export interface Backdrop {
  ensureBackdrop: (rect: BackdropRect, gameW: number) => void
  frameContainer: Container
  notificationContainer: Container
  updateNotification: (timestamp: number) => void
  setFrameTheme: (theme: FrameTheme) => void
}

export function createBackdrop(container: Container): Backdrop {
  const mask = new Graphics()
  let frameSprite: Sprite | null = null
  let currentTheme: FrameTheme = 'default'

  // Separate container for frame overlay (renders on top of tiles)
  const frameContainer = new Container()

  // Empty container (notification removed)
  const notificationContainer = new Container()

  // Don't add mask to children, just use it for masking
  container.mask = mask

  /**
   * Get the grid frame texture key based on theme
   */
  function getGridTextureKey(theme: FrameTheme): string {
    if (theme === 'default') return 'background_grid.webp'
    return `${theme}_background_grid.webp`
  }

  /**
   * Set the frame theme (for high-value symbol wins)
   */
  function setFrameTheme(theme: FrameTheme): void {
    if (theme === currentTheme) return
    currentTheme = theme

    const textureKey = getGridTextureKey(theme)
    const texture = getBackgroundTexture(textureKey)
    if (texture && frameSprite) {
      frameSprite.texture = texture
    }
  }

  function ensureBackdrop(rect: BackdropRect, gameW: number): void {
    // Create frame sprite if not exists (renders ON TOP of tiles)
    if (!frameSprite) {
      const textureKey = getGridTextureKey(currentTheme)
      const texture = getBackgroundTexture(textureKey)
      if (texture) {
        frameSprite = new Sprite(texture)
        frameContainer.addChild(frameSprite)
      }
    }

    // Update frame sprite to cover the tile grid
    // Frame design: 2536x2579, inner area: 2270x2415
    if (frameSprite && frameSprite.texture.width > 0) {
      // Frame asset dimensions
      const frameAssetWidth = 2536
      const frameAssetHeight = 2579

      // Frame larger than game width - scale down more on smartphone
      const isSmartphone = gameW < 600
      const frameScale = isSmartphone ? 1.4 : 1.5
      const frameWidth = gameW * frameScale
      const frameHeight = frameWidth * (frameAssetHeight / frameAssetWidth)

      frameSprite.width = frameWidth
      frameSprite.height = frameHeight

      // Center the INNER CONTENT area over the grid, not the frame itself
      // Grid center is at rect.x + gameW / 2
      const gridCenterX = rect.x + gameW / 2
      const gridCenterY = rect.y + rect.h / 2
      const yOffset = isSmartphone ? 5 : 20  // Raised higher on smartphone

      // Fine-tune horizontal alignment:
      // Positive xOffset = move frame RIGHT (grid appears more left)
      // Negative xOffset = move frame LEFT (grid appears more right)
      // Adjust this value to fix frame/grid alignment
      const xOffset = -5

      frameSprite.x = gridCenterX - frameWidth / 2 + xOffset
      frameSprite.y = gridCenterY - frameHeight / 2 + yOffset
    }

    // Clip reels to the grid area (top and bottom)
    const isSmartphone = gameW < 600
    const topOffset = isSmartphone ? 0 : 0  // No offset, show all rows
    const extraHeight = isSmartphone ? 20 : 70  // Tighter mask on smartphone

    mask.clear()
    mask.rect(rect.x, rect.y + topOffset, gameW, rect.h + extraHeight - topOffset)
    mask.fill(0xffffff)
  }

  function updateNotification(timestamp: number): void {
    // GSAP handles the animation, nothing needed here
  }

  return { ensureBackdrop, frameContainer, notificationContainer, updateNotification, setFrameTheme }
}
