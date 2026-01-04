import type { Graphics, Sprite } from 'pixi.js'

export interface FooterRect {
  x: number
  y: number
  w: number
  h: number
}

export interface FooterHandlers {
  spin?: () => void
  increaseBet?: () => void
  decreaseBet?: () => void
  onGameModeClick?: () => void
}

export interface Particle extends Graphics {
  vx: number
  vy: number
  life: number
  maxLife: number
}

export interface LightningRays extends Graphics {
  rotation: number
}

export interface ScalableSprite extends Sprite {
  originalScale?: { x: number; y: number }
}

// Configurable constants
export const SPIN_BUTTON_CIRCLE_RADIUS_PER_MENU_HEIGHT = 0.45
export const SPIN_BUTTON_EFFECT_DURATION = 1
export const SPIN_BUTTON_EFFECT_FADEOUT = 0.3
export const NOTIFICATION_TEXT_SCALE = 0.7
export const WIN_TEXT_SCALE = 0.6
export const BIG_WIN_TEXT_SCALE = 0.42
