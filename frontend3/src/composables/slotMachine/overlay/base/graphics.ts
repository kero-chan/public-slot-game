// @ts-nocheck
import { Container, Graphics, Sprite, Text, Texture } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import gsap from 'gsap'
import { ASSETS } from '@/config/assets'
import { getGlyphSprite, getGlyphTexture } from '@/config/spritesheet'
import { formatWinAmount } from '@/composables/slotMachine/footer/modules/textureUtils'

/**
 * Recursively destroy a display object and all its children while preserving textures
 */
function destroyRecursive(obj: any): void {
  // First destroy all children recursively
  if (obj.children && obj.children.length > 0) {
    while (obj.children.length > 0) {
      const child = obj.children[0]
      obj.removeChild(child)
      destroyRecursive(child)
    }
  }

  // Then destroy the object itself, preserving textures
  if (obj.destroy) {
    // Use explicit options to ensure textures are never destroyed
    obj.destroy({ children: false, texture: false, textureSource: false })
  }
}

/**
 * Properly clear a container by destroying all children and killing GSAP tweens
 * This prevents memory leaks from orphaned display objects and running tweens
 */
export function clearContainer(container: Container): void {
  // Kill GSAP tweens on all children recursively
  const killTweensRecursive = (obj: Container) => {
    gsap.killTweensOf(obj)
    gsap.killTweensOf(obj.scale)
    gsap.killTweensOf(obj.position)
    if (obj.children) {
      for (const child of obj.children) {
        if (child instanceof Container) {
          killTweensRecursive(child)
        } else {
          gsap.killTweensOf(child)
          gsap.killTweensOf(child.scale)
        }
      }
    }
  }

  killTweensRecursive(container)

  // Destroy all children recursively while preserving textures
  while (container.children.length > 0) {
    const child = container.children[0]
    container.removeChild(child)
    destroyRecursive(child)
  }
}

/**
 * Create a dark overlay background
 */
export function createDarkOverlay(
  container: Container,
  width: number,
  height: number,
  color = 0x000000,
  alpha = 0.6,
  animate = true
): Graphics {
  const bg = new Graphics()
  bg.rect(0, 0, width, height)
  bg.fill({ color, alpha })

  if (animate) {
    bg.alpha = 0
    gsap.to(bg, { alpha: 1, duration: 0.3, ease: 'power2.out' })
  }

  container.addChild(bg)
  return bg
}

/**
 * Create rotating starburst rays background
 */
export function createStarburstRays(
  container: Container,
  centerX: number,
  centerY: number,
  rayCount = 24,
  color = 0xFFD700,
  alpha = 0.1
): Container {
  const raysContainer = new Container()
  raysContainer.x = centerX
  raysContainer.y = centerY

  const length = Math.max(centerX * 2, centerY * 2)

  for (let i = 0; i < rayCount; i++) {
    const ray = new Graphics()
    const angle = (i / rayCount) * Math.PI * 2

    ray.moveTo(0, 0)
    ray.lineTo(Math.cos(angle - 0.03) * length, Math.sin(angle - 0.03) * length)
    ray.lineTo(Math.cos(angle + 0.03) * length, Math.sin(angle + 0.03) * length)
    ray.closePath()
    ray.fill({ color, alpha })
    ray.blendMode = BLEND_MODES.ADD as any

    raysContainer.addChild(ray)
  }

  // Slow rotation animation
  gsap.to(raysContainer, {
    rotation: Math.PI * 2,
    duration: 30,
    repeat: -1,
    ease: 'none'
  })

  container.addChild(raysContainer)
  return raysContainer
}

/**
 * Create a glowing panel with border
 */
