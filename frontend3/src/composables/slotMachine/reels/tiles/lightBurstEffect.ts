// @ts-nocheck
import { Sprite, Container, Graphics, RenderTexture, Application } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'

// Configuration for the light burst effect - enhanced for bonus tiles
const CONFIG = {
  // Ray configuration
  rayCount: 48,              // More rays for fuller effect
  rayLengthRatio: 0.65,      // Longer rays extending beyond tile
  rayWidthBase: 6,           // Wider rays at center (pixels)
  rayWidthTip: 0,            // Width at the tip (tapered)

  // Colors (bright golden/white theme for maximum shine)
  innerColor: 0xffffff,      // Bright white center
  midColor: 0xffec8b,        // Light golden yellow (brighter)
  outerColor: 0xffd700,      // Golden fade

  // Animation
  rotationSpeed: 0.001,      // Slightly faster rotation
  pulseSpeed: 1500,          // Faster pulse cycle (ms)
  pulseMin: 0.7,             // Higher minimum alpha
  pulseMax: 1.0,             // Full alpha at peak

  // Glow ring
  glowRings: 4,              // More glow rings for fuller effect
  glowAlpha: 0.5,            // Stronger glow

  // Size
  sizeMultiplier: 1.6,       // Larger size relative to tile
}

// Cached textures for rays and glow - generated once per size
const rayTextureCache = new Map<number, RenderTexture>()
const glowTextureCache = new Map<number, RenderTexture>()
let pixiApp: Application | null = null

/**
 * Light burst entry for caching - now uses Sprites instead of Graphics
 */
interface LightBurstEntry {
  container: Container
  raysSprite: Sprite        // Sprite using cached ray texture
  glowSprite: Sprite        // Sprite using cached glow texture
  baseSize: number
}

/**
 * Light burst manager interface
 */
export interface LightBurstManager {
  container: Container
  updateBurst: (key: string, tileSprite: Sprite, shouldShow: boolean, timestamp: number) => void
  cleanup: (usedKeys: Set<string>) => void
  clear: () => void
  initTextures: (app: Application) => void
}

/**
 * Draw tapered light rays to a graphics object (called once to generate texture)
 */
function drawRaysGraphics(size: number): Graphics {
  const graphics = new Graphics()
  const rayLength = size * CONFIG.rayLengthRatio
  const angleStep = (Math.PI * 2) / CONFIG.rayCount

  for (let i = 0; i < CONFIG.rayCount; i++) {
    const angle = i * angleStep  // No rotation - we'll rotate the sprite

    // Alternate ray lengths for variety
    const lengthMod = i % 2 === 0 ? 1.0 : 0.7
    const actualLength = rayLength * lengthMod

    // Calculate ray points (tapered triangle)
    const cos = Math.cos(angle)
    const sin = Math.sin(angle)
    const perpCos = Math.cos(angle + Math.PI / 2)
    const perpSin = Math.sin(angle + Math.PI / 2)

    const halfWidth = CONFIG.rayWidthBase / 2

    // Three points: two at base, one at tip
    const x1 = perpCos * halfWidth
    const y1 = perpSin * halfWidth
    const x2 = -perpCos * halfWidth
    const y2 = -perpSin * halfWidth
    const x3 = cos * actualLength
    const y3 = sin * actualLength

    // Draw ray with gradient-like effect using multiple layers - enhanced brightness
    // Outer layer (golden)
    graphics.poly([x1, y1, x2, y2, x3, y3])
    graphics.fill({ color: CONFIG.outerColor, alpha: 0.6 })

    // Middle layer (light golden) - slightly smaller
    const midScale = 0.85
    graphics.poly([
      x1 * midScale, y1 * midScale,
      x2 * midScale, y2 * midScale,
      x3 * midScale, y3 * midScale
    ])
    graphics.fill({ color: CONFIG.midColor, alpha: 0.8 })

    // Inner layer (white core) - smallest, brightest
    const innerScale = 0.5
    graphics.poly([
      x1 * innerScale, y1 * innerScale,
      x2 * innerScale, y2 * innerScale,
      x3 * innerScale * 0.6, y3 * innerScale * 0.6
    ])
    graphics.fill({ color: CONFIG.innerColor, alpha: 1.0 })
  }

  return graphics
}

/**
 * Draw concentric glow rings to a graphics object (called once to generate texture)
 * Enhanced with brighter, more prominent glow
 */
function drawGlowGraphics(size: number): Graphics {
  const graphics = new Graphics()
  const baseRadius = size * 0.55  // Slightly larger glow

  for (let i = CONFIG.glowRings; i >= 1; i--) {
    const ringRadius = baseRadius * (0.3 + (i / CONFIG.glowRings) * 0.7)
    // Enhanced alpha - brighter inner rings
    const ringAlpha = CONFIG.glowAlpha * (1.2 - (i - 1) / CONFIG.glowRings * 0.8)

    // Gradient from white center to golden edge
    let color: number
    if (i === 1) {
      color = CONFIG.innerColor  // White center
    } else if (i === 2) {
      color = 0xfffacd  // Lemon chiffon - very light yellow
    } else if (i === 3) {
      color = CONFIG.midColor  // Light golden
    } else {
      color = CONFIG.outerColor  // Golden
    }

    graphics.circle(0, 0, ringRadius)
    graphics.fill({ color, alpha: Math.min(ringAlpha, 0.7) })
  }

  return graphics
}

/**
 * Get or create cached ray texture for a given size
 */
