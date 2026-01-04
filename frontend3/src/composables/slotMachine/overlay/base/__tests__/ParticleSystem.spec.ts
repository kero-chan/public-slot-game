import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { Container, Texture } from 'pixi.js'
import { ParticleSystem, createParticleSystem } from '../ParticleSystem'

// Mock pixi.js to avoid WebGL context issues in tests
vi.mock('pixi.js', async () => {
  const actual = await vi.importActual('pixi.js')
  return {
    ...actual,
    Container: class MockContainer {
      children: any[] = []
      addChild(child: any) {
        this.children.push(child)
        child.parent = this
        return child
      }
      removeChild(child: any) {
        const index = this.children.indexOf(child)
        if (index > -1) {
          this.children.splice(index, 1)
          child.parent = null
        }
      }
    },
    Graphics: class MockGraphics {
      parent: any = null
      x = 0
      y = 0
      alpha = 1
      blendMode = 0
      rotation = 0
      vx = 0
      vy = 0
      gravity = 0
      rotationSpeed = 0
      life = 1
      maxLife = 1
      born = 0
      delay = 0
      color = 0

      circle() { return this }
      rect() { return this }
      moveTo() { return this }
      lineTo() { return this }
      closePath() { return this }
      fill() { return this }
      stroke() { return this }
      clear() { return this }
      destroy() { }
    },
    Sprite: class MockSprite {
      parent: any = null
      x = 0
      y = 0
      alpha = 1
      rotation = 0
      scale = { x: 1, y: 1, set: (v: number) => { this.scale.x = v; this.scale.y = v } }
      anchor = { set: () => {} }
      texture: any = null
      vx = 0
      vy = 0
      gravity = 0
      rotationSpeed = 0
      life = 1
      born = 0

      destroy() { }
    },
    Texture: {
      from: () => ({})
    }
  }
})

describe('ParticleSystem', () => {
  let container: any
  let particleSystem: ParticleSystem

  beforeEach(() => {
    vi.useFakeTimers()
    const MockContainer = vi.mocked(Container)
    container = new MockContainer()
    particleSystem = createParticleSystem(container as any)
  })

  afterEach(() => {
    particleSystem.clear()
    vi.useRealTimers()
  })

  describe('spawnParticles', () => {
    it('should add particles to container', () => {
      particleSystem.spawnParticles({
        count: 5,
        x: 100,
        y: 100
      })

      expect(container.children.length).toBe(5)
    })

    it('should position particles at spawn point', () => {
      particleSystem.spawnParticles({
        count: 1,
        x: 200,
        y: 150
      })

      expect(container.children[0].x).toBe(200)
      expect(container.children[0].y).toBe(150)
    })

    it('should use custom colors when provided', () => {
      particleSystem.spawnParticles({
        count: 10,
        x: 0,
        y: 0,
        colors: [0xFF0000, 0x00FF00]
      })

      expect(container.children.length).toBe(10)
    })
  })

  describe('spawnFirework', () => {
    it('should spawn multiple particles', () => {
      particleSystem.spawnFirework(100, 100, 20)

      expect(container.children.length).toBe(20)
    })
  })

  describe('spawnShockwave', () => {
    it('should add shockwave graphics to container', () => {
      particleSystem.spawnShockwave(100, 100, 0xFFD700, 400)

      expect(container.children.length).toBe(1)
    })
  })

  describe('spawnConfetti', () => {
    it('should spawn confetti particles', () => {
      particleSystem.spawnConfetti(800, 600, 50)

      expect(container.children.length).toBe(50)
    })
  })

  describe('clear', () => {
    it('should remove all particles from container', () => {
      particleSystem.spawnParticles({ count: 10, x: 0, y: 0 })
      particleSystem.spawnShockwave(0, 0)
      particleSystem.spawnFirework(0, 0, 5)

      expect(container.children.length).toBeGreaterThan(0)

      particleSystem.clear()

      expect(particleSystem.getCounts()).toEqual({ particles: 0, coins: 0, shockwaves: 0 })
    })
  })

  describe('getCounts', () => {
    it('should return correct counts', () => {
      particleSystem.spawnParticles({ count: 5, x: 0, y: 0 })
      particleSystem.spawnShockwave(0, 0)

      const counts = particleSystem.getCounts()

      expect(counts.particles).toBe(5)
      expect(counts.shockwaves).toBe(1)
    })
  })

  describe('update', () => {
    it('should not throw when called with no particles', () => {
      expect(() => {
        particleSystem.update(600)
      }).not.toThrow()
    })

    it('should update particle positions', () => {
      particleSystem.spawnParticles({
        count: 1,
        x: 100,
        y: 100,
        gravity: 0.1
      })

      const particle = container.children[0]
      const initialY = particle.y

      // Manually set physics properties
      particle.vx = 1
      particle.vy = 1
      particle.gravity = 0.1
      particle.born = Date.now()
      particle.maxLife = 10 // Long life so it doesn't die

      particleSystem.update(800)

      // Particle should have moved
      expect(particle.y).not.toBe(initialY)
    })
  })
})
