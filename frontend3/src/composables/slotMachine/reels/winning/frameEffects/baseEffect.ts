import { Container, Graphics } from 'pixi.js'
import type { Sprite } from 'pixi.js'

/**
 * Base interface for all winning frame effects
 */
export interface WinningFrameEffect {
  container: Container
  update: (timestamp: number, sprite: Sprite, x: number, y: number, isBonus: boolean, symbol?: string) => void
  show: () => void
  hide: () => void
  destroy: () => void
}

/**
 * Base class for winning frame effects
 */
export abstract class BaseFrameEffect implements WinningFrameEffect {
  container: Container
  protected width: number = 0
  protected height: number = 0
  protected isVisible: boolean = false

  constructor() {
    this.container = new Container()
    this.container.visible = false
  }

  /**
   * Update the effect (called every frame when active)
   */
  abstract update(timestamp: number, sprite: Sprite, x: number, y: number, isBonus: boolean, symbol?: string): void

  /**
   * Show the effect
   */
  show(): void {
    this.isVisible = true
    this.container.visible = true
  }

  /**
   * Hide the effect
   */
  hide(): void {
    this.isVisible = false
    this.container.visible = false
  }

  /**
   * Destroy and cleanup
   */
  destroy(): void {
    this.container.removeChildren()
    this.container.destroy({ children: true })
  }

  /**
   * Update dimensions based on sprite
   */
  protected updateDimensions(sprite: Sprite): void {
    this.width = sprite.texture.width
    this.height = sprite.texture.height
  }

  /**
   * Update container position and scale to match sprite
   */
  protected matchSpriteTransform(sprite: Sprite, x: number, y: number): void {
    this.container.x = x
    this.container.y = y
    this.container.scale.x = sprite.scale.x
    this.container.scale.y = sprite.scale.y
    this.container.alpha = sprite.alpha
    this.container.zIndex = sprite.zIndex || 0
  }
}
