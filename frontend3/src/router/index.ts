import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores'
import LoginView from '@/views/LoginView.vue'
import RegisterView from '@/views/RegisterView.vue'
import GameView from '@/views/GameView.vue'
import TrialView from '@/views/TrialView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/game',
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { requiresAuth: false },
    },
    {
      path: '/register',
      name: 'register',
      component: RegisterView,
      meta: { requiresAuth: false },
    },
    {
      path: '/game',
      name: 'game',
      component: GameView,
      meta: { requiresAuth: true },
    },
    {
      path: '/trial',
      name: 'trial',
      component: TrialView,
      meta: { requiresAuth: false, isTrial: true },
    },
  ],
})

// Navigation guard for authentication
router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.meta.requiresAuth

  // If route requires authentication
  if (requiresAuth) {
    if (!authStore.isAuthenticated) {
      // Not authenticated, redirect to login
      next({ name: 'login', query: { redirect: to.fullPath } })
      return
    }

    // Authenticated but no player data - try to fetch profile
    if (authStore.sessionToken && !authStore.player) {
      try {
        await authStore.fetchProfile()
      } catch (error) {
        // Failed to fetch profile (token likely expired)
        // Clear auth state and redirect to login
        authStore.logout()
        next({ name: 'login', query: { redirect: to.fullPath } })
        return
      }
    }
  } else {
    // Route doesn't require auth, but redirect to game if already logged in
    if (authStore.isAuthenticated && (to.name === 'login' || to.name === 'register')) {
      next({ name: 'game' })
      return
    }
  }

  next()
})

export default router
