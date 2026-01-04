import { Container, Graphics } from 'pixi.js'
import type { Particle, LightningRays, ScalableSprite } from './types'
import { SPIN_BUTTON_EFFECT_DURATION, SPIN_BUTTON_EFFECT_FADEOUT } from './types'

export interface ParticleState {
  spinParticlesContainer: Container | null
  spinParticles: Particle[]
  spinLightningContainer: Container | null
  lightningBolts: LightningRays[]
  jackpotParticlesContainer: Container | null
  jackpotParticles: Particle[]
  lightningStartTime: number
  wasSpinning: boolean
}

export function createParticleState(): ParticleState {
  return {
    spinParticlesContainer: null,
    spinParticles: [],
    spinLightningContainer: null,
    lightningBolts: [],
    jackpotParticlesContainer: null,
    jackpotParticles: [],
    lightningStartTime: 0,
    wasSpinning: false
  }
}

export function spawnSpinButtonParticle(
  state: ParticleState,
  spinBtnSprite: ScalableSprite
): void {
  if (!state.spinParticlesContainer || !spinBtnSprite) return

  const particle = new Graphics() as Particle
  const size = 3 + Math.random() * 5
  particle.circle(0, 0, size)
  particle.fill({ color: 0xffd700, alpha: 0.9 })
  particle.blendMode = 'add'

  const angle = Math.random() * Math.PI * 2
  const radius = spinBtnSprite.height * 0.5
  particle.x = Math.cos(angle) * radius
  particle.y = Math.sin(angle) * radius

  particle.vx = (Math.random() - 0.5) * 0.3
  particle.vy = -0.3 - Math.random() * 0.3
  particle.life = 120
  particle.maxLife = 120

  state.spinParticlesContainer.addChild(particle)
  state.spinParticles.push(particle)
}

export function updateSpinButtonParticles(state: ParticleState): void {
  if (!state.spinParticlesContainer) return

  for (let i = state.spinParticles.length - 1; i >= 0; i--) {
    const p = state.spinParticles[i]
    p.life--

    p.x += p.vx
    p.y += p.vy

    const lifePercent = p.life / p.maxLife
    p.alpha = lifePercent * 0.9

    if (p.life <= 0) {
      state.spinParticlesContainer.removeChild(p)
      p.destroy()
      state.spinParticles.splice(i, 1)
    }
  }
}

export function spawnJackpotParticle(
  state: ParticleState,
  footerWidth: number
): void {
  if (!state.jackpotParticlesContainer) return

  const particle = new Graphics() as Particle
  const size = 3 + Math.random() * 6
  particle.circle(0, 0, size)
  particle.fill({ color: 0xffd700, alpha: 0.8 })
  particle.blendMode = 'add'

  particle.x = footerWidth / 2 + (Math.random() - 0.5) * 200
  particle.y = 50 + Math.random() * 10

  particle.vx = (Math.random() - 0.5) * 3.5
  particle.vy = 0.5 + Math.random() * 1.2
  particle.life = 150 + Math.random() * 100
  particle.maxLife = particle.life

  state.jackpotParticlesContainer.addChild(particle)
  state.jackpotParticles.push(particle)
}

export function updateJackpotParticles(
  state: ParticleState,
  isFreeSpinMode: boolean,
  footerWidth: number,
  footerHeight: number
): void {
  if (!state.jackpotParticlesContainer) return

  if (isFreeSpinMode && Math.random() < 0.15) {
    spawnJackpotParticle(state, footerWidth)
  }

  for (let i = state.jackpotParticles.length - 1; i >= 0; i--) {
    const p = state.jackpotParticles[i]
    p.life--

    p.x += p.vx + Math.sin(p.life * 0.05) * 0.2
    p.y += p.vy

    const lifePercent = p.life / p.maxLife
    p.alpha = lifePercent * 0.8

    if (p.life <= 0 || p.y > footerHeight + 10) {
      state.jackpotParticlesContainer.removeChild(p)
      p.destroy()
      state.jackpotParticles.splice(i, 1)
    }
  }

  if (!isFreeSpinMode && state.jackpotParticles.length > 0) {
    for (const p of state.jackpotParticles) {
      state.jackpotParticlesContainer.removeChild(p)
      p.destroy()
    }
    state.jackpotParticles.length = 0
  }
}

