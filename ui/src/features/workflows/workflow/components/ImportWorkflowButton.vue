<template>
  <el-button dashed @click="triggerImport" :loading="loading">Import</el-button>
  <input ref="fileInputRef" type="file" accept=".json" hidden @change="handleFileSelect">
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { workflowApi as api } from '@workflows/workflow/api'
import { useWorkflowStore } from '@workflows/workflow/stores/workflow'
import { sanitizeWorkflowPayload } from '@workflows/workflow/utils/workflowMappingUtils'

const emit = defineEmits<{
  success: []
}>()

const store = useWorkflowStore()
const loading = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)

function triggerImport() {
  fileInputRef.value?.click()
}

async function handleFileSelect(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  try {
    loading.value = true
    const text = await file.text()
    const workflow: any = sanitizeWorkflowPayload(JSON.parse(text))
    delete workflow.id
    delete workflow.created_at
    delete workflow.updated_at
    delete workflow.last_run_at
    workflow.status = 'idle'
    
    await api.createWorkflow(workflow)
    await store.fetchWorkflows()
    
    ElMessage.success('Imported successfully')
    emit('success')
  } catch (err: any) {
    ElMessage.error('Import failed: ' + err.message)
  } finally {
    loading.value = false
    if (fileInputRef.value) {
      fileInputRef.value.value = ''
    }
  }
}
</script>
