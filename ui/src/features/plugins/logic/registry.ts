import type { Component } from 'vue'
import { pluginsApi as api } from '@plugins/api'
import type { PluginSummary } from '@plugins/api'
import type {
  PluginNodeConfigModule,
  NodeConfigRegistrar,
  PluginNodeConfigAPI,
} from '@plugins/types/plugin'
import type { NodeFormProps } from '@workflows/node/types/node'

/**
 * @deprecated Use NodeConfigRegistrar from '../types/plugin' instead
 */
export type PluginNodeConfigRegistrar = NodeConfigRegistrar

export type { PluginNodeConfigModule, PluginNodeConfigAPI }

const registry: Record<string, Component> = {}

/**
 * Register a node config component for a specific node type.
 * This is called by plugin UI modules during registration.
 * 
 * @param nodeType - The node type identifier
 * @param component - Vue component implementing NodeConfigComponent interface
 */
export function registerNodeConfig(nodeType: string, component: Component<NodeFormProps>) {
  console.log(`[Plugin Registry] Registering node config for type: ${nodeType}`)
  registry[nodeType] = component
}

/**
 * Unregister a node config component.
 * Used during plugin reload or cleanup.
 * 
 * @param nodeType - The node type identifier to unregister
 */
export function unregisterNodeConfig(nodeType: string) {
  delete registry[nodeType]
}

/**
 * Get the registered config component for a node type.
 * Returns null if no custom component is registered.
 *
 * NOTE: There is NO generic fallback. If this returns null, the UI will throw
 * an error. Every node (internal or plugin) MUST have a configuration component.
 *
 * @param type - The node type identifier
 * @returns The registered component or null
 */
export function getPluginComponent(type: string): Component | null {
  return registry[type] || null
}

/**
 * Load all plugin node config UI modules.
 * Called during app bootstrap to register custom components from plugins.
 */
export async function loadPluginNodeConfigs() {
  const plugins = await api.getPlugins()
  // Relaxing the filter to include any plugin with UI that has node types
  const candidates = plugins.filter((plugin) => 
    plugin.status === 'loaded' && 
    plugin.has_ui && 
    plugin.ui_module_url &&
    (plugin.node_types && plugin.node_types.length > 0)
  )

  for (const plugin of candidates) {
    await loadPluginNodeConfig(plugin)
  }
}

/**
 * Load a single plugin's node config UI module.
 * 
 * @param plugin - Plugin summary with UI module information
 * @throws Error if the plugin module doesn't export a valid register function
 */
async function loadPluginNodeConfig(plugin: PluginSummary) {
  const moduleURL = `${plugin.ui_module_url}?t=${Date.now()}`
  const loaded = await import(/* @vite-ignore */ moduleURL) as PluginNodeConfigModule
  const register = loaded.register || loaded.default

  if (typeof register !== 'function') {
    throw new Error(`Plugin UI module '${plugin.id}' must export a register function`)
  }

  const api: PluginNodeConfigAPI = {
    registerNodeConfig
  }

  register(api)
}

export default registry
