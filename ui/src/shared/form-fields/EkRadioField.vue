<template>
  <EkFormItem 
    :label="field?.label" 
    :helper="field?.helper"
  >
    <ElRadioGroup 
      :model-value="modelValue" 
      :disabled="field?.disabled"
      @update:model-value="$emit('update:modelValue', $event)"
    >
      <ElRadio 
        v-for="opt in options" 
        :key="opt.value" 
        :value="opt.value"
      >
        {{ opt.label }}
      </ElRadio>
    </ElRadioGroup>
  </EkFormItem>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElRadioGroup, ElRadio } from 'element-plus'
import EkFormItem from './EkFormItem.vue'

interface FieldDef {
  key: string
  label?: string
  helper?: string
  disabled?: boolean
  options?: string[] | Array<{ label: string; value: string }>
  [key: string]: any
}

const props = defineProps<{
  modelValue?: string | number | boolean
  field?: FieldDef
  item?: any
}>()

defineEmits<{
  'update:modelValue': [value: string | number | boolean | undefined]
}>()

const options = computed(() => {
  const opts = props.field?.options || []
  if (opts.length > 0 && typeof opts[0] === 'string') {
    return (opts as string[]).map((o) => ({ label: o, value: o }))
  }
  return opts as Array<{ label: string; value: string }>
})
</script>
