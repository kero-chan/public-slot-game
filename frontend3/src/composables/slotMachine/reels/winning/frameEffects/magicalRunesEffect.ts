import { Graphics, Text } from 'pixi.js'
import type { Sprite } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import { BaseFrameEffect } from './baseEffect'
import { FRAME_EFFECT_CONFIG, WinningFrameEffectType } from '../frameEffectsConfig'

/**
 * Effect 3: Magical Runes Frame
 * - Ancient runes/symbols rotating around the tile
 * - Glowing mystical patterns that appear and fade
 * - Different symbols cycle through
 */
export class MagicalRunesEffect extends BaseFrameEffect {
  private border: Graphics
  private runes: Text[] = []
  private runeAngles: number[] = []
  private runeAlphas: number[] = []
  private startTime: number = 0
  private config = FRAME_EFFECT_CONFIG[WinningFrameEffectType.MAGICAL_RUNES]

  constructor() {
    super()

    // Create mystical border
    this.border = new Graphics()
    this.container.addChild(this.border)

    // Create rune symbols
    for (let i = 0; i < this.config.runeCount; i++) {
      const symbol = this.config.runeSymbols[i % this.config.runeSymbols.length]
      const rune = new Text(symbol, {
        fontSize: this.config.runeSize,
        fill: 0xffd700,
        fontWeight: 'bold' as const
      })
      rune.anchor.set(0.5)
      rune.blendMode = BLEND_MODES.ADD as any
      this.container.addChild(rune)
      this.runes.push(rune)
      this.runeAngles.push((i / this.config.runeCount) * Math.PI * 2)
      this.runeAlphas.push(Math.random())
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

    // Clear border
    this.border.clear()

    // Draw mystical border with pulse
    const elapsed = (timestamp - this.startTime) / 1000
    const pulse = Math.sin(elapsed * 2) * 0.3 + 0.7

    const offsetX = -w / 2
    const offsetY = -h / 2

    // Outer mystical glow
    const glowWidth = w * 0.06
    this.border.roundRect(
      offsetX - glowWidth,
      offsetY - glowWidth,
      w + glowWidth * 2,
      h + glowWidth * 2,
      cornerRadius + glowWidth
    )
    this.border.fill({ color: colors[0], alpha: 0.15 * pulse })

    // Inner border
    this.border.roundRect(offsetX, offsetY, w, h, cornerRadius)
    this.border.stroke({ color: colors[1], width: w * 0.02, alpha: 0.8 })

    // Update runes
    const orbitRadius = Math.max(w, h) * 0.65
    this.runes.forEach((rune, i) => {
      // Rotate around the tile
      this.runeAngles[i] += this.config.rotationSpeed

      // Position on orbit
      const angle = this.runeAngles[i]
      rune.x = Math.cos(angle) * orbitRadius
      rune.y = Math.sin(angle) * orbitRadius

      // Rotate the rune itself
      rune.rotation = angle + elapsed

      // Fade in/out animation
      this.runeAlphas[i] += this.config.fadeSpeed * 0.01
      const fade = Math.sin(this.runeAlphas[i]) * 0.5 + 0.5
      rune.alpha = fade * 0.9

      // Color
      const color = colors[i % colors.length]
      rune.style.fill = color

      // Change symbol occasionally
      if (Math.random() < 0.01) {
        const newSymbol = this.config.runeSymbols[Math.floor(Math.random() * this.config.runeSymbols.length)]
        rune.text = newSymbol
      }
    })
  }

  show(): void {
    super.show()
    this.startTime = performance.now()
  }
}
