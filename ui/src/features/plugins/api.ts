import { request } from '@shared/api/request'

export interface NodeSpec {
  type: string
  label: string
  icon?: string
  description?: string
  kind: string
}

export interface OtherSpec {
  name: string
}

export interface Manifest {
  source: string
  kind: string
  spec?: unknown
  node?: NodeSpec
  other?: OtherSpec
}

export interface PluginSummary {
  id: string
  manifest: Manifest
  source_path: string
  status: string
  reason?: string
  is_installed: boolean
  is_enabled: boolean
  has_ui?: boolean
  ui_module_url?: string
  node_types?: string[]
}

export interface RegistryRepo {
  url: string
  about?: string
  author?: string
}

export interface RegistryArtifact {
  os: string
  arch: string
  download_url: string
  checksum: string
}

export interface RegistryPluginSummary {
  id: string
  source: string
  name: string
  kind: string
  version: string
  description?: string
  kind_meta?: Record<string, unknown>
  repo: RegistryRepo
  artifacts: RegistryArtifact[]
  is_installed: boolean
  is_enabled: boolean
  status?: string
  local_version?: string
}

export interface RegistryResponse {
  schema_version: string
  message?: string
  plugins: RegistryPluginSummary[]
}

export type InstallStatus =
  | 'queued'
  | 'downloading'
  | 'verifying'
  | 'extracting'
  | 'installing'
  | 'completed'
  | 'failed'
  | 'canceled'

export interface InstallTask {
  plugin_id: string
  status: InstallStatus
  progress: number
  bytes_received: number
  bytes_total?: number
  error?: string
}

export const pluginsApi = {
  getPlugins: () => request<PluginSummary[]>('/plugins'),
  reloadPlugins: () => request<{ reloaded: boolean }>('/plugins/reload', { method: 'POST' }),
  managePlugin: (id: string, action: string) => request<any>(`/plugins/${id}/${action}`, { method: 'POST' }),
  getRegistry: () => request<RegistryResponse>('/plugins/registry'),
  installPlugin: (id: string) => request<InstallTask>(`/plugins/registry/${id}/install`, { method: 'POST' }),
  getInstallStatus: (id: string) => request<InstallTask>(`/plugins/registry/${id}/install`),
  stopInstall: (id: string) => request<InstallTask>(`/plugins/registry/${id}/install`, { method: 'DELETE' }),
  uninstallPlugin: (id: string) => request<{ uninstalled: boolean }>(`/plugins/${id}`, { method: 'DELETE' }),
}
