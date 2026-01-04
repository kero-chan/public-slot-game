import type { Sprite, Container } from 'pixi.js'
import { Sprite as PixiSprite } from 'pixi.js'
import { gsap } from 'gsap'
import { getTextureForSymbol } from '../textures'
import { applyTileVisuals, applyAnticipationVisuals } from '../visuals'
import { isBonusTile, isTileWildcard, istileWilden } from '@/utils/tileHelpers'
import { WINNING_STATES } from '@/stores'
import { CONFIG } from '@/config/constants'
import type { ShatterAnimationManager } from '../winning/shatterAnimation'
import type { GoldToWildAnimationManager } from '../winning/goldToWildAnimation'

export interface TileRenderParams {
  col: number
  visualRow: number
  gridRow: number
  symbol: string
  xCell: number
  yCell: number
  scaledTileW: number
  scaledTileH: number
  BLEED: number
  spinning: boolean
  winning: boolean
  hasHighlights: boolean
  winningState: string | null
  winningStateTime: number
  timingStore: any
  dropAnimations: any
  bumpAnimations: any
  popAnimations: any
  anticipationEffects: any
  columnIsSpinning: boolean
  grid: string[][]
  inCascadeWindow: boolean
  shatterAnimations: ShatterAnimationManager
  goldToWildAnimations: GoldToWildAnimationManager
  goldTransformedPositions?: Set<string>  // Positions where gold tiles transform to wild (from backend)
  isLightningAnimating?: boolean  // True when lightning links animation is active
  isOtherWinningTile?: boolean    // True for winning tiles that aren't the current animating symbol
}

