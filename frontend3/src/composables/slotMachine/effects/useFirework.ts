import { AnimatedSprite, Texture, Rectangle, type Container } from 'pixi.js'
import { ASSETS } from '@/config/assets'

/**
 * Options for creating a firework sprite
 */
export interface FireworkSpriteOptions {
  frameWidth?: number
  frameHeight?: number
  frameCount?: number
  columns?: number
  animationSpeed?: number
  loop?: boolean
}

/**
 * Options for playing a firework effect
 */
export interface FireworkEffectOptions extends FireworkSpriteOptions {
  scale?: number
  onComplete?: (() => void) | null
}

/**
 * Options for creating a looping firework
 */
export interface LoopingFireworkOptions extends FireworkSpriteOptions {
  scale?: number
  autoPlay?: boolean
}

/**
 * Create a firework animation sprite
 * @param options - Configuration options
 * @returns Animated sprite or null if texture not available
 */
export function createFireworkSprite(options: FireworkSpriteOptions = {}): AnimatedSprite | null {
  const {
    frameWidth = 128,
    frameHeight = 128,
    frameCount = 16,
    columns = 4,
    animationSpeed = 0.2, // Slower default for smoother animation
    loop = true
  } = options

  const fireworkTexture = ASSETS.loadedImages?.firework
  if (!fireworkTexture) {
    console.warn('Firework texture not loaded')
    return null
  }

  // Create frames from spritesheet
  const frames: Texture[] = []
  for (let i = 0; i < frameCount; i++) {
    const col = i % columns
    const row = Math.floor(i / columns)
    const x = col * frameWidth
    const y = row * frameHeight

    const frame = new Texture({
      source: fireworkTexture.source,
      frame: new Rectangle(x, y, frameWidth, frameHeight)
    })
    frames.push(frame)
  }

  // Create animated sprite
  const animatedSprite = new AnimatedSprite(frames)
  animatedSprite.anchor.set(0.5)
  animatedSprite.animationSpeed = animationSpeed
  animatedSprite.loop = loop

  // Apply blend mode for black background
  animatedSprite.blendMode = 'screen'

  return animatedSprite
}

/**
 * Create a firework effect that auto-plays and removes itself when done
 * @param container - Parent container to add the firework to
 * @param x - X position
 * @param y - Y position
 * @param options - Additional options (scale, etc.)
 */
export function playFireworkEffect(
  container: Container,
  x: number,
  y: number,
  options: FireworkEffectOptions = {}
): AnimatedSprite | undefined {
  const {
    scale = 1,
    onComplete = null,
    ...spriteOptions
  } = options

  const firework = createFireworkSprite({
    loop: false, // Don't loop for one-time effect
    ...spriteOptions
  })

  if (!firework) return

  firework.x = x
  firework.y = y
  firework.scale.set(scale)

  // Add to container
  container.addChild(firework)

  // Handle completion
  firework.onComplete = () => {
    if (onComplete) onComplete()
    container.removeChild(firework)
    firework.destroy()
  }

  // Start animation
  firework.play()

  return firework
}

/**
 * Create a looping firework sprite for continuous effect
 * @param options - Configuration options
 * @returns Animated sprite ready to be added to container
 */
export function createLoopingFirework(options: LoopingFireworkOptions = {}): AnimatedSprite | null {
  const {
    scale = 1,
    autoPlay = true,
    ...spriteOptions
  } = options

  const firework = createFireworkSprite({
    loop: true,
    ...spriteOptions
  })

  if (!firework) return null

  firework.scale.set(scale)

  if (autoPlay) {
    firework.play()
  }

  return firework
}
