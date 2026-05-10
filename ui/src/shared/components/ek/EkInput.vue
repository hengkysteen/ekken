<script setup lang="ts">
import { inject } from 'vue'
import { ElInput, ElIcon, ElDropdown, ElDropdownMenu, ElDropdownItem } from 'element-plus'
import { Folder, Key, Document, ArrowDown } from '@element-plus/icons-vue'

defineProps<{
  modelValue?: string | number
  type?: string
  placeholder?: string
  disabled?: boolean
  clearable?: boolean
  showPassword?: boolean
  maxlength?: number
  rows?: number
}>()

defineEmits<{
  'update:modelValue': [value: string | number]
}>()

const formItem = inject<any>('ekFormItem', null)
</script>

<template>
  <ElInput
    :model-value="modelValue"
    :type="type"
    :placeholder="placeholder"
    :disabled="disabled"
    :clearable="clearable"
    :show-password="showPassword"
    :maxlength="maxlength"
    :rows="rows"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <template v-if="$slots.prefix" #prefix>
      <slot name="prefix" />
    </template>
    
    <template #suffix>
      <slot name="suffix" />
      
      <!-- Dropdown for multiple picker types -->
      <template v-if="formItem?.nativeFilePicker && formItem?.nativeDirectoryPicker">
        <ElDropdown trigger="click" @command="(cmd) => cmd === 'file' ? formItem.onFilePick() : formItem.onDirectoryPick()">
          <ElIcon style="cursor: pointer; margin-left: 8px;">
            <Folder />
          </ElIcon>
          <template #dropdown>
            <ElDropdownMenu>
              <ElDropdownItem command="file" :icon="Document">Select File</ElDropdownItem>
              <ElDropdownItem command="folder" :icon="Folder">Select Folder</ElDropdownItem>
            </ElDropdownMenu>
          </template>
        </ElDropdown>
      </template>

      <!-- Single File Picker -->
      <template v-else-if="formItem?.nativeFilePicker">
        <ElIcon style="cursor: pointer; margin-left: 8px;" @click="formItem.onFilePick">
          <Document />
        </ElIcon>
      </template>

      <!-- Single Directory Picker -->
      <template v-else-if="formItem?.nativeDirectoryPicker">
        <ElIcon style="cursor: pointer; margin-left: 8px;" @click="formItem.onDirectoryPick">
          <Folder />
        </ElIcon>
      </template>

      <ElIcon v-if="formItem?.credentialPicker" style="cursor: pointer; margin-left: 8px;" @click="formItem.onCredentialPick">
        <Key />
      </ElIcon>
    </template>
  </ElInput>
</template>
