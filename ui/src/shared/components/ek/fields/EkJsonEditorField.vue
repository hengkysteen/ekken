<template>
  <EkFormItem 
    :label="field?.label" 
    :helper="field?.helper"
  >
    <SimpleJsonEditor 
      :model-value="modelValue || ''" 
      :height="field?.height || '200px'" 
      @update:model-value="emit('update:modelValue', $event)" 
    />
  </EkFormItem>
</template>

<script setup lang="ts">
import EkFormItem from '../EkFormItem.vue'
import SimpleJsonEditor from '@shared/components/SimpleJsonEditor.vue'

interface FieldDef {
  key: string
  label?: string
  helper?: string
  height?: string
  [key: string]: any
}

defineProps<{
  modelValue?: string
  field?: FieldDef
  item?: any
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>