export function createGlowPanel(
  container: Container,
  centerX: number,
  centerY: number,
  width: number,
  height: number,
  options: {
    glowColor?: number
    bgColor?: number
    borderWidth?: number
    cornerRadius?: number
    glowLayers?: number
  } = {}
): Graphics {
  const {
    glowColor = 0xFFD700,
    bgColor = 0x1a0a2a,
    borderWidth = 4,
    cornerRadius = 20,
    glowLayers = 4
  } = options

  // Outer glow layers
  for (let i = glowLayers; i > 0; i--) {
    const glow = new Graphics()
    const padding = i * 20
    const alpha = 0.15 * (1 - i / (glowLayers + 1))
    glow.roundRect(
      centerX - width / 2 - padding,
      centerY - height / 2 - padding,
      width + padding * 2,
      height + padding * 2,
      cornerRadius + 10
    )
    glow.fill({ color: glowColor, alpha })
    glow.blendMode = BLEND_MODES.ADD as any
    container.addChild(glow)
  }

  // Main panel
  const panel = new Graphics()
  panel.roundRect(centerX - width / 2, centerY - height / 2, width, height, cornerRadius)
  panel.fill({ color: bgColor, alpha: 0.9 })

  // Border
  panel.roundRect(centerX - width / 2, centerY - height / 2, width, height, cornerRadius)
  panel.stroke({ color: glowColor, width: borderWidth, alpha: 0.9 })

  // Inner highlight
  panel.roundRect(
    centerX - width / 2 + 4,
    centerY - height / 2 + 4,
    width - 8,
    height - 8,
    cornerRadius - 4
  )
  panel.stroke({ color: glowColor, width: 2, alpha: 0.5 })

  container.addChild(panel)
  return panel
}

/**
 * Create number display using gold glyph images
 */
export function createNumberDisplay(
  amount: number,
  prefixOrOptions?: string | {
    scale?: number
    letterSpacing?: number
    prefix?: string
    baseFontSize?: number
  }
): Container {
  // Handle both old signature (amount, options) and new signature (amount, prefix)
  let scale = 1
  let letterSpacing = 0.95
  let prefix = ''
  let baseFontSize: number | undefined

  if (typeof prefixOrOptions === 'string') {
    prefix = prefixOrOptions
  } else if (prefixOrOptions) {
    scale = prefixOrOptions.scale ?? 1
    letterSpacing = prefixOrOptions.letterSpacing ?? 0.95
    prefix = prefixOrOptions.prefix ?? ''
    baseFontSize = prefixOrOptions.baseFontSize
  }

  const numContainer = new Container()
  const formatted = prefix + formatWinAmount(amount)
  const digits = formatted.split('')

  // Get reference height from first digit
  let referenceHeight = 0
  let originalHeight = 0
  for (const d of digits) {
    if (d >= '0' && d <= '9') {
      const tex = getGlyphTexture(`glyph_${d}.webp`)
      if (tex) {
        // Get height directly from texture instead of creating a sprite
        originalHeight = tex.height
        referenceHeight = baseFontSize ? baseFontSize : originalHeight
        break
      }
    }
  }

  const glyphScale = baseFontSize && originalHeight ? (referenceHeight / originalHeight) : 1
  
  // Scale multiplier for punctuation (comma and dot) - visible but smaller than digits
  const PUNCTUATION_SCALE_MULTIPLIER = 0.5

  let offsetX = 0
  for (const d of digits) {
    if (d === '+') {
      // Create styled text for '+' character
      const plusText = new Text({
        text: '+',
        style: {
          fontFamily: 'Arial Black, sans-serif',
          fontSize: referenceHeight * 0.8,
          fontWeight: 'bold',
          fill: '#FFD700',
          stroke: { color: '#8B4513', width: 4 }
        }
      })
      plusText.x = offsetX
      plusText.y = referenceHeight * 0.1
      numContainer.addChild(plusText)
      offsetX += plusText.width * 0.9
    } else {
      const isPunctuation = d === ',' || d === '.'
      const key =
        d === ','
          ? 'glyph_comma.webp'
          : d === '.'
          ? 'glyph_dot.webp'
          : `glyph_${d}.webp`
      const sprite = getGlyphSprite(key)
      if (sprite) {
        // Use linear filtering with mipmaps for high-quality rendering on retina displays
        if (sprite.texture?.source) {
          sprite.texture.source.scaleMode = 'linear'
          sprite.texture.source.autoGenerateMipmaps = true
        }
        sprite.anchor.set(0)
        const finalScale = isPunctuation ? glyphScale * PUNCTUATION_SCALE_MULTIPLIER : glyphScale
        sprite.scale.set(finalScale)
        sprite.x = offsetX
        if (isPunctuation) {
          const baselineY = referenceHeight * 0.85 // Baseline position (85% from top)
          const punctuationHeight = sprite.height * finalScale
          sprite.y = baselineY - punctuationHeight // Position so bottom of punctuation aligns with baseline
        } else {
          sprite.y = 0 // Digits stay at top
        }
        numContainer.addChild(sprite)
        // sprite.width already includes the scale, so don't multiply by glyphScale again
        offsetX += sprite.width * letterSpacing
      }
    }
  }

  if (scale !== 1) {
    numContainer.scale.set(scale)
  }

  return numContainer
}

/**
 * Create text with Chinese/English support
 */
