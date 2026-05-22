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
      <el-tooltip v-if="profileStore.profile.pin_enabled" content="Lock app" placement="bottom">
        <el-button circle plain :icon="Lock" @click="profileStore.lockApp()" />
      </el-tooltip>
      <ThemeSwitcher />
    </el-space>
  </el-row>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { DArrowRight, Lock } from '@element-plus/icons-vue'
import ThemeSwitcher from './ThemeSwitcher.vue'
import ResourceMonitor from './ResourceMonitor.vue'
import { titleRegistry } from '@shared/utils/titleRegistry'
import { useAppSettingsStore } from '../stores/useAppSettingsStore'
import { useProfileStore } from '@profile/stores/profile'

const route = useRoute()
const router = useRouter()
const settingsStore = useAppSettingsStore()
const profileStore = useProfileStore()

onMounted(() => {
  settingsStore.startPolling()
})

type BreadcrumbItem = {
  label: string
  path?: string
}

const resolveBreadcrumbLabel = (label: string) => {
  if (!label.startsWith(':')) return label

  const idValue = route.params[label.slice(1)] as string
  return titleRegistry[idValue] || idValue || label
}

const breadcrumbs = computed(() => {
  const items: BreadcrumbItem[] = [{ label: 'Ekken', path: '/' }]
  const currentRecord = route.matched[route.matched.length - 1]
  if (!currentRecord) return items

  const explicitBreadcrumbs = currentRecord.meta.breadcrumbs as BreadcrumbItem[] | undefined
  if (explicitBreadcrumbs?.length) {
    explicitBreadcrumbs.forEach((breadcrumb, index) => {
      const path = breadcrumb.path ?? (index === explicitBreadcrumbs.length - 1 ? route.path : undefined)
      items.push({
        label: resolveBreadcrumbLabel(breadcrumb.label),
        path
      })
    })

    return items
  }

  const parentPath = currentRecord.meta.parent as string | undefined
  if (parentPath && parentPath !== currentRecord.path) {
    const parentRecord = router.getRoutes().find(r => r.path === parentPath)
    if (parentRecord && parentRecord.meta?.breadcrumb) {
      items.push({
        label: resolveBreadcrumbLabel(parentRecord.meta.breadcrumb as string),
        path: parentPath
      })
    }
  }

  if (currentRecord.meta.breadcrumb) {
    const label = resolveBreadcrumbLabel(currentRecord.meta.breadcrumb as string)

    if (!items.find(i => i.path === route.path && i.label === label)) {
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
