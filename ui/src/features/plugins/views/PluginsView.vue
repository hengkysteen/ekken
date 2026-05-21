<template>
  <AppPage title="Plugins" subtitle="Manage and monitor the status of installed plugins and their node types."
    scrollable>
    <template #header-extra>
      <el-button :icon="Cloudy" @click="$router.push({ name: 'plugins-hub' })"> Hub </el-button>
    </template>
    <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom: 20px;" />
    <div class="plugin-list-container">
      <el-empty v-if="initialized && !loading && plugins.length === 0" description="No plugins found" />
      <el-table v-else :data="plugins" style="width: 100%" class="plugin-table">
        <el-table-column label="Plugin Name" min-width="100">
          <template #default="{ row }">
            <div class="plugin-identity">
              <el-image style="width: 18px; height: 18px" :src="row.icon">
              </el-image>
              <span class="plugin-label">
                {{ row.id }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Source">
          <template #default="{ row }">
            {{ row.manifest.source }}
          </template>
        </el-table-column>
        <el-table-column label="Kind">
          <template #default="{ row }">
            {{ row.manifest.kind }}
          </template>
        </el-table-column>
        <el-table-column label="Status" align="center">
          <template #default="{ row }">
            {{ row.status }}
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="200" align="center">
          <template #default="{ row }">
            <el-button size="small" :type="row.is_enabled ? 'info' : 'primary'"
              @click="handleAction(row, row.is_enabled ? 'disable' : 'enable')">
              {{ row.is_enabled ? 'Disable' : 'Enable' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </AppPage>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Cloudy, } from '@element-plus/icons-vue'
import { pluginsApi as api } from '@plugins/api'
import type { PluginSummary } from '@plugins/api'
import AppPage from '@shared/components/AppPage.vue'
const loading = ref(false)
const initialized = ref(false)
const error = ref('')
const plugins = ref<PluginSummary[]>([])
async function loadPlugins() {
  loading.value = true
  error.value = ''
  try {
    plugins.value = await api.getPlugins()
  } catch (err: any) {
    error.value = err.message || 'Failed to load plugins'
  } finally {
    loading.value = false
    initialized.value = true
  }
}
const handleAction = async (row: PluginSummary, action: string) => {
  try {
    await api.managePlugin(row.id, action)
    ElMessage.success(`Plugin ${action}d successfully.`)
    loadPlugins() // Refresh list
  } catch (err: any) {
    ElMessage.error(err.message || `Failed to ${action} plugin`)
  }
}
onMounted(() => {
  loadPlugins()
})
</script>
<style scoped>
.plugin-identity {
  display: flex;
  align-items: center;
  gap: 10px;
}

.plugin-label {
  font-weight: 500;
}
</style>
