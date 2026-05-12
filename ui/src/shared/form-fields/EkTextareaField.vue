<template>
  <EkFormItem 
    :label="field?.label" 
    :helper="field?.helper"
  >
    <ElInput 
      :model-value="modelValue" 
      type="textarea"
      :placeholder="field?.placeholder"
      :disabled="field?.disabled"
      :rows="field?.rows || 3"
      :maxlength="field?.maxlength"
      @update:model-value="$emit('update:modelValue', $event)"
    />
  </EkFormItem>
</template>

<script setup lang="ts">
import { ElInput } from 'element-plus'
import EkFormItem from './EkFormItem.vue'

interface FieldDef {
  key: string
  label?: string
  helper?: string
  placeholder?: string
  disabled?: boolean
  rows?: number
  maxlength?: number
  [key: string]: any
}

defineProps<{
  modelValue?: string
  field?: FieldDef
  item?: any
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>
