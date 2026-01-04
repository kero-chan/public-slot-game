import { Container, Sprite } from 'pixi.js'
import { gsap } from 'gsap'
import { getTextureForSymbol } from '../textures'
import type { SymbolSprite, ReelSpinData } from './types'
import type { GridState } from '@/types/global'

export function createReelScroller(
  reelContainers: Container[],
  reelSpinData: Map<number, ReelSpinData>,
  gridState: GridState
) {
  function startGSAPReelScroll(
    col: number,
    strip: string[],
    xPos: number,
    startY: number,
    stepY: number,
    spriteWidth: number,
    spriteHeight: number,
    targetDistance: number,
    duration: number,
    delay: number,
    onComplete?: (col: number) => void
  ): gsap.core.Timeline {
    const reelContainer = reelContainers[col]

    // Clear previous sprites
    while (reelContainer.children.length > 0) {
      const child = reelContainer.children[0]
      reelContainer.removeChild(child)
      if (child.destroy) child.destroy({ children: true, texture: false })
    }

    // Create sprites for the entire strip
    const stripSprites: Sprite[] = []
    const stripLength = strip.length

    for (let i = 0; i < stripLength; i++) {
      const symbol = strip[i]
      const tex = getTextureForSymbol(symbol)
      if (!tex) continue

      const sprite = new Sprite(tex) as SymbolSprite
      sprite.anchor.set(0.5, 0.5)
      sprite.x = xPos
      sprite.y = i * stepY

      const scaleX = spriteWidth / tex.width
      const scaleY = spriteHeight / tex.height
      sprite.scale.set(scaleX, scaleY)
      sprite._symbolData = symbol

      reelContainer.addChild(sprite)
      stripSprites.push(sprite)
    }

    reelContainer.y = startY

    const scrollDistance = targetDistance * stepY

    const timeline = gsap.timeline({
      onComplete: () => {
        reelSpinData.delete(col)
        if (onComplete) onComplete(col)
      }
    })

    timeline.to(reelContainer, {
      y: `+=${scrollDistance}`,
      duration: duration,
      delay: delay,
      ease: 'power2.out',
      onUpdate: () => {
        const progress = timeline.progress()
        const remaining = (1 - progress) * scrollDistance
        const velocity = remaining > 10 ? 10 : Math.max(0.01, remaining / 2)

        if (gridState.spinVelocities) {
          gridState.spinVelocities[col] = velocity
        }
      }
    })

    reelSpinData.set(col, {
      gsapTimeline: timeline,
      isSpinning: true,
      stripSprites
    })

    return timeline
  }

  function stopGSAPReelScroll(col: number): void {
    const spinData = reelSpinData.get(col)
    if (spinData) {
      if (spinData.gsapTimeline) {
        spinData.gsapTimeline.kill()
      }
      const reelContainer = reelContainers[col]
      while (reelContainer.children.length > 0) {
        const child = reelContainer.children[0]
        reelContainer.removeChild(child)
        if (child.destroy) child.destroy({ children: true, texture: false })
      }
      reelContainer.y = 0
      reelSpinData.delete(col)
    }
  }

  function stopAllGSAPReelScrolls(): void {
    for (let col = 0; col < reelContainers.length; col++) {
      stopGSAPReelScroll(col)
    }
  }

  return {
    startGSAPReelScroll,
    stopGSAPReelScroll,
    stopAllGSAPReelScrolls
  }
}
