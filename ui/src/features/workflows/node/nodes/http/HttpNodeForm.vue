<template>
  <el-form label-position="top" class="http-config" @submit.prevent>
    <!-- URL Bar -->
    <div class="url-bar">
      <el-select v-model="config.method" class="method-select" placeholder="Method">
        <el-option v-for="opt in methods" :key="opt.value" :label="opt.label" :value="opt.value" />
      </el-select>
      <el-input v-model="config.url" placeholder="https://api.example.com/data" class="url-input" />
      <el-button type="primary" :loading="sending" @click="sendRequest" :disabled="!config.url">
        <el-icon>
          <Promotion />
        </el-icon>
        <span class="btn-text">Test</span>
      </el-button>
      <el-button @click="showCurlModal = true" title="Import cURL">
        <el-icon>
          <Monitor />
        </el-icon>
      </el-button>
      <el-button type="danger" plain @click="resetForm" title="Reset All Fields">
        <el-icon>
          <RefreshRight />
        </el-icon>
      </el-button>
    </div>

    <!-- Tabs: Params / Headers / Body / Auth / Settings -->
    <el-tabs v-model="activeTab" class="config-tabs">
      <el-tab-pane label="Params" name="params">
        <KeyValueEditor v-model="paramsList" placeholder-key="Key" placeholder-value="Value"
          add-label="Add Parameter" />
      </el-tab-pane>

      <el-tab-pane label="Headers" name="headers">
        <KeyValueEditor v-model="headersList" placeholder-key="Header" placeholder-value="Value"
          add-label="Add Header" />
      </el-tab-pane>

      <el-tab-pane label="Body" name="body">
        <div class="body-section">
          <div class="body-type-selector">
            <el-radio-group v-model="bodyType" size="small">
              <el-radio-button value="json">JSON</el-radio-button>
              <el-radio-button value="raw">Raw</el-radio-button>
              <el-radio-button value="none">None</el-radio-button>
            </el-radio-group>
          </div>
          <SimpleJsonEditor v-if="bodyType === 'json'" v-model="config.body" height="220px" />
          <el-input v-else-if="bodyType === 'raw'" v-model="config.body" type="textarea" :rows="10"
            placeholder="Raw request body" class="body-textarea" />
        </div>
      </el-tab-pane>

      <el-tab-pane label="Auth" name="auth">
        <div class="auth-section">
          <el-select v-model="authType" placeholder="No Auth" class="auth-type-select" clearable>
            <el-option v-for="opt in authTypes" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>

          <div v-if="authType" class="auth-fields">
            <template v-if="authType === 'bearer'">
              <el-form-item label="Token">
                <el-input v-model="auth.bearer.token" placeholder="Bearer token" type="password" show-password />
              </el-form-item>
            </template>

            <template v-if="authType === 'basic'">
              <el-form-item label="Username">
                <el-input v-model="auth.basic.username" placeholder="Username" />
              </el-form-item>
              <el-form-item label="Password">
                <el-input v-model="auth.basic.password" placeholder="Password" type="password" show-password />
              </el-form-item>
            </template>

            <template v-if="authType === 'apikey'">
              <el-form-item label="Key Name">
                <el-input v-model="auth.apikey.key" placeholder="X-API-Key" />
              </el-form-item>
              <el-form-item label="Value">
                <el-input v-model="auth.apikey.value" placeholder="Your API key" type="password" show-password />
              </el-form-item>
              <el-form-item label="Add to">
                <el-select v-model="auth.apikey.in" style="width: 100%">
                  <el-option v-for="opt in apikeyLocations" :key="opt.value" :label="opt.label" :value="opt.value" />
                </el-select>
              </el-form-item>
            </template>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="Settings" name="settings">
        <div class="settings-section">
          <el-form-item label="Timeout (Seconds)">
            <el-input-number v-model="config.timeout" :min="1" :max="300" style="width: 100%" />
            <div class="form-item-hint">
              Maximum time to wait for a response (1-300 seconds).
            </div>
          </el-form-item>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- Response Panel -->
    <el-divider />
    <div class="response-panel">
      <div class="response-header">
        <div class="response-meta">
          <el-tag v-if="response" :type="responseStatusType" size="small">{{ response.status }}</el-tag>
          <span v-if="response && response.statusText" class="status-text">{{ cleanStatusText(response.statusText)
          }}</span>
          <span v-if="response" class="time-text">{{ response.time }}ms</span>
          <span v-if="response?.size" class="size-text">{{ formatSize(response.size) }}</span>
          <span v-else class="no-response-text">No response sent yet</span>
        </div>
        <span class="response-label">Response</span>
      </div>
      <el-tabs v-model="activeResponseTab" size="small">
        <el-tab-pane label="Body" name="body">
          <el-input :model-value="formattedResponseBody" readonly type="textarea" :rows="8" class="response-body" />
        </el-tab-pane>
        <el-tab-pane label="Headers" name="headers">
          <el-input :model-value="responseHeadersText" readonly type="textarea" :rows="6" class="response-headers" />
        </el-tab-pane>
      </el-tabs>
    </div>
  </el-form>

  <!-- cURL Modal -->
  <el-dialog v-model="showCurlModal" title="Import from cURL" width="600px">
    <p class="dialog-hint">Paste your cURL command below.</p>
    <el-input v-model="curlText" type="textarea" :rows="12" :placeholder="curlPlaceholder" class="curl-textarea" />
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="showCurlModal = false">Cancel</el-button>
        <el-button type="primary" @click="handleImportCurl">Import</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import { Promotion, Monitor, RefreshRight } from '@element-plus/icons-vue'