export function createLightRays(spinBtnSprite: ScalableSprite): LightningRays | null {
  if (!spinBtnSprite) return null

  const rays = new Graphics() as LightningRays
  rays.blendMode = 'add'

  const numRays = 1000
  const innerRadius = spinBtnSprite.height * 0.45
  const outerRadius = spinBtnSprite.height * 0.65

  for (let i = 0; i < numRays; i++) {
    const angle = (i / numRays) * Math.PI * 2

    const innerX1 = Math.cos(angle - 0.02) * innerRadius
    const innerY1 = Math.sin(angle - 0.02) * innerRadius
    const innerX2 = Math.cos(angle + 0.02) * innerRadius
    const innerY2 = Math.sin(angle + 0.02) * innerRadius
    const outerX = Math.cos(angle) * outerRadius
    const outerY = Math.sin(angle) * outerRadius

    rays.poly([
      { x: innerX1 * 1.2, y: innerY1 * 1.2 },
      { x: innerX2 * 1.2, y: innerY2 * 1.2 },
      { x: outerX, y: outerY }
    ])
    rays.fill({ color: 0xffeb3b, alpha: 0.02 })

    rays.poly([
      { x: innerX1, y: innerY1 },
      { x: innerX2, y: innerY2 },
      { x: outerX * 0.9, y: outerY * 0.9 }
    ])
    rays.fill({ color: 0xffd700, alpha: 0.04 })

    rays.poly([
      { x: innerX1 * 0.8, y: innerY1 * 0.8 },
      { x: innerX2 * 0.8, y: innerY2 * 0.8 },
      { x: outerX * 0.8, y: outerY * 0.8 }
    ])
    rays.fill({ color: 0xffffff, alpha: 0.05 })
  }

  rays.rotation = 0
  return rays
}

export function updateLightning(
  state: ParticleState,
  isSpinning: boolean,
  spinBtnSprite: ScalableSprite,
  timestamp: number = 0
): void {
  if (!state.spinLightningContainer) return

  // Use passed timestamp instead of Date.now() for better performance
  const now = timestamp || performance.now()

  if (isSpinning && !state.wasSpinning) {
    state.lightningStartTime = now
  }
  state.wasSpinning = isSpinning

  const elapsedSeconds = (now - state.lightningStartTime) / 1000

  if (isSpinning) {
    if (elapsedSeconds < SPIN_BUTTON_EFFECT_DURATION) {
      if (state.lightningBolts.length === 0) {
        const rays = createLightRays(spinBtnSprite)
        if (rays) {
          state.spinLightningContainer.addChild(rays)
          state.lightningBolts.push(rays)
        }
      } else {
        for (const rays of state.lightningBolts) {
          rays.rotation += 0.02
          rays.alpha = 1
        }
      }
    } else if (elapsedSeconds < SPIN_BUTTON_EFFECT_DURATION + SPIN_BUTTON_EFFECT_FADEOUT) {
      const fadeProgress = (elapsedSeconds - SPIN_BUTTON_EFFECT_DURATION) / SPIN_BUTTON_EFFECT_FADEOUT
      for (const rays of state.lightningBolts) {
        rays.rotation += 0.02
        rays.alpha = 1 - fadeProgress
      }
    } else {
      clearLightningBolts(state)
    }
  } else {
    clearLightningBolts(state)
  }
}

function clearLightningBolts(state: ParticleState): void {
  for (const bolt of state.lightningBolts) {
    state.spinLightningContainer?.removeChild(bolt)
    bolt.destroy()
  }
  state.lightningBolts.length = 0
}
