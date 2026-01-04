import { Graphics } from 'pixi.js'
import type { Sprite } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import { BaseFrameEffect } from './baseEffect'
import { FRAME_EFFECT_CONFIG, WinningFrameEffectType } from '../frameEffectsConfig'

interface Ember {
  x: number
  y: number
  vy: number
  alpha: number
  size: number
  color: number
}

/**
 * Effect 5: Dragon Scales/Fire Frame
 * - Dragon scale pattern that shimmers around the border
 * - Small fire particles/embers floating upward
 * - Fits with Asian/dragon theme
 */
export class DragonScalesEffect extends BaseFrameEffect {
  private scales: Graphics[] = []
  private scalePositions: { x: number; y: number; angle: number }[] = []
  private embers: Ember[] = []
  private startTime: number = 0
  private config = FRAME_EFFECT_CONFIG[WinningFrameEffectType.DRAGON_SCALES]

  constructor() {
    super()

    // Create dragon scales
    for (let i = 0; i < this.config.scaleCount; i++) {
      const scale = new Graphics()
      this.container.addChild(scale)
      this.scales.push(scale)
      this.scalePositions.push({ x: 0, y: 0, angle: 0 })
    }

    this.startTime = performance.now()
  }

  update(timestamp: number, sprite: Sprite, x: number, y: number, isBonus: boolean, _symbol?: string): void {
    if (!this.isVisible) return

    this.updateDimensions(sprite)
    this.matchSpriteTransform(sprite, x, y)

    const w = this.width
    const h = this.height
    const colors = isBonus ? this.config.colors.bonus : this.config.colors.normal
    const elapsed = (timestamp - this.startTime) / 1000

    // Calculate border path for scales
    this.updateScalePositions(w, h)

    // Shimmer effect
    const shimmer = Math.sin(elapsed * this.config.shimmerSpeed) * 0.5 + 0.5

    // Draw scales
    this.scales.forEach((scale, i) => {
      scale.clear()
      const pos = this.scalePositions[i]

      // Alternate colors for shimmer
      const colorIndex = Math.floor((i + shimmer * 2) % colors.length)
      const color = colors[colorIndex]

      // Draw scale shape (rounded triangle/arc)
      scale.x = pos.x
      scale.y = pos.y
      scale.rotation = pos.angle

      const scaleSize = this.config.scaleSize

      // Scale body (semicircle)
      scale.arc(0, 0, scaleSize, 0, Math.PI)
      scale.fill({ color, alpha: 0.7 + shimmer * 0.3 })

      // Scale edge highlight
      scale.arc(0, 0, scaleSize, 0, Math.PI)
      scale.stroke({ color: 0xffffff, width: 1, alpha: 0.5 * shimmer })
    })

    // Spawn embers occasionally
    if (Math.random() < 0.1) {
      const offsetX = -w / 2
      const offsetY = -h / 2
      const emberX = offsetX + Math.random() * w
      const emberY = offsetY + h // Start from bottom

      this.embers.push({
        x: emberX,
        y: emberY,
        vy: -0.5 - Math.random() * 1.5, // Float upward
        alpha: 1,
        size: 2 + Math.random() * 4,
        color: colors[Math.floor(Math.random() * colors.length)]
      })
    }

    // Update and draw embers
    const emberGraphics = new Graphics()
    emberGraphics.blendMode = BLEND_MODES.ADD as any
    this.container.addChild(emberGraphics)

    this.embers.forEach((ember, i) => {
      ember.y += ember.vy
      ember.alpha -= 0.01
      ember.size *= 0.98

      if (ember.alpha > 0) {
        // Draw ember
        emberGraphics.circle(ember.x, ember.y, ember.size)
        emberGraphics.fill({ color: ember.color, alpha: ember.alpha })

        // Glow
        emberGraphics.circle(ember.x, ember.y, ember.size * 2)
        emberGraphics.fill({ color: ember.color, alpha: ember.alpha * 0.3 })
      }
    })

    // Remove dead embers
    this.embers = this.embers.filter(e => e.alpha > 0)

    // Limit ember count
    if (this.embers.length > this.config.emberCount) {
      this.embers.shift()
    }
  }

  private updateScalePositions(w: number, h: number): void {
    const offsetX = -w / 2
    const offsetY = -h / 2
    const perimeter = (w + h) * 2
    const spacing = perimeter / this.config.scaleCount

    this.scalePositions.forEach((pos, i) => {
      const distance = i * spacing
      let x: number, y: number, angle: number

      if (distance < w) {
        // Top edge
        x = offsetX + distance
        y = offsetY
        angle = 0
      } else if (distance < w + h) {
        // Right edge
        x = offsetX + w
        y = offsetY + (distance - w)
        angle = Math.PI / 2
      } else if (distance < w * 2 + h) {
        // Bottom edge
        x = offsetX + w - (distance - w - h)
        y = offsetY + h
        angle = Math.PI
      } else {
        // Left edge
        x = offsetX
        y = offsetY + h - (distance - w * 2 - h)
        angle = -Math.PI / 2
      }

      pos.x = x
      pos.y = y
      pos.angle = angle
    })
  }

  show(): void {
    super.show()
    this.startTime = performance.now()
    this.embers = []
  }
}
