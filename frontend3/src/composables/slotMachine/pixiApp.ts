import { Application } from 'pixi.js'
import { watch, type Ref } from 'vue'
import { useGameStore } from '@/stores'
import { CONFIG } from '@/config/constants'

/**
 * Canvas state interface
 */
export interface CanvasState {
  canvas: Ref<HTMLCanvasElement | null>
}

/**
 * Pending size for renderer resize
 */
interface PendingSize {
  width: number
  height: number
}

/**
 * Pixi app composable interface
 */
export interface UsePixiApp {
  ensure: (width: number, height: number) => Promise<void> | null
  isReady: () => boolean
  getApp: () => Application | null
  getCanvas: () => HTMLCanvasElement | null
  updateCanvasPosition: () => void
  destroy: () => void
}

export function usePixiApp(canvasState: CanvasState): UsePixiApp {
  const gameStore = useGameStore()
  let app: Application | null = null
  let initPromise: Promise<void> | null = null
  let canvasEl: HTMLCanvasElement | null = null
  let pendingSize: PendingSize | null = null

  // Canvas is now always visible since background is rendered in PixiJS
  // No need to toggle visibility based on start screen

  function ensure(width: number, height: number): Promise<void> | null {
    if (!app) {
      app = new Application()
      canvasEl = document.createElement('canvas')

      const vw = window.innerWidth
      const vh = window.innerHeight

      // Canvas is now FULL SCREEN - covers entire viewport
      // Game content will be centered within the canvas
      const canvasWidth = vw
      const canvasHeight = vh

      // Position canvas to cover entire viewport
      canvasEl.style.position = 'fixed'
      canvasEl.style.left = '0'
      canvasEl.style.top = '0'
      canvasEl.style.width = `${canvasWidth}px`
      canvasEl.style.height = `${canvasHeight}px`
      canvasEl.style.pointerEvents = 'auto'
      canvasEl.style.zIndex = '1' // Behind Vue UI components but above body background
      canvasEl.style.visibility = 'visible' // Always visible - background renders in canvas
      // Ensure crisp rendering on high-DPI displays
      canvasEl.style.imageRendering = 'auto'
      document.body.appendChild(canvasEl)

      if (canvasState.canvas.value) {
        canvasState.canvas.value.style.display = 'none'
      }

      pendingSize = { width, height }

      // Use full device pixel ratio for crisp rendering
      const isMobile = vw < 600
      const resolution = window.devicePixelRatio || 2

      initPromise = app.init({
        width,
        height,
        antialias: true, // Always enable antialiasing for quality
        resolution,
        autoDensity: true,
        backgroundAlpha: 0,
        canvas: canvasEl,
        preference: 'webgl', // Prefer WebGL for better quality
        roundPixels: true, // Snap to pixel grid for sharp rendering
        // Disable texture garbage collection to prevent crashes during HMR
        // TextureGCSystem can crash when textures are destroyed during hot reload
        textureGCActive: false,
      })
        .then(() => {
          if (pendingSize && app?.renderer) {
            app.renderer.resize(pendingSize.width, pendingSize.height)
            pendingSize = null
          }
          // Completely disable texture GC to prevent glyph textures from being unloaded
          // This fixes the "Cannot read properties of null" error in TextureGCSystem
          // The error occurs when TextureGCSystem unloads texture sources that are still in use
          if (app?.renderer?.textureGC) {
            // Set thresholds to prevent automatic cleanup
            app.renderer.textureGC.checkCountMax = Infinity
            app.renderer.textureGC.maxIdle = Infinity
            // Override the run method to completely prevent GC from running
            app.renderer.textureGC.run = () => {}
          }
        })
        .catch(() => {})
    } else {
      // Canvas already exists, update position and size
      updateCanvasPosition()
      
      if (app.renderer) {
        app.renderer.resize(width, height)
      } else {
        pendingSize = { width, height }
      }
    }
    return initPromise
  }

  function updateCanvasPosition(): void {
    if (!canvasEl) return

    const vw = window.innerWidth
    const vh = window.innerHeight

    // Canvas is FULL SCREEN - always covers entire viewport
    canvasEl.style.left = '0'
    canvasEl.style.top = '0'
    canvasEl.style.width = `${vw}px`
    canvasEl.style.height = `${vh}px`
  }

  const isReady = (): boolean => !!app?.renderer
  const getApp = (): Application | null => app
  const getCanvas = (): HTMLCanvasElement | null => canvasEl

  function destroy(): void {
    // Destroy the PixiJS app
    if (app) {
      try {
        // Destroy app but KEEP textures intact
        // Textures are cached globally in spritesheet.ts and ASSETS.loadedImages
        // Destroying them causes issues when the app is re-created (e.g., after logout/login)
        app.destroy(true, {
          children: true,
          texture: false,       // IMPORTANT: Don't destroy cached textures
          textureSource: false, // IMPORTANT: Don't destroy texture sources
          context: true
        })
      } catch {
        // Silent fail
      }
      app = null
    }

    // Remove canvas from DOM
    if (canvasEl) {
      try {
        if (canvasEl.parentNode) {
          canvasEl.parentNode.removeChild(canvasEl)
        }
      } catch {
        // Silent fail
      }
      canvasEl = null
    }

    // Reset state
    initPromise = null
    pendingSize = null
  }

  return { ensure, isReady, getApp, getCanvas, updateCanvasPosition, destroy }
}
