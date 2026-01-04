import { Graphics } from 'pixi.js'
import type { Sprite } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import { BaseFrameEffect } from './baseEffect'
import { FRAME_EFFECT_CONFIG, WinningFrameEffectType } from '../frameEffectsConfig'

interface Sparkle {
  x: number
  y: number
  alpha: number
  life: number
  size: number
}

/**
 * Effect 6: Crystalline Shard Frame
 * - Crystal shards that rotate and orbit around the tile
 * - Reflects light with sparkles
 * - Ties into crystal theme
 */
export class CrystallineShardsEffect extends BaseFrameEffect {
  private border: Graphics
  private shards: Graphics[] = []
  private shardAngles: number[] = []
  private sparkles: Sparkle[] = []
  private startTime: number = 0
  private config = FRAME_EFFECT_CONFIG[WinningFrameEffectType.CRYSTALLINE_SHARDS]

  constructor() {
    super()

    // Create crystal border
    this.border = new Graphics()
    this.container.addChild(this.border)

    // Create crystal shards
    for (let i = 0; i < this.config.shardCount; i++) {
      const shard = new Graphics()
      shard.blendMode = BLEND_MODES.ADD as any
      this.container.addChild(shard)
      this.shards.push(shard)
      this.shardAngles.push((i / this.config.shardCount) * Math.PI * 2)
    }

    this.startTime = performance.now()
  }

  update(timestamp: number, sprite: Sprite, x: number, y: number, isBonus: boolean, _symbol?: string): void {
    if (!this.isVisible) return

    this.updateDimensions(sprite)
    this.matchSpriteTransform(sprite, x, y)

    const w = this.width
    const h = this.height
    const cornerRadius = w * 0.15
    const colors = isBonus ? this.config.colors.bonus : this.config.colors.normal
    const elapsed = (timestamp - this.startTime) / 1000

    // Clear border
    this.border.clear()

    // Draw crystalline border with faceted look
    const offsetX = -w / 2
    const offsetY = -h / 2

    // Outer crystal glow
    const glowWidth = w * 0.06
    this.border.roundRect(
      offsetX - glowWidth,
      offsetY - glowWidth,
      w + glowWidth * 2,
      h + glowWidth * 2,
      cornerRadius + glowWidth
    )
    this.border.fill({ color: colors[0], alpha: 0.2 })

    // Main crystal border (faceted effect with multiple thin lines)
    for (let i = 0; i < 3; i++) {
      const offset = i * 2
      this.border.roundRect(
        offsetX - offset,
        offsetY - offset,
        w + offset * 2,
        h + offset * 2,
        cornerRadius + offset
      )
      this.border.stroke({ color: colors[1], width: 1, alpha: 0.6 - i * 0.2 })
    }

    // Update orbiting crystal shards
    const orbitRadius = Math.max(w, h) * 0.7
    this.shards.forEach((shard, i) => {
      shard.clear()

      // Rotate around tile
      this.shardAngles[i] += this.config.orbitSpeed

      // Position on orbit
      const angle = this.shardAngles[i]
      const shardX = Math.cos(angle) * orbitRadius
      const shardY = Math.sin(angle) * orbitRadius

      const color = colors[i % colors.length]
      const size = this.config.shardSize

      // Draw crystal shard (elongated diamond)
      const points = [
        { x: shardX, y: shardY - size },              // Top point
        { x: shardX + size * 0.4, y: shardY },        // Right
        { x: shardX, y: shardY + size },              // Bottom
        { x: shardX - size * 0.4, y: shardY }         // Left
      ]

      shard.poly(points)
      shard.fill({ color, alpha: 0.7 })
      shard.stroke({ color: 0xffffff, width: 1, alpha: 0.9 })

      // Rotate shard itself
      shard.rotation = angle + elapsed * 2

      // Spawn sparkles from shards occasionally
      if (Math.random() < 0.05) {
        this.sparkles.push({
          x: shardX + (Math.random() - 0.5) * 10,
          y: shardY + (Math.random() - 0.5) * 10,
          alpha: 1,
          life: 1,
          size: 2 + Math.random() * 3
        })
      }
    })

    // Update and draw sparkles
    const sparkleGraphics = new Graphics()
    sparkleGraphics.blendMode = BLEND_MODES.ADD as any
    this.container.addChild(sparkleGraphics)

    this.sparkles.forEach(sparkle => {
      sparkle.life -= 0.02
      sparkle.alpha = sparkle.life

      if (sparkle.alpha > 0) {
        // Draw sparkle as a star
        sparkleGraphics.star(sparkle.x, sparkle.y, 4, sparkle.size, sparkle.size * 0.5)
        sparkleGraphics.fill({ color: 0xffffff, alpha: sparkle.alpha })
      }
    })

    // Remove dead sparkles
    this.sparkles = this.sparkles.filter(s => s.alpha > 0)

    // Limit sparkle count
    if (this.sparkles.length > this.config.sparkleCount) {
      this.sparkles.shift()
    }
  }

  show(): void {
    super.show()
    this.startTime = performance.now()
    this.sparkles = []
  }
}
