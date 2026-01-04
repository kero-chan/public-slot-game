import { isBonusTile } from '@/utils/tileHelpers'
import type { GridState } from '@/types/global'
import type { SpinConfig } from './types'

export function buildReelStrips(
  gridState: GridState,
  config: SpinConfig,
  getRandomSymbol: (opts: any) => string
): void {
  const { cols, totalRows, stripLength, maxBonusPerColumn, winCheckStartRow, winCheckEndRow, bufferOffset } = config
  const protectedPositions = new Map<number, Set<number>>()

  for (let col = 0; col < cols; col++) {
    const strip: string[] = Array(stripLength).fill(null).map(() =>
      getRandomSymbol({ col, allowGold: true, allowBonus: true })
    )

    const reelTopAtStart = 0
    for (let row = 0; row < totalRows; row++) {
      const stripIdx = ((reelTopAtStart - row) % stripLength + stripLength) % stripLength
      const currentSymbol = gridState.grid?.[col]?.[row]
      if (currentSymbol && typeof currentSymbol === 'string') {
        strip[stripIdx] = currentSymbol
      }
    }

    gridState.reelStrips[col] = strip

    const protectedSet = new Set<number>()
    for (let row = 0; row < totalRows; row++) {
      const stripIdx = ((reelTopAtStart - row) % stripLength + stripLength) % stripLength
      protectedSet.add(stripIdx)
    }
    protectedPositions.set(col, protectedSet)
  }

  enforceBonusLimits(gridState, config, protectedPositions, getRandomSymbol)
  gridState.reelStrips = [...gridState.reelStrips]
}

function enforceBonusLimits(
  gridState: GridState,
  config: SpinConfig,
  protectedPositions: Map<number, Set<number>>,
  getRandomSymbol: (opts: any) => string
): void {
  const { cols, stripLength, maxBonusPerColumn, winCheckStartRow, winCheckEndRow, bufferOffset } = config

  for (let col = 0; col < cols; col++) {
    const strip = gridState.reelStrips[col]
    const protectedSet = protectedPositions.get(col)!

    for (let reelTop = 0; reelTop < stripLength; reelTop++) {
      const bonusPositions: Array<{ idx: number; gridRow: number }> = []

      for (let gridRow = winCheckStartRow; gridRow <= winCheckEndRow; gridRow++) {
        const stripIdx = ((reelTop - gridRow) % strip.length + strip.length) % strip.length
        if (isBonusTile(strip[stripIdx])) {
          bonusPositions.push({ idx: stripIdx, gridRow })
        }
      }

      if (bonusPositions.length > maxBonusPerColumn) {
        let replaced = 0
        const requiredReplacements = bonusPositions.length - maxBonusPerColumn

        for (let i = bonusPositions.length - 1; i >= 0 && replaced < requiredReplacements; i--) {
          const { idx, gridRow } = bonusPositions[i]
          if (protectedSet.has(idx)) continue

          const visualRow = gridRow - bufferOffset
          strip[idx] = getRandomSymbol({ col, visualRow, allowGold: true, allowBonus: false })
          replaced++
        }
      }
    }
  }
}
