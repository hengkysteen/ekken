<template>
  <div class="master-layout" :class="{ 'sidebar-collapsed': isCollapsed }">
    <!-- 1. Global Floating Sidebar -->
    <aside class="floating-sidebar floating-card">
      <AppSidebar :is-collapsed="isCollapsed" @toggle="isCollapsed = !isCollapsed" />
    </aside>
    <!-- 2. Main Area -->
    <main class="main-workspace">
      <!-- 2.1 Floating TopBar -->
      <header class="floating-topbar floating-card">
        <AppTopBar />
      </header>
      <!-- 2.2 Content Area -->
      <div class="content-viewport">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useStorage } from '@vueuse/core'
import AppSidebar from '../shared/components/AppSidebar.vue'
import AppTopBar from '../shared/components/AppTopBar.vue'
import { StorageKeys } from '../shared/utils/storage'

const isCollapsed = useStorage<boolean>(StorageKeys.SIDEBAR_COLLAPSED, false)
</script>

<style scoped>
.master-layout {
  height: 100vh;
  width: 100vw;
  padding: var(--page-padding);
  display: flex;
  gap: var(--page-padding);
  overflow: hidden;
  background-color: var(--el-bg-color-page);
}

.floating-sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow: hidden;
  width: auto;
}


.main-workspace {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--page-padding);
  min-width: 0;
}

.content-viewport {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
</style>
