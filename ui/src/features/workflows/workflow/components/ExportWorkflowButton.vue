<template>
  <el-button dashed @click.stop="handleExport" :loading="loading">Export</el-button>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { workflowApi as api } from '@workflows/workflow/api'
import { sanitizeWorkflowPayload } from '@workflows/workflow/utils/workflowMappingUtils'

const props = defineProps<{
  workflowId: string
  workflowName: string
  data?: any
}>()

const loading = ref(false)

async function handleExport() {
  try {
    loading.value = true
    let exportData = props.data

    if (!exportData) {
      exportData = await api.getWorkflow(props.workflowId)
    }
    exportData = sanitizeWorkflowPayload(exportData)

    const json = JSON.stringify(exportData, null, 2)
    const blob = new Blob([json], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${props.workflowName.replace(/[^a-zA-Z0-9_-]/g, '_')}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (err: any) {
    ElMessage.error('Export Error: ' + err.message)
  } finally {
    loading.value = false
  }
}
</script>
