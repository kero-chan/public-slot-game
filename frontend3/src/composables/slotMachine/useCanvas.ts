import { ref, type Ref } from 'vue'
import { CONFIG } from '@/config/constants'

/**
 * Button configuration interface
 */
export interface ButtonConfig {
  x: number
  y: number
  width?: number
  height?: number
  radius?: number
}

/**
 * Buttons state interface
 */
export interface Buttons {
  spin: ButtonConfig
  betPlus: ButtonConfig
  betMinus: ButtonConfig
  start: ButtonConfig
}

/**
 * Reel offset interface
 */
export interface ReelOffset {
  x: number
  y: number
}

/**
 * Canvas composable interface
 */
export interface UseCanvas {
  canvas: Ref<HTMLCanvasElement | null>
  canvasWidth: Ref<number>
  canvasHeight: Ref<number>
  gameWidth: Ref<number>
  gameHeight: Ref<number>
  gameOffsetX: Ref<number>
  gameOffsetY: Ref<number>
  scale: Ref<number>
  reelOffset: Ref<ReelOffset>
  buttons: Ref<Buttons>
  setupCanvas: () => Promise<void>
}

/**
 * useCanvas - PixiJS-only canvas state management
 * NO 2D canvas creation - only calculations and state
 * PixiJS handles all canvas creation and rendering
 */
export function useCanvas(canvasRef: HTMLCanvasElement | null): UseCanvas {
  // State only - no actual canvas element created here
  const canvas = ref<HTMLCanvasElement | null>(canvasRef)  // Just store the ref for potential future use

  const canvasWidth = ref(0)   // Full viewport width
  const canvasHeight = ref(0)  // Full viewport height
  const gameWidth = ref(0)     // Game area width (maintains aspect ratio)
  const gameHeight = ref(0)    // Game area height (maintains aspect ratio)
  const gameOffsetX = ref(0)   // Horizontal offset to center game
  const gameOffsetY = ref(0)   // Vertical offset to center game vertically
  const scale = ref(1)
  const reelOffset = ref<ReelOffset>({ x: 0, y: 0 })

  const buttons = ref<Buttons>({
    spin: { x: 0, y: 0, radius: 50 },
    betPlus: { x: 0, y: 0, width: 60, height: 50 },
    betMinus: { x: 0, y: 0, width: 60, height: 50 },
    start: { x: 0, y: 0, width: 0, height: 0 }
  })

  const setupCanvas = async (): Promise<void> => {
    // Canvas is now FULL SCREEN - covers entire viewport
    // Game content is centered within the full-screen canvas

    const vw = window.innerWidth
    const vh = window.innerHeight

    // Canvas dimensions = full viewport
    canvasWidth.value = vw
    canvasHeight.value = vh

    // Calculate game dimensions maintaining aspect ratio
    // Try fitting based on width first
    const gw1 = vw
    const gh1 = gw1 / CONFIG.canvas.aspectRatio
    
    // Try fitting based on height
    const gh2 = vh
    const gw2 = gh2 * CONFIG.canvas.aspectRatio
    
    // Use the dimensions that fit within viewport (smaller of the two)
    let gw: number
    let gh: number
    
    if (gh1 <= vh) {
      // Width-based calculation fits within viewport
      gw = gw1
      gh = gh1
    } else {
      // Height-based calculation fits within viewport
      gw = gw2
      gh = gh2
    }

    // Store game dimensions
    gameWidth.value = Math.floor(gw)
    gameHeight.value = Math.floor(gh)
    gameOffsetX.value = Math.floor((vw - gw) / 2)
    gameOffsetY.value = Math.floor((vh - gh) / 2)  // Center vertically

    // Scale based on game area dimensions
    const scaleX = gw / CONFIG.canvas.baseWidth
    const scaleY = gh / CONFIG.canvas.baseHeight
    scale.value = Math.min(scaleX, scaleY)

    // Reel positioning (centered in game area)
    const symbolSize = Math.floor(CONFIG.reels.symbolSize * scale.value)
    const spacing = Math.floor(CONFIG.reels.spacing * scale.value)
    const reelAreaWidth = symbolSize * CONFIG.reels.count + spacing * (CONFIG.reels.count - 1)

    reelOffset.value.x = gameOffsetX.value + (gw - reelAreaWidth) / 2
    reelOffset.value.y = gameOffsetY.value + Math.floor(gh * 0.25)

    // Button positions (centered in game area)
    buttons.value.spin.x = gameOffsetX.value + Math.floor(gw / 2)  // Center of game area
    buttons.value.spin.y = gameOffsetY.value + gh - Math.floor(160 * scale.value)
    buttons.value.spin.radius = Math.floor(85 * scale.value)

    const controlSize = Math.floor(72 * scale.value)
    const gap = Math.floor(36 * scale.value)

    // Minus (left of spin)
    buttons.value.betMinus.width = controlSize
    buttons.value.betMinus.height = controlSize
    buttons.value.betMinus.x = buttons.value.spin.x - (buttons.value.spin.radius || 0) - gap - controlSize
    buttons.value.betMinus.y = buttons.value.spin.y - Math.floor(controlSize / 2)

    // Plus (right of spin)
    buttons.value.betPlus.width = controlSize
    buttons.value.betPlus.height = controlSize
    buttons.value.betPlus.x = buttons.value.spin.x + (buttons.value.spin.radius || 0) + gap
    buttons.value.betPlus.y = buttons.value.spin.y - Math.floor(controlSize / 2)

    // Start screen button placement (bottom center of game area)
    const sbWidth = Math.floor(280 * scale.value)
    const sbHeight = Math.floor(64 * scale.value)
    buttons.value.start.x = gameOffsetX.value + Math.floor((gw - sbWidth) / 2)
    buttons.value.start.y = gameOffsetY.value + gh - Math.floor(24 * scale.value) - sbHeight
    buttons.value.start.width = sbWidth
    buttons.value.start.height = sbHeight
  }

  return {
    canvas,
    canvasWidth,
    canvasHeight,
    gameWidth,
    gameHeight,
    gameOffsetX,
    gameOffsetY,
    scale,
    reelOffset,
    buttons,
    setupCanvas
  }
}
