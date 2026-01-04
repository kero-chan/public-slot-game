import { Container } from 'pixi.js'

export interface FreeSpinHeaderDisplay {
  container: Container
  update: (freeSpins: number, inFreeSpinMode: boolean, gridX: number, gridY: number, gridH: number, gameW: number) => void
  clear: () => void
}

/**
 * Creates a permanent free spin counter display pinned to the top of the grid frame
 */
export function createFreeSpinHeaderDisplay(): FreeSpinHeaderDisplay {
  const container = new Container()
  container.zIndex = 1000

  let lastFreeSpins = -1
  let lastInFreeSpinMode = false

  /**
   * Update the free spin counter display
   */
  function update(freeSpins: number, inFreeSpinMode: boolean, gridX: number, gridY: number, gridH: number, gameW: number): void {
    // Hide when not in free spin mode
    if (!inFreeSpinMode) {
      container.visible = false
      lastInFreeSpinMode = false
      return
    }

    container.visible = true

    // Only rebuild if values changed
    if (freeSpins === lastFreeSpins && inFreeSpinMode === lastInFreeSpinMode) {
      return
    }

    lastFreeSpins = freeSpins
    lastInFreeSpinMode = inFreeSpinMode

    // Clear existing sprites
    container.removeChildren()

    if (freeSpins <= 0) {
      return
    }

    // No background sprites needed - free spin counter is now shown in footer
    // Keep the header empty/hidden during free spins
  }

  /**
   * Clear the display
   */
  function clear(): void {
    container.removeChildren()
    lastFreeSpins = -1
    lastInFreeSpinMode = false
  }

  return {
    container,
    update,
    clear
  }
}
