# Winning Frame Effects System

A flexible, configurable system for winning tile frame animations with 6 different visual effects.

## Quick Start

To switch between effects, simply change one line in `frameEffectsConfig.ts`:

```typescript
export const ACTIVE_WINNING_FRAME_EFFECT: WinningFrameEffectType = WinningFrameEffectType.GLOWING_ENERGY
```

## Available Effects

### 1. GLOWING_ENERGY (Default)
**Glowing Energy Border with Orbiting Particles**
- Pulsing glow border
- 12 orbiting energy particles
- Color-coded: Cyan/blue (normal) | Purple/pink (bonus)
- Smooth, mesmerizing motion

### 2. LIGHTNING
**Electric Arcs and Lightning Bolts**
- Crackling lightning dancing around borders
- 8 electric arcs with jagged paths
- Random flashing effects
- Color-coded: Yellow/white (normal) | Purple (bonus)

### 3. MAGICAL_RUNES
**Rotating Mystical Symbols**
- Ancient runes orbiting the tile
- Symbols: ⚡ ✦ ◆ ✧ ⬡ ◉
- Fade in/out animations
- Auto-changing symbols
- Color-coded: Gold (normal) | Purple (bonus)

### 4. RING_WAVES
**Expanding Shockwave Rings**
- Ripple-in-water effect
- Multiple waves expanding outward
- Continuous wave spawning
- Color-coded: Gold waves (normal) | Purple waves (bonus)

### 5. DRAGON_SCALES
**Dragon Scale Pattern with Fire**
- Shimmering dragon scales around border
- Fire embers floating upward
- Asian/dragon theme
- Color-coded: Orange/gold (normal) | Red/orange (bonus)

### 6. CRYSTALLINE_SHARDS
**Orbiting Crystal Shards**
- Crystal shards rotating around tile
- Sparkle particles
- Faceted border effect
- Color-coded: Cyan/white (normal) | Purple/white (bonus)

### 7. CLASSIC
**Original Simple Glow**
- The original winning frame effect
- Simple glowing border
- Lightweight, proven design

## Configuration

Each effect has detailed configuration in `frameEffectsConfig.ts`:

```typescript
export const FRAME_EFFECT_CONFIG = {
  [WinningFrameEffectType.GLOWING_ENERGY]: {
    particleCount: 12,
    particleSize: 4,
    orbitSpeed: 0.02,
    pulseSpeed: 2,
    colors: {
      normal: [0x00ffff, 0x0099ff, 0x66ccff],
      bonus: [0xff00ff, 0xaa00ff, 0xff66ff]
    }
  },
  // ... more configs
}
```

### Customizable Properties:
- **Particle/Element counts**: How many particles/shards/runes to show
- **Animation speeds**: Orbit, rotation, pulse, shimmer speeds
- **Sizes**: Particle sizes, shard sizes, rune sizes
- **Colors**: Separate color palettes for normal and bonus tiles
- **Timing**: Flash intervals, wave intervals, fade speeds

## Architecture

### File Structure
```
frameEffects/
├── README.md                      # This file
├── baseEffect.ts                  # Base class for all effects
├── glowingEnergyEffect.ts         # Effect 1 implementation
├── lightningEffect.ts             # Effect 2 implementation
├── magicalRunesEffect.ts          # Effect 3 implementation
├── ringWavesEffect.ts             # Effect 4 implementation
├── dragonScalesEffect.ts          # Effect 5 implementation
├── crystallineShardsEffect.ts     # Effect 6 implementation
└── index.ts                       # Factory and exports
```

### How It Works

1. **Configuration** (`frameEffectsConfig.ts`)
   - Set `ACTIVE_WINNING_FRAME_EFFECT` to choose effect
   - Customize effect parameters

2. **Factory** (`frameEffects/index.ts`)
   - `createWinningFrameEffect()` creates the active effect
   - Returns null for CLASSIC mode

3. **Manager** (`winningComposer.ts`)
   - Creates effect instances per winning tile
   - Calls `update(timestamp)` every frame
   - Manages lifecycle (show/hide/cleanup)

4. **Rendering** (`reels/index.ts`)
   - Calls `winningFrames.update(timestamp)` in draw loop
   - Effects animate smoothly at 60fps

## Creating Custom Effects

To create a new effect:

1. Create a new file in `frameEffects/` (e.g., `myCustomEffect.ts`)
2. Extend `BaseFrameEffect` class
3. Implement the `update()` method
4. Add to `WinningFrameEffectType` enum
5. Add config to `FRAME_EFFECT_CONFIG`
6. Add case in factory function

Example:
```typescript
export class MyCustomEffect extends BaseFrameEffect {
  update(timestamp: number, sprite: Sprite, x: number, y: number, isBonus: boolean): void {
    this.updateDimensions(sprite)
    this.matchSpriteTransform(sprite, x, y)
    // Your custom animation here
  }
}
```

## Performance Notes

- Effects use BLEND_MODES.ADD for glowing effects
- Particle counts are optimized for 60fps
- Effects are only updated when visible
- Cleanup automatically removes unused effects

## Tips

- **Testing**: Try all effects to find the best fit for your game theme
- **Colors**: Adjust color palettes to match your game's aesthetic
- **Speed**: Tweak animation speeds for desired intensity
- **Particles**: Reduce particle counts if performance is an issue

## Examples

### Subtle Effect (Low Intensity)
```typescript
export const ACTIVE_WINNING_FRAME_EFFECT: WinningFrameEffectType = WinningFrameEffectType.RING_WAVES
```

### Dramatic Effect (High Intensity)
```typescript
export const ACTIVE_WINNING_FRAME_EFFECT: WinningFrameEffectType = WinningFrameEffectType.LIGHTNING
```

### Thematic Effect (Asian Theme)
```typescript
export const ACTIVE_WINNING_FRAME_EFFECT: WinningFrameEffectType = WinningFrameEffectType.DRAGON_SCALES
```

---

**Need help?** Check the config file for all customization options!
