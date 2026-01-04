import { Container, Graphics } from 'pixi.js'

/**
 * Graphics-based lightning link effect that connects winning tiles
 * Uses multiple layered beams for a glowing effect
 * Each symbol type has unique colors matching their tile design
 *
 * Based on frontend2's implementation with symbol-specific colors added
 */

export interface TilePosition {
  x: number
  y: number
  col: number
  row: number
  symbol?: string
}

export interface ImageLightningLinksManager {
  container: Container
  update: (timestamp: number, winsBySymbol: Map<string, TilePosition[]>) => void
  clear: () => void
  startSequentialAnimation: (winsBySymbol: Map<string, TilePosition[]>) => Promise<void>
  isAnimating: () => boolean
  getCurrentAnimatingSymbol: () => string | null
  getStoredPositionsForSymbol: (symbol: string) => TilePosition[] | null
}

// Animation constants
const TOTAL_PATH_ANIMATION_DURATION = 400 // ms for entire path to animate
const HOLD_TIME = 200 // ms to hold completed path before next symbol

// Timing values (removed console.log for performance)

// Visual constants - smooth glowing beam
const BEAM_CORE_WIDTH = 4
const BEAM_INNER_WIDTH = 12
const BEAM_OUTER_WIDTH = 24
const BEAM_GLOW_WIDTH = 40

// Symbol-specific color schemes matching tile designs
interface ColorScheme {
  core: number
  inner: number
  outer: number
  glow: number
}

const SYMBOL_COLORS: Record<string, ColorScheme> = {
  // FA (Phoenix) - Fiery orange/red energy
  fa: {
    core: 0xffffff,
    inner: 0xffaa44,
    outer: 0xff6622,
    glow: 0xcc3300
  },
  // ZHONG (Green Rune) - Bright green energy
  zhong: {
    core: 0xffffff,
    inner: 0x88ff88,
    outer: 0x44cc44,
    glow: 0x228822
  },
  // BAI (White/Silver) - White/cyan energy
  bai: {
    core: 0xffffff,
    inner: 0xeeffff,
    outer: 0xaaddee,
    glow: 0x66aacc
  },
  // BAWAN (Brown/Cream) - Golden brown energy
  bawan: {
    core: 0xffffff,
    inner: 0xeedd99,
    outer: 0xccaa66,
    glow: 0x997744
  },
  // Default - Golden energy (for other symbols)
  default: {
    core: 0xffffff,
    inner: 0xffee88,
    outer: 0xffcc44,
    glow: 0xff9900
  }
}

/**
 * Get color scheme for a symbol
 */
function getColorsForSymbol(symbol: string): ColorScheme {
  // Check for exact match first
  if (SYMBOL_COLORS[symbol]) {
    return SYMBOL_COLORS[symbol]
  }
  // Check if symbol starts with known prefix (e.g., fa_gold -> fa)
  const baseSymbol = symbol.replace(/_gold$/, '')
  if (SYMBOL_COLORS[baseSymbol]) {
    return SYMBOL_COLORS[baseSymbol]
  }
  return SYMBOL_COLORS.default
}

// Line smoothness - more points = smoother animation
const POINTS_PER_SEGMENT = 30

interface PathData {
  tilePoints: { x: number; y: number }[]
  pathPoints: { x: number; y: number }[]
  totalLength: number
  cumulativeLengths: number[]
}

interface SymbolPathState {
  leftPath: PathData | null
  rightPath: PathData | null
  leftGraphics: Graphics
  rightGraphics: Graphics
  lastPositionsHash: string
  startTime: number
  progress: number
  symbol: string
}

/**
 * Ease-out cubic - starts fast, slows down at the end
 */
function easeOutCubic(t: number): number {
  return 1 - Math.pow(1 - t, 3)
}

/**
 * Calculate distance between two points
 */
