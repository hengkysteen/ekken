import type { NodeDefinition } from '../types/node'
import { getActionType } from './node'

export function matchesSearch(node: NodeDefinition, query: string): boolean {
  if (!query) return true
  const q = query.toLowerCase()
  
  const matchesBasic = 
    node.label.toLowerCase().includes(q) ||
    node.type.toLowerCase().includes(q) ||
    (node.tags && node.tags.some(t => t.toLowerCase().includes(q)))
    
  if (matchesBasic) return true
  
  // Check actions
  if (node.actions && node.actions.some(a => 
    a.label.toLowerCase().includes(q) ||
    a.type.toLowerCase().includes(q)
  )) {
    return true
  }
  
  return false
}

export function matchesMyNodeSearch(item: any, query: string): boolean {
  if (!query) return true
  const q = query.toLowerCase()
  
  return (
    (item.name && item.name.toLowerCase().includes(q)) ||
    item.type.toLowerCase().includes(q) ||
    (item.tags && item.tags.some((t: string) => t.toLowerCase().includes(q))) ||
    getActionType(item.action).toLowerCase().includes(q)
  )
}
