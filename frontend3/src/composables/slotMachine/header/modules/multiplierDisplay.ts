import { Container, Sprite, Text, Texture, ColorMatrixFilter } from 'pixi.js'
import { getMultiplierTexture } from './textureUtils'
import { getBackgroundTexture } from '@/config/spritesheet'
import type { HeaderRect, MultiplierItem } from './types'
import gsap from 'gsap'

export interface MultiplierDisplayConfig {
  inFreeSpinMode: boolean
  currentMultiplier: number
}

export interface MultiplierDisplayState {
  sprites: MultiplierItem[]
}

export function getMultiplierList(inFreeSpinMode: boolean): number[] {
  return inFreeSpinMode ? [2, 4, 6, 10] : [1, 2, 3, 5]
}

function getMultiplierWidth(
  multiplier: number,
  isActive: boolean,
  defaultWidth: number,
  activeWidth: number,
  x10Width: number
): number {
  if (multiplier === 10) return x10Width
  if (isActive) return activeWidth
  return defaultWidth
}

function calculateTotalContentWidth(
  multipliers: number[],
  displayedMult: number,
  defaultWidth: number,
  activeWidth: number,
  x10Width: number,
  gapWidth: number
): number {
  let totalWidth = (multipliers.length - 1) * gapWidth

  multipliers.forEach((mult) => {
    const isActive = mult === displayedMult
    totalWidth += getMultiplierWidth(mult, isActive, defaultWidth, activeWidth, x10Width)
  })

  return totalWidth
}

function createMultiplierSprite(
  texture: Texture,
  centerX: number,
  centerY: number,
  bg02Height: number,
  isActive: boolean
): Container {
  const container = new Container()
  container.x = centerX
  container.y = centerY

  const targetHeight = bg02Height * 0.5
  const baseScale = targetHeight / texture.height
  // Active multiplier is slightly larger
  const scale = isActive ? baseScale * 1.15 : baseScale

  if (isActive) {
    // Use active_status_bg asset for glow effect
    const activeStatusTexture = getBackgroundTexture('active_status_bg.webp')
    if (activeStatusTexture) {
      const activeStatusSprite = new Sprite(activeStatusTexture)
      activeStatusSprite.anchor.set(0.5)
      // Scale the background to fit nicely behind the multiplier
      const bgTargetSize = targetHeight * 2.0
      const bgScale = bgTargetSize / Math.max(activeStatusTexture.width, activeStatusTexture.height)
      activeStatusSprite.scale.set(bgScale)
      activeStatusSprite.blendMode = 'add'
      container.addChild(activeStatusSprite)
    }
  }

  // Main sprite
  const sprite = new Sprite(texture)
  sprite.anchor.set(0.5)
  sprite.scale.set(scale)

  // Apply visual style based on active state
  applyMultiplierStyle(sprite, isActive)
  container.addChild(sprite)

  return container
}

/**
 * Apply brightness/saturation style to multiplier sprite
 */
function applyMultiplierStyle(sprite: Sprite, isActive: boolean): void {
  if (isActive) {
    // Active: brighter and more saturated
    const brightnessFilter = new ColorMatrixFilter()
    brightnessFilter.brightness(1.5, false)
    brightnessFilter.saturate(0.4, false)
    sprite.filters = [brightnessFilter]
  }
  // Inactive multipliers keep their original appearance (no dimming)
  sprite.alpha = 1
}

function createMultiplierText(
  multiplier: number,
  centerX: number,
  centerY: number,
  isActive: boolean,
  multiplierWidth: number,
  spriteHeight: number
): Text {
  const color = isActive ? 0xffd04d : 0x666666
  const text = new Text({
    text: `X${multiplier}`,
    style: {
      fill: color,
      fontSize: Math.floor(Math.min(multiplierWidth, spriteHeight) * 0.4),
      fontWeight: 'bold',
    },
  })
  text.anchor.set(0.5)
  text.x = centerX
  text.y = centerY
  return text
}

export function buildMultipliers(
  container: Container,
  rect: HeaderRect,
  bg02Y: number,
  bg02Height: number,
  config: MultiplierDisplayConfig
): MultiplierDisplayState {
  const { x, w } = rect
  const { inFreeSpinMode, currentMultiplier } = config

  const multipliers = getMultiplierList(inFreeSpinMode)
  const sprites: MultiplierItem[] = []

  const defaultMultiplierWidth = w * 0.10
  const activeMultiplierWidth = w * 0.10
  const x10MultiplierWidth = w * 0.10
  const gapWidth = w * 0.02
  const spriteHeight = Math.floor(bg02Height * 0.8)

  // Calculate total content width and starting position
  const totalContentWidth = calculateTotalContentWidth(
    multipliers,
    currentMultiplier,
    defaultMultiplierWidth,
    activeMultiplierWidth,
    x10MultiplierWidth,
    gapWidth
  )
  let currentX = x + (w - totalContentWidth) / 2

  // Build each multiplier
  multipliers.forEach((mult) => {
    const isActive = mult === currentMultiplier
    const multiplierWidth = getMultiplierWidth(
      mult,
      isActive,
      defaultMultiplierWidth,
      activeMultiplierWidth,
      x10MultiplierWidth
    )

    const centerX = currentX + multiplierWidth / 2
    const centerY = bg02Y + bg02Height * 0.5

    // Create sprite or text
    const texture = getMultiplierTexture(mult, isActive, inFreeSpinMode)
    if (texture) {
      const sprite = createMultiplierSprite(texture, centerX, centerY, bg02Height, isActive)
      container.addChild(sprite)
      sprites.push({ sprite, multiplier: mult })
    } else {
      const text = createMultiplierText(
        mult,
        centerX,
        centerY,
        isActive,
        multiplierWidth,
        spriteHeight
      )
      container.addChild(text)
      sprites.push({ sprite: text, multiplier: mult })
    }

    currentX += multiplierWidth + gapWidth
  })

  return { sprites }
}

