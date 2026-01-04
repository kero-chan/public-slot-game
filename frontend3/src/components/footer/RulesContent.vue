<template>
  <div class="rules-content">
    <PayTableConfig />
    <GoldenSymbol />
    <DoubleBonus />
    <FreeSpinMode />
    <Road1024 />
  </div>
</template>

<script setup lang="ts">
import PayTableConfig from "./rules/PayTableConfig.vue";
import GoldenSymbol from "./rules/GoldenSymbol.vue";
import DoubleBonus from "./rules/DoubleBonus.vue";
import FreeSpinMode from "./rules/FreeSpinMode.vue";
import Road1024 from "./rules/1024Road.vue";
</script>

<style scoped lang="scss">
.rules-content {
  color: #e0e0e0;
}
</style>
