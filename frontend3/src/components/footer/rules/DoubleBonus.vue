<template>
  <RuleSection title="Double Bonus">
    <div class="multiplier-description">
      <ul class="multiplier-rules-list">
        <li>
          During any spin, all wins are multiplied by the multiplier value displayed above the reels, starting at x1 for the first round.
        </li>
        <li>Any win in the first round increases the multiplier to x2 for the second round.</li>
        <li>Any win in the second round increases the multiplier to x3 for the third round.</li>
        <li>
          Any win in the third round increases the multiplier to x5 for the fourth round and beyond.
        </li>
      </ul>
    </div>
  </RuleSection>
</template>

<script setup lang="ts">
import RuleSection from "./RuleSection.vue";
</script>

<style scoped lang="scss">
.multiplier-description {
  margin-top: 8px;
  text-align: left;
}

.multiplier-rules-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;

  li {
    font-size: 15px;
    line-height: 1.6;
    color: #E0E0E0;
    font-weight: 400;
    padding: 0;
    margin: 0;
    text-align: left;
  }
}

/* Responsive Design */
@media (max-width: 768px) {
  .multiplier-rules-list {
    gap: 14px;

    li {
      font-size: 14px;
      line-height: 1.5;
    }
  }
}

@media (max-width: 480px) {
  .multiplier-rules-list {
    gap: 12px;

    li {
      font-size: 13px;
      line-height: 1.4;
    }
  }
}
</style>
