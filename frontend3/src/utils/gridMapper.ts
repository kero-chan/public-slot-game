/**
 * Grid Format Conversion Utilities
 * Converts between backend grid format and frontend tile format
 * Per spec: specs/07-frontend-integration.md
 */

import type { Grid } from '@/features/spin/types'

/**
 * Tile object with symbol and golden flag
 */
export interface TileObject {
  symbol: string
  isGolden: boolean
}

/**
 * Convert backend grid to frontend tile format
 * Backend: ["fa", "zhong_gold", "wild"]
 * Frontend: Keep as strings (NO LONGER converted to objects)
 *
 * @param backendGrid - 5 columns, 6 rows each
 * @returns Frontend grid with symbol strings
 */
export function mapBackendGridToFrontend(backendGrid: Grid): Grid {
  if (!backendGrid || !Array.isArray(backendGrid)) {
    console.error('Invalid backend grid:', backendGrid)
    return []
  }

  // FIXED: Return strings as-is, don't convert to objects
  // The rest of the codebase expects strings, not objects
  return backendGrid.map(column =>
    column.map(symbol => {
      if (typeof symbol === 'string') {
        return symbol
      }

      // If somehow we received an object, convert it back to string
      if (typeof symbol === 'object' && (symbol as any).symbol) {
        const obj = symbol as TileObject
        return obj.isGolden ? `${obj.symbol}_gold` : obj.symbol
      }

      console.error('Invalid symbol type:', symbol)
      return 'fa' // Fallback
    })
  )
}

/**
 * Convert frontend grid to backend string format
 * Frontend: [{ symbol: "fa", isGolden: false }, { symbol: "zhong", isGolden: true }, ...]
 * Backend: ["fa", "zhong_gold", "wild"]
 *
 * @param frontendGrid - Frontend tile objects or strings
 * @returns Backend grid format
 */
export function mapFrontendGridToBackend(frontendGrid: (string | TileObject)[][]): Grid {
  if (!frontendGrid || !Array.isArray(frontendGrid)) {
    console.error('Invalid frontend grid:', frontendGrid)
    return []
  }

  return frontendGrid.map(column =>
    column.map(tile => {
      if (typeof tile === 'string') {
        // Already in backend format
        return tile
      }

      // Convert tile object to string
      return tile.isGolden ? `${tile.symbol}_gold` : tile.symbol
    })
  )
}

/**
 * Parse a single symbol string from backend
 *
 * @param symbolString - e.g., "fa", "zhong_gold"
 * @returns Parsed symbol object
 */
export function parseSymbol(symbolString: string): TileObject {
  const isGolden = symbolString.endsWith('_gold')
  const symbol = symbolString.replace('_gold', '')

  return { symbol, isGolden }
}

/**
 * Format a tile object to backend string
 *
 * @param tile - Tile object or string
 * @returns Backend string format
 */
export function formatSymbol(tile: string | TileObject): string {
  if (typeof tile === 'string') return tile
  return tile.isGolden ? `${tile.symbol}_gold` : tile.symbol
}
