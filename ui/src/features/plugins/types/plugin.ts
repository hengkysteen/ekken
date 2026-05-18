/**
 * Plugin component contract definitions.
 * These types define the interface that plugin UI components must implement.
 */

import type { Component } from 'vue'
import type { NodeFormProps, NodeFormComponent } from '@workflows/node/types/node'

// Re-export for backward compatibility
export type { NodeFormProps, NodeFormComponent }

/**
 * API provided to plugin registration functions.
 * Plugins use this to register their custom config components.
 */
export interface PluginNodeConfigAPI {
  /**
   * Register a custom config component for a node type.
   *
   * @param nodeType - The node type identifier (must match plugin.json)
   * @param component - Vue component that implements NodeFormComponent interface
   *
   * @example
   * ```typescript
   * api.registerNodeConfig('my_node', MyConfigComponent)
   * ```
   */
  registerNodeConfig: (
    nodeType: string,
    component: Component<NodeFormProps>
  ) => void
}

/**
 * Function signature for plugin registration.
 * Plugin UI modules must export a function with this signature.
 *
 * @param api - API object for registering components
 *
 * @example
 * ```typescript
 * export function register(api: PluginNodeConfigAPI) {
 *   api.registerNodeConfig('my_node', MyConfigComponent)
 * }
 * ```
 */
export type NodeConfigRegistrar = (api: PluginNodeConfigAPI) => void

/**
 * Plugin UI module structure.
 * The module can export either a default function or a named 'register' function.
 */
export interface PluginNodeConfigModule {
  /** Default export (if using default export) */
  default?: NodeConfigRegistrar

  /** Named export (preferred) */
  register?: NodeConfigRegistrar
}
