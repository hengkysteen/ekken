<template>
  <AppPage title="Plugin Hub" subtitle="Browse registry plugins and manage remote installs." scrollable>
    <template #header-extra>
      <el-button @click="$router.push({ name: 'plugins' })">Back to Plugins</el-button>
    </template>

    <div class="hub-container">
      <el-alert
        v-if="registry?.message"
        :title="registry.message"
        type="info"
        :closable="false"
        show-icon
        class="mb-6"
      />

      <el-alert
        v-if="error"
        :title="error"
        type="error"
        :closable="false"
        show-icon
        class="mb-6"
      />

      <div v-loading="loading" class="hub-content">
        <el-empty
          v-if="initialized && !loading && plugins.length === 0"
          description="No plugins available in the registry."
        />

        <el-row v-else :gutter="24">
          <el-col
            v-for="item in plugins"
            :key="item.id"
            :xs="24" :sm="12" :md="12" :lg="8" :xl="6"
            class="mb-6"
          >
            <el-card
              shadow="never"
              class="plugin-card"
              :body-style="{ padding: '20px', display: 'flex', flexDirection: 'column', height: '100%' }"
            >
              <div class="card-inner flex-1 flex flex-column">
                <!-- Header: Stable Layout -->
                <div class="flex gap-4 mb-5">
                  <el-avatar :size="48" :src="kindIcon(item)" shape="square" class="shrink-0 border-lighter">
                    <el-icon><Connection /></el-icon>
                  </el-avatar>
                  <div class="flex-1 min-w-0">
                    <div class="flex justify-between items-start">
                      <el-tooltip :content="item.name" placement="top" :disabled="item.name.length < 20">
                        <el-text strong size="large" truncated class="mr-2">{{ item.name }}</el-text>
                      </el-tooltip>
                      <div class="flex items-center gap-2 shrink-0">
                        <el-tag size="small" type="info" effect="plain" disable-transitions>{{ item.kind }}</el-tag>
                        <el-button v-if="item.repo?.url" :icon="Link" link class="p-0 h-auto" @click="openExternal(item.repo.url)" />
                      </div>
                    </div>
                    <div class="flex items-center gap-1 mt-1">
                      <el-text size="small" type="secondary">v{{ item.version || '0.0.0' }}</el-text>
                      <el-divider v-if="item.repo?.author" direction="vertical" class="mx-1" />
                      <el-text v-if="item.repo?.author" size="small" type="secondary" truncated>{{ item.repo.author }}</el-text>
                    </div>
                  </div>
                </div>

                <!-- Description Area -->
                <div class="flex-1">
                  <el-text
                    class="mb-4 description-text"
                    :line-clamp="2"
                    :type="item.description ? '' : 'placeholder'"
                  >
                    {{ item.description || 'No description provided for this plugin.' }}
                  </el-text>

                  <!-- Meta/Actions -->
                  <div class="mb-6">
                    <div v-if="kindActions(item).length" class="flex flex-column gap-1">
                      <div v-for="action in kindActions(item)" :key="action.type">
                        <el-text size="small" truncated class="block">
                          <b class="mr-1">{{ action.type }}</b>
                          <span class="opacity-70">{{ action.description }}</span>
                        </el-text>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Footer Status & Progress -->
                <div class="pt-4 border-t border-lighter mt-auto">
                  <div v-if="!hasCompatibleArtifact(item)" class="mb-3">
                    <el-text type="warning" size="small" class="flex items-center gap-1">
                      <el-icon><Warning /></el-icon>
                      Incompatible device
                    </el-text>
                  </div>

                  <div v-if="shouldShowStatus(item)" class="flex items-center justify-between mb-2">
                    <el-text size="small" type="secondary" class="uppercase tracking-wider font-bold">Status</el-text>
                    <el-tag :type="statusTag(item)" effect="dark" size="small" disable-transitions>
                      {{ statusLabel(item) }}
                    </el-tag>
                  </div>

                  <div v-if="installTasks[item.id]" class="mt-3">
                    <el-progress
                      :percentage="progressPercent(installTasks[item.id])"
                      :status="progressStatus(installTasks[item.id])"
                      :stroke-width="4"
                      :show-text="false"
                    />
                    <el-text v-if="installTasks[item.id]?.error" type="danger" size="small" class="mt-1 block">
                      {{ installTasks[item.id].error }}
                    </el-text>
                  </div>

                  <div class="mt-4">
                    <el-button
                      v-if="isInstalling(item.id)"
                      type="warning"
                      class="w-full"
                      @click="stopInstall(item.id)"
                    >
                      Stop
                    </el-button>
                    <el-button
                      v-else-if="item.is_installed"
                      type="danger"
                      plain
                      class="w-full"
                      @click="uninstallPlugin(item.id)"
                    >
                      Uninstall
                    </el-button>
                    <el-button
                      v-else
                      type="primary"
                      class="w-full"
                      :disabled="!hasCompatibleArtifact(item)"
                      @click="installPlugin(item.id)"
                    >
                      Install
                    </el-button>
                  </div>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </div>
  </AppPage>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { Connection, Link, Warning } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { pluginsApi as api } from '@plugins/api'
