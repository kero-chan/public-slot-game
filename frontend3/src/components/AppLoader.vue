<template>
  <div class="app-loader">
    <!-- Animated background particles -->
    <div class="particles">
      <div v-for="n in 20" :key="n" class="particle" :style="getParticleStyle(n)"></div>
    </div>

    <!-- Decorative frame -->
    <div class="loader-frame">
      <div class="frame-corner top-left"></div>
      <div class="frame-corner top-right"></div>
      <div class="frame-corner bottom-left"></div>
      <div class="frame-corner bottom-right"></div>

      <div class="loader-content">
        <!-- Game logo/title -->
        <div class="logo-container">
          <div class="logo-glow"></div>
          <h1 class="game-title">{{ gameName }}</h1>
          <div class="title-underline"></div>
        </div>

        <!-- Loading spinner -->
        <div v-if="status !== 'error'" class="spinner-container">
          <div class="spinner-outer">
            <div class="spinner-inner"></div>
          </div>
          <div class="spinner-center">
            <span class="progress-number">{{ Math.floor(progress) }}</span>
            <span class="progress-percent">%</span>
          </div>
        </div>

        <!-- Progress bar -->
        <div class="loading-container">
          <div class="progress-bar-bg">
            <div class="progress-bar-fill" :style="{ width: `${progress}%` }">
              <div class="progress-shine"></div>
            </div>
          </div>
          <p class="loading-text">
            <span class="loading-dots" v-if="status === 'loading'">
              {{ statusText }}<span class="dots">...</span>
            </span>
            <span v-else>{{ statusText }}</span>
          </p>
        </div>

        <!-- Error message -->
        <div v-if="error" class="error-message">
          <div class="error-icon">!</div>
          <p>{{ error }}</p>
          <button class="retry-btn" @click="$emit('retry')">Retry</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ASSETS } from '@/config/assets'

const props = defineProps<{
  progress: number
  status: 'loading' | 'error' | 'complete'
  error?: string | null
}>()

// Get game name from API assets (reactive)
const gameName = computed(() => ASSETS.gameName || 'Loading...')

defineEmits<{
  retry: []
}>()

const statusText = computed(() => {
  if (props.status === 'error') return 'Failed to load'
  if (props.progress < 30) return 'Connecting'
  if (props.progress < 60) return 'Loading assets'
  if (props.progress < 90) return 'Preparing game'
  if (props.progress < 100) return 'Almost ready'
  return 'Ready'
})

// Generate random styles for floating particles
function getParticleStyle(index: number) {
  const size = 4 + (index % 3) * 4
  const left = (index * 5) % 100
  const delay = (index * 0.3) % 5
  const duration = 10 + (index % 5) * 3
  return {
    width: `${size}px`,
    height: `${size}px`,
    left: `${left}%`,
    animationDelay: `${delay}s`,
    animationDuration: `${duration}s`
  }
}
</script>

<style scoped lang="scss">
.app-loader {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(ellipse at center, #1e3a5f 0%, #0d1b2a 50%, #0a0f1a 100%);
  z-index: 99999;
  overflow: hidden;
}

// Floating particles
.particles {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}

.particle {
  position: absolute;
  bottom: -20px;
  background: radial-gradient(circle, rgba(255, 215, 0, 0.6) 0%, transparent 70%);
  border-radius: 50%;
  animation: float-up linear infinite;
  opacity: 0.5;
}

@keyframes float-up {
  0% {
    transform: translateY(0) rotate(0deg);
    opacity: 0;
  }
  10% {
    opacity: 0.6;
  }
  90% {
    opacity: 0.3;
  }
  100% {
    transform: translateY(-100vh) rotate(720deg);
    opacity: 0;
  }
}

// Frame styling
.loader-frame {
  position: relative;
  padding: clamp(30px, 6vw, 60px);
  background: linear-gradient(180deg, rgba(30, 58, 95, 0.8) 0%, rgba(13, 27, 42, 0.9) 100%);
  border: 2px solid rgba(255, 215, 0, 0.3);
  border-radius: 20px;
  box-shadow:
    0 0 60px rgba(255, 180, 0, 0.15),
    inset 0 0 30px rgba(0, 0, 0, 0.3);
}

.frame-corner {
  position: absolute;
  width: 30px;
  height: 30px;
  border: 3px solid #ffd700;

  &.top-left {
    top: -3px;
    left: -3px;
    border-right: none;
    border-bottom: none;
    border-radius: 8px 0 0 0;
  }

  &.top-right {
    top: -3px;
    right: -3px;
    border-left: none;
    border-bottom: none;
    border-radius: 0 8px 0 0;
  }

  &.bottom-left {
    bottom: -3px;
    left: -3px;
    border-right: none;
    border-top: none;
    border-radius: 0 0 0 8px;
  }

  &.bottom-right {
    bottom: -3px;
    right: -3px;
    border-left: none;
    border-top: none;
    border-radius: 0 0 8px 0;
  }
}

.loader-content {
  text-align: center;
  min-width: clamp(280px, 50vw, 400px);
}

// Logo styling
.logo-container {
  position: relative;
  margin-bottom: 40px;
}

.logo-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 200%;
  height: 100px;
  background: radial-gradient(ellipse, rgba(255, 215, 0, 0.2) 0%, transparent 70%);
  pointer-events: none;
}

