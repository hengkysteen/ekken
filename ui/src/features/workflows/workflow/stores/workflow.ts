import { defineStore } from 'pinia'
import { ref } from 'vue'
import { workflowApi as api } from '@workflows/workflow/api'
import type { Workflow } from '@workflows/workflow/api'
import { Storage } from '@shared/utils/storage'
import { useWorkflowStatusSSE } from '@workflows/workflow/composables/useWorkflowStatusSSE'
import { ElNotification } from 'element-plus'
import { truncate } from '@/shared/utils/string'

export const useWorkflowStore = defineStore('workflow', () => {
  // --- State ---
  const workflows = ref<Workflow[]>([])
  const statuses = ref<Record<string, string>>({})
  const loading = ref(false)
  const initialized = ref(false)
  const sse = useWorkflowStatusSSE()

  // --- SSE & Status Management ---

  function connectSSE() {
    if (sse.connected.value) return
    sse.connect((id, status, name) => {
      statuses.value[id] = status

      // Notify on terminal states (done, error, stopped)
      if (status === 'done' || status === 'error' || status === 'stopped') {
        const displayName = name || id

        ElNotification({
          title: `Workflow ${status}`,
          message: `"${truncate(displayName, 30)}" ${status}`,
          type: status === 'done' ? 'success' : status === 'error' ? 'error' : 'info',
          position: 'top-right',
          showClose: true,
          duration: 6000,
        })
      }
    })
  }

  function disconnectSSE() {
    sse.disconnect()
  }

  function getStatus(id: string) {
    return statuses.value[id] || 'idle'
  }

  function getStatusSeverity(id: string) {
    const status = getStatus(id)
    const severityMap: Record<string, string> = {
      idle: 'secondary',
      running: 'success',
      error: 'danger',
    }
    return severityMap[status] || 'secondary'
  }

  // --- CRUD Actions ---

  async function fetchWorkflows() {
    loading.value = true
    try {
      const data = await api.getWorkflows()
      workflows.value = data || []

      // Sync initial statuses from backend
      for (const wf of workflows.value) {
        const backendStatus = wf.status || 'idle'
        if (statuses.value[wf.id] === undefined) {
          statuses.value[wf.id] = backendStatus
        }
      }
    } catch (err) {
      console.error('Failed to load workflows:', err)
    } finally {
      loading.value = false
      initialized.value = true
    }
  }

  async function createWorkflow(name: string) {
    const wf: Partial<Workflow> = {
      name: name.trim(),
      nodes: [],
    }
    const response = await api.createWorkflow(wf)
    await fetchWorkflows()
    return response.id
  }

  async function deleteWorkflow(id: string) {
    await api.deleteWorkflow(id)
    Storage.clearWorkflowData(id)
    await fetchWorkflows()
  }

  return {
    // State
    workflows,
    statuses,
    loading,
    initialized,

    // SSE
    connectSSE,
    disconnectSSE,

    // Getters/Helpers
    getStatus,
    getStatusSeverity,

    // CRUD
    fetchWorkflows,
    createWorkflow,
    deleteWorkflow,

  }
})
