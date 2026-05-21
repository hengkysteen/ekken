<template>
  <AppPage title="Credentials" subtitle="Manage your API keys and sensitive credentials securely.">

    <template #header-extra>
      <el-button :icon="Plus" type="primary" @click="openCreate"> Create </el-button>
    </template>

    <el-alert v-if="error" :title="error" type="error" :closable="false" />

    <div v-if="credentials.length > 0" class="credentials-container">

      <el-input v-model="searchQuery" placeholder="Search ..." clearable :prefix-icon="Search"
        style="width: 200px; margin: 0px 12px 12px 0px;" />

      <el-button dashed type="danger" style="margin: 0px 12px 12px 0px;" gosh v-show="selectedIds.length > 0"
        @click="confirmMassDelete">
        Delete ({{ selectedIds.length }})
      </el-button>


      <div class="table-container">

        <el-table stripe border size="large" :data="filteredCredentials" height="100%" style="width: 100%"
          @selection-change="handleSelectionChange">

          <el-table-column type="selection" width="55" align="center" />

          <el-table-column label="Reference" min-width="300">
            <template #default="{ row }">
              <el-space>
                <el-text type="primary">
                  {{ formatRef(row.key) }}
                </el-text>
                <el-button text class="row-copy-btn" size="small" @click="copyRef(row.key)">copy</el-button>
              </el-space>
            </template>
          </el-table-column>

          <el-table-column label="Tags" align="center" min-width="200">
            <template #default="{ row }">
              <template v-if="normalizeTags(row.tags).length > 0">
                <el-space :size="4">
                  <el-tag disable-transitions v-for="(tag, idx) in normalizeTags(row.tags).slice(0, 2)" :key="idx"
                    size="small" round effect="plain" type="info">
                    {{ tag }}
                  </el-tag>
                  <el-popover v-if="normalizeTags(row.tags).length > 2" placement="top" trigger="hover" width="auto">
                    <template #reference>
                      <el-tag size="small" round effect="plain" type="info" style="cursor: pointer">
                        +{{ normalizeTags(row.tags).length - 2 }}
                      </el-tag>
                    </template>
                    <el-space :size="4" wrap style="max-width: 200px">
                      <el-tag v-for="(tag, idx) in normalizeTags(row.tags).slice(2)" :key="idx" size="small" round
                        effect="plain" type="info">
                        {{ tag }}
                      </el-tag>
                    </el-space>
                  </el-popover>
                </el-space>
              </template>
              <el-text v-else type="info">—</el-text>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <el-empty :image-size="120" v-else-if="initialized && !loading" description="No credentials yet">
      <template #extra>
        <el-button type="primary" @click="openCreate">Create Your First Credential</el-button>
      </template>
    </el-empty>

    <!-- Create Dialog -->
    <el-dialog v-model="showDialog" title="New Credential" width="480px" destroy-on-close>
      <el-form :model="form" label-position="top">


        <el-form-item label="Name" required>
          <el-input v-model="form.name" placeholder="Key name" clearable @input="onNameInput" />
        </el-form-item>

        <el-form-item label="Reference">
          <el-text size="small" type="primary" strong>
            {{ formatRef(form.key || 'KEY_NAME') }}
          </el-text>
        </el-form-item>

        <el-form-item label="Value" required>
          <el-input v-model="form.value" type="password" show-password placeholder="Enter secret value" clearable />
          <el-text size="small" type="info">Once saved, this value will never be shown again.</el-text>
        </el-form-item>

        <el-form-item label="Tags">
          <el-space direction="vertical" fill :size="8">
            <el-space v-if="form.tags.length" :size="4" wrap>
              <el-tag v-for="(tag, idx) in form.tags" :key="idx" closable size="small" @close="removeTag(tag)">{{ tag
                }}</el-tag>
            </el-space>
            <el-input v-model="tagInput" placeholder="tags" @keydown.enter.prevent="addTag" @input="onTagInput">
              <template #append>
                <el-button @click="addTag" :disabled="!tagInput.trim()">Add</el-button>
              </template>
            </el-input>
          </el-space>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showDialog = false">Cancel</el-button>
        <el-button type="primary" :loading="saving" @click="saveCredential">
          Create
        </el-button>
      </template>
    </el-dialog>
  </AppPage>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { credentialsApi } from '@credentials/api'
