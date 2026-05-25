<template>
  <el-dialog v-model="isVisible" title="Configure" width="520px" class="node-config-dialog"
    @closed="$emit('after-leave')">
    <div>
      <el-space alignment="center" :size="12">
        <el-image v-if="nodeIconUrl" :src="nodeIconUrl" fit="contain" :style="nodeIconStyle" />
        <el-space direction="vertical" alignment="start" :size="2">
          <el-text tag="strong">{{ nodeLabel }}</el-text>
          <el-text tag="span" size="small" type="info">{{ nodeTypeDisplay }}</el-text>
          <el-text v-if="nodeDescription" tag="span" size="small" type="info">{{ nodeDescription }}</el-text>
        </el-space>
      </el-space>
      <el-divider />
      <el-alert v-if="node?.data?.needsReview" type="warning" :closable="false" title="Version mismatch" class="mb-4">
        loaded {{ formatVersion(node.data.version) }}, installed {{ formatVersion(node.data.installedVersion) }}.
        Please check your data.
      </el-alert>
      <div>
        <component :is="specificComponent" v-if="node" ref="nodeCompRef" :node="node"
          @action-change="handleActionChange" />
      </div>
      <template v-if="currentActionHasResponse">
        <el-form class="mt-4" label-position="top">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="Response variable">
                <el-input v-model="responseVar" placeholder="e.g. my_data" />
                <el-text tag="div" size="small" type="info">
                  Save this node's response to a variable.
                </el-text>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="Response type">
                <el-input :model-value="responseTypeDisplay" readonly disabled />
                <el-text tag="div" size="small" type="info">
                  The media type of this action's response.
                </el-text>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </template>
    </div>
    <template #footer>
      <el-space class="mt-4">
        <el-button @click="onClose">Cancel</el-button>
        <el-button type="primary" :icon="Check" @click="onSave">Save Configure</el-button>
      </el-space>
    </template>
  </el-dialog>
</template>
<script setup>
import { ref, computed, watch, shallowRef, provide } from 'vue'
import { ElMessage } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import { getNodeFormComponent } from '@workflows/node/nodes/registry'
import { validateNodeConfig } from '@workflows/node/utils/validation'
import { getActionBlueprint, getActionType } from '@workflows/node/utils/node'
import { useTheme } from '@shared/composables/useTheme'
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
const nodeDescription = computed(() => props.node?.data?.description || '')
const { isDarkMode: isDark } = useTheme()
const nodeIconStyle = computed(() => ({
  width: '32px',
  height: '32px',
  filter: isDark.value ? 'invert(1) brightness(1.5)' : 'none',
}))
const activeAction = ref(null)
function handleActionChange(action) {
  activeAction.value = action
}
const activeActionType = computed(() => {
  return activeAction.value?.type || getActionType(props.node?.data?.action)
})
const currentActionHasResponse = computed(() => {
  if (!props.node) return false
  const nodeDef = props.catalog.find(n => n.type === nodeTypeDisplay.value)
  return getActionBlueprint(nodeDef, activeActionType.value)?.has_response || false
})
const currentActionResponseType = computed(() => {
  if (!props.node) return null
  const nodeDef = props.catalog.find(n => n.type === nodeTypeDisplay.value)
  return getActionBlueprint(nodeDef, activeActionType.value)?.response_type || null
})
const responseTypeDisplay = computed(() => {
  const rt = currentActionResponseType.value
  if (!rt) return 'dynamic'
  const parts = []
  if (rt.mime) parts.push(rt.mime)
  if (rt.charset) parts.push(`charset=${rt.charset}`)
  if (rt.encoding) parts.push(`encoding=${rt.encoding}`)
  return parts.join('; ') || 'dynamic'
})
watch(() => props.node, (newNode) => {
  if (newNode) {
    const type = newNode.data?.nodeType || newNode.type
    specificComponent.value = getNodeFormComponent(type)
    responseVar.value = newNode.data?.action?.response_var || ''
  }
}, { immediate: true })
// Sync responseVar when child component's action changes (e.g., action switch)
watch(() => activeAction.value?.response_var, (newResponseVar) => {
  responseVar.value = newResponseVar || ''
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
  const nodeDef = props.catalog.find(n => n.type === nodeType)
  if (nodeDef) {
    const validation = validateNodeConfig(finalAction, nodeDef)
    if (!validation.valid) {
      ElMessage.error({
        message: `Invalid configuration:\n${validation.errors.join('\n')}`,
        duration: 5000
      })
      return
    }
  }
  emit('save', {
    id: props.node?.id,
    action: finalAction,
    label: nodeLabel.value
  })
  isVisible.value = false
}
function formatVersion(version) {
  if (!version) return 'unknown version'
  return String(version).startsWith('v') ? String(version) : `v${version}`
}
</script>