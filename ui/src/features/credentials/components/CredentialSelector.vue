<template>
  <el-dialog :model-value="modelValue" @update:model-value="$emit('update:modelValue', $event)"
    title="Select Credential" width="480px" append-to-body destroy-on-close class="credential-selector-dialog">

    <!-- Search & Filter Header -->
    <div style="padding-bottom: 16px; border-bottom: 1px solid var(--el-border-color-lighter); margin-bottom: 8px;">
      <el-input v-model="search" placeholder="Search by name, key, or tag..." :prefix-icon="Search" clearable
        style="margin-bottom: 12px;" />

      <div v-if="uniqueTags.length > 0">
        <el-scrollbar>
          <div style="display: flex; gap: 8px; padding-bottom: 8px; white-space: nowrap;">
            <el-check-tag v-for="tag in uniqueTags" :key="tag" :checked="activeTag === tag" @change="toggleTag(tag)"
              style="flex-shrink: 0;">
              {{ tag }}
            </el-check-tag>
          </div>
        </el-scrollbar>
      </div>
    </div>

    <!-- Scrollable Results -->
    <el-scrollbar height="400px">
      <div v-loading="loading">
        <div v-if="filteredCredentials.length > 0" style="padding-right: 12px;">
          <div v-for="c in filteredCredentials" :key="c.id" style="margin-bottom: 4px;">
            <ListTile clickable @click="handleSelect(c)">
              <template #leading>
                <el-avatar :size="32" shape="square"
                  style="background-color: var(--el-color-primary-light-9); color: var(--el-color-primary);">
                  <el-icon>
                    <Key />
                  </el-icon>
                </el-avatar>
              </template>

              <template #title>
                <el-text strong>{{ c.name }}</el-text>
              </template>

              <template #subtitle>
                <el-space :size="8">
                  <el-text type="info" size="small" tag="code">{{ c.key }}</el-text>
                  <el-space :size="4" v-if="c.tags?.length">
                    <el-tag v-for="t in c.tags" :key="t" size="small" effect="plain" round type="info">
                      {{ t }}
                    </el-tag>
                  </el-space>
                </el-space>
              </template>

              <template #trailing>
                <el-icon color="var(--el-text-color-placeholder)">
                  <ArrowRight />
                </el-icon>
              </template>
            </ListTile>
          </div>
        </div>

        <el-empty v-else :image-size="80"
          :description="search || activeTag ? 'No matching credentials' : 'No credentials found'">
          <el-button v-if="!search && !activeTag" type="primary" @click="goToCredentials">
            Go to Credentials
          </el-button>
        </el-empty>
      </div>
    </el-scrollbar>

    <template #footer>
      <el-row justify="space-between" align="middle">
        <el-text type="info" size="small">{{ filteredCredentials.length }} credentials available</el-text>
        <el-button @click="$emit('update:modelValue', false)">Cancel</el-button>
      </el-row>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search, ArrowRight, Key } from '@element-plus/icons-vue'
import { credentialsApi } from '@credentials/api'
import type { Credential } from '@credentials/api'
import ListTile from '@shared/components/ListTile.vue'

defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', val: boolean): void
  (e: 'select', val: Credential): void
}>()

const loading = ref(false)
const credentials = ref<Credential[]>([])
const search = ref('')
const activeTag = ref<string | null>(null)
const router = useRouter()

const goToCredentials = () => {
  emit('update:modelValue', false)
  router.push('/credentials')
}

const fetchCredentials = async () => {
  loading.value = true
  try {
    credentials.value = await credentialsApi.list()
  } catch (e) {
    console.error('Failed to load credentials', e)
  } finally {
    loading.value = false
  }
}

const uniqueTags = computed(() => {
  const tags = new Set<string>()
  credentials.value.forEach(c => {
    if (Array.isArray(c.tags)) {
      c.tags.forEach(t => tags.add(t))
    }
  })
  return Array.from(tags).sort()
})

const filteredCredentials = computed(() => {
  return credentials.value.filter(c => {
    // Tag filter
    if (activeTag.value && (!c.tags || !c.tags.includes(activeTag.value))) {
      return false
    }

    // Search filter
    if (search.value) {
      const s = search.value.toLowerCase()
      const matchName = c.name.toLowerCase().includes(s)
      const matchKey = c.key.toLowerCase().includes(s)
      const matchTags = c.tags?.some(t => t.toLowerCase().includes(s))
      return matchName || matchKey || matchTags
    }

    return true
  })
})

const toggleTag = (tag: string) => {
  if (activeTag.value === tag) {
    activeTag.value = null
  } else {
    activeTag.value = tag
  }
}

const handleSelect = (c: Credential) => {
  emit('select', c)
  emit('update:modelValue', false)
}

onMounted(() => {
  fetchCredentials()
})
</script>

<style scoped>
.credential-selector-dialog :deep(.el-dialog__body) {
  padding-top: 10px;
}

/* Ensure smooth hover on tiles */
:deep(.list-tile) {
  transition: all 0.2s ease;
}
</style>
