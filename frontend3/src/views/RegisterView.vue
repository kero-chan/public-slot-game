<template>
  <div class="auth-container" :style="containerStyle">
    <div class="auth-card">
      <h1>Welcome</h1>
      <h2>Create Account</h2>

      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label for="username">Username</label>
          <input
            id="username"
            v-model="username"
            type="text"
            required
            minlength="3"
            maxlength="50"
            placeholder="Choose a username"
            :disabled="authStore.isLoading"
          />
        </div>

        <div class="form-group">
          <label for="email">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            placeholder="Enter your email"
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
            minlength="8"
            placeholder="Choose a password (min. 8 characters)"
            :disabled="authStore.isLoading"
          />
        </div>

        <div class="form-group">
          <label for="confirmPassword">Confirm Password</label>
          <input
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            required
            placeholder="Re-enter your password"
            :disabled="authStore.isLoading"
          />
        </div>

        <div v-if="validationError" class="error-message">
          {{ validationError }}
        </div>

        <div v-if="authStore.error" class="error-message">
          {{ authStore.error }}
        </div>

        <button type="submit" :disabled="authStore.isLoading" class="btn-primary">
          {{ authStore.isLoading ? 'Creating account...' : 'Register' }}
        </button>
      </form>

      <p class="auth-link">
        Already have an account?
        <router-link to="/login">Sign in here</router-link>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores'
import { ASSETS } from '@/config/assets'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')

const validationError = computed(() => {
  if (password.value && confirmPassword.value && password.value !== confirmPassword.value) {
    return 'Passwords do not match'
  }
  return null
})

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

async function handleRegister() {
  if (validationError.value) {
    return
  }

  try {
    await authStore.register({
      username: username.value,
      email: email.value,
      password: password.value,
    })

    // Registration successful - redirect to login page
    // User must login to get session token
    // Use replace to prevent back button from returning to completed form
    await router.replace({ path: '/login', query: { registered: 'true' } })
  } catch (error) {
    // Error is handled by the store
    console.error('Registration failed:', error)
  }
}
</script>

<style scoped lang="scss">
@use "@/assets/styles/views/RegisterView.scss";
</style>
