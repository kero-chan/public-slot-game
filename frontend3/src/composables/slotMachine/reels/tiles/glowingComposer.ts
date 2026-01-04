// @ts-nocheck
import { Container, Sprite, Texture, Graphics, RenderTexture, Application } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import type { Ref } from 'vue'

// Basic reel layout
const COLS = 5
const ROWS_FULL = 4
const TOP_PARTIAL = 0.15

// Detect mobile for performance optimization
const IS_MOBILE = typeof window !== 'undefined' && window.innerWidth < 600

// Sparkle tuning: bottom → top star dots
// Reduced on mobile for better performance
const DOTS_PER_TILE = IS_MOBILE ? 15 : 40
const DOT_SPAWN_RATE = IS_MOBILE ? 0.30 : 0.60
const DOT_MIN_SIZE = IS_MOBILE ? 8 : 10
const DOT_MAX_SIZE = IS_MOBILE ? 20 : 26

// Lifetime (ms)
const DOT_MIN_LIFE = 800
const DOT_MAX_LIFE = IS_MOBILE ? 2500 : 4000

// Speeds
const DOT_UP_SPEED_MIN = 0.6  // px/frame upward
const DOT_UP_SPEED_MAX = 1.3
const DOT_DRIFT_MAX = 0.4     // px/frame horizontal drift

// How high dots travel relative to tile height
const DOT_TRAVEL_FACTOR = 1.8

// Approx frame time for converting frames → ms
const FRAME_MS_APPROX = 16.7

// Star configuration
const STAR_CONFIG = {
  points: 4,                // Number of star points
  outerRadius: 16,          // Outer radius of star
  innerRadius: 6,           // Inner radius (valley between points)
  glowRadius: 24,           // Glow radius around star

  // Colors
  coreColor: 0xffffff,      // Bright white core
  innerColor: 0xffffd0,     // Pale yellow
  outerColor: 0xffd700,     // Golden yellow
  glowColor: 0xffaa00,      // Orange glow

  // Alpha values
  coreAlpha: 1.0,
  innerAlpha: 0.9,
  outerAlpha: 0.6,
  glowAlpha: 0.3,
}

/**
 * Sparkle dot interface
 */
interface SparkleDot {
  sprite: Sprite
  born: number
  life: number
  vx: number
  vy: number
  // Track relative offset from tile center for following during spin
  offsetX: number
  offsetY: number
  // Rotation speed for twinkle effect
  rotationSpeed: number
}

/**
 * Dot entry interface - now tracks the tile sprite for following
 */
interface DotEntry {
  list: SparkleDot[]
  lastTileX: number
  lastTileY: number
  tileW: number
  tileH: number
}

/**
 * Main rect interface
 */
interface MainRect {
  x: number
  y: number
}

/**
 * Tile size interface
 */
interface TileSize {
  w: number
  h: number
}

/**
 * Grid state interface
 */
interface GridState {
  spinOffsets?: number[]
  reelStrips?: string[][]
  reelTopIndex?: number[]
  grid?: string[][]
}

/**
 * Game state interface
 */
interface GameState {
  isSpinning?: Ref<boolean>
}

/**
 * Tile info for sparkle updates
 */
export interface SparkTileInfo {
  key: string
  x: number  // tile center x
  y: number  // tile center y
  width: number
  height: number
}

/**
 * Glow overlay interface
 */
export interface GlowOverlay {
  container: Container
  draw: (mainRect: MainRect, tileSize: number | TileSize, timestamp: number, canvasW: number) => void
  updateTileSparkles: (tiles: SparkTileInfo[], timestamp: number) => void
  cleanup: (usedKeys: Set<string>) => void
  initTextures: (app: Application) => void
}

/**
 * Draw a star shape with the given parameters
 */
function drawStar(
  graphics: Graphics,
  cx: number,
  cy: number,
  points: number,
  outerR: number,
  innerR: number,
  color: number,
  alpha: number
): void {
  const angleStep = Math.PI / points
  const path: number[] = []

  for (let i = 0; i < points * 2; i++) {
    const angle = i * angleStep - Math.PI / 2
    const radius = i % 2 === 0 ? outerR : innerR
    path.push(cx + Math.cos(angle) * radius)
    path.push(cy + Math.sin(angle) * radius)
  }

  graphics.poly(path)
  graphics.fill({ color, alpha })
}

/**
 * Create a procedural star texture using Graphics
 */
