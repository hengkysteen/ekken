/**
 * Maps our internal node types to Vue Flow component types.
 * Currently all nodes use the 'customNode' component.
 */
export function getVueFlowType(_type: string): string {
  return 'customNode'
}
