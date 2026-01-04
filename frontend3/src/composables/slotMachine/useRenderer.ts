// @ts-nocheck
import { Container, Sprite, type Application } from 'pixi.js'
import { watch } from 'vue'
import { usePixiApp, type UsePixiApp } from './pixiApp'
import { useBackground, type UseBackground } from './background'
import { useHeader, type UseHeader, type HeaderRect } from './header'
import { useReels, type UseReels, type ReelsRect, type TileSize, type FrameTheme } from './reels'
import { useFooter, type UseFooter, type FooterRect } from './footer'
import { useGlowOverlay } from '@/composables/slotMachine/reels/tiles/glowingComposer'
import { createBonusOverlay } from '@/composables/slotMachine/overlay/bonusOverlay'
import { createRetriggerOverlay } from '@/composables/slotMachine/overlay/retriggerOverlay'
import { createJackpotTriggerAnimation } from '@/composables/slotMachine/overlay/jackpotTriggerAnimation'
import { createBonusTilePopAnimation } from '@/composables/slotMachine/overlay/bonusTilePopAnimation'
import { createJackpotResultOverlay } from '@/composables/slotMachine/overlay/jackpotResultOverlay'
import { createWinningSparkles } from '@/composables/slotMachine/reels/winning/winningSparkles'
import { createWinAnimationManager } from '@/composables/slotMachine/overlay/winAnimationManager'
import { createSymbolWinAnimation, hasCardImages } from '@/composables/slotMachine/effects/symbolWinAnimation'
import { createAmbientParticles, type AmbientParticles } from '@/composables/slotMachine/effects/ambientParticles'
import { createFreeSpinCountdown, type FreeSpinCountdownOverlay } from '@/composables/slotMachine/overlay/freeSpinCountdown'
import { createGameModeSelectionOverlay, type GameModeSelectionOverlay } from '@/composables/slotMachine/overlay/gameModeSelectionOverlay'
import { useGameStore, useFreeSpinsStore, useUIStore, GAME_STATES } from '@/stores'
import { CONFIG } from '@/config/constants'
import { getIconSprite } from '@/config/spritesheet'
import { audioEvents, AUDIO_EVENTS } from '@/composables/audioEventBus'
import { howlerAudio } from '@/composables/useHowlerAudio'
import type { UseGameState } from '@/composables/slotMachine/useGameState'
import type { UseCanvas } from '@/composables/slotMachine/useCanvas'
import type { GridState } from '@/types/global'

/**
 * Control handlers for footer buttons
 */
export interface ControlHandlers {
  spin?: (gameMode?: string) => void | Promise<void>
  increaseBet?: () => void
  decreaseBet?: () => void
  onGameModeClick?: () => void
}

/**
 * Layout computation result
 */
interface LayoutResult {
  headerRect: HeaderRect
  mainRect: ReelsRect
  footerRect: FooterRect
  tileSize: TileSize
}

/**
 * Overlay manager interface
 */
interface OverlayManager {
  container: Container
  show?: (...args: any[]) => void
  update?: (timestamp: number) => void
  build?: (width: number, height: number) => void
  isShowing?: () => boolean
}

/**
 * Renderer composable interface
 */
export interface UseRenderer {
  init: () => Promise<void>
  render: () => void
  startAnimation: () => void
  stopAnimation: () => void
  pauseForVisibility: () => void
  resumeFromVisibility: () => void
  setControls: (handlers: ControlHandlers) => void
  showWinOverlay: (intensity: 'small' | 'medium' | 'big' | 'mega', amount: number) => void
  showFinalJackpotResult: (amount: number) => void
  showSymbolWinAnimation: (symbol: HighValueSymbol) => Promise<void>
  hideSymbolWinAnimation: () => void
  showFreeSpinCountdown: (spinsRemaining: number) => Promise<void>
  showGameModeSelection: () => void
  initializeComposedTextures: () => Promise<void>
  rebuildBackground: () => void
  getReels: () => UseReels | null
  getPixiCanvas: () => HTMLCanvasElement | null
}

