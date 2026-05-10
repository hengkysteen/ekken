<template>
  <div class="ek-dynamic-form">
    <ElRow v-for="(row, rowIdx) in form || []" :key="rowIdx" :gutter="16">
      <ElCol v-for="item in row" :key="item.key" :span="getSpan(item.flex, row)">
        <!-- Input components with v-model -->
        <component 
          v-if="isInputComponent(item.component)"
          :is="getComponent(item.component, item.key)" 
          v-model="formData[item.key]"
          :field="getField(item.key)"
          :item="normalizeItem(item)"
        />
        <!-- Non-input components without v-model -->
        <component 
          v-else
          :is="getComponent(item.component, item.key)" 
          :field="getField(item.key)"
          :item="normalizeItem(item)"
        />
      </ElCol>
    </ElRow>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElRow, ElCol } from 'element-plus'
import EkInputField from './fields/EkInputField.vue'
import EkNumberField from './fields/EkNumberField.vue'
import EkSelectField from './fields/EkSelectField.vue'
import EkTextareaField from './fields/EkTextareaField.vue'
import EkRadioField from './fields/EkRadioField.vue'
import EkSliderField from './fields/EkSliderField.vue'
import EkSwitchField from './fields/EkSwitchField.vue'
import EkJsonEditorField from './fields/EkJsonEditorField.vue'
import EkColorPickerField from './fields/EkColorPickerField.vue'
import EkDatePickerField from './fields/EkDatePickerField.vue'
import EkTimePickerField from './fields/EkTimePickerField.vue'
import EkTextDisplay from './fields/EkTextDisplay.vue'
import type { NodeField, Form } from '@workflows/node/types/node'

const props = defineProps<{
  form?: Form[][]
  fields?: NodeField[]
  modelValue: Record<string, any>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, any>]
}>()

const formData = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const componentMap: Record<string, any> = {
  input: EkInputField,
  number: EkNumberField,
  'number-s1': EkNumberField,
  'number-s2': EkNumberField,
  select: EkSelectField,
  textarea: EkTextareaField,
  radio: EkRadioField,
  slider: EkSliderField,
  switch: EkSwitchField,
  jsonEditor: EkJsonEditorField,
  colorPicker: EkColorPickerField,
  datePicker: EkDatePickerField,
  timePicker: EkTimePickerField,
  text: EkTextDisplay,
}

function getSpan(flex: number, row: Form[]): number {
  const total = row.reduce((sum, item) => sum + (item.flex || 0), 0)
  return Math.round((flex / total) * 24)
}

function getComponent(component: string | undefined, key: string): any {
  const field = getField(key)
  const compName = component || getDefaultComponent(field?.type)
  return componentMap[compName] || EkInputField
}

function getDefaultComponent(type?: string): string {
  if (type === 'number') return 'number'
  if (type === 'select') return 'select'
  return 'input'
}

function getField(key: string): any {
  const field = (props.fields || []).find((f) => f.key === key)
  const item = (props.form || []).flat().find((i) => i.key === key)
  
  // Merge item.form_options ke field
  if (field && item?.form_options) {
    return { ...field, ...item.form_options }
  }
  
  return field
}

function normalizeItem(item: Form): any {
  return {
    ...item,
    options: item.form_options || {}
  }
}

function isInputComponent(component?: string): boolean {
  // Non-input components yang tidak butuh v-model
  const nonInputComponents = ['text', 'button', 'divider', 'alert']
  return !nonInputComponents.includes(component || '')
}
</script>

<style scoped>
.ek-dynamic-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
