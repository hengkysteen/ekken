import { request } from '@shared/api/request'
import type { NodeDefinition } from '@workflows/node/types/node'

export const nodeApi = {
  getCatalog: () => request<NodeDefinition[]>('/nodes/catalog'),
}

