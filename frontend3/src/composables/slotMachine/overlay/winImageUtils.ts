/**
 * Win Image Utilities
 * Shared utilities for rendering win announcement images
 * Ensures images display fully and nicely in viewport regardless of screen size
 */

import { Sprite, Texture } from 'pixi.js'
import { ASSETS } from '@/config/assets'
import { getWinAnnouncementSprite } from '@/config/spritesheet'

/**
 * Win image configuration
 */
export interface WinImageConfig {
  /** Asset key for the win image */
  imageKey: string
  /** Target height as percentage of canvas height (0-1), defaults to 0.25 */
  targetHeightPercent?: number
  /** Max width as percentage of canvas width (0-1), defaults to 0.85 */
  maxWidthPercent?: number
  /** Legacy: fixed target height (deprecated, use targetHeightPercent instead) */
  targetHeight?: number
}

/**
 * Result of creating a win image sprite
 */
export interface WinImageResult {
  sprite: Sprite
  scale: number
}

/**
 * Calculate the scale for a win image to fit within canvas bounds
 * Ensures image displays fully within viewport regardless of screen size
 * @param spriteWidth - Original sprite width
 * @param spriteHeight - Original sprite height
 * @param canvasWidth - Canvas width
 * @param canvasHeight - Canvas height
 * @param maxWidthPercent - Max width as percentage of canvas (0-1)
 * @param maxHeightPercent - Max height as percentage of canvas (0-1)
 * @returns The scale factor to apply
 */
export function calculateWinImageScale(
  spriteWidth: number,
  spriteHeight: number,
  canvasWidth: number,
  canvasHeight: number,
  maxWidthPercent: number = 0.85,
  maxHeightPercent: number = 0.25
): number {
  const maxWidth = canvasWidth * maxWidthPercent
  const maxHeight = canvasHeight * maxHeightPercent
  const scaleByWidth = maxWidth / spriteWidth
  const scaleByHeight = maxHeight / spriteHeight
  // Use the smaller scale to ensure image fits both dimensions
  return Math.min(scaleByWidth, scaleByHeight)
}

/**
 * Create a win image sprite with proper scaling for any screen size
 * @param config - Win image configuration
 * @param canvasWidth - Canvas width for scaling
 * @param canvasHeight - Canvas height for scaling (optional for backward compatibility)
 * @returns WinImageResult or null if image not found
 */
export function createWinImageSprite(
  config: WinImageConfig,
  canvasWidth: number,
  canvasHeight?: number
): WinImageResult | null {
  const sprite = getWinAnnouncementSprite(config.imageKey)
  if (!sprite) {
    return null
  }

  // Use canvas height if provided, otherwise estimate from width (16:9 ratio fallback)
  const effectiveCanvasHeight = canvasHeight || canvasWidth * (16 / 9)

  // Determine max height percent - use config value or calculate from legacy targetHeight
  let maxHeightPercent = config.targetHeightPercent ?? 0.25
  if (config.targetHeight && !config.targetHeightPercent) {
    // Legacy support: convert fixed height to percentage based on typical canvas height
    maxHeightPercent = Math.min(config.targetHeight / effectiveCanvasHeight, 0.35)
  }

  const scale = calculateWinImageScale(
    sprite.width,
    sprite.height,
    canvasWidth,
    effectiveCanvasHeight,
    config.maxWidthPercent ?? 0.85,
    maxHeightPercent
  )

  // Center anchor for proper positioning
  sprite.anchor.set(0.5)

  return { sprite, scale }
}
