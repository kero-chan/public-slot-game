/**
 * Ambient Particles System - Performance Optimized
 * A skin-agnostic particle system for subtle background ambient motion
 *
 * Optimizations:
 * - Object pooling for sprites (no create/destroy during gameplay)
 * - Pre-allocated arrays (no allocations in update loop)
 * - Inlined math operations
 * - Minimal property access
 */

import { Assets, Container, Sprite, Texture } from 'pixi.js'

// Import particle textures
import particle1 from '@/assets/images/japaneseOmakase/particles/particle1.webp'

// ============================================================================
// Types
// ============================================================================

export type MotionProfileId = 'DRIFT' | 'SWEEP' | 'FLOAT' | 'FALL' | 'LEAF'

export interface MotionProfile {
  id: MotionProfileId
  lifeMs: [number, number]
  alpha: [number, number]
  scale: [number, number]
  driftSpeed: [number, number]
  bobAmp: [number, number]
  bobFreq: [number, number]
  rotSpeed: [number, number]
  sweep?: { curvature: [number, number] }
  blendMode?: 'normal' | 'add'
}

export interface SpawnZone {
  id: string
  rect: { x: number; y: number; w: number; h: number }
  weight?: number
  profile?: MotionProfileId
}

export interface Budget {
  target: number
  min: number
  max: number
}

export type ParticleEvent =
  | { type: 'BOOST'; amount: number; durationMs: number }
  | { type: 'INTENSIFY'; mul: number; durationMs: number }
  | { type: 'BURST'; count: number }
  | { type: 'RESET' }

// Flat particle structure for cache efficiency
interface Particle {
  sprite: Sprite
  active: boolean
  birthTime: number
  lifeMs: number
  profileId: number // 0=DRIFT, 1=SWEEP, 2=FLOAT, 3=FALL
  startX: number
  startY: number
  endX: number
  endY: number
  ctrlX: number
  ctrlY: number
  baseScale: number
  alphaMin: number
  alphaMax: number
  driftSpeed: number
  driftAngle: number
  bobAmp: number
  bobFreq: number
  bobPhase: number
  rotSpeed: number
}

export interface AmbientParticlesOptions {
  parent: Container
  textures?: Texture[]
  spawnZones: SpawnZone[]
  profiles?: MotionProfile[]
  budget?: Budget
  bounds: { w: number; h: number }
  uiExclusionRect?: { x: number; y: number; w: number; h: number }
  zIndex?: number
  maxHardCap?: number
}

// ============================================================================
// Constants
// ============================================================================

const PROFILE_DRIFT = 0
const PROFILE_SWEEP = 1
const PROFILE_FLOAT = 2
const PROFILE_FALL = 3

const FADE_IN_MS = 500
const FADE_OUT_MS = 800
const SPAWN_THROTTLE_MS = 100
const PI2 = Math.PI * 2

// Reference resolution for scale calculation
const REF_WIDTH = 1920
const REF_HEIGHT = 1080

const DEFAULT_PROFILES: MotionProfile[] = [
  {
    id: 'DRIFT',
    lifeMs: [8000, 15000],
    alpha: [0.2, 0.5],
    scale: [0.08, 0.2],
    driftSpeed: [10, 30],
    bobAmp: [5, 15],
    bobFreq: [0.5, 1.5],
    rotSpeed: [-0.5, 0.5],
    blendMode: 'normal'
  },
  {
    id: 'SWEEP',
    lifeMs: [6000, 12000],
    alpha: [0.15, 0.4],
    scale: [0.1, 0.25],
    driftSpeed: [40, 80],
    bobAmp: [0, 5],
    bobFreq: [0.3, 0.8],
    rotSpeed: [-1, 1],
    sweep: { curvature: [0.2, 0.6] },
    blendMode: 'normal'
  },
  {
    id: 'FLOAT',
    lifeMs: [10000, 20000],
    alpha: [0.25, 0.5],
    scale: [0.05, 0.15],
    driftSpeed: [2, 8],
    bobAmp: [3, 10],
    bobFreq: [0.2, 0.6],
    rotSpeed: [-0.2, 0.2],
    blendMode: 'normal'
  },
  {
    // Leaf-falling effect: gentle downward drift with natural sway
    id: 'FALL',
    lifeMs: [6000, 12000],      // Time to traverse screen
    alpha: [0.4, 0.7],          // Visible but not distracting
    scale: [0.02, 0.04],        // Small leaves (textures are 1024px, scaled 50%)
    driftSpeed: [50, 100],      // Fall speed (pixels per second)
    bobAmp: [20, 50],           // Horizontal sway amplitude (leaf flutter)
    bobFreq: [0.5, 1.0],        // Sway frequency (gentle oscillation)
    rotSpeed: [-1.2, 1.2],      // Rotation as leaf tumbles
    blendMode: 'normal'
  }
]

