import { Container, Graphics, Sprite, Texture, RenderTexture, Application } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'

// Cached particle texture - generated once and reused
let cachedParticleTexture: Texture | null = null
let pixiApp: Application | null = null

/**
 * Particle for rising sparkle effect - now uses Sprite instead of Graphics
 */
interface Particle {
  sprite: Sprite  // Changed from Graphics to Sprite
  x: number
  y: number
  vx: number
  vy: number
  life: number
  maxLife: number
  size: number
}

/**
 * Highlight elements for a column
 */
interface ColumnHighlight {
  col: number
  left: Graphics
  right: Graphics
  topEdge: Graphics
  bottomEdge: Graphics
}

/**
 * Column highlight manager interface
 */
export interface ColumnHighlightManager {
  container: Container
  initialize: (
    cols: number,
    columnWidth: number,
    columnHeight: number,
    columnX: number,
    columnY: number,
    stepX: number
  ) => void
  showHighlight: (col: number, alpha?: number) => void
  hideHighlight: (col: number) => void
  hideAll: () => void
  update: (activeSlowdownColumn: number) => void
  initTextures: (app: Application) => void
}

/**
 * Create and cache the particle texture (called once)
 */
function createParticleTexture(app: Application): Texture {
  const graphics = new Graphics()
  const size = 8  // Standard size, will be scaled per-particle

  // Draw 4-point star shape
  const points = 4
  const outerRadius = size
  const innerRadius = size * 0.4
  graphics.moveTo(0, -outerRadius)
  for (let i = 0; i < points * 2; i++) {
    const radius = i % 2 === 0 ? outerRadius : innerRadius
    const angle = (i * Math.PI) / points - Math.PI / 2
    graphics.lineTo(Math.cos(angle) * radius, Math.sin(angle) * radius)
  }
  graphics.closePath()
  graphics.fill({ color: 0xFFFFFF, alpha: 0.95 })

  // Add golden glow halo
  graphics.circle(0, 0, size * 1.8)
  graphics.fill({ color: 0xFFD700, alpha: 0.4 })

  const texture = app.renderer.generateTexture({
    target: graphics,
    resolution: 2,
  })
  graphics.destroy()
  return texture
}

/**
 * Column Highlight Manager
 * Enhanced visual effects for slowing-down columns during anticipation mode:
 * 1. Full column background glow
 * 2. Intense border streams with multiple glow layers
 * 3. Rising sparkle particle effects
 */
