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
        class="hub-notice"
      />

      <el-alert
        v-if="error"
        :title="error"
        type="error"
        :closable="false"
        show-icon
        class="hub-notice"
      />

      <div v-loading="loading" class="hub-content">
        <el-empty
          v-if="initialized && !loading && plugins.length === 0"
          description="No plugins available in the registry."
        />

        <div v-else class="plugin-grid">
          <section v-for="item in plugins" :key="item.id" class="plugin-card">
            <header class="card-header">
              <div class="plugin-icon">
                <img v-if="kindIcon(item)" :src="kindIcon(item)" :alt="`${item.name} icon`" />
                <el-icon v-else><Connection /></el-icon>
              </div>
              <div class="plugin-heading">
                <div class="title-row">
                  <h3>{{ item.name }}</h3>
                  <el-tag size="small" effect="plain">{{ item.kind }}</el-tag>
                </div>
                <div class="meta-row">
                  <span>v{{ item.version || '0.0.0' }}</span>
                  <span v-if="item.repo?.author">By {{ item.repo.author }}</span>
                </div>
              </div>
            </header>

            <p class="description">{{ item.description || 'No description' }}</p>

            <div v-if="!hasCompatibleArtifact(item) || shouldShowStatus(item)" class="info-strip">
              <div v-if="!hasCompatibleArtifact(item)">
                <span class="info-label">Compatibility</span>
                <el-tag type="warning" effect="plain">Not available for this device</el-tag>
              </div>
              <div v-if="shouldShowStatus(item)">
                <span class="info-label">Status</span>
                <el-tag :type="statusTag(item)" effect="plain">{{ statusLabel(item) }}</el-tag>
              </div>
            </div>

            <div v-if="kindTags(item).length || kindActions(item).length" class="kind-meta">
              <div v-if="kindTags(item).length" class="kind-tags">
                <el-tag v-for="tag in kindTags(item)" :key="tag" size="small" effect="plain">
                  {{ tag }}
                </el-tag>
              </div>
              <div v-if="kindActions(item).length" class="kind-actions">
                <span v-for="action in kindActions(item)" :key="action.key" class="kind-action">
                  <strong>{{ action.key }}</strong>
                  <span>{{ action.description }}</span>
                </span>
              </div>
            </div>

            <div v-if="installTasks[item.id]" class="progress-block">
              <el-progress
                :percentage="progressPercent(installTasks[item.id])"
                :status="progressStatus(installTasks[item.id])"
                :stroke-width="8"
              />
              <span v-if="installTasks[item.id]?.error" class="install-error">
                {{ installTasks[item.id].error }}
              </span>
            </div>

            <footer class="card-actions">
              <el-button
                v-if="isInstalling(item.id)"
                type="warning"
                plain
                @click="stopInstall(item.id)"
              >
                Stop
              </el-button>
              <el-button
                v-else-if="item.is_installed"
                type="danger"
                plain
                @click="uninstallPlugin(item.id)"
              >
                Uninstall
              </el-button>
              <el-button
                v-else
                type="primary"
                :disabled="!hasCompatibleArtifact(item)"
                @click="installPlugin(item.id)"
              >
                Install
              </el-button>
              <el-button
                v-if="item.repo?.url"
                :icon="Link"
                circle
                @click="openExternal(item.repo.url)"
              />
            </footer>
          </section>
        </div>
      </div>
    </div>
  </AppPage>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { Connection, Link } from '@element-plus/icons-vue'
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

function kindTags(row: RegistryPluginSummary) {
  const tags = row.kind_meta?.tags
  return Array.isArray(tags) ? tags.filter((tag): tag is string => typeof tag === 'string') : []
}

function kindIcon(row: RegistryPluginSummary) {
  const icon = row.kind_meta?.icon
  return typeof icon === 'string' && icon.trim() ? icon : ''
}

function kindActions(row: RegistryPluginSummary) {
  const actions = row.kind_meta?.actions
  if (!Array.isArray(actions)) return []
  return actions
    .filter((action): action is { key: string; description?: string } => {
      return typeof action === 'object' && action !== null && typeof (action as any).key === 'string'
    })
    .map(action => ({
      key: action.key,
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
.hub-container {
  padding: 4px 0 24px;
}

.hub-notice {
  margin-bottom: 16px;
}

.hub-content {
  min-height: 260px;
}

.plugin-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.plugin-card {
  display: flex;
  flex-direction: column;
  min-height: 280px;
  padding: 16px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  background: var(--el-bg-color);
}

.card-header {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.plugin-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 40px;
  width: 40px;
  height: 40px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  color: var(--el-color-primary);
  background: var(--el-fill-color-lighter);
}

.plugin-icon img {
  width: 24px;
  height: 24px;
  object-fit: contain;
}

.plugin-heading {
  min-width: 0;
  flex: 1;
}

.title-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.title-row h3 {
  margin: 0;
  font-size: 16px;
  line-height: 1.3;
  color: var(--el-text-color-primary);
}

.meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.description {
  display: -webkit-box;
  min-height: 62px;
  margin: 14px 0;
  overflow: hidden;
  color: var(--el-text-color-regular);
  line-height: 1.45;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}

.info-strip {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 10px;
  margin-top: auto;
  padding: 10px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background: var(--el-fill-color-lighter);
}

.info-strip > div {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.info-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.info-strip strong {
  overflow: hidden;
  color: var(--el-text-color-primary);
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.kind-meta {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 12px;
}

.kind-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.kind-actions {
  display: grid;
  gap: 6px;
}

.kind-action {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.kind-action strong {
  color: var(--el-text-color-primary);
}

.progress-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 12px;
}

.install-error {
  color: var(--el-color-danger);
  font-size: 12px;
  line-height: 1.3;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 14px;
}
</style>
