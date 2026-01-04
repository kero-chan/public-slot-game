import { Texture } from 'pixi.js'
import { getGlyphTexture } from '@/config/spritesheet'

export function getMultiplierTexture(
  multiplier: number,
  _isActive: boolean,
  _inFreeSpinMode: boolean
): Texture | null {
  // Simplified: single version for all multipliers (no color variants)
  const assetKey = `glyph_x${multiplier}.webp`
  return getGlyphTexture(assetKey)
}
