import { request } from '@shared/api/request'
import type { MyNodesItem } from '@workflows/mynode/types'

export const mynodeApi = {
  getMyNodesItems: () => request<MyNodesItem[]>('/mynodes'),
  saveMyNodesItem: (item: Partial<MyNodesItem>) =>
    request<MyNodesItem>('/mynodes', { method: 'POST', body: JSON.stringify(item) }),
  deleteMyNodesItem: (id: string) =>
    request<void>(`/mynodes/${id}`, { method: 'DELETE' }),
  updateMyNodesItem: (id: string, item: Partial<MyNodesItem>) =>
    request<MyNodesItem>(`/mynodes/${id}`, { method: 'PUT', body: JSON.stringify(item) }),
}
