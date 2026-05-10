import type { WorkflowNode } from '@workflows/node/types/node'

/**
 * MyNodesItem represents a saved component/node template that can be reused
 */
export interface MyNodesItem extends WorkflowNode {
  /** Name given by user when saving to my nodes */
  name: string

  /** My nodes item creation timestamp */
  created_at: string
}
