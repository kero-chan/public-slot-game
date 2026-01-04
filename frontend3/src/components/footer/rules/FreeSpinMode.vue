<template>
  <RuleSection title="Free Spin Mode">
    <div class="freespin-container">
      <div class="freespin-title-section">
        <div class="freespin-title">Free Spin Mode</div>

        <div class="scatter-symbols-display">
          <div v-for="index in 3" :key="index" class="tile-container">
            <img :src="tileBonus" class="bonus-tile-image" alt="Bonus" />
          </div>
        </div>

        <div class="freespin-description">
          <ul class="freespin-rules-list">
            <li>
              When 3 Scatter symbols appear anywhere on the reels, Free Spin Mode is triggered and awards 12 free spins. Each additional Scatter symbol triggers 2 more free spins.
            </li>
          </ul>
        </div>

        <div class="freespin-mode-display">
          <div class="content">
            <GridHeader mode="normal" :value="1" />
            <TileGrid
              :grid-data="bonusGridData"
              :show-positions="showPositions"
              :tile-size="tileSize"
            />
          </div>
        </div>

        <div class="freespin-description">
          <ul class="freespin-rules-list">
            <li>
              During Free Spin Mode, the multipliers above the reels increase to x2, x4, x6, and x10 respectively.
            </li>
            <li>Free spins can be re-triggered.</li>
          </ul>
        </div>
      </div>
    </div>
  </RuleSection>
</template>

<script setup lang="ts">
import { ref } from "vue";
import RuleSection from "./RuleSection.vue";
import TileGrid from "../../game/TileGrid.vue";
import GridHeader from "../../game/GridHeader.vue";
// Import tile image directly
import tileBonus from "@/assets/images/japaneseOmakase/tiles/tile_bonus.webp";

const bonusGridData = [
  ["liangtong", "liangtong", "gold", "bonus", "wutong"],
  ["bai", "fa", "bonus", "fa", "zhong"],
  ["liangtong", "wutong", "wutong", "liangtong", "bawan"],
  ["bai", "fa", "fa_gold", "bonus", "bai"],
];

const showPositions = ref(false);
const tileSize = ref(60);
</script>

<style scoped lang="scss">
.freespin-container {
  width: 100%;
}

.freespin-title-section {
  display: inline-block;
  width: 100%;
  border-bottom: 1px solid rgb(40, 40, 52);
  margin-bottom: 20px;
  text-align: center;
}

.freespin-title {
  font-size: 14px;
  color: rgb(255, 255, 255);
  font-weight: normal;
  margin-bottom: 16px;
  line-height: 145%;
  padding: 0 20px;
  direction: ltr;
}

.scatter-symbols-display {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
  padding: 0 20px;
}

.tile-container {
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(
    135deg,
    rgba(255, 215, 0, 0.1),
    rgba(255, 165, 0, 0.05)
  );
  border: 1px solid rgba(255, 215, 0, 0.3);
  border-radius: 8px;
  padding: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3),
    inset 0 1px 2px rgba(255, 255, 255, 0.1);
}

.bonus-tile-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
}

.freespin-mode-display {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  padding: 24px;
  background: linear-gradient(
    135deg,
    rgba(255, 215, 0, 0.08) 0%,
    rgba(255, 165, 0, 0.04) 100%
  );
  border-radius: 12px;
  border: 2px solid rgba(255, 215, 0, 0.2);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
  transition: all 0.3s ease;
}

.freespin-mode-wrapper {
  min-width: 300px;
  width: 300px;
  min-height: 264px;
  height: 264px;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  background: linear-gradient(
    135deg,
    rgba(255, 215, 0, 0.1),
    rgba(255, 165, 0, 0.05)
  );
  border: 1px solid rgba(255, 215, 0, 0.2);
  border-radius: 12px;
}

.freespin-mode-sprite {
  display: block;
  transform: scale(0.5);
  position: relative;
  right: 150px;
  bottom: 132px;
  min-width: 600px;
  width: 600px;
  min-height: 528px;
  height: 528px;
  background-position: -1px -1px;

  // Apply the freespins background class
  &.paytable_win_sprites.freespins {
    background-position: -1px -1px;
  }

  // Placeholder styling for demonstration
  &::before {
    content: "🎰 FREE SPINS 🎰";
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%) scale(1);
    font-size: 1.5rem;
    font-weight: bold;
    color: #ffd700;
    text-shadow: 0 0 10px rgba(255, 215, 0, 0.8),
      0 0 20px rgba(255, 215, 0, 0.6);
    animation: freespin-glow 2s ease-in-out infinite alternate;
  }
}

.freespin-description {
  margin-top: 8px;
  text-align: left;
}

.freespin-rules-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;

  li {
    font-size: 15px;
    line-height: 1.6;
    color: #e0e0e0;
    font-weight: 400;
    padding: 0;
    margin: 0;
    text-align: left;
  }
}

/* Animations */
@keyframes bonus-glow {
  0% {
    opacity: 0.6;
    filter: blur(1px);
  }
  100% {
    opacity: 1;
    filter: blur(0px);
  }
}

@keyframes freespin-glow {
  0% {
    text-shadow: 0 0 10px rgba(255, 215, 0, 0.8),
      0 0 20px rgba(255, 215, 0, 0.6);
  }
  100% {
    text-shadow: 0 0 20px rgba(255, 215, 0, 1), 0 0 40px rgba(255, 215, 0, 0.8),
      0 0 60px rgba(255, 215, 0, 0.6);
  }
}

/* Responsive Design */
@media (max-width: 768px) {
  .freespin-title {
    font-size: 13px;
    padding: 0 15px;
  }

  .scatter-symbols-display {
    gap: 6px;
    margin-bottom: 15px;
    padding: 0 15px;
  }

  .freespin-mode-wrapper {
    min-width: 250px;
    width: 250px;
    min-height: 220px;
    height: 220px;
  }

  .freespin-mode-sprite {
    transform: scale(0.4);
    right: 125px;
    bottom: 110px;

    &::before {
      font-size: 1.2rem;
    }
  }

  .freespin-description {
    font-size: 11px;
    margin: 15px 15px 15px 25px;
  }

  .freespin-rules-list {
    gap: 12px;

    li {
      font-size: 13px;
      line-height: 1.4;
    }
  }
}
</style>