function distance(x1: number, y1: number, x2: number, y2: number): number {
  const dx = x2 - x1
  const dy = y2 - y1
  return Math.sqrt(dx * dx + dy * dy)
}

/**
 * Calculate cumulative lengths along a path
 */
function calculateCumulativeLengths(points: { x: number; y: number }[]): { totalLength: number; cumulativeLengths: number[] } {
  const cumulativeLengths: number[] = [0]
  let totalLength = 0

  for (let i = 1; i < points.length; i++) {
    const segmentLength = distance(points[i - 1].x, points[i - 1].y, points[i].x, points[i].y)
    totalLength += segmentLength
    cumulativeLengths.push(totalLength)
  }

  return { totalLength, cumulativeLengths }
}

/**
 * Sort positions left-to-right by column, then top-to-bottom by row
 * All winning tiles are included to show the complete winning combination
 */
function sortPositions(positions: TilePosition[]): TilePosition[] {
  return [...positions].sort((a, b) => {
    if (a.col !== b.col) return a.col - b.col
    return a.row - b.row
  })
}

/**
 * Split positions into left and right groups that animate towards the middle
 * Both groups include ALL middle column tiles so no tiles are excluded
 * Left group: columns 0, 1, 2 → sorted left to right, middle column top to bottom
 * Right group: columns 4, 3, 2 → sorted right to left, middle column bottom to top
 * This creates a natural flow where both paths meet in the middle column
 */
function splitPositionsForMiddleConverge(positions: TilePosition[]): { left: TilePosition[], right: TilePosition[] } {
  // Find min and max columns
  const cols = positions.map(p => p.col)
  const minCol = Math.min(...cols)
  const maxCol = Math.max(...cols)

  // Calculate middle column (the meeting point)
  const midCol = Math.floor((minCol + maxCol) / 2)

  // Left group: all tiles with col <= midCol
  // Sort: left to right by column, middle column tiles go top to bottom
  const left = positions
    .filter(p => p.col <= midCol)
    .sort((a, b) => {
      if (a.col !== b.col) return a.col - b.col
      // For middle column, sort top to bottom (ascending row)
      return a.row - b.row
    })

  // Right group: all tiles with col >= midCol
  // Sort: right to left by column, middle column tiles go bottom to top
  const right = positions
    .filter(p => p.col >= midCol)
    .sort((a, b) => {
      if (a.col !== b.col) return b.col - a.col // Reverse column order
      // For middle column, sort bottom to top (descending row)
      return b.row - a.row
    })

  return { left, right }
}

/**
 * Create hash of positions for change detection
 */
function hashPositions(positions: TilePosition[]): string {
  return positions.map(p => `${p.col},${p.row}`).join('|')
}

/**
 * Create straight line path through tile centers
 */
function createStraightLinePath(tilePoints: { x: number; y: number }[]): { x: number; y: number }[] {
  if (tilePoints.length < 2) return tilePoints

  const pathPoints: { x: number; y: number }[] = []

  for (let i = 0; i < tilePoints.length - 1; i++) {
    const start = tilePoints[i]
    const end = tilePoints[i + 1]

    const isLastSegment = i === tilePoints.length - 2
    const pointCount = isLastSegment ? POINTS_PER_SEGMENT + 1 : POINTS_PER_SEGMENT

    for (let j = 0; j < pointCount; j++) {
      const t = j / POINTS_PER_SEGMENT
      pathPoints.push({
        x: start.x + (end.x - start.x) * t,
        y: start.y + (end.y - start.y) * t
      })
    }
  }

  return pathPoints
}