import { useNodeStore } from '@workflows/node/stores/node'
import { getActionBlueprint, getActionValue, serializeActionForSave } from '@workflows/node/utils/node'
import type { NodeFormProps } from '@workflows/node/types/node'
import KeyValueEditor from '@shared/components/KeyValueEditor.vue'
import SimpleJsonEditor from '@shared/components/SimpleJsonEditor.vue'


type HttpMethod = string
type BodyType = 'json' | 'raw' | 'none'
type AuthType = string
type ApiKeyLocation = string

interface KeyValueItem {
  key: string
  value: string
}

interface HttpConfig {
  method: HttpMethod
  url: string
  headers: string
  body: string
  timeout: number
  params: KeyValueItem[]
  auth: {
    type: AuthType
    bearer: { token: string }
    basic: { username: string; password: string }
    apikey: { key: string; value: string; in: ApiKeyLocation }
  }
}

interface HttpResponse {
  status: number
  statusText: string
  body: string
  headers: Record<string, string>
  time: number
  size: number
}

const props = defineProps<{
  node: NodeFormProps['node']
}>()

const nodeStore = useNodeStore()
const nodeDef = computed(() => nodeStore.findDef('http'))

const activeTab = ref('body')
const activeResponseTab = ref('body')
const showCurlModal = ref(false)
const curlText = ref('')
const curlPlaceholder = `curl --request POST \\
  --url https://api.example.com/data \\
  --header 'Content-Type: application/json' \\
  --data '{"foo":"bar"}'`

const methods = computed(() => {
  const methodField = nodeDef.value?.actions?.[0]?.fields?.find((f: any) => f.key === 'method')
  return methodField?.options?.map((m: string) => ({ label: m, value: m })) || [
    { label: 'GET', value: 'GET' },
    { label: 'POST', value: 'POST' }
  ]
})

const authTypes = [
  { label: 'Bearer Token', value: 'bearer' },
  { label: 'Basic Auth', value: 'basic' },
  { label: 'API Key', value: 'apikey' }
]

const apikeyLocations = [
  { label: 'Header', value: 'header' },
  { label: 'Query Param', value: 'query' }
]

const config = ref<HttpConfig>({} as HttpConfig)

const getFieldDefault = (key: string) => {
  const field = nodeDef.value?.actions?.[0]?.fields?.find((f: any) => f.key === key)
  return field?.default
}

