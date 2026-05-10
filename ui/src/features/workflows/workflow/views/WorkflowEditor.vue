<template>
  <AppPage :scrollable="false" :no-padding="true" header-padding="14px 24px">

    <template #title>
      <el-input v-model="workflowName" placeholder="Workflow Name" />
    </template>

    <template #header-extra>
      <el-space :size="18">
        <el-button dashed :icon="DocumentChecked" @click="handleSave">Save</el-button>
        <el-button dashed :icon="Delete" @click="confirmClear">Clear</el-button>
        <el-divider direction="vertical" />
        <el-button :icon="Monitor" @click="showLogsDrawer = true">Logs</el-button>
        <el-button v-if="workflowStore.getStatus(workflowId) !== 'running'" :icon="VideoPlay" type="primary"
          :disabled="isSubmitting || !hasTrigger" @click="handleRun">Run</el-button>
        <el-button v-if="workflowStore.getStatus(workflowId) === 'running'" :icon="VideoPause" type="danger"
          :disabled="isSubmitting" @click="handleStop">Stop</el-button>
      </el-space>
    </template>

    <div class="editor-canvas-wrapper">
      <el-container class="canvas-layout-container">
        <el-main class="canvas-main-area">

          <FlowCanvas ref="canvasRef" :workflow-id="workflowId" />

          <el-button circle size="large" class="canvas-sidebar-toggle"
            @click="isSidebarCollapsed = !isSidebarCollapsed">
            <el-icon>
              <ArrowLeft v-if="isSidebarCollapsed" />
              <ArrowRight v-else />
            </el-icon>
          </el-button>
        </el-main>

        <el-aside class="sidebar-transition" :width="isSidebarCollapsed ? '0' : '230px'">
          <EditorSidebar v-show="!isSidebarCollapsed"
            :workflow-nodes="editor.flowNodes" :mynodes-items="mynodeStore.items"
            :mynodes-tab-key="editor.mynodesTabKey" @add-node="onAddFromSidebar"
            @delete-mynodes-item="onDeleteMyNodesItem" />
        </el-aside>
      </el-container>
    </div>

    <!-- Modals & Drawers -->
    <WorkflowLogsDrawer v-model:show="showLogsDrawer" :logs="logs" @update:logs="setLogs" :workflow-id="workflowId"
      :workflow-name="workflowName" />

  </AppPage>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, onBeforeRouteLeave } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useWorkflowStore } from '@workflows/workflow/stores/workflow'
import { useWorkflowEditorStore } from '@workflows/workflow/stores/workflowEditor'
import { useMyNodeStore } from '@workflows/mynode/stores/mynode'
import { useNodeStore } from '@workflows/node/stores/node'
import { useWorkflowRunner } from '@workflows/workflow/composables/useWorkflowRunner'
import EditorSidebar from '@workflows/workflow/components/EditorSidebar.vue'
import FlowCanvas from '@workflows/workflow/components/FlowCanvas.vue'
import {
  Delete, DocumentChecked, VideoPlay, VideoPause, ArrowRight, ArrowLeft, Monitor
} from '@element-plus/icons-vue'
import { workflowApi as api } from '@workflows/workflow/api'
import WorkflowLogsDrawer from '@workflows/workflow/components/WorkflowLogsDrawer.vue'
import { setTitle } from '@shared/utils/titleRegistry'
import AppPage from '@shared/components/AppPage.vue'

const route = useRoute()
const editor = useWorkflowEditorStore()
const mynodeStore = useMyNodeStore()
const workflowStore = useWorkflowStore()
const nodeStore = useNodeStore()

const workflowId = ref((route.params.id || '') as string)
const originalId = ref((route.params.id || '') as string)
const workflowName = ref('')
const originalName = ref('')

const isSidebarCollapsed = computed({
  get: () => !editor.showSidebar,
  set: (val) => { editor.showSidebar = !val }
})

const canvasRef = ref<any>(null)
const { isRunning, isSubmitting, syncStatus, connectSSE, logs, setLogs, handleRun: runnerHandleRun, handleStop: runnerHandleStop } = useWorkflowRunner(workflowId as any)
const showLogsDrawer = ref(false)

const hasTrigger = computed(() => {
  return editor.flowNodes.some(node => node.data?.tags?.includes('Trigger'))
})

function onAddFromSidebar({ type, label, tags, icon, position, config, response_var, sourceType, name }: any) {
  const pos = position || (() => {
    const center = canvasRef.value?.viewportCenter || { x: 100, y: 200 }
    return { x: center.x + Math.random() * 50, y: center.y + Math.random() * 50 }
  })()
  editor.addNode({ type, label, tags, icon, position: pos, config, response_var, sourceType, name })
}

function onDeleteMyNodesItem(id: string) {
  ElMessageBox.confirm('Delete this item?', 'Delete', { type: 'warning' }).then(async () => {
    try {
      await mynodeStore.deleteItem(id)
      ElMessage.success('Deleted')
    } catch (err: any) {
      ElMessage.error('Failed: ' + err.message)
    }
  }).catch(() => { })
}

function confirmClear() {
  ElMessageBox.confirm('Clear workflow?', 'Clear').then(() => {
    editor.clearWorkflow()
    ElMessage.success('Cleared')
  }).catch(() => { })
}

async function handleSave() {
  try {
    const isRename = originalName.value !== workflowName.value
    if (isRename) {
      if (editor.workflow) editor.workflow.name = workflowName.value
      await editor.saveWorkflow(originalId.value)
      originalName.value = workflowName.value
      await workflowStore.fetchWorkflows()
    } else {
      await editor.saveWorkflow(originalId.value)
    }
    ElMessage.success('Saved')
  } catch (err: any) {
    ElMessage.error({
      message: '<strong>Save Error:</strong><br/>' + err.message.replace(/\n/g, '<br/>'),
      dangerouslyUseHTMLString: true,
      duration: 5000
    })
  }
}

async function handleRun() {
  await runnerHandleRun(async () => {
    await editor.saveWorkflowSilent(originalId.value)
  })
}

async function handleStop() {
  await runnerHandleStop()
}

onMounted(async () => {
  try {
    await nodeStore.loadCatalog()
    await editor.loadWorkflow(workflowId.value)
    await mynodeStore.loadItems()
    
    workflowName.value = editor.workflow?.name || ''
    originalName.value = editor.workflow?.name || ''

    // Load initial logs
    const initialLogs = await api.getWorkflowLogs(workflowId.value)
    setLogs(initialLogs || [])
  } catch (err) { }
  
  await syncStatus()
  if (isRunning.value) connectSSE()
})

onBeforeUnmount(() => {
  editor.resetState()
})

onBeforeRouteLeave(async (to, _from, next) => {
  if (to.name !== null) await editor.saveWorkflowSilent(originalId.value)
  next()
})

watch(workflowName, (val) => {
  if (val) setTitle(workflowId.value, val)
}, { immediate: true })
</script>

<style scoped>
.editor-canvas-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}

.canvas-layout-container {
  height: 100%;
}

.canvas-main-area {
  position: relative;
  padding: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.canvas-sidebar-toggle {
  position: absolute;
  top: 50%;
  right: 24px;
  transform: translateY(-50%);
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
</style>