export function useRenderer(
  canvasState: UseCanvas,
  gameState: UseGameState,
  gridState: GridState,
  controls?: ControlHandlers
): UseRenderer {
  // Composables
  const gameStore = useGameStore()
  const pixiApp: UsePixiApp = usePixiApp(canvasState as any)

  // Scene graph
  let app: Application | null = null
  let root: Container | null = null
  let background: UseBackground | null = null
  let header: UseHeader | null = null
  let reels: UseReels | null = null
  let footer: UseFooter | null = null
  let glowOverlay: ReturnType<typeof useGlowOverlay> | null = null
  let winningSparkles: ReturnType<typeof createWinningSparkles> | null = null
  let bonusOverlay: OverlayManager | null = null
  let retriggerOverlay: OverlayManager | null = null
  let bonusTilePopAnimation: OverlayManager | null = null
  let jackpotTriggerAnimation: OverlayManager | null = null
  let jackpotResultOverlay: OverlayManager | null = null
  let winAnimationManager: ReturnType<typeof createWinAnimationManager> | null = null
  let symbolWinAnimation: ReturnType<typeof createSymbolWinAnimation> | null = null
  let freeSpinCountdown: FreeSpinCountdownOverlay | null = null
  let gameModeSelectionOverlay: GameModeSelectionOverlay | null = null
  let ambientParticles: AmbientParticles | null = null

  // Fixed position UI elements
  let settingsButtonContainer: Container | null = null
  let settingsButton: Sprite | null = null

  // Track last layout for rebuilds
  let lastW = 0
  let lastH = 0

  // Cached layout result - only recompute when dimensions change
  let cachedLayout: LayoutResult | null = null

  // RAF handle
  let animationFrameId: number | null = null

  // Track if animation was running before visibility change (for pause/resume)
  let wasAnimatingBeforePause = false

  // Layout constants
  const MARGIN_X = 10
  const COLS = 5
  const ROWS_FULL = 4
  const TOP_PARTIAL = 0  // No partial row - show exactly 4 full rows

  // Track control handlers for footer buttons
  let controlHandlers: ControlHandlers = controls || {}
  let bambooComposed = false

  // Track if high-value symbol background is currently active (persists until another high-value win)
  let highValueBackgroundActive = false

  // Symbols that have background images available
  const SYMBOLS_WITH_BACKGROUNDS = ['fa', 'zhong', 'bai', 'bawan'] as const

  /**
   * Check if a symbol has a background image available
   */
  function hasBackgroundImage(symbol: string): boolean {
    return SYMBOLS_WITH_BACKGROUNDS.includes(symbol as any)
  }


  /**
   * Compute layout - uses cached result when dimensions haven't changed.
   * Header sticks to top (20px gap), Footer sticks to bottom (20px gap)
   * Reels fill the middle space.
   */
  function computeLayout(w: number, h: number): LayoutResult {
    // Get game dimensions (centered within full-screen canvas)
    const gw = canvasState.gameWidth.value || w
    const gh = canvasState.gameHeight.value || h
    const offsetX = canvasState.gameOffsetX.value || 0
    const offsetY = canvasState.gameOffsetY.value || 0

    // Return cached layout if dimensions haven't changed
    if (cachedLayout && w === lastW && h === lastH) {
      return cachedLayout
    }

    // Fixed gaps from top and bottom (within game area)
    const TOP_GAP = 10  // Raised header
    const BOTTOM_GAP = 5

    // Header sticks to top with gap (within game area)
    const headerH = Math.round(gh * 0.12)
    const headerY = offsetY + TOP_GAP

    // Footer sticks to bottom with gap (within game area)
    const footerH = Math.round(gh * 0.25)
    const footerY = offsetY + gh - footerH - BOTTOM_GAP

    // Main (reels) fills the space between header and footer
    const mainY = headerY + headerH
    const mainH = footerY - mainY

    // Tile width from GAME width (not canvas width); tile height from required ratio
    const tileW = (gw - MARGIN_X * 2) / COLS
    const tileH = tileW * CONFIG.canvas.tileAspectRatio

    cachedLayout = {
      headerRect: { x: offsetX, y: headerY, w: gw, h: headerH },
      mainRect:   { x: offsetX, y: mainY, w: gw, h: mainH },
      footerRect: { x: offsetX, y: footerY, w: gw, h: footerH },
      tileSize:   { w: tileW, h: tileH }
    }

    return cachedLayout
  }

  // PixiJS-only renderer: header, reels, footer
  async function ensureStage(w: number, h: number): Promise<boolean> {
    await pixiApp.ensure(w, h)
    app = pixiApp.getApp()
    if (!pixiApp.isReady()) {
      return false
    }

    if (!root) {
      root = new Container()
      root.sortableChildren = true // Enable zIndex sorting on root container
      app!.stage.sortableChildren = true
      app!.stage.addChild(root)

      // Initialize background (behind everything)
      background = useBackground()
      background.container.zIndex = -1000
      // Force initial build with current dimensions
      background.build(w, h)

      // Initialize ambient particles as falling leaves from the sky
      ambientParticles = createAmbientParticles({
        parent: root,
        spawnZones: [
          // Spawn zone is above the screen - leaves fall from the sky
          { id: 'sky', rect: { x: -50, y: -100, w: w + 100, h: 50 }, weight: 1, profile: 'FALL' }
        ],
        bounds: { w, h },
        budget: { target: 8, min: 5, max: 12 }, // Keep count low for performance
        zIndex: -900,
        maxHardCap: 15 // Limit max particles
      })
      ambientParticles.start()

      header = useHeader(gameState)
      reels = useReels(gameState, gridState)
      footer = useFooter(gameState)
      glowOverlay = useGlowOverlay(gameState, gridState)
      // Initialize star texture for efficient sparkle sprite creation
      if (app) glowOverlay.initTextures(app)
      winningSparkles = createWinningSparkles()

      // Connect winningSparkles callback to footer for balance sync
      // When particles reach wallet, trigger the pending balance update
      winningSparkles.setOnParticlesReachedWallet(() => {
        footer?.onParticlesReachedWallet()
      })

      bonusOverlay = createBonusOverlay(gameState as any)
      retriggerOverlay = createRetriggerOverlay()
      bonusTilePopAnimation = createBonusTilePopAnimation(gridState, reels)
      // Pass reels reference so jackpot trigger animation can clone actual tile sprites
      jackpotTriggerAnimation = createJackpotTriggerAnimation(reels)
      jackpotResultOverlay = createJackpotResultOverlay(gameState as any)
      // Pass reels reference so win animations can clone actual tile sprites
      winAnimationManager = createWinAnimationManager(reels)
      symbolWinAnimation = createSymbolWinAnimation()
      freeSpinCountdown = createFreeSpinCountdown()
      gameModeSelectionOverlay = createGameModeSelectionOverlay()

      // Note: Background persists after symbol win animation hides
      // It only changes when another high-value symbol win is detected

      if (controlHandlers && footer?.setHandlers) {
        // Merge control handlers with internal onGameModeClick handler
        footer.setHandlers({
          ...controlHandlers,
          onGameModeClick: showGameModeSelection
        })
      }

      // Create settings button (fixed at top-right of viewport, below overlays)
      settingsButtonContainer = new Container()
      settingsButtonContainer.zIndex = 5  // Low z-index so it stays under overlays
      settingsButton = getIconSprite('icon_btn_setting_menu.webp')
      if (settingsButton) {
        settingsButton.anchor.set(1, 0) // Anchor top-right
        settingsButton.eventMode = 'static'
        settingsButton
        const uiStore = useUIStore()
        settingsButton.on('pointerdown', () => {
          // Play generic UI sound with pitch randomization
          const howl = howlerAudio.getHowl('generic_ui')
          if (howl) {
            const randomPitch = 0.6 + Math.random() * 0.8
            howl.rate(randomPitch)
            audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { audioKey: 'generic_ui', volume: 0.6 })
          }
          uiStore.openSettings()
        })
        settingsButtonContainer.addChild(settingsButton)
      }

      // Add background first (bottom layer)
      root.addChild(background.container)
      root.addChild(symbolWinAnimation.container) // Symbol win animation above background, below game UI
      root.addChild(header.container)
      root.addChild(reels.container)
      root.addChild(glowOverlay.container)
      root.addChild(winningSparkles.container)
      root.addChild(footer.container)
      root.addChild(settingsButtonContainer) // Settings button above footer but below all overlays
      root.addChild(jackpotResultOverlay.container)
      root.addChild(bonusTilePopAnimation.container)
      root.addChild(jackpotTriggerAnimation.container)
      root.addChild(bonusOverlay.container)
      root.addChild(retriggerOverlay.container)
      root.addChild(freeSpinCountdown.container)
      root.addChild(winAnimationManager.container)
      root.addChild(gameModeSelectionOverlay.container)

      // Watch for bonus overlay state
      watch(() => gameStore.gameFlowState, (newState) => {
        const w = canvasState.canvasWidth.value
        const h = canvasState.canvasHeight.value

        // Show tile pop animation first
        if (newState === GAME_STATES.POPPING_BONUS_TILES && bonusTilePopAnimation) {
          const { mainRect, tileSize } = computeLayout(w, h)
          bonusTilePopAnimation.show?.(w, h, mainRect, tileSize, () => {
            try {
              gameStore.completeBonusTilePop()
            } catch (error) {
              console.error('Error completing bonus tile pop:', error)
            }
          })
        }

        // Show jackpot trigger animation after pop animation
        if (newState === GAME_STATES.SHOWING_JACKPOT_ANIMATION && jackpotTriggerAnimation) {
          // Get bonus tile positions from the grid for the animation
          // Include width, height, col, row for pre-overlay tile effects
          const bonusTilePositions: Array<{x: number, y: number, width: number, height: number, col: number, row: number}> = []
          const { mainRect, tileSize } = computeLayout(w, h)

          // Find bonus tiles in the current grid
          for (let col = 0; col < 5; col++) {
            for (let row = 0; row < 4; row++) {
              const symbol = gridState.grid[col]?.[row + 4] // Visible rows start at index 4
              if (symbol === 'bonus') {
                // Calculate screen position (center of tile)
                const x = mainRect.x + col * tileSize.tileW + tileSize.tileW / 2
                const y = mainRect.y + row * tileSize.tileH + tileSize.tileH / 2
                bonusTilePositions.push({
                  x,
                  y,
                  width: tileSize.tileW,
                  height: tileSize.tileH,
                  col,
                  row: row + 4 // Actual row in grid including buffer
                })
              }
            }
          }

          jackpotTriggerAnimation.show?.(w, h, bonusTilePositions, () => {
            try {
              gameStore.completeJackpotAnimation()
            } catch (error) {
              console.error('Error completing jackpot animation:', error)
            }
          })
        }

        // Then show bonus overlay after animation
        if (newState === GAME_STATES.SHOWING_BONUS_OVERLAY && bonusOverlay) {
          const freeSpinsCount = gameStore.freeSpins
          bonusOverlay.show?.(freeSpinsCount, w, h, () => {
            try {
              gameStore.completeBonusOverlay()
            } catch (error) {
              console.error('Error completing bonus overlay:', error)
            }
          })
        }

        // Show retrigger overlay when free spins are retriggered
        if (newState === GAME_STATES.SHOWING_RETRIGGER_OVERLAY && retriggerOverlay) {
          const freeSpinsStore = useFreeSpinsStore()
          const additionalSpins = freeSpinsStore.pendingRetriggerSpins
          retriggerOverlay.show?.(additionalSpins, w, h, () => {
            try {
              gameStore.completeRetriggerOverlay()
            } catch (error) {
              console.error('Error completing retrigger overlay:', error)
            }
          })
        }
      })
    }

    // Initialize consecutive wins composed textures (removed from here - will be called after asset loading)
    return true
  }

  async function initializeComposedTextures(): Promise<void> {
    if (header?.initializeComposedTextures && app) {
      header.initializeComposedTextures(app)
    }
  }

  function updateOnce(timestamp: number = 0): void {
    const w = canvasState.canvasWidth.value
    const h = canvasState.canvasHeight.value
    if (!w || !h) return

    // Check if size changed and need to resize PixiJS
    const resized = w !== lastW || h !== lastH
    if (resized) {
        // Resize PixiJS app to match new canvas dimensions
        pixiApp.ensure(w, h)
        // Invalidate layout cache before updating lastW/lastH
        cachedLayout = null
        lastW = w
        lastH = h
    }

    // If stage not ready, skip this frame (will be called again)
    if (!pixiApp.isReady()) return

    // Ensure stage is initialized (sync call after first init)
    if (!root) {
      // This shouldn't happen if init() was called properly
      return
    }

    const { headerRect, mainRect, footerRect, tileSize } = computeLayout(w, h)

    // Get game width for content sizing (distinct from full canvas width)
    const gw = canvasState.gameWidth.value || w

    // Always update background (visible on both start screen and game)
    if (background) {
      background.container.visible = true
      // Switch texture based on game state
      // High-value background persists until explicitly changed by another high-value win
      const showStart = !!gameState.showStartScreen.value
      if (showStart) {
        // Start screen always uses start background
        background.setTexture('start')
        highValueBackgroundActive = false
      } else if (!highValueBackgroundActive) {
        // Only set main background if no high-value background is active
        background.setTexture('main')
      }
      // If highValueBackgroundActive is true, keep the current symbol background
      background.update(w, h)
    }

    const showStart = !!gameState.showStartScreen.value

    // Update ambient particles (falling leaves from sky) - only when game is visible
    if (ambientParticles) {
      ambientParticles.setVisible(!showStart)
      if (!showStart) {
        if (resized) {
          ambientParticles.setBounds({ w, h })
          // Spawn zone above the screen - leaves fall from the sky
          ambientParticles.setSpawnZones([
            { id: 'sky', rect: { x: -50, y: -100, w: w + 100, h: 50 }, weight: 1, profile: 'FALL' }
          ])
        }
        ambientParticles.update(16) // ~60fps frame time
      }
    }
    if (header?.container) header.container.visible = !showStart
    if (reels?.container) reels.container.visible = !showStart
    if (glowOverlay?.container) glowOverlay.container.visible = !showStart
    if (winningSparkles?.container) winningSparkles.container.visible = !showStart
    if (footer?.container) footer.container.visible = !showStart
    // Win overlay visibility is controlled by its own show/hide methods

    // Update settings button position (always visible, fixed at top-right of viewport)
    if (settingsButton && settingsButtonContainer) {
      settingsButtonContainer.visible = !showStart
      const isSmartphone = w < 600
      const padding = isSmartphone ? 5 : 10
      const buttonScale = isSmartphone ? 0.04 : 0.08  // Smaller on mobile
      settingsButton.scale.set(buttonScale)
      settingsButton.position.set(w - padding, padding)
    }

    if (!showStart) {
      // Game is visible - build and render everything

      // Build header/footer when play screen becomes visible (even without resize)
      if ((resized || header?.container.children.length === 0) && header) {
        header.build(headerRect)
      }
      if (header) header.updateValues()

      // Always draw reels so spin/cascade state is reflected - use game width, not canvas width
      if (reels) reels.draw(mainRect, tileSize, timestamp, gw)

      // PERFORMANCE: Only update glow effects if there are glowable tiles (bonus/wild) visible
      // Skips expensive particle updates when no bonus/wild tiles present
      if (glowOverlay) {
        // Always update mask
        glowOverlay.draw(mainRect, tileSize as any, timestamp, gw)
        // Update sparkles with tile positions from reels (works during spin too)
        if (gridState.glowableTileInfos && gridState.glowableTileInfos.length > 0) {
          glowOverlay.updateTileSparkles(gridState.glowableTileInfos, timestamp)
        } else {
          // Cleanup when no glowable tiles
          glowOverlay.cleanup(new Set())
        }
      }

      // Get wallet position for particle fly-to animation
      const walletTarget = footer?.getWalletPosition() || null
      if (winningSparkles) {
        winningSparkles.draw(mainRect, tileSize, timestamp, gw, gridState, walletTarget)
        // Update footer's particle state for balance sync
        footer?.setParticlesActive(winningSparkles.hasActiveParticles())
      }

      if ((resized || footer?.container.children.length === 0) && footer) {
        footer.build(footerRect)
      }
      // Update footer every frame: arrow rotation + values refresh
      if (footer?.updateValues) footer.updateValues()
      if (footer?.update) footer.update(timestamp)

      // Update bonus overlay animation
      if (bonusOverlay) {
        bonusOverlay.update?.(timestamp)
        if (resized && bonusOverlay.container.visible) {
          bonusOverlay.build?.(w, h)
        }
      }

      // Update retrigger overlay animation
      if (retriggerOverlay) {
        retriggerOverlay.update?.(timestamp)
        if (resized && retriggerOverlay.container.visible) {
          retriggerOverlay.build?.(w, h)
        }
      }

      // Update bonus tile pop animation
      if (bonusTilePopAnimation) {
        bonusTilePopAnimation.update?.(timestamp)
        if (resized && bonusTilePopAnimation.container.visible) {
          bonusTilePopAnimation.build?.(w, h)
        }
      }

      // Update jackpot trigger animation
      if (jackpotTriggerAnimation) {
        jackpotTriggerAnimation.update?.(timestamp)
        if (resized && jackpotTriggerAnimation.container.visible) {
          jackpotTriggerAnimation.build?.(w, h)
        }
      }

      // Update jackpot result overlay
      if (jackpotResultOverlay) {
        jackpotResultOverlay.update?.(timestamp)
        if (resized && jackpotResultOverlay.container.visible) {
          jackpotResultOverlay.build?.(w, h)
        }
      }

      // Update win animation manager
      if (winAnimationManager) {
        winAnimationManager.update?.(timestamp)
        if (resized && winAnimationManager.isShowing?.()) {
          winAnimationManager.build?.(w, h)
        }
      }

      // Update game mode selection overlay
      if (gameModeSelectionOverlay) {
        gameModeSelectionOverlay.update?.(timestamp)
        if (resized && gameModeSelectionOverlay.isShowing?.()) {
          gameModeSelectionOverlay.build?.(w, h)
        }
      }
    }
  }

  const renderFrame = (timestamp: number = 0): void => {
    updateOnce(timestamp)
    const appInstance = pixiApp.getApp()
    if (appInstance?.renderer) {
      appInstance.renderer.render(appInstance.stage)
    }
    animationFrameId = requestAnimationFrame(renderFrame)
  }

  const init = async (): Promise<void> => {
    // Initialize stage
    const w = canvasState.canvasWidth.value
    const h = canvasState.canvasHeight.value

    if (w > 0 && h > 0) {
      await ensureStage(w, h)
    } else {
      // Try to create stage anyway for later use
      await ensureStage(1, 1)
    }
  }

  const render = (): void => {
    const ts = performance.now()
    updateOnce(ts)
    const appInstance = pixiApp.getApp()
    if (appInstance?.renderer) {
      appInstance.renderer.render(appInstance.stage)
    }
  }

  const startAnimation = (): void => {
    if (!animationFrameId) {
      animationFrameId = requestAnimationFrame(renderFrame)
    }
  }

  const stopAnimation = (): void => {
    if (animationFrameId) {
      cancelAnimationFrame(animationFrameId)
      animationFrameId = null
    }
  }

  /**
   * Pause animation when tab becomes hidden (visibility change)
   * Stops RAF loop and ambient particles to save CPU in background tabs
   */
  const pauseForVisibility = (): void => {
    wasAnimatingBeforePause = animationFrameId !== null
    if (animationFrameId) {
      cancelAnimationFrame(animationFrameId)
      animationFrameId = null
    }
    // Stop ambient particles
    if (ambientParticles) {
      ambientParticles.stop()
    }
  }

  /**
   * Resume animation when tab becomes visible again
   * Only resumes if animation was running before pause
   */
  const resumeFromVisibility = (): void => {
    if (wasAnimatingBeforePause && !animationFrameId) {
      animationFrameId = requestAnimationFrame(renderFrame)
    }
    // Restart ambient particles if game is visible
    if (ambientParticles && !gameState.showStartScreen.value) {
      ambientParticles.start()
    }
  }

  // Allow wiring control handlers after construction
  const setControls = (handlers: ControlHandlers): void => {
    controlHandlers = handlers
    if (footer?.setHandlers) {
      // Merge handlers with internal onGameModeClick handler
      footer.setHandlers({
        ...handlers,
        onGameModeClick: showGameModeSelection
      })
    }
  }

  // Helper to get tile positions for animations
  const getTilePositions = () => {
    if (!reels) return []

    const positions = []
    const { mainRect, tileSize } = computeLayout(canvasState.canvasWidth.value, canvasState.canvasHeight.value)
    const gameW = canvasState.gameWidth.value || canvasState.canvasWidth.value

    // CRITICAL: Match the exact tile sizing calculation from reels/index.ts draw() function
    const TILE_SPACING = 1  // Spacing between tiles (all sides)
    // Larger margin on smartphone to keep tiles within frame's visible area
    const isSmartphone = gameW < 600
    const margin = isSmartphone ? 22 : 10
    const totalSpacingX = TILE_SPACING * (COLS - 1)
    const availableWidth = gameW - (margin * 2) - totalSpacingX
    const scaledTileW = availableWidth / COLS  // Actual rendered tile width
    const scaledTileH = scaledTileW * (tileSize.h / tileSize.w)  // Maintain aspect ratio
    const stepX = scaledTileW + TILE_SPACING  // Tile width + spacing
    const stepY = scaledTileH + TILE_SPACING  // Tile height + spacing
    // originX includes mainRect.x offset for game centering
    // Shift tiles on smartphone to align with frame's visible area
    const tileOffset = isSmartphone ? -3 : 0
    const originX = mainRect.x + margin + tileOffset

    // Calculate positions for all tiles in the grid (5 columns x 4 rows)
    // This matches how tiles are positioned in reels/index.ts
    for (let col = 0; col < COLS; col++) {
      for (let row = 0; row < ROWS_FULL; row++) {
        const xCell = originX + col * stepX
        const yCell = mainRect.y + (row + TOP_PARTIAL) * stepY

        // Position at center of tile (matching reels positioning)
        const BLEED = 2
        const w = scaledTileW + BLEED * 2
        const h = scaledTileH + BLEED * 2
        const x = Math.round(xCell) - BLEED + w / 2
        const y = yCell - BLEED + h / 2

        positions.push({
          x,
          y,
          width: scaledTileW,  // Use actual scaled width (without BLEED)
          height: scaledTileH, // Use actual scaled height (without BLEED)
          col,
          row: row + 4  // Convert visual row to grid row (add BUFFER_OFFSET)
        })
      }
    }

    return positions
  }

  // Expose win overlay for game logic to trigger
  // After win overlay completes, shows high-value card animation if applicable
  const showWinOverlay = (intensity: 'small' | 'medium' | 'big' | 'mega', amount: number): void => {
    const w = canvasState.canvasWidth.value
    const h = canvasState.canvasHeight.value

    if (winAnimationManager) {
      // Show appropriate win animation based on intensity
      const tilePositions = getTilePositions()
      winAnimationManager.show(w, h, tilePositions, intensity, amount, async () => {
        try {
          // After win overlay dismisses, check for winning symbol
          const winningSymbol = gameStore.highValueWinSymbol

          // Show card animation if symbol has card images (front/back)
          if (winningSymbol && hasCardImages(winningSymbol) && symbolWinAnimation) {
            // Get grid rect for positioning the card over the reels
            const { mainRect } = computeLayout(w, h)
            const gridRect = { x: mainRect.x, y: mainRect.y, w: mainRect.w, h: mainRect.h }

            // Show the winning card animation (user must click skip to dismiss)
            await symbolWinAnimation.show(winningSymbol, w, h, gridRect)

            // AFTER card dismissed, change background/frame ONLY if symbol has background image
            if (hasBackgroundImage(winningSymbol)) {
              if (background) {
                background.setTexture(winningSymbol)
                background.update(w, h)
                highValueBackgroundActive = true
              }

              // Update all three frame themes
              const frameTheme = winningSymbol as FrameTheme
              if (reels) reels.setFrameTheme(frameTheme)
              if (header) header.setFrameTheme(frameTheme)
              if (footer) footer.setFrameTheme(frameTheme)
            }
          }

          // Clear the symbol after processing
          gameStore.setHighValueWinSymbol(null)
        } catch (error) {
          console.error('Error in win overlay callback:', error)
        } finally {
          // Always notify game state when done, even if an error occurred
          gameStore.hideWinOverlay() // Clear the showingWinOverlay flag
          gameStore.completeWinOverlay()
        }
      })
      gameStore.showWinOverlay()
    }
  }

  // Expose final jackpot result overlay for game logic to trigger
  const showFinalJackpotResult = (amount: number): void => {
    const w = canvasState.canvasWidth.value
    const h = canvasState.canvasHeight.value

    if (jackpotResultOverlay) {
      jackpotResultOverlay.show?.(amount, w, h)
      gameStore.showWinOverlay()
    }
  }

  // Expose symbol win animation for symbols with card images
  // Flow: Show card (user clicks skip), THEN change background + frames if applicable
  const showSymbolWinAnimation = async (symbol: string): Promise<void> => {
    // Only show animation if symbol has card images
    if (!hasCardImages(symbol)) {
      return
    }

    const w = canvasState.canvasWidth.value
    const h = canvasState.canvasHeight.value

    // Get grid rect for positioning the card over the reels
    const { mainRect } = computeLayout(w, h)
    const gridRect = { x: mainRect.x, y: mainRect.y, w: mainRect.w, h: mainRect.h }

    // Show the winning card animation (user must click skip to dismiss)
    if (symbolWinAnimation) {
      await symbolWinAnimation.show(symbol, w, h, gridRect)
    }

    // AFTER card dismissed, change background/frame ONLY if symbol has background image
    if (hasBackgroundImage(symbol)) {
      if (background) {
        background.setTexture(symbol)
        background.update(w, h)
        highValueBackgroundActive = true
      }

      // Update all three frame themes to match the winning symbol
      const frameTheme = symbol as FrameTheme
      if (reels) {
        reels.setFrameTheme(frameTheme)
      }
      if (header) {
        header.setFrameTheme(frameTheme)
      }
      if (footer) {
        footer.setFrameTheme(frameTheme)
      }
    }
  }

  // Hide symbol win animation (background and frame themes persist until next high-value win)
  const hideSymbolWinAnimation = (): void => {
    if (symbolWinAnimation) {
      symbolWinAnimation.hide()
    }
  }

  // Show free spin countdown overlay before each free spin
  // DISABLED: Free spin counter is now permanently pinned in the header
  const showFreeSpinCountdown = async (spinsRemaining: number): Promise<void> => {
    // Countdown overlay disabled - counter is now in header
    return
  }

  // Show game mode selection overlay
  const showGameModeSelection = (): void => {
    // Don't show during free spin mode or when game is not idle
    if (gameState.inFreeSpinMode.value) return
    if (gameStore.gameFlowState !== GAME_STATES.IDLE) return

    if (gameModeSelectionOverlay) {
      const w = canvasState.canvasWidth.value
      const h = canvasState.canvasHeight.value
      gameModeSelectionOverlay.show(w, h, (modeId: string) => {
        // Call spin with the selected game mode
        if (controlHandlers.spin) {
          controlHandlers.spin(modeId)
        }
      })
    }
  }

  // Expose reels API for GSAP-driven animations
  const getReels = (): UseReels | null => reels

  // Expose PixiJS canvas for click handling
  const getPixiCanvas = (): HTMLCanvasElement | null => pixiApp.getCanvas()

  // Rebuild background after assets are loaded
  const rebuildBackground = (): void => {
    if (background) {
      const w = canvasState.canvasWidth.value
      const h = canvasState.canvasHeight.value
      if (w > 0 && h > 0) {
        background.build(w, h)
      }
    }
  }

  // Destroy function to cleanup PixiJS resources
  const destroy = (): void => {
    stopAnimation()

    // Destroy all containers and overlays
    if (root) {
      root.destroy({ children: true })
      root = null
    }

    // Destroy ambient particles explicitly (has its own destroy method)
    if (ambientParticles) {
      ambientParticles.destroy()
      ambientParticles = null
    }

    // Reset all references
    background = null
    header = null
    reels = null
    footer = null
    glowOverlay = null
    winningSparkles = null
    bonusOverlay = null
    retriggerOverlay = null
    bonusTilePopAnimation = null
    jackpotTriggerAnimation = null
    jackpotResultOverlay = null
    winAnimationManager = null
    symbolWinAnimation = null
    freeSpinCountdown = null
    gameModeSelectionOverlay = null
    app = null

    // Destroy pixiApp last
    pixiApp.destroy()
  }

  return {
    init,
    render,
    startAnimation,
    stopAnimation,
    pauseForVisibility,
    resumeFromVisibility,
    setControls,
    showWinOverlay,
    showFinalJackpotResult,
    showSymbolWinAnimation,
    hideSymbolWinAnimation,
    showFreeSpinCountdown,
    showGameModeSelection,
    initializeComposedTextures,
    rebuildBackground,
    getReels,
    getPixiCanvas,
    destroy
  }
}
