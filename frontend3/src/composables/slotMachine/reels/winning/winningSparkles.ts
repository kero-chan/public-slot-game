// @ts-nocheck
/**
 * Manages sparkling effects for winning tiles during their highlight/flip animations
 * Particles burst from tiles then fly toward the wallet to show money being won
 */
import { Container, Sprite, Texture } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import { useTimingStore, useWinningStore, useGameStore } from '@/stores'
import type { WinCombination } from '@/features/spin/types'
import { getWinAnnouncementTexture } from '@/config/spritesheet'

// Basic reel layout
const COLS = 5
const ROWS_FULL = 4
const TOP_PARTIAL = 0.15
const BUFFER_OFFSET = 4

// Sparkle tuning for winning tiles - balanced for performance
const DOTS_PER_TILE = 6       // Reduced for performance
const DOT_SPAWN_RATE = 1.0    // Always spawn when called
const DOT_MIN_SIZE = 35       // Larger base size
const DOT_MAX_SIZE = 55       // Larger max size

// Scale multipliers for animation phases
const BURST_SCALE_START = 1.5   // Start bigger during burst
const BURST_SCALE_END = 2.0     // Grow during burst
const ARC_SCALE = 1.8           // Maintain size during arc
const FLY_SCALE_START = 2.2     // Largest at start of fly
const FLY_SCALE_END = 0.3       // Shrink as approaching wallet

// Lifetime phases (ms)
const BURST_PHASE_DURATION = 250  // Initial burst upward
const ARC_PHASE_DURATION = 400    // Arc/curve phase (peak then descend)
const FLY_PHASE_DURATION = 350    // Final fly to wallet
const TOTAL_LIFE = BURST_PHASE_DURATION + ARC_PHASE_DURATION + FLY_PHASE_DURATION

// Speeds
const BURST_SPEED_MIN = 3.0  // px/frame outward during burst
const BURST_SPEED_MAX = 6.0
const FLY_SPEED = 15  // px/frame toward wallet

// Approx frame time
const FRAME_MS_APPROX = 16.7

/**
 * Sparkle dot interface
 */
interface SparkleDot {
  sprite: Sprite
  born: number
  life: number
  vx: number
  vy: number
  startX: number  // Original spawn position
  startY: number
  peakX: number   // Position at peak of arc
  peakY: number   // Position at peak of arc
  phase: 'burst' | 'arc' | 'fly'  // Current animation phase
  baseScale: number  // Initial scale for size calculations
}

/**
 * Dot entry interface
 */
