<template>
  <el-dialog v-model="isVisible" :title="`Configure : ${nodeLabel}`" class="node-config-dialog"
    @closed="$emit('after-leave')">
    <div class="node-config-base">
      <div class="header-info">
        <img v-if="nodeIconUrl" :src="nodeIconUrl" class="config-icon-img" :style="iconStyle" width="32" height="32" />
        <div class="label-info">
          <span class="node-label-text">{{ nodeLabel }}</span>
          <span class="node-type-text">{{ nodeTypeDisplay }}</span>
        </div>
      </div>

      <el-divider />

      <div class="specific-config-container">
        <component :is="specificComponent" v-if="node" ref="nodeCompRef" :node="node" />
      </div>

      <template v-if="nodeTypeDisplay !== 'timer' && currentActionHasOutput">
        <el-divider />
        <el-form label-position="top">
          <el-form-item label="Result name">
            <el-input v-model="responseVar" placeholder="e.g. my_data" />
            <div class="form-item-hint">
              Save this node's output to a variable.
            </div>
          </el-form-item>
        </el-form>
      </template>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="onClose">Cancel</el-button>
        <el-button type="primary" @click="onSave">
          <el-icon>
            <Check />
          </el-icon>
          <span class="btn-text">Save</span>
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, shallowRef, provide } from 'vue'
import { ElMessage } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import { getNodeFormComponent } from '../registry'
import { validateNodeConfig } from '@workflows/node/utils/validation'

const props = defineProps({
  visible: { type: Boolean, default: false },
  node: { type: Object, default: null },
  catalog: { type: Array, default: () => [] },
})

provide('nodeCatalog', props.catalog)

const emit = defineEmits(['update:visible', 'save', 'close', 'after-leave'])

const nodeCompRef = ref(null)
const specificComponent = shallowRef(null)
const responseVar = ref('')

const isVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value),
})

const nodeLabel = computed(() => props.node?.data?.label || 'Node')
const nodeTypeDisplay = computed(() => props.node?.data?.nodeType || props.node?.type || 'unknown')
const nodeIconUrl = computed(() => props.node?.data?.icon || null)

import { useTheme } from '../../../../../shared/composables/useTheme'
const { isDarkMode: isDark } = useTheme()

const iconStyle = computed(() => ({
  filter: isDark.value ? 'invert(1) brightness(1.5)' : 'none'
}))

const currentActionHasOutput = computed(() => {
  if (!props.node) return false
  const action = props.node.data?.action
  return action?.has_response || false
})

watch(() => props.node, (newNode) => {
  if (newNode) {
    const type = newNode.data?.nodeType || newNode.type
    specificComponent.value = getNodeFormComponent(type)
    responseVar.value = newNode.data?.action?.response_var || ''
  }
}, { immediate: true })

// Sync responseVar when child component's action changes (e.g., action switch)
watch(() => {
  if (nodeCompRef.value?.getData) {
    return nodeCompRef.value.getData()?.response_var
  }
  return undefined
}, (newResponseVar) => {
  if (newResponseVar) {
    responseVar.value = newResponseVar
  }
})

function onClose() {
  emit('close')
  isVisible.value = false
}

function onSave() {
  let finalAction = null
  if (nodeCompRef.value?.getData) {
    finalAction = nodeCompRef.value.getData()
  }

  if (!finalAction) {
    ElMessage.error("Failed to collect node configuration")
    return
  }

  // Ensure response_var is saved inside the action object
  finalAction.response_var = responseVar.value

  const nodeType = nodeTypeDisplay.value
  if (nodeType !== 'timer') {
    const nodeDef = props.catalog.find(n => n.type === nodeType)

    if (nodeDef) {
      // Note: We might need to update validateNodeConfig to handle NodeAction structure
      const validation = validateNodeConfig(finalAction, nodeDef)
      if (!validation.valid) {
        ElMessage.error({
          message: `Invalid configuration:\n${validation.errors.join('\n')}`,
          duration: 5000
        })
        return
      }
    }
  }

  emit('save', {
    id: props.node?.id,
    action: finalAction,
    label: nodeLabel.value
  })

  isVisible.value = false
}
</script>

<style scoped>
.node-config-base {
  padding: 0 4px;
}

.header-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: -8px;
}

.config-icon-img {
  object-fit: contain !important;
}

.label-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.node-label-text {
  font-weight: bold;
  font-size: 16px;
  color: var(--el-text-color-primary);
}

.node-type-text {
  font-size: 11px;
  text-transform: uppercase;
  color: var(--el-text-color-secondary);
  letter-spacing: 0.5px;
}

.specific-config-container {
  min-height: 100px;
}

.form-item-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
  margin-top: 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn-text {
  margin-left: 6px;
}

:deep(.el-divider--horizontal) {
  margin: 20px 0;
}

:deep(.el-dialog__body) {
  padding-top: 10px;
}
</style>
