<template>
  <div class="payout-data">
    <div
      v-for="(payout, index) in payoutData"
      :key="index"
      class="payout-row"
      :class="{ highlighted: payout.highlighted }"
    >
      <span class="count">{{ payout.count }}</span>
      <span class="value">{{ payout.value }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
interface PayoutItem {
  count: number;
  value: number;
  highlighted?: boolean;
}

defineProps<{
  payoutData: PayoutItem[]
}>();
</script>

<style scoped lang="scss">
.payout-data {
  .payout-row {
    height: 16px;
    display: flex;
    align-items: center;
    margin-bottom: 4px;

    &.highlighted {
      color: rgb(180, 120, 80);
      font-weight: bold;
    }

    &:last-child {
      margin-bottom: 0;
    }

    .count {
      text-align: left;
      padding-right: 10px;
      min-width: 20px;
    }

    .value {
      text-align: left;
    }
  }
}
</style>
