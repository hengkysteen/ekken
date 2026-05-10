<template>
  <div class="node-card-wrapper" :class="{ selected: isSelected }" @contextmenu.prevent.stop="onContextMenu">
    <NodeContextMenu v-model:show="showMenu" :x="menuX" :y="menuY" :node-type="data.nodeType" :index="props.index"
      :disabled-keys="disabledKeys" @select="handleMenuSelect" />
    <!-- Input Handle -->
    <Handle v-if="!data.tags?.includes('Trigger')" id="input" type="target" :position="Position.Left"
      class="handle-input" />
    <div class="node-card" @click="handleClick">
      <!-- "MY NODE" Badge -->
      <div v-if="data.sourceType === 'mynodes'" class="mynodes-badge">
        {{ data.name || 'MY NODE' }}
      </div>
      <div class="node-header">
        <div class="icon-wrapper">
          <el-image v-if="data.icon" :src="data.icon" class="node-icon" fit="contain" :style="iconStyle" />
          <el-icon v-else class="node-icon-placeholder">
            <Grid />
          </el-icon>
        </div>
        <div class="header-info">
          <div class="node-label">{{ data.label }}</div>
          <div class="node-tag">{{ data.tags ? data.tags[0] : '' }}</div>
        </div>
      </div>
      <div class="node-body">
        <div class="action-section">
          <span class="action-name">{{ data.config?.action || data.nodeType }}</span>
        </div>
      </div>
    </div>
    <!-- Output Handles -->
    <Handle v-for="(output) in data.outputs" :key="output.key" :id="output.key" type="source"
      :position="output.tone === 'error' ? Position.Bottom : Position.Right" :style="getOutputHandleStyle(output)"
      :class="['handle-output', output.tone || 'success']" @click.stop="onAddNode(output.key)">
      <el-icon class="handle-add-icon">
        <Plus />
      </el-icon>
    </Handle>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import { Plus, Grid } from '@element-plus/icons-vue'
import NodeContextMenu from './NodeContextMenu.vue'
const props = defineProps<{
  id: string
  data: any
  selected?: boolean
  index?: number
}>()
const emit = defineEmits<{
  (e: 'node-click', data: { node: any; originalEvent?: MouseEvent }): void
  (e: 'configure', data: { id: string } & any): void
  (e: 'delete', data: { id: string }): void
  (e: 'duplicate', data: { id: string }): void
  (e: 'add-node', data: { id: string; handleType: string }): void
  (e: 'save-to-mynodes', data: { id: string; data: any }): void
}>()
const showMenu = ref(false)
const menuX = ref(0)
const menuY = ref(0)
const isSelected = computed(() => props.selected)
// Close menu when node is deselected
watch(isSelected, (selected) => {
  if (!selected) {
    showMenu.value = false
  }
})
import { useTheme } from '../../../../shared/composables/useTheme'
const { isDarkMode: isDark } = useTheme()
const iconStyle = computed(() => ({
  filter: isDark.value ? 'invert(1) brightness(1.5)' : 'none'
}))
const disabledKeys = computed(() => {
  return []
})
const getOutputHandleStyle = (output: any) => {
  if (output.tone === 'error') return {}
  const validOutputs = props.data.outputs.filter((o: any) => o.tone !== 'error')
  const index = validOutputs.findIndex((o: any) => o.key === output.key)
  const count = validOutputs.length
  if (count === 1) return { top: '50%' }
  const spacing = 100 / (count + 1)
  return { top: `${(index + 1) * spacing}%` }
}
const handleClick = (e: MouseEvent) => {
  emit('node-click', { node: { id: props.id, ...props.data }, originalEvent: e })
}
const handleMenuSelect = (key: string) => {
  if (key === 'configure') {
    emit('configure', { id: props.id, ...props.data })
  } else if (key === 'delete') {
    emit('delete', { id: props.id })
  } else if (key === 'duplicate') {
    emit('duplicate', { id: props.id })
  } else if (key === 'save_to_mynodes') {
    emit('save-to-mynodes', { id: props.id, data: props.data })
  }
}
function onContextMenu(e: MouseEvent) {
  menuX.value = e.clientX
  menuY.value = e.clientY
  showMenu.value = true
  emit('node-click', { node: { id: props.id, ...props.data }, originalEvent: e })
}
const onAddNode = (handleType: string) => emit('add-node', { id: props.id, handleType })
</script>
<style scoped>
.node-card-wrapper {
  position: relative;
  width: 180px;
  height: 140px;
}

