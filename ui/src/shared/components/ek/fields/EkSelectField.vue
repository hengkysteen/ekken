<template>
  <EkFormItem 
    :label="field?.label" 
    :helper="field?.helper"
  >
    <ElSelect 
      :model-value="modelValue" 
      :placeholder="field?.placeholder"
      :disabled="field?.disabled"
      :clearable="field?.clearable"
      style="width: 100%"
      @update:model-value="$emit('update:modelValue', $event)"
    >
      <ElOption 
        v-for="opt in options" 
        :key="opt.value" 
        :label="opt.label" 
        :value="opt.value" 
      />
    </ElSelect>
  </EkFormItem>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElSelect, ElOption } from 'element-plus'
import EkFormItem from '../EkFormItem.vue'

interface FieldDef {
  key: string
  label?: string
  helper?: string
  placeholder?: string
  disabled?: boolean
  clearable?: boolean
  options?: string[] | Array<{ label: string; value: string }>
  [key: string]: any
}

const props = defineProps<{
  modelValue?: string
  field?: FieldDef
  item?: any
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()

const options = computed(() => {
  const opts = props.field?.options || []
  if (opts.length > 0 && typeof opts[0] === 'string') {
    return (opts as string[]).map((o) => ({ label: o, value: o }))
  }
  return opts as Array<{ label: string; value: string }>
})
</script>
