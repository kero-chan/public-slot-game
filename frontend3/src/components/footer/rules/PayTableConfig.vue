<template>
  <RuleSection title="Symbol Payout Values">
    <div class="paytable-container">
      <!-- Special Symbols Section -->
      <div class="special-symbols-section">
        <h3 class="section-title">Special Symbols</h3>
        <div class="special-symbols-grid">
          <!-- Wild Symbol Card -->
          <div class="symbol-card special-card wild-card">
            <div class="symbol-header">
              <div class="symbol-icon">
                <img
                  :src="tileWild"
                  class="symbol-sprite-image"
                  alt="Wild"
                  @error="onImageError"
                />
              </div>
              <div class="symbol-badge wild-badge">WILD</div>
            </div>
            <div class="symbol-description">Substitutes for other symbols to help you win more</div>
          </div>

          <!-- Scatter Symbol Card -->
          <div class="symbol-card special-card scatter-card">
            <div class="symbol-header">
              <div class="symbol-icon">
                <img
                  :src="tileBonus"
                  class="symbol-sprite-image"
                  alt="Bonus"
                  @error="onImageError"
                />
              </div>
              <div class="symbol-badge scatter-badge">BONUS</div>
            </div>
            <div class="symbol-description">
              {{ FREE_SPINS_CONFIG.minScatters }} or more triggers
              {{ FREE_SPINS_CONFIG.awards[FREE_SPINS_CONFIG.minScatters] }}
              free spins
            </div>
          </div>
        </div>
      </div>

      <!-- Regular Symbols Section -->
      <div class="regular-symbols-section">
        <h3 class="section-title">Symbol Payouts</h3>
        <div class="symbols-grid">
          <div
            v-for="(symbolData, symbolKey) in paytableSymbols"
            :key="symbolKey"
            class="symbol-card payout-card"
          >
            <div class="symbol-icon">
              <img
                :src="symbolData.imagePath"
                class="symbol-sprite-image"
                :alt="symbolData.title"
                @error="onImageError"
              />
            </div>

            <div class="symbol-info">
              <div class="symbol-name">{{ symbolData.title }}</div>
              <div class="payout-table">
                <div
                  v-for="count in [5, 4, 3]"
                  :key="count"
                  class="payout-row"
                  v-show="getPayoutValue(symbolKey, count)"
                >
                  <span class="payout-multiplier">{{ count }}x</span>
                  <span class="payout-dots"></span>
                  <span class="payout-value">{{
                    getPayoutValue(symbolKey, count)
                  }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </RuleSection>
</template>

<script setup lang="ts">
import RuleSection from "./RuleSection.vue";
import {
  PAYTABLE_CONFIG,
  FREE_SPINS_CONFIG,
  type PaytableSymbol,
  type SymbolCount,
} from "@/config/constants";
import { computed } from "vue";

// Import tile images directly
import tileFa from "@/assets/images/japaneseOmakase/tiles/tile_fa.webp";
import tileZhong from "@/assets/images/japaneseOmakase/tiles/tile_zhong.webp";
import tileBai from "@/assets/images/japaneseOmakase/tiles/tile_bai.webp";
import tileBawan from "@/assets/images/japaneseOmakase/tiles/tile_bawan.webp";
import tileWusuo from "@/assets/images/japaneseOmakase/tiles/tile_wusuo.webp";
import tileWutong from "@/assets/images/japaneseOmakase/tiles/tile_wutong.webp";
import tileLiangsuo from "@/assets/images/japaneseOmakase/tiles/tile_liangsuo.webp";
import tileLiangtong from "@/assets/images/japaneseOmakase/tiles/tile_liangtong.webp";
import tileWild from "@/assets/images/japaneseOmakase/tiles/tile_wild.webp";
import tileBonus from "@/assets/images/japaneseOmakase/tiles/tile_bonus.webp";

// Symbol configuration with direct image paths
const paytableSymbols = computed(() => ({
  fa: {
    title: "Phoenix Fire",
    fallback: "Phoenix Fire",
    imagePath: tileFa,
  },
  zhong: {
    title: "Rune Gem",
    fallback: "Rune Gem",
    imagePath: tileZhong,
  },
  bai: {
    title: "White Dragon",
    fallback: "White Dragon",
    imagePath: tileBai,
  },
  bawan: {
    title: "Wind Spirit",
    fallback: "Wind Spirit",
    imagePath: tileBawan,
  },
  wusuo: {
    title: "Earthy Quartz",
    fallback: "Earthy Quartz",
    imagePath: tileWusuo,
  },
  wutong: {
    title: "Fire Spirit",
    fallback: "Fire Spirit",
    imagePath: tileWutong,
  },
  liangsuo: {
    title: "Pale Breeze",
    fallback: "Pale Breeze",
    imagePath: tileLiangsuo,
  },
  liangtong: {
    title: "Cloud Essence",
    fallback: "Cloud Essence",
    imagePath: tileLiangtong,
  },
}));

// Helper function to get payout value safely
const getPayoutValue = (symbolKey: string, count: number): number => {
  const symbol = symbolKey as PaytableSymbol;
  const symbolCount = count as SymbolCount;
  return PAYTABLE_CONFIG[symbol]?.[symbolCount] || 0;
};

/**
 * Handle image load errors by hiding the image and showing fallback
 */
function onImageError(event: Event) {
  const img = event.target as HTMLImageElement;
  const container = img.closest(".symbol-image-container");
  if (container) {
    img.style.display = "none";
    const fallback = container.querySelector(".fallback-symbol") as HTMLElement;
    if (fallback) {
      fallback.style.display = "block";
    }
  }
}
</script>

<style scoped lang="scss">
.paytable-container {
  width: 100%;
  color: #fff;
  padding: 0 8px;
}

.special-symbols-section,
.regular-symbols-section {
  margin-bottom: 32px;

  &:last-child {
    margin-bottom: 0;
  }
}

.section-title {
  font-size: 1.2rem;
  color: #ffd700;
  font-weight: 700;
  margin-bottom: 24px;
  text-align: center;
  text-transform: uppercase;
  letter-spacing: 1px;
  background: linear-gradient(45deg, #ffd700, #ff8c00);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  position: relative;

  &::after {
    content: "";
    position: absolute;
    bottom: -8px;
    left: 50%;
    transform: translateX(-50%);
    width: 60px;
    height: 2px;
    background: linear-gradient(90deg, #ffd700, #ff8c00);
    border-radius: 1px;
    box-shadow: 0 2px 8px rgba(255, 215, 0, 0.3);
  }
}

.special-symbols-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 8px;
}

.symbols-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 12px;
  max-width: 100%;
}

.symbol-card {
  background: linear-gradient(
    145deg,
    rgba(26, 26, 46, 0.95) 0%,
    rgba(22, 33, 62, 0.9) 50%,
    rgba(15, 52, 96, 0.85) 100%
  );
  border-radius: 16px;
  padding: 16px;
  border: 2px solid rgba(52, 152, 219, 0.2);
  backdrop-filter: blur(10px);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;

  &::before {
    content: "";
    position: absolute;
    top: 0;
    left: -100%;
    width: 100%;
    height: 100%;
    background: linear-gradient(
      90deg,
      transparent,
      rgba(255, 255, 255, 0.1),
      transparent
    );
    transition: left 0.6s ease;
  }

  &:hover {
    transform: translateY(-4px);
    border-color: rgba(52, 152, 219, 0.5);
    box-shadow: 0 8px 25px rgba(52, 152, 219, 0.15),
      0 0 20px rgba(52, 152, 219, 0.1);

    &::before {
      left: 100%;
    }
  }

  &:active {
    transform: translateY(-2px);
  }
}

.special-card {
  text-align: center;

  .symbol-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
  }

  .symbol-description {
    font-size: 0.85rem;
    color: #bdc3c7;
    line-height: 1.4;
    text-align: center;
  }
}

