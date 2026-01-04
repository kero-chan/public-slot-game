import { Graphics } from 'pixi.js'
import type { Sprite } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import { BaseFrameEffect } from './baseEffect'
import { FRAME_EFFECT_CONFIG, WinningFrameEffectType } from '../frameEffectsConfig'

/**
 * Effect 1: Glowing Energy Border with Orbiting Particles
 * OPTIMIZED: Reduced redraws, uses alpha/scale animation
 * - Pulsing glow border
 * - Energy particles orbiting around the tile
 * - Color-coded for normal/bonus tiles
 */
export class GlowingEnergyEffect extends BaseFrameEffect {
  private glowBorder: Graphics
  private particles: Graphics[] = []
  private particleAngles: number[] = []
  private startTime: number = 0
  private config = FRAME_EFFECT_CONFIG[WinningFrameEffectType.GLOWING_ENERGY]
  private lastBorderDimensions: { w: number; h: number; isBonus: boolean } | null = null
  private lastParticleUpdate: number = 0

  // OPTIMIZATION: Throttle particle position updates
  private static readonly PARTICLE_UPDATE_INTERVAL = 33 // ~30fps for particles

  constructor() {
    super()

    // Create glow border
    this.glowBorder = new Graphics()
    this.container.addChild(this.glowBorder)

    // OPTIMIZATION: Use fewer particles (cap at 8)
    const particleCount = Math.min(this.config.particleCount, 8)
    for (let i = 0; i < particleCount; i++) {
      const particle = new Graphics()
      particle.blendMode = BLEND_MODES.ADD as any
      this.container.addChild(particle)
      this.particles.push(particle)
      this.particleAngles.push((i / particleCount) * Math.PI * 2)
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

    // OPTIMIZATION: Only redraw border when dimensions change
    const dimensionsChanged = !this.lastBorderDimensions ||
      Math.abs(this.lastBorderDimensions.w - w) > 1 ||
      Math.abs(this.lastBorderDimensions.h - h) > 1 ||
      this.lastBorderDimensions.isBonus !== isBonus

    if (dimensionsChanged) {
      this.drawBorder(w, h, colors)
      this.lastBorderDimensions = { w, h, isBonus }
    }

    // OPTIMIZATION: Animate pulse using alpha (no redraw)
    const elapsed = (timestamp - this.startTime) / 1000
    const pulse = Math.sin(elapsed * this.config.pulseSpeed) * 0.3 + 0.7
    this.glowBorder.alpha = pulse

    // OPTIMIZATION: Throttle particle position updates
    const now = performance.now()
    if (now - this.lastParticleUpdate > GlowingEnergyEffect.PARTICLE_UPDATE_INTERVAL) {
      this.updateParticles(w, h, colors, pulse)
      this.lastParticleUpdate = now
    } else {
      // Still update particle alpha for smooth pulsing
      this.particles.forEach(particle => {
        particle.alpha = 0.6 + pulse * 0.4
      })
    }
  }

  /**
   * Draw the static border (only when dimensions change)
   */
  private drawBorder(w: number, h: number, colors: number[]): void {
    this.glowBorder.clear()

    const cornerRadius = w * 0.15
    const glowWidth = w * 0.08  // Static width, pulse via alpha
    const offsetX = -w / 2
    const offsetY = -h / 2
    const borderColor = colors[0]
    const glowColor = colors[1]

    // Outer glow
    this.glowBorder.roundRect(
      offsetX - glowWidth,
      offsetY - glowWidth,
      w + glowWidth * 2,
      h + glowWidth * 2,
      cornerRadius + glowWidth
    )
    this.glowBorder.fill({ color: glowColor, alpha: 0.1 })

    // Mid glow
    const midGlow = glowWidth * 0.6
    this.glowBorder.roundRect(
      offsetX - midGlow,
      offsetY - midGlow,
      w + midGlow * 2,
      h + midGlow * 2,
      cornerRadius + midGlow
    )
    this.glowBorder.fill({ color: glowColor, alpha: 0.2 })

    // Inner glow
    const innerGlow = glowWidth * 0.3
    this.glowBorder.roundRect(
      offsetX - innerGlow,
      offsetY - innerGlow,
      w + innerGlow * 2,
      h + innerGlow * 2,
      cornerRadius + innerGlow
    )
    this.glowBorder.fill({ color: borderColor, alpha: 0.3 })

    // Main border
    this.glowBorder.roundRect(offsetX, offsetY, w, h, cornerRadius)
    this.glowBorder.stroke({ color: borderColor, width: w * 0.025, alpha: 1.0 })
  }

  /**
   * Update particle positions (called at throttled intervals)
   */
  private updateParticles(w: number, h: number, colors: number[], pulse: number): void {
    const orbitRadius = Math.max(w, h) * 0.6

    this.particles.forEach((particle, i) => {
      particle.clear()

      // Update angle for orbit
      this.particleAngles[i] += this.config.orbitSpeed

      // Calculate position on orbit path
      const angle = this.particleAngles[i]
      const particleX = Math.cos(angle) * orbitRadius
      const particleY = Math.sin(angle) * orbitRadius

      const size = this.config.particleSize
      const particleColor = colors[i % colors.length]

      // Draw glowing particle (simplified - only 2 layers instead of redrawing with pulse)
      particle.circle(particleX, particleY, size * 1.5)
      particle.fill({ color: particleColor, alpha: 0.3 })

      particle.circle(particleX, particleY, size)
      particle.fill({ color: particleColor, alpha: 0.8 })
    })
  }

  show(): void {
    super.show()
    this.startTime = performance.now()
    this.lastBorderDimensions = null // Force border redraw
    this.lastParticleUpdate = 0
  }
}
