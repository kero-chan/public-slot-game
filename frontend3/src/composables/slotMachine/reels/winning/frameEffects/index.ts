/**
 * Winning Frame Effects - Index
 * Exports all frame effect implementations and factory
 */

export { BaseFrameEffect, type WinningFrameEffect } from './baseEffect'
export { GlowingEnergyEffect } from './glowingEnergyEffect'
export { LightningEffect } from './lightningEffect'
export { MagicalRunesEffect } from './magicalRunesEffect'
export { RingWavesEffect } from './ringWavesEffect'
export { DragonScalesEffect } from './dragonScalesEffect'
export { CrystallineShardsEffect } from './crystallineShardsEffect'
export { ImageFrameEffect } from './imageFrameEffect'

import { WinningFrameEffectType, ACTIVE_WINNING_FRAME_EFFECT } from '../frameEffectsConfig'
import type { WinningFrameEffect } from './baseEffect'
import { GlowingEnergyEffect } from './glowingEnergyEffect'
import { LightningEffect } from './lightningEffect'
import { MagicalRunesEffect } from './magicalRunesEffect'
import { RingWavesEffect } from './ringWavesEffect'
import { DragonScalesEffect } from './dragonScalesEffect'
import { CrystallineShardsEffect } from './crystallineShardsEffect'
import { ImageFrameEffect } from './imageFrameEffect'

/**
 * Factory function to create the active winning frame effect
 * Uses the ACTIVE_WINNING_FRAME_EFFECT from config
 */
export function createWinningFrameEffect(): WinningFrameEffect {
  switch (ACTIVE_WINNING_FRAME_EFFECT) {
    case WinningFrameEffectType.GLOWING_ENERGY:
      return new GlowingEnergyEffect()

    case WinningFrameEffectType.LIGHTNING:
      return new LightningEffect()

    case WinningFrameEffectType.MAGICAL_RUNES:
      return new MagicalRunesEffect()

    case WinningFrameEffectType.RING_WAVES:
      return new RingWavesEffect()

    case WinningFrameEffectType.DRAGON_SCALES:
      return new DragonScalesEffect()

    case WinningFrameEffectType.CRYSTALLINE_SHARDS:
      return new CrystallineShardsEffect()

    case WinningFrameEffectType.IMAGE_FRAME:
    default:
      return new ImageFrameEffect()
  }
}
