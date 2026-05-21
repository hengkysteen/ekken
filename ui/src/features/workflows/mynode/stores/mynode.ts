import { defineStore } from 'pinia'
import { ref } from 'vue'
import { mynodeApi as api } from '@workflows/mynode/api'
import type { MyNodesItem } from '@workflows/mynode/types'
import { serializeActionForSave } from '@workflows/node/utils/node'

export const useMyNodeStore = defineStore('mynode', () => {
  const items = ref<MyNodesItem[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  async function loadItems() {
    isLoading.value = true
    error.value = null
    try {
      items.value = await api.getMyNodesItems()
    } catch (err: any) {
      console.error('Failed to load my nodes:', err)
      error.value = err?.message || 'Failed to load my nodes'
      items.value = []
    } finally {
      isLoading.value = false
    }
  }

  async function saveItem(name: string, nodeData: any) {
    const payload = {
      name,
      type: nodeData.nodeType,
      label: nodeData.label,
      tags: nodeData.tags,
      icon: nodeData.icon,
      action: serializeActionForSave(nodeData.action),
    }
    const saved = await api.saveMyNodesItem(payload)
    items.value = [saved, ...items.value]
    return saved
  }



  async function deleteItem(id: string) {
    await api.deleteMyNodesItem(id)
    items.value = items.value.filter(item => item.id !== id)
  }

  return {
    items,
    isLoading,
    error,
    loadItems,
    saveItem,
    deleteItem,
  }
})
