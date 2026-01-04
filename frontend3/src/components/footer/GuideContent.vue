<template>
  <div class="guide-content">
    <div class="rules-segment">
      <div class="rules-header">
        <div class="rules-title">Omakase Fortune</div>
        <div class="title-decoration"></div>
      </div>

      <div class="rules-container">
        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              Omakase Fortune is a 5-reel, 4-row video slot game featuring transforming Wild symbols and Free Spins with increasing multipliers.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              The game features 1,024 ways to win (fixed) with a base bet of 20, bet multipliers from 1 to 10, and bet sizes ranging from 0.03 to 0.90.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">Bet size can be set in the "Bet Size" option.</div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">Bet multiplier can be set in the "Bet Multiplier" option.</div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">Total bet can be set in the "Total Bet" option.</div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              "Cash Wallet" displays the available cash balance for the player.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              "Auto Spin" automatically plays the game for the number of rounds selected by the player.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              Winning combinations and payouts are awarded according to the "Paytable".
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              A winning way pays an amount equal to the symbol payout value shown in the "Paytable" multiplied by the bet size and bet multiplier.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              Winning ways pay from left to right on the reels.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              Winning ways are calculated by multiplying the number of matching symbols from left to right on adjacent reels by the number of ways for that symbol.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              A winning symbol's payout is the symbol's payout value multiplied by the number of winning ways.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">Wins on different ways are added together.</div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              After each round ends and payouts are awarded, all winning symbols explode and disappear, allowing symbols above to drop down for a new round.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              Additional winning combinations are added to each spin's total until no more winning combinations occur.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">All wins are displayed in cash.</div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              Wild symbols substitute for all symbols except Scatter symbols.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              Golden Mahjong symbols only appear on reels 2, 3, and 4.
            </div>
          </div>
        </div>

        <div class="rules-text-container">
          <div class="rule-card">
            <div class="rule-icon">💎</div>
            <div class="rule-content">
              Bet size and bet multiplier in Free Spin Mode remain the same as when Free Spin Mode was triggered.
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts"></script>

<style scoped lang="scss">
.guide-content {
  width: 100%;
  height: 100%;
}

.rules-segment {
  position: relative;
  z-index: 1;
  min-height: fit-content;
}

.rules-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 24px;
  position: relative;
}

.mahjong-icon {
  font-size: 32px;
  margin-bottom: 8px;
  animation: float 3s ease-in-out infinite;
}

.rules-title {
  font-size: 20px;
  font-weight: 700;
  color: #ffffff;
  text-align: center;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
  margin-bottom: 12px;
  letter-spacing: 2px;
}

.title-decoration {
  width: 80px;
  height: 3px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    #ffd700 50%,
    transparent 100%
  );
  border-radius: 2px;
  position: relative;

  &::after {
    content: "";
    position: absolute;
    top: -2px;
    left: 50%;
    transform: translateX(-50%);
    width: 8px;
    height: 8px;
    background: #ffd700;
    border-radius: 50%;
    box-shadow: 0 0 8px rgba(255, 215, 0, 0.6);
  }
}

.rules-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rules-text-container {
  opacity: 0;
  animation: slideInUp 0.6s ease forwards;

  &:nth-child(1) {
    animation-delay: 0.1s;
  }
  &:nth-child(2) {
    animation-delay: 0.15s;
  }
  &:nth-child(3) {
    animation-delay: 0.2s;
  }
  &:nth-child(4) {
    animation-delay: 0.25s;
  }
  &:nth-child(5) {
    animation-delay: 0.3s;
  }
  &:nth-child(6) {
    animation-delay: 0.35s;
  }
  &:nth-child(7) {
    animation-delay: 0.4s;
  }
  &:nth-child(8) {
    animation-delay: 0.45s;
  }
  &:nth-child(9) {
    animation-delay: 0.5s;
  }
  &:nth-child(10) {
    animation-delay: 0.55s;
  }
  &:nth-child(11) {
    animation-delay: 0.6s;
  }
  &:nth-child(12) {
    animation-delay: 0.65s;
  }
  &:nth-child(13) {
    animation-delay: 0.7s;
  }
  &:nth-child(14) {
    animation-delay: 0.75s;
  }
  &:nth-child(15) {
    animation-delay: 0.8s;
  }
  &:nth-child(16) {
    animation-delay: 0.85s;
  }
  &:nth-child(17) {
    animation-delay: 0.9s;
  }
  &:nth-child(18) {
    animation-delay: 0.95s;
  }
}

