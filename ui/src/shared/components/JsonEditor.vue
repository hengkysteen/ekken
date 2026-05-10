<!-- <template>
  <div class="json-editor" :class="{ 'has-error': hasError && modelValue.trim() }">
    <div class="json-editor-toolbar">
      <div class="toolbar-content">
        <span class="toolbar-label">JSON</span>
        <el-space>
          <el-button link size="small" @click="format" :disabled="!modelValue" title="Format">
            <el-icon><Tools /></el-icon>
            Format
          </el-button>
          <el-button link size="small" @click="minify" :disabled="!modelValue" title="Minify">
            <el-icon><Download /></el-icon>
            Minify
          </el-button>
        </el-space>
      </div>
    </div>
    <div ref="editorRef" class="json-editor-container" :style="{ height: editorHeight }"></div>
    <el-alert v-if="hasError && modelValue.trim()" :title="errorMessage" type="error" :closable="false" style="margin-top: 4px" show-icon />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { EditorView, basicSetup } from 'codemirror'
import { json } from '@codemirror/lang-json'
import { Tools, Download } from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  modelValue: string
  height?: string
}>(), {
  height: '200px'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editorRef = ref<HTMLElement | null>(null)
let view: EditorView | null = null
const editorHeight = ref(props.height)

const hasError = ref(false)
const errorMessage = ref('')

const format = () => {
  try {
    const parsed = JSON.parse(props.modelValue)
    emit('update:modelValue', JSON.stringify(parsed, null, 2))
    hasError.value = false
    errorMessage.value = ''
  } catch (e: any) {
    // keep as-is
  }
}

const minify = () => {
  try {
    const parsed = JSON.parse(props.modelValue)
    emit('update:modelValue', JSON.stringify(parsed))
    hasError.value = false
    errorMessage.value = ''
  } catch (e: any) {
    // keep as-is
  }
}

// Build a CodeMirror theme that matches Element Plus
function elementPlusTheme() {
  return EditorView.theme({
    '&': {
      height: editorHeight.value,
      fontSize: '13px',
      fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', 'SF Mono', monospace",
      backgroundColor: 'var(--el-bg-color)',
      color: 'var(--el-text-color-primary)',
    },
    '.cm-scroller': {
      overflow: 'auto',
      lineHeight: '1.6',
    },
    '.cm-content': {
      padding: '8px',
      caretColor: 'var(--el-text-color-primary) !important',
    },
    '.cm-cursor': {
      borderLeftColor: 'var(--el-text-color-primary) !important',
    },
    '.cm-activeLine': {
      backgroundColor: 'var(--el-fill-color-light)',
    },
    '.cm-selectionBackground': {
      backgroundColor: 'var(--el-color-primary-light-7)',
    },
    '.cm-gutters': {
      backgroundColor: 'var(--el-fill-color-blank)',
      borderRight: '1px solid var(--el-border-color-lighter)',
      color: 'var(--el-text-color-placeholder)',
    },
    '.cm-gutterElement': {
      color: 'inherit',
    },
  })
}

// JSON syntax colors based on Element Plus colors
function elementPlusJsonTheme() {
  return EditorView.theme({
    '.cm-json-key': { color: 'var(--el-color-primary)' },
    '.cm-json-string': { color: 'var(--el-color-success)' },
    '.cm-json-number': { color: 'var(--el-color-warning)' },
    '.cm-json-true': { color: 'var(--el-color-info)' },
    '.cm-json-false': { color: 'var(--el-color-info)' },
    '.cm-json-null': { color: 'var(--el-text-color-placeholder)' },
  })
}

onMounted(() => {
  view = new EditorView({
    doc: props.modelValue,
    extensions: [
      basicSetup,
      json(),
      elementPlusTheme(),
      elementPlusJsonTheme(),
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          const newVal = update.state.doc.toString()
          emit('update:modelValue', newVal)
          try {
            if (newVal.trim()) {
              JSON.parse(newVal)
              hasError.value = false
              errorMessage.value = ''
            } else {
              hasError.value = false
              errorMessage.value = ''
            }
          } catch (e: any) {
            hasError.value = true
            errorMessage.value = e.message
          }
        }
      }),
    ],
    parent: editorRef.value || undefined,
  })
})

watch(() => props.modelValue, (newVal) => {
  if (view && newVal !== view.state.doc.toString()) {
    if (!view.hasFocus) {
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: newVal },
      })
    }
  }
})

watch(() => props.height, (newH) => {
  editorHeight.value = newH
  if (view) {
    view.dom.style.height = newH
  }
})

onBeforeUnmount(() => {
  if (view) {
    view.destroy()
    view = null
  }
})
</script>

<style scoped>
.json-editor {
  display: flex;
  flex-direction: column;
  gap: 0;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  overflow: hidden;
}

.json-editor.has-error {
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

.json-editor-container {
  min-height: 100px;
}

:deep(.cm-editor) {
  border: none !important;
  outline: none !important;
}

:deep(.cm-focused) {
  outline: none !important;
}

.json-editor-container {
  position: relative;
}
</style> -->
