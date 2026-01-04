import { useGameStore, useTimingStore } from '@/stores'
import type { GridState, WinCombination, CascadeData } from '@/types/global'
import { audioEvents, AUDIO_EVENTS } from '@/composables/audioEventBus'
import { howlerAudio } from '@/composables/useHowlerAudio'

/**
 * Play tile break sound with pitch randomization (0.75-1.25)
 */
function playTileBreakSound(): void {
  const howl = howlerAudio.getHowl('tile_break')
  if (!howl) return

  // Apply pitch randomization 0.75-1.25 to prevent audio fatigue
  const randomPitch = 0.75 + Math.random() * 0.5
  howl.rate(randomPitch)

  audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { audioKey: 'tile_break', volume: 0.6 })
}

export interface CascadeAnimationConfig {
  gridState: GridState
  render: () => void
  getCurrentCascade: () => CascadeData | null
  currentCascadeIndex: number
}

export interface UseCascadeAnimation {
  computeGoldTransformedPositions: (wins: WinCombination[]) => void
  animateDisappear: (wins: WinCombination[]) => Promise<void>
  cascadeSymbols: (wins: WinCombination[]) => Promise<void>
}

export function useCascadeAnimation(config: CascadeAnimationConfig): UseCascadeAnimation {
  const gameStore = useGameStore()
  const timingStore = useTimingStore()
  const { gridState, render, getCurrentCascade } = config

  /**
   * Compute gold transformed positions from backend's is_gold_to_wild flag
   * This should be called BEFORE transformGoldTilesToWild so the positions are available
   * for both the transformation animation and the cascade exclusion
   *
   * The backend marks positions with is_gold_to_wild=true when a gold tile should transform to wild
   */
  function computeGoldTransformedPositions(wins: WinCombination[]): void {
    // Read gold transformed positions from backend's is_gold_to_wild flag
    const goldTransformedPositions = new Set<string>()
    wins.forEach(win => {
      win.positions.forEach(pos => {
        if (pos.is_gold_to_wild) {
          goldTransformedPositions.add(`${pos.reel},${pos.row}`)
        }
      })
    })
    gridState.goldTransformedPositions = goldTransformedPositions
  }

  function animateDisappear(wins: WinCombination[]): Promise<void> {
    const DISAPPEAR_MS = timingStore.DISAPPEAR_WAIT

    // Collect positions to disappear, excluding gold-to-wild transformations
    // Gold tiles that transform to wild should NOT disappear - they flip in place
    const positions: Array<[number, number]> = []
    wins.forEach(win => {
      win.positions.forEach(pos => {
        // Skip positions that are gold-to-wild transformations
        if (!pos.is_gold_to_wild) {
          positions.push([pos.reel, pos.row])
        }
      })
    })

    // Play tile break sound when tiles are removed
    if (positions.length > 0) {
      playTileBreakSound()
    }

    gridState.disappearPositions = new Set(positions.map(([c, r]) => `${c},${r}`))
    gridState.disappearAnim = { start: Date.now(), duration: DISAPPEAR_MS }

    return new Promise(resolve => setTimeout(resolve, DISAPPEAR_MS))
  }

  function animateCascade(): Promise<void> {
    const startTime = Date.now()
    const MAX_WAIT = timingStore.CASCADE_MAX_WAIT

    return new Promise(resolve => {
      const animate = (): void => {
        render()
        const elapsed = Date.now() - startTime

        if (gridState.isDropAnimating && elapsed < MAX_WAIT) {
          requestAnimationFrame(animate)
        } else {
          resolve()
        }
      }
      animate()
    })
  }

  async function cascadeSymbols(wins: WinCombination[]): Promise<void> {
    const currentCascade = getCurrentCascade()

    if (!currentCascade?.grid_after) {
      console.error('No cascade grid_after data from backend')
      return
    }

    if (!Array.isArray(currentCascade.grid_after) || currentCascade.grid_after.length !== 5) {
      console.error('Invalid grid_after structure')
      return
    }

    for (let col = 0; col < 5; col++) {
      if (!Array.isArray(currentCascade.grid_after[col]) || currentCascade.grid_after[col].length < 6) {
        console.error(`Invalid column ${col} in grid_after`)
        return
      }
    }

    gridState.previousGridSnapshot = gridState.grid.map(col => [...col])

    // Use goldTransformedPositions computed earlier (by computeGoldTransformedPositions)
    // These positions should NOT cascade (they stay in place after transformation)
    const goldTransformedPositions = gridState.goldTransformedPositions || new Set<string>()

    // Collect removed positions but exclude gold-to-wild transformations
    const removedPositions = new Set<string>()
    wins.forEach(win => {
      win.positions.forEach(pos => {
        const posKey = `${pos.reel},${pos.row}`
        // Only add to removed if not a gold transformation
        if (!goldTransformedPositions.has(posKey)) {
          removedPositions.add(posKey)
        }
      })
    })
    gridState.lastRemovedPositions = removedPositions

    const visibleGrid: string[][] = []
    const bufferRows: string[][] = []

    for (let col = 0; col < 5; col++) {
      const columnData = currentCascade.grid_after[col]
      const columnLength = columnData.length

      visibleGrid[col] = []
      const expectedRows = Math.min(10, columnLength)
      for (let row = 0; row < expectedRows; row++) {
        const symbol = columnData[row]
        visibleGrid[col][row] = (symbol && typeof symbol === 'string') ? symbol : 'fa'
      }

      bufferRows[col] = []
      if (columnLength > 10) {
        for (let row = 10; row < columnLength; row++) {
          if (columnData[row]) bufferRows[col].push(columnData[row])
        }
      }
    }

    gridState.bufferRows = bufferRows

    const backendRowCount = visibleGrid[0]?.length || 0
    for (let col = 0; col < 5; col++) {
      for (let row = 0; row < backendRowCount; row++) {
        gridState.grid[col][row] = visibleGrid[col][row]
      }
    }
    gridState.lastCascadeTime = Date.now()

    await animateCascade()

    // Log cascade verification
    const cascadeNum = currentCascade.cascade_number ?? '?'
    console.log(`🎰 CASCADE #${cascadeNum} COMPLETE - Verification:`)
    console.log('  Backend grid_after:', JSON.stringify(currentCascade.grid_after))
    console.log('  Frontend grid:     ', JSON.stringify(gridState.grid))

    // Check for mismatches
    let mismatches = 0
    for (let col = 0; col < 5; col++) {
      for (let row = 0; row < backendRowCount; row++) {
        const backendSymbol = currentCascade.grid_after[col]?.[row]
        const frontendSymbol = gridState.grid[col]?.[row]
        if (backendSymbol !== frontendSymbol) {
          console.warn(`  ❌ Mismatch at [${col},${row}]: backend="${backendSymbol}" frontend="${frontendSymbol}"`)
          mismatches++
        }
      }
    }
    if (mismatches === 0) {
      console.log('  ✅ Grid matches backend perfectly!')
    } else {
      console.error(`  ⚠️ ${mismatches} mismatches found!`)
    }

    // Log win amount verification
    if (currentCascade.total_cascade_win > 0) {
      const serverCascadeWin = currentCascade.total_cascade_win
      console.log(`  💰 Win Amount - Server cascade_win: ${serverCascadeWin}`)
      gameStore.setCurrentWin(serverCascadeWin)
      gameStore.addCredits(serverCascadeWin)
      gameStore.setShowAmountNotification(true)
    }
  }

  return {
    computeGoldTransformedPositions,
    animateDisappear,
    cascadeSymbols
  }
}
