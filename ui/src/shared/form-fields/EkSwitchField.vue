<template>
  <EkFormItem 
    :label="field?.label" 
    :helper="field?.helper"
  >
    <ElSwitch 
      class="ek-switch-field__input"
      :model-value="modelValue" 
      :disabled="field?.disabled"
      @update:model-value="$emit('update:modelValue', $event)"
    />
  </EkFormItem>
</template>

<script setup lang="ts">
import { ElSwitch } from 'element-plus'
import EkFormItem from './EkFormItem.vue'

interface FieldDef {
  key: string
  label?: string
  helper?: string
  disabled?: boolean
  [key: string]: any
}

defineProps<{
  modelValue?: string | number | boolean
  field?: FieldDef
  item?: any
}>()

defineEmits<{
  'update:modelValue': [value: string | number | boolean]
}>()
</script>

<style scoped>
.ek-switch-field__input {
  align-self: flex-start;
}
</style>
