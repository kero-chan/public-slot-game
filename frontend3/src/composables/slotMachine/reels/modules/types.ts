import type { Sprite, Container } from 'pixi.js'
import type { gsap } from 'gsap'

export interface ReelsRect {
  x?: number
  y: number
  w?: number
  h: number
}

export interface TileSize {
  w: number
  h: number
}

export interface SymbolSprite extends Sprite {
  _symbolData?: string
}

export interface ReelSpinData {
  gsapTimeline: gsap.core.Timeline
  isSpinning: boolean
  stripSprites: Sprite[]
}

export interface ScrollData {
  position: number
  velocity: number
}

export interface ReelBlur {
  strength: number
}

export const COLS = 5
export const ROWS_FULL = 6
export const VISIBLE_ROWS = 4
export const TOP_PARTIAL = 0
export const BLEED = 2
export const TILE_SPACING = 5
export const BUFFER_OFFSET = 4