.game-title {
  position: relative;
  color: #ffd700;
  font-size: clamp(28px, 7vw, 42px);
  font-weight: 700;
  margin: 0;
  text-shadow:
    0 0 10px rgba(255, 215, 0, 0.8),
    0 0 20px rgba(255, 180, 0, 0.5),
    0 0 40px rgba(255, 150, 0, 0.3),
    0 4px 8px rgba(0, 0, 0, 0.5);
  letter-spacing: 3px;
  animation: title-pulse 2s ease-in-out infinite;
}

@keyframes title-pulse {
  0%, 100% {
    text-shadow:
      0 0 10px rgba(255, 215, 0, 0.8),
      0 0 20px rgba(255, 180, 0, 0.5),
      0 0 40px rgba(255, 150, 0, 0.3),
      0 4px 8px rgba(0, 0, 0, 0.5);
  }
  50% {
    text-shadow:
      0 0 20px rgba(255, 215, 0, 1),
      0 0 40px rgba(255, 180, 0, 0.7),
      0 0 60px rgba(255, 150, 0, 0.5),
      0 4px 8px rgba(0, 0, 0, 0.5);
  }
}

.title-underline {
  margin: 12px auto 0;
  width: 60%;
  height: 2px;
  background: linear-gradient(90deg, transparent, #ffd700, transparent);
}

// Spinner
.spinner-container {
  position: relative;
  width: 100px;
  height: 100px;
  margin: 0 auto 30px;
}

.spinner-outer {
  position: absolute;
  inset: 0;
  border: 3px solid rgba(255, 215, 0, 0.1);
  border-top-color: #ffd700;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.spinner-inner {
  position: absolute;
  inset: 10px;
  border: 3px solid rgba(255, 180, 0, 0.1);
  border-bottom-color: #ffaa00;
  border-radius: 50%;
  animation: spin 0.8s linear infinite reverse;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.spinner-center {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
}

.progress-number {
  color: #ffd700;
  font-size: 24px;
  font-weight: 700;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.5);
}

.progress-percent {
  color: #ffb870;
  font-size: 14px;
  font-weight: 500;
}

// Progress bar
.loading-container {
  width: 100%;
}

.progress-bar-bg {
  height: 10px;
  background: rgba(0, 0, 0, 0.4);
  border-radius: 5px;
  overflow: hidden;
  border: 1px solid rgba(255, 215, 0, 0.2);
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.3);
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #ff8c00, #ffd700, #ffea00);
  border-radius: 5px;
  transition: width 0.3s ease;
  position: relative;
  overflow: hidden;
}

.progress-shine {
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4), transparent);
  animation: shine 2s ease-in-out infinite;
}

@keyframes shine {
  0% {
    left: -100%;
  }
  50%, 100% {
    left: 100%;
  }
}

.loading-text {
  margin-top: 16px;
  color: #ffb870;
  font-size: 14px;
  font-weight: 500;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.5);
  letter-spacing: 1px;
}

.loading-dots .dots {
  display: inline-block;
  animation: dots 1.5s steps(4, end) infinite;
  width: 20px;
  text-align: left;
}

@keyframes dots {
  0% { content: ''; }
  25% { content: '.'; }
  50% { content: '..'; }
  75% { content: '...'; }
  100% { content: ''; }
}

// Error styling
.error-message {
  margin-top: 24px;
  padding: 20px;
  background: rgba(180, 40, 40, 0.2);
  border: 1px solid rgba(255, 100, 100, 0.4);
  border-radius: 12px;
}

.error-icon {
  width: 40px;
  height: 40px;
  margin: 0 auto 12px;
  background: linear-gradient(135deg, #ff6b6b, #c0392b);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 24px;
  font-weight: 700;
  box-shadow: 0 4px 12px rgba(192, 57, 43, 0.4);
}

.error-message p {
  margin: 0 0 16px;
  color: #ffaaaa;
  font-size: 14px;
}

.retry-btn {
  padding: 10px 30px;
  background: linear-gradient(135deg, #c0392b, #e74c3c);
  border: none;
  border-radius: 20px;
  color: white;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 12px rgba(192, 57, 43, 0.3);

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(192, 57, 43, 0.4);
  }

  &:active {
    transform: translateY(0);
  }
}
</style>
