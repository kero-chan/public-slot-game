<template>
  <RuleSection title="Golden Mahjong Symbols">
    <div class="golden-symbol-container">
      <!-- Grid Display -->
      <div class="golden-symbol-display-wrapper">
        <GoldenSymbolGrid />
      </div>

      <!-- Rules Description -->
      <div class="golden-symbol-description">
        <ul class="golden-rules-list">
          <li>
            During any spin, certain symbols (excluding Wild and Scatter symbols) on reels 2, 3, and/or 4 may appear as golden.
          </li>
          <li>
            After each cascade of new symbols, golden mahjong symbols that were part of the previous winning combination will be converted to Wild symbols.
          </li>
        </ul>
      </div>
    </div>
  </RuleSection>
</template>

<script setup lang="ts">
import RuleSection from "./RuleSection.vue";
import GoldenSymbolGrid from "../../game/GoldenSymbolGrid.vue";
</script>

<style scoped lang="scss">
.golden-symbol-container {
  display: flex;
  flex-direction: column;
  gap: 32px;
  max-width: 100%;
  margin: 0 auto;
}

.golden-symbol-display-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  padding: 24px;
  background: linear-gradient(135deg, 
    rgba(255, 215, 0, 0.08) 0%,
    rgba(255, 165, 0, 0.04) 100%
  );
  border-radius: 12px;
  border: 2px solid rgba(255, 215, 0, 0.2);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
  transition: all 0.3s ease;

  &:hover {
    border-color: rgba(255, 215, 0, 0.3);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  }
}

.golden-symbol-description {
  margin-top: 8px;
  text-align: left;
}

.golden-rules-list {
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

/* Mobile Responsive Design */
@media (max-width: 768px) {
  .golden-symbol-container {
    gap: 24px;
    padding: 0 8px;
  }

  .golden-symbol-display-wrapper {
    padding: 20px 16px;
  }

  .golden-rules-list {
    gap: 14px;

    li {
      font-size: 14px;
      line-height: 1.5;
    }
  }
}

@media (max-width: 480px) {
  .golden-symbol-container {
    gap: 20px;
    padding: 0 4px;
  }

  .golden-symbol-display-wrapper {
    padding: 16px 12px;
  }

  .golden-rules-list {
    gap: 12px;

    li {
      font-size: 13px;
      line-height: 1.4;
    }
  }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .golden-symbol-display-wrapper {
    transition: none;
  }

  .golden-symbol-display-wrapper:hover {
    transform: none;
  }
}
</style>
