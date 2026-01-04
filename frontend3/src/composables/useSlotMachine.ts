// @ts-nocheck
import { nextTick, watch } from 'vue'
import { useGameState, type UseGameState } from '@/composables/slotMachine/useGameState'
import { useCanvas, type UseCanvas } from '@/composables/slotMachine/useCanvas'
import { useRenderer, type UseRenderer } from '@/composables/slotMachine/useRenderer'
import { useGameLogic, type UseGameLogic } from '@/composables/slotMachine/useGameLogic'
import { useGameFlowController, type UseGameFlowController } from '@/composables/slotMachine/useGameFlowController'
import { loadAllAssets, AssetInitializationError } from '@/utils/imageLoader'
import { useGameStore, GAME_STATES, useGridStore } from '@/stores'
import { useUIStore } from '@/stores/ui/uiStore'
import type { GridState } from '@/types/global'

/**
 * Slot machine composable interface
 */
export interface UseSlotMachine {
  gameState: UseGameState
  gridState: GridState
  canvasState: UseCanvas
  init: () => Promise<void>
  render: () => void
  handleResize: () => Promise<void>
  handleCanvasClick: (e: MouseEvent) => void
  handleCanvasTouch: (e: TouchEvent) => void
  handleKeydown: (e: KeyboardEvent) => void
  spin: () => Promise<void>
  increaseBet: () => void
  decreaseBet: () => void
  start: () => void
  stopAnimation: () => void
  pauseForVisibility: () => void
  resumeFromVisibility: () => void
}

