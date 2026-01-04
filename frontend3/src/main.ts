import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import { gsap } from 'gsap'
import App from './App.vue'
import router from './router'

// Configure GSAP for smoother animations
// Disable lag smoothing to prevent GSAP from "catching up" on dropped frames
gsap.ticker.lagSmoothing(0)

// Create Vue application
const app = createApp(App)

// Register Pinia for state management
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)
app.use(pinia)

// Register Vue Router
app.use(router)

// Mount application
app.mount('#app')
