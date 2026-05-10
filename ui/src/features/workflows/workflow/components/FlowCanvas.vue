<template>
  <div ref="canvasContainerRef" class="flow-canvas" @dragover="onDragOver" @drop="onDrop">
    <VueFlow v-model:nodes="editor.flowNodes" v-model:edges="editor.flowEdges" :min-zoom="0.1" :max-zoom="3"
      :nodes-connectable="true" :connect-on-click="true" :default-edge-type="currentEdgeType" :selection-key="null"
      :multi-selection-key="null" :delete-key="null" :multi-selection-active="false" @nodes-change="onNodesChange"
      @edges-change="onEdgesChange" @connect="onConnect" @pane-click="onPaneClick">

      <template #node-customNode="nodeProps">
        <NodeCardRaw :id="nodeProps.id" :data="nodeProps.data" :selected="selectedNodeId === nodeProps.id"
          :index="editor.flowNodes.findIndex((n: any) => n.id === nodeProps.id)" @node-click="onNodeClick"
          @configure="onConfigureNode" @delete="onDeleteNode" @duplicate="onDuplicateNode"
          @add-node="onAddNodeFromHandle" @save-to-mynodes="onSaveToMyNodes" />
      </template>

      <Background :pattern-color="gridColor" :gap="20" />
      <Controls :show-fit-view="true" :fit-view-options="{ padding: 0.3, maxZoom: 0.5 }">
        <ControlButton @click="showMinimap = !showMinimap" :title="showMinimap ? 'Hide Map' : 'Show Map'"
          :class="{ 'active-map': showMinimap }">
          <el-icon>
            <MapLocation />
          </el-icon>
        </ControlButton>
      </Controls>
      <MiniMap v-if="showMinimap" pannable zoomable />
    </VueFlow>

    <!-- Overlay & Modals -->
    <TriggerOverlay v-if="!editor.isLoading && editor.flowNodes.length === 0" @add-trigger="addFirstTrigger" />

    <BaseNodeForm v-model:visible="showConfigModal" :node="selectedNodeData" :catalog="nodeStore.catalog"
      @save="onConfigSave" @after-leave="selectedNodeData = null" />

    <NodePicker :show="showNodePicker" :catalog-groups="nodeStore.pickerGroups" :mynodes-items="mynodeStore.items"
      @pick-node="onPickNode" @after-leave="onPickerAfterLeave" @update:show="showNodePicker = $event" />

    <el-dialog v-model="showSaveToMyNodesModal" title="Save to My Nodes" width="340px"
      @after-leave="onMyNodesModalClose">
      <el-form label-position="top">
        <el-form-item label="Name">
          <el-input v-model="mynodesItemName" placeholder="Name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSaveToMyNodesModal = false">Cancel</el-button>
        <el-button type="primary" :disabled="!mynodesItemName.trim()" @click="confirmSaveToMyNodes">Save</el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch, markRaw } from 'vue'
import { VueFlow, useVueFlow, type EdgeChange, type NodeChange } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls, ControlButton } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import { storeToRefs } from 'pinia'
import { useMouseInElement } from '@vueuse/core'
import { useTheme } from '../../../../shared/composables/useTheme'
import NodeCard from '@workflows/node/components/NodeCard.vue'
import { MapLocation } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

import TriggerOverlay from '@workflows/workflow/components/TriggerOverlay.vue'
import NodePicker from '@workflows/node/components/NodePicker.vue'
import BaseNodeForm from '@workflows/node/nodes/common/BaseNodeForm.vue'
import { useWorkflowEditorStore } from '@workflows/workflow/stores/workflowEditor'
import { useMyNodeStore } from '@workflows/mynode/stores/mynode'
import { useNodeStore } from '@workflows/node/stores/node'

// Mark NodeCard as raw to avoid reactive overhead on component definition
const NodeCardRaw = markRaw(NodeCard)

const props = defineProps<{
  workflowId: string
}>()

const emit = defineEmits<{
  (e: 'canvas-ready', data: any): void
}>()

const editor = useWorkflowEditorStore()
const mynodeStore = useMyNodeStore()
const nodeStore = useNodeStore()
const { isDarkMode: isDark } = useTheme()
const { onMoveEnd, viewport, dimensions, setViewport, applyNodeChanges, applyEdgeChanges } = useVueFlow()
const { edgeStyle, edgeAnimated } = storeToRefs(editor)

