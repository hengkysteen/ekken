<template>
  <Teleport to="body">
    <Transition name="fade">
      <div v-if="show" class="node-context-menu" :style="menuStyle" @click.stop @contextmenu.prevent>
        <ul class="el-dropdown-menu">
          <li class=" el-dropdown-menu__item" :class="{ 'is-disabled': isDisabled('configure') }"
            @click="!isDisabled('configure') && handleSelect('configure')">
            <el-icon>
              <Setting />
            </el-icon>
            <span>Configure</span>
          </li>
          <li class="el-dropdown-menu__item" :class="{ 'is-disabled': isDisabled('duplicate') }"
            @click="!isDisabled('duplicate') && handleSelect('duplicate')">
            <el-icon>
              <CopyDocument />
            </el-icon>
            <span>Duplicate</span>
          </li>
          <li class="el-dropdown-menu__item" :class="{ 'is-disabled': isDisabled('save_to_mynodes') }"
            @click="!isDisabled('save_to_mynodes') && handleSelect('save_to_mynodes')">
            <el-icon>
              <Memo />
            </el-icon>
            <span>Save to My Nodes</span>
          </li>
          <li class="el-dropdown-menu__item el-dropdown-menu__item--divided danger" @click="handleSelect('delete')">
            <el-icon>
              <Delete />
            </el-icon>
            <span>Delete</span>
          </li>
        </ul>
      </div>
    </Transition>
  </Teleport>



</template>
<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { Setting, Delete, CopyDocument, Memo, } from '@element-plus/icons-vue'
const props = defineProps<{
  show: boolean
  x: number
  y: number
  nodeType?: string
  index?: number
  disabledKeys?: string[]
}>()
const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'select', key: string): void
}>()
const menuStyle = computed(() => {
  // Estimasi dimensi menu untuk kalkulasi posisi aman
  const menuWidth = 180
  const menuHeight = 180
  let x = props.x
  let y = props.y
  // Pastikan tidak keluar ke kanan
  if (x + menuWidth > window.innerWidth) {
    x = window.innerWidth - menuWidth - 5
  }
  // Pastikan tidak keluar ke bawah
  if (y + menuHeight > window.innerHeight) {
    y = window.innerHeight - menuHeight - 5
  }
  return {
    position: 'fixed' as const,
    top: `${y}px`,
    left: `${x}px`,
    zIndex: 9999,
  }
})
function isDisabled(key: string) {
  return props.disabledKeys?.includes(key) || false
}
function handleSelect(key: string) {
  emit('select', key)
  close()
}
function close() {
  emit('update:show', false)
}
const handleGlobalClose = () => {
  if (props.show) close()
}
onMounted(() => {
  window.addEventListener('click', handleGlobalClose)
  window.addEventListener('contextmenu', handleGlobalClose)
  window.addEventListener('scroll', handleGlobalClose, true)
  window.addEventListener('resize', handleGlobalClose)
})
onUnmounted(() => {
  window.removeEventListener('click', handleGlobalClose)
  window.removeEventListener('contextmenu', handleGlobalClose)
  window.removeEventListener('scroll', handleGlobalClose, true)
  window.removeEventListener('resize', handleGlobalClose)
})

</script>



<style scoped>
.node-context-menu {
  min-width: 180px;
  background-color: var(--el-bg-color-overlay);
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  box-shadow: var(--el-box-shadow-light);
  padding: 4px 0;
  pointer-events: auto;
  user-select: none;
}

.el-dropdown-menu__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  font-size: 13px;
  line-height: 1;
  color: var(--el-text-color-regular);
  cursor: pointer;
  transition: background-color 0.2s, color 0.2s;
}

.el-dropdown-menu__item:hover:not(.is-disabled) {
  background-color: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}

.el-dropdown-menu__item.is-disabled {
  color: var(--el-text-color-placeholder);
  cursor: not-allowed;
}

.el-dropdown-menu__item.danger {
  color: var(--el-color-danger);
}

.el-dropdown-menu__item.danger:hover:not(.is-disabled) {
  background-color: var(--el-color-danger-light-9);
}

.el-dropdown-menu__item--divided {
  margin-top: 5px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.el-icon {
  font-size: 16px;
}

/* Override Element Plus default dropdown styles for our custom div */
.el-dropdown-menu {
  padding: 0;
  margin: 0;
  list-style: none;
}

/* Simple Fade Transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
