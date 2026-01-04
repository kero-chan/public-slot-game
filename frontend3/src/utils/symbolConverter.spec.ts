import { describe, it, expect } from 'vitest'
import {
  numberToSymbol,
  symbolToNumber,
  convertGridToSymbols,
  convertGridToNumbers,
  isValidSymbolNumber,
  isValidSymbolString,
  getAllSymbolNumbers,
  getAllSymbolStrings,
} from './symbolConverter'

describe('symbolConverter', () => {
  describe('numberToSymbol', () => {
    it('should convert basic symbols correctly', () => {
      expect(numberToSymbol(0)).toBe('wild')
      expect(numberToSymbol(1)).toBe('bonus')
      expect(numberToSymbol(2)).toBe('fa')
      expect(numberToSymbol(3)).toBe('zhong')
      expect(numberToSymbol(4)).toBe('bai')
      expect(numberToSymbol(5)).toBe('bawan')
      expect(numberToSymbol(6)).toBe('wusuo')
      expect(numberToSymbol(7)).toBe('wutong')
      expect(numberToSymbol(8)).toBe('liangsuo')
      expect(numberToSymbol(9)).toBe('liangtong')
    })

    it('should convert gold variant symbols correctly', () => {
      expect(numberToSymbol(12)).toBe('fa_gold')
      expect(numberToSymbol(13)).toBe('zhong_gold')
      expect(numberToSymbol(14)).toBe('bai_gold')
      expect(numberToSymbol(15)).toBe('bawan_gold')
      expect(numberToSymbol(16)).toBe('wusuo_gold')
      expect(numberToSymbol(17)).toBe('wutong_gold')
      expect(numberToSymbol(18)).toBe('liangsuo_gold')
      expect(numberToSymbol(19)).toBe('liangtong_gold')
    })

    it('should return default for unknown numbers', () => {
      expect(numberToSymbol(99)).toBe('liangtong')
      expect(numberToSymbol(-1)).toBe('liangtong')
    })
  })

  describe('symbolToNumber', () => {
    it('should convert basic symbols correctly', () => {
      expect(symbolToNumber('wild')).toBe(0)
      expect(symbolToNumber('bonus')).toBe(1)
      expect(symbolToNumber('fa')).toBe(2)
      expect(symbolToNumber('zhong')).toBe(3)
      expect(symbolToNumber('bai')).toBe(4)
      expect(symbolToNumber('bawan')).toBe(5)
      expect(symbolToNumber('wusuo')).toBe(6)
      expect(symbolToNumber('wutong')).toBe(7)
      expect(symbolToNumber('liangsuo')).toBe(8)
      expect(symbolToNumber('liangtong')).toBe(9)
    })

    it('should convert gold variant symbols correctly', () => {
      expect(symbolToNumber('fa_gold')).toBe(12)
      expect(symbolToNumber('zhong_gold')).toBe(13)
      expect(symbolToNumber('bai_gold')).toBe(14)
      expect(symbolToNumber('bawan_gold')).toBe(15)
      expect(symbolToNumber('wusuo_gold')).toBe(16)
      expect(symbolToNumber('wutong_gold')).toBe(17)
      expect(symbolToNumber('liangsuo_gold')).toBe(18)
      expect(symbolToNumber('liangtong_gold')).toBe(19)
    })

    it('should return default for unknown symbols', () => {
      expect(symbolToNumber('unknown')).toBe(9)
      expect(symbolToNumber('')).toBe(9)
    })
  })

  describe('convertGridToSymbols', () => {
    it('should convert a grid of numbers to symbols', () => {
      const numberGrid = [
        [2, 3, 4],
        [0, 1, 5],
      ]
      const symbolGrid = convertGridToSymbols(numberGrid)

      expect(symbolGrid).toEqual([
        ['fa', 'zhong', 'bai'],
        ['wild', 'bonus', 'bawan'],
      ])
    })

    it('should handle gold variants in grid', () => {
      const numberGrid = [
        [2, 12],
        [3, 13],
      ]
      const symbolGrid = convertGridToSymbols(numberGrid)

      expect(symbolGrid).toEqual([
        ['fa', 'fa_gold'],
        ['zhong', 'zhong_gold'],
      ])
    })

    it('should handle empty grid', () => {
      const numberGrid: number[][] = []
      const symbolGrid = convertGridToSymbols(numberGrid)

      expect(symbolGrid).toEqual([])
    })
  })

  describe('convertGridToNumbers', () => {
    it('should convert a grid of symbols to numbers', () => {
      const symbolGrid = [
        ['fa', 'zhong', 'bai'],
        ['wild', 'bonus', 'bawan'],
      ]
      const numberGrid = convertGridToNumbers(symbolGrid)

      expect(numberGrid).toEqual([
        [2, 3, 4],
        [0, 1, 5],
      ])
    })

    it('should handle gold variants in grid', () => {
      const symbolGrid = [
        ['fa', 'fa_gold'],
        ['zhong', 'zhong_gold'],
      ]
      const numberGrid = convertGridToNumbers(symbolGrid)

      expect(numberGrid).toEqual([
        [2, 12],
        [3, 13],
      ])
    })

    it('should handle empty grid', () => {
      const symbolGrid: string[][] = []
      const numberGrid = convertGridToNumbers(symbolGrid)

      expect(numberGrid).toEqual([])
    })
  })

  describe('isValidSymbolNumber', () => {
    it('should validate valid symbol numbers', () => {
      expect(isValidSymbolNumber(0)).toBe(true)
      expect(isValidSymbolNumber(1)).toBe(true)
      expect(isValidSymbolNumber(9)).toBe(true)
      expect(isValidSymbolNumber(12)).toBe(true)
      expect(isValidSymbolNumber(19)).toBe(true)
    })

    it('should reject invalid symbol numbers', () => {
      expect(isValidSymbolNumber(10)).toBe(false)
      expect(isValidSymbolNumber(11)).toBe(false)
      expect(isValidSymbolNumber(20)).toBe(false)
      expect(isValidSymbolNumber(-1)).toBe(false)
      expect(isValidSymbolNumber(999)).toBe(false)
    })
  })

  describe('isValidSymbolString', () => {
    it('should validate valid symbol strings', () => {
      expect(isValidSymbolString('wild')).toBe(true)
      expect(isValidSymbolString('bonus')).toBe(true)
      expect(isValidSymbolString('fa')).toBe(true)
      expect(isValidSymbolString('fa_gold')).toBe(true)
      expect(isValidSymbolString('liangtong')).toBe(true)
    })

    it('should reject invalid symbol strings', () => {
      expect(isValidSymbolString('unknown')).toBe(false)
      expect(isValidSymbolString('')).toBe(false)
      expect(isValidSymbolString('WILD')).toBe(false)
      expect(isValidSymbolString('fa gold')).toBe(false)
    })
  })

  describe('getAllSymbolNumbers', () => {
    it('should return all valid symbol numbers', () => {
      const numbers = getAllSymbolNumbers()

      expect(numbers).toContain(0) // wild
      expect(numbers).toContain(1) // bonus
      expect(numbers).toContain(9) // liangtong
      expect(numbers).toContain(12) // fa_gold
      expect(numbers).toContain(19) // liangtong_gold
      expect(numbers.length).toBeGreaterThan(10)
    })
  })

  describe('getAllSymbolStrings', () => {
    it('should return all valid symbol strings', () => {
      const symbols = getAllSymbolStrings()

      expect(symbols).toContain('wild')
      expect(symbols).toContain('bonus')
      expect(symbols).toContain('fa')
      expect(symbols).toContain('fa_gold')
      expect(symbols).toContain('liangtong')
      expect(symbols.length).toBeGreaterThan(10)
    })
  })

  describe('round-trip conversion', () => {
    it('should preserve data through number->symbol->number conversion', () => {
      const numbers = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 13, 14, 15, 16, 17, 18, 19]

      for (const num of numbers) {
        const symbol = numberToSymbol(num)
        const backToNumber = symbolToNumber(symbol)
        expect(backToNumber).toBe(num)
      }
    })

    it('should preserve data through symbol->number->symbol conversion', () => {
      const symbols = [
        'wild',
        'bonus',
        'fa',
        'fa_gold',
        'zhong',
        'zhong_gold',
        'bai',
        'bai_gold',
        'bawan',
        'bawan_gold',
        'wusuo',
        'wusuo_gold',
        'wutong',
        'wutong_gold',
        'liangsuo',
        'liangsuo_gold',
        'liangtong',
        'liangtong_gold',
      ]

      for (const symbol of symbols) {
        const num = symbolToNumber(symbol)
        const backToSymbol = numberToSymbol(num)
        expect(backToSymbol).toBe(symbol)
      }
    })
  })
})
