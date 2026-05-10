/**
 * Centralized Storage Utility for Ekken
 * Provides typed access and debug logging for localStorage.
 */

export const DEFAULT_SETTINGS_NAME = 'ekken-workflow-settings'

export const StorageKeys = {
  THEME_DARK: 'ekken-theme-dark',
  THEME_MODE: 'ekken-theme-mode',
  SIDEBAR_COLLAPSED: 'ekken-sidebar-collapsed',
  SIDEBAR_RIGHT_COLLAPSED: 'ekken-sidebar-right-collapsed',
  ASSISTANT_SIDEBAR_COLLAPSED: 'ekken-assistant-sidebar-collapsed',
  ASSISTANT_NEXT_GREETING: 'ekken-assistant-next-greeting',
  CANVAS_DRAFT: (id: string) => `ekken-canvas-${id}`,
  EDGE_STYLE: 'ekken-edge-style',
  EDGE_ANIMATED: 'ekken-edge-animated',
  DEFAULT_SETTINGS_NAME: 'ekken-workflow-settings',
}

export const Storage = {
  /**
   * Get and parse data from localStorage
   */
  get<T>(key: string): T | null {
    try {
      const val = localStorage.getItem(key)
      if (val === null || val === 'undefined') return null
      
      // Try parsing as JSON first
      try {
        const parsed = JSON.parse(val)
        return parsed as T
      } catch {
        // Fallback for simple strings that are not JSON
        return val as unknown as T
      }
    } catch (e) {
      console.error(`[Storage] Error reading key "${key}":`, e)
      return null
    }
  },

  /**
   * Set data to localStorage with automatic stringification
   */
  set<T>(key: string, val: T): void {
    try {
      const raw = typeof val === 'string' ? val : JSON.stringify(val)
      localStorage.setItem(key, raw)
    } catch (e) {
      console.error(`[Storage] Error writing key "${key}":`, e)
    }
  },

  /**
   * Remove a specific key
   */
  remove(key: string): void {
    localStorage.removeItem(key)
  },

  /**
   * Migrate settings from old unified key to new granular keys
   */
  migrateOldSettings(): void {
    const oldKey = 'ekken-workflow-settings'
    const oldData = this.get<any>(oldKey)
    if (oldData && typeof oldData === 'object') {
      console.log('[Storage] Migrating old settings...', oldData)
      if (oldData.edgeStyle) this.set(StorageKeys.EDGE_STYLE, oldData.edgeStyle)
      if (oldData.edgeAnimated !== undefined) this.set(StorageKeys.EDGE_ANIMATED, oldData.edgeAnimated)
      
      // Migrate theme mode if it was 'system'
      const oldTheme = localStorage.getItem(StorageKeys.THEME_MODE)
      if (oldTheme === 'system') {
        this.set(StorageKeys.THEME_MODE, 'auto')
      }
      // Better keep it for safety for now but mark as migrated
      this.set(oldKey + '-migrated', true)
    }
  },

  /**
   * Clear all frontend data associated with a specific workflow
   */
  clearWorkflowData(id: string): void {
    this.remove(StorageKeys.CANVAS_DRAFT(id))
  }
}

// Auto-run migration on load
Storage.migrateOldSettings()
