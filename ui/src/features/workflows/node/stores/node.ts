import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { nodeApi } from '../api'
import type { NodeDefinition } from '../types/node'

export const useNodeStore = defineStore('node', () => {
  const catalog = ref<NodeDefinition[]>([])
  const catalogFilterGroup = ref<string | null>(null)

  async function loadCatalog() {
    const data = await nodeApi.getCatalog()
    catalog.value = data || []
  }

  function findDef(type: string) {
    return catalog.value.find((c) => c.type === type)
  }

  const pickerGroups = computed(() => {
    const groups: Record<string, { name: string; nodes: NodeDefinition[] }> = {}

    const nodesToGroup = catalogFilterGroup.value
      ? catalog.value.filter(n => n.tags?.includes(catalogFilterGroup.value as any))
      : catalog.value

    for (const node of nodesToGroup) {
      const g = (node.tags && node.tags.length > 0) ? node.tags[0] : 'Other'
      if (!groups[g]) groups[g] = { name: g, nodes: [] }
      groups[g].nodes.push(node)
    }
    return Object.values(groups)
  })

  return {
    catalog,
    catalogFilterGroup,
    pickerGroups,
    loadCatalog,
    findDef
  }
})