export function useSlotMachine(canvasRef: HTMLCanvasElement | null): UseSlotMachine {
  const gameStore = useGameStore()
  const uiStore = useUIStore()
  const gameState = useGameState()
  const gridState = useGridStore()
  const canvasState = useCanvas(canvasRef)
  const renderer = useRenderer(canvasState, gameState, gridState)
  const gameLogic = useGameLogic(
    gameState,
    gridState,
    renderer.render,
    renderer.showWinOverlay,
    renderer.showFinalJackpotResult,
    renderer.showSymbolWinAnimation,
    renderer.hideSymbolWinAnimation
  )

  const flowController: UseGameFlowController = useGameFlowController(gameLogic, gridState, renderer.render, gameState)
  let unwatchFlow: (() => void) | null = null

  renderer.setControls({
    spin: gameLogic.spin,
    increaseBet: gameLogic.increaseBet,
    decreaseBet: gameLogic.decreaseBet
  })

  // Flag to track if reels have been wired
  let reelsWired = false

  const init = async (): Promise<void> => {
    try {
      await nextTick()

      await canvasState.setupCanvas()

      await nextTick() // Give canvas time to settle

      // Initialize renderer first
      await renderer.init()

      // Wire up reels reference to game logic for GSAP-driven animations
      if (renderer.getReels && gameLogic.setReels && !reelsWired) {
        const reelsAPI = renderer.getReels()
        if (reelsAPI) {
          gameLogic.setReels(reelsAPI)
          reelsWired = true
        }
      }

      // Load assets with progress tracking
      await loadAllAssets((loaded: number, total: number) => {
        gameStore.updateLoadingProgress(loaded, total)
      })

      // Rebuild background AFTER assets are loaded (textures are now available)
      renderer.rebuildBackground()

      // Initialize composed textures AFTER assets are loaded
      await renderer.initializeComposedTextures()

      unwatchFlow = flowController.startWatching()
      renderer.startAnimation()

      // Watch for free spin auto-roll - use a robust lock mechanism
      let freeSpinLock = false // True while countdown+spin sequence is in progress
      let pendingFreeSpinTrigger = false // True if a FREE_SPINS_ACTIVE trigger is pending

      /**
       * Execute the countdown and spin sequence
       * IMPORTANT: This must complete the countdown BEFORE spinning
       */
      async function executeCountdownAndSpin(): Promise<void> {
        // Already in progress - don't start another
        if (freeSpinLock) {
          pendingFreeSpinTrigger = true
          return
        }

        // Check preconditions
        if (!gameStore.inFreeSpinMode || gameStore.freeSpins <= 0) {
          return
        }

        freeSpinLock = true
        pendingFreeSpinTrigger = false

        try {
          // Wait for state to become IDLE (flow controller does this after 100ms)
          let attempts = 0
          while (gameStore.gameFlowState !== GAME_STATES.IDLE && attempts < 20) {
            await new Promise(resolve => setTimeout(resolve, 100))
            attempts++
          }

          // Final check before countdown
          if (!gameStore.inFreeSpinMode || gameStore.freeSpins <= 0 || gameStore.gameFlowState !== GAME_STATES.IDLE) {
            return
          }

          // Show countdown overlay - MUST wait for it to complete
          await renderer.showFreeSpinCountdown(gameStore.freeSpins)

          // Extra wait to ensure overlay is fully hidden
          await new Promise(resolve => setTimeout(resolve, 300))

          // Final check after countdown - state must still be IDLE
          if (gameStore.gameFlowState !== GAME_STATES.IDLE) {
            return
          }

          // Now spin
          if (gameStore.inFreeSpinMode && gameStore.freeSpins > 0) {
            gameLogic.spin()
          }
        } finally {
          // Reset lock after spin starts (give it time to transition out of IDLE)
          setTimeout(() => {
            freeSpinLock = false
            // If another trigger came in while we were busy, handle it
            if (pendingFreeSpinTrigger && gameStore.inFreeSpinMode && gameStore.freeSpins > 0) {
              pendingFreeSpinTrigger = false
              // Check state - only proceed if back in FREE_SPINS_ACTIVE or IDLE
              if (gameStore.gameFlowState === GAME_STATES.IDLE) {
                executeCountdownAndSpin()
              }
            }
          }, 500)
        }
      }

      // Watch for start screen changes to ensure render and handle free spin resume
      watch(() => gameState.showStartScreen.value, async (isShowing) => {
        if (!isShowing) {
          // Force a render when game becomes visible
          await nextTick()
          renderer.render()

          // Check if we're resuming a free spin session
          if (gameStore.inFreeSpinMode && gameStore.freeSpins > 0) {
            // Wait for UI to settle, then start countdown+spin
            setTimeout(() => {
              if (gameStore.inFreeSpinMode && gameStore.freeSpins > 0) {
                executeCountdownAndSpin()
              }
            }, 1000)
          }
        }
      })

      watch(() => gameStore.gameFlowState, async (newState, oldState) => {
        // When FREE_SPINS_ACTIVE state is entered, trigger countdown+spin sequence
        if (newState === GAME_STATES.FREE_SPINS_ACTIVE &&
            gameStore.inFreeSpinMode &&
            gameStore.freeSpins > 0) {
          // Small delay to let flow controller transition to IDLE
          setTimeout(() => {
            executeCountdownAndSpin()
          }, 200)
        }
      })
    } catch (err) {
      // Handle initialization errors
      if (err instanceof AssetInitializationError) {
        console.error('❌ Asset initialization failed:', err.message)
        uiStore.setInitializationError(err.message)
      } else if (err instanceof Error) {
        console.error('❌ Initialization failed:', err.message)
        uiStore.setInitializationError('Failed to initialize game. Please refresh the page.')
      } else {
        console.error('❌ Unknown initialization error:', err)
        uiStore.setInitializationError('An unexpected error occurred.')
      }
    }
  }

  const handleResize = async (): Promise<void> => {
    await canvasState.setupCanvas()
    // Force renderer to rebuild layout and rerender
    renderer.render()
  }

  const start = (): void => {
    if (gameState.showStartScreen.value) {
      gameStore.hideStartScreen()
    }
  }

  // Debounce mechanism to prevent double-click/touch processing
  let isProcessingClick = false
  let lastSpinTime = 0
  const SPIN_DEBOUNCE_MS = 500

  const processClick = (x: number, y: number): void => {
    if (gameState.showStartScreen.value) {
      const sb = canvasState.buttons.value.start
      const hit = x >= sb.x && x <= sb.x + sb.width &&
                  y >= sb.y && y <= sb.y + sb.height
      if (hit) start()
      return
    }

    // Spin button hit-test using circular geometry from canvasState
    const spin = canvasState.buttons.value.spin
    const dx = x - spin.x
    const dy = y - spin.y
    const insideSpin = dx * dx + dy * dy <= (spin.radius || 0) * (spin.radius || 0)
    if (insideSpin) {
      // Prevent rapid double-clicks with both flag and timestamp
      const now = Date.now()
      if (isProcessingClick || (now - lastSpinTime < SPIN_DEBOUNCE_MS)) {
        return
      }
      isProcessingClick = true
      lastSpinTime = now
      gameLogic.spin()
      // Reset after a short delay
      setTimeout(() => {
        isProcessingClick = false
      }, SPIN_DEBOUNCE_MS)
      return
    }

    const minusBtn = canvasState.buttons.value.betMinus
    if (x >= minusBtn.x && x <= minusBtn.x + (minusBtn.width || 0) &&
        y >= minusBtn.y && y <= minusBtn.y + (minusBtn.height || 0)) {
      gameLogic.decreaseBet()
      return
    }

    const plusBtn = canvasState.buttons.value.betPlus
    if (x >= plusBtn.x && x <= plusBtn.x + (plusBtn.width || 0) &&
        y >= plusBtn.y && y <= plusBtn.y + (plusBtn.height || 0)) {
      gameLogic.increaseBet()
      return
    }
  }

  const handleCanvasClick = (e: MouseEvent): void => {
    // Get PixiJS canvas element
    const canvasEl = renderer.getPixiCanvas()
    if (!canvasEl) return
    const rect = canvasEl.getBoundingClientRect()
    const x = e.clientX - rect.left
    const y = e.clientY - rect.top
    processClick(x, y)
  }

  const handleCanvasTouch = (e: TouchEvent): void => {
    // Get PixiJS canvas element
    const canvasEl = renderer.getPixiCanvas()
    if (!canvasEl) return
    const touch = e.changedTouches[0]
    if (touch) {
      const rect = canvasEl.getBoundingClientRect()
      const x = touch.clientX - rect.left
      const y = touch.clientY - rect.top
      processClick(x, y)
      // Prevent synthetic click event from firing after touch
      e.preventDefault()
    }
  }

  const handleKeydown = (e: KeyboardEvent): void => {
    // Only handle start screen - keyboard spin is disabled
    if (e.key === ' ' || e.key === 'Enter') {
      e.preventDefault()
      if (gameState.showStartScreen.value) {
        start()
      }
      // Spin via keyboard is disabled - use on-screen buttons only
    }
  }

  const cleanup = (): void => {
    if (unwatchFlow) unwatchFlow()
    flowController.clearActiveTimer()
    renderer.stopAnimation()
    renderer.destroy()
  }

  return {
    gameState,
    gridState,
    canvasState,
    init,
    render: renderer.render,
    handleResize,
    handleCanvasClick,
    handleCanvasTouch,
    handleKeydown,
    spin: gameLogic.spin,
    increaseBet: gameLogic.increaseBet,
    decreaseBet: gameLogic.decreaseBet,
    start,
    stopAnimation: cleanup,
    pauseForVisibility: renderer.pauseForVisibility,
    resumeFromVisibility: renderer.resumeFromVisibility
  }
}
