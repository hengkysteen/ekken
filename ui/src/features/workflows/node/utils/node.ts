import type { NodeAction, NodeDefinition, NodeField, WorkflowNode } from '../types/node'




export function generateNodeId(): string {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  const arr = new Uint8Array(5)
  crypto.getRandomValues(arr)
  return Array.from(arr, b => chars[b % chars.length]).join('')
}





export function getActionType(action?: any): string {
  return action?.type || ''
}

export function getActionBlueprint(def?: NodeDefinition, actionType?: string): NodeAction | undefined {
  if (!def) return undefined
  const selectedActionType = actionType || def.default_action || def.actions?.[0]?.type
  return def.actions?.find(a => a.type === selectedActionType) || def.actions?.[0]
}

export function fieldsToValueMap(fields?: Array<Partial<NodeField>>): Record<string, any> {
  const result: Record<string, any> = {}
  for (const field of fields || []) {
    if (!field?.key) continue
    result[field.key] = field.value
  }
  return result
}

export function getActionValue(action: any, key: string, fallback?: any): any {
  const field = action?.fields?.find((f: any) => f.key === key)
  if (field && field.value !== undefined) return field.value
  if (field && field.default !== undefined) return field.default
  return fallback
}

export function hydrateFieldsForForm(
  blueprintFields?: NodeField[],
  instanceFields?: Array<Partial<NodeField>>
): NodeField[] {
  const values = fieldsToValueMap(instanceFields)
  return (blueprintFields || []).map(field => ({
    ...field,
    value: values[field.key] !== undefined
      ? values[field.key]
      : field.default !== undefined ? field.default : undefined
  }))
}

export function hydrateActionForForm(action: any, def?: NodeDefinition, actionType?: string): any {
  const selectedActionType = actionType || getActionType(action)
  const blueprint = getActionBlueprint(def, selectedActionType)
  if (!blueprint) return action || { type: selectedActionType || '', fields: [] }

  return {
    ...blueprint,
    type: blueprint.type,
    response_var: action?.response_var || blueprint.response_var,
    fields: hydrateFieldsForForm(blueprint.fields, action?.fields)
  }
}

export function serializeActionForSave(action: any): any {
  if (!action) return { type: '', fields: [] }

  const result: any = {
    type: getActionType(action),
    fields: (action.fields || [])
      .filter((field: any) => field?.key)
      .map((field: any) => ({ key: field.key, value: field.value }))
  }

  if (action.response_var) {
    result.response_var = action.response_var
  }

  return result
}

/**
 * Builds a minimal NodeAction instance based on a node definition and selected action type.
 */
export function buildActionInstance(def?: NodeDefinition, actionType?: string): any {
  if (!def) return { type: '', fields: [] }

  const actionDef = getActionBlueprint(def, actionType)
  const selectedActionType = actionDef?.type || actionType || def.default_action || ''

  if (!actionDef) return { type: selectedActionType, fields: [] }

  const action: any = {
    type: selectedActionType,
    fields: (actionDef.fields || []).map(f => ({
      key: f.key,
      value: f.default !== undefined ? f.default : undefined
    }))
  }

  if (actionDef.has_response) {
    action.response_var = `${def.type}.${selectedActionType}_${generateNodeId()}`
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
  } else if (action) {
    action = serializeActionForSave(action)
  }

  const outputs = calculateNodeOutputs(node.type!, action, def)
  const actionBlueprint = getActionBlueprint(def, getActionType(action))

  return {
    label: node._label || node.label || def?.label || node.type,
    nodeType: node.type,
    tags: node._tags || node.tags || def?.tags || [],
    icon: node._icon || node.icon || def?.icon || '',
    action: action,
    outputs: outputs,
    action_has_response: actionBlueprint?.has_response ?? false,
    id: node.id,
    name: node.name,
    sourceType: node.sourceType,
  }
}
