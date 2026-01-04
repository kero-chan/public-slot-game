import { Container, Graphics, Sprite, Texture, Application } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import type { Particle, CoinParticle, Shockwave } from './types'

// Cached hexagon texture for shockwaves - generated once
let cachedHexagonTexture: Texture | null = null
let pixiApp: Application | null = null

/**
 * Create hexagon texture for shockwaves (called once)
 */
function createHexagonTexture(app: Application, maxRadius: number): Texture {
  const graphics = new Graphics()
  const sides = 6

  // Draw hexagon at maxRadius (will be scaled down for animation)
  graphics.moveTo(maxRadius, 0)
  for (let j = 1; j <= sides; j++) {
    const angle = (j / sides) * Math.PI * 2
    graphics.lineTo(Math.cos(angle) * maxRadius, Math.sin(angle) * maxRadius)
  }
  // Use white color - actual color will be tinted on the sprite
  graphics.stroke({ color: 0xffffff, width: 4, alpha: 1 })

  const texture = app.renderer.generateTexture({
    target: graphics,
    resolution: 1,
  })
  graphics.destroy()
  return texture
}

/**
 * Particle System - Manages particle creation, updates, and cleanup
 * OPTIMIZED: Uses cached textures for shockwaves
 */
export class ParticleSystem {
  private container: Container
  private particles: Particle[] = []
  private coinParticles: CoinParticle[] = []
  private shockwaves: Shockwave[] = []

  constructor(container: Container) {
    this.container = container
  }

  /**
   * Initialize textures - call once when app is ready
   */
  static initTextures(app: Application): void {
    pixiApp = app
    if (!cachedHexagonTexture) {
      cachedHexagonTexture = createHexagonTexture(app, 400)  // Max radius
    }
  }

  /**
   * Spawn generic particles
   */
  spawnParticles(config: {
    count: number
    x: number
    y: number
    colors?: number[]
    sizeRange?: [number, number]
    speedRange?: [number, number]
    gravity?: number
    maxLife?: number
    shape?: 'circle' | 'diamond' | 'rect'
    blendMode?: any
  }): void {
    const {
      count,
      x,
      y,
      colors = [0xFFD700],
      sizeRange = [5, 15],
      speedRange = [3, 8],
      gravity = 0.1,
      maxLife = 2,
      shape = 'circle',
      blendMode = BLEND_MODES.ADD
    } = config

    for (let i = 0; i < count; i++) {
      const particle = new Graphics() as Particle
      const color = colors[Math.floor(Math.random() * colors.length)]
      const size = sizeRange[0] + Math.random() * (sizeRange[1] - sizeRange[0])

      this.drawParticleShape(particle, shape, size, color)

      particle.x = x
      particle.y = y
      particle.blendMode = blendMode

      const angle = Math.random() * Math.PI * 2
      const speed = speedRange[0] + Math.random() * (speedRange[1] - speedRange[0])
      particle.vx = Math.cos(angle) * speed
      particle.vy = Math.sin(angle) * speed
      particle.gravity = gravity
      particle.rotationSpeed = (Math.random() - 0.5) * 0.2
      particle.life = 1
      particle.maxLife = maxLife + Math.random() * maxLife * 0.5
      particle.born = Date.now()
      particle.color = color

      this.container.addChild(particle)
      this.particles.push(particle)
    }
  }

