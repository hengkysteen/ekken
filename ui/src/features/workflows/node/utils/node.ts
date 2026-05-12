import type { NodeDefinition, WorkflowNode } from '../types/node'




export function generateNodeId(): string {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  const arr = new Uint8Array(5)
  crypto.getRandomValues(arr)
  return Array.from(arr, b => chars[b % chars.length]).join('')
}





/**
 * Builds a NodeAction instance based on a node definition and selected action key.
 * Populates default values and generates response_var if needed.
 * Does NOT merge global fields - that happens at save time.
 */
export function buildActionInstance(def?: NodeDefinition, actionKey?: string): any {
  if (!def) return { key: '', fields: [] }

  const selectedActionKey = actionKey || def.default_action || (def.actions?.[0]?.key) || ''
  const actionDef = def.actions?.find(a => a.key === selectedActionKey)

  if (!actionDef) return { key: selectedActionKey, fields: [] }

  // Clone actionDef to avoid mutating registry
  const action = { ...actionDef, fields: [...(actionDef.fields || [])] }

  // Populate default values for action fields
  action.fields = action.fields.map(f => ({
    ...f,
    value: f.default !== undefined ? f.default : undefined
  }))

  // Always generate a unique response_var for nodes that produce output
  if (action.has_response) {
    action.response_var = `${action.key}_${generateNodeId()}`
  }

  return action
}

export function calculateNodeOutputs(_type: string, _action: any, def?: NodeDefinition): any[] {
  return def?.outputs || []
}

export function buildNodeData(
  node: Partial<WorkflowNode> & { _label?: string; _tags?: string[]; _icon?: string },
  def?: NodeDefinition
) {
  // Ensure we have an action object
  let action = node.action
  if (!action && def) {
    action = buildActionInstance(def)
  }

  const outputs = calculateNodeOutputs(node.type!, action, def)

  return {
    label: node._label || def?.label || node.type,
    nodeType: node.type,
    tags: node.tags || def?.tags || [],
    icon: node._icon || def?.icon || '',
    action: action,
    outputs: outputs,
    action_has_response: action?.has_response ?? false,
    id: node.id,
    name: node.name,
    sourceType: node.sourceType,
  }
}
