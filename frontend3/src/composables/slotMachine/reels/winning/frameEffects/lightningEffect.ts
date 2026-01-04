import { Graphics } from 'pixi.js'
import type { Sprite } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import { BaseFrameEffect } from './baseEffect'
import { FRAME_EFFECT_CONFIG, WinningFrameEffectType } from '../frameEffectsConfig'

/**
 * Effect 2: Lightning/Electric Frame
 * OPTIMIZED: Reduced draw calls, throttled bolt regeneration
 * - Electric arcs dancing around the tile borders
 * - Crackling lightning bolts
 * - Random flashing effect
 */
export class LightningEffect extends BaseFrameEffect {
  private border: Graphics
  private bolts: Graphics[] = []
  private boltPaths: { points: { x: number; y: number }[]; alpha: number }[] = []
  private flashTimer: number = 0
  private config = FRAME_EFFECT_CONFIG[WinningFrameEffectType.LIGHTNING]
  private lastBoltUpdate: number = 0
  private lastBorderDimensions: { w: number; h: number; isBonus: boolean } | null = null

  // OPTIMIZATION: Throttle regeneration to every ~50ms (20fps for lightning)
  private static readonly BOLT_UPDATE_INTERVAL = 50

  constructor() {
    super()

    // Create electric border
    this.border = new Graphics()
    this.container.addChild(this.border)

    // Create lightning bolts - use fewer bolts for better performance
    const boltCount = Math.min(this.config.boltCount, 4) // Cap at 4 bolts
    for (let i = 0; i < boltCount; i++) {
      const bolt = new Graphics()
      bolt.blendMode = BLEND_MODES.ADD as any
      this.container.addChild(bolt)
      this.bolts.push(bolt)
      this.boltPaths.push({ points: [], alpha: 0 })
    }
  }

  update(timestamp: number, sprite: Sprite, x: number, y: number, isBonus: boolean, _symbol?: string): void {
    if (!this.isVisible) return

    this.updateDimensions(sprite)
    this.matchSpriteTransform(sprite, x, y)

    const w = this.width
    const h = this.height
    const colors = isBonus ? this.config.colors.bonus : this.config.colors.normal

    // Update flash timer
    this.flashTimer += this.config.flashSpeed

    // OPTIMIZATION: Only redraw border if dimensions changed
    const dimensionsChanged = !this.lastBorderDimensions ||
      Math.abs(this.lastBorderDimensions.w - w) > 1 ||
      Math.abs(this.lastBorderDimensions.h - h) > 1 ||
      this.lastBorderDimensions.isBonus !== isBonus

    if (dimensionsChanged) {
      this.drawBorder(w, h, colors)
      this.lastBorderDimensions = { w, h, isBonus }
    }

    // OPTIMIZATION: Animate border alpha without redrawing
    const pulse = Math.sin(this.flashTimer * 2) * 0.3 + 0.7
    this.border.alpha = pulse

    // OPTIMIZATION: Only regenerate bolts every BOLT_UPDATE_INTERVAL ms
    const now = performance.now()
    if (now - this.lastBoltUpdate > LightningEffect.BOLT_UPDATE_INTERVAL) {
      this.regenerateBolts(w, h)
      this.drawBolts(colors)
      this.lastBoltUpdate = now
    }

    // OPTIMIZATION: Update bolt alpha without redrawing
    this.bolts.forEach((bolt, i) => {
      const path = this.boltPaths[i]
      const flash = Math.sin(this.flashTimer * 3 + i * 1.2) * 0.3 + 0.7
      const intenseBurst = Math.random() < 0.05 ? 1.3 : 1.0 // Reduced frequency
      bolt.alpha = flash * intenseBurst * path.alpha
    })
  }

