<template>
  <el-dialog :model-value="modelValue" @update:model-value="$emit('update:modelValue', $event)"
    title="Assistant Providers" width="850px" append-to-body destroy-on-close>
    <el-scrollbar height="550px">
      <el-space direction="vertical" fill :size="24" style="padding: 24px; width: 100%; box-sizing: border-box;">
        <!-- Header Row -->
        <el-row justify="space-between" align="middle">
          <el-space direction="vertical" alignment="flex-start" :size="2">
            <el-text size="large" bold tag="h3" style="margin: 0">LLM Providers</el-text>
            <el-text type="info" size="small">Manage your AI providers and credentials.</el-text>
          </el-space>
          <el-button type="primary" :icon="Plus" @click="openAddDialog">Add Provider</el-button>
        </el-row>
        <!-- Loading State -->
        <el-skeleton v-if="loadingProviders" :rows="3" animated />
        <!-- Empty State -->
        <el-empty v-if="!loadingProviders && providerList.length === 0" description="No providers configured yet.">
          <el-button type="primary" @click="openAddDialog">Connect your first provider</el-button>
        </el-empty>
        <!-- Providers List -->
        <el-space direction="vertical" fill :size="16">
          <el-card v-for="item in providerList" :key="item.provider_id" shadow="never" style="border-radius: 12px">
            <template #header>
              <el-row justify="space-between" align="middle">
                <el-space :size="12" alignment="flex-start">
                  <el-image style="width: 30px; height: 30px" :src="item.logo" fit="contain" />
                  <el-space direction="vertical" alignment="flex-start" :size="2">
                    <el-text bold style="line-height: 1.2">
                      {{ item.name }}
                    </el-text>
                    <el-link :href="item.official_url" target="_blank" style="font-size: 11px">
                      {{ item.official_url }}
                    </el-link>
                  </el-space>
                </el-space>
                <el-button type="danger" :icon="Delete" link @click="removeProvider(item.provider_id)" />
              </el-row>
            </template>
            <el-form label-position="top">
              <div v-for="field in getFields(item)" :key="field">
                <el-form-item :label="field" required style="margin-bottom: 12px">
                  <el-input v-model="item.config[field]" type="password" show-password
                    :placeholder="`Enter ${field}`">
                    <template #suffix>
                      <el-button :icon="Key" link @click="triggerSelect(item, field)" />
                    </template>
                  </el-input>
                </el-form-item>
              </div>
            </el-form>
          </el-card>
        </el-space>
      </el-space>
    </el-scrollbar>
    <template #footer>
      <el-row justify="end" style="padding: 10px 0">
        <el-button @click="$emit('update:modelValue', false)">Cancel</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          Save All Changes
        </el-button>
      </el-row>
    </template>
    <!-- Dialogs -->
    <el-dialog v-model="showAddDialog" title="Select Provider" width="500px" append-to-body>
      <el-scrollbar max-height="400px">
        <el-space direction="vertical" fill :size="8" style="padding: 12px; width: 100%; box-sizing: border-box;">
          <el-card v-for="p in availableProviders" :key="p.id" shadow="hover"
            style="cursor: pointer; border-radius: 8px" @click="addProvider(p)">
            <el-row justify="space-between" align="middle">
              <el-space :size="12">
                <el-image style="width: 30px; height: 30px" :src="p.logo" fit="contain" />
                <el-space direction="vertical" alignment="flex-start" :size="0">
                  <el-text bold>{{ p.name }}</el-text>
                  <el-text type="info" size="small">{{ p.official_url }}</el-text>
                </el-space>
              </el-space>
              <el-icon>
                <ArrowRight />
              </el-icon>
            </el-row>
          </el-card>
        </el-space>
      </el-scrollbar>
    </el-dialog>
    <!-- Reusable Credential Selector -->
    <CredentialSelector v-model="showSelector" @select="onCredentialSelected" />
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete, ArrowRight, Key } from '@element-plus/icons-vue'
import CredentialSelector from '@credentials/components/CredentialSelector.vue'
import { useAssistantStore } from '../store/useAssistantStore'