.rule-card {
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.08) 0%,
    rgba(255, 255, 255, 0.02) 100%
  );
  border: 1px solid rgba(255, 215, 0, 0.2);
  border-radius: 12px;
  padding: 16px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;

  &::before {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      135deg,
      rgba(255, 215, 0, 0.05) 0%,
      transparent 100%
    );
    opacity: 0;
    transition: opacity 0.3s ease;
  }

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 215, 0, 0.4);
    border-color: rgba(255, 215, 0, 0.5);

    &::before {
      opacity: 1;
    }

    .rule-icon {
      transform: scale(1.1) rotate(5deg);
      filter: drop-shadow(0 0 8px rgba(255, 215, 0, 0.6));
    }

    .rule-content {
      color: #ffffff;
    }
  }

  &:active {
    transform: translateY(0);
    transition-duration: 0.1s;
  }
}

.rule-icon {
  font-size: 20px;
  min-width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin-top: 2px;
}

.rule-content {
  flex: 1;
  color: #e0e0e0;
  font-size: 13px;
  line-height: 1.6;
  transition: color 0.3s ease;
  word-wrap: break-word;
  hyphens: auto;
}

/* Animations */
@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes float {
  0%,
  100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-8px);
  }
}

@keyframes glow {
  0%,
  100% {
    box-shadow: 0 0 5px rgba(255, 215, 0, 0.3);
  }
  50% {
    box-shadow: 0 0 20px rgba(255, 215, 0, 0.6);
  }
}

/* Responsive Design */
@media (max-width: 768px) {
  .mahjong-icon {
    font-size: 28px;
  }

  .rules-title {
    font-size: 18px;
    letter-spacing: 1.5px;
  }

  .title-decoration {
    width: 60px;
    height: 2px;
  }

  .rules-container {
    gap: 10px;
  }

  .rule-card {
    padding: 12px;
    gap: 10px;
    border-radius: 10px;
  }

  .rule-icon {
    font-size: 18px;
    min-width: 18px;
    height: 18px;
  }

  .rule-content {
    font-size: 12px;
    line-height: 1.5;
  }
}

@media (max-width: 480px) {
  .mahjong-icon {
    font-size: 24px;
    margin-bottom: 6px;
  }

  .rules-title {
    font-size: 16px;
    letter-spacing: 1px;
    margin-bottom: 8px;
  }

  .title-decoration {
    width: 50px;

    &::after {
      width: 6px;
      height: 6px;
    }
  }

  .rules-container {
    gap: 8px;
  }

  .rule-card {
    padding: 10px;
    gap: 8px;
    border-radius: 8px;

    &:hover {
      transform: translateY(-1px);
    }
  }

  .rule-icon {
    font-size: 16px;
    min-width: 16px;
    height: 16px;
  }

  .rule-content {
    font-size: 11px;
    line-height: 1.4;
  }
}

.rule-card {
  &:focus-within {
    outline: 2px solid rgba(255, 215, 0, 0.6);
    outline-offset: 2px;
  }
}

/* Dark mode enhancements */
@media (prefers-color-scheme: dark) {
  .rule-card {
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.06) 0%,
      rgba(255, 255, 255, 0.01) 100%
    );
    border-color: rgba(255, 215, 0, 0.25);

    &:hover {
      border-color: rgba(255, 215, 0, 0.6);
    }
  }
}

/* Print styles */
@media print {
  .rule-card {
    background: #f8f9fa !important;
    border: 1px solid #dee2e6 !important;
    break-inside: avoid;
  }

  .rule-content {
    color: black !important;
  }
}
</style>
