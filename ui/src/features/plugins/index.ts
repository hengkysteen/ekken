import { markRaw } from 'vue'
import { Connection } from '@element-plus/icons-vue'
import type { EkkenModule } from '../../core/types/module'
import { loadPluginNodeConfigs, getPluginComponent } from './logic/registry'
import { registerNodeResolver } from "@workflows/node/nodes/registry";


const pluginsModule: EkkenModule = {
  id: 'plugins',
  name: 'Plugins',

  routes: [
    {
      path: 'plugins',
      name: 'plugins',
      component: () => import('./views/PluginsView.vue'),
      meta: { breadcrumb: 'Plugins' }
    },
    {
      path: 'plugins/hub',
      name: 'plugins-hub',
      component: () => import('./views/HubView.vue'),
      meta: { breadcrumb: 'Hub', parent: '/plugins' }
    }
  ],

  sidebarItems: [
    {
      label: 'Plugins',
      icon: markRaw(Connection),
      path: '/plugins',
      name: 'plugins',
      order: 100 // High order to put it in "System" section-like area if we want
    }
  ],

  onStartup: async () => {
    console.log('[PluginsModule] Initializing plugin node configs...')
    try {
      await loadPluginNodeConfigs()
      // Inject our component resolver into the workflow registry
      // Convert null to undefined for type compatibility
      registerNodeResolver((type) => getPluginComponent(type) ?? undefined)
    } catch (err) {
      console.error('[PluginsModule] Failed to load plugin node configs:', err)
    }
  }
}

export default pluginsModule
