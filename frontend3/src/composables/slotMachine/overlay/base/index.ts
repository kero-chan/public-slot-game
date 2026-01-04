/**
 * Overlay Base Utilities
 *
 * Shared utilities for overlay animations:
 * - TimerManager: Session-based timer management to prevent stale callbacks
 * - ParticleSystem: Unified particle creation and physics
 * - ScreenShake: Screen shake effect management
 * - Graphics helpers: Common visual elements (backgrounds, panels, text, buttons)
 */

// Types
export * from './types'

// Classes
export { TimerManager, createTimerManager } from './TimerManager'
export { ParticleSystem, createParticleSystem } from './ParticleSystem'
export { ScreenShake, createScreenShake } from './ScreenShake'

// Graphics helpers
export {
  clearContainer,
  createDarkOverlay,
  createStarburstRays,
  createGlowPanel,
  createNumberDisplay,
  createStyledText,
  createAssetSprite,
  createCornerSparkles,
  createGlowButton
} from './graphics'
