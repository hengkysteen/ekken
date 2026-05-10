<template>
  <EkFormItem 
    :label="field?.label" 
    :helper="field?.helper"
  >
    <ElTimePicker 
      :model-value="modelValue" 
      :placeholder="field?.placeholder"
      :disabled="field?.disabled"
      :format="field?.format"
      :value-format="field?.value_format"
      style="width: 100%"
      @update:model-value="$emit('update:modelValue', $event)"
    />
  </EkFormItem>
</template>

<script setup lang="ts">
import { ElTimePicker } from 'element-plus'
import EkFormItem from '../EkFormItem.vue'

interface FieldDef {
  key: string
  label?: string
  helper?: string
  placeholder?: string
  disabled?: boolean
  format?: string
  value_format?: string
  [key: string]: any
}

defineProps<{
  modelValue?: string | Date
  field?: FieldDef
  item?: any
}>()

defineEmits<{
  'update:modelValue': [value: string | Date]
}>()
</script>
