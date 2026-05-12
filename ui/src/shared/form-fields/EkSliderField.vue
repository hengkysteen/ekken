<template>
  <EkFormItem 
    :label="field?.label" 
    :helper="field?.helper"
  >
    <ElSlider 
      :model-value="modelValue" 
      :disabled="field?.disabled"
      :min="field?.min || 0"
      :max="field?.max || 100"
      :step="field?.step || 1"
      :show-input="field?.show_input"
      @update:model-value="$emit('update:modelValue', $event)"
    />
  </EkFormItem>
</template>

<script setup lang="ts">
import { ElSlider } from 'element-plus'
import EkFormItem from './EkFormItem.vue'

interface FieldDef {
  key: string
  label?: string
  helper?: string
  disabled?: boolean
  min?: number
  max?: number
  step?: number
  show_input?: boolean
  [key: string]: any
}

defineProps<{
  modelValue?: number | number[]
  field?: FieldDef
  item?: any
}>()

defineEmits<{
  'update:modelValue': [value: number | number[]]
}>()
</script>