export function createImageLightningLinksManager(): ImageLightningLinksManager {
  const container = new Container()
  container.sortableChildren = true

  const symbolStates = new Map<string, SymbolPathState>()

  // Sequential animation state
  let sequentialMode = false
  let sequentialSymbols: string[] = []
  let currentSymbolIndex = 0
  let sequentialAnimationResolve: (() => void) | null = null

  // Store positions when they have real coordinates so we don't lose them when highlightWins clears
  const storedPositions = new Map<string, TilePosition[]>()

  /**
   * Draw a smooth glowing beam path with symbol-specific colors
   */
  function drawBeamPath(g: Graphics, points: { x: number; y: number }[], colors: ColorScheme): void {
    if (points.length < 2) return

    // Layer 1: Outer glow
    g.moveTo(points[0].x, points[0].y)
    for (let i = 1; i < points.length; i++) {
      g.lineTo(points[i].x, points[i].y)
    }
    g.stroke({ color: colors.glow, width: BEAM_GLOW_WIDTH, alpha: 0.15, cap: 'round', join: 'round' })

    // Layer 2: Mid glow
    g.moveTo(points[0].x, points[0].y)
    for (let i = 1; i < points.length; i++) {
      g.lineTo(points[i].x, points[i].y)
    }
    g.stroke({ color: colors.outer, width: BEAM_OUTER_WIDTH, alpha: 0.4, cap: 'round', join: 'round' })

    // Layer 3: Inner beam
    g.moveTo(points[0].x, points[0].y)
    for (let i = 1; i < points.length; i++) {
      g.lineTo(points[i].x, points[i].y)
    }
    g.stroke({ color: colors.inner, width: BEAM_INNER_WIDTH, alpha: 0.8, cap: 'round', join: 'round' })

    // Layer 4: Core
    g.moveTo(points[0].x, points[0].y)
    for (let i = 1; i < points.length; i++) {
      g.lineTo(points[i].x, points[i].y)
    }
    g.stroke({ color: colors.core, width: BEAM_CORE_WIDTH, alpha: 1, cap: 'round', join: 'round' })
  }

  /**
   * Draw a single path up to current progress
   */
  function drawSinglePath(g: Graphics, pathData: PathData, progress: number, symbol: string): void {
    g.clear()

    const { pathPoints, totalLength, cumulativeLengths } = pathData

    if (pathPoints.length < 2 || totalLength === 0) return

    const currentDistance = totalLength * progress

    const pointsToDraw: { x: number; y: number }[] = [pathPoints[0]]

    for (let i = 1; i < pathPoints.length; i++) {
      const segmentEndDistance = cumulativeLengths[i]

      if (currentDistance >= segmentEndDistance) {
        pointsToDraw.push(pathPoints[i])
      } else {
        const segmentStartDistance = cumulativeLengths[i - 1]
        const segmentLength = segmentEndDistance - segmentStartDistance
        const progressInSegment = (currentDistance - segmentStartDistance) / segmentLength

        const prevPoint = pathPoints[i - 1]
        const currPoint = pathPoints[i]
        const partialX = prevPoint.x + (currPoint.x - prevPoint.x) * progressInSegment
        const partialY = prevPoint.y + (currPoint.y - prevPoint.y) * progressInSegment

        pointsToDraw.push({ x: partialX, y: partialY })
        break
      }
    }

    if (pointsToDraw.length < 2) return

    const colors = getColorsForSymbol(symbol)
    drawBeamPath(g, pointsToDraw, colors)
  }

  /**
   * Draw both left and right paths up to current progress
   */
  function drawPath(state: SymbolPathState): void {
    if (state.leftPath) {
      drawSinglePath(state.leftGraphics, state.leftPath, state.progress, state.symbol)
    }
    if (state.rightPath) {
      drawSinglePath(state.rightGraphics, state.rightPath, state.progress, state.symbol)
    }
  }

  /**
   * Helper function to create PathData from positions
   */
  function createPathData(positions: TilePosition[]): PathData | null {
    if (positions.length < 2) return null
    const tilePoints = positions.map(p => ({ x: p.x, y: p.y }))
    const pathPoints = createStraightLinePath(tilePoints)
    const { totalLength, cumulativeLengths } = calculateCumulativeLengths(pathPoints)
    return { tilePoints, pathPoints, totalLength, cumulativeLengths }
  }

  /**
   * Helper function to cleanup state graphics
   */
  function cleanupState(state: SymbolPathState): void {
    container.removeChild(state.leftGraphics)
    container.removeChild(state.rightGraphics)
    state.leftGraphics.destroy()
    state.rightGraphics.destroy()
  }

  /**
   * Update lightning links - matching frontend2's approach
   */
  function update(timestamp: number, winsBySymbol: Map<string, TilePosition[]>): void {
    if (sequentialMode) {
      updateSequential(timestamp, winsBySymbol)
      return
    }

    // Normal mode: show all symbols at once
    const activeSymbols = new Set<string>()

    for (const [symbol, positions] of winsBySymbol.entries()) {
      if (positions.length < 2) continue

      activeSymbols.add(symbol)

      const sortedPositions = sortPositions(positions)
      const positionsHash = hashPositions(sortedPositions)

      let state = symbolStates.get(symbol)

      if (!state || state.lastPositionsHash !== positionsHash) {
        if (state) {
          cleanupState(state)
        }

        // Split positions into left and right groups
        const { left, right } = splitPositionsForMiddleConverge(sortedPositions)

        const leftGraphics = new Graphics()
        const rightGraphics = new Graphics()
        container.addChild(leftGraphics)
        container.addChild(rightGraphics)

        state = {
          leftPath: createPathData(left),
          rightPath: createPathData(right),
          leftGraphics,
          rightGraphics,
          lastPositionsHash: positionsHash,
          startTime: timestamp,
          progress: 0,
          symbol
        }
        symbolStates.set(symbol, state)
      }

      const elapsed = timestamp - state.startTime
      const linearProgress = Math.min(elapsed / TOTAL_PATH_ANIMATION_DURATION, 1)
      state.progress = easeOutCubic(linearProgress)
      drawPath(state)
    }

    for (const [symbol, state] of symbolStates.entries()) {
      if (!activeSymbols.has(symbol)) {
        cleanupState(state)
        symbolStates.delete(symbol)
      }
    }
  }

  /**
   * Update in sequential mode - store positions when they have real coordinates
   */
  function updateSequential(timestamp: number, winsBySymbol: Map<string, TilePosition[]>): void {
    // Store positions from winsBySymbol when they have real coordinates
    // This ensures we don't lose them when highlightWins clears
    for (const [symbol, positions] of winsBySymbol.entries()) {
      if (positions.length >= 2) {
        // Check if positions have real coordinates (not placeholders)
        const hasRealCoords = positions.some(p => p.x > 0 && p.y > 0)
        if (hasRealCoords && !storedPositions.has(symbol)) {
          storedPositions.set(symbol, [...positions])
        }
      }
    }

    // Clear other symbols' graphics
    for (const [sym, st] of symbolStates.entries()) {
      if (sym !== sequentialSymbols[currentSymbolIndex]) {
        cleanupState(st)
        symbolStates.delete(sym)
      }
    }

    if (currentSymbolIndex >= sequentialSymbols.length) {
      finishSequentialAnimation()
      return
    }

    const currentSymbol = sequentialSymbols[currentSymbolIndex]

    // Use stored positions first, fall back to winsBySymbol
    let positions = storedPositions.get(currentSymbol) || winsBySymbol.get(currentSymbol)

    if (!positions || positions.length < 2) {
      // Skip symbol with no positions
      currentSymbolIndex++
      if (currentSymbolIndex >= sequentialSymbols.length) {
        finishSequentialAnimation()
      }
      return
    }

    const sortedPositions = sortPositions(positions)
    const positionsHash = hashPositions(sortedPositions)

    let state = symbolStates.get(currentSymbol)

    if (!state) {
      // Split positions into left and right groups
      const { left, right } = splitPositionsForMiddleConverge(sortedPositions)

      const leftGraphics = new Graphics()
      const rightGraphics = new Graphics()
      container.addChild(leftGraphics)
      container.addChild(rightGraphics)

      state = {
        leftPath: createPathData(left),
        rightPath: createPathData(right),
        leftGraphics,
        rightGraphics,
        lastPositionsHash: positionsHash,
        startTime: timestamp,
        progress: 0,
        symbol: currentSymbol
      }
      symbolStates.set(currentSymbol, state)
    } else if (state.lastPositionsHash !== positionsHash) {
      // Rebuild paths with new positions
      const { left, right } = splitPositionsForMiddleConverge(sortedPositions)
      state.leftPath = createPathData(left)
      state.rightPath = createPathData(right)
      state.lastPositionsHash = positionsHash
    }

    const elapsed = timestamp - state.startTime
    const linearProgress = Math.min(elapsed / TOTAL_PATH_ANIMATION_DURATION, 1)
    state.progress = easeOutCubic(linearProgress)

    drawPath(state)

    if (linearProgress >= 1) {
      const totalElapsed = timestamp - state.startTime
      if (totalElapsed >= TOTAL_PATH_ANIMATION_DURATION + HOLD_TIME) {
        cleanupState(state)
        symbolStates.delete(currentSymbol)

        currentSymbolIndex++
        if (currentSymbolIndex >= sequentialSymbols.length) {
          finishSequentialAnimation()
        }
      }
    }
  }

  /**
   * Finish sequential animation
   */
  function finishSequentialAnimation(): void {
    // Sequential animation complete
    sequentialMode = false
    sequentialSymbols = []
    currentSymbolIndex = 0
    storedPositions.clear()

    for (const state of symbolStates.values()) {
      cleanupState(state)
    }
    symbolStates.clear()

    if (sequentialAnimationResolve) {
      // Resolve animation promise
      sequentialAnimationResolve()
      sequentialAnimationResolve = null
    }
  }

  /**
   * Clear all lightning links
   */
  function clear(): void {
    for (const state of symbolStates.values()) {
      cleanupState(state)
    }
    symbolStates.clear()
    storedPositions.clear()

    sequentialMode = false
    sequentialSymbols = []
    currentSymbolIndex = 0

    if (sequentialAnimationResolve) {
      sequentialAnimationResolve()
      sequentialAnimationResolve = null
    }
  }

  /**
   * Start sequential animation mode - matching frontend2's logic
   */
  function startSequentialAnimation(winsBySymbol: Map<string, TilePosition[]>): Promise<void> {
    // Start sequential animation
    clear()

    sequentialSymbols = []
    for (const [symbol, positions] of winsBySymbol.entries()) {
      // Symbol has positions
      if (positions.length >= 2) {
        sequentialSymbols.push(symbol)
      }
    }

    // Sequential symbols determined

    if (sequentialSymbols.length === 0) {
      // No symbols to animate
      return Promise.resolve()
    }

    sequentialMode = true
    currentSymbolIndex = 0

    // Starting sequential animation
    return new Promise((resolve) => {
      sequentialAnimationResolve = resolve
    })
  }

  /**
   * Check if sequential animation is currently active
   */
  function isAnimating(): boolean {
    return sequentialMode
  }

  /**
   * Get the symbol currently being animated
   */
  function getCurrentAnimatingSymbol(): string | null {
    if (!sequentialMode || currentSymbolIndex >= sequentialSymbols.length) {
      return null
    }
    return sequentialSymbols[currentSymbolIndex]
  }

  /**
   * Get stored positions for a symbol (used for highlighting during animation)
   */
  function getStoredPositionsForSymbol(symbol: string): TilePosition[] | null {
    return storedPositions.get(symbol) || null
  }

  return {
    container,
    update,
    clear,
    startSequentialAnimation,
    isAnimating,
    getCurrentAnimatingSymbol,
    getStoredPositionsForSymbol
  }
}
