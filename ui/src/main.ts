import { createApp, defineComponent, h, reactive, watch } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import * as ElementPlusComponents from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'

import App from './App.vue'
import router from './router'
import { loadModules } from './core/moduleLoader'
import './assets/main.css'

declare global {
  interface Window {
    EkkenVue?: {
      defineComponent: typeof defineComponent
      h: typeof h
      reactive: typeof reactive
      watch: typeof watch
    }
    EkkenUI?: any
  }
}

async function bootstrap() {
  const app = createApp(App)

  app.use(ElementPlus)

  // Register all icons
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
  }

  const pinia = createPinia()
  app.use(pinia)

  // Load dynamic modules BEFORE mounting and using router
  await loadModules({ app, router })

  // Now attach the router so it triggers initial navigation with all routes present
  app.use(router)

  window.EkkenVue = {
    defineComponent,
    h,
    reactive,
    watch,
  }

  // Map common components for backward compatibility (best effort)
  window.EkkenUI = {
    // We expose actual ElementPlus component objects
    ...ElementPlusComponents,

    // Nxxx aliases for plugins that expect them (mapped to actual component objects)
    NAlert: ElementPlusComponents.ElAlert,
    NCollapse: ElementPlusComponents.ElCollapse,
    NCollapseItem: ElementPlusComponents.ElCollapseItem,
    NForm: ElementPlusComponents.ElForm,
    NFormItem: ElementPlusComponents.ElFormItem,
    NInput: ElementPlusComponents.ElInput,
    NInputNumber: ElementPlusComponents.ElInputNumber,
    NSelect: ElementPlusComponents.ElSelect,
    NSwitch: ElementPlusComponents.ElSwitch,
  }

  // Dynamic modules handle their own initialization in onStartup

  app.mount('#app')
}

bootstrap()