.wild-card {
  border-color: rgba(155, 89, 182, 0.4);
  background: linear-gradient(
    145deg,
    rgba(142, 68, 173, 0.1) 0%,
    rgba(155, 89, 182, 0.05) 100%
  );
  
  .symbol-icon {
    width: 30%;

    @media screen and (max-width: 300px) {
      width: 50%;
    }
  }
}

.scatter-card {
  border-color: rgba(231, 76, 60, 0.4);
  background: linear-gradient(
    145deg,
    rgba(231, 76, 60, 0.1) 0%,
    rgba(243, 156, 18, 0.05) 100%
  );

  .symbol-icon {
    width: 30%;

    @media screen and (max-width: 300px) {
      width: 50%;
    }
  }
}

.payout-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 12px;
}

.symbol-icon {
  width: 50%;
  aspect-ratio: 455 / 593;
  transition: all 0.3s ease;
  overflow: hidden;

  &:hover {
    transform: scale(1.1);
  }
}

.symbol-sprite-image {
  display: block;
  image-rendering: crisp-edges;
  transition: all 0.3s ease;
  width: 100%;
  height: 100%;

  &:hover {
    filter: brightness(1.1) drop-shadow(0 0 8px rgba(255, 215, 0, 0.6));
  }
}

.fallback-symbol {
  font-size: 1.5rem;
  color: #ffd700;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.5);
  font-weight: bold;
  display: none;
  animation: pulse 2s infinite;
}

