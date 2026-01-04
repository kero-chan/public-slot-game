import { Graphics } from 'pixi.js'
import type { Sprite } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import { BaseFrameEffect } from './baseEffect'
import { FRAME_EFFECT_CONFIG, WinningFrameEffectType } from '../frameEffectsConfig'

interface Wave {
  radius: number
  alpha: number
  speed: number
  color: number
  startTime: number
}

/**
 * Effect 4: Expanding Ring Waves
 * - Shockwave rings that expand outward from tile center
 * - Multiple waves with different colors/speeds
 * - Creates a "ripple in water" effect
 */
export class RingWavesEffect extends BaseFrameEffect {
  private border: Graphics
  private waves: Wave[] = []
  private waveGraphics: Graphics[] = []
  private lastWaveTime: number = 0
  private startTime: number = 0
  private config = FRAME_EFFECT_CONFIG[WinningFrameEffectType.RING_WAVES]

  constructor() {
    super()

    // Create border
    this.border = new Graphics()
    this.container.addChild(this.border)

    // Create wave graphics
    for (let i = 0; i < this.config.waveCount; i++) {
      const waveGraphic = new Graphics()
      waveGraphic.blendMode = BLEND_MODES.ADD as any
      this.container.addChild(waveGraphic)
      this.waveGraphics.push(waveGraphic)
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

    // Draw simple border
    const offsetX = -w / 2
    const offsetY = -h / 2

    this.border.roundRect(offsetX, offsetY, w, h, cornerRadius)
    this.border.stroke({ color: colors[0], width: w * 0.02, alpha: 0.6 })

    // Spawn new waves periodically
    if (elapsed - this.lastWaveTime > this.config.waveInterval) {
      this.lastWaveTime = elapsed
      const color = colors[this.waves.length % colors.length]
      this.waves.push({
        radius: 0,
        alpha: 1,
        speed: this.config.waveSpeed,
        color,
        startTime: elapsed
      })

      // Remove old waves
      if (this.waves.length > this.config.waveCount) {
        this.waves.shift()
      }
    }

    // Update and draw waves
    this.waves.forEach((wave, i) => {
      const waveGraphic = this.waveGraphics[i]
      if (!waveGraphic) return

      waveGraphic.clear()

      // Expand radius
      const waveElapsed = elapsed - wave.startTime
      wave.radius = waveElapsed * wave.speed

      // Fade out as it expands
      const maxRadius = Math.max(w, h) * 0.8
      const progress = Math.min(wave.radius / maxRadius, 1)
      wave.alpha = 1 - progress

      if (wave.alpha > 0 && wave.radius > 0) {
        // Draw expanding ring
        waveGraphic.circle(0, 0, wave.radius)
        waveGraphic.stroke({
          color: wave.color,
          width: 4,
          alpha: wave.alpha * 0.8
        })

        // Draw outer glow
        waveGraphic.circle(0, 0, wave.radius)
        waveGraphic.stroke({
          color: wave.color,
          width: 12,
          alpha: wave.alpha * 0.3
        })
      }
    })

    // Remove fully faded waves
    this.waves = this.waves.filter(w => w.alpha > 0.01)
  }

  show(): void {
    super.show()
    this.startTime = performance.now()
    this.lastWaveTime = 0
    this.waves = []
  }
}
