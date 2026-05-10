<template>
  <EkFormItem 
    :label="field?.label" 
    :helper="field?.helper"
    :native-file-picker="field?.native_file_picker"
    :native-directory-picker="field?.native_file_picker_directory"
    :multiple="field?.native_file_picker_multiple"
    :credential-picker="field?.credential_picker"
    @file-select="handleFileSelect"
    @credential-select="handleCredentialSelect"
  >
    <EkInput 
      :model-value="modelValue" 
      :placeholder="field?.placeholder"
      :disabled="field?.disabled"
      @update:model-value="$emit('update:modelValue', $event)"
    />
  </EkFormItem>
</template>

<script setup lang="ts">
import EkFormItem from '../EkFormItem.vue'
import EkInput from '../EkInput.vue'
import type { Credential } from '@credentials/api'

interface FieldDef {
  key: string
  label?: string
  helper?: string
  placeholder?: string
  disabled?: boolean
  native_file_picker?: boolean
  native_file_picker_multiple?: boolean
  native_file_picker_directory?: boolean
  credential_picker?: boolean
  [key: string]: any
}

defineProps<{
  modelValue?: string
  field?: FieldDef
  item?: any
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

function handleFileSelect(path: string | string[]) {
  if (Array.isArray(path)) {
    emit('update:modelValue', path.join(', '))
  } else {
    emit('update:modelValue', path)
  }
}

function handleCredentialSelect(credential: Credential) {
  emit('update:modelValue', `{{ ${credential.key} }}`)
}
</script>
