import { Container, Sprite } from 'pixi.js'
import { getGlyphSprite } from '@/config/spritesheet'

// Digit to texture key mapping for glyphs
const DIGIT_TEXTURE_KEYS: Record<string, string> = {
  '0': 'glyph_0.webp',
  '1': 'glyph_1.webp',
  '2': 'glyph_2.webp',
  '3': 'glyph_3.webp',
  '4': 'glyph_4.webp',
  '5': 'glyph_5.webp',
  '6': 'glyph_6.webp',
  '7': 'glyph_7.webp',
  '8': 'glyph_8.webp',
  '9': 'glyph_9.webp',
  ',': 'glyph_comma.webp',
  '.': 'glyph_dot.webp',
}

export interface GlyphNumberDisplay {
  container: Container
  sprites: Sprite[]
  update: (value: string, targetHeight: number, align?: 'center' | 'left') => void
}

/**
 * Create a glyph-based number display
 */
export function createGlyphNumber(): GlyphNumberDisplay {
  const container = new Container()
  let sprites: Sprite[] = []

  function update(value: string, targetHeight: number, align: 'center' | 'left' = 'center'): void {
    // Guard against destroyed container (can happen when GSAP animations continue after sign out)
    if (!container || container.destroyed) return

    // Clear existing sprites
    for (const sprite of sprites) {
      sprite.destroy()
    }
    sprites = []
    container.removeChildren()

    const digits = value.split('')
    let offsetX = 0

    // Scale multipliers for different character types
    const numberScale = 1.2  // Numbers scaled up
    const punctuationScale = 0.5  // Dots and commas scaled down

    for (const d of digits) {
      const textureKey = DIGIT_TEXTURE_KEYS[d]
      if (!textureKey) continue

      const sprite = getGlyphSprite(textureKey)
      if (!sprite) continue

      // Use linear filtering with mipmaps for high-quality rendering on retina displays
      if (sprite.texture?.source) {
        sprite.texture.source.scaleMode = 'linear'
        sprite.texture.source.autoGenerateMipmaps = true
      }

      const isPunctuation = d === '.' || d === ','
      const scaleMultiplier = isPunctuation ? punctuationScale : numberScale
      const baseScale = targetHeight / sprite.height
      const finalScale = baseScale * scaleMultiplier

      sprite.anchor.set(0, 0.5)
      sprite.scale.set(finalScale)
      // Use rounded positions for crisp rendering
      sprite.x = Math.round(offsetX)
      // Adjust Y position for punctuation to align better with baseline
      sprite.y = Math.round(isPunctuation ? targetHeight * 0.35 : 0)
      sprite.roundPixels = true

      container.addChild(sprite)
      sprites.push(sprite)
      offsetX += sprite.width
    }

    // Adjust pivot based on alignment (rounded)
    if (align === 'left') {
      container.pivot.set(0, 0)  // Left-align: no pivot offset
    } else {
      container.pivot.set(Math.round(offsetX / 2), 0)  // Center-align
    }
  }

  return {
    container,
    sprites,
    update
  }
}
