import { useEventSource } from '@vueuse/core'
import { watch, ref } from 'vue'

export function useWorkflowStatusSSE() {
  const url = ref<string | undefined>(undefined)
  
  // VueUse useEventSource menangani auto-reconnect dan lifecycle secara internal
  const { eventSource, status, data, close } = useEventSource(url, ['status_update'], {
    autoReconnect: true,
  })

  function connect(onStatusUpdate: (id: string, status: string, name?: string) => void) {
    url.value = '/api/workflows/status'
    
    // Perhatikan: status_update ditangkap lewat watcher data
    watch(data, (newVal) => {
      if (!newVal) return
      try {
        const parsed = JSON.parse(newVal)
        onStatusUpdate(parsed.id, parsed.status, parsed.name)
      } catch (err) {
        console.error('Failed to parse SSE status_update:', err)
      }
    })
  }

  function disconnect() {
    url.value = undefined
    close()
  }

  return {
    eventSource,
    connected: computed(() => status.value === 'OPEN'),
    connect,
    disconnect
  }
}

import { computed } from 'vue'