  /**
   * Draw the static border (only when dimensions change)
   */
  private drawBorder(w: number, h: number, colors: number[]): void {
    this.border.clear()

    const cornerRadius = w * 0.15
    const offsetX = -w / 2
    const offsetY = -h / 2
    const glowWidth = w * 0.08

    // OPTIMIZATION: Reduced from 3 layers to 2 for better performance
    this.border.roundRect(
      offsetX - glowWidth,
      offsetY - glowWidth,
      w + glowWidth * 2,
      h + glowWidth * 2,
      cornerRadius + glowWidth
    )
    this.border.fill({ color: colors[0], alpha: 0.12 })

    this.border.roundRect(
      offsetX - glowWidth * 0.5,
      offsetY - glowWidth * 0.5,
      w + glowWidth,
      h + glowWidth,
      cornerRadius + glowWidth * 0.5
    )
    this.border.fill({ color: colors[0], alpha: 0.2 })
  }

  /**
   * Draw all lightning bolts (called at throttled intervals)
   */
  private drawBolts(colors: number[]): void {
    this.bolts.forEach((bolt, i) => {
      bolt.clear()
      const path = this.boltPaths[i]

      if (path.points.length > 1) {
        const color = colors[i % colors.length]

        // OPTIMIZATION: Reduced from 3 stroke passes to 2
        // Draw outer glow
        bolt.moveTo(path.points[0].x, path.points[0].y)
        for (let j = 1; j < path.points.length; j++) {
          bolt.lineTo(path.points[j].x, path.points[j].y)
        }
        bolt.stroke({ color, width: 16, alpha: 0.4 })

        // Draw bright white core
        bolt.moveTo(path.points[0].x, path.points[0].y)
        for (let j = 1; j < path.points.length; j++) {
          bolt.lineTo(path.points[j].x, path.points[j].y)
        }
        bolt.stroke({ color: 0xffffff, width: 4, alpha: 0.9 })

        path.alpha = 1
      }
    })
  }

  private regenerateBolts(w: number, h: number): void {
    const offsetX = -w / 2
    const offsetY = -h / 2

    // Each bolt is a complete loop around the tile perimeter with animated waves
    this.boltPaths.forEach((path, boltIndex) => {
      path.points = []

      const perimeter = (w + h) * 2
      // OPTIMIZATION: Increased segment length for fewer points (12 instead of 6)
      const segmentLength = 12
      const totalSegments = Math.floor(perimeter / segmentLength)

      // Different wave parameters for each bolt
      const waveFrequency = 0.08 + boltIndex * 0.02
      const waveSpeed = 2 + boltIndex * 0.5
      const wavePhase = this.flashTimer * waveSpeed

      // Create a seamless wavy lightning loop around the entire tile
      for (let i = 0; i <= totalSegments; i++) {
        const distance = (i / totalSegments) * perimeter
        let x: number, y: number

        // Determine position along the perimeter
        if (distance < w) {
          x = offsetX + distance
          y = offsetY
        } else if (distance < w + h) {
          x = offsetX + w
          y = offsetY + (distance - w)
        } else if (distance < w * 2 + h) {
          x = offsetX + w - (distance - w - h)
          y = offsetY + h
        } else {
          x = offsetX
          y = offsetY + h - (distance - w * 2 - h)
        }

        // Create animated wavy offset using sine waves
        const wave1 = Math.sin(i * waveFrequency + wavePhase) * 8
        const wave2 = Math.sin(i * waveFrequency * 2 - wavePhase * 1.5) * 4
        const offsetAmount = (wave1 + wave2) * this.config.arcIntensity

        // Determine which edge we're on for perpendicular offset
        if (distance < w) {
          y += offsetAmount
        } else if (distance < w + h) {
          x += offsetAmount
        } else if (distance < w * 2 + h) {
          y += offsetAmount
        } else {
          x += offsetAmount
        }

        path.points.push({ x, y })
      }
    })
  }

  show(): void {
    super.show()
    this.lastBorderDimensions = null // Force border redraw
    this.lastBoltUpdate = 0 // Force bolt regeneration
  }
}