export function renderTile(
  params: TileRenderParams,
  spriteCache: Map<string, Sprite>,
  tilesContainer: Container
): { sprite: Sprite; hasGlowableTile: boolean; shouldShowBonusEffects: boolean; shouldShowLightBurst: boolean; shouldShowAnticipationBurst: boolean } | null {
  const {
    col,
    visualRow,
    gridRow,
    symbol,
    xCell,
    yCell,
    scaledTileW,
    scaledTileH,
    BLEED,
    spinning,
    winning,
    hasHighlights,
    winningState,
    winningStateTime,
    timingStore,
    dropAnimations,
    bumpAnimations,
    popAnimations,
    anticipationEffects,
    columnIsSpinning,
    grid,
    inCascadeWindow,
    shatterAnimations,
    goldToWildAnimations,
    goldTransformedPositions,
    isLightningAnimating = false,
    isOtherWinningTile = false
  } = params

  const cellKey = `${col}:${visualRow}`
  const tex = getTextureForSymbol(symbol)
  if (!tex) return null

  // Only show glow effects for bonus/wild tiles in visible rows (1-4)
  const isInVisibleRows = visualRow >= CONFIG.reels.visualWinStartRow && visualRow <= CONFIG.reels.visualWinEndRow
  const hasGlowableTile = (isBonusTile(symbol) || isTileWildcard(symbol)) && isInVisibleRows

  // Check if gold-to-wild animation is active for this cell
  const isGoldToWildAnimating = goldToWildAnimations.isAnimating(cellKey)

  let sp = spriteCache.get(cellKey)
  let symbolChanged = false

  // Don't reset sprite state if gold-to-wild animation is controlling it
  if (sp && sp.texture !== tex && !isGoldToWildAnimating) {
    symbolChanged = true
    if (bumpAnimations.isAnimating(cellKey) || bumpAnimations.isAlreadyBumped(cellKey)) {
      bumpAnimations.reset(cellKey)
    }
    sp.scale.x = 1
    sp.scale.y = 1
    sp.alpha = 1
  }

  const w = scaledTileW + BLEED * 2
  const h = scaledTileH + BLEED * 2

  if (!sp) {
    sp = new PixiSprite(tex)
    sp.anchor.set(0.5, 0.5)
    sp.roundPixels = false // Disable pixel rounding for smoother high-res rendering
    spriteCache.set(cellKey, sp)
  } else if (sp.texture !== tex && !isGoldToWildAnimating) {
    // Don't override texture if gold-to-wild animation is controlling it
    sp.texture = tex
  }

  sp.anchor.set(0.5, 0.5)

  const scaleX = w / sp.texture.width
  const scaleY = h / sp.texture.height

  const isBumping = bumpAnimations.isAnimating(cellKey)
  const isPopping = popAnimations.isAnimating(cellKey)
  const isCurrentlyDropping = dropAnimations.isDropping(cellKey)
  const hasActiveWinningState = winningState && winningState !== WINNING_STATES.IDLE

  // Apply scale and alpha based on state
  applyTileState(sp, {
    cellKey,
    scaleX,
    scaleY,
    winningState,
    winningStateTime,
    timingStore,
    isBumping,
    isPopping,
    isCurrentlyDropping,
    symbolChanged,
    inCascadeWindow,
    hasActiveWinningState,
    shatterAnimations,
    goldToWildAnimations,
    scaledTileW,
    scaledTileH,
    BLEED,
    isGoldToWildAnimating,
    symbol,
    winning,
    col,
    gridRow,
    grid,
    goldTransformedPositions
  })

  // Check for bonus effects (bump animation and light burst are only for bonus tiles)
  const isBonus = isBonusTile(symbol)
  const isVisibleRow = visualRow >= CONFIG.reels.visualWinStartRow && visualRow <= CONFIG.reels.visualWinEndRow
  const hasActiveDrops = dropAnimations.hasActiveDrops()
  // Bump animation only when stationary - only for bonus tiles
  const shouldShowBonusEffects = isBonus && isVisibleRow && !spinning && !isCurrentlyDropping && !hasActiveDrops
  // Light burst follows bonus tiles even during spin - only for bonus tiles
  const shouldShowLightBurst = isBonus && isVisibleRow && !isCurrentlyDropping

  // Anticipation effects
  const anticipationState = anticipationEffects.getTileVisualState(col, gridRow, symbol, columnIsSpinning, grid)
  const shouldShowAnticipationBurst = anticipationEffects.isActive() && anticipationState.highlight

  // Check if shatter animation has completed (tile should be fully hidden)
  const isShatterCompleted = shatterAnimations.hasCompleted(cellKey)

  // Use separate visual functions for anticipation vs normal winning
  // Skip visual updates only if shatter animation completed (tile is gone)
  // During animation, mask handles visibility so we still apply visuals
  if (!isShatterCompleted) {
    if (anticipationEffects.isActive()) {
      // Anticipation mode: use stronger contrast visuals
      applyAnticipationVisuals(sp, anticipationState.highlight, anticipationState.shouldDim)
    } else {
      // Normal winning: use standard visuals
      // Pass lightning state to darken non-winning tiles during lightning animation
      // Also darken other winning tiles (from different symbols) during lightning
      const shouldDarkenForLightning = isLightningAnimating && !winning
      applyTileVisuals(sp, 1, winning, hasHighlights, shouldDarkenForLightning, isOtherWinningTile)
    }
  }

  // Trigger bump for bonus tiles
  if (shouldShowBonusEffects && !bumpAnimations.isAlreadyBumped(cellKey) && !bumpAnimations.isAnimating(cellKey)) {
    bumpAnimations.startBump(cellKey, sp)
  }

  // Position - avoid Math.round() during animations for smoother movement
  sp.x = xCell - BLEED + w / 2
  if (!isCurrentlyDropping) {
    sp.y = yCell - BLEED + h / 2
  }

  // Z-index
  const shouldElevate = shouldShowBonusEffects || shouldShowAnticipationBurst
  sp.zIndex = shouldElevate ? 100 : 0

  if (!sp.parent) tilesContainer.addChild(sp)

  return { sprite: sp, hasGlowableTile, shouldShowBonusEffects, shouldShowLightBurst, shouldShowAnticipationBurst }
}

