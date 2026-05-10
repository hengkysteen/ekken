import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAppSidebarStore } from './useAppSidebarStore'

describe('useAppSidebarStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('should initialize with empty items', () => {
    const store = useAppSidebarStore()
    expect(store.items).toEqual([])
  })

  it('should add sidebar items', () => {
    const store = useAppSidebarStore()
    const mockItem = {
      label: 'Test Item',
      icon: {},
      path: '/test',
      name: 'test',
      order: 10
    }

    store.addItems([mockItem])
    expect(store.items).toHaveLength(1)
    expect(store.items[0].name).toBe('test')
  })

  it('should sort sidebar items by order', () => {
    const store = useAppSidebarStore()
    const items = [
      { label: 'B', icon: {}, path: '/b', name: 'b', order: 50 },
      { label: 'A', icon: {}, path: '/a', name: 'a', order: 10 },
      { label: 'C', icon: {}, path: '/c', name: 'c', order: 100 }
    ]

    store.addItems(items)
    const sorted = store.sortedItems

    expect(sorted[0].name).toBe('a')
    expect(sorted[1].name).toBe('b')
    expect(sorted[2].name).toBe('c')
  })
})
