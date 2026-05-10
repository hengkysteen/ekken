import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAppSettingsStore } from './useAppSettingsStore'
import { markRaw } from 'vue'

describe('useAppSettingsStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('should initialize with empty settings tabs', () => {
    const store = useAppSettingsStore()
    expect(store.settingsTabs).toEqual([])
  })

  it('should add settings tabs', () => {
    const store = useAppSettingsStore()
    const mockTab = {
      id: 'test',
      label: 'Test Tab',
      icon: {},
      component: markRaw({ template: '<div>Test</div>' }),
      order: 10
    }

    store.addSettingsTabs([mockTab])
    expect(store.settingsTabs).toHaveLength(1)
    expect(store.settingsTabs[0].id).toBe('test')
  })

  it('should sort settings tabs by order', () => {
    const store = useAppSettingsStore()
    const tabs = [
      { id: 'b', label: 'B', icon: {}, component: {}, order: 50 },
      { id: 'a', label: 'A', icon: {}, component: {}, order: 10 },
      { id: 'c', label: 'C', icon: {}, component: {}, order: 100 }
    ]

    store.addSettingsTabs(tabs)
    const sorted = store.sortedSettingsTabs

    expect(sorted[0].id).toBe('a')
    expect(sorted[1].id).toBe('b')
    expect(sorted[2].id).toBe('c')
  })

  it('should use default order if not provided', () => {
    const store = useAppSettingsStore()
    const tabs = [
      { id: 'high', label: 'High', icon: {}, component: {}, order: 10 },
      { id: 'none', label: 'None', icon: {}, component: {} }, // Default 99
    ]

    store.addSettingsTabs(tabs)
    const sorted = store.sortedSettingsTabs

    expect(sorted[0].id).toBe('high')
    expect(sorted[1].id).toBe('none')
  })
})
