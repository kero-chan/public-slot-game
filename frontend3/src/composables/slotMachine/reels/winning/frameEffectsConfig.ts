/**
 * Configuration for winning frame effects
 * Switch between different visual styles by changing the ACTIVE_EFFECT
 */

export enum WinningFrameEffectType {
  GLOWING_ENERGY = 'GLOWING_ENERGY',           // Glowing energy border with orbiting particles
  LIGHTNING = 'LIGHTNING',                     // Electric arcs and lightning bolts
  MAGICAL_RUNES = 'MAGICAL_RUNES',            // Rotating mystical runes/symbols
  RING_WAVES = 'RING_WAVES',                  // Expanding shockwave rings
  DRAGON_SCALES = 'DRAGON_SCALES',            // Dragon scale pattern with fire
  CRYSTALLINE_SHARDS = 'CRYSTALLINE_SHARDS',  // Orbiting crystal shards
  IMAGE_FRAME = 'IMAGE_FRAME'                 // Image-based frames (symbol-specific)
}

/**
 * CONFIGURATION: Change this to switch between different winning frame effects
 */
export const ACTIVE_WINNING_FRAME_EFFECT: WinningFrameEffectType = WinningFrameEffectType.IMAGE_FRAME

/**
 * Effect-specific configuration
 */
export const FRAME_EFFECT_CONFIG = {
  [WinningFrameEffectType.GLOWING_ENERGY]: {
    particleCount: 12,           // Number of orbiting particles
    particleSize: 4,             // Base particle size
    orbitSpeed: 0.02,            // Rotation speed
    pulseSpeed: 2,               // Pulse animation speed
    colors: {
      normal: [0x00ffff, 0x0099ff, 0x66ccff],  // Cyan/blue particles
      bonus: [0xff00ff, 0xaa00ff, 0xff66ff]    // Purple/pink particles
    }
  },

  [WinningFrameEffectType.LIGHTNING]: {
    boltCount: 8,                // Number of lightning bolts
    boltSegments: 5,             // Segments per bolt
    arcIntensity: 0.3,           // How jagged the lightning is
    flashSpeed: 0.1,             // Speed of lightning flashes
    colors: {
      normal: [0xffff00, 0xffffff],  // Yellow/white lightning
      bonus: [0xff00ff, 0xff66ff]    // Purple lightning
    }
  },

  [WinningFrameEffectType.MAGICAL_RUNES]: {
    runeCount: 4,                // Number of runes
    runeSize: 16,                // Size of each rune
    rotationSpeed: 0.01,         // Rune rotation speed
    fadeSpeed: 2,                // Appear/disappear speed
    runeSymbols: ['⚡', '✦', '◆', '✧', '⬡', '◉'],  // Symbols to use
    colors: {
      normal: [0xffd700, 0xffaa00],  // Gold runes
      bonus: [0xff00ff, 0xaa00ff]    // Purple runes
    }
  },

  [WinningFrameEffectType.RING_WAVES]: {
    waveCount: 3,                // Number of simultaneous waves
    waveInterval: 0.5,           // Seconds between waves
    maxRadius: 100,              // Maximum expansion radius
    waveSpeed: 80,               // Expansion speed (pixels/second)
    colors: {
      normal: [0xffdd00, 0xffaa00, 0xff8800],  // Gold waves
      bonus: [0xff00ff, 0xaa00ff, 0xff66ff]    // Purple waves
    }
  },

  [WinningFrameEffectType.DRAGON_SCALES]: {
    scaleCount: 16,              // Number of scales around border
    scaleSize: 8,                // Size of each scale
    shimmerSpeed: 3,             // Shimmer animation speed
    emberCount: 6,               // Number of fire embers
    colors: {
      normal: [0xff6600, 0xffaa00, 0xffd700],  // Orange/gold scales
      bonus: [0xff0000, 0xff6600, 0xffaa00]    // Red/orange scales
    }
  },

  [WinningFrameEffectType.CRYSTALLINE_SHARDS]: {
    shardCount: 8,               // Number of crystal shards
    shardSize: 12,               // Size of each shard
    orbitSpeed: 0.03,            // Orbit rotation speed
    sparkleCount: 16,            // Number of sparkles
    colors: {
      normal: [0x00ffff, 0x66ccff, 0xffffff],  // Cyan/white crystals
      bonus: [0xff00ff, 0xcc66ff, 0xffffff]    // Purple/white crystals
    }
  }
}
