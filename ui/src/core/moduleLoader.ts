import type { ModuleContext, EkkenModule } from './types/module'
import { useAppSidebarStore } from '../shared/stores/useAppSidebarStore'
import { useAppSettingsStore } from '../shared/stores/useAppSettingsStore'

/**
 * Dynamically discovers and loads all modules from the features directory.
 * Each feature must have an index.ts file that exports an EkkenModule as default.
 */
export const loadModules = async (ctx: ModuleContext) => {
  const sidebarStore = useAppSidebarStore()
  const settingsStore = useAppSettingsStore()

  const moduleFiles = import.meta.glob('../features/*/index.ts', { eager: true })

  const modules: EkkenModule[] = []

  for (const path in moduleFiles) {
    const mod = (moduleFiles[path] as any).default as EkkenModule
    if (mod) {
      modules.push(mod)
    }
  }

  modules.forEach(m => {
    if (m.routes) {
      m.routes.forEach(route => {
        ctx.router.addRoute('main', route)
      })
    }
  })

  modules.forEach(m => {
    if (m.sidebarItems) {
      sidebarStore.addItems(m.sidebarItems)
    }
  })

  modules.forEach(m => {
    if (m.settingsTabs) {
      settingsStore.addSettingsTabs(m.settingsTabs)
    }
  })

  for (const m of modules) {
    if (m.onStartup) {
      try {
        await m.onStartup(ctx)
      } catch (err) {
        console.error(`[ModuleLoader] Failed to start module ${m.id}:`, err)
      }
    }
  }
}
