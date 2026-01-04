import type { Sprite } from 'pixi.js'

export function applyTileVisuals(
  sprite: Sprite | null | undefined,
  alpha: number = 1,
  highlight: boolean = false,
  hasHighlights: boolean = false,
  isLightningAnimating: boolean = false,
  isOtherWinningTile: boolean = false
): void {
  if (!sprite) return
  sprite.alpha = alpha

  // Apply tint based on state
  if (highlight) {
      sprite.tint = 0xffffff // Bright white tint for winning tiles
  } else if (isLightningAnimating) {
      // During lightning animation, darken all non-highlighted tiles
      // This includes both non-winning tiles AND other winning tiles from different symbols
      sprite.tint = 0x404040 // Dark gray for dramatic lightning effect
  } else if (isOtherWinningTile) {
      // This shouldn't normally be reached since isOtherWinningTile is only set during lightning
      sprite.tint = 0x404040
  } else if (hasHighlights) {
      // Dim for non-winning tiles when there are winning tiles (~40% brightness)
      sprite.tint = 0x666666 // Medium-dark gray
  } else {
      sprite.tint = 0xffffff // White (normal)
  }
}

/**
 * Apply anticipation-specific visuals with stronger contrast
 * - Highlighted bonus tiles: Brighter with light yellow tint
 * - Dimmed non-bonus tiles: Much darker for dramatic effect
 */
export function applyAnticipationVisuals(
  sprite: Sprite | null | undefined,
  highlight: boolean = false,
  shouldDim: boolean = false
): void {
  if (!sprite) return

  if (highlight) {
      // Bright light yellow/gold tint for bonus tiles during anticipation
      sprite.tint = 0xffffcc // Light yellow-white for extra brightness
      sprite.alpha = 1
  } else if (shouldDim) {
      // Very strong dim for non-bonus tiles during anticipation (~12% brightness)
      sprite.tint = 0x202020 // Near-black for dramatic contrast
      sprite.alpha = 1
  } else {
      sprite.tint = 0xffffff // White (normal)
      sprite.alpha = 1
  }
}
