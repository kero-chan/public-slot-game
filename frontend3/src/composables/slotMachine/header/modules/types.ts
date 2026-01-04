import type { Container, Sprite, Text, Texture } from 'pixi.js'

export interface HeaderRect {
  x: number
  y: number
  w: number
  h: number
}

export interface TextureConfig {
  texture: Texture | null
  scale: number
}

export interface MultiplierItem {
  sprite: Sprite | Text | Container
  multiplier: number
}