.symbol-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.wild-badge {
  background: linear-gradient(135deg, #8e44ad, #9b59b6);
  color: #fff;
  border: 1px solid rgba(155, 89, 182, 0.5);
}

.scatter-badge {
  background: linear-gradient(135deg, #e74c3c, #f39c12);
  color: #fff;
  border: 1px solid rgba(231, 76, 60, 0.5);
}

.symbol-info {
  flex: 1;
  width: 100%;
}

.symbol-name {
  font-size: 1rem;
  color: #ffd700;
  font-weight: 600;
  margin-bottom: 12px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
}

.payout-table {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-height: 60px;
}

.payout-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 8px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 8px;
  transition: all 0.2s ease;

  &:hover {
    background: rgba(52, 152, 219, 0.1);
    transform: translateX(2px);
  }
}

.payout-multiplier {
  font-size: 0.85rem;
  color: #95e1d3;
  font-weight: 600;
  min-width: 24px;
  text-align: left;
  background: rgba(149, 225, 211, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
}

.payout-dots {
  flex: 1;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(189, 195, 199, 0.3) 25%,
    rgba(189, 195, 199, 0.6) 50%,
    rgba(189, 195, 199, 0.3) 75%,
    transparent 100%
  );
  margin: 0 8px;
  position: relative;

  &::after {
    content: "";
    position: absolute;
    top: -1px;
    left: 50%;
    transform: translateX(-50%);
    width: 2px;
    height: 3px;
    background: rgba(189, 195, 199, 0.6);
    border-radius: 1px;
  }
}

.payout-value {
  font-size: 0.9rem;
  color: #f39c12;
  font-weight: 700;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
  min-width: 24px;
  text-align: right;
  background: rgba(243, 156, 18, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  transition: all 0.3s ease;

  &:hover {
    color: #fff;
    background: rgba(243, 156, 18, 0.2);
    text-shadow: 0 0 8px rgba(243, 156, 18, 0.6);
  }
}

/* Pulse animation for fallback symbols */
@keyframes pulse {
  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }

  50% {
    opacity: 0.7;
    transform: scale(1.05);
  }
}

/* Mobile Responsive Design */
@media (max-width: 768px) {
  .paytable-container {
    padding: 0 4px;
  }

  .section-title {
    font-size: 1.1rem;
    margin-bottom: 12px;
  }

  .special-symbols-grid {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .symbols-grid {
    grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
    gap: 8px;
  }

  .symbol-card {
    padding: 12px;
    border-radius: 12px;
  }

  .symbol-name {
    font-size: 0.9rem;
    margin-bottom: 8px;
  }

  .payout-row {
    padding: 3px 6px;
  }

  .payout-multiplier,
  .payout-value {
    font-size: 0.8rem;
    padding: 1px 4px;
  }

  .payout-dots {
    margin: 0 4px;
  }

  .special-card .symbol-description {
    font-size: 0.8rem;
  }

  .symbol-badge {
    font-size: 0.7rem;
    padding: 3px 8px;
  }
}

@media (max-width: 480px) {
  .symbols-grid {
    grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
    gap: 6px;
  }

  .symbol-card {
    padding: 10px;
  }

  .symbol-name {
    font-size: 0.85rem;
  }

  .payout-table {
    gap: 3px;
    min-height: 50px;
  }

  .special-symbols-section,
  .regular-symbols-section {
    margin-bottom: 24px;
  }
}

/* Dark theme enhancement */
@media (prefers-color-scheme: dark) {
  .symbol-card {
    background: linear-gradient(
      145deg,
      rgba(13, 13, 23, 0.95) 0%,
      rgba(11, 17, 31, 0.9) 50%,
      rgba(8, 26, 48, 0.85) 100%
    );
    border-color: rgba(52, 152, 219, 0.15);
  }
}

/* High contrast mode support */
@media (prefers-contrast: high) {
  .symbol-card {
    border-width: 3px;
    border-color: #3498db;
  }

  .symbol-badge {
    border-width: 2px;
  }

  .payout-value {
    color: #fff;
    background: #f39c12;
  }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .symbol-card,
  .symbol-icon,
  .symbol-sprite-image,
  .payout-row,
  .payout-value {
    transition: none;
  }

  .symbol-card::before {
    display: none;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
    }

    50% {
      opacity: 0.8;
    }
  }
}
</style>
