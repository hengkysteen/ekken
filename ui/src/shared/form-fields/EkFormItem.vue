<script setup lang="ts">
import { ref, provide } from 'vue'
import { ElMessage } from 'element-plus'
import CredentialSelector from '@credentials/components/CredentialSelector.vue'
import { request } from '@shared/api/request'
import type { Credential } from '@credentials/api'

const props = defineProps<{
  label?: string
  helper?: string
  nativeFilePicker?: boolean
  nativeDirectoryPicker?: boolean
  multiple?: boolean
  credentialPicker?: boolean
}>()

const emit = defineEmits<{
  'file-select': [path: string | string[]]
  'credential-select': [credential: Credential]
}>()

const showCredentialSelector = ref(false)

const handleFilePick = async () => {
  const title = props.multiple ? 'Choose Files' : 'Choose File'
  await executePick({ multiple: props.multiple, title })
}

const handleDirectoryPick = async () => {
  await executePick({ directory: true, multiple: props.multiple, title: 'Choose Folder' })
}

const executePick = async (options: { multiple?: boolean, directory?: boolean, title?: string }) => {
  try {
    const query = new URLSearchParams()
    if (options.multiple) query.append('multiple', 'true')
    if (options.directory) query.append('directory', 'true')

    const result = await request<string[]>(`/system/file-picker?${query.toString()}`, {
      method: 'POST'
    })

    if (result && result.length > 0) {
      if (props.multiple) {
        emit('file-select', result)
      } else {
        emit('file-select', result[0])
      }
    }
  } catch (e: any) {
    if (e.error === 'cancelled' || (e.message && e.message.toLowerCase().includes('cancel'))) {
      return
    }
    ElMessage.error(e.message || 'Failed to pick item')
  }
}

const handleCredentialPick = () => {
  showCredentialSelector.value = true
}

const handleCredentialSelect = (credential: Credential) => {
  emit('credential-select', credential)
}

// Provide picker actions to child input components
provide('ekFormItem', {
  nativeFilePicker: props.nativeFilePicker,
  nativeDirectoryPicker: props.nativeDirectoryPicker,
  credentialPicker: props.credentialPicker,
  onFilePick: handleFilePick,
  onDirectoryPick: handleDirectoryPick,
  onCredentialPick: handleCredentialPick
})
</script>

<template>
  <div class="ek-form-item">
    <label v-if="label" class="ek-form-item__label">{{ label }}</label>
    
    <slot />
    
    <span v-if="helper" class="ek-form-item__helper">{{ helper }}</span>

    <CredentialSelector 
      v-if="credentialPicker"
      v-model="showCredentialSelector"
      @select="handleCredentialSelect"
    />
  </div>
</template>

<style scoped>
.ek-form-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ek-form-item__label {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-regular);
}

.ek-form-item__helper {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
