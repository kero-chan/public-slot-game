import type { Container, Graphics, Sprite } from 'pixi.js'

/**
 * Base overlay interface - all overlays must implement this
 */
export interface BaseOverlay {
  container: Container
  show: (...args: any[]) => void
  hide: () => void
  update: (timestamp: number) => void
  build: (width: number, height: number) => void
  isShowing: () => boolean
}

/**
 * Particle with physics properties
 */
export interface Particle extends Graphics {
  vx: number
  vy: number
  gravity?: number
  rotationSpeed?: number
  life: number
  maxLife: number
  born: number
  delay?: number
  pulseSpeed?: number
  color?: number
}

/**
 * Coin particle (sprite-based)
 */
export interface CoinParticle extends Sprite {
  vx: number
  vy: number
  gravity: number
  rotationSpeed: number
  life: number
  born: number
}

/**
 * Shockwave effect
 */
export interface Shockwave {
  graphics: Graphics
  startTime: number
  maxRadius: number
  duration: number
  color: number
}

/**
 * Lightning bolt effect
 */
export interface LightningBolt {
  graphics: Graphics
  startX: number
  startY: number
  endX: number
  endY: number
  life: number
  born: number
}

/**
 * Screen shake state
 */
export interface ScreenShakeState {
  active: boolean
  intensity: number
  decay: number
  originalX: number
  originalY: number
}

/**
 * Canvas dimensions
 */
export interface CanvasDimensions {
  width: number
  height: number
  centerX: number
  centerY: number
}

/**
 * Animation state
 */
export interface AnimationState {
  isAnimating: boolean
  isFadingOut: boolean
  startTime: number
  canSkip: boolean
}

/**
 * Timer references for cleanup
 */
export interface TimerRefs {
  timeouts: ReturnType<typeof setTimeout>[]
  intervals: ReturnType<typeof setInterval>[]
}
