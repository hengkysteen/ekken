import { type Ref } from 'vue'
import { type Node, type XYPosition } from '@vue-flow/core'
import type { Workflow } from '@workflows/workflow/api'
import type { WorkflowNode } from '@workflows/node/types/node'
import { generateNodeId, buildDefaultConfig, calculateNodeOutputs, buildNodeData } from '@workflows/node/utils/node'
import { getVueFlowType } from '@workflows/workflow/utils/vueFlowUtils'
import { useNodeStore } from '@workflows/node/stores/node'

export interface AddNodeParams {
  type: string
  label?: string
  tags?: string[]
  icon?: string
  position?: XYPosition
  config?: Record<string, any>
  response_var?: string
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

  function addNode({ type, label, tags, icon, position, config, response_var, sourceType, name }: AddNodeParams) {
    const id = generateNodeId()
    const def = nodeStore.findDef(type)
    let finalConfig = config || {}

    if (!config) {
      finalConfig = buildDefaultConfig(def)
    }

    let defaultResponseVar = ''
    if (def) {
      let actionKey = finalConfig.action || def.default_action
      if (!actionKey && def.actions && def.actions.length > 0) {
        actionKey = def.actions[0].key
      }
      
      const actionDef = def.actions?.find((a: any) => a.key === actionKey)
      if (actionDef?.has_response) {
        const cleanId = id.replace(/-/g, '_')
        defaultResponseVar = `${actionKey}_${cleanId}`
      }
    }

    const rawNode: WorkflowNode = {
      id,
      type,
      label: label || def?.label || type,
      tags: tags || def?.tags || [],
      icon: icon || def?.icon || '',
      version: def?.version,
      config: finalConfig,
      response_var: sourceType === 'mynodes' ? defaultResponseVar : (response_var || defaultResponseVar)
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
        config: finalConfig,
        tags: rawNode.tags,
        label: rawNode.label,
        icon: rawNode.icon,
        response_var: rawNode.response_var,
        name,
        sourceType
      }, def),
    }

    if (label) flowNode.data.label = label
    if (tags) flowNode.data.tags = tags
    if (icon) flowNode.data.icon = icon
    if (sourceType) flowNode.data.sourceType = sourceType

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
      config: nodeDef.config,
      response_var: nodeDef.response_var,
      sourceType: nodeDef.sourceType,
    })

    addEdge(sourceId, handleType, flowNode.id)
    return flowNode
  }

  function updateNodeConfig(id: string, { config, response_var, label }: { config: any; response_var: string; label?: string }) {
    const flowNode = flowNodes.value.find((n) => n.id === id)
    if (flowNode) {
      flowNode.data.config = { ...config }
      flowNode.data.response_var = response_var
      if (label) flowNode.data.label = label

      const def = nodeStore.findDef(flowNode.data.nodeType)
      if (def?.actions && config?.action) {
        const actionDef = def.actions.find(a => a.key === config.action)
        flowNode.data.action_has_output = actionDef?.has_output ?? false
      }

      flowNode.data.outputs = calculateNodeOutputs(flowNode.data.nodeType, config, def)
    }

    const rawNode = workflow.value?.nodes?.find((n) => n.id === id)
    if (rawNode) {
      rawNode.config = { ...config }
      rawNode.response_var = response_var
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
    let newResponseVar = ''
    if (def) {
      const config = sourceNode.data.config || {}
      let actionKey = config.action || def.default_action
      if (!actionKey && def.actions && def.actions.length > 0) {
        actionKey = def.actions[0].key
      }
      
      const actionDef = def.actions?.find((a: any) => a.key === actionKey)
      if (actionDef?.has_response) {
        newResponseVar = `${actionKey}_${newId.replace(/-/g, '_')}`
      }
    }

    const rawNode: WorkflowNode = {
      id: newId,
      type: sourceNode.data.nodeType,
      label: sourceNode.data.label,
      tags: sourceNode.data.tags || [],
      icon: sourceNode.data.icon || '',
      config: { ...(sourceNode.data.config || {}) },
      response_var: newResponseVar,
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
        response_var: rawNode.response_var
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
    updateNodeConfig,
    removeNode,
    duplicateNode,
    clearWorkflow,
    resetState
  }
}
