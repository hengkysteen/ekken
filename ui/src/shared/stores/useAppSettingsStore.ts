import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { SettingsTab } from '../../core/types/module'
import { systemApi as api } from '../api/system'
import type { DeviceInfo } from '../api/system'

export const useAppSettingsStore = defineStore('appSettings', () => {
  const settingsTabs = ref<SettingsTab[]>([])

  // Resource Monitoring Settings
  const enableMonitoring = ref(localStorage.getItem('settings_enable_monitoring') !== 'false')
  const enablePolling = ref(localStorage.getItem('settings_enable_polling') !== 'false')
  
  // Fix: Handle migration from ms to seconds
  const savedInterval = Number(localStorage.getItem('settings_polling_interval'))
  const pollingInterval = ref(savedInterval >= 500 ? savedInterval / 1000 : (savedInterval || 2))

  const deviceInfo = ref<DeviceInfo>({
    os: '', arch: '', hostname: '', cpu_model: '', cpu_cores: 0, cpu_usage: 0,
    ram_total: 0, ram_used: 0, ram_free: 0, ram_usage: 0, uptime: 0
  })

  let pollInterval: any = null

  const saveSettings = () => {
    localStorage.setItem('settings_enable_monitoring', enableMonitoring.value.toString())
    localStorage.setItem('settings_enable_polling', enablePolling.value.toString())
    localStorage.setItem('settings_polling_interval', pollingInterval.value.toString())
  }

  const fetchDeviceInfo = async () => {
    try {
      deviceInfo.value = await api.getDeviceInfo()
    } catch (e) { }
  }

  const startPolling = () => {
    // If master switch is off, do absolutely nothing
    if (!enableMonitoring.value) return

    // Fetch once
    fetchDeviceInfo()

    if (pollInterval || !enablePolling.value) return
    pollInterval = setInterval(fetchDeviceInfo, pollingInterval.value * 1000)
  }

  const stopPolling = () => {
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
  }

  const restartPolling = () => {
    stopPolling()
    if (enablePolling.value) {
      startPolling()
    }
  }

  const addSettingsTabs = (tabs: SettingsTab[]) => {
    settingsTabs.value.push(...tabs)
  }

  const sortedSettingsTabs = computed(() => {
    return [...settingsTabs.value].sort((a, b) => (a.order ?? 99) - (b.order ?? 99))
  })

  return {
    settingsTabs,
    enableMonitoring,
    enablePolling,
    pollingInterval,
    deviceInfo,
    saveSettings,
    startPolling,
    stopPolling,
    restartPolling,
    addSettingsTabs,
    sortedSettingsTabs
  }
})