const canvasContainerRef = ref<HTMLElement | null>(null)
const { elementX, elementY } = useMouseInElement(canvasContainerRef)
const showMinimap = ref(false)

// Update existing edges whenever style settings change
watch([edgeStyle, edgeAnimated], () => {
  if (editor.flowEdges.length > 0) {
    editor.flowEdges = editor.flowEdges.map(applyEdgeSettings)
  }
}, { immediate: true })

const currentEdgeType = computed(() => edgeStyle.value)
const selectedNodeId = ref<string | null>(null)

// UI State for Modals
const showConfigModal = ref(false)
const selectedNodeData = ref<any>(null)
const showNodePicker = ref(false)
const isPicking = ref(false)
const pendingSourceNodeId = ref<string | null>(null)
const pendingHandleType = ref<string | null>(null)

// Save to My Nodes state
const showSaveToMyNodesModal = ref(false)
const mynodesItemName = ref('')
const pendingSaveToMyNodesNode = ref<any>(null)

const viewportCenter = computed(() => ({
  x: (-viewport.value.x + dimensions.value.width / 2) / viewport.value.zoom,
  y: (-viewport.value.y + dimensions.value.height / 2) / viewport.value.zoom
}))

// Watch saved viewport and apply it
watch(() => editor.savedViewport, (newVal) => {
  if (newVal && typeof newVal.x === 'number') {
    setViewport({ x: newVal.x, y: newVal.y, zoom: newVal.zoom })
  }
}, { deep: true })

onMoveEnd(({ flowTransform }) => {
  editor.savedViewport = {
    x: flowTransform.x,
    y: flowTransform.y,
    zoom: flowTransform.zoom
  }
})

const gridColor = computed(() => isDark.value ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)')

function applyEdgeSettings(edge: any) {
  return {
    ...edge,
    type: edgeStyle.value,
    animated: edgeAnimated.value,
    style: { ...edge.style }
  }
}


function onNodesChange(changes: NodeChange[]) {
  // Enforce single selection: if any node is being selected, deselect all others in the changes array
  const selectionChange = changes.find(c => c.type === 'select' && c.selected) as any
  if (selectionChange) {
    changes.forEach((c: any) => {
      if (c.type === 'select' && c.id !== selectionChange.id) {
        c.selected = false
      }
    })
  }
  editor.flowNodes = applyNodeChanges(changes) as any[]
}

function onEdgesChange(changes: EdgeChange[]) {
  editor.flowEdges = applyEdgeChanges(changes) as any[]
}

function onConnect(connection: any) {
  editor.addEdge(connection.source, connection.sourceHandle, connection.target)
}

function onNodeClick(event: any) {
  const id = event.node ? event.node.id : (typeof event === 'string' ? event : event.id)
  selectedNodeId.value = id
}

function onPaneClick() {
  selectedNodeId.value = null
}

function onConfigureNode(nodeData: any) {
  selectedNodeId.value = nodeData.id
  selectedNodeData.value = { id: nodeData.id, data: { ...nodeData } }
  showConfigModal.value = true
}

function onConfigSave({ id, config, response_var, label }: any) {
  editor.updateNodeConfig(id, { config, response_var, label })
}

function onDeleteNode({ id }: any) {
  editor.removeNode(id)
  ElMessage.success('Node deleted')
}

function onDuplicateNode({ id }: any) {
  editor.duplicateNode(id)
  ElMessage.success('Node duplicated')
}

function onAddNodeFromHandle(event: any) {
  if (isPicking.value || showNodePicker.value) return
  pendingSourceNodeId.value = event.id
  pendingHandleType.value = event.type || event.handleType
  showNodePicker.value = true
}

function onPickNode(nodeDef: any) {
  if (isPicking.value) return
  isPicking.value = true

  const sourceId = pendingSourceNodeId.value
  const handleType = pendingHandleType.value

  showNodePicker.value = false
  setTimeout(() => {
    if (sourceId && handleType) {
      editor.addNodeFromHandle(sourceId, handleType, nodeDef)
    } else {
      editor.addNode({
        type: nodeDef.type,
        label: nodeDef.label,
        tags: nodeDef.tags,
        icon: nodeDef.icon,
        position: viewportCenter.value,
        config: nodeDef.config,
        response_var: nodeDef.response_var,
        sourceType: nodeDef.sourceType,
        name: nodeDef.name
      })
    }
  }, 50)
}

