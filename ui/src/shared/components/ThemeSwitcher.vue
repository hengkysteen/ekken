<template>
  <el-tooltip :content="'Theme: ' + selectedMode" placement="bottom">
    <el-button circle plain class="theme-toggle-cycle" @click="toggleTheme">
      <div class="icon-wrapper">
        <Moon v-if="selectedMode === 'dark'" class="icon-pos" key="dark" />
        <Sunny v-else-if="selectedMode === 'light'" class="icon-pos" key="light" />
        <Monitor v-else class="icon-pos" key="auto" />
      </div>
    </el-button>
  </el-tooltip>
</template>

<script setup lang="ts">
import { Moon, Sunny, Monitor } from '@element-plus/icons-vue'
import { useStorage } from '@vueuse/core'
import { StorageKeys } from '../utils/storage'

const selectedMode = useStorage(StorageKeys.THEME_MODE, 'auto')

const toggleTheme = () => {
  if (selectedMode.value === 'light') {
    selectedMode.value = 'dark'
  } else if (selectedMode.value === 'dark') {
    selectedMode.value = 'auto'
  } else {
    selectedMode.value = 'light'
  }
}
</script>

<style scoped>
.theme-toggle-cycle {
  border: none;
  background: transparent;
  color: var(--el-text-color-secondary);
  width: 40px;
  height: 40px;
  padding: 0;
}

.theme-toggle-cycle:hover {
  background: var(--el-fill-color-light);
  color: var(--el-text-color-primary);
}

.icon-wrapper {
  position: relative;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
}

.icon-pos {
  position: absolute;
  width: 20px;
  height: 20px;
}
</style>
