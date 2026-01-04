import type { Container } from 'pixi.js'
import type { ScreenShakeState } from './types'

/**
 * Screen Shake Manager - Handles screen shake effects
 */
export class ScreenShake {
  private container: Container
  private state: ScreenShakeState = {
    active: false,
    intensity: 0,
    decay: 0.95,
    originalX: 0,
    originalY: 0
  }
  private timeout: ReturnType<typeof setTimeout> | null = null

  constructor(container: Container) {
    this.container = container
  }

  /**
   * Start screen shake effect
   */
  start(intensity = 15, duration = 500, decay = 0.95): void {
    this.state.active = true
    this.state.intensity = intensity
    this.state.decay = decay
    this.state.originalX = this.container.x
    this.state.originalY = this.container.y

    // Clear any existing timeout
    if (this.timeout) {
      clearTimeout(this.timeout)
    }

    // Auto-stop after duration
    this.timeout = setTimeout(() => {
      this.stop()
    }, duration)
  }

  /**
   * Stop screen shake and reset position
   */
  stop(): void {
    this.state.active = false
    this.container.x = this.state.originalX
    this.container.y = this.state.originalY

    if (this.timeout) {
      clearTimeout(this.timeout)
      this.timeout = null
    }
  }

  /**
   * Update shake - call this every frame
   */
  update(): void {
    if (!this.state.active) return

    // Apply random offset
    this.container.x = this.state.originalX + (Math.random() - 0.5) * this.state.intensity * 2
    this.container.y = this.state.originalY + (Math.random() - 0.5) * this.state.intensity * 2

    // Decay intensity
    this.state.intensity *= this.state.decay
  }

  /**
   * Check if shake is active
   */
  isActive(): boolean {
    return this.state.active
  }

  /**
   * Clean up
   */
  destroy(): void {
    this.stop()
  }
}

/**
 * Create a screen shake instance
 */
export function createScreenShake(container: Container): ScreenShake {
  return new ScreenShake(container)
}