export function createStyledText(
  text: string,
  options: {
    fontSize?: number
    color?: number | string
    strokeColor?: number | string
    strokeWidth?: number
    dropShadow?: boolean
  } = {}
): Text {
  const {
    fontSize = 32,
    color = '#FFD700',
    strokeColor = '#8B4513',
    strokeWidth = 4,
    dropShadow = true
  } = options

  const style: any = {
    fontFamily: 'Arial Black, sans-serif',
    fontSize,
    fontWeight: 'bold',
    fill: color,
    stroke: { color: strokeColor, width: strokeWidth }
  }

  if (dropShadow) {
    style.dropShadow = {
      color: 0x000000,
      blur: 4,
      distance: 2,
      alpha: 0.7
    }
  }

  const textObj = new Text({ text, style })
  textObj.anchor.set(0.5)
  return textObj
}

/**
 * Create sprite from asset with fallback
 */
export function createAssetSprite(
  assetKey: string,
  fallbackText?: string
): Sprite | Text | null {
  const imgSrc = ASSETS.loadedImages?.[assetKey] || ASSETS.imagePaths?.[assetKey]

  if (imgSrc) {
    const texture = imgSrc instanceof Texture ? imgSrc : Texture.from(imgSrc)
    const sprite = new Sprite(texture)
    sprite.anchor.set(0.5)
    return sprite
  }

  if (fallbackText) {
    return createStyledText(fallbackText, { fontSize: 64 })
  }

  return null
}

/**
 * Create animated corner sparkles
 */
export function createCornerSparkles(
  container: Container,
  corners: Array<{ x: number; y: number }>,
  size = 15
): void {
  corners.forEach((corner, i) => {
    const star = new Graphics()

    // 4-point star shape
    star.moveTo(size, 0)
    for (let j = 0; j < 8; j++) {
      const angle = (j / 8) * Math.PI * 2
      const r = j % 2 === 0 ? size : size * 0.4
      star.lineTo(Math.cos(angle) * r, Math.sin(angle) * r)
    }
    star.closePath()
    star.fill({ color: 0xFFFFFF, alpha: 0.9 })

    star.x = corner.x
    star.y = corner.y
    star.blendMode = BLEND_MODES.ADD as any
    container.addChild(star)

    // Rotation animation
    gsap.to(star, {
      rotation: Math.PI * 2,
      duration: 3,
      repeat: -1,
      ease: 'none'
    })

    // Pulse animation
    gsap.to(star, {
      alpha: 0.4,
      duration: 0.5 + i * 0.1,
      repeat: -1,
      yoyo: true,
      ease: 'sine.inOut'
    })
  })
}

/**
 * Create a button with glow effect
 */
export function createGlowButton(
  centerX: number,
  centerY: number,
  textKey: string,
  onClick: () => void,
  options: {
    width?: number
    height?: number
    bgColor?: number
    glowColor?: number
  } = {}
): Container {
  const {
    width = 200,
    height = 60,
    bgColor = 0xdc143c,
    glowColor = 0xFFD700
  } = options

  const btnContainer = new Container()
  btnContainer.x = centerX
  btnContainer.y = centerY

  // Glow
  const glow = new Graphics()
  glow.roundRect(-width / 2 - 10, -height / 2 - 10, width + 20, height + 20, 20)
  glow.fill({ color: glowColor, alpha: 0.3 })
  glow.blendMode = BLEND_MODES.ADD as any
  btnContainer.addChild(glow)

  // Background
  const bg = new Graphics()
  bg.roundRect(-width / 2, -height / 2, width, height, 15)
  bg.fill({ color: bgColor })
  bg.stroke({ color: glowColor, width: 3 })
  btnContainer.addChild(bg)

  // Text
  const btnText = getGlyphSprite(textKey)
  btnContainer.addChild(btnText)

  // Interactivity
  btnContainer.eventMode = 'static'
  btnContainer

  btnContainer.on('pointerover', () => {
    gsap.to(btnContainer.scale, { x: 1.1, y: 1.1, duration: 0.2 })
  })

  btnContainer.on('pointerout', () => {
    gsap.to(btnContainer.scale, { x: 1, y: 1, duration: 0.2 })
  })

  btnContainer.on('pointerdown', onClick)

  // Pulsing glow
  gsap.to(glow, {
    alpha: 0.5,
    duration: 0.6,
    repeat: -1,
    yoyo: true,
    ease: 'sine.inOut'
  })

  return btnContainer
}
