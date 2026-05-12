import { type XYPosition } from '@vue-flow/core'
import type { Workflow, WorkflowEdge } from '@workflows/workflow/api'
import type { WorkflowNode } from '@workflows/node/types/node'
import { buildNodeData } from '@workflows/node/utils/node'
import { getVueFlowType } from '@workflows/workflow/utils/vueFlowUtils'
import { useNodeStore } from '@workflows/node/stores/node'

export function buildSavePayload(
  workflowId: string,
  workflowName: string,
  workflowStatus: string,
  createdAt: number,
  flowNodes: any[],
  flowEdges: any[]
): Workflow {
  const nodes: WorkflowNode[] = flowNodes.map((n) => ({
    id: n.id,
    type: n.data.nodeType || n.type,
    label: n.data.label,
    icon: n.data.icon,
    tags: n.data.tags,
    action: n.data.action,
  }))

  const edges: WorkflowEdge[] = flowEdges.map((e) => ({
    source: e.source,
    sourceHandle: (e.sourceHandle as string) || 'success',
    target: e.target,
  }))

  const positions: Record<string, XYPosition> = {}
  for (const n of flowNodes) {
    if (n.position) positions[n.id] = { x: n.position.x, y: n.position.y }
  }

  return {
    id: workflowId,
    name: workflowName,
    status: workflowStatus,
    created_at: createdAt,
    nodes,
    edges,
    positions,
  }
}

export function mapNodesToFlow(nodes: WorkflowNode[], positions: Record<string, XYPosition>) {
  const nodeStore = useNodeStore()
  return nodes.map((n, i) => {
    const pos = positions[n.id] || { x: Math.random() * 400 + 50, y: Math.random() * 300 + 50 }
    const def = nodeStore.findDef(n.type)
    return {
      id: n.id || `node_${i}`,
      type: getVueFlowType(n.type),
      position: { x: pos.x, y: pos.y },
      data: buildNodeData(n, def),
    }
  })
}

export function mapEdgesToFlow(edges: WorkflowEdge[], edgeStyle: string, edgeAnimated: boolean) {
  return edges.map((e) => ({
    id: `e-${e.source}-${e.target}-${e.sourceHandle || 'success'}`,
    source: e.source,
    sourceHandle: e.sourceHandle || 'success',
    target: e.target,
    type: edgeStyle,
    animated: edgeAnimated,
    style: { strokeWidth: 2 },
    markerEnd: { type: 'arrowclosed' } as any,
  }))
}
