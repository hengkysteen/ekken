import { request } from '@shared/api/request'

export interface SystemConfig {
  app_name: string
  app_version: string
  mode: string
  data_dir: string
  plugin_dir: string
  address: string
  repo_url: string
  author: string

}

export interface DeviceInfo {
  os: string
  arch: string
  hostname: string
  cpu_model: string
  cpu_cores: number
  cpu_usage: number
  ram_total: number
  ram_used: number
  ram_free: number
  ram_usage: number
  uptime: number
}

export const systemApi = {
  // System Metrics
  getSystemConfig: () => request<SystemConfig>('/system/config'),
  restartServer: () => request<any>('/system/restart', { method: 'POST' }),
  getDeviceInfo: () => request<DeviceInfo>('/system/device'),
  openFilePicker: () => request<string>('/system/file-picker', { method: 'POST' }),
}
