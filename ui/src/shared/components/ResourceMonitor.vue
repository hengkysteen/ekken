<template>
  <div v-if="settingsStore.enableMonitoring" class="resource-container">
    <!-- CPU Tooltip -->
    <el-tooltip effect="dark" placement="bottom">
      <template #content>
        <div class="tooltip-content">
          <div class="tooltip-title">{{ settingsStore.deviceInfo.cpu_model }}</div>
          <div class="tooltip-detail">{{ settingsStore.deviceInfo.cpu_cores }} Cores | {{
            settingsStore.deviceInfo.cpu_usage.toFixed(1) }}% Usage</div>
        </div>
      </template>
      <div class="resource-item">
        <el-icon :size="14" :class="{ 'status-warning': settingsStore.deviceInfo.cpu_usage > 80 }">
          <Cpu />
        </el-icon>
        <span class="usage-value">{{ cpuDisplay.toFixed(0) }}%</span>
      </div>
    </el-tooltip>

    <!-- RAM Tooltip -->
    <el-tooltip effect="dark" placement="bottom">
      <template #content>
        <div class="tooltip-content">
          <div class="tooltip-title">Memory Usage</div>
          <div class="tooltip-detail">
            {{ formatBytesToGB(settingsStore.deviceInfo.ram_used) }} / {{
              formatBytesToGB(settingsStore.deviceInfo.ram_total) }} GB
          </div>
        </div>
      </template>
      <div class="resource-item">
        <el-icon :size="14" :class="{ 'status-warning': settingsStore.deviceInfo.ram_usage > 90 }">
          <Odometer />
        </el-icon>
        <span class="usage-value">{{ ramDisplay.toFixed(0) }}%</span>
      </div>
    </el-tooltip>
  </div>
</template>

<script setup lang="ts">
import { computed, type ComputedRef } from 'vue'
import { useAppSettingsStore } from '../stores/useAppSettingsStore'
import { Cpu, Odometer } from '@element-plus/icons-vue'
import { useTransition, TransitionPresets } from '@vueuse/core'

const settingsStore = useAppSettingsStore()

// Smooth transitions for numbers
const cpuDisplay = useTransition(computed(() => settingsStore.deviceInfo.cpu_usage), {
  duration: 1000,
  transition: TransitionPresets.easeOutCubic,
}) as ComputedRef<number>

const ramDisplay = useTransition(computed(() => settingsStore.deviceInfo.ram_usage), {
  duration: 1000,
  transition: TransitionPresets.easeOutCubic,
}) as ComputedRef<number>

function formatBytesToGB(bytes: number) {
  if (!bytes) return '0'
  return (bytes / (1024 * 1024 * 1024)).toFixed(1)
}
</script>

<style scoped>
.resource-container {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 8px;
}

.resource-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--el-text-color-secondary);
  cursor: default;
}

.usage-value {
  font-size: 12px;
  font-weight: 600;
  font-family: var(--el-font-family-mono);
  min-width: 28px;
}

.status-warning {
  color: var(--el-color-danger);
}

.tooltip-content {
  padding: 4px;
}

.tooltip-title {
  font-weight: bold;
  margin-bottom: 4px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  padding-bottom: 4px;
}

.tooltip-detail {
  font-size: 12px;
  opacity: 0.9;
}
</style>
