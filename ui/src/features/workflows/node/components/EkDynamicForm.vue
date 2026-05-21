<template>
  <div class="ek-dynamic-form">
    <ElRow v-for="(row, rowIdx) in layout || []" :key="rowIdx" :gutter="16">
      <ElCol v-for="item in row" :key="item.key" :span="getSpan(item.flex, row)">
        <!-- Input components with v-model -->
        <component v-if="isInputComponent(item.component)" :is="getComponent(item.component, item.key)"
          :model-value="modelValue ? modelValue[item.key] : getFieldFromAll(item.key).value"
          @update:model-value="updateField(item.key, $event)"
          :field="{ ...getFieldFromAll(item.key), value: (modelValue ? modelValue[item.key] : getFieldFromAll(item.key).value), ...(item.options || {}) }"
          :item="normalizeItem(item)" />
        <!-- Non-input components without v-model -->
        <component v-else :is="getComponent(item.component, item.key)"
          :field="{ ...getFieldFromAll(item.key), value: (modelValue ? modelValue[item.key] : getFieldFromAll(item.key).value), ...(item.options || {}) }"
          :item="normalizeItem(item)" />
      </ElCol>
    </ElRow>

    <!-- Render global fields not in form layout -->
    <ElCollapse v-if="unrenderedGlobalFields.length > 0" class="advanced-settings">
      <ElCollapseItem title="Advanced Settings" name="advanced">
        <ElRow v-for="field in unrenderedGlobalFields" :key="field.key" :gutter="16">
          <ElCol :span="24">
            <component :is="getComponent(undefined, field.key)"
              :model-value="modelValue ? modelValue[field.key] : field.value"
              @update:model-value="updateGlobalField(field.key, $event)"
              :field="{ ...field, value: (modelValue ? modelValue[field.key] : field.value) }"
              :item="{ key: field.key, component: getDefaultComponent(field.type) }" />
          </ElCol>
        </ElRow>
      </ElCollapseItem>
    </ElCollapse>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElRow, ElCol, ElCollapse, ElCollapseItem } from 'element-plus'
import EkInputField from '@shared/form-fields/EkInputField.vue'
import EkNumberField from '@shared/form-fields/EkNumberField.vue'
import EkSelectField from '@shared/form-fields/EkSelectField.vue'
import EkTextareaField from '@shared/form-fields/EkTextareaField.vue'
import EkRadioField from '@shared/form-fields/EkRadioField.vue'
import EkSliderField from '@shared/form-fields/EkSliderField.vue'
import EkSwitchField from '@shared/form-fields/EkSwitchField.vue'
import EkJsonEditorField from '@shared/form-fields/EkJsonEditorField.vue'
import EkColorPickerField from '@shared/form-fields/EkColorPickerField.vue'
import EkDatePickerField from '@shared/form-fields/EkDatePickerField.vue'
import EkTimePickerField from '@shared/form-fields/EkTimePickerField.vue'
import EkTextDisplay from '@shared/form-fields/EkTextDisplay.vue'
import type { NodeField, AutoLayout } from '@workflows/node/types/node'

const props = defineProps<{
  modelValue?: Record<string, any>
  layout?: AutoLayout[][]
  fields?: NodeField[]
  globalFields?: NodeField[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, any>]
  'update:field': [key: string, value: any]
  'update:global-field': [key: string, value: any]
}>()

function updateField(key: string, value: any) {
  if (props.modelValue) {
    emit('update:modelValue', { ...props.modelValue, [key]: value })
  }
  emit('update:field', key, value)
}

function updateGlobalField(key: string, value: any) {
  if (props.modelValue) {
    emit('update:modelValue', { ...props.modelValue, [key]: value })
  }
  emit('update:global-field', key, value)
}

// Merge global fields with action fields for rendering
const allFields = computed(() => {
  const fields = [...(props.fields || [])]

  if (props.globalFields) {
    props.globalFields.forEach(gf => {
      if (!fields.find(f => f.key === gf.key)) {
        fields.push(gf)
      }
    })
  }

  return fields
})

// Find global fields that are not in form layout
const unrenderedGlobalFields = computed(() => {
  if (!props.globalFields || !props.layout) return props.globalFields || []

  const renderedKeys = new Set<string>()
  props.layout.forEach(row => {
    row.forEach(item => renderedKeys.add(item.key))
  })

  return props.globalFields.filter(f => !renderedKeys.has(f.key))
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

function getSpan(flex: number, row: AutoLayout[]): number {
  const total = row.reduce((sum, item) => sum + (item.flex || 0), 0)
  return Math.round((flex / total) * 24)
}

function getComponent(component: string | undefined, key: string): any {
  const field = getFieldFromAll(key)
  const compName = component || getDefaultComponent(field?.type)
  return componentMap[compName] || EkInputField
}

function getDefaultComponent(type?: string): string {
  if (type === 'number') return 'number'
  if (type === 'select') return 'select'
  return 'input'
}

function getFieldFromAll(key: string): any {
  const field = allFields.value.find((f) => f.key === key)
  if (!field) return { key, type: 'string', label: key, value: '' }

  return field.value === undefined ? { ...field, value: '' } : field
}

function normalizeItem(item: AutoLayout): any {
  return {
    ...item,
    options: item.options || {}
  }
}

function isInputComponent(component?: string): boolean {
  // Non-input components that don't need v-model
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
