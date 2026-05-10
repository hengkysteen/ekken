<template>
  <el-drawer v-model="isOpen" direction="btt" size="600" destroy-on-close resizable :show-close="false"
    class="log-drawer">
    <template #header="{ close }">
      <div style="display: flex; align-items: center; justify-content: space-between; width: 100%;">
        <div style="display: flex; align-items: center; gap: 8px;">
          <el-icon :size="16" style="display: flex;">
            <Monitor />
          </el-icon>
          <span style="font-weight: 600;">{{ workflowName ? `${workflowName} Logs` : 'Workflow Logs' }}</span>
        </div>
        <div v-loading="loading" style="display: flex; align-items: center; gap: 8px;">
          <el-button link @click="copyAllLogs" :disabled="loading || !logs?.length">
            Copy
          </el-button>
          <el-button link @click="clearLogs" :disabled="loading">
            Clear
          </el-button>
          <el-button link @click="deleteLogs" :disabled="loading">
            Delete
          </el-button>
          <el-button link @click="close" style="margin-left: 8px;">
            <el-icon :size="20">
              <Close />
            </el-icon>
          </el-button>
        </div>
      </div>
    </template>
    <el-scrollbar ref="scrollbarRef" class="log-scrollbar">
      <div v-loading="loading" style="padding: 20px; min-height: 100%;">
        <div v-if="!loading && (!logs || logs.length === 0)"
          style="display: flex; align-items: center; justify-content: center; min-height: 400px; width: 100%;">
          <el-empty description="Empty" :image-size="80">
            <template #image>
              <el-icon color="var(--el-text-color-placeholder)" :size="64">
                <Monitor />
              </el-icon>
            </template>
          </el-empty>
        </div>
        <div v-else-if="!loading" class="log-container">
          <div v-for="(entry, index) in logs" :key="index" class="log-entry">
            <!-- Normal Text Log -->
            <div v-if="!entry.raw" class="text-log">
              <span class="log-time">{{ formatTime(entry.time) }}</span>
              <span class="log-level" :style="{ color: getLevelColor(entry.level) }">[{{ (entry.level ||
                'INFO').toUpperCase() }}]</span>
              <span class="log-msg">{{ entry.message }}</span>
            </div>

            <!-- Expandable Response Log -->
            <div v-else class="response-log-manual">
              <div class="collapse-header text-log" @click="toggleExpand(index)">
                <span class="log-time">{{ formatTime(entry.time) }}</span>
                <span class="log-level" :style="{ color: getLevelColor(entry.level) }">[{{ (entry.level ||
                  'INFO').toUpperCase() }}]</span>
                <span class="log-msg">{{ entry.message }}</span>
                <el-icon class="expand-icon" :class="{ 'is-active': expandedIndex === index }">
                  <ArrowRight />
                </el-icon>
              </div>
              <div v-if="expandedIndex === index" class="collapse-content-manual">
                <div class="action-bar">
                  <el-button :icon="CopyDocument" size="small" link @click.stop="copyRaw(entry.raw)">
                    Copy
                  </el-button>
                </div>
                <pre class="json-code"><code>{{ formatRaw(entry.raw) }}</code></pre>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-scrollbar>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed, ref, nextTick, watch } from 'vue'
import { Monitor, CopyDocument, Close, ArrowRight } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { workflowApi as api } from '@workflows/workflow/api'

export interface LogEntry {
  time: string
  level: string
  message: string
  raw?: string
}

const props = defineProps<{
  show: boolean
  workflowId: string
  workflowName?: string
  logs: LogEntry[]
}>()

const emit = defineEmits<{
  (e: 'update:show', val: boolean): void
  (e: 'update:logs', val: LogEntry[]): void
}>()

const loading = ref(false)
const scrollbarRef = ref()
const expandedIndex = ref<number | null>(null)

function toggleExpand(index: number) {
  if (expandedIndex.value === index) {
    expandedIndex.value = null
  } else {
    expandedIndex.value = index
  }
}

