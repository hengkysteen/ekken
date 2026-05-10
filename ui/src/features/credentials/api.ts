import { request } from '@shared/api/request'

export interface Credential {
  id: string
  name: string
  key: string
  value?: string   // Only present when fetched by ID
  tags: string[]
  created_at?: string
  updated_at?: string
}

export interface CreateCredentialPayload {
  name: string
  key: string
  value: string
  tags: string[]
}

export interface UpdateCredentialPayload {
  name: string
  key: string
  value: string
  tags: string[]
}

export const credentialsApi = {
  list: () => request<Credential[]>('/credentials'),
  get: (id: string) => request<Credential>(`/credentials/${id}`),
  create: (payload: CreateCredentialPayload) =>
    request<Credential>('/credentials', { method: 'POST', body: JSON.stringify(payload) }),
  update: (id: string, payload: UpdateCredentialPayload) =>
    request<Credential>(`/credentials/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),
  delete: (id: string) =>
    request<{ deleted: string }>(`/credentials/${id}`, { method: 'DELETE' }),
}