import type { Credential } from '@credentials/api'
import AppPage from '@shared/components/AppPage.vue'

const loading = ref(false)
const initialized = ref(false)
const saving = ref(false)
const error = ref('')
const credentials = ref<Credential[]>([])
const searchQuery = ref('')
const selectedIds = ref<string[]>([])

const showDialog = ref(false)
const tagInput = ref('')

const form = ref({
  name: '',
  key: '',
  value: '',
  tags: [] as string[],
})

const filteredCredentials = computed(() => {
  if (!searchQuery.value) return credentials.value
  const query = searchQuery.value.toLowerCase()
  return credentials.value.filter(c =>
    c.key.toLowerCase().includes(query) ||
    normalizeTags(c.tags).some(t => t.toLowerCase().includes(query))
  )
})

function handleSelectionChange(selection: Credential[]) {
  selectedIds.value = selection.map(c => c.id)
}

function formatRef(key: string): string {
  if (!key || key === 'cred.') return '{{ cred.KEY_NAME }}'
  return `{{ ${key} }}`
}

function normalizeTags(tags: any): string[] {
  if (!tags) return []
  if (Array.isArray(tags)) return tags.filter(Boolean)
  if (typeof tags === 'string') {
    try {
      const parsed = JSON.parse(tags)
      return Array.isArray(parsed) ? parsed : [tags]
    } catch {
      return tags.split(',').map(t => t.trim()).filter(Boolean)
    }
  }
  return []
}

function onNameInput() {
  const base = form.value.name
    .trim()
    .toUpperCase()
    .replace(/\s+/g, '_')
    .replace(/[^A-Z0-9_]/g, '')

  form.value.key = `cred.${base}`
}

function onTagInput() {
  tagInput.value = tagInput.value
    .toUpperCase()
    .replace(/\s+/g, '_')
    .replace(/[^A-Z0-9_]/g, '')
}

async function copyRef(key: string) {
  try {
    await navigator.clipboard.writeText(formatRef(key))
    ElMessage.success('Copied to clipboard')
  } catch {
    ElMessage.error('Failed to copy')
  }
}

function resetForm() {
  tagInput.value = ''
  form.value = { name: '', key: 'cred.', value: '', tags: [] }
}

function openCreate() {
  resetForm()
  showDialog.value = true
}

function addTag() {
  const tag = tagInput.value.trim()
  if (!tag || form.value.tags.includes(tag)) return
  form.value.tags = [...form.value.tags, tag]
  tagInput.value = ''
}

function removeTag(tag: string) {
  form.value.tags = form.value.tags.filter(t => t !== tag)
}

async function loadCredentials() {
  loading.value = true
  error.value = ''
  try {
    credentials.value = await credentialsApi.list()
  } catch (err: any) {
    error.value = err.message || 'Failed to load credentials'
  } finally {
    loading.value = false
    initialized.value = true
  }
}

async function saveCredential() {
  if (!form.value.name.trim()) return ElMessage.warning('Name is required')
  if (!form.value.key.trim()) return ElMessage.warning('Key is required')
  if (!form.value.value.trim()) return ElMessage.warning('Value is required')

  saving.value = true
  try {
    await credentialsApi.create({
      name: form.value.name,
      key: form.value.key,
      value: form.value.value,
      tags: form.value.tags,
    })
    ElMessage.success('Credential created')
    showDialog.value = false
    await loadCredentials()
  } catch (err: any) {
    ElMessage.error(err.message || 'Failed to save credential')
  } finally {
    saving.value = false
  }
}

function confirmMassDelete() {
  ElMessageBox.confirm(
    `Delete ${selectedIds.value.length} selected credentials?`,
    'Delete',
    { confirmButtonText: 'Delete', type: 'warning' }
  ).then(() => deleteSelected()).catch(() => { })
}

async function deleteSelected() {
  loading.value = true
  try {
    await Promise.all(selectedIds.value.map(id => credentialsApi.delete(id)))
    ElMessage.success('Credentials deleted')
    selectedIds.value = []
    await loadCredentials()
  } catch (err: any) {
    ElMessage.error(err.message || 'Failed to delete some credentials')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadCredentials()
})

</script>

<style scoped>
.row-copy-btn {
  opacity: 0;
  transition: all 0.2s ease;

}

:deep(.el-table__row:hover) .row-copy-btn {
  opacity: 1;
}
</style>
