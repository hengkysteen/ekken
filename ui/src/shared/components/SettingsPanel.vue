<template>
  <el-container class="settings-panel">
    <el-aside width="180px" class="settings-sidebar">
      <el-menu :default-active="selectedKey" @select="(val: any) => (selectedKey = val)" style="border-right: none">
        <el-menu-item index="general">
          <el-icon>
            <Grid />
          </el-icon>
          <span>General</span>
        </el-menu-item>
        <el-menu-item v-for="tab in dynamicTabs" :key="tab.id" :index="tab.id">
          <el-icon>
            <component :is="tab.icon" v-if="tab.icon" />
            <Files v-else />
          </el-icon>
          <span>{{ tab.label }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-main class="settings-content">
      <div v-if="selectedKey === 'general'">
        <el-space direction="vertical" alignment="start" :size="4" style="margin-bottom: 32px">
          <el-text size="large" strong style="font-size: 20px">General</el-text>
          <el-text type="info">System information and application details</el-text>
        </el-space>

        <div class="settings-section">
          <el-text class="section-title" strong>Resource Monitoring</el-text>
          <el-card shadow="never" class="info-card">
            <div style="display: flex; flex-direction: column; gap: 16px">
              <div style="display: flex; align-items: center; justify-content: space-between">
                <div>
                  <el-text strong>Enable Monitoring</el-text>
                  <div style="margin-top: 4px">
                    <el-text type="info" size="small">Activate system resources tracking</el-text>
                  </div>
                </div>
                <el-switch v-model="settingsStore.enableMonitoring" @change="handleMasterToggle" />
              </div>
              <template v-if="settingsStore.enableMonitoring">
                <div style="display: flex; align-items: center; justify-content: space-between">
                  <div>
                    <el-text strong>Background Polling</el-text>
                    <div style="margin-top: 4px">
                      <el-text type="info" size="small">Fetch system metrics periodically</el-text>
                    </div>
                  </div>
                  <el-switch v-model="settingsStore.enablePolling" @change="handlePollingToggle" />
                </div>
                <div v-if="settingsStore.enablePolling"
                  style="display: flex; align-items: center; justify-content: space-between">
                  <div>
                    <el-text strong>Polling Interval (s)</el-text>
                    <div style="margin-top: 4px">
                      <el-text type="info" size="small">Frequency of updates (min: 1s)</el-text>
                    </div>
                  </div>
                  <el-input-number v-model="settingsStore.pollingInterval" :min="1" :step="1" size="small"
                    @change="handleIntervalChange" />
                </div>
              </template>
            </div>
          </el-card>
        </div>

        <div class="settings-section">
          <el-text class="section-title" strong>Server</el-text>
          <el-card shadow="never" class="info-card" style="margin-bottom: 24px">
            <div style="display: flex; align-items: center; justify-content: space-between">
              <div>
                <el-text strong>Restart</el-text>
                <div style="margin-top: 4px">
                  <el-text type="info" size="small">
                    Force the server to restart.
                  </el-text>
                </div>
              </div>
              <el-button plain @click="handleRestart" :loading="isRestarting">
                Restart
              </el-button>
            </div>
          </el-card>

          <el-text class="section-title" strong>About</el-text>
          <el-card shadow="never" class="info-card">
            <el-descriptions :column="1" border>
              <el-descriptions-item label="Application">
                <el-text strong>{{ systemConfig.app_name }}</el-text>
                <el-text type="info" size="small" style="margin-left: 8px">{{ systemConfig.app_version }}</el-text>
              </el-descriptions-item>
              <el-descriptions-item label="Server Address">
                <el-text tag="code">{{ systemConfig.address }}</el-text>
              </el-descriptions-item>
              <el-descriptions-item label="Environment">
                <el-text :type="systemConfig.mode === 'production' ? 'success' : 'warning'" strong>
                  {{ formatMode(systemConfig.mode) }}
                </el-text>
              </el-descriptions-item>
              <el-descriptions-item label="Data Directory">
                <el-text size="small" tag="code" type="info">{{ systemConfig.data_dir }}</el-text>
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <div style="margin-top: 16px">
            <el-button :icon="Link" link tag="a" :href="systemConfig.repo_url" target="_blank">
              Github
            </el-button>
          </div>
        </div>
      </div>

      <div v-else v-for="tab in dynamicTabs" :key="tab.id">
        <div v-if="selectedKey === tab.id">
          <el-space direction="vertical" alignment="start" :size="4" style="margin-bottom: 32px">
            <el-text size="large" strong style="font-size: 20px">{{ tab.label }}</el-text>
            <el-text type="info">{{ tab.description || tab.label.toLowerCase() }}</el-text>
          </el-space>
          <component :is="tab.component" />
        </div>
      </div>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Grid, Files, Link } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { systemApi as api } from '../api/system'
import type { SystemConfig } from '../api/system'
import { useAppSettingsStore } from '../stores/useAppSettingsStore'

const selectedKey = ref('general')
const settingsStore = useAppSettingsStore()
const dynamicTabs = computed(() => settingsStore.sortedSettingsTabs)
const systemConfig = ref<SystemConfig>({
  app_name: '', app_version: '', mode: '', data_dir: '', plugin_dir: '', address: '', repo_url: '', author: ''
})
const isRestarting = ref(false)

function handleMasterToggle() {
  settingsStore.saveSettings()
  if (settingsStore.enableMonitoring) {
    settingsStore.startPolling()
  } else {
    settingsStore.stopPolling()
  }
}

function handlePollingToggle() {
  settingsStore.saveSettings()
  if (settingsStore.enablePolling) {
    settingsStore.startPolling()
  } else {
    settingsStore.stopPolling()
  }
}

function handleIntervalChange() {
  settingsStore.saveSettings()
  settingsStore.restartPolling()
}

async function handleRestart() {
  try {
    await ElMessageBox.confirm(
      'Are you sure you want to restart the server? This will close all active connections.',
      'Restart Server',
      {
        confirmButtonText: 'Restart',
        cancelButtonText: 'Cancel',
        type: 'warning',
        confirmButtonClass: 'el-button--danger'
      }
    )

    isRestarting.value = true
    await api.restartServer()
    ElMessage.success('Restart signal sent. Please wait...')

    setTimeout(() => {
      window.location.reload()
    }, 3000)
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('Failed to restart server')
    }
  } finally {
    isRestarting.value = false
  }
}

onMounted(async () => {
  try {
    systemConfig.value = await api.getSystemConfig()
  } catch (err) { }
})

function formatMode(mode: string) {
  if (!mode) return ''
  return mode.charAt(0).toUpperCase() + mode.slice(1)
}
</script>

<style scoped>
.settings-panel {
  height: 100%;
}

.settings-sidebar {
  border-right: 1px solid var(--el-border-color-lighter);
}

.settings-content {
  height: 100%;
  overflow-y: auto;
  padding: 0 40px 40px 40px;
}

:deep(.el-main) {
  --el-main-padding: 0;
}

.settings-section {
  margin-bottom: 32px;
}

.section-title {
  display: block;
  margin-bottom: 16px;
  font-size: 16px;
  color: var(--el-text-color-primary);
}

.info-card {
  border-radius: 8px;
  max-width: 600px;
}

:deep(.el-descriptions__label) {
  width: 150px;
  font-weight: 600;
  color: var(--el-text-color-regular);
}
</style>
