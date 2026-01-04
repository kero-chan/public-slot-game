<template>
  <div class="grid-header">
    <div class="header-background" :class="headerBackgroundClass" :style="headerStyle">
      <div class="header-content">
        <!-- Multiplier Values Row -->
        <div class="multipliers-row">
          <div
            v-for="multiplier in multiplierValues"
            :key="multiplier"
            class="multiplier-item"
            :class="{ active: multiplier === value }"
          >
            <img
              v-if="getMultiplierImagePath(multiplier)"
              :src="getMultiplierImagePath(multiplier)"
              class="multiplier-sprite"
              :alt="`x${multiplier}`"
              @error="onImageError"
            />
            <div v-else class="fallback-multiplier">x{{ multiplier }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

// Import glyph images directly (unified naming - no color variants)
import glyphX1 from '@/assets/images/japaneseOmakase/glyphs/glyph_x1.webp'
import glyphX2 from '@/assets/images/japaneseOmakase/glyphs/glyph_x2.webp'
import glyphX3 from '@/assets/images/japaneseOmakase/glyphs/glyph_x3.webp'
import glyphX4 from '@/assets/images/japaneseOmakase/glyphs/glyph_x4.webp'
import glyphX5 from '@/assets/images/japaneseOmakase/glyphs/glyph_x5.webp'
import glyphX6 from '@/assets/images/japaneseOmakase/glyphs/glyph_x6.webp'
import glyphX10 from '@/assets/images/japaneseOmakase/glyphs/glyph_x10.webp'
import headerBg from "@/assets/images/japaneseOmakase/background/background_multiplier.webp";

interface GridHeaderProps {
  mode: "normal" | "free_spin";
  value: number;
  multiplierValues?: number[];
}

const props = withDefaults(defineProps<GridHeaderProps>(), {
  mode: "normal",
  value: 1,
  multiplierValues: () => [1, 2, 3, 5],
});

// Multiplier image mappings (single version for all states)
const multiplierImages: Record<number, string> = {
  1: glyphX1,
  2: glyphX2,
  3: glyphX3,
  4: glyphX4,
  5: glyphX5,
  6: glyphX6,
  10: glyphX10,
};

const headerBgSrc = headerBg;

// Helper function to get multiplier image path
const getMultiplierImagePath = (multiplierValue: number): string => {
  return multiplierImages[multiplierValue] || glyphX1;
};

// Computed class for header background styling
const headerBackgroundClass = computed(() => {
  return {
    "free-spin-mode": props.mode === "free_spin",
    "normal-mode": props.mode === "normal",
  };
});

const headerStyle = computed(() => ({
  backgroundImage: `url(${headerBgSrc})`,
  backgroundSize: "cover",
  backgroundPosition: "center",
  backgroundRepeat: "no-repeat",
}));

/**
 * Handle image load errors by showing fallback
 */
function onImageError(event: Event) {
  const img = event.target as HTMLImageElement;
  const container = img.closest(".multiplier-item");
  if (container) {
    img.style.display = "none";
    const fallback = container.querySelector(
      ".fallback-multiplier"
    ) as HTMLElement;
    if (fallback) {
      fallback.style.display = "block";
    }
  }
}
</script>

<style scoped lang="scss">
.grid-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding-bottom: 12px;
}

.header-background {
  width: 100%;
  min-height: 70px;
  background-size: 100% 100% !important;
  background-position: center !important;
  background-repeat: no-repeat !important;
  display: flex;
  align-items: center;
  justify-content: center;
  
  // Responsive adjustments for mobile
  @media (max-width: 768px) {
    min-height: 60px;
  }
  
  @media (max-width: 480px) {
    min-height: 50px;
  }
}

.header-content {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  
  @media (max-width: 768px) {
    padding: 6px 12px;
  }
  
  @media (max-width: 480px) {
    padding: 4px 8px;
  }
}

.multipliers-row {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  gap: 2%;
  max-width: 600px;
  
  @media (max-width: 768px) {
    max-width: 100%;
    gap: 1.5%;
  }
}

.multiplier-item {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 12%;
  flex-shrink: 0;

  img {
    width: 100%;
    height: auto;
    max-width: 100%;
  }
  
  @media (max-width: 480px) {
    width: 14%;
  }
}

.multiplier-sprite {
  width: 100%;
  height: auto;
  display: block;
  image-rendering: crisp-edges;
}

.fallback-multiplier {
  display: none;
  color: #ffffff;
  font-weight: bold;
  font-size: 14px;
  
  @media (max-width: 768px) {
    font-size: 12px;
  }
  
  @media (max-width: 480px) {
    font-size: 10px;
  }
}
</style>
