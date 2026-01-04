<template>
  <div class="tile-grid-golden">
    <!-- Main Grid Display -->
    <GridHeader mode="normal" :value="1" />
    <TileGrid
      :grid-data="goldenGridData"
      :show-positions="showPositions"
      :tile-size="tileSize"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import TileGrid from "./TileGrid.vue";
import GridHeader from "./GridHeader.vue";

const goldenGridData = [
  ["liangtong", "liangtong", "gold", "bonus", "wutong"],
  ["bai", "fa", "wusuo_gold", "fa", "zhong"],
  ["liangtong", "wutong", "wutong", "liangtong", "bawan"],
  ["bai", "fa", "fa_gold", "wutong", "bai"],
];

// Reactive state
const showPositions = ref(false);
const tileSize = ref(60);
</script>

<style scoped lang="scss"></style>
