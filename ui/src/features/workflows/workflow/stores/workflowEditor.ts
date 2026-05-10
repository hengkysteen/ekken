import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useStorage } from '@vueuse/core'
import { StorageKeys } from '@shared/utils/storage'
import { buildNodeData } from '@workflows/node/utils/node'
import { DEFAULT_EDGE_STYLE, DEFAULT_EDGE_ANIMATED } from '@workflows/workflow/utils/workflowSettings'
import { getVueFlowType } from '@workflows/workflow/utils/vueFlowUtils'

// Composables
import { useWorkflowDraft, type Viewport } from '../composables/useWorkflowDraft'
import { useWorkflowFile } from '../composables/useWorkflowFile'
import { useWorkflowActions } from '../composables/useWorkflowActions'
import { useWorkflowApi } from '../composables/useWorkflowApi'
import type { Workflow } from '../api'

export const useWorkflowEditorStore = defineStore('workflowEditor', () => {
  // --- Appearance Settings ---
  const edgeStyle = useStorage(StorageKeys.EDGE_STYLE, DEFAULT_EDGE_STYLE)
  const edgeAnimated = useStorage(StorageKeys.EDGE_ANIMATED, DEFAULT_EDGE_ANIMATED)

  // --- State ---
  const workflow = ref<Workflow | null>(null)
  const flowNodes = ref<any[]>([])
  const flowEdges = ref<any[]>([])
  const savedViewport = ref<Viewport>({ x: 0, y: 0, zoom: 1 })
  const isLoading = ref(true)
  const showSidebar = useStorage<boolean>(StorageKeys.SIDEBAR_RIGHT_COLLAPSED, true)
  const mynodesTabKey = ref(0)

  // --- Initialize Composables ---
  const { getDraft } = useWorkflowDraft(
    computed(() => workflow.value?.id),
    flowNodes,
    flowEdges,
    savedViewport
  )

  const { exportWorkflowToFile: exportToFile, importWorkflowFromJson: importFromJson } = useWorkflowFile()

  const actions = useWorkflowActions(
    workflow,
    flowNodes,
    flowEdges,
    edgeStyle,
    edgeAnimated,
    isLoading,
    savedViewport
  )

  const apiActions = useWorkflowApi(
    workflow,
    flowNodes,
    flowEdges,
    savedViewport,
    isLoading,
    edgeStyle,
    edgeAnimated,
    getDraft
  )

  // --- Store Local Actions (Mappings) ---
  function importWorkflow(imported: Workflow) {
    flowNodes.value = imported.nodes.map((n, i) => {
      return {
        id: n.id,
        type: getVueFlowType(n.type),
        position: n.position || { x: 100 + i * 250, y: 200 },
        data: buildNodeData(n),
      }
    })

    flowEdges.value = (imported.edges || []).map((e) => ({
      id: `e-${e.source}-${e.target}-${e.sourceHandle}`,
      source: e.source,
      sourceHandle: e.sourceHandle,
      target: e.target,
      type: 'default',
      style: { strokeWidth: 2 },
      markerEnd: { type: 'arrowclosed' } as any,
    }))
  }

  function exportWorkflowToFile(filename: string) {
    const payload = apiActions.buildSavePayload()
    if (payload) exportToFile(payload, filename)
  }

  function importWorkflowFromJson(jsonString: string) {
    const imported = importFromJson(jsonString)
    importWorkflow(imported)
    return imported
  }

  return {
    // State
    edgeStyle, edgeAnimated, workflow, flowNodes, flowEdges, savedViewport, isLoading, showSidebar, mynodesTabKey,

    // Actions from Composables
    ...actions,
    ...apiActions,

    // Overrides/Mappings
    exportWorkflowToFile,
    importWorkflowFromJson,
  }
})