const isOpen = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val)
})



async function deleteLogs() {
  try {
    await ElMessageBox.confirm('Are you sure you want to delete all logs?', 'Delete Logs', { type: 'warning' })
    await api.deleteWorkflowLogs(props.workflowId)
    emit('update:logs', [])
    ElMessage.success('Logs deleted')
  } catch (err: any) { }
}

async function fetchLogs() {
  if (!props.workflowId) return
  loading.value = true
  try {
    // Fetch historical logs from persistence (file).
    // This is required when the drawer is opened to retrieve past logs,
    // or to restore the view after the UI has been cleared (as long as they are not deleted from the server),
    // because SSE only broadcasts real-time data during an active connection.
    const data = await api.getWorkflowLogs(props.workflowId)
    emit('update:logs', data || [])
  } catch (err) {
    console.error('Failed to fetch logs:', err)
  } finally {
    loading.value = false
    scrollToBottom()
  }
}

function clearLogs() {
  emit('update:logs', [])
}

function copyAllLogs() {
  if (!props.logs) return
  const content = props.logs.map(entry => {
    const time = formatTime(entry.time)
    const level = (entry.level || 'INFO').toUpperCase()
    const message = entry.message || ''
    let logLine = `${time} [${level}] ${message}`

    // Include raw JSON data if available
    if (entry.raw) {
      logLine += `\n${formatRaw(entry.raw)}`
    }
    return logLine
  }).join('\n')

  navigator.clipboard.writeText(content)
  ElMessage.success('All logs (including responses) copied')
}

function scrollToBottom() {
  nextTick(() => {
    if (scrollbarRef.value) {
      const wrapRef = scrollbarRef.value.wrapRef
      if (wrapRef) {
        wrapRef.scrollTop = wrapRef.scrollHeight
      }
    }
  })
}

watch(isOpen, (val) => {
  if (val) {
    fetchLogs()
  }
})

watch(() => props.logs, () => {
  scrollToBottom()
}, { deep: true })

function getLevelColor(level: string): string {
  const l = level.toLowerCase()
  if (l === 'error') return 'var(--el-color-danger)'
  if (l === 'debug') return 'var(--el-color-primary)'
  return 'inherit'
}

function copyRaw(raw: string | undefined) {
  if (!raw) return
  navigator.clipboard.writeText(raw)
  ElMessage.success('Copied to clipboard')
}

function formatTime(timeStr: string) {
  try {
    const date = new Date(timeStr)
    return date.toLocaleTimeString('en-GB', {
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      fractionalSecondDigits: 3
    })
  } catch (e) {
    return timeStr
  }
}

function formatRaw(raw: string | undefined) {
  if (!raw) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch (e) {
    return raw
  }
}
</script>
<style>
.log-drawer .el-drawer__header {
  margin-bottom: 0 !important;
  padding-bottom: 12px !important;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.log-drawer .el-drawer__body {
  padding: 0;
  overflow: hidden;
}

.log-scrollbar {
  height: 100%;
}

.log-container {
  display: flex;
  flex-direction: column;
}

.text-log {
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  display: flex;
  align-items: center;
}

.log-time {
  width: 110px;
  flex-shrink: 0;
}

.log-level {
  width: 90px;
  flex-shrink: 0;
}

.response-log-manual {
  display: flex;
  flex-direction: column;
}

.collapse-header {
  display: flex;
  align-items: center;
  width: 100%;
  cursor: pointer;
}

.expand-icon {
  margin-left: 4px;
  transition: transform 0.2s;
  font-size: 12px;
  color: var(--el-color-primary);
}

.expand-icon.is-active {
  transform: rotate(90deg);
}

.collapse-content-manual {
  padding: 8px 0 8px 110px;
  position: relative;
}

.action-bar {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 10;
}

.log-msg {
  margin-left: 8px;
}

.json-code {
  background-color: var(--el-bg-color-page);
  padding: 12px;
  border-radius: 4px;
  font-size: 12px;
  overflow-x: auto;
}
</style>