const DEFAULT_BUDGET: Budget = {
  target: 10,
  min: 8,
  max: 14
}

// ============================================================================
// AmbientParticles Class
// ============================================================================

export class AmbientParticles {
  private container: Container
  private textures: Texture[] = []
  private spawnZones: SpawnZone[]
  private profiles: MotionProfile[]
  private budget: Budget
  private bounds: { w: number; h: number }
  private uiExclusionRect?: { x: number; y: number; w: number; h: number }
  private maxHardCap: number

  // Object pool
  private pool: Particle[] = []
  private activeCount = 0
  private running = false
  private spawnThrottle = 0
  private texturesLoaded = false

  // Event state
  private boostAmount = 0
  private boostEndTime = 0
  private intensifyMul = 1
  private intensifyEndTime = 0

  // Pre-computed zone weights
  private zoneTotalWeight = 0
  private zoneWeights: number[] = []

  // Viewport-based scale factor
  private scaleFactor = 1

  constructor(opts: AmbientParticlesOptions) {
    // Regular Container - efficient enough for 10-50 particles
    this.container = new Container()
    this.container.zIndex = opts.zIndex ?? -100

    opts.parent.addChild(this.container)

    this.spawnZones = opts.spawnZones
    this.budget = opts.budget ?? { ...DEFAULT_BUDGET }
    this.bounds = opts.bounds
    this.uiExclusionRect = opts.uiExclusionRect
    this.maxHardCap = opts.maxHardCap ?? 50
    this.profiles = opts.profiles ?? DEFAULT_PROFILES

    this.computeZoneWeights()
    this.computeScaleFactor()

    // Load textures
    if (opts.textures && opts.textures.length > 0) {
      this.textures = opts.textures
      this.texturesLoaded = true
      this.initPool()
    } else {
      this.loadDefaultTextures()
    }
  }

  private computeZoneWeights(): void {
    this.zoneWeights = []
    this.zoneTotalWeight = 0
    for (const zone of this.spawnZones) {
      const w = zone.weight ?? 1
      this.zoneTotalWeight += w
      this.zoneWeights.push(w)
    }
  }

  private computeScaleFactor(): void {
    // Scale particles relative to viewport size
    // Use the smaller dimension ratio to ensure particles look right on any aspect ratio
    const scaleX = this.bounds.w / REF_WIDTH
    const scaleY = this.bounds.h / REF_HEIGHT
    this.scaleFactor = Math.min(scaleX, scaleY)
  }

  private async loadDefaultTextures(): Promise<void> {
    try {
      const texturePaths = [
        { alias: 'ambient_particle1', src: particle1 },
      ]

      for (const { alias, src } of texturePaths) {
        if (!Assets.resolver.hasKey(alias)) {
          Assets.add({ alias, src })
        }
      }

      const loaded = await Assets.load(texturePaths.map(t => t.alias))
      this.textures = Object.values(loaded).filter(t => t instanceof Texture) as Texture[]
      this.texturesLoaded = this.textures.length > 0

      if (this.texturesLoaded) {
        this.initPool()
      }
    } catch (e) {
      // Silent fail - particles are optional ambient effect
    }
  }

  private initPool(): void {
    // Pre-allocate all particle objects
    for (let i = 0; i < this.maxHardCap; i++) {
      const sprite = new Sprite(this.textures[0])
      sprite.anchor.set(0.5)
      sprite.visible = false
      sprite.alpha = 0
      this.container.addChild(sprite)

      this.pool.push({
        sprite,
        active: false,
        birthTime: 0,
        lifeMs: 0,
        profileId: 0,
        startX: 0,
        startY: 0,
        endX: 0,
        endY: 0,
        ctrlX: 0,
        ctrlY: 0,
        baseScale: 0,
        alphaMin: 0,
        alphaMax: 0,
        driftSpeed: 0,
        driftAngle: 0,
        bobAmp: 0,
        bobFreq: 0,
        bobPhase: 0,
        rotSpeed: 0
      })
    }
  }

  start(): void {
    this.running = true
  }

  stop(): void {
    this.running = false
  }

  destroy(): void {
    this.stop()
    this.container.destroy({ children: true })
    this.pool = []
  }

  setSpawnZones(zones: SpawnZone[]): void {
    this.spawnZones = zones
    this.computeZoneWeights()
  }

  setBounds(bounds: { w: number; h: number }): void {
    this.bounds = bounds
    this.computeScaleFactor()
  }

  setUiExclusionRect(rect: { x: number; y: number; w: number; h: number } | undefined): void {
    this.uiExclusionRect = rect
  }

