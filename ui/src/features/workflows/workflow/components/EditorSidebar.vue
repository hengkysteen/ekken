<template>
  <div class="node-sidebar">
    <el-tabs v-model="activeTab" class="sidebar-tabs" stretch>
      <el-tab-pane label="Nodes" name="nodes">
        <div class="tab-content">
          <div class="sidebar-header">
            <el-input v-model="search" placeholder="Search nodes..." clearable>
              <template #prefix>
                <el-icon>
                  <Search />
                </el-icon>
              </template>
            </el-input>
          </div>
          <el-scrollbar class="sidebar-body">
            <el-collapse v-model="expandedNames">
              <el-collapse-item v-for="group in groupedNodes" :key="group.name" :title="group.name" :name="group.name">
                <div class="group-items">
                  <div v-for="node in group.nodes" :key="node.type"
                    :draggable="!(isInitialNodeRequired && !node.tags?.includes('Trigger'))"
                    @dragstart="onDragStart($event, node)">
                    <list-tile dense :disabled="isInitialNodeRequired && !node.tags?.includes('Trigger')">
                      <template #leading>
                        <img v-if="node.icon" :src="node.icon" alt="" class="sidebar-node-icon" :style="iconStyle" />
                      </template>
                      <template #title>
                        <el-text size="small" bold>{{ node.label }}</el-text>
                      </template>
                      <template #subtitle>
                        <el-text size="small" type="info" truncated>{{ node.description }}</el-text>
                      </template>
                    </list-tile>
                  </div>
                </div>
              </el-collapse-item>
            </el-collapse>
          </el-scrollbar>
        </div>
      </el-tab-pane>
      <el-tab-pane label="My Nodes" name="mynodes">
        <div class="tab-content">
          <div class="sidebar-header">
            <el-input v-model="mynodesSearch" placeholder="Search my nodes..." clearable>
              <template #prefix>
                <el-icon>
                  <Search />
                </el-icon>
              </template>
            </el-input>
          </div>
          <el-scrollbar class="sidebar-body">
            <div v-if="filteredMyNodesItems.length === 0" class="mynodes-empty">
              <el-empty :image-size="60" v-if="!mynodesSearch" description="No saved nodes yet">

              </el-empty>
              <el-empty v-else description="No matching nodes" />
            </div>
            <div v-else class="mynodes-items">
              <div v-for="item in filteredMyNodesItems" :key="item.id" draggable="true"
                @dragstart="onMyNodesDragStart($event, item)">
                <list-tile dense>
                  <template #leading>
                    <img v-if="item.icon" :src="item.icon" alt="" class="sidebar-node-icon" :style="iconStyle" />
                  </template>
                  <template #title>
                    <el-text size="small" bold>{{ item.name }}</el-text>
                  </template>
                  <template #subtitle>
                    <el-space :size="4">
                      <el-text size="small" type="primary" bold>{{ item.action?.key }}</el-text>
                    </el-space>
                  </template>
                  <template #trailing>
                    <el-button link circle size="small" @click.stop="onDeleteMyNodesItem(item.id)">
                      <el-icon>
                        <Delete />
                      </el-icon>
                    </el-button>
                  </template>
                </list-tile>
              </div>
            </div>
          </el-scrollbar>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Search, Delete } from '@element-plus/icons-vue'
import type { NodeDefinition } from '@workflows/node/types/node'
import type { MyNodesItem } from '@workflows/mynode/types'
import ListTile from '@shared/components/ListTile.vue'
import { useNodeStore } from '@workflows/node/stores/node'

type CatalogNode = NodeDefinition & { icon?: string }
type WorkflowNodeSummary = { id: string }
type GroupedNodes = { name: string; nodes: CatalogNode[] }

const nodeStore = useNodeStore()

const props = withDefaults(defineProps<{
  workflowNodes?: WorkflowNodeSummary[]
  mynodesItems?: MyNodesItem[]
  mynodesTabKey?: number
}>(), {
  workflowNodes: () => [],
  mynodesItems: () => [] as MyNodesItem[],
  mynodesTabKey: 0
})

