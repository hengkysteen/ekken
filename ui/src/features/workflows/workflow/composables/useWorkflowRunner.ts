import { ref, computed, watch, onBeforeUnmount, type Ref } from 'vue'
import { useTimeoutFn } from '@vueuse/core'
import { ElMessage } from 'element-plus'
import { workflowApi as api } from '@workflows/workflow/api'
import { useWorkflowStore } from '@workflows/workflow/stores/workflow'

export interface LogEntry {
  time: string
  level: string
  message: string
  raw?: string
}

export function useWorkflowRunner(workflowIdRef: Ref<string>) {
  const workflowStore = useWorkflowStore()
  const isRunning = computed(() => workflowStore.getStatus(workflowIdRef.value) === 'running')
  
  // isSubmitting murni untuk mencegah klik dobel saat API request (bukan urusan SSE/validasi)
  const isSubmitting = ref(false)
  const logs = ref<LogEntry[]>([])

  const message = ElMessage
  let eventSource: EventSource | null = null

  const { start: startSync } = useTimeoutFn(() => syncStatus(), 1000, { immediate: false })

  function syncStatus() {
    if (!workflowIdRef.value) return
    api.getWorkflowStatus(workflowIdRef.value).catch(() => {})
  }

  function connectSSE() {
    if (!workflowIdRef.value) return
    disconnectSSE()

    const url = `/api/workflows/${workflowIdRef.value}/events`
    eventSource = new EventSource(url)

    eventSource.addEventListener('log_entry', (e: MessageEvent) => {
      try {
        const log: LogEntry = JSON.parse(e.data)
        const newLogs = [...logs.value, log]
        if (newLogs.length > 1000) newLogs.shift()
        logs.value = newLogs
      } catch { }
    })

    eventSource.onerror = () => {
      const status = workflowStore.getStatus(workflowIdRef.value)
      const terminalStatuses = ['done', 'error', 'stopped', 'idle']
      if (terminalStatuses.includes(status)) {
        disconnectSSE()
      } else {
        console.warn('SSE Connection lost, attempting to reconnect...')
      }
    }
  }

  // SSOT: Close log connection only when status is truly terminal (done, error, stopped, idle)
  watch(() => workflowStore.getStatus(workflowIdRef.value), (status) => {
    const terminalStatuses = ['done', 'error', 'stopped', 'idle']
    if (terminalStatuses.includes(status)) {
      disconnectSSE()
      startSync()
    }
  })

  function disconnectSSE() {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
  }

  async function handleRun(beforeRunCb?: () => Promise<void>) {
    if (isSubmitting.value || isRunning.value) return
    isSubmitting.value = true
    try {
      if (beforeRunCb) await beforeRunCb()
      connectSSE()
      await api.runWorkflow(workflowIdRef.value)
      message.info('Workflow execution started')
      useTimeoutFn(() => syncStatus(), 5000)
    } catch (err: any) {
      message.error({
        message: '<strong>Execution Failed:</strong><br/>' + (err.message || 'Unknown error').replace(/\n/g, '<br/>'),
        dangerouslyUseHTMLString: true,
        duration: 5000
      })
    } finally {
      isSubmitting.value = false
    }
  }

  async function handleStop() {
    if (isSubmitting.value || !isRunning.value) return
    isSubmitting.value = true
    try {
      await api.stopWorkflow(workflowIdRef.value)
      message.success('Workflow stopped successfully')
      syncStatus()
    } catch (err) {
      message.error('Failed to stop workflow')
    } finally {
      isSubmitting.value = false
    }
  }

  onBeforeUnmount(() => disconnectSSE())

  return {
    isRunning,
    isSubmitting,
    logs,
    setLogs: (newLogs: LogEntry[]) => { logs.value = newLogs },
    syncStatus,
    connectSSE,
    disconnectSSE,
    handleRun,
    handleStop
  }
}
