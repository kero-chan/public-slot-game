/**
 * Rectangle configuration
 */
export interface Rectangle {
  x: number
  y: number
  w: number
  h: number
}

/**
 * Source sprite icon configuration
 */
export interface SourceSpriteIcon {
  sourceSprite: string
  icon: Rectangle
  scale: number
}

/**
 * Tile layer configuration
 */
export interface TileLayer {
  sourceSprite: string
  icon: Rectangle
  scale: number
  offsetX?: number
  offsetY?: number
}

/**
 * Tile configuration
 */
export interface TileConfig {
  base?: Rectangle | SourceSpriteIcon
  outSize: number
  layers: TileLayer[]
}

/**
 * Tile slices configuration for all symbols
 */
export interface TileSlices {
  fa: TileConfig
  fa_gold: TileConfig
  wutong: TileConfig
  wusuo_gold: TileConfig
  wutong_gold: TileConfig
  wusuo: TileConfig
  bawan: TileConfig
  bawan_gold: TileConfig
  liangtong: TileConfig
  liangtong_gold: TileConfig
  liangsuo: TileConfig
  liangsuo_gold: TileConfig
  bai: TileConfig
  bai_gold: TileConfig
  zhong: TileConfig
  zhong_gold: TileConfig
  gold: TileConfig
  bonus: TileConfig
}

export const BASE_RECT: Rectangle = { x: 819, y: 0, w: 162, h: 188 }
export const BASE_GOLD_RECT: Rectangle = { x: 661, y: 0, w: 155, h: 188 }

export const TILE_SLICES: TileSlices = {
  fa: {
    base: BASE_RECT,
    outSize: 512,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 25, y: 211, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  fa_gold: {
    base: BASE_GOLD_RECT,
    outSize: 256,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 25, y: 211, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  wutong: {
    base: BASE_RECT,
    outSize: 512,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 353, y: 211, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  wusuo_gold: {
    base: BASE_GOLD_RECT,
    outSize: 256,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 514, y: 213, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  wutong_gold: {
    base: BASE_GOLD_RECT,
    outSize: 256,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 353, y: 211, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  wusuo: {
    base: BASE_RECT,
    outSize: 512,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 514, y: 213, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  bawan: {
    base: BASE_RECT,
    outSize: 512,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 25, y: 19, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  bawan_gold: {
    base: BASE_GOLD_RECT,
    outSize: 256,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 25, y: 19, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  liangtong: {
    base: BASE_RECT,
    outSize: 512,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 353, y: 24, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  liangtong_gold: {
    base: BASE_GOLD_RECT,
    outSize: 256,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 353, y: 24, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  liangsuo: {
    base: BASE_RECT,
    outSize: 512,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 516, y: 23, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  liangsuo_gold: {
    base: BASE_GOLD_RECT,
    outSize: 256,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 516, y: 23, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  bai: {
    base: BASE_RECT,
    outSize: 512,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 191, y: 210, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  bai_gold: {
    base: BASE_GOLD_RECT,
    outSize: 256,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 191, y: 210, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  zhong: {
    base: BASE_RECT,
    outSize: 512,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 189, y: 22, w: 120, h: 134 }, scale: 0.78 }
    ]
  },
  zhong_gold: {
    base: BASE_GOLD_RECT,
    outSize: 256,
    layers: [
      { sourceSprite: 'tiles_50', icon: { x: 189, y: 22, w: 120, h: 134 }, scale: 0.78 }
    ]
  },

  gold: {
    outSize: 510,
    layers: [
      { sourceSprite: 'tiles_29', icon: { x: 170, y: 6, w: 152, h: 85 }, scale: 0.4, offsetY: -0.27 },
      { sourceSprite: 'tiles_50', icon: { x: 660, y: 270, w: 152, h: 97 }, scale: 0.4, offsetY: 0.08 },
    ]
  },

  bonus: {
    base: { sourceSprite: 'tiles_34', icon: { x: 416, y: 16, w: 45, h: 53 }, scale: 1.8 },
    outSize: 512,
    layers: [
      { sourceSprite: 'tiles_29', icon: { x: 0, y: 1, w: 165, h: 180 }, scale: 2 },
    ]
  }
}

const DEFAULT_LAYER_OFFSET_X = 0.02
const DEFAULT_LAYER_OFFSET_Y = -0.05
const EXEMPT_SYMBOLS = new Set(['gold', 'bonus'])

for (const [symbol, cfg] of Object.entries(TILE_SLICES)) {
  if (EXEMPT_SYMBOLS.has(symbol)) continue
  if (Array.isArray(cfg.layers)) {
    for (const layer of cfg.layers) {
      if (!layer) continue
      if (layer.offsetX === undefined) layer.offsetX = DEFAULT_LAYER_OFFSET_X
      if (layer.offsetY === undefined) layer.offsetY = DEFAULT_LAYER_OFFSET_Y
    }
  }
}