defineProps<{
  modelValue: boolean
}>()
const emit = defineEmits(['update:modelValue', 'saved'])
const store = useAssistantStore()
const providerList = ref<any[]>([])
const loadingProviders = ref(false)
const saving = ref(false)
const availableProviders = ref<any[]>([])
const loadingAvailable = ref(false)
const showAddDialog = ref(false)

// Credential selector state
const showSelector = ref(false)
const activeProvider = ref<any>(null)
const activeField = ref<string>('')

const getFields = (item: any) => {
  const p = availableProviders.value.find(ap => ap.id === item.provider_id)
  return p?.config_fields || []
}

const fetchProviders = async () => {
  loadingProviders.value = true
  try {
    const res = await fetch('/api/assistant/providers')
    const json = await res.json()
    providerList.value = json.data || []
  } catch (e) {
    ElMessage.error('Failed to fetch providers')
  } finally {
    loadingProviders.value = false
  }
}

const fetchCatalogs = async () => {
  loadingAvailable.value = true
  try {
    const res = await fetch('/api/assistant/catalogs')
    const json = await res.json()
    availableProviders.value = json.data || []
  } catch (e) {
    ElMessage.error('Failed to fetch catalogs')
  } finally {
    loadingAvailable.value = false
  }
}

const openAddDialog = () => {
  showAddDialog.value = true
}

const addProvider = (p: any) => {
  if (providerList.value.some(item => item.provider_id === p.id)) {
    ElMessage.warning(`${p.name} is already added`)
    showAddDialog.value = false
    return
  }

  const config: Record<string, string> = {}
  p.config_fields?.forEach((f: string) => { config[f] = '' })

  providerList.value.push({
    provider_id: p.id,
    name: p.name,
    official_url: p.official_url,
    logo: p.logo,
    config: config
  })
  showAddDialog.value = false
}

const triggerSelect = (item: any, field: string) => {
  activeProvider.value = item
  activeField.value = field
  showSelector.value = true
}

const onCredentialSelected = (cred: any) => {
  if (activeProvider.value && activeField.value) {
    activeProvider.value.config[activeField.value] = `{{ ${cred.key} }}`
  }
}


const removeProvider = (providerID: string) => {
  providerList.value = providerList.value.filter(item => item.provider_id !== providerID)
}



const handleSave = async () => {
  for (const p of providerList.value) {
    const fields = getFields(p)
    for (const field of fields) {
      if (!p.config[field]) {
        ElMessage.warning(`Please fill in ${field} for ${p.name}`)
        return
      }
    }
  }

  saving.value = true
  try {
    const resProviders = await fetch('/api/assistant/providers')
    const jsonProviders = await resProviders.json()
    const serverProviders = jsonProviders.data || []
    const serverIDs = serverProviders.map((p: any) => p.provider_id)

    const currentIDs = providerList.value.map(p => p.provider_id)
    const toDelete = serverIDs.filter((id: string) => !currentIDs.includes(id))

    for (const id of toDelete) {
      await fetch(`/api/assistant/providers/${id}`, { method: 'DELETE' })
    }

    for (const p of providerList.value) {
      const res = await fetch(`/api/assistant/providers/setup`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          provider_id: p.provider_id,
          config: p.config
        })
      })
      if (!res.ok) {
        const err = await res.json()
        throw new Error(err.error || `Failed to save ${p.name}`)
      }
    }

    ElMessage.success('Settings saved successfully')
    await store.fetchProviderList()
    emit('saved')
    emit('update:modelValue', false)
  } catch (e: any) {
    ElMessage.error(`Failed to save: ${e.message}`)
  } finally {
    saving.value = false
    fetchProviders()
  }
}

onMounted(async () => {
  await fetchCatalogs()
  await fetchProviders()
})
</script>

<style scoped></style>