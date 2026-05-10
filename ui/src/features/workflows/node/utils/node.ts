import type { NodeDefinition, WorkflowNode } from '../types/node'




export function generateNodeId(): string {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  const arr = new Uint8Array(5)
  crypto.getRandomValues(arr)
  return Array.from(arr, b => chars[b % chars.length]).join('')
}





export function buildDefaultConfig(def?: NodeDefinition): Record<string, any> {
  const finalConfig: Record<string, any> = {}

  if (!def) return finalConfig

  const selectedAction = def.default_action || (def.actions?.[0]?.key) || ''
  finalConfig.action = selectedAction

  if (def.global_fields) {
    def.global_fields.forEach(field => {
      if (field.default !== undefined) {
        finalConfig[field.key] = field.default
      }
    })
  }

  if (def.actions) {
    const actionDef = def.actions.find(a => a.key === selectedAction)
    if (actionDef) {
      actionDef.fields.forEach(field => {
        if (field.default !== undefined) {
          finalConfig[field.key] = field.default
        }
      })
    }
  }


  return finalConfig
}

export function calculateNodeOutputs(_type: string, _config: any, def?: NodeDefinition): any[] {
  return def?.outputs || []
}

export function buildNodeData(
  node: Partial<WorkflowNode> & { _label?: string; _tags?: string[]; _icon?: string },
  def?: NodeDefinition
) {
  const outputs = calculateNodeOutputs(node.type!, node.config, def)
  const finalConfig = { ...(node.config || {}) }

  if (!finalConfig.action && def) {
    if (def.default_action) {
      finalConfig.action = def.default_action
    } else if (def.actions && def.actions.length > 0) {
      finalConfig.action = def.actions[0].key
    }
  }

  let actionHasResponse = false
  if (def?.actions && finalConfig.action) {
    const actionDef = def.actions.find(a => a.key === finalConfig.action)
    actionHasResponse = actionDef?.has_response ?? false
  }

  return {
    label: node._label || def?.label || node.type,
    nodeType: node.type,
    tags: node.tags || def?.tags || [],
    icon: node._icon || def?.icon || '',
    config: finalConfig,
    response_var: node.response_var || '',
    outputs: outputs,
    action_has_response: actionHasResponse,
    id: node.id,
    name: node.name,
    sourceType: node.sourceType,
  }
}