export function createColumnHighlightManager(): ColumnHighlightManager {
  const container = new Container()
  const highlightElements: ColumnHighlight[] = []
  const particles: Particle[] = []
  const particleContainer = new Container()

  let animationTime = 0
  let lastActiveColumn = -1

  // Store dimensions for particle spawning
  let storedColumnWidth = 0
  let storedColumnHeight = 0
  let storedColumnX = 0
  let storedColumnY = 0
  let storedStepX = 0

  container.addChild(particleContainer)

  /**
   * Create enhanced glowing stream graphic for borders
   */
  function createGlowingStream(width: number, height: number, isLeft: boolean): Graphics {
    const graphics = new Graphics()
    const numSegments = 50
    const segmentHeight = height / numSegments

    for (let i = 0; i < numSegments; i++) {
      const y = i * segmentHeight
      // Create wave pattern for more dynamic look
      const waveAlpha = Math.sin((i / numSegments) * Math.PI) * 0.5 + 0.5

      // Layer 1: Outermost soft glow (very wide)
      if (isLeft) {
        graphics.rect(-width * 1.5, y, width * 4, segmentHeight)
      } else {
        graphics.rect(-width * 1.5, y, width * 4, segmentHeight)
      }
      graphics.fill({ color: 0xFFD700, alpha: waveAlpha * 0.2 })

      // Layer 2: Outer glow
      if (isLeft) {
        graphics.rect(-width * 0.5, y, width * 2.5, segmentHeight)
      } else {
        graphics.rect(-width, y, width * 2.5, segmentHeight)
      }
      graphics.fill({ color: 0xFFA500, alpha: waveAlpha * 0.4 }) // Orange tint

      // Layer 3: Middle glow (gold)
      graphics.rect(0, y, width * 1.8, segmentHeight)
      graphics.fill({ color: 0xFFD700, alpha: waveAlpha * 0.7 })

      // Layer 4: Inner bright glow
      graphics.rect(width * 0.2, y, width * 1.2, segmentHeight)
      graphics.fill({ color: 0xFFE44D, alpha: waveAlpha * 0.9 }) // Bright gold

      // Layer 5: Core - white hot center
      graphics.rect(width * 0.4, y, width * 0.5, segmentHeight)
      graphics.fill({ color: 0xFFFFFF, alpha: waveAlpha * 1.0 })
    }

    return graphics
  }

  /**
   * Create horizontal edge glow (top or bottom)
   */
  function createEdgeGlow(width: number, edgeHeight: number, isTop: boolean): Graphics {
    const graphics = new Graphics()
    const numSegments = 30
    const segmentWidth = width / numSegments

    for (let i = 0; i < numSegments; i++) {
      const x = i * segmentWidth
      // Fade from edges to center
      const distFromCenter = Math.abs(i - numSegments / 2) / (numSegments / 2)
      const alpha = (1 - distFromCenter * 0.5) * 0.8

      // Outer glow
      const outerY = isTop ? -edgeHeight * 1.5 : 0
      graphics.rect(x, outerY, segmentWidth, edgeHeight * 2.5)
      graphics.fill({ color: 0xFFD700, alpha: alpha * 0.3 })

      // Inner glow
      const innerY = isTop ? -edgeHeight * 0.5 : edgeHeight * 0.3
      graphics.rect(x, innerY, segmentWidth, edgeHeight * 1.2)
      graphics.fill({ color: 0xFFE44D, alpha: alpha * 0.6 })

      // Core
      const coreY = isTop ? 0 : edgeHeight * 0.5
      graphics.rect(x, coreY, segmentWidth, edgeHeight * 0.5)
      graphics.fill({ color: 0xFFFFFF, alpha: alpha * 0.8 })
    }

    return graphics
  }

  /**
   * Create a sparkle particle using cached texture (OPTIMIZED)
   */
  function createParticle(x: number, y: number): Particle {
    const size = 3 + Math.random() * 5

    // Use cached texture if available, otherwise create a simple sprite
    const sprite = cachedParticleTexture
      ? new Sprite(cachedParticleTexture)
      : new Sprite()  // Fallback (should rarely happen)

    sprite.anchor.set(0.5)
    sprite.blendMode = BLEND_MODES.ADD as any
    sprite.x = x
    sprite.y = y

    // Scale based on desired size
    const scale = size / 8  // 8 is the base texture size
    sprite.scale.set(scale)

    return {
      sprite,
      x,
      y,
      vx: (Math.random() - 0.5) * 1.2, // Slight horizontal drift
      vy: 2 + Math.random() * 3, // Fall downward (same as spin direction)
      life: 1,
      maxLife: 1,
      size
    }
  }

  /**
   * Spawn particles along column borders (full height, flowing down)
   * OPTIMIZED: Reduced spawn rate slightly for better performance
   */
  function spawnParticles(col: number): void {
    if (storedColumnWidth === 0) return

    const colX = storedColumnX + col * storedStepX

    // Spawn along left border (full column height) - reduced rate from 0.35 to 0.25
    if (Math.random() < 0.25) {
      const spawnY = storedColumnY + Math.random() * storedColumnHeight * 0.7
      const particle = createParticle(colX + Math.random() * 12, spawnY)
      particles.push(particle)
      particleContainer.addChild(particle.sprite)
    }

    // Spawn along right border (full column height) - reduced rate from 0.35 to 0.25
    if (Math.random() < 0.25) {
      const spawnY = storedColumnY + Math.random() * storedColumnHeight * 0.7
      const particle = createParticle(colX + storedColumnWidth - Math.random() * 12, spawnY)
      particles.push(particle)
      particleContainer.addChild(particle.sprite)
    }
  }

  /**
   * Update particles (falling downward)
   */
  function updateParticles(): void {
    for (let i = particles.length - 1; i >= 0; i--) {
      const p = particles[i]

      // Update position (falling down)
      p.x += p.vx
      p.y += p.vy
      p.sprite.x = p.x
      p.sprite.y = p.y

      // Add slight horizontal drift
      p.vx += (Math.random() - 0.5) * 0.08

      // Rotate the star as it falls
      p.sprite.rotation += 0.05

      // Fade out
      p.life -= 0.012
      p.sprite.alpha = p.life * 0.9

      // Scale variation as it falls (multiply by base scale)
      const baseScale = p.size / 8
      const lifeFactor = 0.6 + p.life * 0.4
      p.sprite.scale.set(baseScale * lifeFactor)

      // Remove dead particles (when they exit bottom or fade out)
      if (p.life <= 0 || p.y > storedColumnY + storedColumnHeight + 30) {
        particleContainer.removeChild(p.sprite)
        p.sprite.destroy()  // Don't destroy texture - it's cached
        particles.splice(i, 1)
      }
    }
  }

  /**
   * Clear all particles
   */
  function clearParticles(): void {
    particles.forEach(p => {
      particleContainer.removeChild(p.sprite)
      p.sprite.destroy()  // Don't destroy texture - it's cached
    })
    particles.length = 0
  }

  /**
   * Initialize textures - call once when app is ready
   */
  function initTextures(app: Application): void {
    pixiApp = app
    if (!cachedParticleTexture) {
      cachedParticleTexture = createParticleTexture(app)
    }
  }

  /**
   * Initialize highlight elements for all columns
   */
  function initialize(
    cols: number,
    columnWidth: number,
    columnHeight: number,
    columnX: number,
    columnY: number,
    stepX: number
  ): void {
    // Store dimensions
    storedColumnWidth = columnWidth
    storedColumnHeight = columnHeight
    storedColumnX = columnX
    storedColumnY = columnY
    storedStepX = stepX

    // Clear existing elements
    highlightElements.forEach(({ left, right, topEdge, bottomEdge }) => {
      container.removeChild(left)
      container.removeChild(right)
      container.removeChild(topEdge)
      container.removeChild(bottomEdge)
    })
    highlightElements.length = 0
    clearParticles()

    const borderWidth = 8 // Thinner borders
    const edgeHeight = 10

    for (let col = 0; col < cols; col++) {
      const colX = columnX + col * stepX

      // Left border stream
      const left = createGlowingStream(borderWidth, columnHeight, true)
      left.x = colX - borderWidth * 0.5
      left.y = columnY
      left.alpha = 0
      container.addChild(left)

      // Right border stream
      const right = createGlowingStream(borderWidth, columnHeight, false)
      right.x = colX + columnWidth - borderWidth * 0.5
      right.y = columnY
      right.alpha = 0
      container.addChild(right)

      // Top edge glow (hidden)
      const topEdge = createEdgeGlow(columnWidth, edgeHeight, true)
      topEdge.x = colX
      topEdge.y = columnY
      topEdge.alpha = 0
      topEdge.visible = false
      container.addChild(topEdge)

      // Bottom edge glow (hidden)
      const bottomEdge = createEdgeGlow(columnWidth, edgeHeight, false)
      bottomEdge.x = colX
      bottomEdge.y = columnY + columnHeight - edgeHeight
      bottomEdge.alpha = 0
      bottomEdge.visible = false
      container.addChild(bottomEdge)

      highlightElements.push({ col, left, right, topEdge, bottomEdge })
    }

    // Make sure particle container is on top
    container.removeChild(particleContainer)
    container.addChild(particleContainer)
  }

  /**
   * Update animation
   */
  function updateAnimation(): void {
    animationTime += 0.12 // Faster animation

    // More dramatic pulse
    const pulse = Math.sin(animationTime * 2) * 0.2 + 0.8 // 0.6-1.0

    highlightElements.forEach(({ left, right, topEdge, bottomEdge, col }) => {
      if (left.alpha > 0) {
        // Borders pulse
        left.alpha = pulse
        right.alpha = pulse

        // Edges hidden (no top/bottom)
        topEdge.alpha = 0
        bottomEdge.alpha = 0

        // Spawn particles for active column
        if (col === lastActiveColumn) {
          spawnParticles(col)
        }
      }
    })

    // Update existing particles
    updateParticles()
  }

  /**
   * Show highlight for a specific column
   */
  function showHighlight(col: number, alpha: number = 1.0): void {
    const highlight = highlightElements.find(h => h.col === col)
    if (highlight) {
      highlight.left.alpha = alpha
      highlight.right.alpha = alpha
      highlight.topEdge.alpha = 0
      highlight.bottomEdge.alpha = 0
      highlight.left.visible = true
      highlight.right.visible = true
      highlight.topEdge.visible = false
      highlight.bottomEdge.visible = false
    }
  }

  /**
   * Hide highlight for a specific column
   */
  function hideHighlight(col: number): void {
    const highlight = highlightElements.find(h => h.col === col)
    if (highlight) {
      highlight.left.alpha = 0
      highlight.right.alpha = 0
      highlight.topEdge.alpha = 0
      highlight.bottomEdge.alpha = 0
    }
  }

  /**
   * Hide all highlights
   */
  function hideAll(): void {
    highlightElements.forEach(({ left, right, topEdge, bottomEdge }) => {
      left.alpha = 0
      right.alpha = 0
      topEdge.alpha = 0
      bottomEdge.alpha = 0
    })
    clearParticles()
    lastActiveColumn = -1
  }

  /**
   * Update highlights based on active slowdown column
   */
  function update(activeSlowdownColumn: number): void {
    if (highlightElements.length === 0) return

    // Track column changes for particle spawning
    if (activeSlowdownColumn !== lastActiveColumn) {
      if (lastActiveColumn >= 0) {
        hideHighlight(lastActiveColumn)
      }
      lastActiveColumn = activeSlowdownColumn
    }

    highlightElements.forEach(({ left, right, topEdge, bottomEdge, col }) => {
      if (col === activeSlowdownColumn) {
        left.alpha = 1.0
        right.alpha = 1.0
        topEdge.alpha = 0
        bottomEdge.alpha = 0
        left.visible = true
        right.visible = true
        topEdge.visible = false
        bottomEdge.visible = false
      } else {
        left.alpha = 0
        right.alpha = 0
        topEdge.alpha = 0
        bottomEdge.alpha = 0
      }
    })

    updateAnimation()
  }

  return {
    container,
    initialize,
    showHighlight,
    hideHighlight,
    hideAll,
    update,
    initTextures
  }
}
