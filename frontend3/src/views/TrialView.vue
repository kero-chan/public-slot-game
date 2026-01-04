<template>
  <div v-if="isLoading" class="trial-loading">
    <div class="loading-content">
      <div class="spinner"></div>
      <p>Starting trial mode...</p>
    </div>
  </div>
  <div v-else-if="error" class="trial-error">
    <div class="error-content">
      <h2>Failed to Start Trial</h2>
      <p>{{ error }}</p>
      <button @click="retryTrial" class="btn-retry">Try Again</button>
      <router-link to="/login" class="link-login">Go to Login</router-link>
    </div>
  </div>
  <GameView v-else />
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores'
import GameView from '@/views/GameView.vue'

const router = useRouter()
const authStore = useAuthStore()

const isLoading = ref(true)
const error = ref<string | null>(null)

async function startTrial() {
  isLoading.value = true
  error.value = null

  try {
    // If already in trial mode, just show the game
    if (authStore.isAuthenticated && authStore.isTrial) {
      isLoading.value = false
      return
    }

    // If logged in as regular user, redirect to game
    if (authStore.isAuthenticated && !authStore.isTrial) {
      router.replace('/game')
      return
    }

    // Start trial session
    await authStore.startTrialMode()
    isLoading.value = false
  } catch (err: any) {
    error.value = err.message || 'Failed to start trial mode'
    isLoading.value = false
  }
}

async function retryTrial() {
  await startTrial()
}

onMounted(() => {
  startTrial()
})
</script>

<style scoped lang="scss">
.trial-loading,
.trial-error {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0a0a15 0%, #1a1a2e 50%, #0f1525 100%);
}

.loading-content,
.error-content {
  text-align: center;
  color: #fff;
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid rgba(255, 255, 255, 0.2);
  border-top-color: rgba(255, 180, 50, 1);
  border-radius: 50%;
  margin: 0 auto 20px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading-content p {
  font-size: 18px;
  color: rgba(255, 255, 255, 0.8);
}

.error-content h2 {
  font-size: 24px;
  margin-bottom: 12px;
  color: rgba(255, 100, 100, 1);
}

.error-content p {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 24px;
}

.btn-retry {
  padding: 12px 32px;
  background: linear-gradient(180deg, rgba(255, 180, 50, 1) 0%, rgba(230, 150, 30, 1) 100%);
  color: #1a1a2e;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  margin-right: 16px;
  transition: transform 0.2s;

  &:hover {
    transform: translateY(-2px);
  }
}

.link-login {
  color: rgba(140, 200, 255, 1);
  text-decoration: none;
  font-weight: 600;

  &:hover {
    text-decoration: underline;
  }
}
</style>
