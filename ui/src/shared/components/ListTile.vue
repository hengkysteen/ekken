<template>
  <div class="list-tile" :class="{
    'is-clickable': clickable && !disabled,
    'is-selected': selected,
    'is-disabled': disabled,
    'is-dense': dense
  }" @click="onTileClick">
    <!-- Leading Slot (Icon/Image) -->
    <div v-if="$slots.leading" class="list-tile-leading">
      <slot name="leading" />
    </div>

    <!-- Content Area -->
    <div class="list-tile-content">
      <div v-if="$slots.title" class="list-tile-title">
        <slot name="title" />
      </div>
      <div v-if="$slots.subtitle" class="list-tile-subtitle">
        <slot name="subtitle" />
      </div>
    </div>

    <!-- Trailing Slot (Buttons/Badge) -->
    <div v-if="$slots.trailing" class="list-tile-trailing" @click.stop>
      <slot name="trailing" />
    </div>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  clickable?: boolean
  selected?: boolean
  disabled?: boolean
  dense?: boolean
}>(), {
  clickable: true,
  selected: false,
  disabled: false,
  dense: false
})

const emit = defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

function onTileClick(event: MouseEvent) {
  if (!props.disabled) {
    emit('click', event)
  }
}
</script>

<style scoped>
.list-tile {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: var(--el-border-radius-base);
  border: 1px solid transparent;
  user-select: none;
  min-width: 0;
}

.list-tile.is-dense {
  padding: 6px 12px;
  gap: 10px;
}

.list-tile.is-clickable {
  cursor: pointer;
}

.list-tile.is-clickable:hover {
  background-color: var(--el-fill-color-light);
  border-color: var(--el-border-color-lighter);
}

.list-tile.is-selected {
  background-color: var(--el-color-primary-light-9);
  border-color: var(--el-color-primary-light-7);
}

.list-tile.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
  filter: grayscale(1);
}

.list-tile-leading {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.list-tile-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.list-tile-title {
  display: flex;
  align-items: center;
}

.list-tile-subtitle {
  display: flex;
  align-items: center;
  margin-top: 2px;
}

.list-tile-trailing {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  gap: 8px;
}

/* Base text styling override to ensure it looks good inside tile */
:deep(.list-tile-title .el-text) {
  font-weight: 600;
}

:deep(.el-text.is-truncated) {
  max-width: 100%;
}
</style>
