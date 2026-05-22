import { request } from '@/shared/api/request'

export interface Conversation {
  id: string
  title: string
  created_at: string
  updated_at: string
}

export interface ConversationDetail {
  conversation: Conversation
  messages: any[]
}

export interface AssistantJobSnapshot {
  conversation_id: string
  status: 'running' | 'done' | 'error' | 'canceled'
  error?: string
  running: boolean
}

export function fetchConversations() {
  return request<Conversation[]>('/conversations')
}

export function createConversation(title?: string) {
  return request<Conversation>('/conversations', {
    method: 'POST',
    body: JSON.stringify({ title }),
  })
}

export function renameConversation(id: string, title: string) {
  return request<void>(`/conversations/${id}/rename`, {
    method: 'PUT',
    body: JSON.stringify({ title }),
  })
}

export function deleteConversation(id: string) {
  return request<void>(`/conversations/${id}`, {
    method: 'DELETE',
  })
}

export function deleteAllConversations() {
  return request<void>('/conversations', {
    method: 'DELETE',
  })
}

export function fetchConversationById(id: string) {
  return request<ConversationDetail>(`/conversations/${id}`)
}

export function addMessageToConversation(id: string, role: string, content: string, provider?: string, model?: string) {
  return request<void>(`/conversations/${id}/messages`, {
    method: 'POST',
    body: JSON.stringify({ role, content, provider, model }),
  })
}

export function chatWithProvider(id: string, model: string, messages: any[], conversationId?: string, agent?: string, thinking?: string) {
  return request<AssistantJobSnapshot>('/assistant/providers/' + id + '/chat', {
    method: 'POST',
    body: JSON.stringify({
      model,
      messages,
      stream: true,
      conversation_id: conversationId,
      agent: agent || 'chat',
      thinking
    }),
  })
}

export function fetchAssistantJob(id: string) {
  return request<AssistantJobSnapshot>(`/assistant/conversations/${id}/job`)
}

export function fetchAssistantJobs() {
  return request<AssistantJobSnapshot[]>('/assistant/jobs')
}

export async function openAssistantStream(id: string, signal?: AbortSignal) {
  const response = await fetch(`/api/assistant/conversations/${id}/stream`, { signal })
  if (!response.ok) {
    throw new Error(`Stream failed: ${response.status} ${response.statusText}`)
  }
  return response
}

export function stopChat(id: string) {
  return request<void>(`/assistant/conversations/${id}/stop`, {
    method: 'POST',
  })
}
