import { gsap } from 'gsap'
import type { ScrollData } from './types'

export interface GSAPSpinManager {
  startColumnSpin: (
    col: number,
    startIndex: number,
    targetIndex: number,
    duration: number,
    delay: number,
    onUpdate?: (position: number, velocity: number) => void,
    onComplete?: () => void
  ) => gsap.core.Tween
  stopColumnSpin: (col: number) => void
  slowDownColumn: (col: number, newDuration: number) => boolean
  isColumnSpinning: (col: number) => boolean
  killAll: () => void
}

export function createGSAPSpinManager(): GSAPSpinManager {
  const activeSpins = new Map<number, gsap.core.Tween>()

  function startColumnSpin(
    col: number,
    startIndex: number,
    targetIndex: number,
    duration: number,
    delay: number,
    onUpdate?: (position: number, velocity: number) => void,
    onComplete?: () => void
  ): gsap.core.Tween {
    const existing = activeSpins.get(col)
    if (existing) {
      existing.kill()
    }

    const scrollData: ScrollData = {
      position: startIndex,
      velocity: 10
    }

    const tween = gsap.to(scrollData, {
      position: targetIndex,
      velocity: 0,
      duration: duration,
      delay: delay,
      ease: 'power2.out',
      onUpdate: () => {
        const remaining = targetIndex - scrollData.position
        const actualVelocity = Math.max(0.01, Math.min(remaining / 2, 10))
        scrollData.velocity = actualVelocity

        if (onUpdate) {
          onUpdate(scrollData.position, actualVelocity)
        }
      },
      onComplete: () => {
        scrollData.position = targetIndex
        scrollData.velocity = 0
        activeSpins.delete(col)

        if (onComplete) {
          onComplete()
        }
      }
    })

    activeSpins.set(col, tween)
    return tween
  }

  function stopColumnSpin(col: number): void {
    const tween = activeSpins.get(col)
    if (tween) {
      tween.kill()
      activeSpins.delete(col)
    }
  }

  function slowDownColumn(col: number, newDuration: number): boolean {
    const tween = activeSpins.get(col)
    if (tween && tween.isActive()) {
      const remainingProgress = 1 - tween.progress()
      const adjustedDuration = remainingProgress * newDuration
      tween.duration(adjustedDuration)
      return true
    }
    return false
  }

  function isColumnSpinning(col: number): boolean {
    const tween = activeSpins.get(col)
    return !!(tween && tween.isActive())
  }

  function killAll(): void {
    for (const [col, tween] of activeSpins.entries()) {
      tween.kill()
    }
    activeSpins.clear()
  }

  return {
    startColumnSpin,
    stopColumnSpin,
    slowDownColumn,
    isColumnSpinning,
    killAll
  }
}