function onPickerAfterLeave() {
  pendingSourceNodeId.value = null
  pendingHandleType.value = null
  nodeStore.catalogFilterGroup = null
  isPicking.value = false
}

function addFirstTrigger() {
  if (isPicking.value || showNodePicker.value) return
  nodeStore.catalogFilterGroup = 'Trigger'
  showNodePicker.value = true
}

function onSaveToMyNodes({ id, data }: any) {
  pendingSaveToMyNodesNode.value = { id, data }
  mynodesItemName.value = ''
  showSaveToMyNodesModal.value = true
}

function onMyNodesModalClose() {
  pendingSaveToMyNodesNode.value = null
  mynodesItemName.value = ''
}

async function confirmSaveToMyNodes() {
  if (!mynodesItemName.value.trim() || !pendingSaveToMyNodesNode.value) return
  try {
    // Pastikan data yang disimpan bersih dari ID instance canvas lama
    const nodeData = { ...pendingSaveToMyNodesNode.value.data }
    if (nodeData.id) delete nodeData.id

    await mynodeStore.saveItem(mynodesItemName.value, nodeData || {})
    ElMessage.success('Saved to My Nodes')
    showSaveToMyNodesModal.value = false

    // Open sidebar and refresh
    editor.showSidebar = true
    editor.mynodesTabKey++
  } catch (err) {
    ElMessage.error('Failed to save')
  }
}

function onDragOver(event: DragEvent) {
  event.preventDefault()
  if (event.dataTransfer) event.dataTransfer.dropEffect = 'move'
}

function onDrop(event: DragEvent) {
  event.preventDefault()
  event.stopPropagation()
  const nodeData = event.dataTransfer?.getData('application/vueflow')
  if (!nodeData) return
  const node = JSON.parse(nodeData)
  const isFromMyNodes = node.sourceType === 'mynodes'
  if (editor.flowNodes.length === 0 && !node.tags?.includes('Trigger')) {
    ElMessage.warning('Please add a Trigger node first to start your workflow.')
    return
  }

  // Use VueUse element coordinates
  const x = elementX.value
  const y = elementY.value
  const { x: panX, y: panY, zoom } = viewport.value
  const position = { x: (x - panX) / zoom, y: (y - panY) / zoom }
  editor.addNode({ ...node, position, sourceType: isFromMyNodes ? 'mynodes' : 'catalog', name: node.name })
}

onMounted(() => {
  emit('canvas-ready', {})
})

defineExpose({ viewportCenter })
</script>

<style scoped>
.flow-canvas {
  flex: 1;
  height: 100%;
  position: relative;
  background-color: var(--el-bg-color);
}

.flow-canvas :deep(.vue-flow__node) {
  cursor: default;
}

.flow-canvas :deep(.vue-flow__edge.selected .vue-flow__edge-path) {
  stroke: var(--el-color-primary) !important;
  stroke-width: 3 !important;
}

.flow-canvas :deep(.vue-flow__edge-path) {
  stroke-width: 2;
  transition: stroke 0.2s ease, stroke-width 0.2s ease;
  stroke: var(--el-border-color-dark);
}

.flow-canvas :deep(.vue-flow__controls) {
  bottom: 16px;
  left: 16px;
  display: flex;
  gap: 8px;
}

.flow-canvas :deep(.vue-flow__controls-button) {
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  color: var(--el-text-color-regular);
  fill: currentColor;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 4px;
}

.flow-canvas :deep(.vue-flow__controls-button:hover) {
  background-color: var(--el-fill-color-light);
  border-color: var(--el-color-primary);
  color: var(--el-color-primary);
}

.flow-canvas :deep(.vue-flow__controls-button.active-map) {
  color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
  border-color: var(--el-color-primary);
}

.flow-canvas :deep(.vue-flow__minimap) {
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color);
  bottom: 60px;
  left: 16px;
  right: auto;
  border-radius: 8px;
  box-shadow: var(--el-box-shadow-light);
}
</style>
