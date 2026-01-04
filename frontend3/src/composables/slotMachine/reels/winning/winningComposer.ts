import { Container } from 'pixi.js'
import type { Sprite } from 'pixi.js'
import { createWinningFrameEffect, type WinningFrameEffect } from './frameEffects'

/**
 * Winning frame manager interface
 */
export interface WinningFrameManager {
  container: Container
  updateFrame: (key: string, sprite: Sprite, highlight: boolean, x: number, y: number, isBonus?: boolean, symbol?: string) => void
  update: (timestamp: number) => void
  cleanup: (usedKeys: Set<string>) => void
}

/**
 * Create a manager for winning frames that doesn't attach to sprites
 */
export function createWinningFrameManager(onNewFrameCallback: ((key: string) => void) | null = null): WinningFrameManager {
  const container = new Container()
  const activeFrames = new Set<string>()
  const effectCache = new Map<string, WinningFrameEffect>()
  const effectData = new Map<string, { sprite: Sprite; x: number; y: number; isBonus: boolean; symbol?: string }>()

  function updateFrame(key: string, sprite: Sprite, highlight: boolean, x: number, y: number, isBonus: boolean = false, symbol?: string): void {
    if (!sprite) return

    if (highlight) {
      const isNewFrame = !activeFrames.has(key)

      // Create effect if it doesn't exist
      if (!effectCache.has(key)) {
        const effect = createWinningFrameEffect()
        container.addChild(effect.container)
        effectCache.set(key, effect)
      }

      const effect = effectCache.get(key)!
      effectData.set(key, { sprite, x, y, isBonus, symbol })
      effect.show()

      if (isNewFrame) {
        activeFrames.add(key)
        if (onNewFrameCallback) {
          onNewFrameCallback(key)
        }
      }
    } else {
      const effect = effectCache.get(key)
      if (effect) {
        effect.hide()
      }
      effectData.delete(key)
      activeFrames.delete(key)
    }
  }

  function update(timestamp: number): void {
    for (const [key, data] of effectData.entries()) {
      const effect = effectCache.get(key)
      if (effect) {
        effect.update(timestamp, data.sprite, data.x, data.y, data.isBonus, data.symbol)
      }
    }
  }

  function cleanup(usedKeys: Set<string>): void {
    for (const [key, effect] of effectCache.entries()) {
      if (!usedKeys.has(key)) {
        effect.destroy()
        effectCache.delete(key)
        effectData.delete(key)
        activeFrames.delete(key)
      }
    }
  }

  return { container, updateFrame, update, cleanup }
}
