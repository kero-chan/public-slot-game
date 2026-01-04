<template>
  <div class="tile-grid-container">
    <div class="tile-grid">
      <div v-for="(row, rowIndex) in gridData" :key="rowIndex" class="tile-row">
        <div v-for="(tile, colIndex) in row.slice(0, 5)" :key="`${rowIndex}-${colIndex}`" class="tile-cell" :class="{
            'special-tile': isSpecialTile(tile),
            'bonus-tile': tile === 'bonus',
          }">
          <div class="tile-wrapper">
            <img v-if="getTileImagePath(tile)" :src="getTileImagePath(tile)" class="tile-sprite" :alt="tile"
              @error="onImageError" />
            <div v-else class="fallback-tile">
              {{ getTileFallback(tile) }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { CONFIG } from "@/config/constants";

// Import tile images directly
import tileBawan from '@/assets/images/japaneseOmakase/tiles/tile_bawan.webp'
import tileWutong from '@/assets/images/japaneseOmakase/tiles/tile_wutong.webp'
import tileWusuo from '@/assets/images/japaneseOmakase/tiles/tile_wusuo.webp'
import tileZhongGold from '@/assets/images/japaneseOmakase/tiles/tile_zhong_gold.webp'
import tileZhong from '@/assets/images/japaneseOmakase/tiles/tile_zhong.webp'
import tileWild from '@/assets/images/japaneseOmakase/tiles/tile_wild.webp'
import tileLiangtong from '@/assets/images/japaneseOmakase/tiles/tile_liangtong.webp'
import tileLiangsuoGold from '@/assets/images/japaneseOmakase/tiles/tile_liangsuo_gold.webp'
import tileBai from '@/assets/images/japaneseOmakase/tiles/tile_bai.webp'
import tileLiangtongGold from '@/assets/images/japaneseOmakase/tiles/tile_liangtong_gold.webp'
import tileFaGold from '@/assets/images/japaneseOmakase/tiles/tile_fa_gold.webp'
import tileWusuoGold from '@/assets/images/japaneseOmakase/tiles/tile_wusuo_gold.webp'
import tileFa from '@/assets/images/japaneseOmakase/tiles/tile_fa.webp'
import tileBonus from '@/assets/images/japaneseOmakase/tiles/tile_bonus.webp'
import tileLiangsuo from '@/assets/images/japaneseOmakase/tiles/tile_liangsuo.webp'
import tileBawanGold from '@/assets/images/japaneseOmakase/tiles/tile_bawan_gold.webp'
import tileBaiGold from '@/assets/images/japaneseOmakase/tiles/tile_bai_gold.webp'
import tileWutongGold from '@/assets/images/japaneseOmakase/tiles/tile_wutong_gold.webp'

interface TileGridProps {
  gridData?: string[][];
  tileSize?: number;
}

const props = withDefaults(defineProps<TileGridProps>(), {
  gridData: () => [
    ["liangtong", "liangtong", "wusuo", "bonus", "wutong"],
    ["bai", "fa", "wusuo_gold", "fa", "zhong"],
    ["liangtong", "wutong", "wutong", "liangtong", "bawan"],
    ["bai", "fa", "fa_gold", "wutong", "bai"],
  ],
  tileSize: 60,
});

// Tile image mappings
const tileImages: Record<string, string> = {
  liangtong: tileLiangtong,
  liangtong_gold: tileLiangtongGold,
  liangsuo: tileLiangsuo,
  liangsuo_gold: tileLiangsuoGold,
  wusuo: tileWusuo,
  wusuo_gold: tileWusuoGold,
  wutong: tileWutong,
  wutong_gold: tileWutongGold,
  bai: tileBai,
  bai_gold: tileBaiGold,
  fa: tileFa,
  fa_gold: tileFaGold,
  zhong: tileZhong,
  zhong_gold: tileZhongGold,
  bawan: tileBawan,
  bawan_gold: tileBawanGold,
  bonus: tileBonus,
  wild: tileWild,
  gold: tileWild,
};

// Helper function to get tile image path
const getTileImagePath = (tileName: string): string => {
  return tileImages[tileName] || '';
};

// Helper function to get fallback display for tiles
const getTileFallback = (tileName: string): string => {
  const fallbackMap: Record<string, string> = {
    liangtong: "二筒",
    liangsuo: "二索",
    wusuo: "五索",
    wusuo_gold: "五索⭐",
    wutong: "五筒",
    bai: "白",
    fa: "發",
    fa_gold: "發⭐",
    zhong: "中",
    bawan: "八萬",
    bonus: "⭐",
    wild: "💎",
  };

  return fallbackMap[tileName] || tileName;
};

// Helper function to check if tile is special
const isSpecialTile = (tileName: string): boolean => {
  return ["bonus", "wild"].includes(tileName);
};

// Helper function to check if tile is gold variant
const isGoldTile = (tileName: string): boolean => {
  return tileName.includes("_gold");
};

/**
 * Handle image load errors by hiding the image and showing fallback
 */
function onImageError(event: Event) {
  const img = event.target as HTMLImageElement;
  const wrapper = img.closest(".tile-wrapper");
  if (wrapper) {
    img.style.display = "none";
    const fallback = wrapper.querySelector(".fallback-tile") as HTMLElement;
    if (fallback) {
      fallback.style.display = "block";
    }
  }
}
</script>

<style scoped lang="scss">
.tile-grid-container {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.tile-grid {
  display: flex;
  flex-direction: column;
}

.tile-row {
  display: flex;
  justify-content: center;
}

.tile-cell {
  position: relative;
  width: v-bind("`${tileSize}px`");
  height: v-bind("`${tileSize * CONFIG.canvas.tileAspectRatio}px`");
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.tile-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.tile-sprite {
  display: block;
  image-rendering: crisp-edges;
  transition: all 0.3s ease;
  width: 100%;
  height: 100%;
}

.fallback-tile {
  font-size: 0.8rem;
  color: #ffd700;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
  font-weight: bold;
  display: none;
  text-align: center;
  line-height: 1;
}

.tile-position {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 10px;
  text-align: center;
  padding: 2px;
  font-family: monospace;
}

/* Mobile Responsive Design */
@media (max-width: 768px) {
  .tile-cell {
    width: v-bind("`${Math.max(tileSize * 0.8, 40)}px`");
    height: v-bind("`${Math.max(tileSize * 0.8, 40) * CONFIG.canvas.tileAspectRatio}px`"
    );
  }

  .fallback-tile {
    font-size: 0.7rem;
  }
}

@media (max-width: 480px) {
  .tile-cell {
    width: v-bind("`${Math.max(tileSize * 0.6, 35)}px`");
    height: v-bind("`${Math.max(tileSize * 0.6, 35) * CONFIG.canvas.tileAspectRatio}px`"
    );
  }

  .fallback-tile {
    font-size: 0.6rem;
  }
}

/* Animation for grid appearance */
@keyframes tileAppear {
  from {
    opacity: 0;
    transform: scale(0.8) translateY(10px);
  }

  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.tile-cell {
  animation: tileAppear 0.3s ease-out;
  animation-delay: calc(var(--row-index, 0) * 0.05s + var(--col-index, 0) * 0.02s);
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .tile-cell {
    transition: none;
    animation: none;
  }

  .tile-sprite {
    transition: none;
  }
}
</style>
