<template>
  <el-config-provider>
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useWorkflowStore } from '@workflows/workflow/stores/workflow'
import { useTheme } from './shared/composables/useTheme'

const workflowStore = useWorkflowStore()
useTheme() // Inisialisasi tema global

const handleUnload = () => {
  workflowStore.disconnectSSE()
}

onMounted(() => {
  // Connect global SSE
  workflowStore.connectSSE()

  window.addEventListener('pagehide', handleUnload)
})

onUnmounted(() => {
  window.removeEventListener('pagehide', handleUnload)
})
</script>