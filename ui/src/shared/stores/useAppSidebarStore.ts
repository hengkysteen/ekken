import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { SidebarItem } from '../../core/types/module'

export const useAppSidebarStore = defineStore('appSidebar', () => {
  const items = ref<SidebarItem[]>([])

  const addItems = (newItems: SidebarItem[]) => {
    items.value.push(...newItems)
  }

  const sortedItems = computed(() => {
    return [...items.value].sort((a, b) => (a.order ?? 99) - (b.order ?? 99))
  })

  return {
    items,
    addItems,
    sortedItems
  }
})