interface DotEntry {
  list: SparkleDot[]
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
 * Highlight animation state
 */
interface HighlightAnim {
  start: number
}

/**
 * Grid state interface
 */
interface GridState {
  highlightWins?: WinCombination[]
  highlightAnim?: HighlightAnim
}

/**
 * Target position for particles to fly toward
 */
export interface ParticleTarget {
  x: number
  y: number
}

/**
 * Callback when particles reach the wallet
 */
export type OnParticlesReachedWallet = () => void

/**
 * Winning sparkles interface
 */
export interface WinningSparkles {
  container: Container
  draw: (
    mainRect: MainRect,
    tileSize: number | TileSize,
    timestamp: number,
    canvasW: number,
    gridState: GridState,
    walletTarget?: ParticleTarget | null
  ) => void
  clear: () => void
  /** Check if there are particles currently flying to wallet */
  hasActiveParticles: () => boolean
  /** Set callback for when particles reach the wallet */
  setOnParticlesReachedWallet: (callback: OnParticlesReachedWallet | null) => void
}

// Cache the star texture
let STAR_TEX: Texture | null = null
function ensureStarTexture(): Texture | null {
  if (STAR_TEX && STAR_TEX.source?.valid) return STAR_TEX

  const loaded = getWinAnnouncementTexture('win_gold.webp')
  if (loaded) {
    STAR_TEX = loaded.source ? loaded : Texture.from(loaded)
    return STAR_TEX
  }

  return null
}

export function createWinningSparkles(): WinningSparkles {
  const container = new Container()
  container.zIndex = 15 // Above glow overlay (10), but below overlays (50+)

  const timingStore = useTimingStore()
  const winningStore = useWinningStore()
  const gameStore = useGameStore()
  const dotsMap = new Map<string, DotEntry>() // key -> { list: [] }

  // Current wallet target position (updated each frame)
  let currentWalletTarget: ParticleTarget | null = null

  // Track if we've already spawned particles for current cascade
  let hasSpawnedForCurrentCascade = false
  let lastCascadeState = ''

  // Callback when particles reach wallet
  let onParticlesReachedWalletCallback: OnParticlesReachedWallet | null = null
  let particlesWereActive = false  // Track if we had active particles last frame
  let reachedWalletCallbackFired = false  // Prevent multiple callbacks per spawn cycle

  function spawnDot(
    key: string,
    tileW: number,
    tileH: number,
    xCell: number,
    yCell: number,
    timestamp: number
  ): void {
    let entry = dotsMap.get(key)
    if (!entry) {
      entry = { list: [] }
      dotsMap.set(key, entry)
    }
    if (entry.list.length >= DOTS_PER_TILE) return
    if (Math.random() > DOT_SPAWN_RATE) return

    const tex = ensureStarTexture() || Texture.WHITE
    const s = new Sprite(tex)
    s.blendMode = BLEND_MODES.ADD
    s.anchor.set(0.5)

    // Use scale instead of width/height for better PixiJS performance
    const pxSize = DOT_MIN_SIZE + Math.random() * (DOT_MAX_SIZE - DOT_MIN_SIZE)
    const baseScale = pxSize / (tex.width || 32)  // Calculate scale based on texture size
    s.scale.set(baseScale)

    // Start near center of tile with slight randomness
    const centerX = xCell + tileW * 0.5
    const centerY = yCell + tileH * 0.5
    const startX = centerX + (Math.random() * 2 - 1) * tileW * 0.2
    const startY = centerY + (Math.random() * 2 - 1) * tileH * 0.2
    s.x = startX
    s.y = startY
    s.alpha = 1.0

    // Initial burst motion: primarily UPWARD with some horizontal spread
    // Angle between -60 and -120 degrees (upward cone)
    const angle = -Math.PI / 2 + (Math.random() - 0.5) * Math.PI * 0.6
    const speed = BURST_SPEED_MIN + Math.random() * (BURST_SPEED_MAX - BURST_SPEED_MIN)
    const vx = Math.cos(angle) * speed
    const vy = Math.sin(angle) * speed  // Negative = upward

    container.addChild(s)
    entry.list.push({
      sprite: s,
      born: timestamp,
      life: TOTAL_LIFE,
      vx,
      vy,
      startX,
      startY,
      peakX: 0,   // Will be set when transitioning to arc phase
      peakY: 0,
      phase: 'burst',
      baseScale
    })
  }

  function updateDots(entry: DotEntry, timestamp: number): void {
    if (!entry) return
    const alive: SparkleDot[] = []

    for (const p of entry.list) {
      const age = timestamp - p.born

      if (age < BURST_PHASE_DURATION) {
        // Phase 1: Burst UPWARD
        p.phase = 'burst'
        p.sprite.x += p.vx
        p.sprite.y += p.vy

        // Slight deceleration during burst (gravity-like effect)
        p.vx *= 0.96
        p.vy *= 0.92  // Slower vertical deceleration to rise higher

        // Full opacity during burst
        p.sprite.alpha = 1.0

        // Scale up during burst phase - grow as they rise
        const burstProgress = age / BURST_PHASE_DURATION
        const burstScale = BURST_SCALE_START + (BURST_SCALE_END - BURST_SCALE_START) * burstProgress
        p.sprite.scale.set(p.baseScale * burstScale)

      } else if (age < BURST_PHASE_DURATION + ARC_PHASE_DURATION) {
        // Phase 2: Arc - curve toward wallet target
        if (p.phase === 'burst') {
          // Just transitioned - save peak position
          p.phase = 'arc'
          p.peakX = p.sprite.x
          p.peakY = p.sprite.y
        }

        const arcProgress = (age - BURST_PHASE_DURATION) / ARC_PHASE_DURATION
        const t = Math.min(arcProgress, 1)

        if (currentWalletTarget) {
          // Bezier curve from peak to wallet
          // Use easeInQuad for smooth acceleration toward wallet
          const easeT = t * t

          // Interpolate from peak position toward wallet
          const targetX = currentWalletTarget.x
          const targetY = currentWalletTarget.y

          p.sprite.x = p.peakX + (targetX - p.peakX) * easeT * 0.6  // Only go 60% toward target in arc phase
          p.sprite.y = p.peakY + (targetY - p.peakY) * easeT * 0.6

          // Add slight wobble for organic motion
          const wobble = Math.sin(age * 0.015) * 3 * (1 - t)
          p.sprite.x += wobble
        } else {
          // No wallet target - gentle float downward
          p.sprite.y += 1.5
          p.sprite.x += Math.sin(age * 0.01) * 0.5
        }

        // Maintain full opacity during arc
        p.sprite.alpha = 1.0

        // Maintain large scale during arc, slight transition to fly scale
        const arcScale = ARC_SCALE + (FLY_SCALE_START - ARC_SCALE) * t
        p.sprite.scale.set(p.baseScale * arcScale)

      } else if (currentWalletTarget) {
        // Phase 3: Final fly toward wallet
        p.phase = 'fly'

        const flyProgress = (age - BURST_PHASE_DURATION - ARC_PHASE_DURATION) / FLY_PHASE_DURATION
        const t = Math.min(flyProgress, 1)

        // Calculate direction to wallet
        const dx = currentWalletTarget.x - p.sprite.x
        const dy = currentWalletTarget.y - p.sprite.y
        const dist = Math.sqrt(dx * dx + dy * dy)

        if (dist > 5) {
          // Move toward wallet with acceleration
          const speed = FLY_SPEED * (1 + t * 3)  // Accelerate toward wallet
          const moveX = (dx / dist) * speed
          const moveY = (dy / dist) * speed
          p.sprite.x += moveX
          p.sprite.y += moveY
        }

        // Fade out as approaching wallet
        p.sprite.alpha = Math.max(0, 1 - t * 0.9)

        // Scale down from large to small as approaching wallet
        const flyScale = FLY_SCALE_START + (FLY_SCALE_END - FLY_SCALE_START) * t
        p.sprite.scale.set(p.baseScale * flyScale)

        // Check if reached wallet or time expired
        if (dist < 15 || t >= 1) {
          p.sprite.parent?.removeChild(p.sprite)
          p.sprite.destroy({ children: true, texture: false, baseTexture: false })
          continue
        }
      } else {
        // No wallet target - just fade out
        const t = age / TOTAL_LIFE
        p.sprite.alpha = (1 - t) * (1 - t)
        // Scale down gradually when no wallet target
        const fallbackScale = BURST_SCALE_END * (1 - t * 0.7)
        p.sprite.scale.set(p.baseScale * fallbackScale)

        if (t >= 1) {
          p.sprite.parent?.removeChild(p.sprite)
          p.sprite.destroy({ children: true, texture: false, baseTexture: false })
          continue
        }
      }

      alive.push(p)
    }
    entry.list = alive
  }

  function draw(
    mainRect: MainRect,
    tileSize: number | TileSize,
    timestamp: number,
    canvasW: number,
    gridState: GridState,
    walletTarget?: ParticleTarget | null
  ): void {
    // Update wallet target for particle fly-to animation
    currentWalletTarget = walletTarget || null

    const tileW = typeof tileSize === 'number' ? tileSize : tileSize.w
    const tileH = typeof tileSize === 'number' ? tileSize : tileSize.h

    // Match reels positioning - larger margin on smartphone
    const isSmartphone = canvasW < 600
    const margin = isSmartphone ? 22 : 10
    const availableWidth = canvasW - (margin * 2)
    const scaledTileW = availableWidth / COLS
    const scaledTileH = scaledTileW * (tileH / tileW)
    const stepX = scaledTileW
    // Shift tiles on smartphone to align with frame's visible area
    const tileOffset = isSmartphone ? -3 : 0
    const originX = mainRect.x + margin + tileOffset

    // With TOP_PARTIAL = 0.15, adjust startY
    const startY = mainRect.y - (1 - TOP_PARTIAL) * scaledTileH

    const used = new Set<string>()

    // Get winning positions from gameStore.currentWins (persists through cascade)
    // gridState.highlightWins is cleared before cascade starts
    const winningPositions = new Set<string>()
    const wins = gameStore.currentWins || []
    wins.forEach((win: any) => {
      win.positions.forEach((pos: any) => {
        // Backend sends positions as objects: {reel, row}
        winningPositions.add(`${pos.reel},${pos.row}`)
      })
    })

    // Use Date.now() for consistency with other timing
    const now = Date.now()

    // Get current game flow state
    const gameFlowState = gameStore.gameFlowState

    // Reset spawn flag when entering a new cascade cycle
    if (gameFlowState !== lastCascadeState) {
      if (gameFlowState === 'cascading' || gameFlowState === 'disappearing_tiles') {
        hasSpawnedForCurrentCascade = false
      }
      lastCascadeState = gameFlowState
    }

    // Spawn sparkles at the start of CASCADING or DISAPPEARING_TILES state
    // Only spawn once per cascade cycle
    const isCascadeState = gameFlowState === 'cascading' || gameFlowState === 'disappearing_tiles'
    const shouldSpawn = isCascadeState && !hasSpawnedForCurrentCascade

    // Spawn sparkles during cascade
    if (winningPositions.size > 0 && shouldSpawn) {
      hasSpawnedForCurrentCascade = true
      reachedWalletCallbackFired = false  // Reset callback flag for new spawn cycle
      for (let col = 0; col < COLS; col++) {
        for (let r = 0; r < ROWS_FULL; r++) {  // Visible rows 0-3
          const gridRow = r + BUFFER_OFFSET
          const posKey = `${col},${gridRow}`

          if (!winningPositions.has(posKey)) continue

          const xCell = originX + col * stepX
          const yCell = startY + (r + 1) * scaledTileH

          const key = `${col}:${r}`

          // Spawn sparkles for this tile (reduced count for performance)
          for (let i = 0; i < 3; i++) {
            spawnDot(key, scaledTileW, scaledTileH, xCell, yCell, now)
          }
          used.add(key)
        }
      }
    }

    // Always update all existing sparkles (even after flip ends)
    for (const [key, entry] of dotsMap.entries()) {
      updateDots(entry, now)  // Use Date.now() for consistency
      // If entry still has particles, keep it alive
      if (entry.list.length > 0) {
        used.add(key)
      }
    }

    // Cleanup only entries with no particles left
    for (const [key, entry] of dotsMap.entries()) {
      if (!used.has(key)) {
        dotsMap.delete(key)
      }
    }

    // Check if particles just finished reaching wallet (transition from active to inactive)
    const currentlyActive = hasActiveParticles()
    if (particlesWereActive && !currentlyActive && !reachedWalletCallbackFired && currentWalletTarget) {
      // Particles just finished flying to wallet - fire callback
      reachedWalletCallbackFired = true
      if (onParticlesReachedWalletCallback) {
        onParticlesReachedWalletCallback()
      }
    }
    particlesWereActive = currentlyActive
  }

  function clear(): void {
    // Clear all sparkles
    for (const [key, entry] of dotsMap.entries()) {
      for (const p of entry.list) {
        p.sprite.parent?.removeChild(p.sprite)
        p.sprite.destroy({ children: true, texture: false, baseTexture: false })
      }
    }
    dotsMap.clear()
  }

  /**
   * Check if there are particles currently in flight
   */
  function hasActiveParticles(): boolean {
    for (const [, entry] of dotsMap.entries()) {
      if (entry.list.length > 0) return true
    }
    return false
  }

  /**
   * Set callback for when particles reach the wallet
   */
  function setOnParticlesReachedWallet(callback: OnParticlesReachedWallet | null): void {
    onParticlesReachedWalletCallback = callback
  }

  return { container, draw, clear, hasActiveParticles, setOnParticlesReachedWallet }
}