const paramsList = ref<KeyValueItem[]>([])
const headersList = ref<KeyValueItem[]>([])
const bodyType = ref<BodyType>('json')
const authType = ref<AuthType>('')
const auth = ref({ bearer: { token: '' }, basic: { username: '', password: '' }, apikey: { key: '', value: '', in: 'header' as ApiKeyLocation } })

// Response state
const sending = ref(false)
const response = ref<HttpResponse | null>(null)
const isInternalUpdate = ref(false)

const responseStatusType = computed(() => {
  if (!response.value) return 'info'
  const s = response.value.status
  if (s >= 200 && s < 300) return 'success'
  if (s >= 400 && s < 500) return 'warning'
  if (s >= 500) return 'danger'
  return 'info'
})

const formattedResponseBody = computed(() => {
  if (!response.value) return ''
  try {
    return JSON.stringify(JSON.parse(response.value.body), null, 2)
  } catch {
    return response.value.body
  }
})

const responseHeadersText = computed(() => {
  if (!response.value?.headers) return ''
  return Object.entries(response.value.headers).map(([k, v]) => `${k}: ${v}`).join('\n')
})

// Sync: props.node → local state
watch(() => props.node.data?.action, (newAction) => {
  if (newAction) {
    config.value = {
      method: (getActionValue(newAction, 'method', getFieldDefault('method') || 'GET')) as HttpMethod,
      url: getActionValue(newAction, 'url', getFieldDefault('url') || '') || '',
      headers: getActionValue(newAction, 'headers', getFieldDefault('headers') || '') || '',
      body: getActionValue(newAction, 'body', getFieldDefault('body') || '') || '',
      timeout: getActionValue(newAction, 'timeout', getFieldDefault('timeout') || 60) || 60,
      params: [], // These might be handled specifically in local state
      auth: { type: '', bearer: { token: '' }, basic: { username: '', password: '' }, apikey: { key: '', value: '', in: 'header' } }
    }

    isInternalUpdate.value = true

    // Restore bodyType from saved body value
    const body = config.value.body
    if (body) {
      try { JSON.parse(body); bodyType.value = 'json' } catch { bodyType.value = 'raw' }
    } else {
      bodyType.value = 'json'
    }

    const headers = config.value.headers
    if (headers) {
      headersList.value = headers.split('\n').filter(l => l.trim()).map(line => {
        const idx = line.indexOf(':')
        if (idx === -1) return { key: line.trim(), value: '' }
        return { key: line.slice(0, idx).trim(), value: line.slice(idx + 1).trim() }
      })
    } else {
      headersList.value = []
    }

    nextTick(() => { isInternalUpdate.value = false })
  }
}, { immediate: true })

watch(paramsList, (val) => {
  if (isInternalUpdate.value) return
  config.value.params = val.filter(p => p.key)
}, { deep: true })

watch(headersList, (val) => {
  if (isInternalUpdate.value) return
  config.value.headers = val.filter(h => h.key).map(h => `${h.key}: ${h.value}`).join('\n')
}, { deep: true })

watch([authType, auth], ([type, a]) => {
  if (isInternalUpdate.value) return
  config.value.auth = { type, bearer: a.bearer, basic: a.basic, apikey: a.apikey }
}, { deep: true })

