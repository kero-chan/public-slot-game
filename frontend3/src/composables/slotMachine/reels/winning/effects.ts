import { createParticleSystem, type ParticleSystem } from './particles'
import type { Sprite } from 'pixi.js'
import type { WinCombination } from '@/features/spin/types'

type EffectIntensity = 'small' | 'medium' | 'big' | 'mega'
type EffectPhase = 'highlight' | 'celebrate' | 'idle'

interface EffectParams {
  glowIntensity: number
  pulseSpeed: number
  pulseScale: number
  rotation: number
  particleCount: number
  particleType: 'sparkle' | 'star'
  flashCount: number
  enableOrbiting: boolean
  orbitStars?: number
}

interface Effect {
  sprite: Sprite
  params: EffectParams
  intensity: EffectIntensity
  startTime: number
  phase: EffectPhase
  originalWidth: number
  originalHeight: number
  originalRotation: number
  flashCounter: number
  burstTriggered: boolean
  orbitTriggered: boolean
  celebrateStartTime?: number
}

export interface WinningEffects {
  container: ParticleSystem['container']
  startEffects: (wins: WinCombination[], tileSprites: Map<string, Sprite>) => void
  update: (deltaTime?: number, tileSprites?: Map<string, Sprite>) => void
  clear: () => void
  isActive: () => boolean
  getEffectIntensity: (win: WinCombination) => EffectIntensity
}