function applyTileState(
  sp: Sprite,
  params: {
    cellKey: string
    scaleX: number
    scaleY: number
    winningState: string | null
    winningStateTime: number
    timingStore: any
    isBumping: boolean
    isPopping: boolean
    isCurrentlyDropping: boolean
    symbolChanged: boolean
    inCascadeWindow: boolean
    hasActiveWinningState: boolean
    shatterAnimations: ShatterAnimationManager
    goldToWildAnimations: GoldToWildAnimationManager
    scaledTileW: number
    scaledTileH: number
    BLEED: number
    isGoldToWildAnimating: boolean
    symbol: string
    winning: boolean
    col: number
    gridRow: number
    grid: string[][]
    goldTransformedPositions?: Set<string>
  }
): void {
  const {
    cellKey,
    scaleX,
    scaleY,
    winningState,
    isBumping,
    isPopping,
    isCurrentlyDropping,
    symbolChanged,
    inCascadeWindow,
    hasActiveWinningState,
    shatterAnimations,
    goldToWildAnimations,
    isGoldToWildAnimating,
    symbol,
    winning,
    col,
    gridRow,
    grid,
    goldTransformedPositions
  } = params

  // Don't touch scale if pop or bump animation is controlling it
  if (isPopping || isBumping) {
    return
  }

  if (isCurrentlyDropping || symbolChanged || (inCascadeWindow && !hasActiveWinningState)) {
    // Kill any GSAP tweens on scale to prevent conflicts
    gsap.killTweensOf(sp.scale)
    // Always reset scale
    sp.scale.x = scaleX
    sp.scale.y = scaleY
    sp.alpha = 1
  } else if (winningState === WINNING_STATES.HIGHLIGHTED) {
    // Kill any GSAP tweens on scale to prevent conflicts
    gsap.killTweensOf(sp.scale)
    // Always reset scale
    sp.scale.x = scaleX
    sp.scale.y = scaleY
    sp.alpha = 1
  } else if (winningState === WINNING_STATES.FLIPPING) {
    // Check if this position is a gold-to-wild transformation (from backend)
    const posKey = `${col},${gridRow}`
    const isGoldToWild = goldTransformedPositions?.has(posKey) ?? false
    // Check if this is a transformed wild - should never shatter
    const isTransformed = shatterAnimations.isTransformedWild(cellKey)

    // Gold-to-wild tiles: use snap animation, not shatter
    if (isGoldToWild) {
      // Start snap animation if not already running
      if (!goldToWildAnimations.isAnimating(cellKey)) {
        goldToWildAnimations.startTransform(
          cellKey, col, gridRow, sp, sp.x, sp.y,
          // Callback to update grid when texture swaps
          () => {
            if (grid[col]) {
              grid[col][gridRow] = 'wild'
            }
          }
        )
      }
      // Keep tile visible during animation
      sp.alpha = 1
    } else if (isTransformed) {
      // Already transformed wilds don't get any animation - they stay visible
      // Do nothing - keep tile visible
    } else {
      // Normal winning tiles get the eating animation
      if (!shatterAnimations.isAnimating(cellKey) && !shatterAnimations.hasCompleted(cellKey)) {
        shatterAnimations.startShatter(
          cellKey, sp, sp.x, sp.y, scaleX, scaleY, false
        )
      }

      // Keep sprite visible during animation (mask handles the eating effect)
      // Only hide after animation completes
      if (shatterAnimations.hasCompleted(cellKey)) {
        sp.alpha = 0
        sp.scale.x = scaleX
        sp.scale.y = scaleY
      }
    }
  } else if (winningState === WINNING_STATES.FLIPPED) {
    // Check if this position is a gold-to-wild transformation
    const posKey = `${col},${gridRow}`
    const isGoldToWild = goldTransformedPositions?.has(posKey) ?? false
    // Transformed wilds stay visible
    const isTransformed = shatterAnimations.isTransformedWild(cellKey)

    // Gold-to-wild and transformed tiles stay visible, normal tiles stay hidden
    if (!isGoldToWild && !isTransformed) {
      sp.alpha = 0
      sp.scale.x = 0
    }
  } else if (winningState === WINNING_STATES.DISAPPEARING) {
    // Check if this position is a gold-to-wild transformation
    const posKey = `${col},${gridRow}`
    const isGoldToWild = goldTransformedPositions?.has(posKey) ?? false
    // Transformed wilds stay visible
    const isTransformed = shatterAnimations.isTransformedWild(cellKey)

    // Gold-to-wild and transformed tiles stay visible, normal tiles stay hidden
    if (!isGoldToWild && !isTransformed) {
      sp.alpha = 0
    }
  } else {
    // Normal state - reset
    // Kill any GSAP tweens on scale to prevent conflicts from shatter animations
    gsap.killTweensOf(sp.scale)
    // Always reset scale first to ensure tiles aren't stuck collapsed from FLIPPED state
    sp.scale.x = scaleX
    sp.scale.y = scaleY
    sp.rotation = 0
    sp.skew.y = 0
    sp.alpha = 1
  }
}
