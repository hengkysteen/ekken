import { type Ref } from 'vue'
import { watchDebounced } from '@vueuse/core'
import { type XYPosition } from '@vue-flow/core'
import { Storage, StorageKeys } from '@shared/utils/storage'

export interface Viewport {
  x: number
  y: number
  zoom: number
}

export interface CanvasDraft {
  positions?: Record<string, XYPosition>
  viewport?: Viewport
}

export function useWorkflowDraft(
  workflowId: Ref<string | undefined>,
  flowNodes: Ref<any[]>,
  flowEdges: Ref<any[]>,
  savedViewport: Ref<Viewport>
) {
  // Load draft from storage
  function getDraft(): CanvasDraft | null {
    if (!workflowId.value) return null
    return Storage.get<CanvasDraft>(StorageKeys.CANVAS_DRAFT(workflowId.value))
  }

  // Save current state to draft
  function saveDraft() {
    if (!workflowId.value) return

    const positions: Record<string, XYPosition> = {}
    for (const n of flowNodes.value) {
      if (n.position) {
        positions[n.id] = { x: n.position.x, y: n.position.y }
      }
    }

    const key = StorageKeys.CANVAS_DRAFT(workflowId.value)
    const existing = Storage.get<CanvasDraft>(key) || {}
    
    Storage.set(key, {
      ...existing,
      positions,
      viewport: savedViewport.value
    })
  }

  // Automatic sync on changes - Debounced to avoid lag during interactions (like selecting nodes)
  watchDebounced([flowNodes, flowEdges, savedViewport], () => {
    saveDraft()
  }, { deep: true, debounce: 1000, maxWait: 3000 })

  return {
    getDraft,
    saveDraft
  }
}
