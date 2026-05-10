<template>
  <EkFormItem 
    :label="field?.label" 
    :helper="field?.helper"
  >
    <ElColorPicker 
      :model-value="modelValue" 
      :disabled="field?.disabled"
      :show-alpha="field?.show_alpha"
      @update:model-value="$emit('update:modelValue', $event || '')"
    />
  </EkFormItem>
</template>

<script setup lang="ts">
import { ElColorPicker } from 'element-plus'
import EkFormItem from '../EkFormItem.vue'

interface FieldDef {
  key: string
  label?: string
  helper?: string
  disabled?: boolean
  show_alpha?: boolean
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