  emit(event: ParticleEvent): void {
    const now = performance.now()

    switch (event.type) {
      case 'BOOST':
        this.boostAmount = event.amount
        this.boostEndTime = now + event.durationMs
        break

      case 'INTENSIFY':
        this.intensifyMul = event.mul
        this.intensifyEndTime = now + event.durationMs
        break

      case 'BURST': {
        const count = Math.min(event.count, this.maxHardCap - this.activeCount)
        for (let i = 0; i < count; i++) {
          this.spawnParticle(now)
        }
        break
      }

      case 'RESET':
        this.boostAmount = 0
        this.boostEndTime = 0
        this.intensifyMul = 1
        this.intensifyEndTime = 0
        // Quick fade out
        for (let i = 0; i < this.pool.length; i++) {
          const p = this.pool[i]
          if (p.active) {
            const age = now - p.birthTime
            p.lifeMs = Math.min(p.lifeMs, age + 500)
          }
        }
        break
    }
  }

  update(dtMs: number): void {
    if (!this.running || !this.texturesLoaded) return
    if (this.spawnZones.length === 0) return

    const now = performance.now()

    // Update event state
    if (now > this.boostEndTime) this.boostAmount = 0
    if (now > this.intensifyEndTime) this.intensifyMul = 1

    // Effective target
    const targetCount = this.budget.target + this.boostAmount
    const maxCount = Math.min(this.budget.max + this.boostAmount, this.maxHardCap)

    // Spawn if needed (throttled)
    this.spawnThrottle -= dtMs
    if (this.spawnThrottle <= 0 && this.activeCount < targetCount) {
      const toSpawn = Math.min(2, targetCount - this.activeCount)
      for (let i = 0; i < toSpawn; i++) {
        this.spawnParticle(now)
      }
      this.spawnThrottle = SPAWN_THROTTLE_MS
    }

    // Update particles
    const boundsW = this.bounds.w
    const boundsH = this.bounds.h
    const intensify = this.intensifyMul
    let excessCount = this.activeCount - maxCount

    for (let i = 0; i < this.pool.length; i++) {
      const p = this.pool[i]
      if (!p.active) continue

      const age = now - p.birthTime
      const lifeMs = p.lifeMs

      // Check lifetime
      if (age >= lifeMs) {
        p.active = false
        p.sprite.visible = false
        this.activeCount--
        continue
      }

      const t = age / lifeMs

      // Alpha with fade in/out
      let alpha = p.alphaMin + (p.alphaMax - p.alphaMin) * (0.5 + 0.5 * Math.sin(t * Math.PI))
      if (age < FADE_IN_MS) {
        alpha *= age / FADE_IN_MS
      } else if (age > lifeMs - FADE_OUT_MS) {
        alpha *= (lifeMs - age) / FADE_OUT_MS
      }
      alpha *= intensify

      // Position based on profile
      let x: number, y: number
      const profileId = p.profileId

      if (profileId === PROFILE_SWEEP) {
        // Quadratic bezier
        const mt = 1 - t
        x = mt * mt * p.startX + 2 * mt * t * p.ctrlX + t * t * p.endX
        y = mt * mt * p.startY + 2 * mt * t * p.ctrlY + t * t * p.endY
      } else if (profileId === PROFILE_FALL) {
        // Leaf falling: downward drift with horizontal sway and slight diagonal movement
        const ageSec = age * 0.001
        // Primary vertical fall
        y = p.startY + p.driftSpeed * ageSec
        // Sinusoidal horizontal sway (leaf flutter) + slight diagonal drift
        const sway = Math.sin(age * p.bobFreq * 0.001 + p.bobPhase) * p.bobAmp
        const horizontalDrift = Math.cos(p.driftAngle) * p.driftSpeed * 0.3 * ageSec // Slight horizontal movement
        x = p.startX + sway + horizontalDrift
      } else {
        // Drift (DRIFT or FLOAT)
        const ageSec = age * 0.001
        x = p.startX + Math.cos(p.driftAngle) * p.driftSpeed * ageSec
        y = p.startY + Math.sin(p.driftAngle) * p.driftSpeed * ageSec
        y += Math.sin(age * p.bobFreq * 0.001 + p.bobPhase) * p.bobAmp
      }

      // Out of bounds check (allow particles above screen for falling effect)
      if (x < -100 || x > boundsW + 100 || y < -150 || y > boundsH + 50) {
        p.active = false
        p.sprite.visible = false
        this.activeCount--
        continue
      }

      // Force early death if over max
      if (excessCount > 0) {
        p.lifeMs = Math.min(lifeMs, age + 300)
        excessCount--
      }

      // Update sprite
      const sprite = p.sprite
      sprite.x = x
      sprite.y = y
      sprite.alpha = alpha > 1 ? 1 : (alpha < 0 ? 0 : alpha)
      sprite.rotation += p.rotSpeed * dtMs * 0.001

      // Breathing scale
      const breathe = 1 + 0.1 * Math.sin(age * 0.002)
      const scale = p.baseScale * breathe
      sprite.scale.x = scale
      sprite.scale.y = scale
    }
  }

