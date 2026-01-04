import type { WinCombination } from '@/types/global'
import type { GameLogicAPI, TilePosition } from './types'
import { numberToSymbol } from '@/utils/symbolConverter'
import { CONFIG } from '@/config/constants'

// ============================================
// 🧪 TEST MODE: Cycle through all win intensities during free spins
// Set TEST_WIN_ANIMATIONS=true in .env to enable
// ============================================
const testWinAnimationsValue = import.meta.env.VITE_TEST_WIN_ANIMATIONS
const TEST_WIN_ANIMATIONS = Boolean(testWinAnimationsValue && testWinAnimationsValue.toLowerCase() === 'true')
const TEST_INTENSITIES = ['small', 'medium', 'big', 'mega'] as const
let testIntensityIndex = 0

/**
 * Handle win checking
 * Also triggers symbol-specific animation for high-value wins (fa, zhong, bai, bawan)
 * Animation is only triggered on FIRST cascade (consecutiveWins === 0)
 */
export async function handleCheckWins(
  gameLogic: GameLogicAPI,
  gameStore: any
): Promise<void> {
  // Get wins from backend (backend provides all win data)
  let wins = gameLogic.getBackendWins()
  if (wins === null) {
    wins = gameLogic.findWinningCombinations()
  }

  if (wins.length > 0) {
    const totalWinAmount = wins.reduce((sum, win) => sum + (win.payout || 0), 0)

    if (gameStore.inFreeSpinMode) {
      gameLogic.playConsecutiveWinSound(gameStore.consecutiveWins, true)
      playWinSoundWithDelay(gameLogic, wins, gameStore.consecutiveWins)
    } else {
      gameLogic.playConsecutiveWinSound(gameStore.consecutiveWins, false)
      playWinSoundWithDelay(gameLogic, wins, gameStore.consecutiveWins)
    }

    // Track winning symbol for animation AFTER all cascades complete
    // Get winning_tile_kind from first cascade in spin response (server provides this)
    // This now tracks ALL winning symbols, not just high-value ones
    const isFirstWin = gameStore.consecutiveWins === 0
    if (isFirstWin) {
      const spinResponse = gameStore.spinResponse
      const firstCascade = spinResponse?.cascades?.[0]
      const winningTileKind = firstCascade?.winning_tile_kind
      if (winningTileKind) {
        // Store the winning symbol for later animation (after all cascades)
        // Background/frame change will only happen if the symbol has a background image
        gameStore.setHighValueWinSymbol(winningTileKind)
      }
    }
  }
  // Note: highValueWinSymbol is NOT cleared here when there are no wins.
  // It persists until shown in handleShowWinOverlay, because we may have
  // set it during an earlier cascade in the same spin.

  gameStore.setWinResults(wins)
}

/**
 * Play win sound with optional delay for consecutive wins
 */
function playWinSoundWithDelay(
  gameLogic: GameLogicAPI,
  wins: WinCombination[],
  consecutiveWins: number
): void {
  if (consecutiveWins > 0) {
    setTimeout(() => {
      gameLogic.playWinSound(wins)
    }, 500)
  } else {
    gameLogic.playWinSound(wins)
  }
}

/**
 * Build winsBySymbol map for lightning links animation
 * Groups winning positions by symbol name
 * Note: x,y coordinates are placeholders - actual positions come from render loop
 */
function buildWinsBySymbol(
  wins: WinCombination[],
  bufferOffset: number
): Map<string, TilePosition[]> {
  const winsBySymbol = new Map<string, TilePosition[]>()

  for (const win of wins) {
    const symbol = numberToSymbol(win.symbol)
    if (!winsBySymbol.has(symbol)) {
      winsBySymbol.set(symbol, [])
    }
    const positions = winsBySymbol.get(symbol)!

    for (const pos of win.positions) {
      const visualRow = pos.row - bufferOffset
      // Only include visible rows (1-4 for 4-row display, matching backend winCheckStartRow/EndRow)
      if (visualRow >= 1 && visualRow <= 4) {
        // Placeholder coordinates - actual positions calculated in render loop
        positions.push({
          x: 0,
          y: 0,
          col: pos.reel,
          row: visualRow
        })
      }
    }
  }

  return winsBySymbol
}

/**
 * Handle win highlighting animation
 */
