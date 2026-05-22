<template>
  <el-config-provider>
    <router-view v-if="profileStore.initialized && !profileStore.requiresUnlock" />
    <AppLockScreen v-if="profileStore.requiresUnlock" />
  </el-config-provider>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useWorkflowStore } from '@workflows/workflow/stores/workflow'
import AppLockScreen from '@profile/components/AppLockScreen.vue'
import { useProfileStore } from '@profile/stores/profile'
import { useTheme } from './shared/composables/useTheme'

const workflowStore = useWorkflowStore()
const profileStore = useProfileStore()
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
