<template>
  <RuleSection title="1,024 Ways">
    <div class="road1024-container">
      <div class="road1024-title-section">
        <div class="road1024-title">1,024 Ways</div>

        <div class="road1024-description">
          <ul class="road1024-rules-list">
            <li>Winning ways pay from left to right on the reels.</li>
          </ul>
        </div>

        <div class="ways-symbols-display">
          <ReelGrid :grid-data="winningGridData" />
          <div class="tick-separator">
            <div class="paytable-icon tick"></div>
          </div>
          <ReelGrid :grid-data="nonWinningGridData" />
          <div class="tick-separator">
            <div class="paytable-icon cross"></div>
          </div>
        </div>

        <div class="road1024-description">
          <ul class="road1024-rules-list">
            <li>
              Winning ways are calculated by multiplying the number of matching
              symbols from left to right on adjacent reels by the number of ways
              for that symbol.
            </li>
          </ul>
        </div>

        <div class="example-calculation">
          <div class="example-text">Referring to the example above:</div>
          <div class="calculation-result">1 x 3 x 2 = 6</div>
        </div>

        <div class="road1024-description">
          <ul class="road1024-rules-list">
            <li>
              A winning symbol's payout is the symbol's payout value multiplied
              by the number of winning ways.
            </li>
          </ul>
        </div>

        <div class="payout-example">
          <div class="symbol-payout-table">
            <div class="symbol-info">
              <div class="symbol-icon">
                <div class="symbol-wrapper">
                  <div class="paytable-icon questionmark">
                    <img
                      src="@/assets/images/japaneseOmakase/menu/symbol_icon.webp"
                      alt="Symbol"
                    />
                  </div>
                </div>
              </div>
              <PayoutTable :payout-data="payoutData" />
            </div>
          </div>
          <div class="total-calculation">
            <div class="total-text">Total win in this example:</div>
            <div class="total-result">10 x 6 = 60</div>
          </div>
        </div>

        <div class="road1024-description">
          <ul class="road1024-rules-list">
            <li>
              After each round ends and payouts are awarded, all winning symbols
              explode and disappear, allowing symbols above to drop down for a
              new round.
            </li>
            <li>
              Additional winning combinations are added to each spin's total
              until no more winning combinations occur.
            </li>
            <li>All wins are displayed in cash.</li>
          </ul>
        </div>
      </div>
    </div>
  </RuleSection>
</template>

<script setup lang="ts">
import { computed } from "vue";
import RuleSection from "./RuleSection.vue";
import ReelGrid from "./ReelGrid.vue";
import PayoutTable from "./PayoutTable.vue";

// Winning grid configuration (1 x 3 x 2)
const winningGridData = computed(() => [
  [true, false, false, false], // Reel 1: 1 symbol
  [true, true, false, true], // Reel 2: 3 symbols
  [false, true, false, true], // Reel 3: 2 symbols
  [false, false, false, false], // Reel 4: empty
  [false, false, false, false], // Reel 5: empty
]);

// Non-winning grid (disconnected symbols)
const nonWinningGridData = computed(() => [
  [true, false, false, false], // Reel 1: 1 symbol
  [true, true, false, true], // Reel 2: 3 symbols
  [false, false, false, false], // Reel 3: empty (breaks the chain)
  [false, true, false, true], // Reel 4: 2 symbols (but no connection)
  [false, false, false, false], // Reel 5: empty
]);

// Payout table data
const payoutData = [
  { count: 5, value: 500 },
  { count: 4, value: 100 },
  { count: 3, value: 10, highlighted: true },
];
</script>

<style scoped lang="scss">
.road1024-container {
  width: 100%;
}

.road1024-title-section {
  display: inline-block;
  width: 100%;
  margin-bottom: 60px;
  text-align: center;
}

.road1024-title {
  font-size: 14px;
  color: rgb(255, 255, 255);
  font-weight: normal;
  margin-bottom: 16px;
  line-height: 145%;
  padding: 0 20px;
  direction: ltr;
}

