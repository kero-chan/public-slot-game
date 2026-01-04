/**
 * Tile object structure and helper functions
 * Utilities for working with tile objects and symbol strings
 */

/**
 * Tile object interface
 */
export interface Tile {
  symbol: string
  isGolden: boolean
  isWildcard: boolean
}

/**
 * Type that can be either a Tile object or a string symbol
 */
export type TileOrString = Tile | string | null

/**
 * Create a tile object
 *
 * @param symbol - The base symbol (e.g., 'fa', 'chu', 'wild')
 * @param isGolden - Whether the tile is golden
 * @returns Tile object
 */
export function createTile(symbol: string, isGolden: boolean = false): Tile {
  return {
    symbol,
    isGolden,
    isWildcard: symbol === 'wild'
  }
}

/**
 * Get the symbol from a tile (handles both object and string format)
 *
 * @param tile - Tile object or string
 * @returns Symbol string or null
 */
export function getTileSymbol(tile: TileOrString): string | null {
  if (!tile) return null
  return typeof tile === 'string' ? tile : tile.symbol
}

/**
 * Check if a tile is golden
 *
 * @param tile - Tile object or string
 * @returns True if golden
 */
export function istileWilden(tile: TileOrString): boolean {
  if (!tile) return false
  if (typeof tile === 'string') return tile.endsWith('_gold')
  return tile.isGolden
}

/**
 * Check if a tile is a wildcard
 *
 * @param tile - Tile object or string
 * @returns True if wildcard
 */
export function isTileWildcard(tile: TileOrString): boolean {
  if (!tile) return false
  const symbol = getTileBaseSymbol(tile)
  return symbol === 'wild'
}

/**
 * Check if a tile is a bonus tile
 *
 * @param tile - Tile object or string
 * @returns True if bonus
 */
export function isBonusTile(tile: TileOrString): boolean {
  if (!tile) return false
  const symbol = getTileBaseSymbol(tile)
  return symbol === 'bonus'
}

/**
 * Get the base symbol (without _gold suffix)
 *
 * @param tile - Tile object or string
 * @returns Base symbol or null
 */
export function getTileBaseSymbol(tile: TileOrString): string | null {
  if (!tile) return null

  if (typeof tile === 'string') {
    return tile.endsWith('_gold') ? tile.slice(0, -5) : tile
  }

  return tile.symbol
}

/**
 * Check if two tiles match (considering wildcards)
 *
 * @param tile1 - First tile
 * @param tile2 - Second tile
 * @param allowWildcard - Whether to allow wildcard matching
 * @returns True if tiles match
 */
export function tilesMatch(
  tile1: TileOrString,
  tile2: TileOrString,
  allowWildcard: boolean = true
): boolean {
  const symbol1 = getTileBaseSymbol(tile1)
  const symbol2 = getTileBaseSymbol(tile2)

  if (!symbol1 || !symbol2) return false
  if (symbol1 === symbol2) return true

  if (allowWildcard) {
    const isWild1 = isTileWildcard(tile1)
    const isWild2 = isTileWildcard(tile2)
    return isWild1 || isWild2
  }

  return false
}

/**
 * Convert tile object to string representation
 *
 * @param tile - Tile object or string
 * @returns String representation or null
 */
export function tileToString(tile: TileOrString): string | null {
  if (!tile) return null
  if (typeof tile === 'string') return tile

  const { symbol, isGolden } = tile
  return isGolden && symbol !== 'wild' ? `${symbol}_gold` : symbol
}

/**
 * Convert string to tile object
 *
 * @param str - String representation
 * @returns Tile object or null
 */
export function stringToTile(str: string | null): Tile | null {
  if (!str) return null
  if (typeof str !== 'string') return str as any

  const isGolden = str.endsWith('_gold')
  const symbol = isGolden ? str.slice(0, -5) : str

  return createTile(symbol, isGolden)
}

/**
 * Convert grid of strings to grid of tile objects
 *
 * @param stringGrid - 2D array of symbol strings
 * @returns 2D array of Tile objects
 */
export function convertGridToTiles(stringGrid: string[][]): Tile[][] {
  return stringGrid.map(column =>
    column.map(cell => stringToTile(cell)!)
  )
}

/**
 * Convert grid of tile objects to grid of strings
 *
 * @param tileGrid - 2D array of Tile objects
 * @returns 2D array of symbol strings
 */
export function convertGridToStrings(tileGrid: Tile[][]): string[][] {
  return tileGrid.map(column =>
    column.map(tile => tileToString(tile)!)
  )
}
