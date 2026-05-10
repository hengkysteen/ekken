import { request } from '@shared/api/request'
import type { WorkflowNode } from '@workflows/node/types/node'

export interface WorkflowEdge {
  source: string
  sourceHandle: string
  target: string
}

export interface Workflow {
  id: string
  name: string
  status: string
  created_at: number
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
  positions?: Record<string, { x: number; y: number }>
  trigger?: any
  last_run_at?: string
}

export interface LogEntry {
  time: string
  level: string
  message: string
}

export const workflowApi = {
  // Workflows
  getWorkflows: () => request<Workflow[]>('/workflows'),
  getWorkflowsStatus: () => request<Array<{ id: string, name: string, status: string }>>('/workflows/status'),
  getWorkflow: (id: string) => request<Workflow>(`/workflows/${id}`),
  createWorkflow: (wf: Partial<Workflow>) =>
    request<Workflow>('/workflows', { method: 'POST', body: JSON.stringify(wf) }),
  updateWorkflow: (id: string, wf: Partial<Workflow>) =>
    request<Workflow>(`/workflows/${id}`, { method: 'PUT', body: JSON.stringify(wf) }),
  deleteWorkflow: (id: string) =>
    request<void>(`/workflows/${id}`, { method: 'DELETE' }),
  deleteAllWorkflows: () =>
    request<void>('/workflows', { method: 'DELETE' }),
  runWorkflow: (id: string) =>
    request<void>(`/workflows/${id}/run`, { method: 'POST' }),
  stopWorkflow: (id: string) =>
    request<void>(`/workflows/${id}/stop`, { method: 'POST' }),
  getWorkflowStatus: (id: string) => request<any>(`/workflows/${id}/status`),
  getWorkflowLogs: (id: string) => request<LogEntry[]>(`/workflows/${id}/logs`),
  deleteWorkflowLogs: (id: string) => request<void>(`/workflows/${id}/logs`, { method: 'DELETE' }),
}