import type { InstallTask, RegistryPluginSummary, RegistryResponse } from '@plugins/api'
import { useAppSettingsStore } from '@shared/stores/useAppSettingsStore'
import AppPage from '@shared/components/AppPage.vue'

const settingsStore = useAppSettingsStore()
const loading = ref(false)
const initialized = ref(false)
const error = ref('')
const registry = ref<RegistryResponse | null>(null)
const installTasks = ref<Record<string, InstallTask>>({})
let pollTimer: number | undefined

const plugins = computed(() => registry.value?.plugins ?? [])

async function fetchRegistry() {
  loading.value = true
  error.value = ''
  try {
    registry.value = await api.getRegistry()
  } catch (err: any) {
    error.value = err.message || 'Failed to load registry'
  } finally {
    loading.value = false
    initialized.value = true
  }
}

async function installPlugin(id: string) {
  try {
    const task = await api.installPlugin(id)
    installTasks.value = { ...installTasks.value, [id]: task }
    ElMessage.success(`Install started: ${id}`)
    startPolling()
  } catch (err: any) {
    ElMessage.error(err.message || 'Failed to start install')
  }
}

async function stopInstall(id: string) {
  try {
    const task = await api.stopInstall(id)
    installTasks.value = { ...installTasks.value, [id]: task }
    ElMessage.success(`Install stopped: ${id}`)
    await fetchRegistry()
  } catch (err: any) {
    ElMessage.error(err.message || 'Failed to stop install')
  }
}

async function uninstallPlugin(id: string) {
  try {
    await ElMessageBox.confirm('Uninstall this plugin?', 'Uninstall', { type: 'warning' })
    await api.uninstallPlugin(id)
    ElMessage.success(`Plugin uninstalled: ${id}`)
    await fetchRegistry()
  } catch (err: any) {
    if (err === 'cancel') return
    ElMessage.error(err.message || 'Failed to uninstall plugin')
  }
}

async function pollInstallTasks() {
  const ids = Object.keys(installTasks.value).filter(id => isInstalling(id))
  if (ids.length === 0) {
    stopPolling()
    return
  }

  await Promise.all(ids.map(async id => {
    try {
      const task = await api.getInstallStatus(id)
      installTasks.value = { ...installTasks.value, [id]: task }
      if (task.status === 'completed') {
        ElMessage.success(`Plugin installed: ${id}`)
        await fetchRegistry()
      } else if (task.status === 'failed') {
        ElMessage.error(task.error || `Install failed: ${id}`)
      }
    } catch {
      stopPolling()
    }
  }))
}

function startPolling() {
  if (pollTimer) return
  pollTimer = window.setInterval(() => {
    pollInstallTasks()
  }, 1000)
}

function stopPolling() {
  if (!pollTimer) return
  window.clearInterval(pollTimer)
  pollTimer = undefined
}

function isInstalling(id: string) {
  const status = installTasks.value[id]?.status
  return status === 'queued'
    || status === 'downloading'
    || status === 'verifying'
    || status === 'extracting'
    || status === 'installing'
}

