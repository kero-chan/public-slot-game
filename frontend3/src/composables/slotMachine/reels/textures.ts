import { Texture } from 'pixi.js'
import { getTileTexture } from '@/config/spritesheet'
import { getTileBaseSymbol, istileWilden } from '@/utils/tileHelpers'

// PERFORMANCE: Memoization cache for texture lookups
// Avoids repeated string manipulation and texture creation (1200x/second → 1x)
const textureCache = new Map<string, Texture | null>()

export function getTextureForSymbol(symbol: string, useGold: boolean = false): Texture | null {
  // PERFORMANCE: Check cache first (O(1) lookup)
  const cacheKey = useGold ? `${symbol}:gold` : symbol
  if (textureCache.has(cacheKey)) {
    return textureCache.get(cacheKey) || null
  }

  // Cache miss - perform lookup
  const baseSymbol = getTileBaseSymbol(symbol)
  const isGold = useGold || istileWilden(symbol)

  // Map symbol name to tile asset key (with .webp extension for spritesheet)
  const tileKey = isGold ? `tile_${baseSymbol}_gold.webp` : `tile_${baseSymbol}.webp`

  // Get texture from spritesheet
  const texture = getTileTexture(tileKey)
  if (texture) {
    textureCache.set(cacheKey, texture)
    return texture
  }

  // If gold version not found, fallback to normal version
  if (isGold) {
    const normalKey = `tile_${baseSymbol}.webp`
    const normalTexture = getTileTexture(normalKey)
    if (normalTexture) {
      textureCache.set(cacheKey, normalTexture)
      return normalTexture
    }
  }

  // Cache null result to avoid repeated failed lookups
  textureCache.set(cacheKey, null)
  return null
}

// Optional: Clear cache if assets are reloaded
export function clearTextureCache(): void {
  textureCache.clear()
}