function updateMultiplierContainer(
  item: Container,
  multiplier: number,
  isActive: boolean,
  inFreeSpinMode: boolean,
  bg02Height: number
): void {
  const texture = getMultiplierTexture(multiplier, isActive, inFreeSpinMode)
  if (!texture) return

  const targetHeight = bg02Height * 0.5
  const baseScale = targetHeight / texture.height
  const scale = isActive ? baseScale * 1.15 : baseScale

  // Clear container and rebuild
  item.removeChildren()

  if (isActive) {
    // Use active_status_bg asset for glow effect
    const activeStatusTexture = getBackgroundTexture('active_status_bg.webp')
    if (activeStatusTexture) {
      const activeStatusSprite = new Sprite(activeStatusTexture)
      activeStatusSprite.anchor.set(0.5)
      // Scale the background to fit nicely behind the multiplier
      const bgTargetSize = targetHeight * 2.0
      const bgScale = bgTargetSize / Math.max(activeStatusTexture.width, activeStatusTexture.height)
      activeStatusSprite.scale.set(bgScale)
      activeStatusSprite.blendMode = 'add'
      item.addChild(activeStatusSprite)
    }
  }

  // Add sprite
  const sprite = new Sprite(texture)
  sprite.anchor.set(0.5)
  sprite.scale.set(scale)
  applyMultiplierStyle(sprite, isActive)
  item.addChild(sprite)
}

function updateTextStyle(
  text: Text,
  isActive: boolean,
  isX10: boolean,
  multiplierWidth: number,
  spriteHeight: number,
  defaultWidth: number,
  activeWidth: number
): void {
  const width = isX10 ? multiplierWidth : (isActive ? activeWidth : defaultWidth)
  text.style.fill = isActive ? 0xffd04d : 0x666666
  text.style.fontSize = Math.floor(Math.min(width, spriteHeight) * 0.4)
}

export function updateMultiplierSprites(
  container: Container,
  state: MultiplierDisplayState,
  config: MultiplierDisplayConfig,
  headerHeight: number
): MultiplierDisplayState {
  const { inFreeSpinMode, currentMultiplier } = config
  const displayedMult = currentMultiplier

  const w = container.getBounds().width
  const bg02Height = headerHeight * 0.5
  const spriteHeight = Math.floor(bg02Height * 0.8)

  const defaultMultiplierWidth = w * 0.10
  const activeMultiplierWidth = w * 0.10
  const x10MultiplierWidth = w * 0.10

  // Update each multiplier sprite
  state.sprites.forEach((item) => {
    const { sprite, multiplier } = item
    const isActive = multiplier === displayedMult
    const isX10 = multiplier === 10

    // Skip if sprite is no longer a child of container (happens during transitions)
    if (!sprite.parent || sprite.parent !== container) {
      return
    }

    // Update sprite or text appearance
    if (sprite instanceof Container && !(sprite instanceof Sprite) && !(sprite instanceof Text)) {
      updateMultiplierContainer(sprite, multiplier, isActive, inFreeSpinMode, bg02Height)
    } else if (sprite instanceof Text) {
      updateTextStyle(
        sprite,
        isActive,
        isX10,
        x10MultiplierWidth,
        spriteHeight,
        defaultMultiplierWidth,
        activeMultiplierWidth
      )
    }
  })

  return { sprites: state.sprites }
}

export function transitionMultipliers(
  container: Container,
  state: MultiplierDisplayState,
  rect: HeaderRect,
  config: MultiplierDisplayConfig,
  bg02Y: number,
  bg02Height: number,
  onComplete: (newState: MultiplierDisplayState) => void
): void {
  const oldSprites = [...state.sprites]
  const spriteObjects = oldSprites.map(item => item.sprite)

  gsap.to(spriteObjects, {
    alpha: 0,
    scale: 0.8,
    duration: 0.3,
    ease: 'power2.in',
    onComplete: () => {
      oldSprites.forEach(item => {
        if (item.sprite.parent) {
          container.removeChild(item.sprite)
        }
      })
    }
  })

  setTimeout(() => {
    const newState = buildMultipliers(container, rect, bg02Y, bg02Height, config)

    const newSprites = newState.sprites.map(item => item.sprite)
    newSprites.forEach(sprite => {
      sprite.alpha = 0
      const originalScaleX = sprite.scale.x
      const originalScaleY = sprite.scale.y
      sprite.scale.x = originalScaleX * 1.2
      sprite.scale.y = originalScaleY * 1.2
    })

    newSprites.forEach((sprite, index) => {
      const originalScaleX = sprite.scale.x / 1.2
      const originalScaleY = sprite.scale.y / 1.2

      // Animate alpha separately
      gsap.to(sprite, {
        alpha: 1,
        duration: 0.4,
        ease: 'back.out(1.4)',
        delay: index * 0.05
      })

      // Animate scale directly on the scale object
      gsap.to(sprite.scale, {
        x: originalScaleX,
        y: originalScaleY,
        duration: 0.4,
        ease: 'back.out(1.4)',
        delay: index * 0.05,
        onComplete: () => {
          if (index === newSprites.length - 1) {
            onComplete(newState)
          }
        }
      })
    })
  }, 150)
}
