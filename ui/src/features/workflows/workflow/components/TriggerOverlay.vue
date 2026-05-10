<template>
  <div class="trigger-overlay" @click.stop.prevent>


    <el-card shadow="never" class="trigger-card">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>Trigger</span>
          <el-icon :size="20" style="color: var(--el-text-color-placeholder);">
            <Lightning />
          </el-icon>
        </div>
      </template>

      <div style="text-align: center;">
        <el-text size="small" type="info">
          Pick a trigger to start.
        </el-text>
      </div>

      <template #footer>
        <div style="display: flex; justify-content: center;">
          <el-button type="primary" @click.stop.prevent="$emit('add-trigger')">
            <el-icon>
              <Plus />
            </el-icon>
            Add Trigger
          </el-button>
        </div>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { Lightning, Plus } from '@element-plus/icons-vue'

defineEmits<{
  (e: 'add-trigger'): void
}>()
</script>

<style scoped>
.trigger-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: transparent;
  z-index: 10;
  pointer-events: none;
}

.trigger-card {
  width: 200px;
  pointer-events: auto;
  padding-bottom: 2px;
}
</style>