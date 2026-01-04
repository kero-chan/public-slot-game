<template>
  <div class="auth-container" :style="containerStyle">
    <div class="auth-card">
      <h1>Welcome</h1>
      <h2>Sign In</h2>

      <div v-if="justRegistered" class="success-message">
        Registration successful! Please sign in to continue.
      </div>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label for="username">Username</label>
          <input
            id="username"
            v-model="username"
            type="text"
            required
            placeholder="Enter your username"
            :disabled="authStore.isLoading"
          />
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            placeholder="Enter your password"
            :disabled="authStore.isLoading"
          />
        </div>

        <div v-if="loginError && !showForceLogoutModal" class="error-message">
          {{ loginError }}
        </div>

        <button type="submit" :disabled="authStore.isLoading" class="btn-primary">
          {{ authStore.isLoading ? 'Signing in...' : 'Sign In' }}
        </button>
      </form>

      <div class="divider">
        <span>or</span>
      </div>

      <button
        type="button"
        class="btn-trial"
        @click="handleTrialMode"
        :disabled="authStore.isLoading"
      >
        Play as Guest
      </button>

      <p class="trial-note">
        Trial mode: 100K credits, no registration required
      </p>

      <p class="auth-link">
        Don't have an account?
        <router-link to="/register">Register here</router-link>
      </p>
    </div>

    <!-- Force Logout Confirmation Modal -->
    <div v-if="showForceLogoutModal" class="modal-overlay" @click.self="closeForceLogoutModal">
      <div class="modal-content">
        <div class="modal-icon">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
          </svg>
        </div>
        <h3 class="modal-title">Account Already Logged In</h3>
        <p class="modal-message">
          Your account is currently logged in on another device.<br/>
          Would you like to log out from that device and continue here?
        </p>
        <div class="modal-actions">
          <button
            class="btn-cancel"
            @click="closeForceLogoutModal"
            :disabled="isForceLoggingIn"
          >
            Cancel
          </button>
          <button
            class="btn-confirm"
            @click="handleForceLogin"
            :disabled="isForceLoggingIn"
          >
            {{ isForceLoggingIn ? 'Signing in...' : 'Confirm' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores'
import { ASSETS } from '@/config/assets'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const loginError = ref<string | null>(null)
const showForceLogoutModal = ref(false)
const isForceLoggingIn = ref(false)

// Check if user just registered
const justRegistered = computed(() => route.query.registered === 'true')

// Dynamic background style from API assets
const containerStyle = computed(() => {
  const bgUrl = ASSETS.imagePaths.backgroundStart
  if (bgUrl) {
    return {
      backgroundImage: `url(${bgUrl})`,
      backgroundSize: 'cover',
      backgroundPosition: 'center',
      backgroundRepeat: 'no-repeat'
    }
  }
  return {}
})

async function handleLogin() {
  loginError.value = null

  try {
    await authStore.login({
      username: username.value,
      password: password.value,
    })

    // Wait for Vue reactivity to propagate state changes before navigating
    await nextTick()

    // Get redirect path from query params, or default to game
    const redirectPath = (route.query.redirect as string) || '/game'

    // Use replace to prevent back button from returning to login page
    await router.replace(redirectPath)
  } catch (error: any) {
    // Check if it's "already logged in" error
    // Note: API client interceptor transforms error to { status, error, message, details }
    if (error.error === 'already_logged_in') {
      showForceLogoutModal.value = true
    } else {
      loginError.value = error.message || 'Login failed'
    }
  }
}

function closeForceLogoutModal() {
  showForceLogoutModal.value = false
}

async function handleForceLogin() {
  isForceLoggingIn.value = true
  loginError.value = null

  try {
    await authStore.login({
      username: username.value,
      password: password.value,
      force_logout: true,
    })

    showForceLogoutModal.value = false

    // Wait for Vue reactivity to propagate state changes before navigating
    await nextTick()

    // Get redirect path from query params, or default to game
    const redirectPath = (route.query.redirect as string) || '/game'

    // Use replace to prevent back button from returning to login page
    await router.replace(redirectPath)
  } catch (error: any) {
    showForceLogoutModal.value = false
    loginError.value = error.message || 'Login failed'
  } finally {
    isForceLoggingIn.value = false
  }
}

function handleTrialMode() {
  // Redirect to /trial which handles trial session setup
  router.replace('/trial')
}
</script>

<style scoped lang="scss">
@use "@/assets/styles/views/LoginView.scss";
</style>