function createStarTexture(): Graphics {
  const g = new Graphics()
  const cx = STAR_CONFIG.glowRadius
  const cy = STAR_CONFIG.glowRadius

  // Outer glow (soft radial gradient approximation)
  for (let i = 5; i >= 1; i--) {
    const radius = STAR_CONFIG.glowRadius * (i / 5)
    const alpha = STAR_CONFIG.glowAlpha * (1 - (i - 1) / 5) * 0.5
    g.circle(cx, cy, radius)
    g.fill({ color: STAR_CONFIG.glowColor, alpha })
  }

  // Main star - outer layer (golden)
  drawStar(g, cx, cy, STAR_CONFIG.points, STAR_CONFIG.outerRadius, STAR_CONFIG.innerRadius, STAR_CONFIG.outerColor, STAR_CONFIG.outerAlpha)

  // Middle layer (pale yellow) - slightly smaller
  drawStar(g, cx, cy, STAR_CONFIG.points, STAR_CONFIG.outerRadius * 0.7, STAR_CONFIG.innerRadius * 0.7, STAR_CONFIG.innerColor, STAR_CONFIG.innerAlpha)

  // Core (white) - small bright center
  drawStar(g, cx, cy, STAR_CONFIG.points, STAR_CONFIG.outerRadius * 0.4, STAR_CONFIG.innerRadius * 0.5, STAR_CONFIG.coreColor, STAR_CONFIG.coreAlpha)

  // Center dot for extra brightness
  g.circle(cx, cy, 2)
  g.fill({ color: STAR_CONFIG.coreColor, alpha: 1 })

  return g
}

// Cache the star texture - generated once and reused for all sprites
let cachedStarTexture: Texture | null = null
let starTextureInitializing = false

/**
 * Initialize the star texture from a renderer
 * This should be called once when the app is ready
 */
function initStarTexture(app: Application): void {
  if (cachedStarTexture || starTextureInitializing) return
  starTextureInitializing = true

  const g = createStarTexture()
  const textureSize = STAR_CONFIG.glowRadius * 2

  // Generate a texture from the graphics - this is done once
  cachedStarTexture = app.renderer.generateTexture({
    target: g,
    resolution: 2, // Higher resolution for crisp scaling
  })

  // Cleanup the graphics after generating texture
  g.destroy()
  starTextureInitializing = false
}

/**
 * Create a new star sprite using the cached texture
 * Much faster than recreating Graphics every time
 */
function createStarSprite(): Sprite {
  // If texture isn't ready yet, create a fallback graphics-based sprite
  // This should rarely happen after initialization
  if (!cachedStarTexture) {
    const g = createStarTexture()
    const cx = STAR_CONFIG.glowRadius
    const cy = STAR_CONFIG.glowRadius
    g.pivot.set(cx, cy)
    const wrapper = g as any
    wrapper.blendMode = BLEND_MODES.ADD
    return wrapper
  }

  // Create sprite from cached texture - very fast
  const sprite = new Sprite(cachedStarTexture)
  sprite.anchor.set(0.5, 0.5)
  sprite.blendMode = BLEND_MODES.ADD

  return sprite
}

