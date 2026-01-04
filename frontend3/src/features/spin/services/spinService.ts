/**
 * Spin Service - Pure business logic for spin operations
 * No Vue dependencies, fully testable
 */

import type { Grid, GridVerification, GridMismatch } from '../types'

/**
 * Validates grid integrity before spin
 * Ensures all cells have valid symbol data
 *
 * @param grid - 2D grid array (5 columns x N rows)
 * @param expectedRows - Expected number of rows per column
 * @param fallbackSymbol - Fallback symbol for invalid cells (default: 'fa')
 * @returns Validated grid with all invalid cells replaced
 *
 * @example
 * ```ts
 * const grid = [['fa', 'chu'], ['fa', ''], ...] // Invalid cell at [1][1]
 * const validated = validateGridIntegrity(grid, 10, 'fa')
 * // Returns grid with '' replaced by 'fa'
 * ```
 */
export function validateGridIntegrity(
  grid: Grid,
  expectedRows: number,
  fallbackSymbol: string = 'fa'
): Grid {
  const validatedGrid: Grid = []

  for (let col = 0; col < 5; col++) {
    if (!grid[col] || grid[col].length === 0) {
      console.warn(`⚠️ Column ${col} is empty, filling with safe fallback symbols`)
      validatedGrid[col] = Array(expectedRows).fill(fallbackSymbol)
    } else {
      validatedGrid[col] = [...grid[col]]

      // Check each cell in the column
      for (let row = 0; row < expectedRows; row++) {
        if (!validatedGrid[col][row] || typeof validatedGrid[col][row] !== 'string') {
          console.warn(
            `⚠️ Cell [${col},${row}] is invalid (${validatedGrid[col][row]}), filling with safe fallback`
          )
          validatedGrid[col][row] = fallbackSymbol
        }
      }
    }
  }

  return validatedGrid
}

/**
 * Verifies that displayed grid matches backend grid
 * Critical security check - ensures server authority
 *
 * @param backendGrid - Grid received from backend (authoritative source)
 * @param displayedGrid - Currently displayed grid in frontend
 * @returns Verification result with mismatch details
 *
 * @example
 * ```ts
 * const backend = [['fa', 'chu'], ['fa', 'fa'], ...]
 * const displayed = [['fa', 'chu'], ['fa', 'zhong'], ...] // Mismatch at [1][1]
 * const result = verifyGridMatch(backend, displayed)
 * // result.allMatch === false
 * // result.mismatches === [{ col: 1, row: 1, backend: 'fa', displayed: 'zhong' }]
 * ```
 */
export function verifyGridMatch(
  backendGrid: Grid,
  displayedGrid: Grid
): GridVerification {
  const mismatches: GridMismatch[] = []
  let allMatch = true

  const backendRowCount = backendGrid[0]?.length || 0

  for (let col = 0; col < 5; col++) {
    for (let row = 0; row < backendRowCount; row++) {
      const backendSymbol = backendGrid[col][row]
      const displayedSymbol = displayedGrid[col][row]

      if (backendSymbol !== displayedSymbol) {
        mismatches.push({
          col,
          row,
          backend: backendSymbol,
          displayed: displayedSymbol,
        })
        allMatch = false
      }
    }
  }

  return { allMatch, mismatches, totalCells: 5 * backendRowCount }
}

/**
 * Logs verification results to console
 * Provides detailed output for debugging grid mismatches
 *
 * @param verification - Result from verifyGridMatch
 * @param context - Context string for logging (e.g., "SPIN" or "CASCADE 1")
 *
 * @example
 * ```ts
 * const verification = verifyGridMatch(backend, displayed)
 * logVerification(verification, 'SPIN COMPLETE')
 * // Logs formatted verification results to console
 * ```
 */
export function logVerification(
  verification: GridVerification,
  context: string = 'VERIFICATION'
): void {
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━')
  console.log(`🔍 ${context}`)
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━')

  if (verification.allMatch) {
    console.log(
      `✅ ${context} PASSED: All ${verification.totalCells} tiles match backend data perfectly!`
    )
  } else {
    console.error(`❌ ${context} FAILED: Displayed grid does NOT match backend!`)
    verification.mismatches.forEach(m => {
      console.error(
        `  ❌ MISMATCH at [${m.col},${m.row}]: Backend="${m.backend}" Display="${m.displayed}"`
      )
    })
  }

  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━')
}
