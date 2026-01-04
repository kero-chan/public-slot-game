import { describe, it, expect } from 'vitest'
import {
  createTile,
  getTileSymbol,
  istileWilden,
  isTileWildcard,
  isBonusTile,
  getTileBaseSymbol,
  tilesMatch,
  tileToString,
  stringToTile,
  convertGridToTiles,
  convertGridToStrings,
  type Tile
} from './tileHelpers'

describe('tileHelpers', () => {
  describe('createTile', () => {
    it('should create a basic tile', () => {
      const tile = createTile('fa')

      expect(tile.symbol).toBe('fa')
      expect(tile.isGolden).toBe(false)
      expect(tile.isWildcard).toBe(false)
    })

    it('should create a golden tile', () => {
      const tile = createTile('chu', true)

      expect(tile.symbol).toBe('chu')
      expect(tile.isGolden).toBe(true)
      expect(tile.isWildcard).toBe(false)
    })

    it('should create a wildcard tile', () => {
      const tile = createTile('wild')

      expect(tile.symbol).toBe('wild')
      expect(tile.isWildcard).toBe(true)
    })

    it('should create a golden wildcard tile', () => {
      const tile = createTile('wild', true)

      expect(tile.symbol).toBe('wild')
      expect(tile.isGolden).toBe(true)
      expect(tile.isWildcard).toBe(true)
    })
  })

  describe('getTileSymbol', () => {
    it('should get symbol from string', () => {
      expect(getTileSymbol('fa')).toBe('fa')
    })

    it('should get symbol from tile object', () => {
      const tile = createTile('chu')
      expect(getTileSymbol(tile)).toBe('chu')
    })

    it('should return null for null input', () => {
      expect(getTileSymbol(null)).toBeNull()
    })
  })

  describe('istileWilden', () => {
    it('should return false for regular string', () => {
      expect(istileWilden('fa')).toBe(false)
    })

    it('should return true for golden string', () => {
      expect(istileWilden('fa_gold')).toBe(true)
    })

    it('should return false for regular tile object', () => {
      const tile = createTile('fa', false)
      expect(istileWilden(tile)).toBe(false)
    })

    it('should return true for golden tile object', () => {
      const tile = createTile('fa', true)
      expect(istileWilden(tile)).toBe(true)
    })

    it('should return false for null', () => {
      expect(istileWilden(null)).toBe(false)
    })
  })

  describe('isTileWildcard', () => {
    it('should return true for wild string', () => {
      expect(isTileWildcard('wild')).toBe(true)
    })

    it('should return false for non-wild string', () => {
      expect(isTileWildcard('fa')).toBe(false)
    })

    it('should return true for wild tile object', () => {
      const tile = createTile('wild')
      expect(isTileWildcard(tile)).toBe(true)
    })

    it('should return false for non-wild tile object', () => {
      const tile = createTile('fa')
      expect(isTileWildcard(tile)).toBe(false)
    })

    it('should return false for null', () => {
      expect(isTileWildcard(null)).toBe(false)
    })
  })

  describe('isBonusTile', () => {
    it('should return true for bonus string', () => {
      expect(isBonusTile('bonus')).toBe(true)
    })

    it('should return false for non-bonus string', () => {
      expect(isBonusTile('fa')).toBe(false)
    })

    it('should return true for bonus tile object', () => {
      const tile = createTile('bonus')
      expect(isBonusTile(tile)).toBe(true)
    })

    it('should return false for non-bonus tile object', () => {
      const tile = createTile('fa')
      expect(isBonusTile(tile)).toBe(false)
    })

    it('should return false for null', () => {
      expect(isBonusTile(null)).toBe(false)
    })
  })

  describe('getTileBaseSymbol', () => {
    it('should return base symbol from regular string', () => {
      expect(getTileBaseSymbol('fa')).toBe('fa')
    })

    it('should return base symbol from golden string', () => {
      expect(getTileBaseSymbol('fa_gold')).toBe('fa')
    })

    it('should return base symbol from tile object', () => {
      const tile = createTile('chu', true)
      expect(getTileBaseSymbol(tile)).toBe('chu')
    })

    it('should return null for null input', () => {
      expect(getTileBaseSymbol(null)).toBeNull()
    })
  })

  describe('tilesMatch', () => {
    it('should match identical strings', () => {
      expect(tilesMatch('fa', 'fa')).toBe(true)
    })

    it('should not match different strings', () => {
      expect(tilesMatch('fa', 'chu')).toBe(false)
    })

    it('should match golden with non-golden of same symbol', () => {
      expect(tilesMatch('fa_gold', 'fa')).toBe(true)
    })

    it('should match wildcard with any symbol', () => {
      expect(tilesMatch('wild', 'fa')).toBe(true)
      expect(tilesMatch('fa', 'wild')).toBe(true)
    })

    it('should not match when allowWildcard is false', () => {
      expect(tilesMatch('wild', 'fa', false)).toBe(false)
    })

    it('should match identical symbols when allowWildcard is false', () => {
      expect(tilesMatch('fa', 'fa', false)).toBe(true)
    })

    it('should return false for null tiles', () => {
      expect(tilesMatch(null, 'fa')).toBe(false)
      expect(tilesMatch('fa', null)).toBe(false)
    })

    it('should match tile objects', () => {
      const tile1 = createTile('fa')
      const tile2 = createTile('fa')
      expect(tilesMatch(tile1, tile2)).toBe(true)
    })

    it('should match mixed types', () => {
      const tile = createTile('fa')
      expect(tilesMatch(tile, 'fa')).toBe(true)
      expect(tilesMatch('fa', tile)).toBe(true)
    })
  })

  describe('tileToString', () => {
    it('should return string as-is', () => {
      expect(tileToString('fa')).toBe('fa')
    })

    it('should convert regular tile to string', () => {
      const tile = createTile('fa')
      expect(tileToString(tile)).toBe('fa')
    })

    it('should convert golden tile to string with _gold suffix', () => {
      const tile = createTile('chu', true)
      expect(tileToString(tile)).toBe('chu_gold')
    })

    it('should not add _gold to golden wild', () => {
      const tile = createTile('wild', true)
      expect(tileToString(tile)).toBe('wild')
    })

    it('should return null for null input', () => {
      expect(tileToString(null)).toBeNull()
    })
  })

  describe('stringToTile', () => {
    it('should convert regular string to tile', () => {
      const tile = stringToTile('fa')

      expect(tile).not.toBeNull()
      expect(tile!.symbol).toBe('fa')
      expect(tile!.isGolden).toBe(false)
    })

    it('should convert golden string to tile', () => {
      const tile = stringToTile('chu_gold')

      expect(tile).not.toBeNull()
      expect(tile!.symbol).toBe('chu')
      expect(tile!.isGolden).toBe(true)
    })

    it('should convert wild string to tile', () => {
      const tile = stringToTile('wild')

      expect(tile).not.toBeNull()
      expect(tile!.symbol).toBe('wild')
      expect(tile!.isWildcard).toBe(true)
    })

    it('should return null for null input', () => {
      expect(stringToTile(null)).toBeNull()
    })
  })

  describe('convertGridToTiles', () => {
    it('should convert string grid to tile grid', () => {
      const stringGrid = [
        ['fa', 'chu'],
        ['fa', 'chu']
      ]

      const tileGrid = convertGridToTiles(stringGrid)

      expect(tileGrid).toHaveLength(2)
      expect(tileGrid[0]).toHaveLength(2)
      expect(tileGrid[0][0].symbol).toBe('fa')
      expect(tileGrid[0][1].symbol).toBe('chu')
    })

    it('should handle golden tiles', () => {
      const stringGrid = [
        ['fa_gold', 'chu_gold']
      ]

      const tileGrid = convertGridToTiles(stringGrid)

      expect(tileGrid[0][0].isGolden).toBe(true)
      expect(tileGrid[0][1].isGolden).toBe(true)
    })

    it('should handle empty grid', () => {
      const tileGrid = convertGridToTiles([])
      expect(tileGrid).toHaveLength(0)
    })
  })

  describe('convertGridToStrings', () => {
    it('should convert tile grid to string grid', () => {
      const tileGrid: Tile[][] = [
        [createTile('fa'), createTile('chu')],
        [createTile('fa'), createTile('chu')]
      ]

      const stringGrid = convertGridToStrings(tileGrid)

      expect(stringGrid).toHaveLength(2)
      expect(stringGrid[0]).toHaveLength(2)
      expect(stringGrid[0][0]).toBe('fa')
      expect(stringGrid[0][1]).toBe('chu')
    })

    it('should handle golden tiles', () => {
      const tileGrid: Tile[][] = [
        [createTile('fa', true), createTile('chu', true)]
      ]

      const stringGrid = convertGridToStrings(tileGrid)

      expect(stringGrid[0][0]).toBe('fa_gold')
      expect(stringGrid[0][1]).toBe('chu_gold')
    })

    it('should handle empty grid', () => {
      const stringGrid = convertGridToStrings([])
      expect(stringGrid).toHaveLength(0)
    })
  })

  describe('integration tests', () => {
    it('should round-trip string to tile and back', () => {
      const original = 'fa_gold'
      const tile = stringToTile(original)
      const result = tileToString(tile)

      expect(result).toBe(original)
    })

    it('should round-trip tile to string and back', () => {
      const original = createTile('chu', true)
      const str = tileToString(original)
      const result = stringToTile(str)

      expect(result).toEqual(original)
    })

    it('should round-trip grid conversions', () => {
      const originalStrings = [
        ['fa', 'chu_gold', 'wild'],
        ['fa_gold', 'chu', 'bonus']
      ]

      const tiles = convertGridToTiles(originalStrings)
      const resultStrings = convertGridToStrings(tiles)

      expect(resultStrings).toEqual(originalStrings)
    })
  })
})
