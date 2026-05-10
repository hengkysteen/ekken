<template>
  <el-row align="middle" justify="space-between" style="height: 52px; padding: 0 24px; width: 100%">
    <el-breadcrumb :separator-icon="DArrowRight">
      <el-breadcrumb-item v-for="(item, idx) in breadcrumbs" :key="idx"
        :to="idx < breadcrumbs.length - 1 ? item.path : undefined">
        <span
          :class="{ 'breadcrumb-current': idx === breadcrumbs.length - 1, 'breadcrumb-truncate': idx === breadcrumbs.length - 1 }">{{
            item.label }}</span>
      </el-breadcrumb-item>
    </el-breadcrumb>

    <el-space>
      <ResourceMonitor />
      <el-divider direction="vertical" />
      <ThemeSwitcher />
    </el-space>
  </el-row>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DArrowRight } from '@element-plus/icons-vue'
import ThemeSwitcher from './ThemeSwitcher.vue'
import ResourceMonitor from './ResourceMonitor.vue'
import { titleRegistry } from '@shared/utils/titleRegistry'
import { useAppSettingsStore } from '../stores/useAppSettingsStore'

const route = useRoute()
const router = useRouter()
const settingsStore = useAppSettingsStore()

onMounted(() => {
  settingsStore.startPolling()
})

const breadcrumbs = computed(() => {
  const items = [{ label: 'Ekken', path: '/' }]
  const currentRecord = route.matched[route.matched.length - 1]
  if (!currentRecord) return items

  const parentPath = currentRecord.meta.parent as string | undefined
  if (parentPath && parentPath !== currentRecord.path) {
    const parentRecord = router.getRoutes().find(r => r.path === parentPath)
    if (parentRecord && parentRecord.meta?.breadcrumb) {
      items.push({
        label: parentRecord.meta.breadcrumb as string,
        path: parentRecord.path
      })
    }
  }

  if (currentRecord.meta.breadcrumb) {
    let label = currentRecord.meta.breadcrumb as string
    if (label.startsWith(':')) {
      const idValue = route.params[label.slice(1)] as string
      // AUTO LOOKUP in our Global Registry
      label = titleRegistry[idValue] || idValue || label
    }

    if (!items.find(i => i.path === route.path)) {
      items.push({ label, path: route.path })
    }
  }

  return items
})
</script>

<style scoped>
.breadcrumb-truncate {
  max-width: 200px;
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
}
</style>