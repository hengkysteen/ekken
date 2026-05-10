<template>
  <div class="ek-text-display" :style="{ textAlign: textAlign }">
    <p :style="textStyle">{{ text }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  item?: any
}>()

const text = computed(() => props.item?.options?.text || props.item?.text)
const textAlign = computed(() => props.item?.options?.align || props.item?.align || 'left')

const textStyle = computed(() => ({
  fontSize: props.item?.options?.size || props.item?.size || '14px',
  color: props.item?.options?.color || props.item?.color || 'var(--el-text-color-regular)',
  fontWeight: (props.item?.options?.bold || props.item?.bold) ? 'bold' : 'normal',
  margin: 0
}))
</script>

<style scoped>
.ek-text-display {
  padding: 8px 0;
}
</style>
