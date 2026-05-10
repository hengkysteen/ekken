<template>
  <AppPage title="Workflows" subtitle="Manage and monitor your automated processes." scrollable>
    <template #header-extra>
      <el-space>
        <ImportWorkflowButton />
        <el-button type="primary" :icon="Plus" @click="showNewDialog = true">
          Create
        </el-button>
      </el-space>
    </template>
    <div>
      <div v-if="store.workflows.length > 0">
        <el-space direction="vertical" fill style="width: 100%;" :size="12">
          <ListTile v-for="row in store.workflows" :key="row.id" clickable @click="router.push(`/workflow/${row.id}`)">
            <template #leading>
              <el-avatar shape="square" :size="40"
                style="background-color: var(--el-color-primary-light-9); color: var(--el-color-primary); border-radius: 12px;">
                <el-icon :size="20">
                  <CopyDocument />
                </el-icon>
              </el-avatar>
            </template>
            <template #title>
              <el-text size="default" strong>{{ truncate(row.name, 30) }}</el-text>
            </template>
            <template #subtitle>
              <el-space :size="14">
                <el-text type="info" size="small">
                  <el-icon>
                    <PriceTag />
                  </el-icon>
                  {{ row.id }}
                </el-text>
                <el-text :type="store.getStatus(row.id) === 'running' ? 'primary' : 'info'" size="small">
                  <el-icon :class="{ 'is-loading': store.getStatus(row.id) === 'running' }">
                    <Loading v-if="store.getStatus(row.id) === 'running'" />
                    <VideoPlay v-else />
                  </el-icon>
                  {{ store.getStatus(row.id) }}
                </el-text>
                <el-text type="info" size="small">
                  <el-icon>
                    <Timer />
                  </el-icon>
                  {{ formatLastRun(row.last_run_at) }}
                </el-text>
              </el-space>
            </template>
            <template #trailing>
              <ExportWorkflowButton :workflow-id="row.id" :workflow-name="row.name" />
              <el-button dashed @click.stop="confirmDelete(row.id, row.name)">Delete</el-button>
            </template>
          </ListTile>
        </el-space>
      </div>
      <el-empty :image-size="120" v-else-if="store.initialized && !store.loading" description="No workflows yet">

      </el-empty>
    </div>
    <!-- Create Dialog -->
    <el-dialog v-model="showNewDialog" title="Create Workflow" width="400px" align-center>
      <el-form label-position="top" @submit.prevent="createWorkflow">
        <el-form-item>
          <el-input v-model="newName" placeholder="Enter workflow name..." autofocus />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showNewDialog = false">Cancel</el-button>
        <el-button type="primary" @click="createWorkflow" :disabled="!newName.trim()">Create</el-button>
      </template>
    </el-dialog>
  </AppPage>
</template>


<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'

// Stores & Composables
import { useWorkflowStore } from '@workflows/workflow/stores/workflow'

// UI Components & Icons
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Timer, VideoPlay, CopyDocument, PriceTag, Loading } from '@element-plus/icons-vue'
import AppPage from '@shared/components/AppPage.vue'
import ListTile from '@shared/components/ListTile.vue'

// Workflow Components
import ImportWorkflowButton from '@workflows/workflow/components/ImportWorkflowButton.vue'
import ExportWorkflowButton from '@workflows/workflow/components/ExportWorkflowButton.vue'

// Utils
import { truncate } from '@/shared/utils/string'
import { setTitle } from '@shared/utils/titleRegistry'

const router = useRouter()
const store = useWorkflowStore()

// --- UI State ---
const showNewDialog = ref(false)
const newName = ref('')

// --- Lifecycle & Watchers ---

// Auto-register names to global registry for flicker-free breadcrumbs
watch(() => store.workflows, (list) => {
  list.forEach(w => setTitle(String(w.id), w.name))
}, { immediate: true })

onMounted(() => {
  store.fetchWorkflows()
})

// --- Actions ---

async function createWorkflow() {
  if (!newName.value.trim()) return
  try {
    const id = await store.createWorkflow(newName.value)
    showNewDialog.value = false
    newName.value = ''
    router.push(`/workflow/${id}`)
  } catch (err) {
    ElMessage.error('Failed to create workflow')
  }
}

function confirmDelete(id: string, name: string) {
  ElMessageBox.confirm(
    `Delete workflow "${name}"? This cannot be undone.`,
    'Confirm Deletion',
    {
      confirmButtonText: 'Delete',
      cancelButtonText: 'Cancel',
      type: 'error',
      buttonSize: 'default'
    }
  ).then(() => deleteWorkflow(id)).catch(() => { })
}

async function deleteWorkflow(id: string) {
  try {
    await store.deleteWorkflow(id)
    ElMessage.success('Workflow deleted')
  } catch (err) {
    ElMessage.error('Failed to delete workflow')
  }
}

// --- Formatters ---

function formatLastRun(isoString?: string): string {
  if (!isoString) return 'Never'
  const date = new Date(isoString)
  const diff = Math.floor((new Date().getTime() - date.getTime()) / 60000)

  if (diff < 1) return 'Just now'
  if (diff < 60) return `${diff}m ago`
  return `${Math.floor(diff / 60)}h ago`
}
</script>