function hasCompatibleArtifact(row: RegistryPluginSummary) {
  const device = settingsStore.deviceInfo
  if (!row.artifacts?.length || !device.os || !device.arch) return false
  return row.artifacts.some(artifact => {
    const osMatch = artifact.os === 'any' || artifact.os === device.os
    const archMatch = artifact.arch === 'any' || artifact.arch === device.arch
    return osMatch && archMatch
  })
}

function statusLabel(row: RegistryPluginSummary) {
  const task = installTasks.value[row.id]
  if (task) return task.status
  if (row.is_installed) return row.is_enabled ? 'enabled' : row.status || 'installed'
  return ''
}

function statusTag(row: RegistryPluginSummary) {
  const task = installTasks.value[row.id]
  if (task?.status === 'failed') return 'danger'
  if (task?.status === 'canceled') return 'warning'
  if (task && isInstalling(row.id)) return 'warning'
  if (row.is_installed) return 'success'
  return 'info'
}

function progressPercent(task?: InstallTask) {
  if (!task) return 0
  if (task.status === 'completed') return 100
  return Math.round((task.progress || 0) * 100)
}

function progressStatus(task?: InstallTask) {
  if (task?.status === 'failed') return 'exception'
  if (task?.status === 'completed') return 'success'
  if (task?.status === 'canceled') return 'warning'
  return undefined
}

function shouldShowStatus(row: RegistryPluginSummary) {
  return Boolean(installTasks.value[row.id] || row.is_installed)
}

function kindIcon(row: RegistryPluginSummary) {
  const icon = row.kind_meta?.icon
  return typeof icon === 'string' && icon.trim() ? icon : ''
}

function kindActions(row: RegistryPluginSummary) {
  const actions = row.kind_meta?.actions
  if (!Array.isArray(actions)) return []
  return actions
    .filter((action): action is { type: string; description?: string } => {
      return typeof action === 'object' && action !== null && typeof (action as any).type === 'string'
    })
    .map(action => ({
      type: action.type,
      description: typeof action.description === 'string' ? action.description : '',
    }))
}

function openExternal(url: string) {
  window.open(url, '_blank', 'noopener,noreferrer')
}

onMounted(() => {
  if (!settingsStore.deviceInfo.os || !settingsStore.deviceInfo.arch) {
    settingsStore.startPolling()
  }
  fetchRegistry()
})

onBeforeUnmount(() => {
  stopPolling()
})
</script>

<style scoped>
 

.hub-content {
  min-height: 300px;
}

.plugin-card {
  height: 100%;
}

.card-inner {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.description-text {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.5;
  height: 3em;
}

.border-lighter {
  border-color: var(--el-border-color-lighter);
}

/* Utility classes to avoid heavy CSS */
.w-full { width: 100%; }
.flex { display: flex; }
.flex-column { flex-direction: column; }
.flex-1 { flex: 1; }
.flex-wrap { flex-wrap: wrap; }
.shrink-0 { flex-shrink: 0; }
.items-center { align-items: center; }
.items-start { align-items: flex-start; }
.justify-between { justify-content: space-between; }
.min-w-0 { min-width: 0; }
.gap-1 { gap: 4px; }
.gap-2 { gap: 8px; }
.gap-4 { gap: 16px; }
.mb-1 { margin-bottom: 4px; }
.mb-2 { margin-bottom: 8px; }
.mb-3 { margin-bottom: 12px; }
.mb-4 { margin-bottom: 16px; }
.mb-5 { margin-bottom: 20px; }
.mb-6 { margin-bottom: 24px; }
.mr-1 { margin-right: 4px; }
.mr-2 { margin-right: 8px; }
.mx-1 { margin-left: 4px; margin-right: 4px; }
.mt-1 { margin-top: 4px; }
.mt-3 { margin-top: 12px; }
.mt-4 { margin-top: 16px; }
.mt-auto { margin-top: auto; }
.pt-4 { padding-top: 16px; }
.p-0 { padding: 0; }
.h-auto { height: auto; }
.block { display: block; }
.border-t { border-top-width: 1px; border-top-style: solid; }
.uppercase { text-transform: uppercase; }
.tracking-wider { letter-spacing: 0.05em; }
.font-bold { font-weight: bold; }
.opacity-70 { opacity: 0.7; }
</style>