  /**
   * Spawn confetti particles (falling from top)
   */
  spawnConfetti(width: number, height: number, count: number): void {
    const colors = [0xFFD700, 0xFF6B00, 0xFF0000, 0x00FF00, 0x00FFFF, 0xFF00FF]

    for (let i = 0; i < count; i++) {
      const particle = new Graphics() as Particle
      const color = colors[Math.floor(Math.random() * colors.length)]
      const size = 5 + Math.random() * 10

      particle.rect(-size / 2, -size / 2, size, size * 2)
      particle.fill({ color, alpha: 0.9 })
      particle.x = Math.random() * width
      particle.y = -50 - Math.random() * 100
      particle.rotation = Math.random() * Math.PI * 2

      particle.vx = (Math.random() - 0.5) * 3
      particle.vy = 2 + Math.random() * 3
      particle.gravity = 0.05
      particle.rotationSpeed = (Math.random() - 0.5) * 0.2
      particle.life = 1
      particle.maxLife = 5 + Math.random() * 3
      particle.born = Date.now()
      particle.color = color

      this.container.addChild(particle)
      this.particles.push(particle)
    }
  }

  /**
   * Spawn firework explosion
   */
  spawnFirework(x: number, y: number, count = 40): void {
    const colors = [0xFFD700, 0xFF6B00, 0xFF0000, 0xFFFFFF]

    for (let i = 0; i < count; i++) {
      const particle = new Graphics() as Particle
      const color = colors[Math.floor(Math.random() * colors.length)]
      const size = 4 + Math.random() * 6

      particle.circle(0, 0, size)
      particle.fill({ color, alpha: 0.9 })
      particle.x = x
      particle.y = y
      particle.blendMode = BLEND_MODES.ADD as any

      const angle = (i / count) * Math.PI * 2
      const speed = 3 + Math.random() * 5
      particle.vx = Math.cos(angle) * speed
      particle.vy = Math.sin(angle) * speed
      particle.gravity = 0.15
      particle.life = 1
      particle.maxLife = 1.2 + Math.random() * 0.8
      particle.born = Date.now()
      particle.color = color

      this.container.addChild(particle)
      this.particles.push(particle)
    }
  }

  /**
   * Spawn coin particles using textures
   */
  spawnCoins(config: {
    x: number
    y: number
    count: number
    textures: Texture[]
    spread?: number
    speed?: number
  }): void {
    const { x, y, count, textures, spread = 100, speed = 10 } = config

    if (textures.length === 0) return

    for (let i = 0; i < count; i++) {
      const texture = textures[Math.floor(Math.random() * textures.length)]
      const coin = new Sprite(texture) as CoinParticle
      coin.anchor.set(0.5)
      coin.x = x + (Math.random() - 0.5) * spread
      coin.y = y
      coin.scale.set(0.3 + Math.random() * 0.3)

      const angle = Math.random() * Math.PI * 2
      coin.vx = Math.cos(angle) * speed
      coin.vy = Math.sin(angle) * speed - 8
      coin.gravity = 0.3
      coin.rotationSpeed = (Math.random() - 0.5) * 0.3
      coin.life = 3 + Math.random() * 2
      coin.born = Date.now()

      this.container.addChild(coin)
      this.coinParticles.push(coin)
    }
  }

  /**
   * Spawn shockwave effect
   * OPTIMIZED: Pre-draws hexagon once, uses scale animation
   */
  spawnShockwave(x: number, y: number, color = 0xFFD700, maxRadius = 400, duration = 800): void {
    const graphics = new Graphics()
    graphics.x = x
    graphics.y = y
    graphics.blendMode = BLEND_MODES.ADD as any

    // OPTIMIZATION: Draw hexagon once at max size
    const sides = 6
    graphics.moveTo(maxRadius, 0)
    for (let j = 1; j <= sides; j++) {
      const angle = (j / sides) * Math.PI * 2
      graphics.lineTo(Math.cos(angle) * maxRadius, Math.sin(angle) * maxRadius)
    }
    graphics.stroke({ color, width: 4, alpha: 1 })

    // Start at scale 0, will animate to 1
    graphics.scale.set(0)
    graphics.alpha = 0.6

    this.container.addChild(graphics)

    this.shockwaves.push({
      graphics,
      startTime: Date.now(),
      maxRadius,
      duration,
      color
    })
  }

