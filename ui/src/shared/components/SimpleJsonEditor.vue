<template>
  <div class="simple-json-editor" :class="{ 'has-error': hasError && modelValue.trim() }">
    <div class="json-editor-toolbar">
      <div class="toolbar-content">
        <span class="toolbar-label">JSON</span>
        <el-space>
          <el-button :icon="Document" link size="small" @click="format" :disabled="!modelValue" title="Pretty">

            Pretty
          </el-button>
          <el-button :icon="Tools" link size="small" @click="minify" :disabled="!modelValue" title="Minify">

            Minify
          </el-button>
        </el-space>
      </div>
    </div>
    <textarea ref="textareaRef" :value="modelValue" @input="handleInput" @keydown="handleKeydown" class="json-textarea"
      :style="{ height }" spellcheck="false" />
    <el-alert v-if="hasError && modelValue.trim()" :title="errorMessage" type="error" :closable="false"
      style="margin-top: 4px" show-icon />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Tools, Document } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  modelValue: string
  height?: string
}>(), {
  height: '200px'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()


const hasError = ref(false)
const errorMessage = ref('')

function handleInput(e: Event) {
  const value = (e.target as HTMLTextAreaElement).value
  emit('update:modelValue', value)
  validateJson(value)
}

function handleKeydown(e: KeyboardEvent) {
  const textarea = e.target as HTMLTextAreaElement

  // Auto-closing brackets
  const pairs: Record<string, string> = {
    '{': '}',
    '[': ']',
    '"': '"'
  }

  if (pairs[e.key]) {
    e.preventDefault()
    const { selectionStart, selectionEnd, value } = textarea
    const selected = value.substring(selectionStart, selectionEnd)
    const newValue = value.substring(0, selectionStart) + e.key + selected + pairs[e.key] + value.substring(selectionEnd)
    emit('update:modelValue', newValue)

    setTimeout(() => {
      textarea.selectionStart = textarea.selectionEnd = selectionStart + 1 + selected.length
    }, 0)
    return
  }

  // Auto-indent on Enter
  if (e.key === 'Enter') {
    const { selectionStart, value } = textarea
    const lineStart = value.lastIndexOf('\n', selectionStart - 1) + 1
    const line = value.substring(lineStart, selectionStart)
    const indent = line.match(/^\s*/)?.[0] || ''

    // Check if previous char is { or [
    const prevChar = value[selectionStart - 1]
    const nextChar = value[selectionStart]
    const extraIndent = (prevChar === '{' || prevChar === '[') ? '  ' : ''

    // If between brackets, add extra line
    if ((prevChar === '{' && nextChar === '}') || (prevChar === '[' && nextChar === ']')) {
      e.preventDefault()
      const newValue = value.substring(0, selectionStart) + '\n' + indent + extraIndent + '\n' + indent + value.substring(selectionStart)
      emit('update:modelValue', newValue)

      setTimeout(() => {
        textarea.selectionStart = textarea.selectionEnd = selectionStart + 1 + indent.length + extraIndent.length
      }, 0)
      return
    }

    e.preventDefault()
    const newValue = value.substring(0, selectionStart) + '\n' + indent + extraIndent + value.substring(selectionStart)
    emit('update:modelValue', newValue)

    setTimeout(() => {
      textarea.selectionStart = textarea.selectionEnd = selectionStart + 1 + indent.length + extraIndent.length
    }, 0)
  }

  // Tab key for indent
  if (e.key === 'Tab') {
    e.preventDefault()
    const { selectionStart, selectionEnd, value } = textarea
    const newValue = value.substring(0, selectionStart) + '  ' + value.substring(selectionEnd)
    emit('update:modelValue', newValue)

    setTimeout(() => {
      textarea.selectionStart = textarea.selectionEnd = selectionStart + 2
    }, 0)
  }
}

function validateJson(value: string) {
  if (!value.trim()) {
    hasError.value = false
    errorMessage.value = ''
    return
  }

  try {
    JSON.parse(value)
    hasError.value = false
    errorMessage.value = ''
  } catch (e: any) {
    hasError.value = true
    errorMessage.value = e.message
  }
}

function format() {
  try {
    const parsed = JSON.parse(props.modelValue)
    emit('update:modelValue', JSON.stringify(parsed, null, 2))
    hasError.value = false
    errorMessage.value = ''
  } catch (e: any) {
    // keep as-is
  }
}

function minify() {
  try {
    const parsed = JSON.parse(props.modelValue)
    emit('update:modelValue', JSON.stringify(parsed))
    hasError.value = false
    errorMessage.value = ''
  } catch (e: any) {
    // keep as-is
  }
}

watch(() => props.modelValue, (val) => {
  validateJson(val)
}, { immediate: true })
</script>

<style scoped>
.simple-json-editor {
  display: flex;
  flex-direction: column;
  gap: 0;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  overflow: hidden;
}

.simple-json-editor.has-error {
  border-color: var(--el-color-danger);
}

.json-editor-toolbar {
  padding: 4px 12px;
  background: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color);
}

.toolbar-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.json-textarea {
  width: 100%;
  min-height: 100px;
  padding: 8px;
  border: none;
  outline: none;
  resize: vertical;
  font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', 'SF Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
  background: var(--el-bg-color);
  color: var(--el-text-color-primary);
}

.json-textarea:focus {
  outline: none;
}
</style>
