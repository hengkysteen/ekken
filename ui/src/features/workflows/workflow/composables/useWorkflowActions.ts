import { type Ref } from 'vue'
import { type Node, type XYPosition } from '@vue-flow/core'
import type { Workflow } from '@workflows/workflow/api'
import type { WorkflowNode } from '@workflows/node/types/node'
import { generateNodeId, buildActionInstance, calculateNodeOutputHandles, buildNodeData, serializeActionForSave } from '@workflows/node/utils/node'
import { getVueFlowType } from '@workflows/workflow/utils/vueFlowUtils'
import { useNodeStore } from '@workflows/node/stores/node'

export interface AddNodeParams {
  type: string
  label?: string
  tags?: string[]
  icon?: string
  position?: XYPosition
  action?: any
  sourceType?: 'catalog' | 'mynodes'
  name?: string
}

export function useWorkflowActions(
  workflow: Ref<Workflow | null>,
  flowNodes: Ref<any[]>,
  flowEdges: Ref<any[]>,
  edgeStyle: Ref<string>,
  edgeAnimated: Ref<boolean>,
  isLoading: Ref<boolean>,
  savedViewport: Ref<any>
) {
  const nodeStore = useNodeStore()

  function addNode({ type, label, tags, icon, position, action, sourceType, name }: AddNodeParams) {
    const id = generateNodeId()
    const def = nodeStore.findDef(type)
    
    // Build default action if not provided
    let finalAction = action ? serializeActionForSave(action) : action
    if (!finalAction && def) {
      finalAction = buildActionInstance(def)
    }

    const rawNode: WorkflowNode = {
      id,
      type,
      label: label || def?.label || type,
      tags: tags || def?.tags || [],
      icon: icon || def?.icon || '',
      version: def?.version,
      action: finalAction,
    }

    if (workflow.value) {
      workflow.value.nodes = [...(workflow.value.nodes || []), rawNode]
    }

    const flowNode: Node = {
      id,
      type: getVueFlowType(type),
      position: position || { x: 100, y: 200 },
      data: buildNodeData({
        id,
        type,
        version: rawNode.version,
        action: finalAction,
        tags: rawNode.tags,
        label: rawNode.label,
        icon: rawNode.icon,
        name,
        sourceType
      }, def),
    }

    flowNodes.value = [...flowNodes.value, flowNode]
    return flowNode
  }

  function addEdge(sourceId: string, sourceHandle: string | null, targetId: string) {
    const handleType = sourceHandle || 'success'
    const edgeId = `e-${sourceId}-${targetId}-${handleType}`

    flowEdges.value = [
      ...flowEdges.value.filter(
        (e) => !(e.source === sourceId && (e.sourceHandle || 'success') === handleType)
      ),
      {
        id: edgeId,
        source: sourceId,
        sourceHandle: handleType,
        target: targetId,
        type: edgeStyle.value,
        animated: edgeAnimated.value,
        style: { strokeWidth: 2 },
        markerEnd: { type: 'arrowclosed' } as any,
      },
    ]
  }

  function addNodeFromHandle(sourceId: string, handleType: string, nodeDef: any) {
    const sourceNode = flowNodes.value.find((n) => n.id === sourceId)
    const sourcePos = sourceNode ? sourceNode.position : { x: 100, y: 100 }

    const position = {
      x: sourcePos.x + 300,
      y: handleType === 'failure' ? sourcePos.y + 150 : sourcePos.y,
    }

    const flowNode = addNode({
      type: nodeDef.type,
      label: nodeDef.label,
      tags: nodeDef.tags,
      icon: nodeDef.icon,
      position,
      action: nodeDef.action,
      sourceType: nodeDef.sourceType,
    })

    addEdge(sourceId, handleType, flowNode.id)
    return flowNode
  }

  function updateNodeAction(id: string, { action, label }: { action: any; label?: string }) {
    const savedAction = serializeActionForSave(action)
    const flowNode = flowNodes.value.find((n) => n.id === id)
    if (flowNode) {
      flowNode.data.action = savedAction
      if (label) flowNode.data.label = label

      const def = nodeStore.findDef(flowNode.data.nodeType)
      if (def?.version) {
        flowNode.data.version = def.version
        flowNode.data.installedVersion = def.version
        flowNode.data.needsReview = false
      }
      flowNode.data.action_has_response = def?.actions?.find(a => a.type === flowNode.data.action?.type)?.has_response ?? false
      flowNode.data.output_handles = calculateNodeOutputHandles(flowNode.data.nodeType, flowNode.data.action, def)
      flowNode.data.hide_input_handles = def?.hide_input_handles || false
    }

    const rawNode = workflow.value?.nodes?.find((n) => n.id === id)
    if (rawNode) {
      rawNode.action = savedAction
      if (label) rawNode.label = label
      const def = nodeStore.findDef(rawNode.type)
      if (def?.version) rawNode.version = def.version
    }
  }

  function removeNode(id: string) {
    if (workflow.value) {
      workflow.value.nodes = (workflow.value.nodes || []).filter((n) => n.id !== id)
    }
    flowNodes.value = flowNodes.value.filter((n) => n.id !== id)
    flowEdges.value = flowEdges.value.filter((e) => e.source !== id && e.target !== id)
    return true
  }

  function duplicateNode(id: string) {
    const sourceNode = flowNodes.value.find((n) => n.id === id)
    if (!sourceNode) return

    const newId = generateNodeId()
    const def = nodeStore.findDef(sourceNode.data.nodeType)
    
    // Deep clone the action and generate new response_var
    const newAction = serializeActionForSave(JSON.parse(JSON.stringify(sourceNode.data.action || {})))
    const actionHasResponse = def?.actions?.find(a => a.type === newAction.type)?.has_response ?? false
    if (actionHasResponse) {
      newAction.response_var = `${sourceNode.data.nodeType}.${newAction.type}_${generateNodeId()}`
    }

    const rawNode: WorkflowNode = {
      id: newId,
      type: sourceNode.data.nodeType,
      label: sourceNode.data.label,
      tags: sourceNode.data.tags || [],
      icon: sourceNode.data.icon || '',
      version: sourceNode.data.version,
      action: newAction,
    }

    if (workflow.value) {
      workflow.value.nodes = [...(workflow.value.nodes || []), rawNode]
    }

    const flowNode: Node = {
      id: newId,
      type: getVueFlowType(sourceNode.data.nodeType),
      position: {
        x: (sourceNode.position?.x || 0) + 150,
        y: sourceNode.position?.y || 0,
      },
      data: buildNodeData({
        ...rawNode,
        label: rawNode.label,
        tags: rawNode.tags,
        icon: rawNode.icon,
      }, def),
    }

    flowNodes.value = [...flowNodes.value, flowNode]
    return flowNode
  }

  function clearWorkflow() {
    if (workflow.value) workflow.value.nodes = []
    flowNodes.value = []
    flowEdges.value = []
    isLoading.value = false
    savedViewport.value = { x: 0, y: 0, zoom: 1 }
  }

  function resetState() {
    workflow.value = null
    flowNodes.value = []
    flowEdges.value = []
    isLoading.value = true
  }

  return {
    addNode,
    addEdge,
    addNodeFromHandle,
    updateNodeConfig: updateNodeAction, // Alias for compatibility
    updateNodeAction,
    removeNode,
    duplicateNode,
    clearWorkflow,
    resetState
  }
}