.road1024-description {
  margin-top: 8px;
  text-align: left;
}

.road1024-rules-list {
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

.ways-symbols-display {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 4px;
  margin: 20px 0;
}

.tick-separator {
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 5px;
}

.paytable-icon {
  display: block;
  position: relative;

  &.tick {
    transform: scale(0.5);
    right: 12px;
    bottom: 9px;
    min-width: 48px;
    width: 48px;
    min-height: 36px;
    height: 36px;

    &::before {
      content: "✓";
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%) scale(2);
      font-size: 2.5rem;
      color: #4caf50;
      font-weight: bold;
      padding-left: 20px;
    }
  }

  &.cross {
    transform: scale(0.5);
    right: 10px;
    bottom: 10px;
    min-width: 40px;
    width: 40px;
    min-height: 40px;
    height: 40px;

    &::before {
      content: "✗";
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%) scale(2);
      font-size: 2.5rem;
      color: #f44336;
      font-weight: bold;
      padding-left: 20px;
    }
  }

  &.questionmark {
    img {
      transform: scale(0.5);
      right: 30px;
      bottom: 30px;
      min-width: 80px;
      width: 80px;
      min-height: 80px;
      height: 80px;
    }
  }
}

.example-calculation {
  color: rgb(255, 255, 255);
  font-size: 16px;
  width: fit-content;
  padding-top: 0;
  padding-bottom: 0;
  margin: 20px auto;
  text-align: center;

  .example-text {
    margin-bottom: 8px;
  }

  .calculation-result {
    color: rgb(180, 120, 80);
    font-weight: bold;
  }
}

.payout-example {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 35px;
  margin: 30px 0;

  @media (max-width: 768px) {
    flex-direction: column;
    gap: 20px;
  }
}

.symbol-payout-table {
  color: rgb(255, 255, 255);
  font-size: 14px;
  padding: 15px;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.symbol-info {
  display: flex;
  align-items: center;
  gap: 6px;
}

.symbol-icon {
  width: 75px;
  text-align: center;

  .symbol-wrapper {
    min-width: 60px;
    width: 60px;
    min-height: 60px;
    height: 60px;
    position: relative;
    display: flex;
    justify-content: center;
    align-items: center;
    margin: 0 auto;
  }
}

.total-calculation {
  color: rgb(255, 255, 255);
  font-size: 14px;
  width: 160px;
  text-align: center;
  padding: 6px;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);

  .total-text {
    margin-bottom: 8px;
  }

  .total-result {
    color: rgb(180, 120, 80);
    font-weight: bold;
  }
}

/* Responsive Design */
@media (max-width: 768px) {
  .road1024-title {
    font-size: 13px;
    padding: 0 15px;
  }

  .ways-symbols-display {
    flex-wrap: wrap;
    gap: 32px;
    justify-content: center;
  }

  .paytable-icon {
    &.tick {
      transform: scale(0.4);

      &::before {
        font-size: 2rem;
      }
    }

    &.cross {
      transform: scale(0.4);

      &::before {
        font-size: 2rem;
      }
    }
  }

  .road1024-description {
    font-size: 11px;
    margin: 15px 15px 15px 25px;
  }

  .road1024-rules-list {
    gap: 12px;

    li {
      font-size: 13px;
      line-height: 1.4;
    }
  }

  .example-calculation,
  .symbol-payout-table,
  .total-calculation {
    font-size: 11px;
  }

  .total-calculation {
    width: 110px;
  }
}

/* Animations */
@keyframes symbol-glow {
  0% {
    text-shadow: 0 0 5px currentColor;
  }
  100% {
    text-shadow: 0 0 15px currentColor, 0 0 25px currentColor;
  }
}

.paytable-icon::before {
  animation: symbol-glow 2s ease-in-out infinite alternate;
}
</style>
