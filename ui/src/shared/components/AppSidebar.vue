<template>
  <div class="sidebar-inner sidebar-transition" :class="{ 'collapsed': isCollapsed }" :style="{ width: sidebarWidth }">
    <div class="sidebar-header" @click="isCollapsed ? emit('toggle') : null">
      <AppLogo class="header-logo" />
      <span class="logo-text">Ekken</span>
      <el-button link class="menu-toggle" @click.stop="emit('toggle')">
        <el-icon :size="20">
          <Expand />
        </el-icon>
      </el-button>
    </div>

    <nav class="sidebar-nav">
      <div class="nav-section">
        <ListTile v-for="item in allMenuItems" :key="item.name" :clickable="true" :selected="activeKey === item.name"
          :style="{ margin: '6px 0' }" @click="router.push(item.path)" class="nav-item">
          <template #leading>
            <el-icon :size="20">
              <component :is="item.icon" />
            </el-icon>
          </template>
          <template #title>
            <span class="nav-text">{{ item.label }}</span>
          </template>
        </ListTile>
      </div>

      <div class="nav-divider">
        <span v-if="!isCollapsed" class="divider-text">System</span>
        <el-divider v-else />
      </div>

      <div class="nav-section">
        <ListTile v-for="item in allSystemItems" :key="item.name" :clickable="true" :selected="activeKey === item.name"
          :style="{ margin: '6px 0' }" @click="router.push(item.path)" class="nav-item">
          <template #leading>
            <el-icon :size="20">
              <component :is="item.icon" />
            </el-icon>
          </template>
          <template #title>
            <span class="nav-text">{{ item.label }}</span>
          </template>
        </ListTile>
      </div>
    </nav>

    <div class="sidebar-footer">
      <ListTile class="user-card" :clickable="true" :style="{ margin: '4px 0' }">
        <template #leading>
          <el-avatar :size="32" src="https://ui-avatars.com/api/?name=J+D&background=0D8ABC&color=fff" />
        </template>
        <template #title>
          <div class="user-info">
            <div class="user-name">Jhon Doe</div>
          </div>
        </template>
      </ListTile>
    </div>
  </div>
</template>



<script setup lang="ts">
import { computed, markRaw } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  House, Setting,
  Expand
} from '@element-plus/icons-vue'
import AppLogo from './AppLogo.vue'
import ListTile from './ListTile.vue'
import { useAppSidebarStore } from '../stores/useAppSidebarStore'

const sidebarStore = useAppSidebarStore()

const router = useRouter()

const props = withDefaults(defineProps<{
  isCollapsed: boolean,
  width?: string | number
}>(), {
  width: 240
})

const emit = defineEmits(['toggle'])
const route = useRoute()

const sidebarWidth = computed(() => {
  if (props.isCollapsed) return '84px'
  if (!props.width) return '260px'
  const w = props.width.toString()
  return w.endsWith('px') || w.endsWith('%') || w.endsWith('rem') ? w : `${w}px`
})

const menuItems = [
  { label: 'Dashboard', icon: markRaw(House), path: '/', name: 'dashboard' },
]

const systemItems = [
  { label: 'Settings', icon: markRaw(Setting), path: '/settings', name: 'settings', order: 200 },
]

// Dynamic items from modules
const dynamicItems = computed(() => sidebarStore.sortedItems)

const allMenuItems = computed(() => {
  // Main items: Hardcoded + Dynamic items with order < 100
  const dynamicMain = dynamicItems.value.filter(item => (item.order || 0) < 100)
  return [...menuItems, ...dynamicMain]
})

const allSystemItems = computed(() => {
  // System items: Hardcoded + Dynamic items with order >= 100
  const dynamicSystem = dynamicItems.value.filter(item => (item.order || 0) >= 100)
  return [...systemItems, ...dynamicSystem].sort((a, b) => (a.order || 0) - (b.order || 0))
})

const activeKey = computed(() => {
  if (route.path === '/') return 'dashboard'
  const found = [...allMenuItems.value, ...allSystemItems.value].find(i => i.path === route.path)
  return found?.name || ''
})
</script>



<style scoped>
.sidebar-inner {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 20px 16px;
}

/* Width is controlled by MainLayout aside, we just ensure content alignment */

/* Stable Layout Anchors */
.sidebar-header {
  display: grid;
  grid-template-columns: 52px 1fr auto;
  align-items: center;
  margin-bottom: 20px;
  height: 42px;
}

.header-logo {
  margin: 0 auto;
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  color: var(--el-color-primary);
}

.collapsed .header-logo {
  cursor: pointer;
}

.logo-text,
.nav-text,
.user-info,
.menu-toggle,
.divider-text {
  white-space: nowrap;
  overflow: hidden;
  opacity: 1;
}

.logo-text {
  font-size: 20px;
  font-weight: 800;
  color: var(--el-text-color-primary);
}

/* Collapsed States: Simply shrink the text column and hide internals */
.collapsed .logo-text,
.collapsed .nav-text,
.collapsed .user-info,
.collapsed .menu-toggle,
.collapsed .divider-text {
  max-width: 0;
  opacity: 0 !important;
  margin: 0;
  padding: 0 !important;
}

/* Force Icons to stay ABSOLUTELY STABLE in 52px column */
:deep(.list-tile-leading) {
  width: 52px;
  justify-content: center;
  flex-shrink: 0;
}

:deep(.list-tile) {
  gap: 0 !important;
  padding-left: 0 !important;
}

.collapsed :deep(.list-tile-content) {
  display: none !important;
}

.collapsed :deep(.list-tile) {
  padding-right: 0 !important;
  justify-content: flex-start;
}

.sidebar-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.nav-item {
  border-radius: 16px !important;
}

.nav-divider {
  height: 44px;
  display: flex;
  align-items: flex-end;
  padding-bottom: 8px;
}

.collapsed .nav-divider {
  align-items: flex-end;
  padding-bottom: 10px;
}

.nav-divider :deep(.el-divider) {
  margin: 0;
}

.divider-text {
  font-size: 11px;
  font-weight: 700;
  color: var(--el-text-color-placeholder);
  text-transform: uppercase;
  padding-left: 12px;
}

.sidebar-footer {
  margin-top: auto;
}

.user-card {
  border-radius: 12px !important;
}

.user-name {
  font-size: 13px;
  font-weight: 700;
}
</style>