export async function handleHighlightWins(
  gameLogic: GameLogicAPI,
  gameStore: any,
  winningStore: any,
  timingStore: any,
  bufferOffset: number
): Promise<void> {
  const wins = gameStore.currentWins

  if (!wins || wins.length === 0) {
    gameStore.completeHighlighting()
    return
  }

  // Clear completed shatter animations from previous win cycle
  // This allows the same cells to animate again in cascades
  if (gameLogic.clearCompletedShatterAnimations) {
    gameLogic.clearCompletedShatterAnimations()
  }

  // Convert wins to cell keys and set to HIGHLIGHTED state
  const cellKeys = winningStore.winsToCellKeys(wins, bufferOffset)
  winningStore.setHighlighted(cellKeys)

  // Start highlight animation (winning frames appear)
  gameLogic.highlightWinsAnimation(wins)

  // Play line win sound when winning line is shown (relatively quiet)
  gameLogic.playLineWinSound()

  // Wait a moment for winning frames to be visible
  await new Promise(resolve => setTimeout(resolve, timingStore.HIGHLIGHT_BEFORE_FLIP))

  // Run lightning links animation RIGHT AFTER highlight frames appear
  // This shows links sequentially for each symbol type, then waits for completion
  // The cascade will only start AFTER this animation finishes
  if (gameLogic.startLightningLinksAnimation) {
    const winsBySymbol = buildWinsBySymbol(wins, bufferOffset)
    console.log('🔗 winHandler: winsBySymbol size =', winsBySymbol.size)
    if (winsBySymbol.size > 0) {
      console.log('🔗 winHandler: BEFORE await startLightningLinksAnimation')
      await gameLogic.startLightningLinksAnimation(winsBySymbol)
      console.log('🔗 winHandler: AFTER await startLightningLinksAnimation - animation done!')
    }
  }

  // Transition to FLIPPING state (after lightning links done)
  winningStore.setFlipping()

  // Wait for flip to complete
  // Note: Gold transform happens in separate TRANSFORMING_GOLD state, not here
  await new Promise(resolve => setTimeout(resolve, timingStore.FLIP_DURATION))

  // Transition to FLIPPED state
  winningStore.setFlipped()

  // Stop highlight animation
  gameLogic.stopHighlightAnimation()

  gameStore.completeHighlighting()
}

/**
 * Handle showing win overlay
 * Uses backend's spin_total_win for per-spin win announcement
 * Flow: Win overlay → (after dismiss) → high-value card → background/frame change
 */
export async function handleShowWinOverlay(
  gameLogic: GameLogicAPI,
  gameStore: any
): Promise<void> {
  // Note: High-value win card animation is shown AFTER the win overlay dismisses
  // This is handled in useRenderer's showWinOverlay callback
  // The highValueWinSymbol is read from gameStore there

  gameStore.markAnimationComplete()
  gameStore.setShowAmountNotification(true)

  let intensity = gameLogic.getWinIntensity(gameStore.allWinsThisSpin)

  // 🧪 TEST MODE: Override intensity to cycle through all types during free spins
  if (TEST_WIN_ANIMATIONS && gameStore.inFreeSpinMode) {
    intensity = TEST_INTENSITIES[testIntensityIndex]
    testIntensityIndex = (testIntensityIndex + 1) % TEST_INTENSITIES.length
  }

  // Use backend's spin_total_win for per-spin win amount
  // Fall back to accumulatedWinAmount if spin_total_win is 0 or missing (trial mode edge case)
  const displayAmount = gameStore.spinResponse?.spin_total_win || gameStore.accumulatedWinAmount

  // Skip showing win overlay if there's no amount to display
  if (displayAmount <= 0) {
    return
  }

  if (gameLogic.showWinOverlay) {
    gameLogic.showWinOverlay(intensity, displayAmount)
  }
}

/**
 * Handle showing final jackpot result
 * Uses freeSpinSessionWinAmount for the full session win amount
 */
export function handleShowFinalJackpotResult(
  gameLogic: GameLogicAPI,
  gameStore: any
): void {
  gameStore.markAnimationComplete()
  gameStore.setShowAmountNotification(true)

  // Use freeSpinSessionWinAmount for the full session win amount
  const totalWin = gameStore.freeSpinSessionWinAmount

  if (gameLogic.showFinalJackpotResult) {
    gameLogic.showFinalJackpotResult(totalWin)
  }
}