const sendRequest = async () => {
  if (!config.value.url) return
  sending.value = true
  response.value = null

  let url = config.value.url
  const params = new URLSearchParams()
  config.value.params?.forEach(p => { if (p.key) params.append(p.key, p.value) })
  if (params.toString()) url += (url.includes('?') ? '&' : '?') + params.toString()

  const headers: Record<string, string> = {}
  config.value.headers?.split('\n').filter(l => l.trim()).forEach(line => {
    const idx = line.indexOf(':')
    if (idx !== -1) headers[line.slice(0, idx).trim()] = line.slice(idx + 1).trim()
  })

  const a = config.value.auth
  if (a?.type === 'bearer' && a.bearer?.token) headers['Authorization'] = `Bearer ${a.bearer.token}`
  if (a?.type === 'basic' && a.basic?.username) headers['Authorization'] = 'Basic ' + btoa(`${a.basic.username}:${a.basic.password}`)
  if (a?.type === 'apikey') {
    if (a.apikey?.in === 'header' && a.apikey?.key) headers[a.apikey.key] = a.apikey.value
    if (a.apikey?.in === 'query' && a.apikey?.key) params.append(a.apikey.key, a.apikey.value)
  }

  try {
    const res = await fetch('/api/http/test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        url,
        method: config.value.method,
        headers,
        body: (config.value.method !== 'GET' && config.value.method !== 'HEAD' && config.value.body) ? config.value.body : undefined,
        timeout: config.value.timeout || 60,
      })
    })

    const data = await res.json()
    response.value = {
      status: data.status,
      statusText: data.status_text,
      body: data.body || '',
      headers: data.headers || {},
      time: data.time_ms || 0,
      size: data.size || 0
    }
  } catch (err) {
    response.value = {
      status: 0,
      statusText: (err as Error).message,
      body: (err as Error).message,
      headers: {},
      time: 0,
      size: 0
    }
  } finally {
    sending.value = false
  }
}

const cleanStatusText = (text: string): string => {
  const parts = text.split(' ')
  if (parts.length > 1 && /^\d+$/.test(parts[0])) {
    return parts.slice(1).join(' ')
  }
  return text
}

const formatSize = (bytes: number): string => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1048576).toFixed(1) + ' MB'
}

const resetForm = () => {
  isInternalUpdate.value = true
  config.value.method = (getFieldDefault('method') || 'GET') as HttpMethod
  config.value.url = ''
  config.value.headers = ''
  config.value.body = ''
  config.value.timeout = getFieldDefault('timeout') || 60
  config.value.params = []
  config.value.auth = { type: '', bearer: { token: '' }, basic: { username: '', password: '' }, apikey: { key: '', value: '', in: 'header' } }

  headersList.value = []
  paramsList.value = []
  authType.value = ''
  auth.value = { bearer: { token: '' }, basic: { username: '', password: '' }, apikey: { key: '', value: '', in: 'header' } }
  bodyType.value = 'json'
  response.value = null

  nextTick(() => { isInternalUpdate.value = false })
}

const handleImportCurl = () => {
  const curl = curlText.value.trim()
  if (!curl) {
    showCurlModal.value = false
    return
  }

  const result = { method: 'GET' as HttpMethod, url: '', headers: [] as string[], body: '' }
  const tokens = tokenizeCurl(curl)

  for (let i = 0; i < tokens.length; i++) {
    const token = tokens[i]
    if ((token === '-X' || token === '--request') && tokens[i + 1]) {
      result.method = tokens[++i].toUpperCase() as HttpMethod
    } else if ((token === '-H' || token === '--header') && tokens[i + 1]) {
      result.headers.push(tokens[++i])
    } else if ((token === '-d' || token === '--data' || token === '--data-raw' || token === '--data-binary') && tokens[i + 1]) {
      result.body = tokens[++i]
      if (result.method === 'GET') result.method = 'POST'
    } else if (token === '--url' && tokens[i + 1]) {
      result.url = tokens[++i]
    } else if (token.startsWith('http') && !result.url) {
      result.url = token
    }
  }

  if (result.url) {
    let finalUrl = result.url
    const params: KeyValueItem[] = []

    try {
      const urlObj = new URL(result.url)
      finalUrl = urlObj.origin + urlObj.pathname
      urlObj.searchParams.forEach((value, key) => {
        params.push({ key, value })
      })
    } catch (e) {
      // Not a full URL, maybe just a path or has template variables
      const qIdx = finalUrl.indexOf('?')
      if (qIdx !== -1) {
        const queryString = finalUrl.slice(qIdx + 1)
        finalUrl = finalUrl.slice(0, qIdx)
        queryString.split('&').forEach(pair => {
          const [key, value] = pair.split('=')
          if (key) params.push({ key, value: value || '' })
        })
      }
    }

    config.value.method = result.method
    config.value.url = finalUrl
    paramsList.value = params.length ? params : [{ key: '', value: '' }]
    headersList.value = result.headers.map(h => {
      const idx = h.indexOf(':')
      return idx === -1 ? { key: h, value: '' } : { key: h.slice(0, idx).trim(), value: h.slice(idx + 1).trim() }
    })
    if (!headersList.value.length) headersList.value = [{ key: '', value: '' }]
    
    config.value.body = result.body
    if (result.body) {
      const isJson = headersList.value.some(h => 
        h.key.toLowerCase() === 'content-type' && 
        h.value.toLowerCase().includes('application/json')
      )
      bodyType.value = isJson ? 'json' : 'raw'
    } else {
      bodyType.value = 'none'
    }
  }

  curlText.value = ''
  showCurlModal.value = false
}