.node-card {
  position: relative;
  background-color: var(--el-bg-color-overlay);
  border: 1px solid var(--el-border-color-dark);
  border-radius: 12px;
  overflow: hidden;
  box-shadow: var(--el-box-shadow-lighter);
  transition: border-color 0.2s, box-shadow 0.2s;
  cursor: pointer;
  user-select: none;
  width: 180px;
  height: 140px;
  display: flex;
  flex-direction: column;
}

.node-card-wrapper.selected .node-card {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 2px var(--el-color-primary-light-8), var(--el-box-shadow-light);
}

.node-header {
  padding-left: 12px;
  display: flex;
  align-items: center;
  gap: 10px;
  background-color: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
  height: 70px;
}

.icon-wrapper {
  width: 32px;
  height: 32px;
  background-color: var(--el-bg-color);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 0 0 1px var(--el-border-color-lighter);
  flex-shrink: 0;
}

.node-icon {
  width: 20px;
  height: 20px;
}

.node-icon-placeholder {
  font-size: 18px;
  color: var(--el-text-color-placeholder);
}

.header-info {
  flex: 1;
  min-width: 0;
}

.node-label {
  font-size: 13px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.2;
}

.node-tag {
  font-size: 10px;
  color: var(--el-text-color-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 500;
  margin-top: 2px;
}

.node-body {
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.action-section {
  display: flex;
  align-items: center;
}

.action-name {
  font-size: 11px;
  color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
  text-transform: uppercase;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-save-info {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 6px;
  background-color: var(--el-color-success-light-9);
  border-radius: 4px;
}

.save-icon {
  font-size: 11px;
  color: var(--el-color-success);
}

.save-target {
  font-size: 11px;
  font-weight: 600;
  color: var(--el-color-success);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Handles Styling */
.handle-input {
  width: 14px;
  height: 14px;
  background: var(--el-color-primary);
  border: 3px solid var(--el-bg-color-overlay);
  box-shadow: 0 0 0 1px var(--el-border-color-light);
  left: -7px;
  top: 50%;
  transform: translateY(-50%);
  z-index: 100;
}

.handle-output {
  width: 24px;
  height: 24px;
  right: -12px;
  border: 4px solid var(--el-bg-color-overlay);
  box-shadow: var(--el-box-shadow-lighter);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  transition: all 0.2s;
  transform: translateY(-50%);
  z-index: 100;
}

.handle-output:hover {
  transform: translateY(-50%) scale(1.15);
  box-shadow: var(--el-box-shadow-light);
}

.handle-add-icon {
  font-size: 14px;
  width: 14px;
  height: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.handle-output.success {
  background-color: var(--el-color-success);
}

.handle-output.error {
  background-color: var(--el-color-danger);
  bottom: -12px;
  transform: translateX(-50%);
  left: 50%;
  right: auto;
  top: auto !important;
}

.handle-output.warning {
  background-color: var(--el-color-warning);
}

.handle-output.info {
  background-color: var(--el-color-info);
}

.handle-output.neutral {
  background-color: var(--el-text-color-secondary);
}

.handle-output.true {
  background-color: var(--el-color-primary);
}

.handle-output.false {
  background-color: var(--el-color-warning);
}

.mynodes-badge {
  position: absolute;
  right: 0;
  background: linear-gradient(15deg, var(--el-color-primary), var(--el-color-success));
  color: white;
  font-size: 9px;
  font-weight: 900;
  padding: 2px 6px;
  border-radius: 0 0 0 10px;
}
</style>