export function createWinningEffects(): WinningEffects {
  const particleSystem = createParticleSystem()
  const activeEffects = new Map<string, Effect>()
  let animationTime = 0

  function getEffectIntensity(win: WinCombination): EffectIntensity {
    // Use backend-provided win_intensity
    return win.win_intensity
  }

  function getEffectParams(intensity: EffectIntensity): EffectParams {
    const params: Record<EffectIntensity, EffectParams> = {
      small: { glowIntensity: 0.3, pulseSpeed: 2, pulseScale: 1.1, rotation: 0.05, particleCount: 10, particleType: 'sparkle', flashCount: 2, enableOrbiting: false },
      medium: { glowIntensity: 0.5, pulseSpeed: 2.5, pulseScale: 1.15, rotation: 0.08, particleCount: 15, particleType: 'sparkle', flashCount: 3, enableOrbiting: false },
      big: { glowIntensity: 0.7, pulseSpeed: 3, pulseScale: 1.2, rotation: 0.1, particleCount: 20, particleType: 'star', flashCount: 4, enableOrbiting: true, orbitStars: 5 },
      mega: { glowIntensity: 1.0, pulseSpeed: 4, pulseScale: 1.3, rotation: 0.15, particleCount: 30, particleType: 'star', flashCount: 5, enableOrbiting: true, orbitStars: 8 }
    }
    return params[intensity] || params.small
  }

  function startEffects(wins: WinCombination[], tileSprites: Map<string, Sprite>): void {
    if (!wins || wins.length === 0) return
    animationTime = 0
    activeEffects.clear()

    wins.forEach(win => {
      const intensity = getEffectIntensity(win)
      const params = getEffectParams(intensity)

      win.positions.forEach(pos => {
        const col = pos.reel
        const row = pos.row
        const key = `${col}:${row}`
        const sprite = tileSprites.get(key)

        if (sprite) {
          activeEffects.set(key, {
            sprite, params, intensity,
            startTime: Date.now(),
            phase: 'highlight',
            originalWidth: sprite.width,
            originalHeight: sprite.height,
            originalRotation: sprite.rotation || 0,
            flashCounter: 0,
            burstTriggered: false,
            orbitTriggered: false
          })
        }
      })
    })
  }

  function applyEffectToSprite(effect: Effect, deltaTime: number): void {
    const { sprite, params, phase, startTime } = effect
    const elapsed = (Date.now() - startTime) / 1000

    if (!sprite || sprite.destroyed) return

    if (elapsed < 0.5) {
      effect.phase = 'highlight'
      const glowPulse = 0.5 + Math.sin(animationTime * params.pulseSpeed) * 0.5
      sprite.tint = 0xffffaa + Math.floor(glowPulse * 0x55)
      const scaleProgress = Math.min(elapsed / 0.5, 1)
      const targetScale = 1 + (params.pulseScale - 1) * scaleProgress
      sprite.width = effect.originalWidth * targetScale
      sprite.height = effect.originalHeight * targetScale
      sprite.rotation = effect.originalRotation + Math.sin(animationTime * 2) * params.rotation * 0.5
    } else if (elapsed < 2.0) {
      if (effect.phase !== 'celebrate') {
        effect.phase = 'celebrate'
        effect.celebrateStartTime = Date.now()
      }
      const celebrateElapsed = (Date.now() - effect.celebrateStartTime!) / 1000
      const glowPulse = 0.5 + Math.sin(animationTime * params.pulseSpeed * 1.5) * 0.5
      sprite.tint = 0xffff88 + Math.floor(glowPulse * 0x77)
      const scalePulse = 1 + Math.sin(animationTime * params.pulseSpeed) * (params.pulseScale - 1)
      sprite.width = effect.originalWidth * scalePulse
      sprite.height = effect.originalHeight * scalePulse
      sprite.rotation = effect.originalRotation + Math.sin(animationTime * params.pulseSpeed * 2) * params.rotation

      if (!effect.burstTriggered) {
        effect.burstTriggered = true
        const centerX = sprite.x + sprite.width / 2
        const centerY = sprite.y + sprite.height / 2
        particleSystem.burst(centerX, centerY, params.particleCount, params.particleType)
      }

      if (Math.random() < 0.15) {
        const centerX = sprite.x + sprite.width / 2
        const centerY = sprite.y + sprite.height / 2
        particleSystem.sparkle(centerX, centerY, 2)
      }

      if (params.enableOrbiting && !effect.orbitTriggered && celebrateElapsed > 0.3 && params.orbitStars) {
        effect.orbitTriggered = true
        const centerX = sprite.x + sprite.width / 2
        const centerY = sprite.y + sprite.height / 2
        particleSystem.orbitingStars(centerX, centerY, params.orbitStars, sprite.width * 0.4)
      }

      const flashInterval = 0.2
      const flashIndex = Math.floor(celebrateElapsed / flashInterval)
      if (flashIndex !== effect.flashCounter && flashIndex < params.flashCount) {
        effect.flashCounter = flashIndex
        sprite.tint = 0xffffff
        setTimeout(() => {
          if (sprite && !sprite.destroyed) sprite.tint = 0xffffcc
        }, 50)
      }
    } else {
      effect.phase = 'idle'
      const fadeElapsed = elapsed - 2.0
      const fadeProgress = Math.min(fadeElapsed / 0.5, 1)
      const targetScale = params.pulseScale - (params.pulseScale - 1) * fadeProgress
      sprite.width = effect.originalWidth * targetScale
      sprite.height = effect.originalHeight * targetScale
      const rotation = params.rotation * (1 - fadeProgress)
      sprite.rotation = effect.originalRotation + Math.sin(animationTime * 2) * rotation
      const tintProgress = 1 - fadeProgress
      sprite.tint = 0xffffff + Math.floor(tintProgress * 0xff) * 0x010100
    }
  }

  function update(deltaTime: number = 1, tileSprites?: Map<string, Sprite>): void {
    animationTime += deltaTime / 60
    activeEffects.forEach((effect, key) => {
      if (effect.sprite && !effect.sprite.destroyed) {
        applyEffectToSprite(effect, deltaTime)
      } else {
        activeEffects.delete(key)
      }
    })
    particleSystem.update(deltaTime)
  }

  function clear(): void {
    activeEffects.forEach((effect) => {
      if (effect.sprite && !effect.sprite.destroyed) {
        effect.sprite.width = effect.originalWidth
        effect.sprite.height = effect.originalHeight
        effect.sprite.rotation = effect.originalRotation
        effect.sprite.tint = 0xffffff
      }
    })
    activeEffects.clear()
    particleSystem.clear()
  }

  function isActive(): boolean {
    return activeEffects.size > 0
  }

  return { container: particleSystem.container, startEffects, update, clear, isActive, getEffectIntensity }
}
