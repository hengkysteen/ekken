<template>
  <el-dialog v-model="isOpen" :title="title" width="480px" @closed="$emit('after-leave')" class="node-picker-dialog"
    destroy-on-close append-to-body>
    <el-container v-loading="loading">
      <el-header height="auto" style="padding: 0 0 16px 0;">
        <el-input v-model="currentSearch" :placeholder="activeTab === 'nodes' ? 'Search nodes...' : 'Search my nodes...'"
          clearable>
          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>
        </el-input>
      </el-header>

      <el-main style="padding: 0;">
        <el-tabs v-model="activeTab" stretch style="margin: 0;">
          <el-tab-pane label="Nodes" name="nodes">
            <el-scrollbar max-height="450px">
              <div v-if="filteredCatalogGroups.length === 0" style="padding: 40px 0; text-align: center;">
                <el-empty :description="search ? 'No matching nodes found' : 'No nodes available'" :image-size="60">
                  <template #image>
                    <el-icon color="var(--el-text-color-placeholder)" :size="64">
                      <Monitor />
                    </el-icon>
                  </template>
                </el-empty>
              </div>
              <div v-else v-for="group in filteredCatalogGroups" :key="group.name" style="margin-bottom: 8px;">
                <div
                  style="padding: 16px 12px 8px; font-size: 11px; font-weight: bold; text-transform: uppercase; color: var(--el-text-color-secondary); letter-spacing: 0.5px;">
                  {{ group.name }}
                </div>

                <div style="display: flex; flex-direction: column; gap: 2px;">
                  <list-tile v-for="node in group.nodes" :key="node.type"
                    @click="$emit('pick-node', { ...node, sourceType: 'catalog' })">
                    <template #leading>
                      <img v-if="node.icon" :src="node.icon" alt=""
                        :style="{ width: '24px', height: '24px', objectFit: 'contain', filter: isDark ? 'invert(1) brightness(1.5)' : 'none' }" />
                    </template>
                    <template #title>
                      <el-text bold>{{ node.label }}</el-text>
                    </template>
                    <template #subtitle>
                      <el-text type="info" size="small">{{ node.description }}</el-text>
                    </template>
                  </list-tile>
                </div>
              </div>
            </el-scrollbar>
          </el-tab-pane>

          <el-tab-pane label="My Nodes" name="mynodes">
            <el-scrollbar max-height="450px">
              <div v-if="filteredMyNodes.length === 0" style="padding: 40px 0; text-align: center;">
                <el-empty :description="mynodesSearch ? 'Workflow tidak ditemukan' : 'No saved nodes yet'" :image-size="60" />
              </div>
              <div v-else style="padding-top: 8px; display: flex; flex-direction: column; gap: 4px;">
                <list-tile v-for="item in filteredMyNodes" :key="item.id"
                  @click="$emit('pick-node', { ...item, sourceType: 'mynodes' })">
                  <template #leading>
                    <img v-if="item.icon" :src="item.icon" alt=""
                      :style="{ width: '24px', height: '24px', objectFit: 'contain', filter: isDark ? 'invert(1) brightness(1.5)' : 'none' }" />
                  </template>
                  <template #title>
                    <el-text bold>{{ item.name }}</el-text>
                  </template>
                  <template #subtitle>
                    <div style="display: flex; align-items: flex-start; gap: 8px; width: 100%; min-width: 0;">
                      <el-tag size="small" effect="plain" style="flex-shrink: 0;">{{ item.label }}</el-tag>
                      <el-text type="info" size="small" style="flex: 1;">{{ item.config?.action }}</el-text>
                    </div>
                  </template>
                </list-tile>
              </div>
            </el-scrollbar>
          </el-tab-pane>
        </el-tabs>
      </el-main>
    </el-container>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search, Monitor } from '@element-plus/icons-vue'
import ListTile from '@shared/components/ListTile.vue'
import type { MyNodesItem } from '@workflows/mynode/types'
import { useTheme } from '../../../../shared/composables/useTheme'
import type { NodeDefinition } from '../types/node'
import { matchesSearch, matchesMyNodeSearch } from '../utils/search'

interface GroupedNodes {
  name: string
  nodes: NodeDefinition[]
}

const props = withDefaults(defineProps<{
  show?: boolean
  loading?: boolean
  title?: string
  catalogGroups?: GroupedNodes[]
  mynodesItems?: MyNodesItem[]
}>(), {
  show: false,
  loading: false,
  title: 'Add Node',
  catalogGroups: () => [],
  mynodesItems: () => [] as MyNodesItem[]
})

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'pick-node', node: any): void
  (e: 'after-leave'): void
}>()

const { isDarkMode: isDarkRef } = useTheme()
const isDark = computed(() => isDarkRef.value)

const search = ref('')
const mynodesSearch = ref('')
const activeTab = ref('nodes')

const currentSearch = computed({
  get: () => activeTab.value === 'nodes' ? search.value : mynodesSearch.value,
  set: (val) => {
    if (activeTab.value === 'nodes') search.value = val
    else mynodesSearch.value = val
  }
})

const isOpen = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val)
})

const filteredCatalogGroups = computed(() => {
  return props.catalogGroups.map(group => ({
    ...group,
    nodes: group.nodes.filter(n => matchesSearch(n, search.value))
  })).filter(group => group.nodes.length > 0)
})

const filteredMyNodes = computed(() => {
  return props.mynodesItems.filter(item => matchesMyNodeSearch(item, mynodesSearch.value))
})
</script>

<style>
.node-picker-dialog .el-dialog__body {
  padding-top: 10px !important;
}

.node-picker-dialog .el-tabs__content {
  padding: 0 !important;
}
</style>
