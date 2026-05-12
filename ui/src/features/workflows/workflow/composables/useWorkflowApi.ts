import { type Ref } from 'vue'
import { type XYPosition } from '@vue-flow/core'
import { workflowApi as api, type Workflow, type WorkflowEdge } from '@workflows/workflow/api'
import { mapNodesToFlow, mapEdgesToFlow, buildSavePayload as createSavePayload } from '../utils/workflowMappingUtils'

export function useWorkflowApi(
  workflow: Ref<Workflow | null>,
  flowNodes: Ref<any[]>,
  flowEdges: Ref<any[]>,
  savedViewport: Ref<any>,
  isLoading: Ref<boolean>,
  edgeStyle: Ref<string>,
  edgeAnimated: Ref<boolean>,
  getDraft: () => any
) {
  async function loadWorkflow(id: string, canvasPositions: Record<string, XYPosition> = {}) {
    isLoading.value = true
    try {
      const data = await api.getWorkflow(id)
      workflow.value = data

      const draft = getDraft() || {}
      const dbPositions = data.positions || {}
      const positions = { ...dbPositions, ...(draft.positions || {}), ...canvasPositions }

      if (draft.viewport) {
        savedViewport.value = draft.viewport
      } else {
        savedViewport.value = { x: 0, y: 0, zoom: 1 }
      }

      flowNodes.value = mapNodesToFlow(data.nodes || [], positions)
      
      const dbEdges = data.edges || []
      const edgesToLoad: WorkflowEdge[] = (draft.edges && draft.edges.length > 0) ? draft.edges : dbEdges
      
      const validEdges = edgesToLoad.filter((e) => {
        const hasSource = flowNodes.value.some(n => n.id === e.source)
        const hasTarget = flowNodes.value.some(n => n.id === e.target)
        return hasSource && hasTarget
      })

      flowEdges.value = mapEdgesToFlow(validEdges, edgeStyle.value, edgeAnimated.value)
    } finally {
      isLoading.value = false
    }
  }

  function buildSavePayload() {
    if (!workflow.value) return null
    return createSavePayload(
      workflow.value.id,
      workflow.value.name,
      workflow.value.status,
      workflow.value.created_at,
      flowNodes.value,
      flowEdges.value
    )
  }

  async function saveWorkflow(originalId: string) {
    const payload = buildSavePayload()
    if (!payload) return null
    const response = await api.updateWorkflow(originalId, payload)
    workflow.value = response
    return response
  }

  async function saveWorkflowSilent(originalId: string) {
    try {
      const payload = buildSavePayload()
      if (!payload) return
      const response = await api.updateWorkflow(originalId, payload)
      workflow.value = response
    } catch (err) {
      console.warn('Silent save failed:', err)
    }
  }

  async function handleWorkflowRename(oldId: string, newId: string) {
    if (workflow.value) workflow.value.name = newId
    return await saveWorkflow(oldId)
  }

  return {
    loadWorkflow,
    saveWorkflow,
    saveWorkflowSilent,
    handleWorkflowRename,
    buildSavePayload
  }
}