function getRayTexture(size: number): RenderTexture | null {
  if (!pixiApp) return null

  // Round size to nearest 10 to reduce cache entries
  const roundedSize = Math.round(size / 10) * 10

  if (!rayTextureCache.has(roundedSize)) {
    const graphics = drawRaysGraphics(roundedSize)
    const texture = pixiApp.renderer.generateTexture({
      target: graphics,
      resolution: 1,
    })
    graphics.destroy()
    rayTextureCache.set(roundedSize, texture)
  }

  return rayTextureCache.get(roundedSize)!
}

/**
 * Get or create cached glow texture for a given size
 */
function getGlowTexture(size: number): RenderTexture | null {
  if (!pixiApp) return null

  // Round size to nearest 10 to reduce cache entries
  const roundedSize = Math.round(size / 10) * 10

  if (!glowTextureCache.has(roundedSize)) {
    const graphics = drawGlowGraphics(roundedSize)
    const texture = pixiApp.renderer.generateTexture({
      target: graphics,
      resolution: 1,
    })
    graphics.destroy()
    glowTextureCache.set(roundedSize, texture)
  }

  return glowTextureCache.get(roundedSize)!
}

/**
 * Creates a new light burst effect container using cached textures
 */
function createBurstEffect(size: number): LightBurstEntry {
  const burstContainer = new Container()
  burstContainer.sortableChildren = true

  // Create glow sprite (behind rays)
  const glowTexture = getGlowTexture(size)
  const glowSprite = new Sprite(glowTexture || undefined)
  glowSprite.anchor.set(0.5)
  glowSprite.zIndex = 0
  glowSprite.blendMode = BLEND_MODES.ADD
  burstContainer.addChild(glowSprite)

  // Create rays sprite
  const rayTexture = getRayTexture(size)
  const raysSprite = new Sprite(rayTexture || undefined)
  raysSprite.anchor.set(0.5)
  raysSprite.zIndex = 1
  raysSprite.blendMode = BLEND_MODES.ADD
  burstContainer.addChild(raysSprite)

  return {
    container: burstContainer,
    raysSprite,
    glowSprite,
    baseSize: size
  }
}

/**
 * Creates rotating golden light burst effects for bonus tiles
 * OPTIMIZED: Uses pre-rendered textures and sprite rotation instead of redrawing
 */
export function createLightBurstManager(): LightBurstManager {
  const container = new Container()
  container.sortableChildren = true

  const burstCache = new Map<string, LightBurstEntry>()

  /**
   * Initialize textures - call once when app is ready
   */
  function initTextures(app: Application): void {
    pixiApp = app
  }

  /**
   * Update or create light burst for a bonus tile
   */
  function updateBurst(key: string, tileSprite: Sprite, shouldShow: boolean, timestamp: number): void {
    if (!tileSprite) return

    let entry = burstCache.get(key)

    if (shouldShow) {
      const tileSize = Math.max(tileSprite.width, tileSprite.height)
      const burstSize = tileSize * CONFIG.sizeMultiplier

      if (!entry) {
        entry = createBurstEffect(burstSize)
        container.addChild(entry.container)
        burstCache.set(key, entry)
      }

      // Update size if tile size changed significantly
      if (Math.abs(entry.baseSize - burstSize) > 10) {
        entry.baseSize = burstSize
        // Update textures
        const rayTexture = getRayTexture(burstSize)
        const glowTexture = getGlowTexture(burstSize)
        if (rayTexture) entry.raysSprite.texture = rayTexture
        if (glowTexture) entry.glowSprite.texture = glowTexture
      }

      entry.container.visible = true

      // Position at center of tile
      entry.container.x = tileSprite.x
      entry.container.y = tileSprite.y

      // Place behind the tile
      entry.container.zIndex = (tileSprite.zIndex || 0) - 1

      // OPTIMIZED: Animate rotation using sprite rotation (no redraw!)
      const rotation = (timestamp * CONFIG.rotationSpeed) % (Math.PI * 2)
      entry.raysSprite.rotation = rotation

      // Pulse alpha based on timestamp
      const pulseProgress = (timestamp % CONFIG.pulseSpeed) / CONFIG.pulseSpeed
      const pulseValue = Math.sin(pulseProgress * Math.PI * 2) * 0.5 + 0.5
      const alpha = CONFIG.pulseMin + pulseValue * (CONFIG.pulseMax - CONFIG.pulseMin)

      entry.container.alpha = alpha

      // OPTIMIZED: Pulse glow using scale instead of redrawing
      const glowScale = 0.85 + pulseValue * 0.3
      entry.glowSprite.scale.set(glowScale)

    } else if (entry) {
      entry.container.visible = false
    }
  }

  /**
   * Clean up bursts that are no longer needed
   */
  function cleanup(usedKeys: Set<string>): void {
    for (const [key, entry] of burstCache.entries()) {
      if (!usedKeys.has(key)) {
        if (entry.container.parent) entry.container.parent.removeChild(entry.container)
        // Don't destroy textures - they're cached and shared
        entry.raysSprite.destroy()
        entry.glowSprite.destroy()
        entry.container.destroy({ children: true })
        burstCache.delete(key)
      }
    }
  }

  /**
   * Clear all bursts
   */
  function clear(): void {
    for (const [, entry] of burstCache.entries()) {
      if (entry.container.parent) entry.container.parent.removeChild(entry.container)
      // Don't destroy textures - they're cached and shared
      entry.raysSprite.destroy()
      entry.glowSprite.destroy()
      entry.container.destroy({ children: true })
    }
    burstCache.clear()
  }

  return { container, updateBurst, cleanup, clear, initTextures }
}
