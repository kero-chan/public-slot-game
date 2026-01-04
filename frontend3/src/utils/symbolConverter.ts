/**
 * Symbol Converter Utilities
 * Converts between symbol numbers (from backend) and symbol strings (for frontend)
 *
 * IMPORTANT: This must match the backend mapping in:
 * backend/internal/game/symbols/symbols.go::SymbolNumber()
 */

/**
 * Symbol number to symbol string mapping
 * Must match backend exactly
 */
const SYMBOL_NUMBER_MAP: Record<number, string> = {
  0: 'wild',
  1: 'bonus',
  2: 'fa',
  3: 'zhong',
  4: 'bai',
  5: 'bawan',
  6: 'wusuo',
  7: 'wutong',
  8: 'liangsuo',
  9: 'liangtong',
  // Gold variants
  12: 'fa_gold',
  13: 'zhong_gold',
  14: 'bai_gold',
  15: 'bawan_gold',
  16: 'wusuo_gold',
  17: 'wutong_gold',
  18: 'liangsuo_gold',
  19: 'liangtong_gold',
}

/**
 * Symbol string to symbol number mapping (reverse of above)
 */
const SYMBOL_STRING_MAP: Record<string, number> = Object.fromEntries(
  Object.entries(SYMBOL_NUMBER_MAP).map(([num, symbol]) => [symbol, Number(num)])
)

/**
 * Convert a symbol number to symbol string
 *
 * @param num - Symbol number from backend (0-19)
 * @returns Symbol string (e.g., 'fa', 'wild', 'fa_gold')
 *
 * @example
 * numberToSymbol(2) // 'fa'
 * numberToSymbol(12) // 'fa_gold'
 * numberToSymbol(0) // 'wild'
 */
export function numberToSymbol(num: number): string {
  const symbol = SYMBOL_NUMBER_MAP[num]
  if (!symbol) {
    console.warn(`Unknown symbol number: ${num}, defaulting to 'liangtong'`)
    return 'liangtong'
  }
  return symbol
}

/**
 * Convert a symbol string to symbol number
 *
 * @param symbol - Symbol string (e.g., 'fa', 'wild')
 * @returns Symbol number (0-19)
 *
 * @example
 * symbolToNumber('fa') // 2
 * symbolToNumber('fa_gold') // 12
 * symbolToNumber('wild') // 0
 */
export function symbolToNumber(symbol: string): number {
  const num = SYMBOL_STRING_MAP[symbol]
  if (num === undefined) {
    console.warn(`Unknown symbol string: ${symbol}, defaulting to 9 (liangtong)`)
    return 9
  }
  return num
}

/**
 * Convert a grid of numbers to a grid of symbols
 *
 * @param numberGrid - 2D array of symbol numbers from backend
 * @returns 2D array of symbol strings for frontend
 *
 * @example
 * const backendGrid = [[2, 3], [0, 1]]
 * const frontendGrid = convertGridToSymbols(backendGrid)
 * // [['fa', 'zhong'], ['wild', 'bonus']]
 */
export function convertGridToSymbols(numberGrid: number[][]): string[][] {
  return numberGrid.map(row => row.map(numberToSymbol))
}

/**
 * Convert a grid of symbols to a grid of numbers
 *
 * @param symbolGrid - 2D array of symbol strings
 * @returns 2D array of symbol numbers
 *
 * @example
 * const frontendGrid = [['fa', 'zhong'], ['wild', 'bonus']]
 * const backendGrid = convertGridToNumbers(frontendGrid)
 * // [[2, 3], [0, 1]]
 */
export function convertGridToNumbers(symbolGrid: string[][]): number[][] {
  return symbolGrid.map(row => row.map(symbolToNumber))
}

/**
 * Validate that a symbol number is valid
 *
 * @param num - Symbol number to validate
 * @returns true if valid, false otherwise
 */
export function isValidSymbolNumber(num: number): boolean {
  return num in SYMBOL_NUMBER_MAP
}

/**
 * Validate that a symbol string is valid
 *
 * @param symbol - Symbol string to validate
 * @returns true if valid, false otherwise
 */
export function isValidSymbolString(symbol: string): boolean {
  return symbol in SYMBOL_STRING_MAP
}

/**
 * Get all valid symbol numbers
 *
 * @returns Array of valid symbol numbers
 */
export function getAllSymbolNumbers(): number[] {
  return Object.keys(SYMBOL_NUMBER_MAP).map(Number)
}

/**
 * Get all valid symbol strings
 *
 * @returns Array of valid symbol strings
 */
export function getAllSymbolStrings(): string[] {
  return Object.keys(SYMBOL_STRING_MAP)
}
