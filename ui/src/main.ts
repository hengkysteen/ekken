import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'

import App from './App.vue'
import router from './router'
import { loadModules } from './core/moduleLoader'
import './assets/main.css'

async function bootstrap() {
  const app = createApp(App)

  const pinia = createPinia()
  app.use(pinia)

  // Load dynamic modules BEFORE mounting and using router
  await loadModules({ app, router })

  // Now attach the router so it triggers initial navigation with all routes present
  app.use(router)

  // Dynamic modules handle their own initialization in onStartup

  app.mount('#app')
}

bootstrap()
