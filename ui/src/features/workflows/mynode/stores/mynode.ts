import { defineStore } from 'pinia'
import { ref } from 'vue'
import { mynodeApi as api } from '@workflows/mynode/api'
import type { MyNodesItem } from '@workflows/mynode/types'
import { serializeActionForSave } from '@workflows/node/utils/node'

export const useMyNodeStore = defineStore('mynode', () => {
  const items = ref<MyNodesItem[]>([])
  const isLoading = ref(false)

  async function loadItems() {
    isLoading.value = true
    try {
      items.value = await api.getMyNodesItems()
    } catch (err) {
      console.error('Failed to load my nodes:', err)
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
    loadItems,
    saveItem,
    deleteItem,
  }
})
