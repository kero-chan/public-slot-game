// @ts-nocheck
import { Container, Sprite } from 'pixi.js'
import { createBackdrop, type FrameTheme } from './backdrop'
import { getTextureForSymbol } from './textures'
import { createWinningEffects } from '@/composables/slotMachine/reels/winning/effects'
import { createWinningFrameManager } from '@/composables/slotMachine/reels/winning/winningComposer'
import { createImageLightningLinksManager, type TilePosition } from '@/composables/slotMachine/reels/winning/imageLightningLinks'
import { createDropAnimationManager } from './dropAnimation'
import { createBumpAnimationManager } from '@/composables/slotMachine/reels/tiles/bumpAnimation'
import { createPopAnimationManager } from '@/composables/slotMachine/reels/tiles/popAnimation'
import { createLightBurstManager } from '@/composables/slotMachine/reels/tiles/lightBurstEffect'
import { createAnticipationEffects } from './anticipationEffects'
import { createColumnHighlightManager } from './columnHighlight'
import { createGoldToWildAnimationManager } from './winning/goldToWildAnimation'
import { createShatterAnimationManager } from './winning/shatterAnimation'
import { createFreeSpinHeaderDisplay } from './freeSpinHeader'
import { useWinningStore, useTimingStore } from '@/stores'
import { CONFIG } from '@/config/constants'
import { useAudioEffects } from '@/composables/useAudioEffects'
import { numberToSymbol } from '@/utils/symbolConverter'
import type { UseGameState } from '@/composables/slotMachine/useGameState'
import type { GridState } from '@/types/global'

import { COLS, ROWS_FULL, VISIBLE_ROWS, TOP_PARTIAL, BLEED, TILE_SPACING, BUFFER_OFFSET } from './modules/types'
import type { ReelsRect, TileSize, ReelSpinData } from './modules/types'
import { createGSAPSpinManager, type GSAPSpinManager } from './modules/gsapSpinManager'
import { createReelScroller } from './modules/reelScroller'
import { detectCascadeAndStartDrops } from './modules/cascadeDetector'
import { renderTile } from './modules/tileRenderer'

export type { ReelsRect, TileSize, FrameTheme }

export interface UseReels {
  container: Container
  tilesContainer: Container
  draw: (mainRect: ReelsRect, tileSize: TileSize | number, timestamp: number, canvasW: number) => void
  winningEffects: ReturnType<typeof createWinningEffects>
  goldToWildAnimations: ReturnType<typeof createGoldToWildAnimationManager>
  triggerPop: (col: number, visualRow: number) => void
  getSpriteCache: () => Map<string, Sprite>
  startGSAPReelScroll: (
    col: number,
    strip: string[],
    xPos: number,
    startY: number,
    stepY: number,
    spriteWidth: number,
    spriteHeight: number,
    targetDistance: number,
    duration: number,
    delay: number,
    onComplete?: (col: number) => void
  ) => gsap.core.Timeline
  stopGSAPReelScroll: (col: number) => void
  stopAllGSAPReelScrolls: () => void
  gsapSpinManager: GSAPSpinManager
  reelContainers: Container[]
  isColumnSpinning: (col: number) => boolean
  lightningLinks: ReturnType<typeof createImageLightningLinksManager>
  /** Set the frame theme (for high-value symbol wins) */
  setFrameTheme: (theme: FrameTheme) => void
}