const emit = defineEmits<{
  (e: 'delete-mynodes-item', id: string): void
  (e: 'add-node', node: any): void
}>()

const activeTab = ref('nodes')
const search = ref('')
const mynodesSearch = ref('')

// Watch mynodesTabKey to auto-switch to mynodes tab
watch(() => props.mynodesTabKey, () => {
  activeTab.value = 'mynodes'
})

const isInitialNodeRequired = computed(() => props.workflowNodes.length === 0)

import { useTheme } from '../../../../shared/composables/useTheme'
const { isDarkMode: isDark } = useTheme()
import { matchesSearch, matchesMyNodeSearch } from '../../node/utils/search'

const iconStyle = computed(() => ({
  filter: isDark.value ? 'invert(1) brightness(1.5)' : 'none'
}))

const filteredMyNodesItems = computed<MyNodesItem[]>(() => {
  return props.mynodesItems.filter(item => matchesMyNodeSearch(item, mynodesSearch.value))
})

const filteredCatalog = computed<CatalogNode[]>(() => {
  return nodeStore.catalog.filter(n => matchesSearch(n, search.value))
})

const groupedNodes = computed<GroupedNodes[]>(() => {
  const groups: Record<string, GroupedNodes> = {}
  for (const node of filteredCatalog.value) {
    const g = (node.tags && node.tags.length > 0) ? node.tags[0] : 'Other'
    if (!groups[g]) groups[g] = { name: g, nodes: [] }
    groups[g].nodes.push(node)
  }
  return Object.values(groups)
})

const expandedNames = ref<string[]>([])

watch(groupedNodes, (val) => {
  expandedNames.value = val.map(g => g.name)
}, { immediate: true })

function onDragStart(event: DragEvent, node: CatalogNode) {
  if (isInitialNodeRequired.value && !node.tags?.includes('Trigger')) {
    event.preventDefault()
    return
  }
  event.dataTransfer?.setData('application/vueflow', JSON.stringify({
    ...node,
    sourceType: 'catalog'
  }))
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
}

function onMyNodesDragStart(event: DragEvent, item: any) {
  const dragData = {
    type: item.type,
    label: item.label,
    tags: item.tags,
    icon: item.icon,
    action: item.action,
    sourceType: 'mynodes',
    name: item.name
  }
  event.dataTransfer?.setData('application/vueflow', JSON.stringify(dragData))
  if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
}

function onDeleteMyNodesItem(id: string) {
  emit('delete-mynodes-item', id)
}
</script>

<style scoped>
.node-sidebar {
  height: 100%;
  background-color: var(--el-bg-color);
  border-left: 1px solid var(--el-border-color);
}

.sidebar-tabs {
  height: 100%;
}

:deep(.el-tabs__header) {
  margin-bottom: 0;
}

:deep(.el-tabs__content),
:deep(.el-tab-pane) {
  height: 100%;
  overflow: hidden;
}

.tab-content {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.sidebar-body {
  flex: 1;
  min-height: 0;
  padding: 0 12px;
}

:deep(.el-collapse) {
  border: none;
}

:deep(.el-collapse-item__header) {
  font-weight: bold;
  font-size: 13px;
  color: var(--el-text-color-regular);
  height: 48px;
  background-color: transparent !important;
}

:deep(.el-collapse-item__wrap) {
  background-color: transparent !important;
  border-bottom: none;
}

:deep(.el-collapse-item__content) {
  padding-bottom: 12px;
  background-color: transparent !important;
}

:deep(.list-tile) {
  background-color: transparent !important;
}

.group-items {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.sidebar-node-icon {
  width: 24px;
  height: 24px;
  object-fit: contain;
  flex-shrink: 0;
}

.empty-hint {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  text-align: center;
  padding: 0 20px;
}

.mynodes-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 300px;
}

.mynodes-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 12px;
}
</style>
