<template>
  <div class="ways-symbols-block">
    <div
      v-for="(column, colIndex) in gridData"
      :key="colIndex"
      class="reel-column"
    >
      <div
        v-for="(hasSymbol, rowIndex) in column"
        :key="rowIndex"
        class="reel-cell"
        :class="{ empty: !hasSymbol }"
      >
        <div v-if="hasSymbol" class="paytable-icon">
          <img
            src="@/assets/images/japaneseOmakase/menu/symbol_icon.webp"
            alt="Symbol"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  gridData: boolean[][];
}>();
</script>

<style scoped lang="scss">
.ways-symbols-block {
  display: flex;
  gap: 4px;
}

.reel-column {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.reel-cell {
  width: 35px;
  height: 35px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 4px;
  display: flex;
  justify-content: center;
  align-items: center;
  background: rgba(0, 0, 0, 0.2);

  &.empty {
    background: rgba(0, 0, 0, 0.1);
  }

  .paytable-icon {
    display: grid;
    img {
      width: 25px;
      height: auto;
    }
  }
}

@media (max-width: 768px) {
  .reel-cell {
    width: 50px;
    height: 42px;

    .paytable-icon img {
      width: 32px;
    }
  }
}
</style>