const tokenizeCurl = (curl: string): string[] => {
  const tokens: string[] = []
  let current = ''
  let inQuote = false
  let quoteChar = ''
  let escaped = false

  for (let i = 0; i < curl.length; i++) {
    const ch = curl[i]
    if (escaped) {
      if (ch === '\n') continue
      current += ch
      escaped = false
      continue
    }
    if (ch === '\\') {
      escaped = true
      continue
    }
    if (inQuote) {
      if (ch === quoteChar) inQuote = false
      else current += ch
      continue
    }
    if (ch === '"' || ch === "'") {
      inQuote = true
      quoteChar = ch
      if (current.trim()) {
        tokens.push(current.trim())
        current = ''
      }
      continue
    }
    if (ch === ' ' || ch === '\n' || ch === '\t') {
      if (current.trim()) {
        tokens.push(current.trim())
        current = ''
      }
      continue
    }
    current += ch
  }
  if (current.trim()) tokens.push(current.trim())
  return tokens
}

const getData = (): any => {
  const originalAction = props.node.data.action || {}
  const actionBlueprint = getActionBlueprint(nodeDef.value, originalAction.key)
  if (!actionBlueprint) return serializeActionForSave(originalAction)

  const action = {
    key: actionBlueprint.key,
    response_var: originalAction.response_var,
    fields: actionBlueprint.fields.map((field: any) => ({
      key: field.key,
      value: getActionValue(originalAction, field.key, field.default)
    }))
  }

  action.fields = action.fields.map((f: any) => {
    if (f.key === 'method') return { key: f.key, value: config.value.method }
    if (f.key === 'url') return { key: f.key, value: config.value.url }
    if (f.key === 'headers') return { key: f.key, value: config.value.headers }
    if (f.key === 'body') return { key: f.key, value: config.value.body }
    if (f.key === 'timeout') return { key: f.key, value: config.value.timeout }
    return f
  })

  return serializeActionForSave(action)
}

defineExpose({ getData })
</script>

<style scoped>
.http-config {
  display: flex;
  flex-direction: column;
}

.url-bar {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 4px;
}

.method-select {
  width: 120px;
  flex-shrink: 0;
}

.url-input {
  flex: 1;
}

.btn-text {
  margin-left: 4px;
}

.config-tabs {
  margin-top: 12px;
}

.body-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.body-type-selector {
  display: flex;
}

.body-textarea {
  font-family: var(--el-font-family-mono);
  font-size: 13px;
}

.settings-section {
  padding: 8px 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.auth-section {
  padding: 8px 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.auth-type-select {
  width: 200px;
  margin-bottom: 16px;
}

.auth-fields {
  display: flex;
  flex-direction: column;
}

.response-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.response-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
}

.response-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-text,
.time-text,
.size-text,
.no-response-text {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.response-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-regular);
}

.response-body,
.response-headers {
  font-family: var(--el-font-family-mono);
  font-size: 12px;
}

.dialog-hint {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 12px;
}

.curl-textarea {
  font-family: var(--el-font-family-mono);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

:deep(.el-tabs__header) {
  margin-bottom: 12px;
}

:deep(.el-form-item) {
  margin-bottom: 16px;
}

.form-item-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>