  private spawnParticle(now: number): void {
    if (this.activeCount >= this.maxHardCap) return
    if (this.textures.length === 0) return

    // Find inactive particle
    let p: Particle | null = null
    for (let i = 0; i < this.pool.length; i++) {
      if (!this.pool[i].active) {
        p = this.pool[i]
        break
      }
    }
    if (!p) return

    // Pick zone by weight
    let rand = Math.random() * this.zoneTotalWeight
    let zoneIdx = 0
    for (let i = 0; i < this.zoneWeights.length; i++) {
      rand -= this.zoneWeights[i]
      if (rand <= 0) {
        zoneIdx = i
        break
      }
    }
    const zone = this.spawnZones[zoneIdx]

    // Pick position avoiding UI
    let x: number, y: number
    let attempts = 0
    const rect = zone.rect
    const excl = this.uiExclusionRect

    do {
      x = rect.x + Math.random() * rect.w
      y = rect.y + Math.random() * rect.h
      attempts++
    } while (
      excl &&
      x >= excl.x && x <= excl.x + excl.w &&
      y >= excl.y && y <= excl.y + excl.h &&
      attempts < 10
    )

    if (excl && x >= excl.x && x <= excl.x + excl.w && y >= excl.y && y <= excl.y + excl.h) {
      return
    }

    // Pick profile
    let profileId: number
    if (zone.profile) {
      profileId = zone.profile === 'DRIFT' ? 0 : zone.profile === 'SWEEP' ? 1 : zone.profile === 'FLOAT' ? 2 : 3
    } else {
      profileId = Math.floor(Math.random() * 4)
    }
    const profile = this.profiles[profileId]

    // Random texture
    const texIdx = Math.floor(Math.random() * this.textures.length)
    p.sprite.texture = this.textures[texIdx]

    // Generate properties inline
    const lifeRange = profile.lifeMs[1] - profile.lifeMs[0]
    const scaleRange = profile.scale[1] - profile.scale[0]
    const speedRange = profile.driftSpeed[1] - profile.driftSpeed[0]
    const bobAmpRange = profile.bobAmp[1] - profile.bobAmp[0]
    const bobFreqRange = profile.bobFreq[1] - profile.bobFreq[0]
    const rotRange = profile.rotSpeed[1] - profile.rotSpeed[0]

    p.active = true
    p.birthTime = now
    p.lifeMs = profile.lifeMs[0] + Math.random() * lifeRange
    p.profileId = profileId
    p.startX = x
    p.startY = y
    p.baseScale = (profile.scale[0] + Math.random() * scaleRange) * this.scaleFactor
    p.alphaMin = profile.alpha[0]
    p.alphaMax = profile.alpha[1]
    p.driftSpeed = profile.driftSpeed[0] + Math.random() * speedRange
    p.driftAngle = Math.random() * PI2
    p.bobAmp = profile.bobAmp[0] + Math.random() * bobAmpRange
    p.bobFreq = profile.bobFreq[0] + Math.random() * bobFreqRange
    p.bobPhase = Math.random() * PI2
    p.rotSpeed = profile.rotSpeed[0] + Math.random() * rotRange

    // Setup sweep path
    if (profileId === PROFILE_SWEEP && profile.sweep) {
      const curvRange = profile.sweep.curvature[1] - profile.sweep.curvature[0]
      const curvature = profile.sweep.curvature[0] + Math.random() * curvRange
      const angle = Math.random() * PI2
      const dist = p.driftSpeed * (p.lifeMs * 0.001)

      p.endX = x + Math.cos(angle) * dist
      p.endY = y + Math.sin(angle) * dist

      const midX = (x + p.endX) * 0.5
      const midY = (y + p.endY) * 0.5
      const perpAngle = angle + Math.PI * 0.5
      const curveDist = dist * curvature * (Math.random() > 0.5 ? 1 : -1)

      p.ctrlX = midX + Math.cos(perpAngle) * curveDist
      p.ctrlY = midY + Math.sin(perpAngle) * curveDist
    }

    // Init sprite
    const sprite = p.sprite
    sprite.x = x
    sprite.y = y
    sprite.scale.set(p.baseScale)
    sprite.alpha = 0
    sprite.rotation = Math.random() * PI2
    sprite.visible = true

    this.activeCount++
  }

  get particleCount(): number {
    return this.activeCount
  }

  get isRunning(): boolean {
    return this.running
  }

  setVisible(visible: boolean): void {
    this.container.visible = visible
  }
}

// ============================================================================
// Factory Function
// ============================================================================

export function createAmbientParticles(opts: AmbientParticlesOptions): AmbientParticles {
  return new AmbientParticles(opts)
}
