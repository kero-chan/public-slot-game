import type { GridState } from '@/types/global'

export interface AnimObject {
  col: number
  position: number
  targetIndex: number
  isSlowingDown: boolean
  hasSlowedDown: boolean
}

export interface SpinConfig {
  cols: number
  totalRows: number
  stripLength: number
  maxBonusPerColumn: number
  winCheckStartRow: number
  winCheckEndRow: number
  bufferOffset: number
}

export interface SpinAnimationState {
  stoppedColumns: Set<number>
  columnTweens: Map<number, any>
  firstBonusColumn: number
  currentSlowdownColumn: number
  backendInjected: boolean
  animObjects: AnimObject[]
  targetIndexes: number[]
}

export interface SpinAnimationDeps {
  gridState: GridState
  backendTargetGrid: string[][] | null
  gameStore: any
  timingStore: any
  playEffect: (name: string) => void
  getRandomSymbol: (opts: any) => string
  gsap: any
}