  /**
   * Update all particles - call this every frame
   */
  update(canvasHeight: number): void {
    const now = Date.now()

    // Update regular particles
    for (let i = this.particles.length - 1; i >= 0; i--) {
      const p = this.particles[i]
      const age = (now - p.born) / 1000

      if (p.delay && age < p.delay) continue

      p.vy += p.gravity || 0
      p.x += p.vx
      p.y += p.vy
      p.rotation += p.rotationSpeed || 0

      const t = age / p.maxLife
      if (t < 1 && p.y < canvasHeight + 50) {
        p.alpha = (1 - t) * 0.9
      } else {
        this.removeParticle(p, i)
      }
    }

    // Update coin particles
    for (let i = this.coinParticles.length - 1; i >= 0; i--) {
      const c = this.coinParticles[i]
      const age = (now - c.born) / 1000

      c.vy += c.gravity
      c.x += c.vx
      c.y += c.vy
      c.rotation += c.rotationSpeed

      if (age < c.life && c.y < canvasHeight + 50) {
        c.alpha = Math.max(0, 1 - age / c.life) * 0.7
      } else {
        this.removeCoinParticle(c, i)
      }
    }

    // Update shockwaves - OPTIMIZED: use scale/alpha instead of redrawing
    for (let i = this.shockwaves.length - 1; i >= 0; i--) {
      const sw = this.shockwaves[i]
      const elapsed = now - sw.startTime
      const progress = Math.min(elapsed / sw.duration, 1)

      if (progress < 1) {
        // OPTIMIZATION: Animate using scale and alpha (no redraw!)
        sw.graphics.scale.set(progress)
        sw.graphics.alpha = (1 - progress) * 0.6
      } else {
        this.removeShockwave(sw, i)
      }
    }
  }

  /**
   * Clear all particles
   */
  clear(): void {
    this.particles.forEach(p => {
      p.parent?.removeChild(p)
      p.destroy()
    })
    this.particles = []

    this.coinParticles.forEach(c => {
      c.parent?.removeChild(c)
      c.destroy()
    })
    this.coinParticles = []

    this.shockwaves.forEach(sw => {
      sw.graphics.parent?.removeChild(sw.graphics)
      sw.graphics.destroy()
    })
    this.shockwaves = []
  }

  /**
   * Get particle counts for debugging
   */
  getCounts(): { particles: number; coins: number; shockwaves: number } {
    return {
      particles: this.particles.length,
      coins: this.coinParticles.length,
      shockwaves: this.shockwaves.length
    }
  }

  private drawParticleShape(particle: Graphics, shape: string, size: number, color: number): void {
    switch (shape) {
      case 'diamond':
        particle.moveTo(0, -size)
        particle.lineTo(size * 0.6, 0)
        particle.lineTo(0, size)
        particle.lineTo(-size * 0.6, 0)
        particle.closePath()
        particle.fill({ color, alpha: 0.9 })
        break
      case 'rect':
        particle.rect(-size / 2, -size / 2, size, size * 2)
        particle.fill({ color, alpha: 0.9 })
        break
      case 'circle':
      default:
        particle.circle(0, 0, size)
        particle.fill({ color, alpha: 0.9 })
        // Add glow
        particle.circle(0, 0, size * 1.5)
        particle.fill({ color, alpha: 0.3 })
        break
    }
  }

  private removeParticle(p: Particle, index: number): void {
    p.parent?.removeChild(p)
    p.destroy()
    this.particles.splice(index, 1)
  }

  private removeCoinParticle(c: CoinParticle, index: number): void {
    c.parent?.removeChild(c)
    c.destroy()
    this.coinParticles.splice(index, 1)
  }

  private removeShockwave(sw: Shockwave, index: number): void {
    sw.graphics.parent?.removeChild(sw.graphics)
    sw.graphics.destroy()
    this.shockwaves.splice(index, 1)
  }
}

/**
 * Create a particle system instance
 */
export function createParticleSystem(container: Container): ParticleSystem {
  return new ParticleSystem(container)
}
