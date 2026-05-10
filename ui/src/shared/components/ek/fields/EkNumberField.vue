<template>
  <EkFormItem 
    :label="field?.label" 
    :helper="field?.helper"
  >
    <ElInputNumber 
      :model-value="modelValue" 
      :placeholder="field?.placeholder"
      :disabled="field?.disabled"
      :min="field?.min"
      :max="field?.max"
      :step="field?.step || 1"
      :controls-position="controlsPosition"
      style="width: 100%"
      @update:model-value="$emit('update:modelValue', $event)"
    />
  </EkFormItem>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElInputNumber } from 'element-plus'
import EkFormItem from '../EkFormItem.vue'

interface FieldDef {
  key: string
  label?: string
  helper?: string
  placeholder?: string
  disabled?: boolean
  min?: number
  max?: number
  step?: number
  [key: string]: any
}

const props = defineProps<{
  modelValue?: number
  field?: FieldDef
  item?: any
}>()

defineEmits<{
  'update:modelValue': [value: number | undefined]
}>()

const controlsPosition = computed(() => {
  // number-s1 = controls right, number-s2 = controls default (both sides)
  if (props.item?.component === 'number-s1') return 'right'
  return undefined
})
</script>