export function useReels(gameState: UseGameState, gridState: GridState): UseReels {
  const container = new Container()
  const tilesContainer = new Container()
  tilesContainer.sortableChildren = true
  const framesContainer = new Container()
  framesContainer.sortableChildren = true

  const winningStore = useWinningStore()
  const timingStore = useTimingStore()
  const { ensureBackdrop, frameContainer, notificationContainer, setFrameTheme } = createBackdrop(tilesContainer)
  const winningEffects = createWinningEffects()
  const { playWinningHighlight } = useAudioEffects()
  const winningFrames = createWinningFrameManager(() => playWinningHighlight())
  const lightningLinks = createImageLightningLinksManager()
  const dropAnimations = createDropAnimationManager()
  const bumpAnimations = createBumpAnimationManager()
  const popAnimations = createPopAnimationManager()
  const lightBursts = createLightBurstManager()
  const anticipationEffects = createAnticipationEffects()
  const columnHighlights = createColumnHighlightManager()
  const goldToWildAnimations = createGoldToWildAnimationManager()
  const shatterAnimations = createShatterAnimationManager()
  const freeSpinHeader = createFreeSpinHeaderDisplay()

  let previousSpinning = false
  let lastCascadeTime = 0
  let columnHighlightsInitialized = false

  // Cached winning positions set - recomputed only when highlightWins changes
  let cachedWinningPositionsSet: Set<string> = new Set()
  let lastHighlightWinsRef: typeof gridState.highlightWins = null

  // Double-buffer pattern for used keys - swap between frames to avoid allocation
  // One set is "current frame", the other is "previous frame"
  let usedKeysA: Set<string> = new Set()
  let usedKeysB: Set<string> = new Set()
  let currentUsedKeys = usedKeysA
  let previousUsedKeys = usedKeysB

  // Reusable array for glowable tile info - avoids allocation every frame
  const glowableTileInfosArray: Array<{ key: string; x: number; y: number; width: number; height: number }> = []

  const reelContainers: Container[] = []
  const reelSpinData = new Map<number, ReelSpinData>()

  for (let col = 0; col < COLS; col++) {
    const reelContainer = new Container()
    reelContainer.sortableChildren = true
    tilesContainer.addChild(reelContainer)
    reelContainers.push(reelContainer)
  }

  const gsapSpinManager = createGSAPSpinManager()
  const reelScroller = createReelScroller(reelContainers, reelSpinData, gridState)

  container.addChild(tilesContainer)
  tilesContainer.addChild(lightBursts.container)
  lightBursts.container.sortableChildren = true
  tilesContainer.addChild(shatterAnimations.container)
  shatterAnimations.container.zIndex = 500 // Above tiles but below overlays
  container.addChild(columnHighlights.container)
  container.addChild(frameContainer) // Grid frame rendered ON TOP of tiles
  container.addChild(notificationContainer) // Notification text in top bar of frame
  container.addChild(freeSpinHeader.container) // Free spin counter in header
  container.addChild(framesContainer)

  const spriteCache = new Map<string, Sprite>()

  function draw(mainRect: ReelsRect, tileSize: TileSize | number, timestamp: number, gameW: number): void {
    ensureBackdrop(mainRect, gameW)

    const tileW = typeof tileSize === 'number' ? tileSize : tileSize.w
    const tileH = typeof tileSize === 'number' ? tileSize : tileSize.h

    // Use larger margin on smartphone to keep tiles within frame's visible area
    const isSmartphone = gameW < 600
    const margin = isSmartphone ? 22 : 10
    const totalSpacingX = TILE_SPACING * (COLS - 1)
    const availableWidth = gameW - (margin * 2) - totalSpacingX
    const scaledTileW = availableWidth / COLS
    const scaledTileH = scaledTileW * (tileH / tileW)
    const stepX = scaledTileW + TILE_SPACING
    const stepY = scaledTileH + TILE_SPACING
    // originX includes mainRect.x offset for centering game content
    // Shift tiles on smartphone to align with frame's visible area
    const tileOffset = isSmartphone ? -3 : 0
    const originX = mainRect.x + margin + tileOffset
    // Adjust grid position based on screen height percentage
    const screenHeight = window.innerHeight
    const yOffsetPercent = 0.08  // 8% of screen height
    const yOffset = Math.round(screenHeight * yOffsetPercent)
    const startY = Math.round(mainRect.y - (1 - TOP_PARTIAL) * scaledTileH) + yOffset
    const spinning = !!gameState.isSpinning?.value

    // Update animations
    // Note: bumpAnimations, popAnimations, goldToWildAnimations use GSAP which handles updates automatically
    // dropAnimations.update() does cleanup of completed states
    // shatterAnimations.update() moves broken tile pieces each frame (physics-based)
    dropAnimations.update()
    shatterAnimations.update()
    gridState.isDropAnimating = dropAnimations.hasActiveDrops()

    // Update free spin header display (pinned in middle of grid during free spin mode)
    const freeSpins = gameState.freeSpins?.value ?? 0
    const inFreeSpinMode = gameState.inFreeSpinMode?.value ?? false
    freeSpinHeader.update(freeSpins, inFreeSpinMode, mainRect.x, mainRect.y, mainRect.h, gameW)

    const hasHighlights = gridState.highlightWins?.length > 0

    // Initialize column highlights
    if (!columnHighlightsInitialized) {
      const rowsToCover = VISIBLE_ROWS + 1 // 5 rows to cover all visible tiles
      const extraPadding = scaledTileH * 0.2 // 20% extra padding for last row
      const columnHeight = Math.min(
        mainRect.h,
        rowsToCover * stepY + extraPadding
      )
      columnHighlights.initialize(COLS, scaledTileW, columnHeight, originX, mainRect.y + scaledTileH * 0.5, stepX)
      columnHighlightsInitialized = true
    }

    columnHighlights.update(gridState.activeSlowdownColumn ?? -1)

    // Clear animations when spinning starts
    if (spinning && !previousSpinning) {
      dropAnimations.clear()
      bumpAnimations.clear()
      popAnimations.clear()
      // Don't clear lightBursts - let them follow tiles during spin
      shatterAnimations.clear()
      columnHighlights.hideAll()
      lastCascadeTime = 0
      gridState.previousGridSnapshot = null
    }

    // Detect cascade
    const cascadeTime = gridState.lastCascadeTime || 0
    if (cascadeTime > lastCascadeTime) {
      detectCascadeAndStartDrops({
        gridState,
        dropAnimations,
        spriteCache,
        COLS,
        ROWS_FULL,
        BUFFER_OFFSET,
        startY,
        stepY,
        originX,
        stepX,
        scaledTileH,
        BLEED,
        getTextureForSymbol
      })
      lastCascadeTime = cascadeTime
    }

    const now = timestamp || performance.now()
    // Note: lastCascadeTime uses Date.now() time base (set in useCascadeAnimation.ts)
    // so we must compare with Date.now() here for consistency
    const timeSinceLastCascade = lastCascadeTime > 0 ? (Date.now() - lastCascadeTime) : Infinity
    const inCascadeWindow = timeSinceLastCascade < timingStore.CASCADE_RESET_WINDOW

    previousSpinning = spinning

    // Double-buffer swap: previous frame's "current" becomes this frame's "previous"
    // Then clear the new "current" for reuse
    const temp = previousUsedKeys
    previousUsedKeys = currentUsedKeys
    currentUsedKeys = temp
    currentUsedKeys.clear()

    // Only rebuild winning positions set when highlightWins reference changes
    // This avoids Set construction overhead on every frame
    if (gridState.highlightWins !== lastHighlightWinsRef) {
      lastHighlightWinsRef = gridState.highlightWins
      cachedWinningPositionsSet = new Set<string>()

      if (gridState.highlightWins && gridState.highlightWins.length > 0) {
        for (const win of gridState.highlightWins) {
          for (const pos of win.positions) {
            cachedWinningPositionsSet.add(`${pos.reel},${pos.row}`)
          }
        }
      }
    }
    const winningPositionsSet = cachedWinningPositionsSet

    // Check lightning animation state once per frame, not per tile
    const isLightningAnimating = lightningLinks.isAnimating()
    const currentAnimatingSymbol = lightningLinks.getCurrentAnimatingSymbol()

    // Get stored positions for the currently animating symbol
    // These persist even when highlightWins is cleared, keeping highlight in sync with link
    let currentAnimatingSymbolPositions: Set<string> | null = null
    if (isLightningAnimating && currentAnimatingSymbol !== null) {
      const storedPositions = lightningLinks.getStoredPositionsForSymbol(currentAnimatingSymbol)
      if (storedPositions && storedPositions.length > 0) {
        currentAnimatingSymbolPositions = new Set<string>()
        for (const pos of storedPositions) {
          // storedPositions use col/row format, need to convert to match winningPositionsSet format
          // winningPositionsSet uses `${col},${gridRow}` where gridRow = visualRow + BUFFER_OFFSET
          const gridRow = pos.row + BUFFER_OFFSET
          currentAnimatingSymbolPositions.add(`${pos.col},${gridRow}`)
        }
      }
    }

    let hasGlowableTiles = false
    // Reuse array instead of allocating new one each frame
    glowableTileInfosArray.length = 0

    winningEffects.container.x = 0
    winningEffects.container.y = mainRect.y

    if (winningEffects.isActive()) {
      winningEffects.clear()
    }

    for (let col = 0; col < COLS; col++) {
      const offsetTiles = gridState.spinOffsets?.[col] ?? 0
      const velocityTiles = gridState.spinVelocities?.[col] ?? 0
      const reelStrip = gridState.reelStrips?.[col] || []
      const reelTop = gridState.reelTopIndex?.[col] ?? 0

      for (let r = -1; r < ROWS_FULL; r++) {
        const xCell = originX + col * stepX
        const yCell = startY + r * stepY + (offsetTiles * stepY)
        const gridRow = r + BUFFER_OFFSET

        if (gridRow < 0 || gridRow >= CONFIG.reels.rows) continue

        const cellKey = `${col}:${r}`

        // Determine symbol
        let symbol: string | undefined
        const animatingSymbol = dropAnimations.getAnimatingSymbol(cellKey)
        const completedSymbol = dropAnimations.getCompletedSymbol(cellKey)

        if (animatingSymbol) {
          symbol = animatingSymbol
        } else if (completedSymbol) {
          symbol = completedSymbol
        } else if (velocityTiles > 0.001) {
          if (reelStrip.length === 0) continue
          const idx = ((reelTop - gridRow) % reelStrip.length + reelStrip.length) % reelStrip.length
          symbol = reelStrip[idx]
        } else {
          symbol = gridState.grid?.[col]?.[gridRow]
        }

        if (!symbol) continue

        const isInVisibleWinRow = r >= CONFIG.reels.visualWinStartRow && r <= CONFIG.reels.visualWinEndRow
        const isWinningTile = isInVisibleWinRow && winningPositionsSet.has(`${col},${gridRow}`)

        // During lightning animation, only highlight tiles from the currently animating symbol
        // Other winning tiles should be darkened (not highlighted)
        let winning: boolean
        if (isLightningAnimating && currentAnimatingSymbolPositions !== null) {
          // Only highlight if this tile is part of the currently animating symbol
          winning = isInVisibleWinRow && currentAnimatingSymbolPositions.has(`${col},${gridRow}`)
        } else {
          // Normal case: all winning tiles are highlighted
          winning = isWinningTile
        }

        // Track if this tile is a winning tile but NOT from the current animating symbol
        // These should be darkened during lightning animation
        const isOtherWinningTile = isLightningAnimating && isWinningTile && !winning

        const winningState = winningStore.getCellState(cellKey)
        const columnIsSpinning = gridState.spinVelocities && gridState.spinVelocities[col] > 0.001

        const result = renderTile(
          {
            col,
            visualRow: r,
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
            winningStateTime: winningStore.stateStartTime,
            timingStore,
            dropAnimations,
            bumpAnimations,
            popAnimations,
            anticipationEffects,
            columnIsSpinning,
            grid: gridState.grid,
            inCascadeWindow,
            shatterAnimations,
            goldToWildAnimations,
            goldTransformedPositions: gridState.goldTransformedPositions,
            isLightningAnimating,
            isOtherWinningTile
          },
          spriteCache,
          tilesContainer
        )

        if (result) {
          const { sprite: sp, hasGlowableTile, shouldShowBonusEffects, shouldShowLightBurst, shouldShowAnticipationBurst } = result

          if (hasGlowableTile) {
            hasGlowableTiles = true
            // Collect glowable tile info for sparkle effects - only for visible rows (1-4)
            if (r >= CONFIG.reels.visualWinStartRow && r <= CONFIG.reels.visualWinEndRow) {
              glowableTileInfosArray.push({
                key: cellKey,
                x: sp.x,
                y: sp.y,
                width: sp.width,
                height: sp.height
              })
            }
          }

          if (r >= 0 && r < ROWS_FULL) {
            winningFrames.updateFrame(cellKey, sp, winning, sp.x, sp.y, false, symbol)
            // Light burst follows bonus tiles during spin, anticipation burst only when stationary
            const shouldShowBurst = shouldShowLightBurst || shouldShowAnticipationBurst
            lightBursts.updateBurst(cellKey, sp, shouldShowBurst, timestamp)
          }

          currentUsedKeys.add(cellKey)
        }
      }
    }

    // Cleanup - only check keys that were used in previous frame but not current
    // This avoids iterating the entire cache on every frame
    for (const key of previousUsedKeys) {
      if (!currentUsedKeys.has(key)) {
        const sprite = spriteCache.get(key)
        if (sprite) {
          if (sprite.parent) sprite.parent.removeChild(sprite)
          sprite.destroy({ children: true, texture: false, baseTexture: false })
          spriteCache.delete(key)
        }
      }
    }
    winningFrames.cleanup(currentUsedKeys)
    lightBursts.cleanup(currentUsedKeys)
    winningFrames.update(timestamp)

    // Update lightning links for winning tiles
    // Keep calling update() while animation is running (uses stored positions)
    // Only clear when animation is NOT running
    const hasHighlightWins = gridState.highlightWins && gridState.highlightWins.length > 0

    if (isLightningAnimating) {
      // Build positions from highlightWins if available
      const winsBySymbol = new Map<string, TilePosition[]>()

      if (hasHighlightWins && !spinning) {
        const seenPositions = new Map<string, Set<string>>()

        for (const win of gridState.highlightWins) {
          const symbol = numberToSymbol(win.symbol)

          if (!winsBySymbol.has(symbol)) {
            winsBySymbol.set(symbol, [])
            seenPositions.set(symbol, new Set())
          }
          const positions = winsBySymbol.get(symbol)!
          const seen = seenPositions.get(symbol)!

          for (const pos of win.positions) {
            const visualRow = pos.row - BUFFER_OFFSET

            // Only include visible rows (1-4 for 4-row display, matching backend winCheckStartRow/EndRow)
            if (visualRow >= 1 && visualRow <= 4) {
              const posKey = `${pos.reel}:${visualRow}`
              // Skip if we've already added this position for this symbol
              if (seen.has(posKey)) continue
              seen.add(posKey)

              const cellKey = `${pos.reel}:${visualRow}`
              const sprite = spriteCache.get(cellKey)

              // Calculate tile center position
              // Use sprite position if available, otherwise calculate from grid coordinates
              let tileX: number
              let tileY: number

              if (sprite) {
                // Sprites use anchor(0.5, 0.5), so sprite.x/y IS already the center
                tileX = sprite.x
                tileY = sprite.y
              } else {
                // Fallback: calculate position from grid coordinates
                // This ensures all winning tiles are included even if sprite is not in cache
                const xCell = originX + pos.reel * stepX
                const yCell = startY + visualRow * stepY
                tileX = xCell + scaledTileW / 2
                tileY = yCell + scaledTileH / 2
              }

              positions.push({
                x: tileX,
                y: tileY,
                col: pos.reel,
                row: visualRow
              })
            }
          }
        }
      }

      // Always call update when animating - it will use stored positions if needed
      lightningLinks.update(now, winsBySymbol)
    } else if (!hasHighlightWins) {
      // Only clear when NOT animating and no highlight wins
      lightningLinks.clear()
    }

    gridState.hasGlowableTiles = hasGlowableTiles
    gridState.glowableTileInfos = glowableTileInfosArray
  }

  container.addChild(winningEffects.container)
  container.addChild(goldToWildAnimations.container)
  framesContainer.addChild(winningFrames.container)
  framesContainer.addChild(lightningLinks.container)
  lightningLinks.container.zIndex = 1000 // Above winning frames

  function triggerPop(col: number, visualRow: number): void {
    const cellKey = `${col}:${visualRow}`
    const sprite = spriteCache.get(cellKey)
    if (sprite) {
      popAnimations.startPop(cellKey, sprite)
    }
  }

  return {
    container,
    tilesContainer,
    draw,
    winningEffects,
    goldToWildAnimations,
    triggerPop,
    getSpriteCache: () => spriteCache,
    startGSAPReelScroll: reelScroller.startGSAPReelScroll,
    stopGSAPReelScroll: reelScroller.stopGSAPReelScroll,
    stopAllGSAPReelScrolls: reelScroller.stopAllGSAPReelScrolls,
    gsapSpinManager,
    reelContainers,
    isColumnSpinning: (col: number) => reelSpinData.has(col),
    clearCompletedShatterAnimations: () => shatterAnimations.clearCompleted(),
    lightningLinks,
    setFrameTheme
  }
}
