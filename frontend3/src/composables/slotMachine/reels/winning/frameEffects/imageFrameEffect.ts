import { Container, Sprite } from 'pixi.js'
import type { Sprite as SpriteType, Texture } from 'pixi.js'
import { BaseFrameEffect } from './baseEffect'
import { getWinningFrameForSymbol } from '@/config/spritesheet'
import gsap from 'gsap'

/**
 * Image-based winning frame effect
 * Uses pre-rendered frame images for each high-value symbol type
 */
export class ImageFrameEffect extends BaseFrameEffect {
  private frameSprite: Sprite | null = null
  private currentSymbol: string = ''
  private pulseTween: gsap.core.Tween | null = null

  constructor() {
    super()
  }

  /**
   * Update the frame effect
   * Shows the appropriate frame image based on the winning symbol
   */
  update(timestamp: number, sprite: SpriteType, x: number, y: number, isBonus: boolean, symbol?: string): void {
    if (!this.isVisible) return

    // Update position to match tile sprite
    this.matchSpriteTransform(sprite, x, y)
    this.updateDimensions(sprite)

    // Get the symbol to use for frame selection
    const frameSymbol = symbol || 'default'

    // Only update texture if symbol changed
    if (frameSymbol !== this.currentSymbol) {
      this.currentSymbol = frameSymbol
      this.updateFrameTexture(frameSymbol)
    }

    // Update frame sprite size to match tile
    if (this.frameSprite) {
      // Scale frame to cover the tile with some padding
      const padding = 0.15 // 15% larger than tile
      const targetWidth = this.width * (1 + padding)
      const targetHeight = this.height * (1 + padding)

      if (this.frameSprite.texture.width > 0) {
        this.frameSprite.scale.x = targetWidth / this.frameSprite.texture.width
        this.frameSprite.scale.y = targetHeight / this.frameSprite.texture.height
      }
    }
  }

  /**
   * Update the frame texture based on symbol
   */
  private updateFrameTexture(symbol: string): void {
    const texture = getWinningFrameForSymbol(symbol)

    if (!texture) {
      console.warn(`[ImageFrameEffect] No texture found for symbol: ${symbol}`)
      return
    }

    if (!this.frameSprite) {
      this.frameSprite = new Sprite(texture)
      this.frameSprite.anchor.set(0.5, 0.5)
      this.container.addChild(this.frameSprite)
    } else {
      this.frameSprite.texture = texture
    }
  }

  /**
   * Show the effect with a pulse animation
   */
  show(): void {
    super.show()

    // Start pulse animation
    if (this.frameSprite && !this.pulseTween) {
      this.pulseTween = gsap.to(this.frameSprite, {
        alpha: 0.7,
        duration: 0.5,
        repeat: -1,
        yoyo: true,
        ease: 'sine.inOut'
      })
    }
  }

  /**
   * Hide the effect
   */
  hide(): void {
    super.hide()

    // Stop pulse animation
    if (this.pulseTween) {
      this.pulseTween.kill()
      this.pulseTween = null
    }

    if (this.frameSprite) {
      this.frameSprite.alpha = 1
    }
  }

  /**
   * Destroy and cleanup
   */
  destroy(): void {
    if (this.pulseTween) {
      this.pulseTween.kill()
      this.pulseTween = null
    }

    if (this.frameSprite) {
      this.frameSprite.destroy()
      this.frameSprite = null
    }

    this.currentSymbol = ''
    super.destroy()
  }
}