export function useGlowOverlay(gameState: GameState, gridState: GridState, options: Record<string, unknown> = {}): GlowOverlay {
  const container = new Container()
  container.zIndex = 10 // Keep behind overlays (symbolWinAnimation=50, winAnimationManager, bonusOverlay, etc.)

  // Add a graphics mask to clip to the board area
  let maskGraphics: Graphics | null = null
  let lastMaskRect = { x: 0, y: 0, w: 0, h: 0 }

  const dotsMap = new Map<string, DotEntry>()

  /**
   * Spawn a new sparkle dot for a tile
   */
  function spawnDot(
    entry: DotEntry,
    timestamp: number
  ): void {
    if (entry.list.length >= DOTS_PER_TILE) return
    if (Math.random() > DOT_SPAWN_RATE) return

    const star = createStarSprite()

    // Random size scaling
    const sizeMin = Math.min(DOT_MIN_SIZE, DOT_MAX_SIZE)
    const sizeMax = Math.max(DOT_MIN_SIZE, DOT_MAX_SIZE)
    const pxSize = sizeMin + Math.random() * (sizeMax - sizeMin)
    const scale = pxSize / (STAR_CONFIG.glowRadius * 2)
    star.scale.set(scale)

    // Spawn relative to tile center - in bottom portion of tile
    const offsetX = entry.tileW * (-0.2 + Math.random() * 0.4)
    const offsetY = entry.tileH * (0.32 + Math.random() * 0.12)  // Bottom band

    // Set initial position
    star.x = entry.lastTileX + offsetX
    star.y = entry.lastTileY + offsetY
    star.alpha = 1.0

    // Random initial rotation
    star.rotation = Math.random() * Math.PI * 2

    // Random rotation speed for twinkle effect
    const rotationSpeed = (Math.random() - 0.5) * 0.1

    // Motion - upward with drift
    const vy = -(DOT_UP_SPEED_MIN + Math.random() * (DOT_UP_SPEED_MAX - DOT_UP_SPEED_MIN))
    const vx = (Math.random() * 2 - 1) * DOT_DRIFT_MAX

    // Life based on travel distance
    const desiredRisePx = entry.tileH * DOT_TRAVEL_FACTOR
    const framesNeeded = desiredRisePx / Math.max(0.001, Math.abs(vy))
    let life = framesNeeded * FRAME_MS_APPROX
    life = Math.max(DOT_MIN_LIFE, Math.min(DOT_MAX_LIFE, life)) * (0.9 + Math.random() * 0.2)

    container.addChild(star)
    entry.list.push({
      sprite: star as any,
      born: timestamp,
      life,
      vx,
      vy,
      offsetX,
      offsetY,
      rotationSpeed
    })
  }

  /**
   * Update dots - move them and handle following the tile
   */
  function updateDots(entry: DotEntry, timestamp: number, tileX: number, tileY: number): void {
    if (!entry) return

    // Calculate how much the tile moved since last frame
    const tileDeltaX = tileX - entry.lastTileX
    const tileDeltaY = tileY - entry.lastTileY

    const alive: SparkleDot[] = []
    for (const p of entry.list) {
      const age = timestamp - p.born
      const t = Math.min(Math.max(age / p.life, 0), 1)

      // Move with tile (follow) + own velocity
      p.sprite.x += tileDeltaX + p.vx
      p.sprite.y += tileDeltaY + p.vy

      // Update offset tracking (for the velocity component)
      p.offsetY += p.vy

      // Rotate for twinkle effect
      p.sprite.rotation += p.rotationSpeed

      // Ease-out fade and slight shrink
      p.sprite.alpha = (1 - t) * (1 - t)
      p.sprite.scale.x *= 0.998
      p.sprite.scale.y *= 0.998

      if (t < 1) {
        alive.push(p)
      } else {
        p.sprite.parent?.removeChild(p.sprite)
        p.sprite.destroy({ children: true })
      }
    }
    entry.list = alive

    // Update last position
    entry.lastTileX = tileX
    entry.lastTileY = tileY
  }

  /**
   * Update sparkles for specific tiles - called from main render loop
   * This allows sparkles to follow tiles during spin
   */
  function updateTileSparkles(tiles: SparkTileInfo[], timestamp: number): void {
    const usedKeys = new Set<string>()

    for (const tile of tiles) {
      usedKeys.add(tile.key)

      let entry = dotsMap.get(tile.key)
      if (!entry) {
        entry = {
          list: [],
          lastTileX: tile.x,
          lastTileY: tile.y,
          tileW: tile.width,
          tileH: tile.height
        }
        dotsMap.set(tile.key, entry)
      }

      // Update tile dimensions if changed
      entry.tileW = tile.width
      entry.tileH = tile.height

      // Spawn new dots
      spawnDot(entry, timestamp)

      // Update existing dots (they follow the tile)
      updateDots(entry, timestamp, tile.x, tile.y)
    }

    // Cleanup dots for tiles no longer visible
    for (const [key, entry] of dotsMap.entries()) {
      if (!usedKeys.has(key)) {
        for (const p of entry.list) {
          p.sprite.parent?.removeChild(p.sprite)
          p.sprite.destroy({ children: true })
        }
        dotsMap.delete(key)
      }
    }
  }

  /**
   * Cleanup specific keys
   */
  function cleanup(usedKeys: Set<string>): void {
    for (const [key, entry] of dotsMap.entries()) {
      if (!usedKeys.has(key)) {
        for (const p of entry.list) {
          p.sprite.parent?.removeChild(p.sprite)
          p.sprite.destroy({ children: true })
        }
        dotsMap.delete(key)
      }
    }
  }

  /**
   * Draw function - now mainly handles mask setup
   * Actual sparkle updates happen via updateTileSparkles
   */
  function draw(mainRect: MainRect, tileSize: number | TileSize, timestamp: number, gameW: number): void {
    const tileW = typeof tileSize === 'number' ? tileSize : tileSize.w
    const tileH = typeof tileSize === 'number' ? tileSize : tileSize.h

    // Update mask so effects never appear in header/footer
    const TILE_SPACING = 5
    // Larger margin on smartphone to keep tiles within frame's visible area
    const isSmartphone = gameW < 600
    const margin = isSmartphone ? 22 : 10
    const totalSpacingX = TILE_SPACING * (COLS - 1)
    const availableWidth = gameW - (margin * 2) - totalSpacingX
    const scaledTileW = availableWidth / COLS
    const scaledTileH = scaledTileW * (tileH / tileW)

    const boardW = COLS * scaledTileW + totalSpacingX
    const boardH = ROWS_FULL * scaledTileH + TILE_SPACING * (ROWS_FULL - 1) + TOP_PARTIAL * scaledTileH

    // Only update mask if dimensions changed
    if (!maskGraphics ||
        lastMaskRect.x !== mainRect.x ||
        lastMaskRect.y !== mainRect.y ||
        lastMaskRect.w !== boardW ||
        lastMaskRect.h !== boardH) {

      if (!maskGraphics) {
        maskGraphics = new Graphics()
        container.mask = maskGraphics
      }
      maskGraphics.clear()
      maskGraphics.rect(mainRect.x, mainRect.y, boardW, boardH)
      maskGraphics.fill(0xffffff)

      lastMaskRect = { x: mainRect.x, y: mainRect.y, w: boardW, h: boardH }
    }
  }

  /**
   * Initialize star texture for efficient sprite creation
   * Should be called once when the PixiJS app is ready
   */
  function initTextures(app: Application): void {
    initStarTexture(app)
  }

  return { container, draw, updateTileSparkles, cleanup, initTextures }
}
