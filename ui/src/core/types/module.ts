import type { App } from 'vue'
import type { Router, RouteRecordRaw } from 'vue-router'

export interface ModuleContext {
  app: App
  router: Router
}

export interface SidebarItem {
  label: string
  icon: any
  path: string
  name: string
  order?: number
}

export interface SettingsTab {
  id: string
  label: string
  description?: string
  icon: any
  component: any
  order?: number
}

export interface EkkenModule {
  id: string
  name: string
  
  // Routes to be injected into the main router
  routes?: RouteRecordRaw[]
  
  // Navigation items for the sidebar
  sidebarItems?: SidebarItem[]

  // Settings tabs to be injected into the settings page
  settingsTabs?: SettingsTab[]
  
  // Execution logic when the app boots up
  onStartup?: (ctx: ModuleContext) => Promise<void>
}
